package service

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

const gracePeriod = 4200 * time.Millisecond
const maxPrizeRetries = 5

type ExamInfo struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Intro       string `json:"intro"`
	LimitTime   int    `json:"limit_time"`
	Random      int    `json:"random"`
	LimitNumber int    `json:"limit_number"`
	Active      bool   `json:"active"`
}

type ProblemItem struct {
	Type  string   `json:"type"`
	Text  string   `json:"text"`
	Data  []string `json:"data"`
	Score int      `json:"score"`
}

type StartResult struct {
	Title     string        `json:"title"`
	Intro     string        `json:"intro"`
	StartTime string        `json:"start_time"`
	EndTime   string        `json:"end_time"`
	LogID     uint          `json:"log_id"`
	Problems  []ProblemItem `json:"problems"`
}

type SubmitResultItem struct {
	AC     bool   `json:"ac"`
	Answer string `json:"answer,omitempty"`
}

type SubmitResult struct {
	FullScore int                `json:"full_score"`
	Score     int                `json:"score"`
	Result    []SubmitResultItem `json:"result"`
	StartTime string             `json:"start_time"`
	EndTime   string             `json:"end_time"`
	Extra     string             `json:"extra"`
}

type ExamService struct {
	examRepo    *repository.ExamRepo
	problemRepo *repository.ProblemRepo
	prizeRepo   *repository.PrizeRepo
	sessionRepo *repository.ExamSessionRepo
	loc         *time.Location
}

func NewExamService(loc *time.Location) *ExamService {
	return &ExamService{
		examRepo:    &repository.ExamRepo{},
		problemRepo: &repository.ProblemRepo{},
		prizeRepo:   &repository.PrizeRepo{},
		sessionRepo: &repository.ExamSessionRepo{},
		loc:         loc,
	}
}

func (s *ExamService) formatTime(t time.Time) string {
	return t.In(s.loc).Format("2006/01/02 15:04:05")
}

func (s *ExamService) Info(examID uint) (*ExamInfo, error) {
	exam, err := s.examRepo.GetByID(examID)
	if err != nil {
		return nil, err
	}
	return &ExamInfo{
		ID:          exam.ID,
		Title:       exam.Title,
		Intro:       exam.Intro,
		LimitTime:   exam.LimitTime,
		Random:      exam.Random,
		LimitNumber: exam.LimitNumber,
		Active:      exam.Active,
	}, nil
}

func (s *ExamService) Start(examID uint, studentID, name string) (*StartResult, error) {
	exam, err := s.examRepo.GetByID(examID)
	if err != nil {
		return nil, fmt.Errorf("exam not found")
	}

	now := time.Now()

	count, err := s.sessionRepo.CountFinishedOrExpired(examID, studentID, now)
	if err != nil {
		return nil, fmt.Errorf("database error")
	}
	if exam.LimitNumber > 0 && count >= int64(exam.LimitNumber) {
		return nil, fmt.Errorf("提交数量已达上限（%d次）！", exam.LimitNumber)
	}

	session, err := s.sessionRepo.FindUnfinished(examID, studentID, now)
	if err == nil && session != nil {
		return s.buildStartResult(session, exam)
	}

	var problems []model.Problem
	if exam.Random > 0 {
		problems, err = s.problemRepo.GetRandomByExamID(examID, exam.Random)
	} else {
		problems, err = s.problemRepo.GetActiveByExamID(examID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load problems")
	}

	fullScore := 0
	for _, p := range problems {
		fullScore += p.Score
	}

	session = &model.ExamSession{
		ExamID:    examID,
		StudentID: studentID,
		Name:      name,
		FullScore: fullScore,
		EndTime:   now.Add(time.Duration(exam.LimitTime)*time.Second + gracePeriod),
	}

	if err := s.sessionRepo.CreateWithProblems(session, problems); err != nil {
		return nil, fmt.Errorf("failed to create session")
	}

	return s.buildStartResult(session, exam)
}

func (s *ExamService) buildStartResult(session *model.ExamSession, exam *model.Exam) (*StartResult, error) {
	// Reload with problems preloaded
	session, err := s.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	items := make([]ProblemItem, len(session.Problems))
	for i, p := range session.Problems {
		items[i] = ProblemItem{
			Type:  p.Type,
			Text:  p.Text,
			Data:  p.DataArray(),
			Score: p.Score,
		}
	}

	return &StartResult{
		Title:     exam.Title,
		Intro:     exam.Intro,
		StartTime: s.formatTime(session.StartTime),
		EndTime:   s.formatTime(session.EndTime.Add(-gracePeriod)),
		LogID:     session.ID,
		Problems:  items,
	}, nil
}

func (s *ExamService) Submit(logID uint, studentID, name, answersJSON string) (*SubmitResult, error) {
	session, err := s.sessionRepo.GetByIDAndName(logID, studentID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.EndTime) {
		return nil, fmt.Errorf("超出时间限制")
	}

	var submitAnswers []string
	if err := json.Unmarshal([]byte(answersJSON), &submitAnswers); err != nil {
		return nil, fmt.Errorf("invalid answers format")
	}

	answersBytes, _ := json.Marshal(submitAnswers)
	session.SubmitAnswers = string(answersBytes)
	session.EndTime = time.Now()
	session.Finish = true

	score := 0
	results := make([]SubmitResultItem, len(session.Problems))

	for i, p := range session.Problems {
		if i < len(submitAnswers) && submitAnswers[i] == p.Answer {
			results[i] = SubmitResultItem{AC: true}
			score += p.Score
		} else {
			results[i] = SubmitResultItem{AC: false, Answer: p.Answer}
		}
	}

	session.Score = &score

	extra := s.runLottery(session)
	session.Extra = extra

	if err := s.sessionRepo.Update(session); err != nil {
		return nil, fmt.Errorf("failed to save session")
	}

	return &SubmitResult{
		FullScore: session.FullScore,
		Score:     score,
		Result:    results,
		StartTime: s.formatTime(session.StartTime),
		EndTime:   s.formatTime(session.EndTime),
		Extra:     extra,
	}, nil
}

func (s *ExamService) runLottery(session *model.ExamSession) string {
	noPrizeMsg := "很遗憾，没有中奖捏:("

	if session.Score == nil || session.FullScore == 0 {
		return noPrizeMsg
	}

	ratio := float64(*session.Score) / float64(session.FullScore)
	threshold := math.Pow(ratio, 3) * float64(session.FullScore)
	roll := rand.Float64() * float64(session.FullScore)

	if roll > threshold {
		return noPrizeMsg
	}

	for attempt := 0; attempt < maxPrizeRetries; attempt++ {
		prizes, err := s.prizeRepo.GetByExamID(session.ExamID)
		if err != nil || len(prizes) == 0 {
			return noPrizeMsg
		}

		target := weightedRandom(prizes)

		tx := db.DB.Begin()
		text, ok, err := s.prizeRepo.AtomicDecrement(tx, target.ID)
		if err != nil {
			tx.Rollback()
			return noPrizeMsg
		}
		if !ok {
			tx.Rollback()
			continue
		}

		tx.Commit()
		return fmt.Sprintf("恭喜获得%s！", text)
	}

	return noPrizeMsg
}

func weightedRandom(prizes []model.Prize) model.Prize {
	total := 0
	for _, p := range prizes {
		total += p.Remain
	}
	if total == 0 {
		return prizes[0]
	}

	r := rand.Float64() * float64(total)
	cum := 0.0
	for _, p := range prizes {
		cum += float64(p.Remain)
		if r < cum {
			return p
		}
	}
	return prizes[len(prizes)-1]
}
