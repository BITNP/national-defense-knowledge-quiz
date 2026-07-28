package repository

import (
	"context"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type ExamRepo struct{}

func (r *ExamRepo) GetByID(ctx context.Context, id uint) (*model.Exam, error) {
	var exam model.Exam
	err := db.DB.WithContext(ctx).Where("id = ? AND active = ?", id, true).First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *ExamRepo) GetAllActive(ctx context.Context) ([]*model.Exam, error) {
	var exams []*model.Exam
	err := db.DB.WithContext(ctx).Where("active = ?", true).Find(&exams).Error
	return exams, err
}
