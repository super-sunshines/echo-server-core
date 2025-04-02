package routers

import (
	"echo-server-core/core"
	"echo-server-core/vben/bo"
	_const "echo-server-core/vben/const"
	"echo-server-core/vben/gorm/model"
	"echo-server-core/vben/vo"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var AllRoleCodes = []string{}

func AddRoleCodes(codes []string) {
	for _, code := range codes {
		AllRoleCodes = slice.AppendIfAbsent(AllRoleCodes, code)
	}
}

var MenuRouterGroup = core.NewRouterGroup("/system/menu", NewMenuRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *MenuRouter) {
		rg.GET("/list", m.list, core.IgnorePermission())
		rg.GET("/simple", m.simpleMenu, core.IgnorePermission())
		rg.GET("/:id", m.detailMenu, core.IgnorePermission())
		rg.PUT("/:id", m.updateMenu, core.IgnorePermission())
		rg.POST("", m.addMenu, core.IgnorePermission())
		rg.DELETE("", m.deleteMenu, core.IgnorePermission())
		rg.GET("/new/codes", m.newCodeList, core.IgnorePermission())
		rg.POST("/code", m.addCode, core.IgnorePermission())
		rg.PUT("/code/:id", m.editCode, core.IgnorePermission())
		rg.GET("/name-exists", m.nameExist, core.IgnorePermission())
		rg.GET("/path-exists", m.pathExist, core.IgnorePermission())
	})
})

type MenuRouter struct {
	MenuService     core.PreGorm[model.SysMenu, vo.SysMenuWithMetaVo]
	MenuMetaService core.PreGorm[model.SysMenuMetum, any]
}

func NewMenuRouter() *MenuRouter {
	return &MenuRouter{
		MenuService:     core.NewService[model.SysMenu, vo.SysMenuWithMetaVo](),
		MenuMetaService: core.NewService[model.SysMenuMetum, any](),
	}
}

// @Summary	获取目录列表
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=[]vo.SysMenuWithMeta}
// @Router		/system/menu/list [GET]
func (r MenuRouter) list(c echo.Context) error {
	context := core.GetContext[any](c)
	var metas []vo.SysMenuWithMeta
	tx := r.MenuService.WithContext(c).SkipGlobalHook().
		Model(vo.SysMenuWithMeta{}).Preload("Meta").Order("`type`").
		Find(&metas)
	context.CheckError(tx.Error)
	userMenuVoList := core.CopyListFrom[vo.SysMenuWithMetaVo](metas)
	return context.Success(vo.BuildTree(userMenuVoList))
}

// @Summary	目录详情
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=vo.SysMenuWithMetaVo}
// @Router		/system/menu/:id [GET]
func (r MenuRouter) detailMenu(c echo.Context) error {
	context := core.GetContext[any](c)
	paramInt64 := context.GetPathParamInt64("id")
	var meta vo.SysMenuWithMeta
	tx := r.MenuService.WithContext(c).SkipGlobalHook().
		Model(vo.SysMenuWithMeta{}).Where("id = ?", paramInt64).Preload("Meta").First(&meta)
	context.CheckError(tx.Error)
	userMenuVo := core.CopyFrom[vo.SysMenuWithMetaVo](meta)
	return context.Success(userMenuVo)
}

// @Summary	名称是否存在
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu/path-exists [GET]
func (r MenuRouter) nameExist(c echo.Context) error {
	context := core.GetContext[any](c)
	return context.Success(false)
}

// @Summary	地址是否存在
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu/path-exists [GET]
func (r MenuRouter) pathExist(c echo.Context) error {
	context := core.GetContext[any](c)
	return context.Success(false)
}

// @Summary	更新系统菜单
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu/:id [PUT]
// @Param		bo	body	bo.UserMenuBo	true	"更新参数"
// @Param		id	path	int				true	"主键"
func (r MenuRouter) updateMenu(c echo.Context) error {
	context := core.GetContext[bo.UserMenuBo](c)
	userMenuBo := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	fromMenu := core.CopyFrom[model.SysMenu](userMenuBo)
	fromMenu.MetaID = userMenuBo.Meta.ID
	err, _ := r.MenuService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(id, fromMenu)
	context.CheckError(err)
	err, _ = r.MenuMetaService.WithContext(c).SkipGlobalHook().
		SaveByPrimaryKey(userMenuBo.Meta.ID, core.CopyFrom[model.SysMenuMetum](userMenuBo.Meta))
	context.CheckError(err)
	return context.Success(true)
}

// @Summary	新增系统菜单
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu [POST]
// @Param		bo	body	bo.UserMenuBo	true	"更新参数"
func (r MenuRouter) addMenu(c echo.Context) error {
	context := core.GetContext[bo.UserMenuBo](c)
	userMenuBo := context.GetBodyAndValid()
	err, meta := r.MenuMetaService.WithContext(c).InsertOne(core.CopyFrom[model.SysMenuMetum](userMenuBo.Meta))
	context.CheckError(err)
	menu := core.CopyFrom[model.SysMenu](userMenuBo)
	menu.MetaID = meta.ID
	err, _ = r.MenuService.WithContext(c).InsertOne(menu)
	context.CheckError(err)
	return context.Success(true)
}

// @Summary	删除系统菜单
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=int}
// @Router		/system/menu [DELETE]
// @Param		bo	body	bo.UserMenuBo	true	"更新参数"
func (r MenuRouter) deleteMenu(c echo.Context) error {
	context := core.GetContext[any](c)
	ids := context.QueryParamIds()
	r.MenuService.WithContext(c).GetModelDb().Where("pid in (?)", ids).Update("pid", 0)
	err, meta := r.MenuService.WithContext(c).DeleteByPrimaryKeys(ids)
	context.CheckError(err)
	return context.Success(meta)
}

// @Summary	简单系统菜单
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=[]vo.SysSimpleMenuVo}
// @Router		/system/menu/simple [GET]
// @Param		bo	query	bo.SimpleTreeBo	true	"更新参数"
func (r MenuRouter) simpleMenu(c echo.Context) error {
	context := core.GetContext[bo.SimpleTreeBo](c)
	params := context.GetQueryParamAndValid()
	var metas []vo.SysMenuWithMeta
	tx := r.MenuService.WithContext(c).GetDb().
		Model(vo.SysMenuWithMeta{}).Preload("Meta").
		Where("type IN (?)", core.BooleanTo(params.IncludePermissions, _const.MenuTreeTypeWithApi, _const.MenuTreeType)).
		Find(&metas)
	context.CheckError(tx.Error)
	sysSimpleMenuVoList := vo.BuildSimpleTree(metas)

	if params.IncludeTopLevel {
		sysSimpleMenuVoList = append(sysSimpleMenuVoList, &vo.SysSimpleMenuVo{
			Name: "顶级目录",
			ID:   0,
		})
	}
	return context.Success(sysSimpleMenuVoList)
}

// @Summary	未录入的权限码
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=[]string}
// @Router		/system/menu/new/codes [GET]
func (r MenuRouter) newCodeList(c echo.Context) error {
	context := core.GetContext[any](c)
	err, notInList := r.MenuService.WithContext(c).FindList(func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", _const.MenuTypeApi)
	})
	context.CheckError(err)
	exitCodes := slice.Map(notInList, func(_ int, item model.SysMenu) string {
		return item.APICode
	})
	return context.Success(slice.Unique(slice.Difference(AllRoleCodes, exitCodes)))
}

// @Summary	录入权限码
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu/code [POST]
// @Param		bo	body	bo.AddCodeListBo	true	"新增参数"
func (r MenuRouter) addCode(c echo.Context) error {
	context := core.GetContext[bo.AddCodeListBo](c)
	body := context.GetBodyAndValid()
	menus := slice.Map(body.List, func(index int, item bo.ApiCodeBo) model.SysMenu {
		return model.SysMenu{
			Pid:            item.Pid,
			APICode:        item.Code,
			APIDescription: item.Description,
			Type:           _const.MenuTypeApi,
			MetaID:         0,
		}
	})
	err, _ := r.MenuService.WithContext(c).InsertBatch(menus)
	context.CheckError(err)
	return context.Success(true)
}

// @Summary 修改权限码
// @Tags		[系统]目录模块
// @Success	200	{object}	core.ResponseSuccess{data=bool}
// @Router		/system/menu/code/:id [PUT]
// @Param		bo	body	bo.ApiCodeBo	true	"新增参数"
func (r MenuRouter) editCode(c echo.Context) error {
	context := core.GetContext[bo.ApiCodeBo](c)
	body := context.GetBodyAndValid()
	id := context.GetPathParamInt64("id")
	err2, menuItem := r.MenuService.WithContext(c).FindOneByPrimaryKey(id)
	context.CheckError(err2)
	menuItem.APICode = body.Code
	menuItem.APIDescription = body.Description
	menuItem.Pid = body.Pid
	err, _ := r.MenuService.WithContext(c).SaveByPrimaryKey(id, menuItem)
	context.CheckError(err)
	return context.Success(true)
}
