package hooks

import (
	context2 "context"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// 这里设计的权限是逐级递增的 再往难的去做也不是很会了
// 一个角色有多个权限的时候 以权限码最大的为准
var (
	PersonalDataOnly = int64(1)
	DepartmentBelow  = int64(2)
	AllData          = int64(3)
)

func queryStrategy(roles []model.SysRole, db *gorm.DB, context *core.XContext[any], departService services.SysDepartmentService) {
	user, err := context.GetLoginUser()
	if err != nil {
		return
	}
	var topQuery = int64(0)
	for _, role := range roles {
		if role.QueryStrategy >= topQuery {
			topQuery = role.QueryStrategy
		}
	}
	switch topQuery {
	case PersonalDataOnly:
		db.Where("create_by = ?", user.UID)
		break
	case DepartmentBelow:
		childrenIds, _ := departService.GetChildren(context, user.DepartmentId)
		db.Where("create_dept IN (?)", childrenIds).Or("create_by = ?", user.UID)
		break
	}
}
func updateStrategy(roles []model.SysRole, db *gorm.DB, context *core.XContext[any], departService services.SysDepartmentService) {
	user, err := context.GetLoginUser()
	if err != nil {
		return
	}
	var topQuery = int64(0)
	for _, role := range roles {
		if role.UpdateStrategy >= topQuery {
			topQuery = role.QueryStrategy
		}
	}
	switch topQuery {
	case PersonalDataOnly:
		db.Where("create_by = ?", user.UID)
		break
	case DepartmentBelow:
		childrenIds, _ := departService.GetChildren(context, user.DepartmentId)
		db.Where("create_dept IN (?)", childrenIds).Or("create_by = ?", user.UID)
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
		user, err := context.GetLoginUser()
		if err != nil {
			zap.L().Error("Gorm Global Update Hook Error:", zap.Error(err))
			_ = context.Fail(err)
			return
		}
		updateStrategy(roleService.GetRoleConfigByCodes(context, user.RoleCodes...), _db, context, departService)
		// 使用 Assign 方法确保多列都能设置
		_db.Statement.Assign(map[string]interface{}{
			"update_by": user.UID,
		})
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
			user, err := context.GetLoginUser()
			if err != nil {
				return
			}
			// 使用 Assign 方法确保多列都能设置
			_db.Statement.Assign(map[string]interface{}{
				"create_by":   user.UID,
				"create_dept": user.DepartmentId,
			})
		}
	})

	// 全局查询策略
	_ = globalDb.Callback().Query().Before("gorm:query").Register("custom:BeforeQuery", func(_db *gorm.DB) {
		// 获取GORM绑定的Context
		ctx := _db.Statement.Context
		// 如果是普通的Context
		if bgContext, ok := ctx.(context2.Context); ok {
			if bgContext.Value(core.GormGlobalSkipHookKey) != nil && bgContext.Value(core.GormGlobalSkipHookKey).(bool) {
				return
			}
		}
		// 检测是不是XContext
		context, ok := ctx.(*core.XContext[any])
		if !ok || context == nil {
			return
		}
		// 获取全局跳过的标记
		get := context.Get(core.GormGlobalSkipHookKey)
		// 如果获取到了且是true
		if get != nil && get.(bool) {
			return
		}
		// 最后进入策略
		user, err := context.GetLoginUser()
		if err != nil {
			zap.L().Error("Gorm Global Query Hook Error", zap.Error(err))
			_ = context.Fail(err)
			return
		}
		codes := roleService.GetRoleConfigByCodes(context, user.RoleCodes...)
		queryStrategy(codes, _db, context, departService)
	})

}
