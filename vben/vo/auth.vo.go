package vo

type LoginVo struct {
	AccessToken string `json:"accessToken"` //token
}

type LoginUserInfoVo struct {
	UserId   int64    `json:"userId"`
	RealName string   `json:"realName"`
	Roles    []string `json:"roles"`
	Username string   `json:"username"`
	HomePath string   `json:"homePath"`
	Avatar   string   `json:"avatar"`
}

type OauthLoginVo struct {
	NeedRegister bool   `json:"needRegister"`
	QywxUid      string `json:"qywxUid"`
	AccessToken  string `json:"accessToken"`
}
