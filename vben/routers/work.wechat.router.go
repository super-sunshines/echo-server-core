package routers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	eventCenter "github.com/super-sunshines/echo-server-core/vben/event"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/helper"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	WorkWechatRouterGroup = core.NewRouterGroup("/work-wechat", NewQywxAuthRouter, func(rg *echo.Group, group *core.RouterGroup) error {
		services.NewTencentWorkWeChatService()
		return group.Reg(func(m *QywxAuthRouter) {
			rg.GET("/login", m.login, core.IgnorePermission())
			rg.GET("/bind", m.bind, core.IgnorePermission(), core.Log("绑定企业微信"))
			rg.GET("/auth-url", m.authUrl, core.IgnorePermission())
			rg.GET("/signature", m.getConfigSignature, core.IgnorePermission())
		})
	})
)

type QywxAuthRouter struct {
	userService              core.PreGorm[model.SysUser, any]
	tencentWorkWeChatService *services.TencentWorkWeChatService
	thirdBindService         services.SysThirdBindService
}

func NewQywxAuthRouter() *QywxAuthRouter {
	return &QywxAuthRouter{
		userService:              core.NewService[model.SysUser, any](),
		tencentWorkWeChatService: services.NewTencentWorkWeChatService(),
		thirdBindService:         services.NewSysThirdBindService(),
	}
}

// @Summary	企业微信oauth2登录
// @Tags		[系统]三方授权
// @Success	200	{object}	core.ResponseSuccess{data=vo.OauthLoginVo}
// @Router		/work-wechat/login [get]
// @Param		code	query	string	true	"用户code"
func (r QywxAuthRouter) login(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	param := context.QueryParam("code")
	workWechat := core.GetConfig().Tencent.WorkWechat
	workWechatUserInfo, err := r.tencentWorkWeChatService.UserInfoByCode(param)
	if err != nil {
		zap.L().Error("获取用户信息失败", zap.Error(err))
		return context.Fail(core.NewFrontShowErrMsg("获取用户信息失败!请联系管理员"))
	}
	uid, exist := r.thirdBindService.ThirdPlatformUidToUid(_const.ThirdPlatformWorkWeChat, workWechatUserInfo.UserID)
	if !exist {
		err, userInfo := r.userService.WithContext(context).SkipGlobalHook().InsertOne(model.SysUser{
			Username:     workWechatUserInfo.UserID,
			Password:     core.HashPassword(workWechatUserInfo.UserID),
			NickName:     workWechatUserInfo.Name,
			RealName:     workWechatUserInfo.Name,
			RoleCodeList: workWechat.DefaultRoles,
			EnableStatus: int64(core.BooleanTo(workWechat.AutoRegister, _const.CommonStateOk, _const.CommonStateBanned)),
		})

		err, _ = r.thirdBindService.WithContext(ec).InsertOne(model.SysUserThirdBind{
			UserID:    userInfo.ID,
			LoginType: _const.ThirdPlatformWorkWeChat,
			Openid:    workWechatUserInfo.UserID,
		})
		if err != nil {
			return err
		}
		eventCenter.TencentWorkWeChatEventBus.Publish(eventCenter.TencentWorkWeChatNewUserEventBusKey, eventCenter.TencentWorkWeChatNewUserEventBusData{
			SysUid:           userInfo.ID,
			WorkWechatName:   workWechatUserInfo.Name,
			WorkWechatUserId: workWechatUserInfo.UserID,
		})
		if !workWechat.AutoRegister {
			return core.NewFrontShowErrMsg("请通知管理员为您开通账号,识别码:" + workWechatUserInfo.UserID)
		}
	}

	uid, _ = r.thirdBindService.ThirdPlatformUidToUid(_const.ThirdPlatformWorkWeChat, workWechatUserInfo.UserID)
	err, useInfo := r.userService.WithContext(context).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", uid).Where("")
	})
	if err != nil {
		return err
	}
	return context.Success(vo.OauthLoginVo{
		AccessToken: helper.GenJwtByUserInfo(context.GetAppPlatformCode(), useInfo),
	})
}

// @Summary	企业微信绑定
// @Tags		[系统]三方授权
// @Success	200	{object}	core.ResponseSuccess{data=string}
// @Router		/work-wechat/bind [post]
func (r QywxAuthRouter) bind(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	return context.Fail(core.NewFrontShowErrMsg("暂未实现！"))
}

// @Summary	获取授权链接
// @Tags		[系统]三方授权
// @Success	200	{object}	core.ResponseSuccess{data=string}
// @Router		/work-wechat/auth-url [get]
// @Param		code	query	string	true	"用户code"
func (r QywxAuthRouter) authUrl(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	path := r.tencentWorkWeChatService.GetAuthUrl()
	return context.Success(path)
}

// @Summary	获取Signature
// @Tags		[系统]三方授权
// @Success	200	{object}	core.ResponseSuccess{data=string}
// @Router		/work-wechat/signature [get]
// @Param		code	query	string	true	"用户code"
func (r QywxAuthRouter) getConfigSignature(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	param := context.QueryParam("url")
	ticket, err := r.tencentWorkWeChatService.GetJSAPITicket()
	timestamp := core.GetNowTimeUnix()
	randomStr := core.GetRandomStr(10)
	fullUrl := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticket, randomStr, timestamp, param)
	encryptedStr := core.SHA1Encrypt(fullUrl)
	if err != nil {
		zap.L().Error("获取jsapi_ticket失败", zap.Error(err))
		return
	}
	return context.Success(vo.SignatureVo{
		Timestamp: timestamp,
		NonceStr:  randomStr,
		Signature: encryptedStr,
	})
}
