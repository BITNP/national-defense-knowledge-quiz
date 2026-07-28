package db

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"national-defense-knowledge-quiz/internal/model"
)

const maxOpenConns = 200
const maxIdleConns = 50
const connMaxLifetime = 30 * time.Minute

var DB *gorm.DB

func Init(dbType, dsn string) {
	var dialector gorm.Dialector
	switch dbType {
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		dialector = sqlite.Open(dsn)
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

	log.Printf("database connected and migrated (type=%s)", dbType)

	if dbType == "postgres" {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("failed to get underlying sql.DB: %v", err)
		}
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
		log.Printf("connection pool configured: max_open=%d max_idle=%d max_lifetime=%s", maxOpenConns, maxIdleConns, connMaxLifetime)
	}
}
