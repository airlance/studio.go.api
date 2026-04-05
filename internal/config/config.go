package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB       DatabaseConfig `envconfig:"DB"`
	Logging  LoggingConfig  `envconfig:"LOG"`
	Server   ServerConfig   `envconfig:"SERVER"`
	Mailer   MailerConfig   `envconfig:"MAILER"`
	Storage  StorageConfig  `envconfig:"STORAGE"`
	Kratos   KratosConfig   `envconfig:"KRATOS"`
	RabbitMQ RabbitMQConfig `envconfig:"RABBITMQ"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DSN" required:"true"`
}

type LoggingConfig struct {
	Level string `envconfig:"LEVEL" default:"debug"`
}

type ServerConfig struct {
	Port               string `envconfig:"PORT" default:"8080"`
	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://dashboard.studio.localhost"`
	DashboardURL       string `envconfig:"DASHBOARD_URL" default:"http://dashboard.studio.localhost"`
}

type MailerConfig struct {
	Provider    string `envconfig:"PROVIDER" default:"log"`
	From        string `envconfig:"FROM" default:"no-reply@studio.localhost"`
	Host        string `envconfig:"HOST" default:"localhost"`
	Port        int    `envconfig:"PORT" default:"1025"`
	Username    string `envconfig:"USERNAME" required:"false"`
	Password    string `envconfig:"PASSWORD" required:"false"`
	LogoPath    string `envconfig:"LOGO_PATH" default:"logo.png"`
	AdminEmails string `envconfig:"ADMIN_EMAILS" required:"false"`
}

type StorageConfig struct {
	Provider      string `envconfig:"PROVIDER" default:"minio"`
	Endpoint      string `envconfig:"ENDPOINT" default:"http://localhost:9000"`
	Region        string `envconfig:"REGION" default:"us-east-1"`
	AccessKeyID   string `envconfig:"ACCESS_KEY_ID" required:"false"`
	SecretKey     string `envconfig:"SECRET_ACCESS_KEY" required:"false"`
	Bucket        string `envconfig:"BUCKET" default:"studio"`
	PublicBaseURL string `envconfig:"PUBLIC_BASE_URL" required:"false"`
	UseSSL        bool   `envconfig:"USE_SSL" default:"false"`
	ForcePath     bool   `envconfig:"FORCE_PATH_STYLE" default:"true"`
	MaxAvatarMB   int64  `envconfig:"MAX_AVATAR_MB" default:"5"`
}

type KratosConfig struct {
	AdminURL  string `envconfig:"ADMIN_URL" default:"http://kratos:4434"`
	PublicURL string `envconfig:"PUBLIC_URL" default:"http://kratos:4433"`
}

type RabbitMQConfig struct {
	URL string `envconfig:"URL" default:"amqp://guest:guest@localhost:5672/"`
}

func (s *ServerConfig) GetCORSAllowedOrigins() []string {
	origins := strings.Split(s.CORSAllowedOrigins, ",")
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			result = append(result, origin)
		}
	}

	return result
}

func Init(ctx context.Context) *Config {
	cfg, err := loadConfig(ctx)
	if err != nil {
		panic(err)
	}

	return cfg
}

func loadConfig(ctx context.Context) (*Config, error) {
	if ctx == nil {
		panic("context must not be nil")
	}

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}
	applyLegacyEmailFallback(&cfg.Mailer)

	return &cfg, nil
}

func applyLegacyEmailFallback(mailer *MailerConfig) {
	if mailer == nil {
		return
	}

	if v := strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")); v != "" && os.Getenv("MAILER_PROVIDER") == "" {
		mailer.Provider = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_FROM")); v != "" && os.Getenv("MAILER_FROM") == "" {
		mailer.From = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_HOST")); v != "" && os.Getenv("MAILER_HOST") == "" {
		mailer.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_PORT")); v != "" && os.Getenv("MAILER_PORT") == "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			mailer.Port = port
		}
	}
	if v := os.Getenv("EMAIL_USERNAME"); v != "" && os.Getenv("MAILER_USERNAME") == "" {
		mailer.Username = v
	}
	if v := os.Getenv("EMAIL_PASSWORD"); v != "" && os.Getenv("MAILER_PASSWORD") == "" {
		mailer.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_ADMIN_EMAILS")); v != "" && os.Getenv("MAILER_ADMIN_EMAILS") == "" {
		mailer.AdminEmails = v
	}
}
