package models

import "time"

type RepeatRule struct {
	ID          int64      `gorm:"primary_key:autoIncrement"`
	RepeatType  int        // 0 none 1 daily 2 weekly 3 monthly
	Interval    int        // 每隔几天/周/月
	Weekdays    string     // 如果是每周，是那几天，1，3，5
	EndType     string     // 0 不结束 1 到了某天 2 到了几次
	EndDate     *time.Time // 到了某天的话哪天结束
	RepeatTimes int        // 多少次的话，到了多少次结束
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	CreatedBy   int
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	UpdatedBy   int
}
