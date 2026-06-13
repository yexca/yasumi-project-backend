package models

import "time"

type Task struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;index"`
	UserID       int64  `gorm:"index"`
	Title        string `gorm:"not null"`
	Description  string
	CategoryID   int `gorm:"default:1"` // 保留字段，不实现功能
	DueDate      *time.Time
	Status       int        `gorm:"default:1"` // 0 已删除 1 未完成 2 已完成
	RemindAt     *time.Time // 防止未设置时间为默认值
	RepeatRuleID int64      `gorm:"default:0"` // 重复规则 0 不重复
	CreatedBy    int64
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedBy    int64
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
