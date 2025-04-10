package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/helper"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
)

var (
	QywxRouterGroup = core.NewRouterGroup("/work/wx", NewQywxAuthRouter, func(rg *echo.Group, group *core.RouterGroup) error {
		helper.InitCacheStore()
		return group.Reg(func(m *AuthRouter) {
			rg.GET("/login", m.userinfo, core.IgnorePermission())
			rg.GET("/bind", m.bind, core.IgnorePermission(), core.Log("绑定企业微信"))
			rg.GET("/auth-url", m.getAuthUrl, core.IgnorePermission())
		})
	})
)

type QywxAuthRouter struct {
	userService core.PreGorm[model.SysUser, any]
}

func NewQywxAuthRouter() *AuthRouter {
	return &AuthRouter{
		userService: core.NewService[model.SysUser, any](),
	}
}

//	@Summary	企业微信oauth2登录
//	@Tags		[系统]三方授权
//	@Success	200	{object}	core.ResponseSuccess{data=vo.OauthLoginVo}
//	@Router		/work/wx/login [get]
//	@Param		code	query	string	true	"用户code"
func (r AuthRouter) userinfo(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	param := context.QueryParam("code")
	qywxUserInfo, err := helper.GetUserInfoByCode(param)
	if err != nil {
		return context.Fail(core.NewFrontShowErrMsg("获取用户信息失败!请联系管理员"))
	}
	count := r.userService.WithContext(context).SkipGlobalHook().Count(func(db *gorm.DB) *gorm.DB {
		return db.Where("qywx_uid = ?", qywxUserInfo.UserID)
	})
	if count == 0 {
		//return context.Success(vo.OauthLoginVo{
		//	NeedRegister: true,
		//	QywxUid:      code.UserID,
		//})
		err, _ := r.userService.WithContext(context).SkipGlobalHook().InsertOne(model.SysUser{
			Username: qywxUserInfo.UserID,
			Password: core.HashPassword(qywxUserInfo.UserID),
			NickName: qywxUserInfo.Name,
			RealName: qywxUserInfo.Name,
		})
		if err != nil {
			return err
		}
	}
	err, useInfo := r.userService.WithContext(context).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("qywx_uid = ?", qywxUserInfo.UserID)
	})
	if err != nil {
		return err
	}
	return context.Success(vo.OauthLoginVo{
		NeedRegister: false,
		AccessToken:  helper.GenJwtByUserInfo(context.GetAppPlatformCode(), useInfo),
	})
}

//	@Summary	企业微信绑定
//	@Tags		[系统]三方授权
//	@Success	200	{object}	core.ResponseSuccess{data=string}
//	@Router		/work/wx/bind [post]
func (r AuthRouter) bind(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	return context.Fail(core.NewFrontShowErrMsg("暂未实现！"))
}

//	@Summary	获取授权链接
//	@Tags		[系统]三方授权
//	@Success	200	{object}	core.ResponseSuccess{data=string}
//	@Router		/work/wx/auth-url [get]
//	@Param		code	query	string	true	"用户code"
func (r AuthRouter) getAuthUrl(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	path := helper.GetAuthUrl()
	return context.Success(path)
}
