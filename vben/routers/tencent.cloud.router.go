package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/services"
)

var TencentCloudRouterGroup = core.NewRouterGroup("/tencent/cloud", NewTencentCloudRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *TencentCloudRouter) {
		rg.GET("/cos/tem-key", m.cosTempKey, core.IgnorePermission())
	})
})

type TencentCloudRouter struct {
	TencentCloudService *services.TencentCloud
}

func NewTencentCloudRouter() *TencentCloudRouter {
	return &TencentCloudRouter{
		TencentCloudService: services.NewTencentCloudService(),
	}
}

// @Summary	COS临时密钥
// @Tags		[系统]腾讯云模块
// @Success	200	{object}	core.ResponseSuccess{data=services.TencentCloudCosTmpKey}
// @Router		/tencent/cloud/cos/tem-key [GET]
func (t TencentCloudRouter) cosTempKey(c echo.Context) error {
	context := core.GetContext[any](c)
	key, err := t.TencentCloudService.GetTempCosKey()
	context.CheckError(err)
	return context.Success(key)
}
