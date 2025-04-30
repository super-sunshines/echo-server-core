package core

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	// OK SUCCESS
	OK uint32 = 200
	// SERVER_COMMON_ERROR 服务器普通错误码
	SERVER_COMMON_ERROR uint32 = 100001
	// REQUEST_PARAM_ERROR 请求参数错误
	REQUEST_PARAM_ERROR uint32 = 100002
	NO_RERMIT_ERROR     uint32 = 100002

	// TOKEN_EXPIRE_ERROR TOKEN相关错误码
	TOKEN_EXPIRE_ERROR   uint32 = 100050
	TOKEN_GENERATE_ERROR uint32 = 100051
	TOKEN_FORMAT_ERROR   uint32 = 100052
	TOKEN_ERROR          uint32 = 100051

	// DB_ERROR DB相关错误
	DB_ERROR                      uint32 = 100100
	DB_UPDATE_AFFECTED_ZERO_ERROR uint32 = 100101

	// PARAM_BIND_ERROR 参数绑定失败
	PARAM_BIND_ERROR     uint32 = 100150
	PARAM_VALIDATE_ERROR uint32 = 100150

	// CAPTCHA_KEY_NOT_FOUND_ERROR  验证码相关
	CAPTCHA_KEY_NOT_FOUND_ERROR uint32 = 100200
	CAPTCHA_VERIFY_ERROR        uint32 = 100201

	// USER_NOT_EXIST_ERROR 用户登录相关
	USER_NOT_EXIST_ERROR uint32 = 100250
	USER_PASSWORD_ERROR  uint32 = 100251
	USER_STARTUS_ERROR   uint32 = 100252

	// CURD_AFFECT_NONE_ERROR 通用增删改查相关
	CURD_AFFECT_NONE_ERROR        uint32 = 101000
	CURD_UPDATE_AFFECT_NONE_ERROR uint32 = 101001
	CURD_DATA_EXIST_ERROR         uint32 = 101010
	CURD_DATA_NOT_EXIST_ERROR     uint32 = 101010

	// DICT_NOT_EXIST_ERROR 字典相关
	DICT_NOT_EXIST_ERROR uint32 = 101200

	// GEN_NOT_EXIST_ERROR 代码生成相关
	GEN_NOT_EXIST_ERROR uint32 = 101300

	// QYWX_PARAM_ERROR 企业微信相关
	QYWX_PARAM_ERROR uint32 = 101400

	// FONT_SHOW_MSG
	FONT_SHOW_MSG uint32 = 110000
)
var message = map[uint32]string{
	OK:                  "SUCCESS",
	SERVER_COMMON_ERROR: "服务器开小差啦,稍后再来试一试",
	//REQUEST_PARAM_ERROR: "参数错误",
	NO_RERMIT_ERROR: "暂无权限！",

	TOKEN_EXPIRE_ERROR:   "token失效，请重新登陆",
	TOKEN_GENERATE_ERROR: "生成token失败",
	TOKEN_FORMAT_ERROR:   "Token格式错误",
	TOKEN_ERROR:          "Token错误",

	DB_ERROR:                      "数据库繁忙,请稍后再试",
	DB_UPDATE_AFFECTED_ZERO_ERROR: "更新数据影响行数为0",

	CAPTCHA_KEY_NOT_FOUND_ERROR: "请完成验证码",
	CAPTCHA_VERIFY_ERROR:        "验证码验证失败",

	USER_NOT_EXIST_ERROR: "用户不存在",
	USER_PASSWORD_ERROR:  "密码错误",
	USER_STARTUS_ERROR:   "用户状态异常",

	CURD_AFFECT_NONE_ERROR:        "未影响行数",
	CURD_UPDATE_AFFECT_NONE_ERROR: "修改失败",
	CURD_DATA_EXIST_ERROR:         "数据已存在",
	CURD_DATA_NOT_EXIST_ERROR:     "数据不存在",

	DICT_NOT_EXIST_ERROR: "字典不存在",

	GEN_NOT_EXIST_ERROR: "要生成的表不存在",

	QYWX_PARAM_ERROR: "参数错误",
}

func MapErrMsg(decode uint32) string {
	if msg, ok := message[decode]; ok {
		return msg
	} else {
		return "服务器开小差啦,稍后再来试一试"
	}
}

// 实现 error 接口

func IsCodeErr(decode uint32) bool {
	if _, ok := message[decode]; ok {
		return true
	} else {
		return false
	}
}

type CodeError struct {
	errCode uint32
	errMsg  string
}

// GetErrCode 返回给前端的错误码
func (e *CodeError) GetErrCode() uint32 {
	return e.errCode
}

// GetErrMsg 返回给前端显示的错误信息
func (e *CodeError) GetErrMsg() string {
	return e.errMsg
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d, ErrMsg:%s", e.errCode, e.errMsg)
}

// NewErrCodeMsg 创建带错误信息的 CodeError
func NewErrCodeMsg(errCode uint32, errMsg string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: errMsg}
}

// NewFrontShowErrMsg 创建前端显示的错误
func NewFrontShowErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: FONT_SHOW_MSG, errMsg: errMsg}
}

// NewErrCode 创建仅带错误码的 CodeError
func NewErrCode(errCode uint32) *CodeError {
	return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

// NewErrMsg 创建仅带错误信息的 CodeError
func NewErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: 10000, errMsg: errMsg}
}

// TransformErr 转换错误为 CodeError
func TransformErr(err error) *CodeError {
	// 判断 err 是否是 *CodeError 类型
	var codeErr *CodeError
	if errors.As(err, &codeErr) {
		return codeErr // 返回 CodeError
	} else {
		return NewFrontShowErrMsg(err.Error())
	}

}

// IsXError 转换错误为 CodeError
func IsXError(err any) bool {
	var codeErr *CodeError
	return errors.As(err.(error), &codeErr)
}
