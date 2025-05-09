package core

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
	"sort"
	"strings"
)

var config = &Config{}

type NewConfigParam struct {
	ConfigPaths []string
	ConfigName  string
	EnvPrefix   string
}

func NewConfig(newConfigParam NewConfigParam) Config {
	_config := Config{}
	vip := viper.New()
	for _, path := range newConfigParam.ConfigPaths {
		vip.AddConfigPath(path)
	}
	vip.SetConfigName(newConfigParam.ConfigName)
	vip.SetConfigType("yaml")
	vip.AutomaticEnv()
	vip.SetEnvPrefix(newConfigParam.EnvPrefix)
	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := vip.ReadInConfig(); err != nil {
		panic(err)
	}
	err := vip.Unmarshal(&_config)
	if err != nil {
		panic(err)
	}
	_config.Instance = vip
	fmt.Println(fmt.Sprintf(`%s==> Server Config All Keys  %s`, logger.Cyan, logger.Reset))
	keys := _config.Instance.AllKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, key := range keys {
		fmt.Println(fmt.Sprintf(`%s==> %s`, logger.Cyan, logger.Reset), key)
	}
	return _config
}

func GetConfig() *Config {
	if config == nil {
		InitConfig()
	}
	return config
}

func InitConfig() {
	_config := NewConfig(NewConfigParam{
		ConfigPaths: []string{
			"./",
		},
		ConfigName: "application",
		EnvPrefix:  "CONFIG",
	})
	config = &_config

}
