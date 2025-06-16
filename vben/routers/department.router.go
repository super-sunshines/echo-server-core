package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
)

var SysDepartmentRouterGroup = core.NewRouterGroup("/system/department", NewSysDepartmentRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysDepartmentRouter) {
		rg.GET("/list", m.SysDepartmentList, core.Log("部门列表"), core.HavePermission("SYS::DEPART::QUERY"))
		rg.GET("/tree", m.SysDepartmentTree, core.Log("部门树形列表"), core.HavePermission("SYS::DEPART::QUERY"))
		rg.GET("/:id", m.SysDepartmentDetail, core.Log("部门详情"), core.HavePermission("SYS::DEPART::QUERY"))
		rg.PUT("/:id", m.SysDepartmentUpdate, core.Log("部门修改"), core.HavePermission("SYS::DEPART::UPDATE"))
		rg.POST("", m.SysDepartmentAdd, core.Log("部门新增"), core.HavePermission("SYS::DEPART::ADD"))
		rg.DELETE("", m.SysDepartmentDelete, core.Log("部门删除"), core.HavePermission("SYS::DEPART:DEL"))
	})
})

type SysDepartmentRouter struct {
	SysDepartmentService core.PreGorm[model.SysDepartment, vo.SysDepartmentVo]
	clearCache           func()
}

func NewSysDepartmentRouter() *SysDepartmentRouter {
	return &SysDepartmentRouter{
		SysDepartmentService: core.NewService[model.SysDepartment, vo.SysDepartmentVo](),
		clearCache:           services.NewDepartmentService().ClearCache,
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
	pageBo, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		return err
	}
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
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().FindOneVoByPrimaryKey(id)
	if err != nil {
		return err
	}
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
	updateBo, err := context.GetBodyAndValid()
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	err, x := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().SaveByPrimaryKey(id, core.CopyFrom[model.SysDepartment](updateBo))
	if err != nil {
		return err
	}
	receiver.clearCache()
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
	addBo, err := context.GetBodyAndValid()

	if err != nil {
		return err
	}

	err, meta := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysDepartment](addBo))

	if err != nil {
		return err
	}
	receiver.clearCache()
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
	ids, err := context.QueryParamIds()
	if err != nil {
		return err
	}
	err, row := receiver.SysDepartmentService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	if err != nil {
		return err
	}
	receiver.clearCache()
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
	if err != nil {
		return err
	}
	return context.Success(vo.ToSysDepartmentTreeVo(departments))
}
