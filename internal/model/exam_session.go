package model

import (
	"encoding/json"
	"time"
)

type ExamSession struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	ExamID        uint      `gorm:"not null;index:idx_exam_sessions_lookup"`
	StudentID     string    `gorm:"type:varchar(255);not null;index:idx_exam_sessions_lookup"`
	Name          string    `gorm:"type:varchar(255);not null"`
	FullScore     int       `gorm:"not null"`
	Score         *int      `gorm:""`
	SubmitAnswers string    `gorm:"type:json"`
	StartTime     time.Time `gorm:"not null"`
	EndTime       time.Time `gorm:"not null;index:idx_exam_sessions_lookup"`
	Finish        bool      `gorm:"not null;index:idx_exam_sessions_lookup"`
	Extra         string    `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (s *ExamSession) SubmitAnswersArray() []string {
	if s.SubmitAnswers == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s.SubmitAnswers), &arr); err != nil {
		return nil
	}
	return arr
}
