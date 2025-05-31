package models

import "time"

type User struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"size:100;not null;index"`
	Password    string `gorm:"not null"`
	Role        int    `gorm:"default:2"` // 1 管理员 2 普通用户
	Status      int    `gorm:"default:1"` // 1 激活 2 不激活
	Email       string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	CreatedBy   int64
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	UpdatedBy   int64
	LastLoginAt time.Time
}
