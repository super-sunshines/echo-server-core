package core

import (
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Config struct {
	Instance        *viper.Viper
	DataBase        GormConfig
	Server          ServerConfig
	Logger          LogConfig
	Redis           RedisConfig
	Jwt             JwtConfig
	Tencent         TencentConfig
	Ip2RegionConfig Ip2RegionConfig
}
type JwtConfig struct {
	JwtKey            string
	Expire            int64
	MaxLoginFailCount int64
}
type LogConfig struct {
	Level         string // Level 最低日志等级，DEBUG<INFO<WARN<ERROR<FATAL 例如：info-->收集info等级以上的日志
	LogFilePath   string // 日志保存地址
	FileName      string // FileName 日志文件位置
	ErrorFileName string //错误日志名称 错误的日志是一定会写入的
	MaxSize       int    // MaxSize 进行切割之前，日志文件的最大大小(MB为单位)，默认为100MB
	MaxAge        int    // MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
	MaxBackups    int    // MaxBackups 是要保留的旧日志文件的最大数量。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
}
type ServerConfig struct {
	Dev           bool
	HttpPort      int
	WebSocketPath string
	GlobalPrefix  string
	SeverDomain   string
	FrontDomain   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Ip2RegionConfig struct {
	FilePath string
}

type TencentConfig struct {
	WorkWechat WorkWechat
	Cos        Cos
	WechatApp  WechatApp
}

type WechatApp struct {
	Review          bool
	AppId           string
	AppSecret       string
	DefaultNickName string
	DefaultAvatar   string
}

type WorkWechat struct {
	AutoRegister   bool
	CorpId         string
	CorpSecret     string
	AgentId        int64
	RedirectionUrl string
	DefaultRoles   []string
}

type Cos struct {
	SecretId  string
	SecretKey string
	AppId     string
	Region    string
	CdnUrl    string
	Bucket    string
	Action    []string
	Resource  []string
}

type OrderParam struct {
	SortName string `json:"sortName" form:"sortName" query:"sortName" zh_comment:"排序字段" en_comment:"sortName" validate:""`                  // 排序的字段
	SortType string `json:"sortType" form:"sortType" query:"sortType" zh_comment:"排序规则" en_comment:"sortType" validate:"oneof=ASC DESC ''"` // ASC DESC
}

func (r OrderParam) Inject(db *gorm.DB) {
	db.Order(LowerCamelCaseToSnake(r.SortName) + " " + r.SortType)
}

type PageParam struct {
	Page     int `json:"page" form:"page" query:"page" zh_comment:"当前页数" en_comment:"page" validate:"required,gte=1"`                 // 必填，页面值>=1
	PageSize int `json:"pageSize" form:"pageSize" query:"pageSize" zh_comment:"每页条数" en_comment:"pageSize" validate:"required,gte=1"` // 必填，每页条数值>=1
}

type PageResult struct {
	PageParam
	Total    int64 `json:"total"`
	LastPage bool  `json:"lastPage"`
}

type PageResultList[T any] struct {
	PageResult
	Items []T `json:"items"`
}

type QueryIds struct {
	Ids []int64 `json:"ids" query:"ids" form:"ids"  zh_comment:"UID" en_comment:"ids" validate:"required"`
}

type BaseModel struct {
	ID           int64          `gorm:"column:id;type:int(255);primaryKey;autoIncrement:true;comment:主键" json:"id"`     // 主键
	EnableStatus int64          `gorm:"column:enable_status;type:int(1);comment:启用状态" json:"enableStatus"`              // 启用状态
	CreateDept   string         `gorm:"column:create_dept;type:varchar(255);comment:创建部门" json:"createDept"`            // 创建部门
	CreateBy     int64          `gorm:"column:create_by;type:int(11);comment:创建者" json:"createBy"`                      // 创建者
	CreateTime   Time           `gorm:"column:create_time;autoCreateTime;type:datetime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateBy     int64          `gorm:"column:update_by;type:int(11);comment:更新者" json:"updateBy"`                      // 更新者
	UpdateTime   Time           `gorm:"column:update_time;autoUpdateTime;type:datetime;comment:更新时间" json:"updateTime"` // 更新时间
	DeleteTime   gorm.DeletedAt `gorm:"column:delete_time;type:datetime;comment:删除时间" json:"deleteTime"`                // 删除时间
}
