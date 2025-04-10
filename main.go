package main

import (
	"github.com/super-sunshines/echo-server-core/core"
	_ "github.com/super-sunshines/echo-server-core/docs"
	"github.com/super-sunshines/echo-server-core/vben"
	"github.com/super-sunshines/echo-server-core/vben/hooks"
)

func main() {
	groups := make([]*core.RouterGroup, 0)
	groups = append(groups, vben.BaseRouters...)
	groups = append(groups, vben.TencentRouters...)
	groups = append(groups, vben.TencentQywxRouters...)
	core.NewServer(groups, core.ServerRunOption{
		GormOptions:        core.InitGormOptions{GormGlobalHook: hooks.GlobalGormHook},
		PermissionsOptions: hooks.RolePermissionHook,
		LoggerOptions: core.LoggerOptions{
			LoggerSaver: hooks.LoggerMiddlewareHook,
		},
	})
}
