package routers

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/bo"
	_const "github.com/XiaoSGentle/echo-server-core/vben/const"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/helper"
	"github.com/XiaoSGentle/echo-server-core/vben/vo"
	"github.com/labstack/echo/v4"
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
	userService core.PreGorm[model.SysUser, any]
	menuService core.PreGorm[model.SysMenu, any]
}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{
		userService: core.NewService[model.SysUser, any](),
		menuService: core.NewService[model.SysMenu, any](),
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
	loginInfo := context.GetBodyAndValid()
	err, a := r.userService.WithContext(context).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", loginInfo.Username)
	})
	if a.Status == _const.CommonStateBanned || a.LoginFailCount >= maxLoginFailCount {
		return context.Fail(core.NewFrontShowErrMsg("账户已锁定，请联系管理员解锁！"))
	}
	if err != nil {
		context.CheckError(core.NewFrontShowErrMsg("用户名或者密码错误！"))
	}
	if core.ComparePasswords(a.Password, loginInfo.Password) {
		// 重置登录失败次数
		r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
			"login_fail_count": 0,
			"last_online":      core.GetNowTimeUnixMilli(),
		})
		return context.Success(vo.LoginVo{AccessToken: helper.GenJwtByUserInfo(context.GetAppPlatformCode(), a)})
	} else {
		r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
			"login_fail_count": gorm.Expr("login_fail_count + 1"),
		})
		if a.LoginFailCount+1 >= maxLoginFailCount {
			r.userService.WithContext(ec).SkipGlobalHook().Where("id = ?", a.ID).UpdateColumns(map[string]any{
				"status": _const.CommonStateBanned,
			})
		}
		return context.Fail(core.NewFrontShowErrMsg("用户名或者密码错误！"))
	}
}

// @Summary	检测token
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/check [get]
func (r AuthRouter) checkToken(ec echo.Context) (err error) {
	context := core.GetContext[bo.LoginBo](ec)
	return context.Success(core.GetTokenManager().ValidToken(context.GetLoginUserUid(), context.GetAppPlatformCode(), context.GetUserToken()))
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
	var metas []vo.SysMenuWithMeta
	codes := context.GetLoginUser().RoleCodes
	menuIds := core.PermissionMange.GetRoleMenuIdList(codes...)
	tx := r.menuService.WithContext(c).SkipGlobalHook().
		Model(vo.SysMenuWithMeta{}).Preload("Meta").
		Where("id in (?)", menuIds).
		Where("type in ?", _const.MenuTreeType).Find(&metas)
	context.CheckError(tx.Error)
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
	err, loginUserInfo := r.userService.WithContext(ec).SkipGlobalHook().FindOneByPrimaryKey(context.GetLoginUserUid())
	roleCodes := context.GetLoginUser().RoleCodes
	return context.Success(vo.LoginUserInfoVo{
		UserId:   loginUserInfo.ID,
		RealName: loginUserInfo.RealName,
		Roles:    roleCodes,
		Username: loginUserInfo.Username,
		HomePath: core.BooleanTo(len(roleCodes) != 0, core.PermissionMange.GetRoleHomePath(roleCodes[0]), "/"),
	})
}
