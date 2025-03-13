package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		DSN string
	}
}

var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("未找到配置文件，使用默认配置")
		} else {
			log.Fatalf("读取配置文件错误: %v", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("解析配置文件错误: %v", err)
	}

	// 环境变量覆盖
	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.Server.Port = port
	}
}
