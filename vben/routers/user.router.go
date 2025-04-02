package routers

import (
	"echo-server-core/core"
	"echo-server-core/vben/bo"
	_const "echo-server-core/vben/const"
	"echo-server-core/vben/gorm/model"
	"echo-server-core/vben/services"
	"echo-server-core/vben/vo"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var SysUserRouterGroup = core.NewRouterGroup("/system/user", NewSysUserRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysUserRouter) {
		rg.GET("/list", m.SysUserList, core.IgnorePermission())
		rg.GET("/options", m.optionsList, core.IgnorePermission())
		rg.GET("/:id", m.SysUserDetail, core.IgnorePermission())
		rg.PUT("/:id", m.SysUserUpdate, core.IgnorePermission())
		rg.POST("", m.SysUserAdd, core.IgnorePermission())
		rg.DELETE("", m.SysUserDelete, core.IgnorePermission())
	})
})

type SysUserRouter struct {
	SysUserService       core.PreGorm[model.SysUser, vo.SysUserVo]
	SysDepartmentService services.SysDepartmentService
}

func NewSysUserRouter() *SysUserRouter {
	return &SysUserRouter{
		SysUserService:       core.NewService[model.SysUser, vo.SysUserVo](),
		SysDepartmentService: services.NewDepartmentService(),
	}
}

// @Summary	用户下拉列表
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=[]vo.UserOptionsVo}
// @Router		/system/user/options [get]
func (receiver SysUserRouter) optionsList(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	err, userList := receiver.SysUserService.WithContext(ec).SkipGlobalHook().
		FindList()
	context.CheckError(err)
	vos := slice.Map(userList, func(index int, item model.SysUser) vo.UserOptionsVo {
		return vo.UserOptionsVo{
			Uid:      item.ID,
			RealName: item.RealName,
		}
	})

	return context.Success(vos)
}

// SysUserList
// @Summary	系统用户列表
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysUserVo]}
// @Router		/system/user/list [GET]
// @Param		bo	query	bo.SysUserPageBo	true	"分页参数"
func (receiver SysUserRouter) SysUserList(c echo.Context) error {
	context := core.GetContext[bo.SysUserPageBo](c)
	pageBo := context.GetQueryParamAndValid()
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
			core.BooleanFun(pageBo.DepartmentId != 0, func() {
				children, err := receiver.SysDepartmentService.GetChildren(c, pageBo.DepartmentId)
				context.CheckError(err)
				db.Where("department_id in (?)", children)
			})
			return db
		})
	context.CheckError(err)
	return context.Success(x)
}

// SysUserDetail
// @Summary	系统用户详情
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysUserVo}
// @Router		/system/user/:id [GET]
// @Param		id	path	int	true	"id"
func (receiver SysUserRouter) SysUserDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	id := context.GetPathParamInt64("id")
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().FindOneVoByPrimaryKey(id)
	context.CheckError(err)
	return context.Success(x)
}

// SysUserUpdate
// @Summary	系统用户更新
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/user/:id [PUT]
// @Param		id	path	int				true	"id"
// @Param		bo	body	bo.SysUserBo	true	"修改参数"
func (receiver SysUserRouter) SysUserUpdate(c echo.Context) error {
	context := core.GetContext[bo.SysUserBo](c)
	updateBo := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	from := core.CopyFrom[model.SysUser](updateBo)
	core.BooleanFun(from.Status == _const.CommonStateBanned, func() {
		core.GetTokenManager().RemoveTokenByUid(id)
	})
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(id, from, "password")
	context.CheckError(err)
	return context.Success(x)
}

// SysUserAdd
// @Summary	系统用户新增
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysUserVo}
// @Router		/system/user [POST]
// @Param		bo	body	bo.SysUserBo	true	"新增参数"
func (receiver SysUserRouter) SysUserAdd(c echo.Context) error {
	context := core.GetContext[bo.SysUserBo](c)
	addBo := context.GetBodyAndValid()
	err, meta := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysUser](addBo))
	context.CheckError(err)
	return context.Success(core.CopyFrom[vo.SysUserVo](meta))
}

// SysUserDelete
// @Summary	系统用户删除
// @Tags		[系统]用户模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/user [DELETE]
// @Param		param	query	core.QueryIds	true	"删除参数"
func (receiver SysUserRouter) SysUserDelete(c echo.Context) error {
	context := core.GetContext[any](c)
	ids := context.QueryParamIds()
	err, row := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	context.CheckError(err)
	return context.Success(row)
}
