package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ExamSession struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	ExamID        uint      `gorm:"not null;index"`
	StudentID     string    `gorm:"type:varchar(255);not null"`
	Name          string    `gorm:"type:varchar(255);not null"`
	FullScore     int       `gorm:"not null"`
	Score         *int      `gorm:""`
	SubmitAnswers string    `gorm:"type:json"`
	StartTime     time.Time `gorm:"not null"`
	EndTime       time.Time `gorm:"not null"`
	Finish        bool      `gorm:"not null"`
	Extra         string    `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Problems []Problem `gorm:"many2many:exam_session_problems;joinForeignKey:ExamSessionID;joinReferences:ProblemID"`
}

func (s *ExamSession) BeforeCreate(tx *gorm.DB) error {
	if s.SubmitAnswers == "" {
		s.SubmitAnswers = "[]"
	}
	return nil
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
