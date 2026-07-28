package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type ExamSessionRepo struct{}

func (r *ExamSessionRepo) Create(ctx context.Context, session *model.ExamSession) error {
	return db.DB.WithContext(ctx).Create(session).Error
}

func (r *ExamSessionRepo) Update(ctx context.Context, session *model.ExamSession) error {
	return db.DB.WithContext(ctx).Save(session).Error
}

func (r *ExamSessionRepo) GetByID(ctx context.Context, id uint) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).Preload("Problems").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) GetByIDAndName(ctx context.Context, id uint, studentID string) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).Preload("Problems").
		Where("id = ? AND student_id = ?", id, studentID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) FindSessionsByExamAndStudent(ctx context.Context, examID uint, studentID string) ([]model.ExamSession, error) {
	var sessions []model.ExamSession
	err := db.DB.WithContext(ctx).Where("exam_id = ? AND student_id = ?", examID, studentID).
		Order("id DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *ExamSessionRepo) CountFinishedOrExpired(ctx context.Context, examID uint, studentID string, now time.Time) (int64, error) {
	var count int64
	err := db.DB.WithContext(ctx).Model(&model.ExamSession{}).
		Where("exam_id = ? AND student_id = ? AND (finish = ? OR end_time <= ?)", examID, studentID, true, now).
		Count(&count).Error
	return count, err
}

func (r *ExamSessionRepo) FindUnfinished(ctx context.Context, examID uint, studentID string, now time.Time) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).Preload("Problems").
		Where("exam_id = ? AND student_id = ? AND finish = ? AND end_time > ?", examID, studentID, false, now).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) CreateWithProblems(ctx context.Context, session *model.ExamSession, problems []model.Problem) error {
	return db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}

		if len(problems) == 0 {
			return nil
		}

		const batchSize = 500
		for start := 0; start < len(problems); start += batchSize {
			end := start + batchSize
			if end > len(problems) {
				end = len(problems)
			}

			sql := strings.Builder{}
			sql.WriteString("INSERT INTO exam_session_problems (exam_session_id, problem_id) VALUES ")

			args := make([]interface{}, 0, (end-start)*2)
			for j := start; j < end; j++ {
				if j > start {
					sql.WriteString(", ")
				}
				sql.WriteString("(?, ?)")
				args = append(args, session.ID, problems[j].ID)
			}

			if err := tx.Exec(sql.String(), args...).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ExamSessionRepo) UpdateWithTx(tx *gorm.DB, session *model.ExamSession) error {
	return tx.Save(session).Error
}
