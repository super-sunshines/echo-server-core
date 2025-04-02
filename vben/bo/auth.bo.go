package bo

type LoginBo struct {
	Username string `valid:"required,min=5" zh_comment:"账号" json:"username" query:"username"` // 账号
	Password string `valid:"required,min=5" zh_comment:"密码" json:"password" query:"password"` // 密码
}
