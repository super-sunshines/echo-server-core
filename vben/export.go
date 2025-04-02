package vben

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/routers"
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
