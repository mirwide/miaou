package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mirwide/miaou/internal/bot"
	"github.com/mirwide/miaou/internal/config"
	"github.com/mirwide/miaou/internal/llm"
	"github.com/mirwide/miaou/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.NewConfig()
	log.Logger = log.With().Caller().Logger()
	switch cfg.LogLevel {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	// Init storage for context messages
	st, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("problem init storage")
	}
	// LLM client
	llm, err := llm.NewLLM()
	if err != nil {
		log.Fatal().Err(err).Msg("problem start ollama client")
	}

	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Pull LLM images if enabled
	wg.Add(1)
	go func() {
		defer wg.Done()
		if cfg.PullImages {
			if err := llm.PullImages(ctx, cfg.Models); err != nil {
				log.Fatal().Err(err).Msg("problem pull image")
			}
		}
	}()

	// Run bot
	miaou, err := bot.NewBot(cfg, llm, st)
	if err != nil {
		log.Fatal().Err(err).Msg("problem start bot")
	}
	miaou.SetDefaultCommands(cfg.Telegram.DefaultLang)
	if err != nil {
		log.Fatal().Err(err).Msg("problem set bot commands")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		miaou.Run(ctx)
	}()

	wg.Wait()
}
