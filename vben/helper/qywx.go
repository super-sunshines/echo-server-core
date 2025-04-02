package helper

import (
	"echo-server-core/core"
	"fmt"
	"github.com/xen0n/go-workwx/v2"
	"net/url"
)

var workWxApp *workwx.WorkwxApp

func InitCacheStore() {
	qywx := core.GetConfig().Tencent.Qywx
	workWxApp = workwx.New(qywx.CorpId).WithApp(qywx.CorpSecret, int64(qywx.AgentId))
	workWxApp.SpawnAccessTokenRefresher()
	workWxApp.SpawnJSAPITicketRefresher()
	workWxApp.SpawnJSAPITicketAgentConfigRefresher()
}

func GetUserInfoByCode(code string) (userInfo *workwx.UserInfo, err error) {
	userIdentityInfo, err := workWxApp.GetUserInfoByCode(code)
	if err != nil {
		err = core.NewFrontShowErrMsg(err.Error())
		return
	}
	userInfo, err = workWxApp.GetUser(userIdentityInfo.UserID)
	if err != nil {
		err = core.NewFrontShowErrMsg(err.Error())
		return
	}
	return
}

func GetAuthUrl() string {
	qywx := core.GetConfig().Tencent.Qywx
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&debug=1&state=#wechat_redirect",
		qywx.CorpId, url.QueryEscape(qywx.RedirectionUrl))
}
