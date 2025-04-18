package eventCenter

import "github.com/asaskevich/EventBus"

const (
	TencentWorkWeChatNewUserEventBusKey = "TencentWorkWeChatNewUserEventBusKey"
)

type TencentWorkWeChatNewUserEventBusData struct {
	SysUid           int64
	WorkWechatName   string
	WorkWechatUserId string
}

var TencentWorkWeChatEventBus = EventBus.New()
