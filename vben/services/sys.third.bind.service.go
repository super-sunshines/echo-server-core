package services

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"gorm.io/gorm"
)

type SysThirdBindService struct {
	core.PreGorm[model.SysUserThirdBind, model.SysUserThirdBind]
}

func NewSysThirdBindService() SysThirdBindService {
	return SysThirdBindService{
		PreGorm: core.NewService[model.SysUserThirdBind, model.SysUserThirdBind](),
	}
}

func (s SysThirdBindService) UidListToWorkWechatUidList(thirdPlatform string, uidList []int64) []string {
	unique := slice.Unique(uidList)
	var users []model.SysUserThirdBind
	core.GetGormDB().Model(model.SysUserThirdBind{}).Where("user_id IN (?)", unique).Where("login_type = ?", thirdPlatform).Find(&users)
	return slice.Map(users, func(index int, item model.SysUserThirdBind) string {
		return item.Openid
	})
}
func (s SysThirdBindService) UidToWorkWechatUid(thirdPlatform string, uid int64) string {
	var user model.SysUserThirdBind
	core.GetGormDB().Model(model.SysUserThirdBind{}).Where("user_id = ?", uid).Where("login_type = ?", thirdPlatform).First(&user)
	return user.Openid
}

func (s SysThirdBindService) WorkWechatUidToUid(thirdPlatform string, workWechatUid string) (uid int64, exist bool) {
	err, bind := s.SetDB(core.GetGormDB()).SkipGlobalHook().FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("openid = ?", workWechatUid).Where("login_type = ?", thirdPlatform)
	})
	return bind.UserID, err == nil
}
