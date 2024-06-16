package db

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"password"`
	DBName   string `envconfig:"DB_NAME" default:"polovni_auto_alert_db"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

func NewConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	return cfg
}
