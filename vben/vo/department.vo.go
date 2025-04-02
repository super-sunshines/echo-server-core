package vo

import (
	"gorm.io/gorm"
	"time"
)

type SysDepartmentVo struct {
	ID          int64          `gorm:"column:id;type:int(255);primaryKey;autoIncrement:true;comment:主键" json:"id"`     // 主键
	Pid         int64          `gorm:"column:pid;type:int(255);comment:父ID" json:"pid"`                                // 父ID
	Name        string         `gorm:"column:name;type:varchar(255);comment:部门名称" json:"name"`                         // 部门名称
	Description string         `gorm:"column:description;type:varchar(500);comment:权限描述" json:"description"`           // 权限描述
	Status      int64          `gorm:"column:status;type:int(1);comment:部门状态" json:"status"`                           // 部门状态
	OrderNum    int64          `gorm:"column:order_num;type:int(11);comment:排序" json:"orderNum"`                       // 排序
	CreateDept  int64          `gorm:"column:create_dept;type:int(11);comment:创建部门" json:"createDept"`                 // 创建部门
	CreateBy    int64          `gorm:"column:create_by;type:int(11);comment:创建者" json:"createBy"`                      // 创建者
	CreateTime  time.Time      `gorm:"column:create_time;autoCreateTime;type:datetime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateBy    int64          `gorm:"column:update_by;type:int(11);comment:更新者" json:"updateBy"`                      // 更新者
	UpdateTime  time.Time      `gorm:"column:update_time;autoUpdateTime;type:datetime;comment:更新时间" json:"updateTime"` // 更新时间
	DeleteTime  gorm.DeletedAt `gorm:"column:delete_time;type:datetime;comment:删除时间" json:"deleteTime"`                // 删除时间
}

type SysDepartmentTreeVo struct {
	SysDepartmentVo
	Children []*SysDepartmentTreeVo `json:"children"`
}

// ToSysDepartmentTreeVo 将平面结构转换为树形结构
func ToSysDepartmentTreeVo(list []SysDepartmentVo) []*SysDepartmentTreeVo {
	// 创建一个映射，用于存储 UID 到 SysDepartmentTreeVo 的关系
	nodeMap := make(map[int64]*SysDepartmentTreeVo)
	// 初始化每个部门节点
	for _, item := range list {
		nodeMap[item.ID] = &SysDepartmentTreeVo{
			SysDepartmentVo: item,
			Children:        []*SysDepartmentTreeVo{},
		}
	}
	var tree []*SysDepartmentTreeVo

	// 递归构建树形结构
	for _, item := range list {
		if item.Pid == 0 {
			// 如果没有父级，添加根节点
			tree = append(tree, nodeMap[item.ID])
		} else {
			// 如果有父级，添加到父级的 Children 中
			if parentNode, exists := nodeMap[item.Pid]; exists {
				parentNode.Children = append(parentNode.Children, nodeMap[item.ID])
			}
		}
	}

	return tree
}
