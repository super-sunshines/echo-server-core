package services

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/vo"
)

type SysRoleService struct {
	core.PreGorm[model.SysRole, vo.SysRoleVo]
	roleCache *core.RedisCache[model.SysRole]
}

func NewSysRoleService() SysRoleService {
	return SysRoleService{
		PreGorm:   core.NewService[model.SysRole, vo.SysRoleVo](),
		roleCache: core.GetRedisCache[model.SysRole]("sys-role-permission-cache"),
	}
}
func (r SysRoleService) RefreshCache() {
	r.roleCache.XDel()
}

func (r SysRoleService) GetAllRole(c echo.Context) map[string]model.SysRole {
	all := r.roleCache.XHGetAll()
	if len(all) == 0 {
		var roles []model.SysRole
		r.WithContext(c).Set(core.GormGlobalSkipHookKey, true).Find(&roles)
		for _, role := range roles {
			r.roleCache.XHSet(role.Code, role)
		}
	}
	return r.roleCache.XHGetAll()
}

func (r SysRoleService) GetRoleConfigByCodes(c echo.Context, codes ...string) []model.SysRole {
	var resultRoles []model.SysRole
	if len(codes) == 0 {
		return resultRoles
	}
	rolesMap := r.GetAllRole(c)
	for _, code := range codes {
		resultRoles = append(resultRoles, rolesMap[code])
	}
	return resultRoles
}
