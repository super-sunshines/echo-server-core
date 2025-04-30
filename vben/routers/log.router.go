package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/bo"
	_ "github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"gorm.io/gorm"
)

var SysLogRouterGroup = core.NewRouterGroup("/system/log", NewLogRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *LogRouter) {
		rg.GET("/list", m.operateLog, core.HavePermission("SYS::LOG::QUERY"))
		rg.GET("/login/list", m.loginLog, core.HavePermission("SYS::LOG::QUERY"))

	})
})

type LogRouter struct {
	operateLogService services.SysLogService
	loginLogService   services.SysLoginInfoService
}

func NewLogRouter() *LogRouter {
	return &LogRouter{
		operateLogService: services.NewSysLogService(),
		loginLogService:   services.NewSysLoginInfoService(),
	}
}

// @Summary	操作日志列表
// @Tags		[系统]日志模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[model.SysOperateLog]}
// @Router		/system/log/list [GET]
// @Param		bo	query	bo.SysLogOperatePageBo	true	"请求参数"
func (r LogRouter) operateLog(ec echo.Context) (err error) {
	context := core.GetContext[bo.SysLogOperatePageBo](ec)
	queryParam, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, list := r.operateLogService.WithContext(ec).SkipGlobalHook().
		FindVoListByPage(queryParam.PageParam, func(db *gorm.DB) *gorm.DB {
			return db.Order("operate_time desc")
		})
	if err != nil {
		return err
	}
	return context.Success(list)
}

// @Summary	登录日志列表
// @Tags		[系统]日志模块
// @Success	200	{object}	core.ResponseSuccess{data=core.PageResultList[model.SysLoginInfo]}
// @Router		/system/log/login/list [GET]
// @Param		bo	query	bo.SysLogLoginPageBo	true	"请求参数"
func (r LogRouter) loginLog(ec echo.Context) (err error) {
	context := core.GetContext[bo.SysLogLoginPageBo](ec)
	queryParam, err := context.GetQueryParamAndValid()
	if err != nil {
		return err
	}
	err, list := r.loginLogService.WithContext(ec).SkipGlobalHook().
		FindVoListByPage(queryParam.PageParam, func(db *gorm.DB) *gorm.DB {
			return db.Order("operate_time desc")
		})
	if err != nil {
		return err
	}
	return context.Success(list)
}
