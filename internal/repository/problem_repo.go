package repository

import (
	"context"

	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type ProblemRepo struct{}

func (r *ProblemRepo) GetActiveByExamID(ctx context.Context, examID uint) ([]model.Problem, error) {
	var problems []model.Problem
	err := db.DB.WithContext(ctx).Where("exam_id = ? AND active = ?", examID, true).Find(&problems).Error
	return problems, err
}

func (r *ProblemRepo) GetRandomByExamID(ctx context.Context, examID uint, limit int) ([]model.Problem, error) {
	var problems []model.Problem
	err := db.DB.WithContext(ctx).Where("exam_id = ? AND active = ?", examID, true).
		Order("RANDOM()").
		Limit(limit).
		Find(&problems).Error
	return problems, err
}

func (r *ProblemRepo) GetAllActive(ctx context.Context) (map[uint][]model.Problem, error) {
	var problems []model.Problem
	if err := db.DB.WithContext(ctx).Where("active = ?", true).Find(&problems).Error; err != nil {
		return nil, err
	}
	result := make(map[uint][]model.Problem, 10)
	for i := range problems {
		result[problems[i].ExamID] = append(result[problems[i].ExamID], problems[i])
	}
	return result, nil
}

func (r *ProblemRepo) GetByIDs(ctx context.Context, ids []uint) ([]model.Problem, error) {
	var problems []model.Problem
	err := db.DB.WithContext(ctx).Where("id IN ?", ids).Order("id ASC").Find(&problems).Error
	return problems, err
}

func (r *ProblemRepo) BulkCreate(tx *gorm.DB, problems []model.Problem) error {
	return tx.Create(&problems).Error
}

func (r *ProblemRepo) DeleteByExamID(tx *gorm.DB, examID uint) error {
	return tx.Where("exam_id = ?", examID).Delete(&model.Problem{}).Error
}
