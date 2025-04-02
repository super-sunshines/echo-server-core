package main

import (
	"fmt"
	"github.com/XiaoSGentle/echo-server-core/core"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"strings"
)

var genModels = []string{
	"sys_department",
	"sys_dict",
	"sys_dict_child",
	"sys_menu",
	"sys_menu_meta",
	"sys_user_department",
	"sys_user_third_bind",
	"sys_role",
	"sys_user",
}

// 有特殊表的生成在此填写
var specialModelTagGenConfig = map[string][]gen.ModelOpt{
	"sys_menu_meta": {
		gen.FieldType("hide_in_menu", "core.IntBool"),
		gen.FieldType("hide_in_tab", "core.IntBool"),
		gen.FieldType("hide_in_breadcrumb", "core.IntBool"),
		gen.FieldType("hide_children_in_menu", "core.IntBool"),
		gen.FieldType("affix_tab", "core.IntBool"),
		gen.FieldType("open_in_new_window", "core.IntBool"),
		gen.FieldType("no_basic_layout", "core.IntBool"),
		gen.FieldType("authority", "core.Array[string]"),
		gen.FieldType("ignore_access", "core.IntBool"),
	},

	"sys_role": {
		gen.FieldType("menu_id_list", "core.Array[int64]"),
		gen.FieldType("enable", "core.IntBool"),
	},

	"sys_user": {
		gen.FieldType("role_code_list", "core.Array[string]"),
	},
}

// 通用配置生成
var allModelTagGenConfig = []gen.ModelOpt{
	gen.FieldGORMTag("create_time", func(tag field.GormTag) field.GormTag {
		tag.Set("column", "create_time;autoCreateTime")
		return tag
	}),
	gen.FieldGORMTag("update_time", func(tag field.GormTag) field.GormTag {
		tag.Set("column", "update_time;autoUpdateTime")
		return tag
	}),
	gen.FieldType("delete_time", "gorm.DeletedAt"),
}

func main() {
	core.InitConfig()
	MysqlConfig := core.GetConfig().DataBase
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", MysqlConfig.User,
		MysqlConfig.Pass, MysqlConfig.Host, MysqlConfig.Port, MysqlConfig.DataBase)
	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Errorf("cannot establish db connection: %w", err))
	}
	// 生成实例
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./gorm/query",
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  true,
		Mode:              gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	// 设置目标 db
	g.UseDB(db)
	// 下划线转驼峰
	g.WithJSONTagNameStrategy(lowUpperFunc)
	// 自定义字段的数据类型
	// 统一数字类型为int64,兼容protobuf
	dataMap := map[string]func(detailType gorm.ColumnType) (dataType string){
		"tinyint":   func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"smallint":  func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"mediumint": func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"bigint":    func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"int":       func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"datetime":  func(detailType gorm.ColumnType) (dataType string) { return "time.Time" },
	}
	// 要先于`ApplyBasic`执行
	g.WithDataTypeMap(dataMap)

	for tableName, option := range specialModelTagGenConfig {
		model := g.GenerateModel(tableName,
			append(
				allModelTagGenConfig, // 保留全局配置
				option...,            // 添加表特有配置
			)...,
		)
		g.ApplyBasic(model)
	}

	for _, tableName := range genModels {
		if options, exists := specialModelTagGenConfig[tableName]; exists {
			g.ApplyBasic(g.GenerateModel(tableName, append(allModelTagGenConfig, options...)...))
		} else {
			g.ApplyBasic(g.GenerateModel(tableName, allModelTagGenConfig...))
		}
	}

	// 确保生成过程正确执行
	g.Execute()
}
func lowUpperFunc(columnName string) (tagContent string) {
	// 去掉下划线和横杠，并将后面的字母改为大写
	var modifiedName string
	for i := 0; i < len(columnName); i++ {
		if columnName[i] == '_' {
			if i+1 < len(columnName) {
				modifiedName += strings.ToUpper(string(columnName[i+1]))
				i++
			}
		} else {
			modifiedName += string(columnName[i])
		}
	}
	return modifiedName
}
