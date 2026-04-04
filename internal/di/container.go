package di

import (
	"context"
	"time"

	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/infrastructure/rabbitmq"
	"github.com/resoul/studio.go.api/internal/infrastructure/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Container struct {
	Config   *config.Config
	DB       *gorm.DB
	Storage  domain.Storage
	RabbitMQ *rabbitmq.Client
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

	storageSvc, err := storage.NewMinioStorage(cfg)
	if err != nil {
		return nil, err
	}

	rbmqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ)
	if err != nil {
		return &Container{
			Config:  cfg,
			DB:      db,
			Storage: storageSvc,
		}, nil
	}

	return &Container{
		Config:   cfg,
		DB:       db,
		Storage:  storageSvc,
		RabbitMQ: rbmqClient,
	}, nil
}

func (c *Container) Close() error {
	if c == nil {
		return nil
	}

	if c.RabbitMQ != nil {
		c.RabbitMQ.Close()
	}

	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}

	return nil
}
