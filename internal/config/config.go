package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram     telegram         `envPrefix:"TG_"`
	Limits       limits           `envPrefix:"LIMITS_"`
	Storege      storage          `envPrefix:"STORAGE_"`
	Redis        redis            `envPrefix:"REDIS_"`
	Models       map[string]Model `env:"MODELS" envKeyValSeparator:"|"`
	DefaultModel string           `env:"DEFAULT_MODEL" envDefault:"gemma2:9b"`
	LogLevel     string           `env:"LOG_LEVEL" envDefault:"INFO"`
}

type Model struct {
	Name   string
	Vision bool
	Tools  bool
}

type telegram struct {
	Token string `env:"TOKEN,required" `
}

type limits struct {
	PerMinute int `env:"PER_MINUTE" envDefault:"2"`
	Tokens    int `env:"TOKEN" envDefault:"2"`
	Db        int `env:"DB" envDefault:"1"`
}

type storage struct {
	History int           `env:"HISTORY" envDefault:"64"`
	TTL     time.Duration `env:"TTL" envDefault:"1h"`
	Db      int           `env:"DB" envDefault:"0"`
}

type redis struct {
	Addr string `env:"ADDR" envDefault:"localhost:6379"`
}

func NewConfig() *Config {
	var cfg Config

	data, err := os.ReadFile("config/miaou.yaml")
	if err != nil {
		log.Warn().Err(err).Msg("config: error load yaml config")
	}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		log.Warn().Err(err).Msg("config: error unmarshal yaml config")
	}
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("config: error load dotenv config")
	}
	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "MIAOU_"}); err != nil {
		log.Fatal().
			Err(err).
			Msg("config: error parse config")
	}
	_, ok := cfg.Models[cfg.DefaultModel]
	if !ok {
		log.Fatal().Msgf("config: not found config for default model %s", cfg.DefaultModel)
	}
	return &cfg
}
