package routers

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/bo"
	_const "github.com/XiaoSGentle/echo-server-core/vben/const"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/vo"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var SysDictRouterGroup = core.NewRouterGroup("/system/dict", NewSysDictRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysDictRouter) {
		rg.GET("/list", m.SysDictList, core.IgnorePermission())
		rg.GET("/code-list", m.SysDictCodeList, core.IgnorePermission())
		rg.GET("/code-exist", m.SysDictExist, core.IgnorePermission())
		rg.GET("/:id", m.SysDictDetail, core.IgnorePermission())
		rg.GET("/code/:code", m.SysDictCodeDetail, core.IgnorePermission())
		rg.PUT("/:id", m.SysDictUpdate, core.IgnorePermission())
		rg.POST("", m.SysDictAdd, core.IgnorePermission())
		rg.DELETE("", m.SysDictDelete, core.IgnorePermission())
		rg.GET("/child/list", m.SysDictChildList, core.IgnorePermission())
		rg.PUT("/child/:code", m.SysDictChildUpdate, core.IgnorePermission())
	})
})

type SysDictRouter struct {
	SysDictService      core.PreGorm[model.SysDict, vo.SysDictVo]
	SysDictChildService core.PreGorm[model.SysDictChild, vo.SysDictChildVo]
	SysDictRedisCache   *core.RedisCache[vo.SysDictVo]
}

func NewSysDictRouter() *SysDictRouter {
	return &SysDictRouter{
		SysDictService:      core.NewService[model.SysDict, vo.SysDictVo](),
		SysDictChildService: core.NewService[model.SysDictChild, vo.SysDictChildVo](),
		SysDictRedisCache:   core.GetRedisCache[vo.SysDictVo]("sys-dict-content"),
	}
}

// SysDictList
// @Summary	[系统]字典列表
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysDictVo]}
// @Router		/system/dict/list [GET]
// @Param		bo	query	bo.SysDictPageBo	true	"分页参数"
func (receiver SysDictRouter) SysDictList(c echo.Context) error {
	context := core.GetContext[bo.SysDictPageBo](c)
	pageBo := context.GetQueryParamAndValid()
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
			core.BooleanFun(pageBo.Module != 0, func() {
				db.Where("module = ?", pageBo.Module)
			})
			return db
		})
	context.CheckError(err)
	return context.Success(x)
}

// SysDictCodeDetail
// @Summary	[系统]字典详情[code]
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
// @Router		/system/dict/code/:code [GET]
// @Param		code	path	string	true	"code"
func (receiver SysDictRouter) SysDictCodeDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	code := context.GetPathParam("code")
	if code == "" {
		context.CheckError(core.NewErrCode(core.PARAM_VALIDATE_ERROR))
	}
	//缓存有的话就直接返回
	if exists := receiver.SysDictRedisCache.XHExists(code); exists {
		return context.Success(receiver.SysDictRedisCache.XHGet(code))
	}
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindOneVo(func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", code).Where("status = ?", _const.DictEnableStatusOK)
		})
	context.CheckError(err)
	err, list := receiver.SysDictChildService.WithContext(c).SkipGlobalHook().
		FindVoList(func(db *gorm.DB) *gorm.DB {
			return db.Where("dict_code = ?", code)
		})
	context.CheckError(err)
	x.Children = list
	receiver.SysDictRedisCache.XHSet(code, x)
	return context.Success(x)
}

// SysDictDetail
// @Summary	[系统]字典详情
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
// @Router		/system/dict/:id [GET]
// @Param		id	path	int	true	"id"
func (receiver SysDictRouter) SysDictDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	id := context.GetPathParamInt64("id")
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindOneVoByPrimaryKey(id)
	context.CheckError(err)
	return context.Success(x)
}

// SysDictCodeList
// @Summary	[系统]所有字典代码
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=[]vo.SysCodeList}
// @Router		/system/dict/code-list [GET]
func (receiver SysDictRouter) SysDictCodeList(c echo.Context) error {
	context := core.GetContext[any](c)
	err, dictList := receiver.SysDictService.WithContext(c).FindList()
	context.CheckError(err)
	lists := slice.Map(dictList, func(index int, item model.SysDict) vo.SysCodeList {
		return vo.SysCodeList{
			Code: item.Code,
			Name: item.Name,
		}
	})
	return context.Success(lists)
}

// SysDictExist
// @Summary	[系统]检测字典代码
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/dict/code-exist [GET]
// @Param		code	query	int	true	"code"
func (receiver SysDictRouter) SysDictExist(c echo.Context) error {
	context := core.GetContext[any](c)
	code := context.QueryParam("code")
	count := receiver.SysDictService.WithContext(c).Count(func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	})
	return context.Success(count > 0)
}

// SysDictUpdate
// @Summary	[系统]字典更新
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/dict/:id [PUT]
// @Param		id	path	int				true	"id"
// @Param		bo	body	bo.SysDictBo	true	"修改参数"
func (receiver SysDictRouter) SysDictUpdate(c echo.Context) error {
	context := core.GetContext[bo.SysDictBo](c)
	updateBo := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	receiver.SysDictRedisCache.XHDel(updateBo.Code)
	err, x := receiver.SysDictService.WithContext(c).SaveByPrimaryKey(id, core.CopyFrom[model.SysDict](updateBo))
	context.CheckError(err)
	return context.Success(x)
}

// SysDictAdd
// @Summary	[系统]字典新增
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
// @Router		/system/dict [POST]
// @Param		bo	body	bo.SysDictBo	true	"新增参数"
func (receiver SysDictRouter) SysDictAdd(c echo.Context) error {
	context := core.GetContext[bo.SysDictBo](c)
	addBo := context.GetBodyAndValid()
	err, meta := receiver.SysDictService.WithContext(c).InsertOne(core.CopyFrom[model.SysDict](addBo))
	context.CheckError(err)
	return context.Success(core.CopyFrom[vo.SysDictVo](meta))
}

// SysDictDelete
// @Summary	[系统]字典删除
// @Tags		[系统]字典模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/dict [DELETE]
// @Param		param	query	core.QueryIds	true	"删除参数"
func (receiver SysDictRouter) SysDictDelete(c echo.Context) error {
	context := core.GetContext[any](c)
	ids := context.QueryParamIds()
	_, deleteRows := receiver.SysDictService.WithContext(c).FindVoList(func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", ids)
	})
	slice.ForEach(deleteRows, func(index int, item vo.SysDictVo) {
		receiver.SysDictRedisCache.XHDel(item.Code)
	})

	err, row := receiver.SysDictService.WithContext(c).DeleteByPrimaryKeys(ids)
	context.CheckError(err)
	return context.Success(row)
}

// SysDictChildList
//
//	@Summary	[系统]字典内容列表
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysDictChildVo]}
//	@Router		/system/dict/child/list [GET]
//	@Param		bo	query	bo.SysDictChildPageBo	true	"分页参数"
func (receiver SysDictRouter) SysDictChildList(c echo.Context) error {
	context := core.GetContext[bo.SysDictChildPageBo](c)
	pageBo := context.GetQueryParamAndValid()
	err, x := receiver.SysDictChildService.WithContext(c).FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
		return db.Where("dict_code = ?", pageBo.DictCode)
	})
	context.CheckError(err)
	return context.Success(x)
}

// SysDictChildUpdate
//
//	@Summary	[系统]字典内容更新
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=bool}
//	@Router		/system/dict/child/:code [PUT]
//	@Param		id	path	int					true	"id"
//	@Param		bo	body	[]bo.SysDictChildBo	true	"修改参数"
func (receiver SysDictRouter) SysDictChildUpdate(c echo.Context) error {
	context := core.GetContext[[]bo.SysDictChildBo](c)
	updateBo := context.GetBodyAndValid()
	code := context.GetPathParam("code")
	modelList := core.CopyListFrom[model.SysDictChild](updateBo)
	receiver.SysDictRedisCache.XHDel(code)
	newCodes := slice.Map(modelList, func(index int, item model.SysDictChild) int64 {
		return item.ID
	})

	_, _ = receiver.SysDictChildService.WithContext(c).DeleteBy(func(db *gorm.DB) *gorm.DB {
		return db.Where("id NOT IN (?)", slice.Unique(newCodes)).Where("dict_code = ?", code)
	})

	slice.ForEach(modelList, func(index int, item model.SysDictChild) {
		if item.ID == 0 {
			item.DictCode = code
			err, _ := receiver.SysDictChildService.WithContext(c).InsertOne(item)
			context.CheckError(err)
		} else {
			item.DictCode = code
			err, _ := receiver.SysDictChildService.WithContext(c).UpdateByPrimaryKey(item.ID, item)
			context.CheckError(err)
		}
	})
	return context.Success(true)
}
