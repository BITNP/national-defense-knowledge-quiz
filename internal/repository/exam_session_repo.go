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
	return db.DB.WithContext(ctx).Model(&model.ExamSession{}).
		Where("id = ?", session.ID).
		UpdateColumns(map[string]interface{}{
			"submit_answers": session.SubmitAnswers,
			"end_time":       session.EndTime,
			"finish":         session.Finish,
			"score":          session.Score,
			"extra":          session.Extra,
		}).Error
}

func (r *ExamSessionRepo) GetByID(ctx context.Context, id uint) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) GetByIDAndName(ctx context.Context, id uint, studentID string) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).
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

type CompletedCount struct {
	ExamID    uint   `gorm:"column:exam_id"`
	StudentID string `gorm:"column:student_id"`
	Count     int64  `gorm:"column:count"`
}

func (r *ExamSessionRepo) CountCompletedGroupBy(ctx context.Context, now time.Time) ([]CompletedCount, error) {
	var rows []CompletedCount
	err := db.DB.WithContext(ctx).Model(&model.ExamSession{}).
		Select("exam_id, student_id, COUNT(*) as count").
		Where("finish = ? OR end_time <= ?", true, now).
		Group("exam_id, student_id").
		Scan(&rows).Error
	return rows, err
}

type ActiveSession struct {
	ID        uint      `gorm:"column:id"`
	ExamID    uint      `gorm:"column:exam_id"`
	StudentID string    `gorm:"column:student_id"`
	EndTime   time.Time `gorm:"column:end_time"`
}

func (r *ExamSessionRepo) ListActiveSessions(ctx context.Context, now time.Time) ([]ActiveSession, error) {
	var rows []ActiveSession
	err := db.DB.WithContext(ctx).Model(&model.ExamSession{}).
		Select("id, exam_id, student_id, end_time").
		Where("finish = ? AND end_time > ?", false, now).
		Scan(&rows).Error
	return rows, err
}

func (r *ExamSessionRepo) FindUnfinished(ctx context.Context, examID uint, studentID string, now time.Time) (*model.ExamSession, error) {
	var session model.ExamSession
	err := db.DB.WithContext(ctx).
		Where("exam_id = ? AND student_id = ? AND finish = ? AND end_time > ?", examID, studentID, false, now).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExamSessionRepo) GetProblemIDsBySessionID(ctx context.Context, sessionID uint) ([]uint, error) {
	var ids []uint
	err := db.DB.WithContext(ctx).Table("exam_session_problems").
		Where("exam_session_id = ?", sessionID).
		Order("problem_id").
		Pluck("problem_id", &ids).Error
	return ids, err
}

func (r *ExamSessionRepo) GetAllProblemIDs(ctx context.Context) (map[uint][]uint, error) {
	type row struct {
		SessionID uint
		ProblemID uint
	}
	var rows []row
	err := db.DB.WithContext(ctx).Table("exam_session_problems").
		Select("exam_session_id AS session_id, problem_id").
		Order("exam_session_id, problem_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint][]uint, len(rows))
	for _, r := range rows {
		result[r.SessionID] = append(result[r.SessionID], r.ProblemID)
	}
	return result, nil
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
