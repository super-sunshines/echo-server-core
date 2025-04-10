package bo

import "github.com/super-sunshines/echo-server-core/core"

type SysRolePageBo struct {
	core.PageParam
	ID           int64             `json:"id"`           // 主键
	Code         string            `json:"code"`         // 权限代码
	Description  string            `json:"description"`  // 权限描述
	HomePath     string            ` json:"homePath"`    // 主页目录
	MenuIDList   core.Array[int64] `json:"menuIdList"`   // 目录列表
	EnableStatus core.IntBool      `json:"enableStatus"` // 启用状态
}
type SysRoleBo struct {
	ID             int64             `json:"id"`             // 主键
	Code           string            `json:"code"`           // 权限代码
	Description    string            `json:"description"`    // 权限描述
	Name           string            `json:"name"`           // 角色名称
	QueryStrategy  int64             `json:"queryStrategy"`  // 查询策略
	UpdateStrategy int64             `json:"updateStrategy"` // 更新策略
	HomePath       string            `json:"homePath"`       // 主页目录
	MenuIDList     core.Array[int64] `json:"menuIdList"`     // 目录列表
	EnableStatus   core.IntBool      `json:"enableStatus"`   // 启用状态
}
