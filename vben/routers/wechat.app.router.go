package routers

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/helper"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
	"time"
)

var (
	WechatAppRouterGroup = core.NewRouterGroup("/wechat-app", NewWechatAppAuthRouter, func(rg *echo.Group, group *core.RouterGroup) error {
		services.NewTencentWorkWeChatService()
		return group.Reg(func(m *WechatAppAuthRouter) {
			rg.GET("/login", m.login, core.IgnorePermission(), core.Log("微信小程序授权登录"))
		})
	})
)

type WechatAppAuthRouter struct {
	userService              core.PreGorm[model.SysUser, any]
	tencentWorkWeChatService *services.TencentWorkWeChatService
	thirdBindService         services.SysThirdBindService
	RequestClient            *resty.Client
}

func NewWechatAppAuthRouter() *WechatAppAuthRouter {
	return &WechatAppAuthRouter{
		userService:              core.NewService[model.SysUser, any](),
		tencentWorkWeChatService: services.NewTencentWorkWeChatService(),
		thirdBindService:         services.NewSysThirdBindService(),
		RequestClient: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(5*time.Second).SetHeader("Content-Type", "application/json"),
	}
}

// @Summary	企业微信oauth2登录
// @Tags		[系统]三方授权
// @Success	200	{object}	core.ResponseSuccess{data=vo.OauthLoginVo}
// @Router		/work/wx/login [get]
// @Param		code	query	string	true	"用户code"
func (r WechatAppAuthRouter) login(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	var Config = core.GetConfig().Tencent.WechatApp
	jsCode := context.QueryParam("code")
	resp, err := r.RequestClient.R().
		SetQueryParams(
			map[string]string{
				"appid":      Config.AppId,
				"secret":     Config.AppSecret,
				"js_code":    jsCode,
				"grant_type": "authorization_code",
			}).
		Get("https://api.weixin.qq.com/sns/jscode2session")
	if err != nil {
		return core.NewFrontShowErrMsg("授权登录失败！")
	}
	var result bo.WechatAppServerResp
	if _ = json.Unmarshal(resp.Body(), &result); result.Errcode != 0 {
		return core.NewFrontShowErrMsg(fmt.Sprintf("授权登录失败！%s", result.Errmsg))
	}
	var code = result.Openid
	uid, exist := r.thirdBindService.ThirdPlatformUidToUid(_const.ThirdPlatformWeChatApp, code)
	if !exist {
		_, userInfo := r.userService.WithContext(context).SkipGlobalHook().InsertOne(model.SysUser{
			Username:     code,
			Password:     core.HashPassword(code),
			NickName:     Config.DefaultNickName,
			RealName:     Config.DefaultNickName,
			Avatar:       Config.DefaultAvatar,
			EnableStatus: _const.CommonStateOk,
		})

		_err, _ := r.thirdBindService.WithContext(ec).InsertOne(model.SysUserThirdBind{
			UserID:    userInfo.ID,
			LoginType: _const.ThirdPlatformWeChatApp,
			Openid:    code,
		})
		if _err != nil {
			return core.NewFrontShowErrMsg("注册失败!")
		}
	}
	uid, _ = r.thirdBindService.ThirdPlatformUidToUid(_const.ThirdPlatformWeChatApp, code)
	err, useInfo := r.userService.WithContext(context).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", uid)
	})
	if useInfo.EnableStatus == _const.CommonStateBanned {
		return core.NewErrCodeMsg(core.USER_STARTUS_ERROR, "账号已被禁用！")
	}
	if err != nil {
		return err
	}
	return context.Success(vo.OauthLoginVo{
		AccessToken: helper.GenJwtByUserInfo(context.GetAppPlatformCode(), useInfo),
	})
}
