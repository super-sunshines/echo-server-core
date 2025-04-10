package routers

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
)

var SysUserRouterGroup = core.NewRouterGroup("/system/user", NewSysUserRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysUserRouter) {
		rg.GET("/list", m.SysUserList, core.HavePermission("SYS::USER::QUERY"), core.Log("用户分页列表"))
		rg.GET("/options", m.optionsList, core.HavePermission("SYS::USER::OPTIONS"), core.Log("用户下拉列表"))
		rg.GET("/:id", m.SysUserDetail, core.HavePermission("SYS::USER::QUERY"), core.Log("查询用户"))
		rg.PUT("/:id", m.SysUserUpdate, core.HavePermission("SYS::USER::UPDATE"), core.Log("修改用户"))
		rg.POST("", m.SysUserAdd, core.HavePermission("SYS::USER::ADD"), core.Log("新增用户"))
		rg.DELETE("", m.SysUserDelete, core.HavePermission("SYS::USER::DEL"), core.Log("删除用户"))
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

//	@Summary	用户下拉列表
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=[]vo.UserOptionsVo}
//	@Router		/system/user/options [get]
func (receiver SysUserRouter) optionsList(ec echo.Context) (err error) {
	context := core.GetContext[any](ec)
	err, userList := receiver.SysUserService.WithContext(ec).SkipGlobalHook().
		FindList()
	if err != nil {
		return err
	}
	vos := slice.Map(userList, func(index int, item model.SysUser) vo.UserOptionsVo {
		return vo.UserOptionsVo{
			Uid:      item.ID,
			RealName: item.RealName,
		}
	})

	return context.Success(vos)
}

// SysUserList
//
//	@Summary	系统用户列表
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysUserVo]}
//	@Router		/system/user/list [GET]
//	@Param		bo	query	bo.SysUserPageBo	true	"分页参数"
func (receiver SysUserRouter) SysUserList(c echo.Context) error {
	context := core.GetContext[bo.SysUserPageBo](c)
	pageBo, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
			core.BooleanFun(pageBo.DepartmentId != 0, func() {
				children, err := receiver.SysDepartmentService.GetChildren(c, pageBo.DepartmentId)
				if err != nil {
					c.Error(err)
					return
				}
				db.Where("department_id in (?)", children)
			})
			return db
		})
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysUserDetail
//
//	@Summary	系统用户详情
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysUserVo}
//	@Router		/system/user/:id [GET]
//	@Param		id	path	int	true	"id"
func (receiver SysUserRouter) SysUserDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().FindOneVoByPrimaryKey(id)
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysUserUpdate
//
//	@Summary	系统用户更新
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/user/:id [PUT]
//	@Param		id	path	int				true	"id"
//	@Param		bo	body	bo.SysUserBo	true	"修改参数"
func (receiver SysUserRouter) SysUserUpdate(c echo.Context) error {
	context := core.GetContext[bo.SysUserBo](c)
	updateBo, err := context.GetBodyAndValid()
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	from := core.CopyFrom[model.SysUser](updateBo)
	core.BooleanFun(from.EnableStatus == _const.CommonStateBanned, func() {
		core.GetTokenManager().RemoveTokenByUid(id)
	})
	err, x := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(id, from, "password")
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysUserAdd
//
//	@Summary	系统用户新增
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysUserVo}
//	@Router		/system/user [POST]
//	@Param		bo	body	bo.SysUserBo	true	"新增参数"
func (receiver SysUserRouter) SysUserAdd(c echo.Context) error {
	context := core.GetContext[bo.SysUserBo](c)
	addBo, err := context.GetBodyAndValid()
	if err != nil {
		return err
	}
	err, meta := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		InsertOne(core.CopyFrom[model.SysUser](addBo))
	if err != nil {
		return err
	}
	return context.Success(core.CopyFrom[vo.SysUserVo](meta))
}

// SysUserDelete
//
//	@Summary	系统用户删除
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/user [DELETE]
//	@Param		param	query	core.QueryIds	true	"删除参数"
func (receiver SysUserRouter) SysUserDelete(c echo.Context) error {
	context := core.GetContext[any](c)
	ids, err := context.QueryParamIds()
	if err != nil {
		return err
	}
	err, row := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		DeleteByPrimaryKeys(ids)
	if err != nil {
		return err
	}
	return context.Success(row)
}
