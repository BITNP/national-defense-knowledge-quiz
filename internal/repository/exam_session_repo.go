package repository

import (
	"time"

	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type ExamSessionRepo struct{}

func (r *ExamSessionRepo) Create(session *model.ExamSession) error {
	return db.DB.Create(session).Error
}

func (r *ExamSessionRepo) Update(session *model.ExamSession) error {
	return db.DB.Save(session).Error
}

func (r *ExamSessionRepo) GetByID(id uint) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.Preload("Problems").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) GetByIDAndName(id uint, studentID string) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.Preload("Problems").
		Where("id = ? AND student_id = ?", id, studentID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) CountFinishedOrExpired(examID uint, studentID string, now time.Time) (int64, error) {
	var count int64
	err := db.DB.Model(&model.ExamSession{}).
		Where("exam_id = ? AND student_id = ? AND (finish = ? OR end_time <= ?)", examID, studentID, true, now).
		Count(&count).Error
	return count, err
}

func (r *ExamSessionRepo) FindUnfinished(examID uint, studentID string, now time.Time) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.Preload("Problems").
		Where("exam_id = ? AND student_id = ? AND finish = ? AND end_time > ?", examID, studentID, false, now).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) CreateWithProblems(session *model.ExamSession, problems []model.Problem) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		return tx.Model(session).Association("Problems").Replace(problems)
	})
}

func (r *ExamSessionRepo) UpdateWithTx(tx *gorm.DB, session *model.ExamSession) error {
	return tx.Save(session).Error
}
