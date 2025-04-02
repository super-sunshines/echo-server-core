package routers

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/bo"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/vo"
	"github.com/labstack/echo/v4"
)

var SysRoleRouterGroup = core.NewRouterGroup("/system/role", NewRoleRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *RoleRouter) {
		rg.GET("/list", m.roleList, core.IgnorePermission())
		rg.GET("/:id", m.roleDetail, core.IgnorePermission())
		rg.POST("", m.addRole, core.IgnorePermission())
		rg.PUT("/:id", m.updateRole, core.IgnorePermission())
		rg.DELETE("", m.deleteRole, core.IgnorePermission())
	})
})

type RoleRouter struct {
	roleService core.PreGorm[model.SysRole, vo.SysRoleVo]
}

func NewRoleRouter() *RoleRouter {
	return &RoleRouter{
		roleService: core.NewService[model.SysRole, vo.SysRoleVo](),
	}
}

// @Summary	角色列表
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysRoleVo]}
// @Router		/system/role/list [GET]
// @Param		bo	query	bo.SysRolePageBo	true	"请求参数"
func (r RoleRouter) roleList(ec echo.Context) (err error) {
	context := core.GetContext[bo.SysRolePageBo](ec)
	queryParam := context.GetQueryParamAndValid()
	err, roleList := r.roleService.WithContext(ec).SkipGlobalHook().
		FindVoListByPage(queryParam.PageParam)
	context.CheckError(err)
	return context.Success(roleList)
}

// @Summary	角色详情
// @Tags		[系统]角色模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysRoleVo}
// @Router		/system/role/:id [GET]
// @Param		id	path	int	true	"id"
func (r RoleRouter) roleDetail(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	paramInt64 := context.GetPathParamInt64("id")
	err, roleVo := r.roleService.WithContext(ec).SkipGlobalHook().
		FindOneVoByPrimaryKey(paramInt64)
	context.CheckError(err)
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
	roleBo := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	var insertValues = core.CopyFrom[model.SysRole](roleBo)
	err, _ := r.roleService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(id, insertValues)
	context.CheckError(err)
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
	roleBo := context.GetBodyAndValid()
	err, meta := r.roleService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysRole](roleBo))
	context.CheckError(err)
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
	ids := context.QueryParamIds()
	err, row := r.roleService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	context.CheckError(err)
	_ = core.PermissionMange.Refresh()
	return context.Success(row)
}
