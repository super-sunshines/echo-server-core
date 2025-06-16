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
	"time"
)

var AuthRouterGroup = core.NewRouterGroup("", NewAuthRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *AuthRouter) {
		rg.POST("/auth/login", m.login, core.IgnorePermission())
		rg.GET("/auth/codes", m.codes)
		rg.GET("/auth/check", m.checkToken, core.IgnorePermission())
		rg.POST("/password/change", m.passwordReset, core.Log("用户重置密码"), core.IgnorePermission())
		rg.GET("/auth/password/reset/check/:code", m.passWordResetCheck, core.IgnorePermission())
		rg.POST("/auth/password/reset", m.passWordReset, core.IgnorePermission())
		rg.POST("/auth/refresh", m.refreshToken, core.IgnorePermission())
		rg.GET("/auth/logout", m.logout, core.IgnorePermission())
		rg.GET("/menu/all", m.menu, core.IgnorePermission())
		rg.GET("/user/info", m.loginUserInfo, core.IgnorePermission())
		rg.PUT("/user/info", m.updateInfo, core.IgnorePermission())
	})
})

type AuthRouter struct {
	services.SysUserService
	menuService         core.PreGorm[model.SysMenu, any]
	userService         core.PreGorm[model.SysUser, any]
	departmentService   services.SysDepartmentService
	loginLogService     services.SysLoginInfoService
	changePasswordCache *core.RedisCache[int64]
}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{
		SysUserService:      services.NewSysUserService(),
		userService:         core.NewService[model.SysUser, any](),
		menuService:         core.NewService[model.SysMenu, any](),
		loginLogService:     services.NewSysLoginInfoService(),
		changePasswordCache: core.GetRedisCache[int64]("user:change:password"),
		departmentService:   services.NewDepartmentService(),
	}
}

// @Summary	用户登录
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.LoginVo}
// @Router		/auth/login [post]
// @Param		loginBo	body	bo.LoginBo	true	"登录参数"
func (r AuthRouter) login(ec echo.Context) (err error) {
	context := core.GetContext[bo.LoginBo](ec)
	platform := context.GetAppPlatformCode()
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

		token, err := helper.GenJwtByUserInfo(platform, a)
		if err != nil {
			return err
		}
		loginVo := vo.LoginVo{
			AccessToken:        token,
			NeedChangePassword: bool(a.NeedChangePassword),
		}
		if a.NeedChangePassword {
			str := core.GetRandomStr(12)
			loginVo.ChangePasswordCode = str
			r.changePasswordCache.Set(context, "sys:user:change:password:"+str, a.ID, 5*time.Minute)
		}
		return context.Success(loginVo)
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
		return context.Fail(core.NewFrontShowErrMsg("用户名或者密码错误！" + fmt.Sprintf("第%d次", a.LoginFailCount+1) + fmt.Sprintf("共%d次", maxLoginFailCount)))
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

// @Summary	检测token
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/check [get]
func (r AuthRouter) refreshToken(ec echo.Context) (err error) {
	context := core.GetContext[bo.LoginBo](ec)
	uid, err := context.GetLoginUserUid()
	return context.Success(core.GetTokenManager().ValidToken(uid, context.GetAppPlatformCode(), context.GetUserToken()))
}

// @Summary	登出
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/logout [get]
// @Param		accessToken	query	string	true	"token"
func (r AuthRouter) logout(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	param := context.QueryParam("accessToken")
	jwt, _ := core.GetTokenManager().ParseJwt(param)
	core.GetTokenManager().RemoveToken(jwt.UID, context.GetAppPlatformCode())
	return context.Success(true)
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
	department := r.departmentService.GetUserDepartment(context)
	return context.Success(vo.LoginUserInfoVo{
		UserId:     loginUserInfo.ID,
		RealName:   loginUserInfo.RealName,
		Roles:      user.RoleCodes,
		Username:   loginUserInfo.Username,
		NickName:   loginUserInfo.NickName,
		Avatar:     loginUserInfo.Avatar,
		HomePath:   HomePath,
		Department: department.Name,
	})
}

// @Summary	更新个人信息
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/user/info [put]
// @Param		bo	body	bo.UpdateUserInfoBo	true	"修改参数"
func (r AuthRouter) updateInfo(ec echo.Context) (err error) {
	context := core.GetContext[bo.UpdateUserInfoBo](ec)
	body, err := context.GetBodyAndValid()
	uid, err := context.GetLoginUserUid()
	if err != nil {
		return err
	}
	r.userService.WithContext(ec).GetModelDb().Where("id = ?", uid).
		Updates(map[string]any{
			"nick_name": body.NickName,
			"avatar":    body.Avatar,
		})
	r.RemoveCacheById(uid)
	return context.Success(true)
}

// @Summary	检测改密码code
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/password/reset/check/:code [get]
// @Param		code	path	string	true	"修改参数"
func (r AuthRouter) passWordResetCheck(c echo.Context) error {
	context := core.GetContext[any](c)
	code := context.GetPathParam("code")
	if code == "" {
		return context.Fail(core.NewFrontShowErrMsg("参数错误"))
	}
	exists := r.changePasswordCache.Exists(context, "sys:user:change:password:"+code)
	result, err := exists.Result()
	if err != nil {
		return err
	}
	return context.Success(result >= 1)
}

// @Summary	新用户密码修改
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/auth/password/reset [post]
// @Param		bo	body	bo.ChangePasswordByCodeBo	true	"修改参数"
func (r AuthRouter) passWordReset(c echo.Context) error {
	context := core.GetContext[bo.ChangePasswordByCodeBo](c)
	body, err := context.GetBodyAndValid()
	if err != nil {
		return err
	}
	var redisKey = "sys:user:change:password:" + body.Code
	exists := r.changePasswordCache.Exists(context, redisKey)
	result, err := exists.Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return context.Fail(core.NewFrontShowErrMsg("校验码失效"))
	}

	get := r.changePasswordCache.Get(context, redisKey)
	uid, err := get.Int64()
	if err != nil {
		return err
	}

	err, user := r.userService.WithContext(context).FindOneByPrimaryKey(uid)
	if err != nil {
		return err
	}

	if !core.ComparePasswords(user.Password, body.OldPassword) {
		return context.Fail(core.NewFrontShowErrMsg("旧密码错误"))
	}

	tx := r.userService.WithContext(context).Where("id = ?", uid).Updates(map[string]any{
		"password":             core.HashPassword(body.NewPassword),
		"need_change_password": core.IntBoolFalse,
	})

	r.changePasswordCache.Del(context, redisKey)
	return context.Success(tx.RowsAffected > 0)
}

func (r AuthRouter) codes(c echo.Context) error {
	context := core.GetContext[any](c)
	return context.Success([]string{""})
}

// @Summary	用户主动修改密码
// @Tags		[系统]授权模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/password/change [post]
// @Param		bo	body	bo.ChangePasswordBo	true	"修改参数"
func (r AuthRouter) passwordReset(c echo.Context) error {
	context := core.GetContext[bo.ChangePasswordBo](c)
	body, err := context.GetBodyAndValid()
	if err != nil {
		return err
	}
	user, err := context.GetLoginUser()
	if err != nil {
		return err
	}
	err, sysUser := r.userService.WithContext(c).FindOneByPrimaryKey(user.UID)
	if err != nil {
		return err
	}
	if core.ComparePasswords(sysUser.Password, body.OldPassword) {
		tx := r.userService.WithContext(c).Where("id = ?", user.UID).Updates(map[string]any{
			"password": core.HashPassword(body.NewPassword),
		})
		return context.Success(tx.RowsAffected > 0)
	} else {
		return core.NewFrontShowErrMsg("旧密码错误！")
	}
}
