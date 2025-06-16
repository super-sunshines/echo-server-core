package bo

type LoginBo struct {
	Username string `validate:"required" zh_comment:"账号" json:"username" query:"username"` // 账号
	Password string `validate:"required" zh_comment:"密码" json:"password" query:"password"` // 密码
}
type UpdateUserInfoBo struct {
	NickName string `gorm:"column:nick_name;type:varchar(255);comment:昵称" json:"nickName"` // 昵称
	Avatar   string `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`      // 头像
}
type ChangePasswordBo struct {
	OldPassword string `validate:"required,min=5" zh_comment:"旧密码" json:"oldPassword" query:"oldPassword"` // 旧密码
	NewPassword string `validate:"required,min=5" zh_comment:"新密码" json:"newPassword" query:"newPassword"`
}
type ChangePasswordByCodeBo struct {
	ChangePasswordBo
	Code string `validate:"required" zh_comment:"验证码" json:"code" query:"code"`
}
