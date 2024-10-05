package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Telegram telegram `envPrefix:"TG_"`
	Limits   limits   `envPrefix:"LIMITS_"`
	Redis    redis    `envPrefix:"REDIS_"`
}

type telegram struct {
	Token string `env:"TOKEN,required" `
}

type limits struct {
	Rate int `env:"RATE" envDefault:"2"`
}

type redis struct {
	Addr string `env:"ADDR" envDefault:"localhost:6379"`
}

func NewConfig() *Config {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		log.Fatal().
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
