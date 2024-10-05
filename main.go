package main

import (
	"github.com/mirwide/tgbot/internal/bot"
	"github.com/mirwide/tgbot/internal/config"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.NewConfig()
	tgbot, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("problem start bot")
	}
	tgbot.Run()

}
