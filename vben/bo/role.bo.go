package bo

import (
	"echo-server-core/core"
)

type SysRolePageBo struct {
	core.PageParam
	ID          int64             `json:"id"`          // 主键
	Code        string            `json:"code"`        // 权限代码
	Description string            `json:"description"` // 权限描述
	HomePath    string            ` json:"homePath"`   // 主页目录
	MenuIDList  core.Array[int64] `json:"menuIdList"`  // 目录列表
	Enable      core.IntBool      `json:"enable"`      // 启用状态
}
type SysRoleBo struct {
	ID                 int64             `json:"id"`                 // 主键
	Code               string            `json:"code"`               // 权限代码
	Description        string            `json:"description"`        // 权限描述
	Name               string            `json:"name"`               // 角色名称
	HomePath           string            `json:"homePath"`           // 主页目录
	DataHandleStrategy int64             `json:"dataHandleStrategy"` // 数据处理策略
	MenuIDList         core.Array[int64] `json:"menuIdList"`         // 目录列表
	Enable             core.IntBool      `json:"enable"`             // 启用状态
}
