package routers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/helper"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
)

var AuthRouterGroup = core.NewRouterGroup("", NewAuthRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *AuthRouter) {
		rg.POST("/auth/login", m.login, core.IgnorePermission())
		rg.GET("/auth/check", m.checkToken, core.IgnorePermission())
		rg.GET("/auth/logout", m.logout, core.IgnorePermission())
		rg.GET("/menu/all", m.menu, core.IgnorePermission())
		rg.GET("/user/info", m.loginUserInfo, core.IgnorePermission())
	})
})

type AuthRouter struct {
	userService     core.PreGorm[model.SysUser, any]
	menuService     core.PreGorm[model.SysMenu, any]
	loginLogService services.SysLoginInfoService
}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{
		userService:     core.NewService[model.SysUser, any](),
		menuService:     core.NewService[model.SysMenu, any](),
		loginLogService: services.NewSysLoginInfoService(),
	}
}

// @Summary	用户登录
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.LoginVo}
// @Router		/auth/login [post]
// @Param		loginBo	body	bo.LoginBo	true	"登录参数"
func (r AuthRouter) login(ec echo.Context) (err error) {
	context := core.GetContext[bo.LoginBo](ec)
	maxLoginFailCount := core.GetConfig().Jwt.MaxLoginFailCount
	loginInfo, err := context.GetBodyAndValid()
	if err != nil {
		return context.Fail(err)
	}
	err, a := r.userService.WithContext(context).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", loginInfo.Username)
	})
	if a.EnableStatus == _const.CommonStateBanned || a.LoginFailCount >= maxLoginFailCount {
		r.loginLogService.AddLog(ec, loginInfo.Username, _const.LoginTypePassword, 2, "账户已锁定，请联系管理员解锁！")
		return context.Fail(core.NewFrontShowErrMsg("账户已锁定，请联系管理员解锁！"))
	}
	if err != nil {
		r.loginLogService.AddLog(ec, loginInfo.Username, _const.LoginTypePassword, 2, "用户名或者密码错误！")
		return context.Fail(core.NewFrontShowErrMsg("用户名或者密码错误！"))
	}
	if core.ComparePasswords(a.Password, loginInfo.Password) {
		// 重置登录失败次数
		r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
			"login_fail_count": 0,
			"last_online":      core.GetNowTimeUnixMilli(),
		})
		r.loginLogService.AddLog(ec, loginInfo.Username, _const.LoginTypePassword, 1, "登录成功")
		return context.Success(vo.LoginVo{AccessToken: helper.GenJwtByUserInfo(context.GetAppPlatformCode(), a)})
	} else {
		r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
			"login_fail_count": gorm.Expr("login_fail_count + 1"),
		})
		if a.LoginFailCount+1 >= maxLoginFailCount {
			r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
				"enable_status": _const.CommonStateBanned,
			})
		}
		r.loginLogService.AddLog(ec, loginInfo.Username, _const.LoginTypePassword, 2, "用户名或者密码错误！"+fmt.Sprintf("尝试第%d次", a.LoginFailCount+1))
		return context.Fail(core.NewFrontShowErrMsg("用户名或者密码错误！"))
	}
}

// @Summary	检测token
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/check [get]
func (r AuthRouter) checkToken(ec echo.Context) (err error) {
	context := core.GetContext[bo.LoginBo](ec)
	uid, err := context.GetLoginUserUid()
	return context.Success(core.GetTokenManager().ValidToken(uid, context.GetAppPlatformCode(), context.GetUserToken()))
}

// @Summary	登出
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/logout [get]
func (r AuthRouter) logout(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	param := context.QueryParam("accessToken")
	if param == "" {
		return context.Success(true)
	} else {
		jwt, _ := core.GetTokenManager().ParseJwt(param)
		core.GetTokenManager().RemoveToken(jwt.UID, context.GetAppPlatformCode())
		return context.Success(true)
	}
}

// @Summary	获取目录列表
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=[]vo.SysMenuWithMetaVo}
// @Router		/menu/all [GET]
func (r AuthRouter) menu(c echo.Context) error {
	context := core.GetContext[any](c)
	user, err := context.GetLoginUser()
	if err != nil || user.RoleCodes == nil {
		return context.Fail(err)
	}
	menuIds := core.PermissionMange.GetRoleMenuIdList(user.RoleCodes...)
	service := core.NewService[vo.SysMenuWithMeta, vo.SysMenuWithMeta]()
	err, metas := service.WithContext(c).SkipGlobalHook().FindList(func(db *gorm.DB) *gorm.DB {
		return db.Where("id in (?)", menuIds).Where("type in ?", _const.MenuTreeType).Preload("Meta")
	})
	if err != nil {
		return context.Fail(err)
	}
	userMenuVoList := core.CopyListFrom[vo.SysMenuWithMetaVo](metas)
	return context.Success(vo.BuildTree(userMenuVoList))
}

// @Summary	用户基本信息
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.LoginUserInfoVo}
// @Router		/user/info [get]
// @Param		loginBo	body	bo.LoginBo	true	"登录参数"
func (r AuthRouter) loginUserInfo(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	user, err := context.GetLoginUser()
	if err != nil {
		return context.Fail(err)
	}
	err, loginUserInfo := r.userService.WithContext(ec).SkipGlobalHook().FindOneByPrimaryKey(user.UID)
	var HomePath string = "/"
	if len(user.RoleCodes) > 0 {
		HomePath = core.PermissionMange.GetRoleHomePath(user.RoleCodes[0])
	}
	return context.Success(vo.LoginUserInfoVo{
		UserId:   loginUserInfo.ID,
		RealName: loginUserInfo.RealName,
		Roles:    user.RoleCodes,
		Username: loginUserInfo.Username,
		HomePath: HomePath,
	})
}
