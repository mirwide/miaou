package main

import (
	"github.com/mirwide/miaou/internal/bot"
	"github.com/mirwide/miaou/internal/config"
	"github.com/mirwide/miaou/internal/storage"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.NewConfig()
	st, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("problem init storage")
	}
	miaou, err := bot.NewBot(cfg, st)
	if err != nil {
		log.Fatal().Err(err).Msg("problem start bot")
	}
	miaou.Run()

}
