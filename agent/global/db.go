package global

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	dbPath := os.Getenv("GAGENT_DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(DataDir, "agent.db")
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newDBLogger(),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("Agent database initialized at: %s", dbPath)
	return nil
}

func newDBLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}
