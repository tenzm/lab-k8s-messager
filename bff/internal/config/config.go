package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	HTTPPort       string `env:"HTTP_PORT" envDefault:"8080"`
	UserServiceURL string `env:"USER_SERVICE_URL,required"`
	MsgServiceURL  string `env:"MSG_SERVICE_URL,required"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
