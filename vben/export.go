package vben

import (
	"echo-server-core/core"
	"echo-server-core/vben/routers"
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
