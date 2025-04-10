package vo

import "github.com/super-sunshines/echo-server-core/core"

type SysRoleVo struct {
	ID             int64             `json:"id"`             // 主键
	Code           string            `json:"code"`           // 权限代码
	Name           string            `json:"name"`           // 角色名称
	Description    string            `json:"description"`    // 权限描述
	MenuIDList     core.Array[int64] `json:"menuIdList"`     // 目录列表
	QueryStrategy  int64             `json:"queryStrategy"`  // 查询策略
	UpdateStrategy int64             `json:"updateStrategy"` // 更新策略
	EnableStatus   core.IntBool      `json:"enableStatus"`   // 启用状态
	HomePath       string            `json:"homePath"`       // 主页目录
}
