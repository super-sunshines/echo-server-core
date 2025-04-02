package core

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"
	"log"
)

var container *DigContainer

type DigContainer struct {
	Container *dig.Container
}

func NewDigContainer() *DigContainer {
	container = &DigContainer{
		Container: dig.New(),
	}
	return container
}

func GetDig() *DigContainer {
	if container == nil {
		container = NewDigContainer()
	}
	return container
}

func (r *DigContainer) Provide(constructor interface{}, opts ...dig.ProvideOption) {
	_ = GetDig().Container.Provide(constructor, opts...)

}
func (r *DigContainer) DI(function interface{}, opts ...dig.InvokeOption) error {
	return GetDig().Container.Invoke(function, opts...)
}

type RouterGroup struct {
	path        string
	initHandle  any
	regHandle   func(g *echo.Group, group *RouterGroup) error
	middleWares []echo.MiddlewareFunc
}

func NewRouterGroup(relativePath string, initHandle interface{}, regHandle func(rg *echo.Group, group *RouterGroup) error,
	middlewares ...echo.MiddlewareFunc) *RouterGroup {
	return &RouterGroup{
		path:        relativePath,
		initHandle:  initHandle,
		regHandle:   regHandle,
		middleWares: middlewares,
	}
}

// RegisterGroup 将路由组注册到gin
func RegisterGroup(rg *echo.Group, group *RouterGroup) {
	r := rg.Group(group.path)
	if len(group.middleWares) > 0 {
		r.Use(group.middleWares...)
	}
	GetDig().Provide(group.initHandle)
	if err := group.regHandle(r, group); err != nil {
		log.Fatalln(err)
	}
}

// Reg registers handle by DI
func (group RouterGroup) Reg(function interface{}, opts ...dig.InvokeOption) error {
	return GetDig().DI(function, opts...)
}
