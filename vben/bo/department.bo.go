package bo

import "github.com/super-sunshines/echo-server-core/core"

type SysDepartmentPageBo struct {
	Pid         int64  `json:"pid"`         // 父ID
	Name        string `json:"name"`        // 部门名称
	Description string `json:"description"` // 权限描述
	Status      int64  `json:"status"`      // 部门状态
	OrderNum    int64  `json:"orderNum"`    // 排序
	core.PageParam
}

type SysDepartmentBo struct {
	Pid         int64  `json:"pid"`         // 父ID
	Name        string `json:"name"`        // 部门名称
	Description string `json:"description"` // 权限描述
	Status      int64  `json:"status"`      // 部门状态
	OrderNum    int64  `json:"orderNum"`    // 排序
}
