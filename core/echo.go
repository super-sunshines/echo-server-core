package core

import (
	"fmt"
	"github.com/jedib0t/go-pretty/table"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var PathFuncStrMap = map[string]string{}

func CheckFile(filePath string) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// 文件不存在，输出消息并终止程序
		fmt.Printf("文件 '%s' 不存在\n", filePath)
		os.Exit(1)
	} else if err != nil {
		// 其他错误
		fmt.Printf("检测文件时发生错误: %v\n", err)
		os.Exit(1)
	}
}

type ServerRunOption struct {
	GormOptions        InitGormOptions       // Gorm  数据库操作全局钩子
	PermissionsOptions PermissionsOptions    // 角色权限全局钩子
	BeforeRun          func(echo *echo.Echo) // ServerRun 之前的钩子
	LoggerOptions      LoggerOptions
}

func NewServer(routerGroup []*RouterGroup, option ServerRunOption) {
	CheckFile("./application.yaml")
	InitConfig()
	CheckFile(config.Ip2RegionConfig.FilePath)
	initGormConfig(option.GormOptions)
	initRolePermission(option.PermissionsOptions)
	initRedis()
	initLogMiddleware(option.LoggerOptions)
	e := echo.New()
	// 关闭Banner
	e.HideBanner = true
	// 全局错误方法
	e.HTTPErrorHandler = EchoError()
	// 全局日志接管Echo 日志
	e.Logger = GetLogger()
	// 使用中间件
	e.Use(SetEchoContext)
	e.Use(RequestLoggerMiddleware)
	e.Use(RecoverMiddleware)
	e.Use(middleware.CORS())
	for _, group := range routerGroup {
		RegisterGroup(e.Group(config.Server.GlobalPrefix), group)
	}
	// 生产环境下不打开Swagger
	if config.Server.Dev {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	//
	ResolveRoutes(e)
	if option.BeforeRun != nil {
		option.BeforeRun(e)
	}
	fmt.Println(fmt.Sprintf(`%s==> Server Started !%s`, logger.Green, logger.Reset))
	err := e.Start(fmt.Sprintf(":%d", config.Server.HttpPort))
	if err != nil {
		panic(err)
	}
	return
}

func EchoError() func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if err != nil && !c.Response().Committed {
			c.Logger().Errorf(fmt.Sprintf("%s", err.Error()))
			codeErr := TransformErr(err)
			fmt.Println(fmt.Sprintf("%+v ", codeErr))
			c.Response().Header().Set("Content-Type", "application/json")
			_ = GetContext[any](c).Fail(codeErr)
			return
		}
	}
}

// RecoverMiddleware Panic recovery middleware
func RecoverMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				if IsXError(r) {
					return
				}
				// 创建一个缓冲区来存储堆栈信息
				var stackBuf strings.Builder
				stackBuf.WriteString("Stack Trace:\n")
				// 获取调用的信息
				pcs := make([]uintptr, 50)
				n := runtime.Callers(2, pcs) // Skip the first two frames
				// 准备帧
				frames := runtime.CallersFrames(pcs[:n])
				// 构造详细日志信息
				stackBuf.WriteString(fmt.Sprintf("Recovered from panic: %v\n", r))
				// 迭代每个帧并收集信息
				for {
					frame, more := frames.Next()
					stackBuf.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
					if !more {
						break
					}
				}
				// 记录日志
				logMessage := stackBuf.String()
				if config.Server.Dev {
					fmt.Println(fmt.Sprintf("%s %s %s", logger.Red, logMessage, logger.Reset))
				}
				c.Logger().Error(logMessage)
				context := GetContext[any](c)
				_ = context.Fail(NewErrCodeMsg(500, "Internal Server Error"))
			}
		}()
		return next(c)
	}
}

// CustomResponseWriter 是一个自定义的 ResponseWriter，用于捕获响应体
type CustomResponseWriter struct {
	http.ResponseWriter
	body []byte // 用于存储响应体
}

// Write 重写 Write 方法，捕获响应体
func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...) // 将响应体内容存储到 body 中
	return w.ResponseWriter.Write(b)
}

// RequestLoggerMiddleware 请求日志中间件
func RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	if config.Server.Dev {
		return func(c echo.Context) error {
			if c.Request().Method == "OPTIONS" {
				return next(c)
			}
			zap.L().Info(fmt.Sprintf("请求: %s %s, 参数: %v, 请求开始\n", c.Request().Method, c.Request().URL.Path, c.QueryParams()))
			start := time.Now() // 记录开始时间
			// 使用自定义的 ResponseWriter 替换原始的 ResponseWriter
			originalWriter := c.Response().Writer
			customWriter := &CustomResponseWriter{ResponseWriter: originalWriter}
			c.Response().Writer = customWriter

			// 执行下一步处理
			err := next(c)

			// 计算响应时间
			duration := time.Since(start)

			// 获取请求信息
			method := c.Request().Method
			path := c.Request().URL.Path
			queryParams := c.QueryParams()

			// 获取响应体
			responseBody := string(customWriter.body)

			// 打印请求和响应信息
			zap.L().Info(fmt.Sprintf("%s请求: %s, 参数: %v, 响应体: %s, 耗时: %v",
				method, path, queryParams, responseBody, duration))
			// 返回处理的错误（如果有的话）
			return err
		}
	}
	return func(c echo.Context) error {
		if c.Request().Method == "OPTIONS" {
			return next(c)
		}
		zap.L().Info(fmt.Sprintf("请求: %s %s, 参数: %v, 请求开始\n", c.Request().Method, c.Request().URL.Path, c.QueryParams()))
		start := time.Now() // 记录开始时间
		// 执行下一步处理
		err := next(c)
		// 计算响应时间
		duration := time.Since(start)
		// 获取请求信息
		method := c.Request().Method
		path := c.Request().URL.Path
		queryParams := c.QueryParams()
		// 获取响应体
		// 打印请求和响应信息
		zap.L().Info(fmt.Sprintf("%s请求: %s, 参数: %v, 耗时: %v",
			method, path, queryParams, duration))
		// 返回处理的错误（如果有的话）
		return err
	}
}

// ResolveRoutes 打印所有注册的路由
func ResolveRoutes(e *echo.Echo) {
	routes := e.Routes()

	fmt.Println(fmt.Sprintf(`%s==> All Routers%s`, logger.Magenta, logger.Reset))
	// 创建一个新的表格
	t := table.NewWriter()
	// 设置表格标题
	t.SetTitle("All Routers")
	// 添加表头
	t.AppendHeader(table.Row{"METHOD", "PATH", "FUN"})
	for _, route := range routes {
		PathFuncStrMap[route.Path] = route.Name
		t.AppendRows([]table.Row{
			{route.Method, route.Path, route.Name},
		})
	}
	// 打印表格
	t.SetOutputMirror(os.Stdout) // 设置输出为标准输出
	t.Render()

}
