# echo-server-core

> 配套的前端地址: https://github.com/super-sunshines/echo-vben-front

## 快速启动

```golang
package main
import (
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben"
	"github.com/super-sunshines/echo-server-core/vben/hooks"
)
func main() {
	groups := make([]*core.RouterGroup, 0)
	groups = append(groups, vben.BaseRouters...)
	groups = append(groups, vben.TencentRouters...)
	groups = append(groups, vben.TencentWorkWechatRouters...)
	core.NewServer(groups, core.ServerRunOption{
		GormOptions:        core.InitGormOptions{GormGlobalHook: hooks.GlobalGormHook},
		PermissionsOptions: hooks.RolePermissionHook,
		LoggerOptions: core.LoggerOptions{
			LoggerSaver: hooks.LoggerMiddlewareHook,
		},
	})
}
```

## Swag生成

```shell
swag init --parseVendor --parseDependency --parseDependencyLevel 2 --parseInternal --parseDepth 1000
```

## 文件配置

```yaml

DataBase:
  Host: 111.231.xx.xx  # 数据库地址
  DataBase: echo-vben-admin #数据库名
  Port: 3306 #端口
  User: echo-vben-admin # 用户名
  Pass: xxx # 密码
# 服务器配置
Server: 
  Dev: true # 是否开发模式
  HttpPort: 7878 # 服务暴露的端口号
  WebSocketPath: /websocket # WeSocket 前缀地址
  GlobalPrefix: /api # 接口全局前缀
  SeverDomain: https://echo-vben-admin.com/ # 服务器域名
  FrontDomain: https://echo-vben-admin.com/ # 前端域名
Redis:  
  Enable: false # 暂时没用
  Addr: 111.231.xxx.xxx:6379 # 服务器地址和端口
  Password: redis_FmXD4d # redis密码
  DB: 4 # DB
Jwt:
  JwtKey: Jwt-JwtKey-SH-TEN # JwtKey
  Expire: 86400 # 过期时间
  MaxLoginFailCount: 5 # 最大登录失败次数
  Strict: false # 是否开启严格模式
  SpecifiedConfig: # 特殊平台的配置
    - Platform: ReacoolMMRXimmerse # 平台
      Expire: 864000 # 过期时间
      Strict: true # 是否开启严格模式
    - Platform: ReacoolMMRAdmin # 平台
      Expire: 864000 # 过期时间
# 日志配置
Logger:
  Level: INFO  # Level 最低日志等级，DEBUG<INFO<WARN<ERROR<FATAL 等级以上的日志才会写入日志文件
  LogFilePath: ./log/ # 日的存储地址
  FileName: app.log # FileName 日志文件位置
  ErrorFileName: error.log # 错误日志配置
  MaxSize: 100    #  MaxSize 进行切割之前，日志文件的最大大小(MB为单位)，默认为100MB
  MaxAge: 30    #  MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
  MaxBackups:  30    #  MaxBackups 是要保留的旧日志文件的最大数量。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
# 离线IP定位库
Ip2RegionConfig:
  FilePath: ./ip2region.xdb
# 腾讯全家桶的配置
Tencent:
  # 微信小程序相关
  WechatApp:
#     自动注册用户
    AutoRegister: true
#     默认的角色权限
    DefaultRoles:
      - SUPER_ADMIN
    AppId: xxx # 小程序appid
    AppSecret: xxx # 小程序appSecret
    DefaultNickName: xxx # 默认昵称
    DefaultAvatar: xxx # 默认头像
  #   企业微信的
  WorkWechat: 
    CorpId: xxx  # 企业微信的corpId
    CorpSecret: xxx # 企业微信的corpSecret
    AgentId: xxx # 企业微信的AgentId
    RedirectionUrl: xxx # 企业微信的RedirectionUrl
  #  存储桶
  Cos:
    SecretId: AKI****************RHs56s********
    SecretKey: bw3ua***********H0TjgRzHLAj***********
    AppId: 130***********
    Region: ap-shanghai
    Bucket: xxx-130xxxx078
    CdnUrl: https://xxx-130xxxx078.cos.ap-shanghai.myqcloud.com
    Action:
      - name/ci:CreateAuditingTextJob
      - name/ci:CreateAuditingPictureJob
    Resource:
      - qcs::ci:%s:uid/%s:bucket/%s/*
```

```go
type Config struct {
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
	AppId           string
	AppSecret       string
	DefaultNickName string
	DefaultAvatar   string
}

type WorkWechat struct {
	CorpId         string
	CorpSecret     string
	AgentId        int64
	RedirectionUrl string
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

```