package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/services"
)

var SysRoleRouterGroup = core.NewRouterGroup("/system/role", NewRoleRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *RoleRouter) {
		rg.GET("/list", m.roleList, core.HavePermission("SYS::ROLE::QUERY"), core.Log("角色列表"))
		rg.GET("/:id", m.roleDetail, core.HavePermission("SYS::ROLE::QUERY"), core.Log("角色详情"))
		rg.POST("", m.addRole, core.HavePermission("SYS::ROLE::ADD"), core.Log("角色新增"))
		rg.PUT("/:id", m.updateRole, core.HavePermission("SYS::ROLE::UPDATE"), core.Log("角色修改"))
		rg.DELETE("", m.deleteRole, core.HavePermission("SYS::ROLE::DEL"), core.Log("角色删除"))
	})
})

type RoleRouter struct {
	roleService services.SysRoleService
}

func NewRoleRouter() *RoleRouter {
	return &RoleRouter{
		roleService: services.NewSysRoleService(),
	}
}

// @Summary	角色列表
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysRoleVo]}
// @Router		/system/role/list [GET]
// @Param		bo	query	bo.SysRolePageBo	true	"请求参数"
func (r RoleRouter) roleList(ec echo.Context) (err error) {
	context := core.GetContext[bo.SysRolePageBo](ec)
	queryParam, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, roleList := r.roleService.WithContext(ec).SkipGlobalHook().
		FindVoListByPage(queryParam.PageParam)
	if err != nil {
		return err
	}
	return context.Success(roleList)
}

// @Summary	角色详情
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysRoleVo}
// @Router		/system/role/:id [GET]
// @Param		id	path	int	true	"id"
func (r RoleRouter) roleDetail(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	paramInt64, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	err, roleVo := r.roleService.WithContext(ec).SkipGlobalHook().
		FindOneVoByPrimaryKey(paramInt64)
	if err != nil {
		return err
	}
	return context.Success(roleVo)
}

// @Summary	更新角色详情
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/role/:id [PUT]
// @Param		bo	body	bo.SysRoleBo	true	"更新考角色详情参数"
// @Param		id	path	int				true	"主键"
func (r RoleRouter) updateRole(c echo.Context) error {
	context := core.GetContext[bo.SysRoleBo](c)
	roleBo, err := context.GetBodyAndValid()

	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	var insertValues = core.CopyFrom[model.SysRole](roleBo)
	err, _ = r.roleService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(id, insertValues)
	if err != nil {
		return err
	}
	r.roleService.RefreshCache()
	_ = core.PermissionMange.Refresh()
	return context.Success(true)
}

// @Summary	新增角色
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysRoleVo}
// @Router		/system/role [POST]
// @Param		bo	body	bo.SysRoleBo	true	"更新参数"
func (r RoleRouter) addRole(c echo.Context) error {
	context := core.GetContext[bo.SysRoleBo](c)
	roleBo, err := context.GetBodyAndValid()
	if err != nil {
		return err
	}
	err, meta := r.roleService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysRole](roleBo))
	if err != nil {
		return err
	}
	r.roleService.RefreshCache()
	_ = core.PermissionMange.Refresh()
	return context.Success(meta)
}

// @Summary	删除角色
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/role [DELETE]
// @Param		param	query	core.QueryIds	true	"删除参数"
func (r RoleRouter) deleteRole(c echo.Context) error {
	context := core.GetContext[any](c)
	ids, err := context.QueryParamIds()
	if err != nil {
		return err
	}

	err, row := r.roleService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	if err != nil {
		return err
	}
	r.roleService.RefreshCache()
	_ = core.PermissionMange.Refresh()
	return context.Success(row)
}
