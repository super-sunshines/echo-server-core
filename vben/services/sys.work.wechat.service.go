package services

import (
	"fmt"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/xen0n/go-workwx/v2"
	"net/url"
)

var tencentWorkWeChatService *TencentWorkWeChatService

type TencentWorkWeChatService struct {
	*workwx.WorkwxApp
}

func NewTencentWorkWeChatService() *TencentWorkWeChatService {
	if tencentWorkWeChatService != nil {
		return tencentWorkWeChatService
	}
	qywx := core.GetConfig().Tencent.WorkWechat
	tencentWorkWeChatService = &TencentWorkWeChatService{
		WorkwxApp: workwx.New(qywx.CorpId).WithApp(qywx.CorpSecret, qywx.AgentId),
	}
	tencentWorkWeChatService.SpawnAccessTokenRefresher()
	tencentWorkWeChatService.SpawnJSAPITicketRefresher()
	tencentWorkWeChatService.SpawnJSAPITicketAgentConfigRefresher()
	return tencentWorkWeChatService
}

func (r *TencentWorkWeChatService) UserInfoByCode(code string) (userInfo *workwx.UserInfo, err error) {
	userIdentityInfo, err := r.GetUserInfoByCode(code)
	if err != nil {
		err = core.NewFrontShowErrMsg(err.Error())
		return
	}
	userInfo, err = r.GetUser(userIdentityInfo.UserID)
	if err != nil {
		err = core.NewFrontShowErrMsg(err.Error())
		return
	}
	return
}

func (r *TencentWorkWeChatService) GetAuthUrl() string {
	qywx := core.GetConfig().Tencent.WorkWechat
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&debug=1&state=#wechat_redirect",
		qywx.CorpId, url.QueryEscape(qywx.RedirectionUrl))
}
