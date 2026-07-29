package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"national-defense-knowledge-quiz/internal/config"
	"national-defense-knowledge-quiz/internal/model"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var dialector gorm.Dialector
	switch cfg.DBType {
	case "postgres":
		dialector = postgres.Open(cfg.DBURL)
	default:
		dialector = sqlite.Open(cfg.DBURL)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: false,
		},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := DB.AutoMigrate(
		&model.Exam{},
		&model.Problem{},
		&model.Prize{},
		&model.ExamSession{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Printf("database connected and migrated (type=%s)", cfg.DBType)

	DB.Exec("DROP INDEX IF EXISTS idx_exam_sessions_exam_id")

	DB.Exec(`CREATE TABLE IF NOT EXISTS exam_session_problems (
		exam_session_id INTEGER NOT NULL,
		problem_id INTEGER NOT NULL
	)`)
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_esp_session_problem ON exam_session_problems(exam_session_id, problem_id)")

	if cfg.DBType == "postgres" {
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_exam_sessions_finished ON exam_sessions(exam_id, student_id) WHERE finish = true")
	}

	if cfg.DBType == "postgres" {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("failed to get underlying sql.DB: %v", err)
		}
		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)
		log.Printf("connection pool configured: max_open=%d max_idle=%d max_lifetime=%s max_idle_time=%s",
			cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLifetime, cfg.DBConnMaxIdleTime)
	}
}
