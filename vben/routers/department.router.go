package routers

import (
	"echo-server-core/core"
	"echo-server-core/vben/bo"
	"echo-server-core/vben/gorm/model"
	"echo-server-core/vben/vo"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var SysDepartmentRouterGroup = core.NewRouterGroup("/system/department", NewSysDepartmentRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysDepartmentRouter) {
		rg.GET("/list", m.SysDepartmentList, core.IgnorePermission())
		rg.GET("/tree", m.SysDepartmentTree, core.IgnorePermission())
		rg.GET("/:id", m.SysDepartmentDetail, core.IgnorePermission())
		rg.PUT("/:id", m.SysDepartmentUpdate, core.IgnorePermission())
		rg.POST("", m.SysDepartmentAdd, core.IgnorePermission())
		rg.DELETE("", m.SysDepartmentDelete, core.IgnorePermission())
	})
})

type SysDepartmentRouter struct {
	SysDepartmentService core.PreGorm[model.SysDepartment, vo.SysDepartmentVo]
}

func NewSysDepartmentRouter() *SysDepartmentRouter {
	return &SysDepartmentRouter{
		SysDepartmentService: core.NewService[model.SysDepartment, vo.SysDepartmentVo](),
	}
}

// SysDepartmentList
//
//	@Summary	系统部门列表
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysDepartmentVo]}
//	@Router		/system/department/list [GET]
//	@Param		bo	query	bo.SysDepartmentPageBo	true	"分页参数"
func (receiver SysDepartmentRouter) SysDepartmentList(c echo.Context) error {
	context := core.GetContext[bo.SysDepartmentPageBo](c)
	pageBo := context.GetQueryParamAndValid()
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
		return db
	})
	context.CheckError(err)
	return context.Success(x)
}

// SysDepartmentDetail
//
//	@Summary	系统部门详情
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysDepartmentVo}
//	@Router		/system/department/:id [GET]
//	@Param		id	path	int	true	"id"
func (receiver SysDepartmentRouter) SysDepartmentDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	id := context.GetPathParamInt64("id")
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().FindOneVoByPrimaryKey(id)
	context.CheckError(err)
	return context.Success(x)
}

// SysDepartmentUpdate
//
//	@Summary	系统部门更新
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/department/:id [PUT]
//	@Param		id	path	int					true	"id"
//	@Param		bo	body	bo.SysDepartmentBo	true	"修改参数"
func (receiver SysDepartmentRouter) SysDepartmentUpdate(c echo.Context) error {
	context := core.GetContext[bo.SysDepartmentBo](c)
	updateBo := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().SaveByPrimaryKey(id, core.CopyFrom[model.SysDepartment](updateBo))
	context.CheckError(err)
	return context.Success(x)
}

// SysDepartmentAdd
//
//	@Summary	系统部门新增
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysDepartmentVo}
//	@Router		/system/department [POST]
//	@Param		bo	body	bo.SysDepartmentBo	true	"新增参数"
func (receiver SysDepartmentRouter) SysDepartmentAdd(c echo.Context) error {
	context := core.GetContext[bo.SysDepartmentBo](c)
	addBo := context.GetBodyAndValid()
	err, meta := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysDepartment](addBo))
	context.CheckError(err)
	return context.Success(core.CopyFrom[vo.SysDepartmentVo](meta))
}

// SysDepartmentDelete
//
//	@Summary	系统部门删除
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/department [DELETE]
//	@Param		param	query	core.QueryIds	true	"删除参数"
func (receiver SysDepartmentRouter) SysDepartmentDelete(c echo.Context) error {
	context := core.GetContext[any](c)
	ids := context.QueryParamIds()
	err, row := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	context.CheckError(err)
	return context.Success(row)
}

// SysDepartmentTree
//
//	@Summary	系统部门树形列表
//	@Tags		[系统]部门模块
//	@Success	200	{object}	core.ResponseSuccess{data=[]vo.SysDepartmentTreeVo}
//	@Router		/system/department/tree [GET]
func (receiver SysDepartmentRouter) SysDepartmentTree(c echo.Context) error {
	context := core.GetContext[any](c)
	err, departments := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().FindVoList()
	context.CheckError(err)
	return context.Success(vo.ToSysDepartmentTreeVo(departments))
}
