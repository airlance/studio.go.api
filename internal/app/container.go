package app

import (
	"context"
	"time"

	"github.com/resoul/studio.go.api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB
}

func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Init(ctx)

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Container{
		Config: cfg,
		DB:     db,
	}, nil
}

func (c *Container) Close() error {
	if c == nil || c.DB == nil {
		return nil
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
