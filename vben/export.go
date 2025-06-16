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
	routers.SysFileRouterGroup,
}

var TencentRouters = []*core.RouterGroup{
	routers.TencentCloudRouterGroup,
}

var TencentWorkWechatRouters = []*core.RouterGroup{
	routers.WorkWechatRouterGroup,
}

var TencentWechatAppRouters = []*core.RouterGroup{
	routers.WechatAppRouterGroup,
}

func AddPermissionCodes(codes []string) {
	routers.AddRoleCodes(codes)
}
