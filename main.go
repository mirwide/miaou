package main

import (
	"github.com/mirwide/tgbot/internal/bot"
	"github.com/mirwide/tgbot/internal/config"
	"github.com/mirwide/tgbot/internal/storage"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.NewConfig()
	st, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("problem init storage")
	}
	tgbot, err := bot.NewBot(cfg, st)
	if err != nil {
		log.Fatal().Err(err).Msg("problem start bot")
	}
	tgbot.Run()

}
