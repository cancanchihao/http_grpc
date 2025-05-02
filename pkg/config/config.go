package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Mysql struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"mysql"`

	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	Grpc struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"grpc"`

	Http struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"http"`
}

var AppConfig *Config

func InitConfig() error {
	viper.SetConfigName("config")       // 不带扩展名
	viper.SetConfigType("yaml")         // 配置类型
	viper.AddConfigPath(".")            // 当前目录
	viper.AddConfigPath("./pkg/config") // 支持 config 子目录

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}

	AppConfig = &conf
	return nil
}

func init() {
	gin.SetMode(gin.ReleaseMode)

	if err := InitConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}
}
