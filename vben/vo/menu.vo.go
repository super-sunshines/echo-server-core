package vo

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/super-sunshines/echo-server-core/core"
	_const "github.com/super-sunshines/echo-server-core/vben/const"
	"github.com/super-sunshines/echo-server-core/vben/gorm/model"
	"sort"
)

type SysMenuWithMeta struct {
	model.SysMenu
	Meta model.SysMenuMetum `gorm:"foreignKey:meta_id" json:"meta" json:"meta"`
}

type SysMenuVo struct {
	ID             int64  `json:"id"`             // id
	Type           int64  `json:"type"`           // 目录类型 0:目录 1:接口
	Pid            int64  `json:"pid"`            // 父目录
	APICode        string `json:"apiCode"`        // 接口代码
	APIDescription string `json:"apiDescription"` // 接口描述
	Component      string `json:"component"`      //组件
	Name           string `json:"name"`           // 路由名称
	Path           string `json:"path"`           // 访问路径
}

type SysMenuMetaVo struct {
	ID                 int64              `json:"id"`                 // id
	MenuID             int64              `json:"menuId"`             // 菜单名称
	Title              string             `json:"title"`              // 路由名称
	Icon               string             `json:"icon"`               // 访问路径
	OrderNum           int64              `json:"order"`              // 排序
	ActiveIcon         string             `json:"activeIcon"`         // 组件地址
	HideInMenu         core.IntBool       `json:"hideInMenu"`         // 隐藏菜单
	HideInTab          core.IntBool       `json:"hideInTab"`          // 标签页隐藏
	HideInBreadcrumb   core.IntBool       `json:"hideInBreadcrumb"`   // 面包屑中隐藏
	HideChildrenInMenu core.IntBool       `json:"hideChildrenInMenu"` // 子菜单隐藏
	Authority          core.Array[string] `json:"authority"`          // 权限代码数组
	ActivePath         string             `json:"activePath"`         // 激活的菜单
	AffixTab           core.IntBool       `json:"affixTab"`           // 固定标签
	AffixTabOrder      int64              `json:"affixTabOrder"`      // 固定标签排序
	IframeSrc          string             `json:"iframeSrc"`          // 内嵌iframe地址
	IgnoreAccess       core.IntBool       `json:"ignoreAccess"`       // 忽略权限
	Link               string             `json:"link"`               // 跳转打开地址
	OpenInNewWindow    core.IntBool       `json:"openInNewWindow"`    // 在新窗口打开
	NoBasicLayout      core.IntBool       `json:"noBasicLayout"`      // 基础布局
}

type UserMenuMetaVo struct {
	SysMenuMetaVo
	BadgeType     string `json:"badgeType"`     // 徽标类型
	BadgeVariants string `json:"badgeVariants"` // 徽标颜色
}

type SysMenuWithMetaVo struct {
	SysMenuVo
	Children []*SysMenuWithMetaVo `json:"children"`
	Meta     SysMenuMetaVo        `json:"meta"`
}

type UserMenuWithMetaVo struct {
	SysMenuVo
	Children []UserMenuWithMetaVo `json:"children"`
	Meta     UserMenuMetaVo       `json:"meta"`
}
type SysSimpleMenuVo struct {
	Name     string             `json:"label"`
	ID       int64              `json:"value"`
	Pid      int64              `json:"-"`
	Type     int64              `json:"-"`
	OrderNum int64              `json:"-"`
	Children []*SysSimpleMenuVo `json:"children"`
}

func BuildSimpleTree(menus []SysMenuWithMeta) []*SysSimpleMenuVo {
	// Map menus to SysSimpleMenuVo
	sysSimpleMenuVoList := slice.Map(menus, func(index int, item SysMenuWithMeta) SysSimpleMenuVo {
		return SysSimpleMenuVo{
			Name:     core.BooleanTo(item.Type != _const.MenuTypeApi, item.Meta.Title, item.APIDescription),
			ID:       item.ID,
			OrderNum: item.Meta.OrderNum,
			Pid:      item.Pid,
			Children: nil, // Initialize Children as nil
		}
	})

	// Create a map to hold menus by their UID for quick access
	menuMap := make(map[int64]*SysSimpleMenuVo)
	for i := range sysSimpleMenuVoList {
		menu := &sysSimpleMenuVoList[i] // Use pointer to the original item
		menuMap[menu.ID] = menu
	}

	var tree []*SysSimpleMenuVo

	// Build the tree structure
	for _, menu := range menuMap {
		if menu.Pid == 0 { // Root menu
			tree = append(tree, menu)
		} else {
			if parent, exists := menuMap[menu.Pid]; exists {
				// Add the child menu as a pointer
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	// Sort the tree by Order
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].OrderNum < tree[j].OrderNum
	})

	return tree
}

// BuildTree converts a flat slice of SysMenuWithMetaVo to a tree structure based on parent-child relationships.
func BuildTree(menus []SysMenuWithMetaVo) []*SysMenuWithMetaVo {
	var tree []*SysMenuWithMetaVo
	menuMap := make(map[int64]*SysMenuWithMetaVo)

	// Initialize map with menu items
	for i := range menus {
		menuMap[menus[i].ID] = &menus[i]
		menuMap[menus[i].ID].Children = []*SysMenuWithMetaVo{} // 使用指针
	}

	// Build tree structure
	for _, menu := range menuMap {
		if menu.Pid == 0 {
			// Top-level menu
			tree = append(tree, menu)
		} else {
			// Add to its parent's children
			if parent, exists := menuMap[menu.Pid]; exists {
				parent.Children = append(parent.Children, menu) // 使用指针
			}
		}

	}
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Meta.OrderNum < tree[j].Meta.OrderNum
	})

	slice.ForEach(tree, func(index int, item *SysMenuWithMetaVo) {
		sort.Slice(item.Children, func(_i, _j int) bool {
			return tree[index].Children[_i].Meta.OrderNum < tree[index].Children[_j].Meta.OrderNum
		})
	})
	return tree
}
