package repository

import (
	"strings"
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
