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
		rg.GET("/list", m.SysUserList, core.Log("用户分页列表"), core.HavePermission("SYS::USER::QUERY"))
		rg.GET("/options", m.optionsList, core.Log("用户下拉列表"), core.HavePermission("SYS::USER::OPTIONS"))
		rg.GET("/:id", m.SysUserDetail, core.HavePermission("SYS::USER::QUERY"), core.Log("查询用户"))
		rg.GET("/simple-list", m.SysUserSimpleList, core.IgnorePermission())
		rg.PUT("/:id", m.SysUserUpdate, core.HavePermission("SYS::USER::UPDATE"), core.Log("修改用户"))
		rg.POST("", m.SysUserAdd, core.HavePermission("SYS::USER::ADD"), core.Log("新增用户"))
		rg.DELETE("", m.SysUserDelete, core.HavePermission("SYS::USER::DEL"), core.Log("删除用户"))
		rg.PUT("/unlock/:id", m.SysUserUnLock, core.HavePermission("SYS::USER::UNLOCK"), core.Log("解锁用户"))
		rg.PUT("/lock/:id", m.SysUserLock, core.HavePermission("SYS::USER::LOCK"), core.Log("封禁用户"))
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

	// 从echo上下文中获取自定义的上下文对象
	context := core.GetContext[any](ec)

	// 调用用户服务获取用户列表
	// WithContext(ec): 将echo上下文传递给服务层
	// SkipGlobalHook(): 跳过全局钩子
	// FindList(): 执行查询获取用户列表
	err, userList := receiver.SysUserService.WithContext(ec).SkipGlobalHook().
		FindList()
	if err != nil {
		return err
	}

	// 将用户模型列表映射为用户选项值对象列表
	// 使用slice.Map函数转换每个用户对象
	vos := slice.Map(userList, func(index int, item model.SysUser) vo.UserOptionsVo {
		return vo.UserOptionsVo{
			Uid:      item.ID,       // 用户ID
			RealName: item.RealName, // 用户真实姓名
		}
	})

	// 返回成功的响应，包含用户选项列表
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
					_ = context.Fail(err)
					return
				}
				db.Where("department_id in (?)", children)
			})
			core.BooleanFun(pageBo.SearchKey != "", func() {
				db.Where("real_name like ?", "%"+pageBo.SearchKey+"%").Or("nick_name like ?", "%"+pageBo.SearchKey+"%")
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
	from := core.CopyFrom[model.SysUser](addBo)
	from.Password = core.HashPassword(from.Username + "123!")
	from.NeedChangePassword = true
	err, meta := receiver.SysUserService.WithContext(c).SkipGlobalHook().
		InsertOne(from)
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

// SysUserUnLock
//
//	@Summary	系统用户解锁
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=bool}
//	@Router		/system/user/unlock/:id [put]
func (receiver SysUserRouter) SysUserUnLock(c echo.Context) error {
	context := core.GetContext[any](c)
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	tx := receiver.SysUserService.WithContext(c).SkipGlobalHook().Where("id = ?", id).Updates(map[string]any{
		"enable_status":    _const.CommonStateOk,
		"login_fail_count": 0,
	})
	return context.Success(tx.RowsAffected > 0)
}

// SysUserLock
//
//	@Summary	系统用户封禁
//	@Tags		[系统]用户模块
//	@Success	200	{object}	core.ResponseSuccess{data=bool}
//	@Router		/system/user/lock/:id [put]
func (receiver SysUserRouter) SysUserLock(c echo.Context) error {
	context := core.GetContext[any](c)
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	tx := receiver.SysUserService.WithContext(c).SkipGlobalHook().Where("id = ?", id).Updates(map[string]any{
		"enable_status":    _const.CommonStateBanned,
		"login_fail_count": 100,
	})
	if tx.RowsAffected == 0 {
		return core.NewFrontShowErrMsg("封禁失败！")
	}
	return context.Success(true)
}

// SysUserSimpleList 系统用户简单列表
func (receiver SysUserRouter) SysUserSimpleList(c echo.Context) error {
	context := core.GetContext[any](c)
	err, users := receiver.SysUserService.WithContext(c).SkipGlobalHook().FindList()
	if err != nil {
		return err
	}
	var userList = core.CopyListFrom[vo.SimpleUserVo](users)
	return context.Success(userList)
}
