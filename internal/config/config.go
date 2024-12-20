package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	APPEnv     string `env:"APP_ENV" envDefault:"local" validate:"required"`
	APPPort    string `env:"APP_PORT" envDefault:"8080" validate:"required,number"`
	DBHost     string `env:"DB_HOST" envDefault:"localhost" validate:"required"`
	DBPort     string `env:"DB_PORT" envDefault:"5434" validate:"required,number"`
	DBUser     string `env:"DB_USER" validate:"required"`
	DBPassword string `env:"DB_PASSWORD" validate:"required"`
	DBName     string `env:"DB_NAME" validate:"required"`
	JWTSecret  string `env:"JWT_SECRET" validate:"required"`
}

func LoadConfig() (*Config, error) {

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var missingVars []string
		for _, err := range err.(validator.ValidationErrors) {
			missingVars = append(missingVars, strings.ToUpper(err.Field()))
		}
		return nil, fmt.Errorf("missing or invalid environment variables: %s", strings.Join(missingVars, ", "))
	}

	return cfg, nil
}

func (c *Config) DatabaseConnectionString() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   fmt.Sprintf("%s:%s", c.DBHost, c.DBPort),
		Path:   c.DBName,
	}

	query := u.Query()
	if strings.ToLower(c.APPEnv) == "local" {
		query.Set("sslmode", "disable")
	} else {
		query.Set("sslmode", "require")
	}
	u.RawQuery = query.Encode()

	return u.String()
}
