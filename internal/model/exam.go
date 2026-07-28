package model

import "time"

type Exam struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"type:text;not null;default:'考试名称'"`
	Intro       string `gorm:"type:text"`
	LimitTime   int    `gorm:"not null;default:300"`
	Random      int    `gorm:"not null;default:0"`
	LimitNumber int    `gorm:"not null;default:0"`
	Active      bool   `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
