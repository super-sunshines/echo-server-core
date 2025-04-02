package main

import (
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben"
)

func main() {
	groups := make([]*core.RouterGroup, 0)
	groups = append(groups, vben.Routers...)
	core.NewServer(groups, core.ServerRunOption{})
}
