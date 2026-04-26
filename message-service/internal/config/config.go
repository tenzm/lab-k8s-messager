package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPPort   string `env:"HTTP_PORT" envDefault:"8082"`
	DBDSN      string `env:"DB_DSN,required"`
	UploadsDir string `env:"UPLOADS_DIR" envDefault:"./uploads"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
