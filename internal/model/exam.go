package model

import "time"

type Exam struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"type:text;not null"`
	Intro       string `gorm:"type:text"`
	LimitTime   int    `gorm:"not null"`
	Random      int    `gorm:"not null"`
	LimitNumber int    `gorm:"not null"`
	Active      bool   `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
