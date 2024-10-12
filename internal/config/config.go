package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Telegram telegram `envPrefix:"TG_"`
	Limits   limits   `envPrefix:"LIMITS_"`
	Storege  storage  `envPrefix:"STORAGE_"`
	Redis    redis    `envPrefix:"REDIS_"`
}

type telegram struct {
	Token string `env:"TOKEN,required" `
}

type limits struct {
	PerMinute int `env:"PER_MINUTE" envDefault:"2"`
	Tokens    int `env:"TOKEN" envDefault:"2"`
}

type storage struct {
	History int           `env:"HISTORY" envDefault:"64"`
	TTL     time.Duration `env:"TTL" envDefault:"1h"`
}

type redis struct {
	Addr string `env:"ADDR" envDefault:"localhost:6379"`
}

func NewConfig() *Config {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		log.Error().
			Err(err).
			Msg("config: error load config")
	}

	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "TGBOT_"}); err != nil {
		log.Fatal().
			Err(err).
			Msg("config: error parse config")
	}
	return &cfg
}
