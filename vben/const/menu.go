package _const

const (
	// MenuTypeApi 接口
	MenuTypeApi = 0
	// MenuTypeMenu 菜单
	MenuTypeMenu = 1
	// MenuTypeCatalogue 目录
	MenuTypeCatalogue = 2
	// MenuTypeIframe 内嵌网页
	MenuTypeIframe = 3
	// MenuTypeLink 外联
	MenuTypeLink = 4
)

var MenuTreeType = []int64{
	// 菜单
	MenuTypeMenu,
	// 目录
	MenuTypeCatalogue,
	// 内嵌网页
	MenuTypeIframe,
	// 外联
	MenuTypeLink,
} //
var MenuTreeTypeWithApi = []int64{
	// 接口
	MenuTypeApi,
	// 菜单
	MenuTypeMenu,
	// 目录
	MenuTypeCatalogue,
	// 内嵌网页
	MenuTypeIframe,
	// 外联
	MenuTypeLink,
}

var DictEnableStatusOK = 1
var DictEnableStatusForbid = 2

const (
	CommonStateOk     = 1
	CommonStateBanned = 2
)
