package bo

import "github.com/super-sunshines/echo-server-core/core"

type UserInfoBo struct {
	UID      int64  `valid:"required,min=5" zh_comment:"账号" json:"uid" query:"uid"`           // 账号
	Password string `valid:"required,min=5" zh_comment:"密码" json:"password" query:"password"` // 密码
}

type SysUserPageBo struct {
	DepartmentId int64 `query:"departmentId"`
	core.PageParam
}

type SysUserBo struct {
	ID           int64              `json:"id"`           // 主键
	Username     string             `json:"username"`     // 用户名
	NickName     string             `json:"nickName"`     // 昵称
	RealName     string             `json:"realName"`     // 真实姓名
	RoleCodeList core.Array[string] `json:"roleCodeList"` // 角色CODE列表
	Email        string             `json:"email"`        // 邮箱地址
	Avatar       string             `json:"avatar"`       // 头像
	DepartmentId int64              `json:"departmentId"`
	Phone        string             `json:"phone"`      // 手机号
	EnableStatus int64              `json:"status"`     // 状态
	LastOnline   int64              `json:"lastOnline"` // 上次在线时间
}
