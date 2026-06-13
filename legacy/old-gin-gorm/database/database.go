package database

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"yasumiProject-Backend/config"
	"yasumiProject-Backend/log"
)

var DB *gorm.DB

func InitDB() {
	cfg := config.Config.Postgresql

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Dbname, cfg.Port, cfg.Sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database: ", zap.Error(err))
	}
	DB = db
}
