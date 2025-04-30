package bo

type LoginBo struct {
	Username string `valid:"required,min=5" zh_comment:"账号" json:"username" query:"username"` // 账号
	Password string `valid:"required,min=5" zh_comment:"密码" json:"password" query:"password"` // 密码
}
type UpdateUserInfoBo struct {
	NickName string `gorm:"column:nick_name;type:varchar(255);comment:昵称" json:"nickName"` // 昵称
	Avatar   string `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`      // 头像
}
