package hooks

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/gorm/query"
	"github.com/super-sunshines/echo-server-core/vben/services"
)

func LoggerMiddlewareHook(info core.RequestInfo, c echo.Context) {
	go func() {
		logQuery := query.Use(core.GetGormDB()).SysOperateLog.WithContext(core.NewSkipGormGlobalHookContext())
		from := core.CopyFrom[model.SysOperateLog](info)
		context := core.GetContext[any](c)
		if context.IsLogin() {
			user, _ := context.GetLoginUser()
			from.OperateName = user.NickName
			from.OperateDepart = services.NewDepartmentService().GetUserDepartment(c).Name
			from.OperateUserID = user.UID
		}
		err := logQuery.Save(&from)
		core.BooleanFun(err != nil, func() {
			fmt.Printf("%#v", err)
		})
	}()
	return
}
