package core

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//const (
//	PermissionOneOf PermissionType = "CheckPermissionOneOf"
//	PermissionAll   PermissionType = "CheckPermissionAll"
//)
//
//type PermissionType string
//// CustomPermissions 自定义权限 permissionType core.PermissionOneOf 或者 core.PermissionAll
//func CustomPermissions(roles []string, permissionType ...PermissionType) echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			selectPermission := PermissionOneOf
//			if len(permissionType) != 0 {
//				selectPermission = permissionType[0]
//			}
//			fmt.Printf("%s => %v \n", selectPermission, roles)
//			return next(c)
//		}
//	}
//}

func HaveOneOfPermissions(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := GetContext[any](c)
			user, err := context.GetLoginUser()
			if err != nil || !PermissionMange.CheckRoleHaveCodePermission(user.RoleCodes, roles, false) {
				context.Error(NewFrontShowErrMsg(fmt.Sprintf("Dont Have OneOf Permission Check Error: %#v", roles)))
			} else {
				return next(c)
			}
			return err
		}
	}
}
func HaveAllPermissions(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := GetContext[any](c)
			user, err := context.GetLoginUser()
			if err != nil || !PermissionMange.CheckRoleHaveCodePermission(user.RoleCodes, roles, true) {
				context.Error(NewFrontShowErrMsg(fmt.Sprintf("Dont All Permission Check Error: %#v", roles)))
			} else {
				return next(c)
			}
			return err
		}
	}
}

func HavePermission(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := GetContext[any](c)
			user, err := context.GetLoginUser()
			if err != nil || !PermissionMange.CheckRoleHaveCodePermission(user.RoleCodes, []string{role}, false) {
				context.Error(NewFrontShowErrMsg(fmt.Sprintf("Dont Have Permission Check Error: %#v", role)))
				return NewFrontShowErrMsg(fmt.Sprintf("Dont Have Permission Check Error: %#v", role))
			}
			return next(c)
		}
	}
}
func IgnorePermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(GormGlobalSkipHookKey, true)
			return next(c)
		}
	}
}

var PermissionMange *RolePermission

type RolePermission struct {
	RoleCodeRedis        *RedisCache[[]string]
	RoleMenuIdRedis      *RedisCache[[]int64]
	RoleHomeRedis        *RedisCache[string]
	getRolePermissionsFn PermissionsOptions
}
type RoleMap map[string]struct {
	Codes    []string
	MenuIds  []int64
	HomePath string
}
type PermissionsOptions func() (RoleMap, error)

func initRolePermission(getRolePermissionsFn PermissionsOptions) {
	PermissionMange = &RolePermission{
		RoleCodeRedis:        GetRedisCache[[]string]("role-code-cache-key"),
		RoleMenuIdRedis:      GetRedisCache[[]int64]("role-menu-cache-key"),
		RoleHomeRedis:        GetRedisCache[string]("role-home-path-cache-key"),
		getRolePermissionsFn: getRolePermissionsFn,
	}
	err := PermissionMange.init()
	if err != nil {
		zap.L().Error("init role permission failed", zap.Error(err))
	}
}

// init 初始化角色权限信息到缓存中。
// 该方法首先调用 getRolePermissionsFn 获取角色权限数据，
// 然后遍历每个角色，将其对应的菜单ID和权限代码转换为JSON字符串，
// 并分别存储到角色菜单缓存、角色代码缓存中。
// 同时，角色的主页路径也被存储到角色主页路径缓存中。
// 这样做是为了在后续查询角色权限时能够快速从缓存中获取，提高性能。
func (r *RolePermission) init() error {
	// 获取角色权限数据。
	rolePermissions, err := r.getRolePermissionsFn()
	r.RoleCodeRedis.XDel()
	r.RoleCodeRedis.XDel()
	r.RoleCodeRedis.XDel()
	if err != nil {
		return err
	}
	// 遍历角色权限数据。
	for key, roles := range rolePermissions {
		// 将角色的菜单ID列表转换为JSON字符串并存储到缓存中。
		r.RoleMenuIdRedis.XHSet(key, roles.MenuIds)
		// 将角色的权限代码列表转换为JSON字符串并存储到缓存中。
		r.RoleCodeRedis.XHSet(key, roles.Codes)
		// 将角色的主页路径存储到缓存中。
		r.RoleHomeRedis.XHSet(key, roles.HomePath)
	}
	return nil
}

// Refresh 用于刷新角色权限信息。
// 该方法通过调用 init 函数来重新初始化角色权限相关数据，
// 以确保角色权限信息是最新的。
func (r *RolePermission) Refresh() error {
	return r.init()
}

func (r *RolePermission) CheckRoleHaveCodePermission(roles []string, codes []string, all bool) bool {
	// 初始化一个空的字符串切片，用于存储从缓存中获取的权限码
	if len(roles) == 0 {
		return false
	}
	if len(codes) == 0 {
		return true
	}
	var cacheCodes []string
	listStrList := r.RoleCodeRedis.XHMGet(roles...)
	slice.ForEach(listStrList, func(index int, listStr []string) {
		if listStr != nil {
			cacheCodes = append(cacheCodes, listStr...)
		}
	})
	intersection := slice.Intersection(cacheCodes, codes)
	return BooleanTo(all, len(intersection) == len(codes), len(intersection) > 0)
}

// GetRoleMenuIdList 获取角色对应的菜单ID列表
// 参数:
//
//	role - 角色代码，用于查询对应的菜单ID列表
//
// 返回值:
//
//	[]int64 - 角色对应的菜单ID列表，以切片形式返回
func (r *RolePermission) GetRoleMenuIdList(roles ...string) []int64 {
	// 初始化一个空的菜单ID切片，用于存储从缓存中获取的菜单ID
	var cacheMenuIds []int64
	listStrList := r.RoleMenuIdRedis.XHMGet(roles...)
	slice.ForEach(listStrList, func(index int, listStr []int64) {
		if listStr != nil {
			cacheMenuIds = append(cacheMenuIds, listStr...)
		}
	})
	// 返回解析后的菜单ID列表
	return slice.Unique(cacheMenuIds)
}

// GetRoleHomePath 获取角色的主页路径
// 该方法通过查询Redis缓存来获取指定角色的主页路径
// 参数:
//
//	role - 角色名称，用于查询主页路径
//
// 返回值:
//
//	string - 角色的主页路径，如果未找到则返回空字符串
func (r *RolePermission) GetRoleHomePath(role string) string {
	return r.RoleHomeRedis.XHGet(role)
}
