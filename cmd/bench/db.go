package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func resetDB(dbType, dbURL string) error {
	var dialector gorm.Dialector
	switch dbType {
	case "postgres":
		dialector = postgres.Open(dbURL)
	default:
		dialector = sqlite.Open(dbURL)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	switch dbType {
	case "postgres":
		err = db.Exec("TRUNCATE TABLE exam_session_problems CASCADE").Error
		if err != nil {
			return err
		}
		return db.Exec("TRUNCATE TABLE exam_sessions CASCADE").Error
	default:
		err = db.Exec("DELETE FROM exam_session_problems").Error
		if err != nil {
			return err
		}
		err = db.Exec("DELETE FROM exam_sessions").Error
		if err != nil {
			return err
		}
		_ = db.Exec("DELETE FROM sqlite_sequence WHERE name IN ('exam_sessions', 'exam_session_problems')")
		return nil
	}
}
