package services

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"github.com/super-sunshines/echo-server-core/vben/gorm/query"
	"github.com/super-sunshines/echo-server-core/vben/vo"
)

type SysUserService struct {
	userService core.PreGorm[model.SysUser, vo.SysUserVo]
	UserCache   *core.RedisCache[model.SysUser]
}

func NewSysUserService() SysUserService {
	return SysUserService{
		userService: core.NewService[model.SysUser, vo.SysUserVo](),
		UserCache:   core.GetRedisCache[model.SysUser]("sys-user-cache"),
	}
}

func (r SysUserService) InitUserCache(uid int64) {
	_uid := fmt.Sprintf("%d", uid)
	sysUserQuery := query.Use(core.GetGormDB()).SysUser
	first, _ := sysUserQuery.Where(sysUserQuery.ID.Eq(uid)).First()
	r.UserCache.XHSet(_uid, *first)
}
func (r SysUserService) GetUserInfo(uid int64) model.SysUser {
	_uid := fmt.Sprintf("%d", uid)
	if r.UserCache.XHExists(_uid) {
		return r.UserCache.XHGet(_uid)
	} else {
		r.InitUserCache(uid)
	}
	return r.UserCache.XHGet(_uid)
}
func (r SysUserService) GetUserName(uid int64) string {
	return r.GetUserInfo(uid).NickName
}

func (r SysUserService) GetRealName(uid int64) string {
	return r.GetUserInfo(uid).RealName
}
func (r SysUserService) FilterUserIdByStr(str string) []int64 {
	patientQuery := query.Use(core.GetGormDB()).SysUser
	find, _ := patientQuery.Where(patientQuery.NickName.Like(str)).Find()
	return slice.Unique(core.GetFieldValueSlice[int64](find))
}
