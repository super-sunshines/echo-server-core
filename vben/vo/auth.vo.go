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
	NickName string   `json:"nickName"`
	Avatar   string   `json:"avatar"`
}

type OauthLoginVo struct {
	AccessToken string `json:"accessToken"`
}
