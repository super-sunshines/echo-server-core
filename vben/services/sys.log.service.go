package services

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"time"
)

type SysLogService struct {
	core.PreGorm[model.SysOperateLog, model.SysOperateLog]
	departmentService SysDepartmentService
}

func NewSysLogService() SysLogService {
	return SysLogService{
		PreGorm:           core.NewService[model.SysOperateLog, model.SysOperateLog](),
		departmentService: NewDepartmentService(),
	}
}

func (r SysLogService) AddLog(c echo.Context, log model.SysOperateLog) {
	go func() {
		from := core.CopyFrom[model.SysOperateLog](log)
		context := core.GetContext[any](c)
		if context.IsLogin() {
			user, _ := context.GetLoginUser()
			from.OperateName = user.NickName
			from.OperateDepart = r.departmentService.GetUserDepartment(c).Name
			from.OperateUserID = user.UID
		}
		// 重写部分内容
		from.CallFunc = core.PathFuncStrMap[c.Path()]
		from.RequestMethod = c.Request().Method
		from.OperateURL = c.Path()
		from.OperateIP = c.RealIP()
		from.OperateLocation, _ = core.IPParse(c.RealIP())
		from.OperateParam = c.QueryParams().Encode()
		from.OperateTime = core.NewTime(time.Now())
		_, _ = r.PreGorm.SetDB(core.GetGormDB()).SkipGlobalHook().InsertOne(from)

	}()
}
func (r SysLogService) AddLogSimple(c echo.Context, title string, content string) {
	go func() {
		var from = model.SysOperateLog{
			Title:       title,
			JSONResult:  content,
			OperateType: 1,
		}
		context := core.GetContext[any](c)
		if context.IsLogin() {
			user, _ := context.GetLoginUser()
			from.OperateName = user.NickName
			from.OperateDepart = r.departmentService.GetUserDepartment(c).Name
			from.OperateUserID = user.UID
		}
		// 重写部分内容
		from.CallFunc = core.PathFuncStrMap[c.Path()]
		from.RequestMethod = c.Request().Method
		from.OperateURL = c.Path()
		from.OperateIP = c.RealIP()
		from.OperateLocation, _ = core.IPParse(c.RealIP())
		from.OperateParam = c.QueryParams().Encode()
		from.OperateTime = core.NewTime(time.Now())
		_, _ = r.PreGorm.SetDB(core.GetGormDB()).SkipGlobalHook().InsertOne(from)

	}()
}
