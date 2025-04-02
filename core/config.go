package core

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
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
	fmt.Println(fmt.Sprintf(`%s==> Server Configs %s`, logger.Cyan, logger.Reset))
	fmt.Println(fmt.Sprintf("%+v", *config))
}
