package main

import "github.com/XiaoSGentle/echo-server-core/core"

func main() {
	groups := make([]*core.RouterGroup, 0)
	core.NewServer(groups, core.ServerRunOption{})
}
