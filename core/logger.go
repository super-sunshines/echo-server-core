package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/xerror"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
	"io"
	"os"
	"time"
)

type Logger struct {
	*log.Logger
	ZapLogger zap.Logger
}

func buildLogString(i ...interface{}) string {
	var result string
	if len(i) > 0 {
		// 尝试将第一个元素转换为 string
		if str, ok := i[0].(string); ok {
			result += str
		}
		// 如果第一个元素是切片，尝试从中获取所有字符串
		if strSlice, ok := i[0].([]interface{}); ok {
			for _, item := range strSlice {
				if str, ok := item.(string); ok {
					result += str + " " // 用空格拼接字符串
				}
			}
		}
	}

	return result
}

func (c *Logger) Prefix() string {
	return c.Logger.Prefix()
}

func (c *Logger) SetPrefix(p string) {
	c.Logger.SetPrefix(p)
}

func (c *Logger) Level() log.Lvl {
	return c.Logger.Level()
}

func (c *Logger) SetLevel(v log.Lvl) {
	c.Logger.SetLevel(v)
}

func (c *Logger) SetHeader(h string) {
	c.Logger.SetHeader(h)
}

func (c *Logger) Print(i ...interface{}) {
	c.ZapLogger.Info(buildLogString(i))
}

func (c *Logger) Printf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Infof(format, args)
}

func (c *Logger) Printj(j log.JSON) {
	c.ZapLogger.Sugar().Info(j)
}

func (c *Logger) Debug(i ...interface{}) {
	c.ZapLogger.Info(buildLogString(i))
}

func (c *Logger) Debugf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Debugf(format, args)
}

func (c *Logger) Debugj(j log.JSON) {
	c.ZapLogger.Sugar().Info(j)
}

func (c *Logger) Info(i ...interface{}) {
	c.ZapLogger.Info(buildLogString(i))
}

func (c *Logger) Infof(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Infof(format, args)
}

func (c *Logger) Infoj(j log.JSON) {
	c.ZapLogger.Sugar().Info(j)
}

func (c *Logger) Warn(i ...interface{}) {
	c.ZapLogger.Warn(buildLogString(i))
}

func (c *Logger) Warnf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Warnf(format, args)
}

func (c *Logger) Warnj(j log.JSON) {
	c.ZapLogger.Sugar().Warn(j)
}

func (c *Logger) Error(i ...interface{}) {
	c.ZapLogger.Error(buildLogString(i))
}

func (c *Logger) Errorf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Errorf(format, args)
}

func (c *Logger) Errorj(j log.JSON) {
	c.ZapLogger.Sugar().Error(j)
}

func (c *Logger) Fatal(i ...interface{}) {
	c.ZapLogger.Fatal(buildLogString(i))
}

func (c *Logger) Fatalj(j log.JSON) {
	c.ZapLogger.Sugar().Fatal(j)
}

func (c *Logger) Fatalf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Fatal(format, args)
}

func (c *Logger) Panic(i ...interface{}) {
	c.ZapLogger.Panic(buildLogString(i))
}

func (c *Logger) Panicf(format string, args ...interface{}) {
	c.ZapLogger.Sugar().Panic(format, args)
}
func (c *Logger) Panicj(j log.JSON) {
	c.ZapLogger.Sugar().Panic(j)
}
func GetLogger() *Logger {
	if zapLogger == nil {
		initLogger()
	}
	var l = &Logger{
		Logger:    log.New("-"),
		ZapLogger: *zapLogger,
	}
	return l
}

var zapLogger *zap.Logger

// BetterColorWriter 是一个自定义的 io.Writer，用于带颜色和格式化的输出
type BetterColorWriter struct {
	writer io.Writer
}

// Write 实现 io.Writer 接口
func (cw *BetterColorWriter) Write(p []byte) (n int, err error) {
	var info map[string]interface{}
	_ = json.Unmarshal(p, &info) // ANSI 颜色代码
	level := info["level"].(string)
	timestamp, _ := time.Parse("2006-01-02T15:04:05.999999999-0700", info["time"].(string))
	caller := info["caller"].(string)
	formatted := fmt.Sprintf(
		"%s[%s]%s %s%-5s%s %s%s%s %s%s%s\n%s%s%s",
		logger.Blue, timestamp.Format(time.DateTime), logger.Reset,
		logger.Green, level, logger.Reset,
		logger.RedBold, caller, logger.Reset,
		logger.Magenta, info["msg"].(string), logger.Reset,
		logger.White, string(p), logger.Reset,
	)
	return cw.writer.Write([]byte(formatted))
}

// 负责设置 encoding 的日志格式
func genZapLogEncoder() zapcore.Encoder {
	// 获取一个指定的的EncoderConfig，进行自定义
	encodeConfig := zap.NewProductionEncoderConfig()
	// 设置每个日志条目使用的键。如果有任何键为空，则省略该条目的部分。
	// 序列化时间。eg: 2022-09-01T19:11:35.921+0800
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// "time":"2022-09-01T19:11:35.921+0800"
	encodeConfig.TimeKey = "time"
	// 将Level序列化为全大写字符串。例如，将info level序列化为INFO。
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 以 package/file:行 的格式 序列化调用程序，从完整路径中删除除最后一个目录外的所有目录。
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

// 负责日志写入的位置
// isStd 是否打印控制台
func genWriteSyncer(filename string, maxsize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 文件位置
		MaxSize:    maxsize,   // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     maxAge,    // 保留旧文件的最大天数
		MaxBackups: maxBackup, // 保留旧文件的最大个数
		Compress:   false,     // 是否压缩/归档旧文件
	}
	syncConsole := BooleanTo(config.Server.Dev,
		zapcore.AddSync(&BetterColorWriter{writer: os.Stderr}),
		zapcore.AddSync(os.Stderr))
	syncFile := zapcore.AddSync(lumberJackLogger)
	// 输出的话打印控制台
	return zapcore.NewMultiWriteSyncer(BooleanTo(config.Server.Dev,
		[]zapcore.WriteSyncer{syncConsole},
		[]zapcore.WriteSyncer{syncConsole, syncFile})...)

}

// InitLogger 初始化Logger std:输出到控制台
func initLogger() {
	var lCfg = config.Logger
	// 获取日志写入位置
	writeSyncer := genWriteSyncer(lCfg.LogFilePath+lCfg.FileName, lCfg.MaxSize, lCfg.MaxBackups, lCfg.MaxAge)
	errWriteSyncer := genWriteSyncer(lCfg.LogFilePath+lCfg.ErrorFileName, lCfg.MaxSize, lCfg.MaxBackups, lCfg.MaxAge)
	// 获取日志编码格式
	encoder := genZapLogEncoder()
	// 获取日志最低等级，即>=该等级，才会被写入。
	var level = new(zapcore.Level)
	err := level.UnmarshalText([]byte(lCfg.Level))
	if err != nil {
		panic(err)
	}
	// 创建一个将日志写入 WriteSyncer 的核心。
	core := zapcore.NewCore(encoder, writeSyncer, level)
	errorCore := zapcore.NewCore(encoder, errWriteSyncer, zapcore.ErrorLevel)
	// 合并核心
	zapLogger = zap.New(zapcore.NewTee(core, errorCore), zap.AddCaller())
	// 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	zap.ReplaceGlobals(zapLogger)
}

type BusinessType int64

const (
	BusinessTypeQuery  BusinessType = 1
	BusinessTypeAdd                 = 2
	BusinessTypeUpdate              = 3
	BusinessTypeDelete              = 4
	BusinessTypeImport              = 5
	BusinessTypeExport              = 6
	BusinessTypeAny                 = 7
)

type LoggerOptions struct {
	LoggerSaver func(RequestInfo, echo.Context)
}

var loggerOptions *LoggerOptions

func initLogMiddleware(options LoggerOptions) {
	loggerOptions = &options
}

type RequestInfo struct {
	Title           string `json:"title"`           // 标题
	BusinessType    int64  `json:"businessType"`    // 业务类型
	CallFunc        string `json:"callFunc"`        // 执行的方法
	RequestMethod   string `json:"requestMethod"`   // 请求方法
	OperateType     int64  `json:"operateType"`     // 操作类型
	OperateName     string `json:"operateName"`     // 操作人员
	OperateDepart   string `json:"operateDepart"`   // 部门名称
	OperateURL      string `json:"operateUrl"`      // 请求地址
	OperateIP       string `json:"operateIp"`       // 请求IP
	OperateLocation string `json:"operateLocation"` // 请求地点
	OperateParam    string `json:"operateParam"`    // 请求参数
	RequestJSONBody string `json:"requestJsonBody"` // 请求体
	JSONResult      string `json:"jsonResult"`      // 返回响应
	ErrorMsg        string `json:"errorMsg"`        // 错误信息
	Status          int64  `json:"status"`          // 操作状态
	OperateTime     Time   `json:"operateTime"`     // 操作时间
	CostTime        int64  `json:"costTime"`
}

var methodBusinessTypeMap = map[string]BusinessType{
	"GET":    BusinessTypeQuery,
	"POST":   BusinessTypeAdd,
	"PUT":    BusinessTypeUpdate,
	"DELETE": BusinessTypeDelete,
}

func Log(title string, operateType ...BusinessType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "OPTIONS" {
				return next(c)
			}
			start := time.Now() // 记录开始时间
			originalWriter := c.Response().Writer
			customWriter := &CustomResponseWriter{ResponseWriter: originalWriter}
			c.Response().Writer = customWriter
			buffer, _ := io.ReadAll(c.Request().Body) //直接读取请求体
			c.Request().Body = io.NopCloser(bytes.NewBuffer(buffer))
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(c.Request().Body)
			err := next(c)
			// 获取响应体
			parse, _ := IPParse(c.RealIP())
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			responseBody := string(customWriter.body)
			resultMap := map[string]any{}
			_ = json.Unmarshal([]byte(responseBody), &resultMap)
			if resultMap["code"] != 200 {
				err = xerror.New(resultMap["msg"].(string))
			}
			if loggerOptions != nil && loggerOptions.LoggerSaver != nil {
				_operateType := methodBusinessTypeMap[c.Request().Method]
				if len(operateType) != 0 {
					_operateType = operateType[0]
				}
				if _operateType == 0 {
					_operateType = BusinessTypeAny
				}
				loggerOptions.LoggerSaver(
					RequestInfo{
						Title:           title,
						BusinessType:    int64(_operateType),
						CallFunc:        PathFuncStrMap[c.Path()],
						RequestMethod:   c.Request().Method,
						OperateType:     int64(_operateType),
						OperateURL:      c.Path(),
						OperateIP:       c.RealIP(),
						OperateLocation: parse,
						OperateParam:    c.QueryParams().Encode(),
						RequestJSONBody: string(buffer),
						JSONResult:      responseBody,
						ErrorMsg:        errMsg,
						Status:          int64(BooleanTo(err == nil, 1, 2)),
						OperateTime:     NewTime(time.Now()),
						CostTime:        time.Since(start).Milliseconds(),
					}, c,
				)
			}

			return err
		}
	}
}
