package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"yasumiProject-Backend/log"
)

//var Config *viper.Viper

type AppConfig struct {
	Server struct {
		Port string
	}
	Postgresql struct {
		Host     string
		Port     string
		User     string
		Password string
		Dbname   string
		Sslmode  string
	}
}

var Config AppConfig

func InitConfig() {
	//Config = viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		//log.Fatalf("配置文件读取失败: %v", err)
		log.Error("配置文件读取失败", zap.Error(err))
	}

	if err := viper.Unmarshal(&Config); err != nil {
		//log.Fatalf("配置文件解析失败: %v", err)
		log.Error("配置文件解析失败", zap.Error(err))
	}
}
