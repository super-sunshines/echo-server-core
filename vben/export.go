package vben

import (
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/routers"
)

var BaseRouters = []*core.RouterGroup{
	routers.AuthRouterGroup,
	routers.MenuRouterGroup,
	routers.SysLogRouterGroup,
	routers.SysDictRouterGroup,
	routers.SysUserRouterGroup,
	routers.SysRoleRouterGroup,
	routers.SysDepartmentRouterGroup,
}

var TencentRouters = []*core.RouterGroup{
	routers.TencentCloudRouterGroup,
}

var TencentQywxRouters = []*core.RouterGroup{
	routers.QywxRouterGroup,
}

func AddPermissionCodes(codes []string) {
	routers.AddRoleCodes(codes)
}
