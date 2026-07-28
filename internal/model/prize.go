package model

import "time"

type Prize struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	ExamID    uint   `gorm:"not null;index"`
	Text      string `gorm:"type:text;not null"`
	Remain    int    `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
