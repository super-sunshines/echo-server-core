package bo

type PermissionBo struct {
	ID          int64  `json:"id"`          // 主键
	Code        string `json:"code"`        // 权限代码
	Description string `json:"description"` // 权限描述
	MenuID      int64  `json:"menuId"`      // 目录ID
}
