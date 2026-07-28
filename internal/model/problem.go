package model

import (
	"encoding/json"
	"time"
)

type Problem struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	ExamID    uint   `gorm:"not null;index"`
	Type      string `gorm:"type:varchar(24);not null"`
	Text      string `gorm:"type:text;not null"`
	Data      string `gorm:"type:json;not null"`
	Answer    string `gorm:"type:text;not null"`
	Score     int    `gorm:"not null"`
	Active    bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Problem) DataArray() []string {
	var arr []string
	if err := json.Unmarshal([]byte(p.Data), &arr); err != nil {
		return nil
	}
	return arr
}
