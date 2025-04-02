package services

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"github.com/XiaoSGentle/echo-server-core/vben/gorm/model"
	"github.com/XiaoSGentle/echo-server-core/vben/vo"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type SysDepartmentService struct {
	core.PreGorm[model.SysDepartment, vo.SysDepartmentVo]
	userService     core.PreGorm[model.SysUser, vo.SysUserVo]
	departmentCache *core.RedisCache[[]model.SysDepartment]
}

func NewDepartmentService() SysDepartmentService {
	return SysDepartmentService{
		core.NewService[model.SysDepartment, vo.SysDepartmentVo](),
		core.NewService[model.SysUser, vo.SysUserVo](),
		core.GetRedisCache[[]model.SysDepartment]("sys-department-cache"),
	}
}

func (r SysDepartmentService) GetAllDepartment(c echo.Context) []model.SysDepartment {
	departments := make([]model.SysDepartment, 0)
	if have, value := r.departmentCache.XGet(); have {
		departments = value
	} else {
		_, departments = r.WithContext(c).SkipGlobalHook().FindList()
		r.departmentCache.XSet(departments)
	}
	return departments
}

func (r SysDepartmentService) GetDepartmentUsers(c echo.Context, id int64) ([]model.SysUser, error) {
	childrenIds, err := r.GetChildren(c, id)
	if err != nil {
		return nil, err
	}
	err, users := r.userService.WithContext(c).SkipGlobalHook().FindList(func(db *gorm.DB) *gorm.DB {
		return db.Where("department_id in ?", childrenIds)
	})
	return users, err
}

func (r SysDepartmentService) GetChildren(c echo.Context, id int64) ([]int64, error) {
	departments := r.GetAllDepartment(c)
	// 创建一个映射来存储每个部门的子部门
	childrenMap := make(map[int64][]model.SysDepartment)
	// 遍历所有部门，构建子部门映射
	for _, dept := range departments {
		childrenMap[dept.Pid] = append(childrenMap[dept.Pid], dept)
	}
	// 创建一个切片来存储子部门的 UID
	var childrenIDs []int64
	// 辅助函数：递归获取子部门的 UID
	var collectChildren func(int64)
	collectChildren = func(deptID int64) {
		for _, child := range childrenMap[deptID] {
			childrenIDs = append(childrenIDs, child.ID)
			collectChildren(child.ID)
		}
	}
	// 从指定部门开始收集子部门的 UID
	collectChildren(id)
	childrenIDs = append(childrenIDs, id)
	return childrenIDs, nil
}
