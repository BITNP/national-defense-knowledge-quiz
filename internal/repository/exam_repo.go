package repository

import (
	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type ExamRepo struct{}

func (r *ExamRepo) GetByID(id uint) (*model.Exam, error) {
	var exam model.Exam
	err := db.DB.Where("id = ? AND active = ?", id, true).First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *ExamRepo) GetAllActive() ([]*model.Exam, error) {
	var exams []*model.Exam
	err := db.DB.Where("active = ?", true).Find(&exams).Error
	return exams, err
}
