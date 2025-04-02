package vo

type SysDictVo struct {
	ID        int64            `json:"id"`        // 主键
	Module    int64            `json:"module"`    // 所属模块
	Code      string           `json:"code"`      // 字典代码
	Name      string           `json:"name"`      // 字典名称
	Describe  string           `json:"describe"`  // 字典描述
	ValueType int64            `json:"valueType"` // 值类型
	Status    int64            `json:"status"`    // 字典状态
	Children  []SysDictChildVo `json:"children"`  // 字典值
}

type SysDictChildVo struct {
	ID        int64  `json:"id"`        // 主键
	DictCode  string `json:"dictCode"`  // 字典代码
	Type      int64  `json:"type"`      // 值类型
	Value     string `json:"value"`     // 值
	Describe  string ` json:"describe"` // 描述
	Label     string `json:"label"`     // 名称
	Style     string `json:"style"`     // 显示的type
	OrderNum  int64  `json:"orderNum"`  // 排序
	ItemClass string `json:"itemClass"` // 样式class
}

type SysCodeList struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
