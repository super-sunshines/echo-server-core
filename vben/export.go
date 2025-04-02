package vben

import (
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/routers"
)

var Routers = []*core.RouterGroup{
	routers.AuthRouterGroup,
	routers.MenuRouterGroup,
	routers.QywxRouterGroup,
	routers.TencentCloudRouterGroup,
	routers.SysDictRouterGroup,
	routers.SysUserRouterGroup,
	routers.SysRoleRouterGroup,
	routers.SysDepartmentRouterGroup,
}

func AddPermissionCodes(codes []string) {
	routers.AddRoleCodes(codes)
}
