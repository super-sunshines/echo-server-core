package hooks

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	_const "github.com/XiaoSGentle/echo-server-core/vben/const"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/query"
	"github.com/duke-git/lancet/v2/slice"
)

// RolePermissionHook 返回一个预定义的管理员角色权限映射
// 该函数无需参数
// 返回值: core.RoleMap 类型的映射
func RolePermissionHook() (roleMap core.RoleMap, err error) {
	roleMap = make(core.RoleMap)
	roleQuery := query.Use(core.GetGormDB()).SysRole.WithContext(core.NewSkipGormGlobalHookContext())
	menuQuery := query.Use(core.GetGormDB()).SysMenu.WithContext(core.NewSkipGormGlobalHookContext())
	allRole, err := roleQuery.Find()
	allMenu, err := menuQuery.Find()
	if err != nil {
		return
	}
	slice.ForEach(allRole, func(index int, itemRole *model.SysRole) {
		filterMenu := slice.Filter(allMenu, func(index int, itemMenu *model.SysMenu) bool {
			return slice.Contain(itemRole.MenuIDList, itemMenu.ID) && itemMenu.Type != _const.MenuTypeApi
		})
		filterMenuIds := slice.Map(filterMenu, func(index int, item *model.SysMenu) int64 {
			return item.ID
		})

		filterRole := slice.Filter(allMenu, func(index int, itemMenu *model.SysMenu) bool {
			return slice.Contain(itemRole.MenuIDList, itemMenu.ID) && itemMenu.Type == _const.MenuTypeApi
		})
		filterRoleCodes := slice.Map(filterRole, func(index int, item *model.SysMenu) string {
			return item.APICode
		})
		roleMap[itemRole.Code] = struct {
			Codes    []string
			MenuIds  []int64
			HomePath string
		}{Codes: filterRoleCodes, MenuIds: filterMenuIds, HomePath: itemRole.HomePath}

	})
	return
}
