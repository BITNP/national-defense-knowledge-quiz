package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/config"
	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

type ProblemConfig struct {
	Type   string   `json:"type"`
	Text   string   `json:"text"`
	Data   []string `json:"data"`
	Answer string   `json:"answer"`
	Score  int      `json:"score"`
	Active bool     `json:"active"`
}

type PrizeConfig struct {
	Text   string `json:"text"`
	Remain int    `json:"remain"`
}

type ExamConfig struct {
	Exam struct {
		Title       string `json:"title"`
		Intro       string `json:"intro"`
		LimitTime   int    `json:"limit_time"`
		Random      int    `json:"random"`
		LimitNumber int    `json:"limit_number"`
		Active      bool   `json:"active"`
	} `json:"exam"`
	Problems []ProblemConfig `json:"problems"`
	Prizes   []PrizeConfig   `json:"prizes"`
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db.Init(cfg.DBType, cfg.DBURL)

	configFile := cfg.ConfigPath
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %s: %v", configFile, err)
	}

	var examCfg ExamConfig
	if err := json.Unmarshal(data, &examCfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	problemRepo := &repository.ProblemRepo{}
	prizeRepo := &repository.PrizeRepo{}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		exam := model.Exam{
			Title:       examCfg.Exam.Title,
			Intro:       examCfg.Exam.Intro,
			LimitTime:   examCfg.Exam.LimitTime,
			Random:      examCfg.Exam.Random,
			LimitNumber: examCfg.Exam.LimitNumber,
			Active:      examCfg.Exam.Active,
		}

		// Upsert exam by title
		var existing model.Exam
		result := tx.Where("title = ?", exam.Title).First(&existing)
		if result.Error == nil {
			exam.ID = existing.ID
			exam.CreatedAt = existing.CreatedAt
			if err := tx.Save(&exam).Error; err != nil {
				return fmt.Errorf("failed to update exam: %w", err)
			}
		} else {
			if err := tx.Create(&exam).Error; err != nil {
				return fmt.Errorf("failed to create exam: %w", err)
			}
		}

		if err := problemRepo.DeleteByExamID(tx, exam.ID); err != nil {
			return fmt.Errorf("failed to delete old problems: %w", err)
		}
		if err := prizeRepo.DeleteByExamID(tx, exam.ID); err != nil {
			return fmt.Errorf("failed to delete old prizes: %w", err)
		}

		problems := make([]model.Problem, len(examCfg.Problems))
		for i, p := range examCfg.Problems {
			dataBytes, _ := json.Marshal(p.Data)
			active := true
			if !p.Active {
				active = false
			}
			problems[i] = model.Problem{
				ExamID: exam.ID,
				Type:   p.Type,
				Text:   p.Text,
				Data:   string(dataBytes),
				Answer: p.Answer,
				Score:  p.Score,
				Active: active,
			}
		}
		if err := problemRepo.BulkCreate(tx, problems); err != nil {
			return fmt.Errorf("failed to create problems: %w", err)
		}

		prizes := make([]model.Prize, len(examCfg.Prizes))
		for i, p := range examCfg.Prizes {
			prizes[i] = model.Prize{
				ExamID: exam.ID,
				Text:   p.Text,
				Remain: p.Remain,
			}
		}
		if err := prizeRepo.BulkCreate(tx, prizes); err != nil {
			return fmt.Errorf("failed to create prizes: %w", err)
		}

		fmt.Printf("seeded exam %d with %d problems and %d prizes\n", exam.ID, len(problems), len(prizes))
		return nil
	})

	if err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}
