package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gowiki/internal/model"
	"gowiki/internal/pkg/timeutil"
)

func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().In(timeutil.Beijing)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := db.AutoMigrate(
		&model.User{},
		&model.Space{},
		&model.Document{},
		&model.DocumentVersion{},
		&model.DocumentOp{},
		&model.Activity{},
		&model.Favorite{},
		&model.RecentView{},
	); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}
