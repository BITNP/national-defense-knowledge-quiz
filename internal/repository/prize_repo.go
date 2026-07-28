package repository

import (
	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

type PrizeRepo struct{}

func (r *PrizeRepo) GetByExamID(examID uint) ([]model.Prize, error) {
	var prizes []model.Prize
	err := db.DB.Where("exam_id = ? AND remain > ?", examID, 0).Find(&prizes).Error
	return prizes, err
}

func (r *PrizeRepo) AtomicDecrement(tx *gorm.DB, id uint) (string, bool, error) {
	var prize model.Prize
	err := tx.Raw(
		"UPDATE prizes SET remain = remain - 1 WHERE id = ? AND remain > 0 RETURNING id, text, remain",
		id,
	).Scan(&prize).Error
	if err != nil {
		return "", false, err
	}
	if prize.ID == 0 {
		return "", false, nil
	}
	return prize.Text, true, nil
}

func (r *PrizeRepo) BulkCreate(tx *gorm.DB, prizes []model.Prize) error {
	return tx.Create(&prizes).Error
}

func (r *PrizeRepo) DeleteByExamID(tx *gorm.DB, examID uint) error {
	return tx.Where("exam_id = ?", examID).Delete(&model.Prize{}).Error
}
