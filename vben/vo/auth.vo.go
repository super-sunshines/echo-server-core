package vo

type LoginVo struct {
	AccessToken        string `json:"accessToken"` //token
	NeedChangePassword bool   `json:"needChangePassword"`
	ChangePasswordCode string `json:"changePasswordCode"`
}

type LoginUserInfoVo struct {
	UserId     int64    `json:"userId"`
	RealName   string   `json:"realName"`
	Roles      []string `json:"roles"`
	Username   string   `json:"username"`
	HomePath   string   `json:"homePath"`
	NickName   string   `json:"nickName"`
	Avatar     string   `json:"avatar"`
	Department string   `json:"department"`
}

type OauthLoginVo struct {
	AccessToken string `json:"accessToken"`
}
