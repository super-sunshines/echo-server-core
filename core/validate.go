package core

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
	"reflect"
)

var validate *Validator

type Validator struct {
	selectTranslator  ut.Translator
	selectValidator   *validator.Validate
	chineseTranslator ut.Translator
	englishTranslator ut.Translator
	chineseValidator  *validator.Validate
	englishValidator  *validator.Validate
}

func GetValidator() *Validator {
	if validate == nil {
		return initValidator()
	}
	return validate
}
func initValidator() *Validator {
	chineseDict := zh.New()
	englishDict := en.New()
	//设置国际化翻译器
	uni := ut.New(chineseDict, englishDict, englishDict)
	//设置验证器
	chineseValidator := validator.New()
	englishValidator := validator.New()
	//根据参数取翻译器实例
	chineseTranslator, _ := uni.GetTranslator("zh")
	englishTranslator, _ := uni.GetTranslator("en")
	//翻译器注册到validator
	err := zhTranslations.RegisterDefaultTranslations(chineseValidator, chineseTranslator)
	if err != nil {
		return nil
	}
	chineseValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("zh_comment")
	})
	err = enTranslations.RegisterDefaultTranslations(englishValidator, englishTranslator)
	if err != nil {
		return nil
	}
	englishValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Name
	})
	var v = &Validator{
		selectTranslator:  chineseTranslator,
		selectValidator:   chineseValidator,
		chineseTranslator: chineseTranslator,
		englishTranslator: englishTranslator,
		chineseValidator:  chineseValidator,
		englishValidator:  englishValidator,
	}
	validate = v
	return v
}

// ValidateStruct 用于绑定地址栏参数，支持结构体和结构体数组的验证
func (v *Validator) ValidateStruct(param interface{}) (err error) {
	if v.selectValidator == nil {
		err = errors.WithStack(errors.New("validate not init!"))
		return
	}

	// 检查 param 是否是切片或数组
	val := reflect.ValueOf(param)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		// 遍历切片或数组中的每个元素
		for i := 0; i < val.Len(); i++ {
			// 获取当前元素
			elem := val.Index(i).Interface()
			// 验证当前元素
			if _err := v.selectValidator.Struct(elem); _err != nil {
				var errs validator.ValidationErrors
				errors.As(_err, &errs)
				err = errors.New(getFirstKeyValue(errs.Translate(v.selectTranslator)))
				return
			}
		}
	} else {
		// 如果不是切片或数组，直接验证结构体
		if _err := v.selectValidator.Struct(param); _err != nil {
			var errs validator.ValidationErrors
			errors.As(_err, &errs)
			err = errors.New(getFirstKeyValue(errs.Translate(v.selectTranslator)))
			return
		}
	}

	err = nil
	return
}

func (v *Validator) ChangeEnglishTranslator() {
	v.selectTranslator = v.englishTranslator
	v.selectValidator = v.englishValidator

}
func (v *Validator) ChangeChineseTranslator() {
	v.selectValidator = v.chineseValidator
	v.selectTranslator = v.chineseTranslator
}
func getFirstKeyValue(m map[string]string) string {
	for _, value := range m {
		return value
	}
	return ""
}
