package bo

import "github.com/super-sunshines/echo-server-core/core"

type SysDictBo struct {
	ID           int64  `json:"id"`           // 主键
	Module       int64  `json:"module"`       // 所属模块
	Code         string `json:"code"`         // 字典代码
	Regular      string `json:"regular"`      // 正则字符串
	ValueType    int64  `json:"valueType"`    // 值类型
	Name         string `json:"name"`         // 字典名称
	Describe     string `json:"describe"`     // 字典描述
	EnableStatus int64  `json:"enableStatus"` // 字典状态
}
type SysDictPageBo struct {
	Module int64 `query:"module"` // 所属模块
	core.PageParam
}

type SysDictChildBo struct {
	ID        int64  `json:"id"`        // 主键
	DictCode  string `json:"dictCode"`  // 字典代码
	Type      int64  `json:"type"`      // 值类型 1 数字 2 字符串
	Value     string `json:"value"`     // 值
	Label     string `json:"label"`     // 名称
	Style     string `json:"style"`     // 显示的type
	OrderNum  int64  `json:"orderNum"`  // 排序
	Describe  string ` json:"describe"` // 描述
	ItemClass string `json:"itemClass"` // 样式class
}

type SysDictChildPageBo struct {
	DictCode string `json:"dictCode" query:"dictCode"`
	core.PageParam
}
