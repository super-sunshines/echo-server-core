package bo

import "github.com/XiaoSGentle/echo-server-core/core"

type SysMenuBo struct {
	ID        int64  `json:"id"`        // id
	Pid       int64  `json:"pid"`       // 父目录
	Name      string `json:"name"`      // 路由名称
	Component string `json:"component"` //组件
	Path      string `json:"path"`      // 访问路径
	MetaID    int64  `json:"metaId"`    // metaID
	Type      int64  `json:"type"`
}

type SysMenuMetaBo struct {
	ID                       int64              `json:"id"`                       // id
	MenuID                   int64              `json:"menuId"`                   // 菜单名称
	Title                    string             `json:"title"`                    // 路由名称
	Icon                     string             `json:"icon"`                     // 访问路径
	Order                    int64              `json:"order"`                    // 排序
	ActiveIcon               string             `json:"activeIcon"`               // 组件地址
	HideInMenu               core.IntBool       `json:"hideInMenu"`               // 隐藏菜单
	HideInTab                core.IntBool       `json:"hideInTab"`                // 标签页隐藏
	HideInBreadcrumb         core.IntBool       `json:"hideInBreadcrumb"`         // 面包屑中隐藏
	HideChildrenInMenu       core.IntBool       `json:"hideChildrenInMenu"`       // 子菜单隐藏
	Authority                core.Array[string] `json:"authority"`                // 权限代码数组
	ActivePath               string             `json:"activePath"`               // 激活的菜单
	AffixTab                 core.IntBool       `json:"affixTab"`                 // 固定标签
	AffixTabOrder            int64              `json:"affixTabOrder"`            // 固定标签排序
	IframeSrc                string             `json:"iframeSrc"`                // 内嵌iframe地址
	IgnoreAccess             core.IntBool       `json:"ignoreAccess"`             // 忽略权限
	Link                     string             `json:"link"`                     // 跳转打开地址
	MenuVisibleWithForbidden string             `json:"menuVisibleWithForbidden"` // 可以看到重定向到403
	OpenInNewWindow          core.IntBool       `json:"openInNewWindow"`          // 在新窗口打开
	NoBasicLayout            core.IntBool       `json:"noBasicLayout"`            // 基础布局
}

type UserMenuBo struct {
	SysMenuBo
	Meta SysMenuMetaBo `json:"meta"`
}

type SimpleTreeBo struct {
	IncludeTopLevel    bool `query:"includeTopLevel"`
	IncludePermissions bool `query:"includePermissions"`
}

type AddCodeListBo struct {
	List []ApiCodeBo `json:"list"`
}

type ApiCodeBo struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Pid         int64  `json:"pid"`
}
