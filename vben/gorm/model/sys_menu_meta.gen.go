// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"github.com/super-sunshines/echo-server-core/core"
	"gorm.io/gorm"
)

const TableNameSysMenuMetum = "sys_menu_meta"

// SysMenuMetum mapped from table <sys_menu_meta>
type SysMenuMetum struct {
	ID                 int64              `gorm:"column:id;type:int(11);primaryKey;autoIncrement:true;comment:id" json:"id"`        // id
	Title              string             `gorm:"column:title;type:varchar(255);comment:路由名称" json:"title"`                         // 路由名称
	Icon               string             `gorm:"column:icon;type:varchar(255);comment:访问路径" json:"icon"`                           // 访问路径
	OrderNum           int64              `gorm:"column:order_num;type:int(11);comment:排序" json:"orderNum"`                         // 排序
	ActiveIcon         string             `gorm:"column:active_icon;type:varchar(255);comment:激活时的Icon" json:"activeIcon"`          // 激活时的Icon
	HideInMenu         core.IntBool       `gorm:"column:hide_in_menu;type:int(1);comment:隐藏菜单" json:"hideInMenu"`                   // 隐藏菜单
	HideInTab          core.IntBool       `gorm:"column:hide_in_tab;type:int(1);comment:标签页隐藏" json:"hideInTab"`                    // 标签页隐藏
	HideInBreadcrumb   core.IntBool       `gorm:"column:hide_in_breadcrumb;type:int(1);comment:面包屑中隐藏" json:"hideInBreadcrumb"`     // 面包屑中隐藏
	HideChildrenInMenu core.IntBool       `gorm:"column:hide_children_in_menu;type:int(1);comment:子菜单隐藏" json:"hideChildrenInMenu"` // 子菜单隐藏
	Authority          core.Array[string] `gorm:"column:authority;type:json;comment:权限代码数组" json:"authority"`                       // 权限代码数组
	ActivePath         string             `gorm:"column:active_path;type:varchar(255);comment:激活的菜单" json:"activePath"`             // 激活的菜单
	AffixTab           core.IntBool       `gorm:"column:affix_tab;type:int(1);comment:固定标签" json:"affixTab"`                        // 固定标签
	AffixTabOrder      int64              `gorm:"column:affix_tab_order;type:int(11);comment:固定标签排序" json:"affixTabOrder"`          // 固定标签排序
	IframeSrc          string             `gorm:"column:iframe_src;type:varchar(500);comment:内嵌iframe地址" json:"iframeSrc"`          // 内嵌iframe地址
	IgnoreAccess       core.IntBool       `gorm:"column:ignore_access;type:int(1);comment:忽略权限" json:"ignoreAccess"`                // 忽略权限
	Link               string             `gorm:"column:link;type:varchar(255);comment:跳转打开地址" json:"link"`                         // 跳转打开地址
	OpenInNewWindow    core.IntBool       `gorm:"column:open_in_new_window;type:int(1);comment:在新窗口打开" json:"openInNewWindow"`      // 在新窗口打开
	NoBasicLayout      core.IntBool       `gorm:"column:no_basic_layout;type:int(1);comment:基础布局" json:"noBasicLayout"`             // 基础布局
	CreateDept         int64              `gorm:"column:create_dept;type:int(11);comment:创建部门" json:"createDept"`                   // 创建部门
	CreateBy           int64              `gorm:"column:create_by;type:int(11);comment:创建者" json:"createBy"`                        // 创建者
	CreateTime         core.Time          `gorm:"column:create_time;autoCreateTime;type:datetime;comment:创建时间" json:"createTime"`   // 创建时间
	UpdateBy           int64              `gorm:"column:update_by;type:int(11);comment:更新者" json:"updateBy"`                        // 更新者
	UpdateTime         core.Time          `gorm:"column:update_time;autoUpdateTime;type:datetime;comment:更新时间" json:"updateTime"`   // 更新时间
	DeleteTime         gorm.DeletedAt     `gorm:"column:delete_time;type:datetime;comment:删除时间" json:"deleteTime"`                  // 删除时间
}

// TableName SysMenuMetum's table name
func (*SysMenuMetum) TableName() string {
	return TableNameSysMenuMetum
}
