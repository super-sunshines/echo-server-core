package services

import (
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"time"
)

type SysLoginInfoService struct {
	core.PreGorm[model.SysLoginInfo, model.SysLoginInfo]
}

func NewSysLoginInfoService() SysLoginInfoService {
	return SysLoginInfoService{
		PreGorm: core.NewService[model.SysLoginInfo, model.SysLoginInfo](),
	}
}

func (r SysLoginInfoService) AddLog(c echo.Context, username string, loginType _const.LoginType, status int, msg string) {
	go func() {
		os, browser, agent := core.GetOs(c)
		parse, _ := core.IPParse(c.RealIP())
		newLogInfo := model.SysLoginInfo{
			LoginType:       int64(loginType),
			RequestMethod:   c.Request().Method,
			UserAgent:       agent,
			OperateName:     username,
			Status:          int64(status),
			Browser:         browser,
			Os:              os,
			OperateIP:       c.RealIP(),
			OperateLocation: parse,
			Msg:             msg,
			OperateTime:     core.NewTime(time.Now()),
		}
		r.PreGorm.WithContext(c).SkipGlobalHook().Save(&newLogInfo)
	}()
}
