package routers

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"gorm.io/gorm"
)

var SysDictRouterGroup = core.NewRouterGroup("/system/dict", NewSysDictRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysDictRouter) {
		rg.GET("/list", m.SysDictList, core.HavePermission("SYS::DICT::QUERY"), core.Log("字典列表"))
		rg.GET("/code-list", m.SysDictCodeList, core.HavePermission("SYS::DICT::QUERY"), core.Log("所有字典代码"))
		rg.GET("/code-exist", m.SysDictExist, core.HavePermission("SYS::DICT::QUERY"), core.Log("代码存在"))
		rg.GET("/:id", m.SysDictDetail, core.HavePermission("SYS::DICT::QUERY"), core.Log("代码内容"))
		rg.GET("/code/:code", m.SysDictCodeDetail, core.IgnorePermission())
		rg.PUT("/:id", m.SysDictUpdate, core.HavePermission("SYS::DICT::UPDATE"), core.Log("字典修改"))
		rg.POST("", m.SysDictAdd, core.HavePermission("SYS::DICT::ADD"), core.Log("字典新增"))
		rg.DELETE("", m.SysDictDelete, core.HavePermission("SYS::DICT::DEL"), core.Log("字典删除"))
		rg.GET("/child/list", m.SysDictChildList, core.HavePermission("SYS::DICT::CHILD::QUERY"), core.Log("字典内容删除"))
		rg.PUT("/child/:code", m.SysDictChildUpdate, core.HavePermission("SYS::DICT::CHILD::UPDATE"), core.Log("字典内容修改"))
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
//
//	@Summary	[系统]字典列表
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[vo.SysDictVo]}
//	@Router		/system/dict/list [GET]
//	@Param		bo	query	bo.SysDictPageBo	true	"分页参数"
func (receiver SysDictRouter) SysDictList(c echo.Context) error {
	context := core.GetContext[bo.SysDictPageBo](c)
	pageBo, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
			core.BooleanFun(pageBo.Module != 0, func() {
				db.Where("module = ?", pageBo.Module)
			})
			return db
		})
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysDictCodeDetail
//
//	@Summary	[系统]字典详情[code]
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
//	@Router		/system/dict/code/:code [GET]
//	@Param		code	path	string	true	"code"
func (receiver SysDictRouter) SysDictCodeDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	code := context.GetPathParam("code")
	if code == "" {
		return core.NewErrCode(core.PARAM_VALIDATE_ERROR)
	}
	//缓存有的话就直接返回
	if exists := receiver.SysDictRedisCache.XHExists(code); exists {
		return context.Success(receiver.SysDictRedisCache.XHGet(code))
	}
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindOneVo(func(db *gorm.DB) *gorm.DB {
			return db.Where("code = ?", code).Where("status = ?", _const.DictEnableStatusOK)
		})
	if err != nil {
		return err
	}
	err, list := receiver.SysDictChildService.WithContext(c).SkipGlobalHook().
		FindVoList(func(db *gorm.DB) *gorm.DB {
			return db.Where("dict_code = ?", code)
		})
	if err != nil {
		return err
	}
	x.Children = list
	receiver.SysDictRedisCache.XHSet(code, x)
	return context.Success(x)
}

// SysDictDetail
//
//	@Summary	[系统]字典详情
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
//	@Router		/system/dict/:id [GET]
//	@Param		id	path	int	true	"id"
func (receiver SysDictRouter) SysDictDetail(c echo.Context) error {
	context := core.GetContext[any](c)
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	err, x := receiver.SysDictService.WithContext(c).SkipGlobalHook().
		FindOneVoByPrimaryKey(id)
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysDictCodeList
//
//	@Summary	[系统]所有字典代码
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=[]vo.SysCodeList}
//	@Router		/system/dict/code-list [GET]
func (receiver SysDictRouter) SysDictCodeList(c echo.Context) error {
	context := core.GetContext[any](c)
	err, dictList := receiver.SysDictService.WithContext(c).FindList()
	if err != nil {
		return err
	}
	lists := slice.Map(dictList, func(index int, item model.SysDict) vo.SysCodeList {
		return vo.SysCodeList{
			Code: item.Code,
			Name: item.Name,
		}
	})
	return context.Success(lists)
}

// SysDictExist
//
//	@Summary	[系统]检测字典代码
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=bool}
//	@Router		/system/dict/code-exist [GET]
//	@Param		code	query	int	true	"code"
func (receiver SysDictRouter) SysDictExist(c echo.Context) error {
	context := core.GetContext[any](c)
	code := context.QueryParam("code")
	count := receiver.SysDictService.WithContext(c).Count(func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	})
	return context.Success(count > 0)
}

// SysDictUpdate
//
//	@Summary	[系统]字典更新
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/dict/:id [PUT]
//	@Param		id	path	int				true	"id"
//	@Param		bo	body	bo.SysDictBo	true	"修改参数"
func (receiver SysDictRouter) SysDictUpdate(c echo.Context) error {
	context := core.GetContext[bo.SysDictBo](c)
	updateBo, err := context.GetBodyAndValid()
	id, err := context.GetPathParamInt64("id")
	if err != nil {
		return err
	}
	receiver.SysDictRedisCache.XHDel(updateBo.Code)
	err, x := receiver.SysDictService.WithContext(c).SaveByPrimaryKey(id, core.CopyFrom[model.SysDict](updateBo))
	if err != nil {
		return err
	}
	return context.Success(x)
}

// SysDictAdd
//
//	@Summary	[系统]字典新增
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=vo.SysDictVo}
//	@Router		/system/dict [POST]
//	@Param		bo	body	bo.SysDictBo	true	"新增参数"
func (receiver SysDictRouter) SysDictAdd(c echo.Context) error {
	context := core.GetContext[bo.SysDictBo](c)
	addBo, err := context.GetBodyAndValid()
	if err != nil {
		return err
	}
	err, meta := receiver.SysDictService.WithContext(c).InsertOne(core.CopyFrom[model.SysDict](addBo))
	if err != nil {
		return err
	}
	return context.Success(core.CopyFrom[vo.SysDictVo](meta))
}

// SysDictDelete
//
//	@Summary	[系统]字典删除
//	@Tags		[系统]字典模块
//	@Success	200	{object}	core.ResponseSuccess{data=int}
//	@Router		/system/dict [DELETE]
//	@Param		param	query	core.QueryIds	true	"删除参数"
func (receiver SysDictRouter) SysDictDelete(c echo.Context) error {
	context := core.GetContext[any](c)
	ids, err := context.QueryParamIds()
	if err != nil {
		return err
	}
	_, deleteRows := receiver.SysDictService.WithContext(c).FindVoList(func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", ids)
	})
	slice.ForEach(deleteRows, func(index int, item vo.SysDictVo) {
		receiver.SysDictRedisCache.XHDel(item.Code)
	})

	err, row := receiver.SysDictService.WithContext(c).DeleteByPrimaryKeys(ids)
	if err != nil {
		return err
	}
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
	pageBo, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, x := receiver.SysDictChildService.WithContext(c).SkipGlobalHook().FindVoListByPage(pageBo.PageParam, func(db *gorm.DB) *gorm.DB {
		return db.Where("dict_code = ?", pageBo.DictCode)
	})
	if err != nil {
		return err
	}
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
	updateBo, err := context.GetBodyAndValid()
	code := context.GetPathParam("code")
	if err != nil {
		return err
	}
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
			if err != nil {
				c.Error(err)
				return
			}
		} else {
			item.DictCode = code
			err, _ := receiver.SysDictChildService.WithContext(c).UpdateByPrimaryKey(item.ID, item)
			if err != nil {
				c.Error(err)
				return
			}

		}
	})
	return context.Success(true)
}
