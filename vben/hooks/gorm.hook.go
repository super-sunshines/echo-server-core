package hooks

import (
	context2 "context"
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/services"

	"gorm.io/gorm"
)

// 这里设计的权限是逐级递增的 再往难的去做也不是很会了
// 一个角色有多个权限的时候 以权限码最大的为准
var (
	PersonalDataOnly = int64(1)
	DepartmentBelow  = int64(2)
	AllData          = int64(3)
)

type DBExecutionStrategy struct {
	Name string `json:"name"` // 策略名
	Code int64  `json:"code"` // 策略标记
}

var DataStrategyMap = map[int64]DBExecutionStrategy{
	PersonalDataOnly: {
		Name: "仅本人数据",
		Code: PersonalDataOnly,
	},
	DepartmentBelow: {
		Name: "部门及以下",
		Code: DepartmentBelow,
	},
	AllData: {
		Name: "所有数据",
		Code: AllData,
	},
}

func queryStrategy(roles []model.SysRole, db *gorm.DB, context *core.XContext[any], departService services.SysDepartmentService) {
	var topQuery = int64(0)
	for _, role := range roles {
		if role.QueryStrategy >= topQuery {
			topQuery = role.QueryStrategy
		}
	}

	switch topQuery {
	case PersonalDataOnly:
		db.Where("create_by = ?", context.GetLoginUserUid())
		break
	case DepartmentBelow:
		childrenIds, _ := departService.GetChildren(context, context.GetLoginUserDepartmentId())
		db.Where("create_dept IN (?)", childrenIds).Or("create_by = ?", context.GetLoginUserUid())
		break
	}
}
func updateStrategy(roles []model.SysRole, db *gorm.DB, context *core.XContext[any], departService services.SysDepartmentService) {
	var topQuery = int64(0)
	for _, role := range roles {
		if role.UpdateStrategy >= topQuery {
			topQuery = role.QueryStrategy
		}
	}
	switch topQuery {
	case PersonalDataOnly:
		db.Where("create_by = ?", context.GetLoginUserUid())
		break
	case DepartmentBelow:
		childrenIds, _ := departService.GetChildren(context, context.GetLoginUserDepartmentId())
		db.Where("create_dept IN (?)", childrenIds).Or("create_by = ?", context.GetLoginUserUid())
		break
	}
}

func GlobalGormHook(globalDb *gorm.DB) {
	roleService := services.NewSysRoleService()
	departService := services.NewDepartmentService()
	_ = globalDb.Callback().Update().Before("gorm:update").Register("custom:BeforeUpdate", func(_db *gorm.DB) {
		ctx := _db.Statement.Context
		if bgContext, ok := ctx.(context2.Context); ok {
			if bgContext.Value(core.GormGlobalSkipHookKey) != nil && bgContext.Value(core.GormGlobalSkipHookKey).(bool) {
				return
			}
		}
		context, ok := ctx.(*core.XContext[any])
		if !ok || context == nil {
			return
		}
		if get := context.Get(core.GormGlobalSkipHookKey); get != nil && get.(bool) {
			return
		}
		codes := context.GetLoginUser().RoleCodes
		updateStrategy(roleService.GetRoleConfigByCodes(context, codes...), _db, context, departService)
		_db.Statement.SetColumn("update_by", context.GetLoginUserUid())
	})

	_ = globalDb.Callback().Create().Before("gorm:create").Register("custom:BeforeCreate", func(_db *gorm.DB) {
		ctx := _db.Statement.Context

		if bgContext, ok := ctx.(context2.Context); ok {
			if bgContext.Value(core.GormGlobalSkipHookKey) != nil && bgContext.Value(core.GormGlobalSkipHookKey).(bool) {
				return
			}
		}

		context, ok := ctx.(*core.XContext[any])
		if !ok || context == nil {
			return
		}

		if get := context.Get(core.GormGlobalSkipHookKey); get != nil && get.(bool) {
			return
		}
		if ok {
			_db.Statement.SetColumn("create_by", context.GetLoginUserUid())
			_db.Statement.SetColumn("create_dept", context.GetLoginUserDepartmentId())
		}
	})

	_ = globalDb.Callback().Query().Before("gorm:query").Register("custom:BeforeQuery", func(_db *gorm.DB) {
		ctx := _db.Statement.Context
		if bgContext, ok := ctx.(context2.Context); ok {
			if bgContext.Value(core.GormGlobalSkipHookKey) != nil && bgContext.Value(core.GormGlobalSkipHookKey).(bool) {
				return
			}
		}
		context, ok := ctx.(*core.XContext[any])
		if !ok || context == nil {
			return
		}
		if get := context.Get(core.GormGlobalSkipHookKey); get != nil && get.(bool) {
			return
		}
		codes := context.GetLoginUser().RoleCodes
		queryStrategy(roleService.GetRoleConfigByCodes(context, codes...), _db, context, departService)
	})

}
