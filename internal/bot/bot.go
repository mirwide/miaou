package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mirwide/tgbot/internal/bot/model"
	"github.com/mirwide/tgbot/internal/bot/msg"
	"github.com/mirwide/tgbot/internal/config"
	"github.com/mirwide/tgbot/internal/storage"
	"github.com/redis/go-redis/v9"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/go-redis/redis_rate/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ollama "github.com/ollama/ollama/api"

	"github.com/rs/zerolog/log"
)

type Bot struct {
	cfg        *config.Config
	tgClient   *tgbotapi.BotAPI
	tgChannel  *tgbotapi.UpdatesChannel
	httpClient *resty.Client
	ollama     *ollama.Client
	limiter    *redis_rate.Limiter
	storage    *storage.Storage
	translator *message.Printer
}

func NewBot(cfg *config.Config, st *storage.Storage) (*Bot, error) {
	tgClient, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Error().Err(err).Msg("bot: problem start telegram client")
		return nil, err
	}
	tgClient.Debug = true
	log.Info().Msgf("bot: authorized on account %s", tgClient.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	tgChannel := tgClient.GetUpdatesChan(u)

	httpClient := resty.New()

	ollama, err := ollama.ClientFromEnvironment()
	if err != nil {
		log.Error().Err(err).Msg("bot: problem start ollama client")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})

	limiter := redis_rate.NewLimiter(rdb)

	lang := language.MustParse("ru-RU")
	translator := message.NewPrinter(lang)

	return &Bot{
		cfg:        cfg,
		tgChannel:  &tgChannel,
		tgClient:   tgClient,
		httpClient: httpClient,
		ollama:     ollama,
		limiter:    limiter,
		storage:    st,
		translator: translator,
	}, nil
}

func (b *Bot) Run() {
	var duration time.Duration
	for {
		update := <-*b.tgChannel
		if update.Message == nil {
			continue
		}
		conv := NewConversation(update.Message.Chat.ID, b)
		text := update.Message.Text
		if b.RateLimited(update.Message.Chat.ID) {
			conv.SendServiceMessage(msg.ToManyRequests)
			continue
		}

		switch update.Message.Command() {
		case "start":
			text = b.translator.Sprintf(msg.Start)

		case "reset":
			if err := conv.Reset(); err != nil {
				conv.SendServiceMessage(msg.ErrorOccurred)
			}
			text = b.translator.Sprintf(msg.Start)
		}

		log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if duration > 5*time.Second {
			conv.SendAction("typing")
		}

		var f bool = false
		var images []ollama.ImageData

		if update.Message.Photo != nil {
			url, err := b.tgClient.GetFileDirectURL(update.Message.Photo[0].FileID)
			if err != nil {
				log.Error().Err(err).Msgf("bot: problem get file %s", update.Message.Chat.Photo.BigFileID)
				continue
			}
			images = append(images, ollama.ImageData(b.GetFile(url)))
			text = text + " " + update.Message.Caption
		}
		b.storage.SaveMessage(update.Message.Chat.ID, ollama.Message{
			Role:    "user",
			Content: text,
			Images:  images,
		})
		messages := b.storage.GetMessages(update.Message.Chat.ID)
		req := &ollama.ChatRequest{
			Model:     model.Gemma2_27b,
			Messages:  messages,
			Stream:    &f,
			KeepAlive: &ollama.Duration{Duration: time.Minute * 60},
		}
		respFunc := func(resp ollama.ChatResponse) error {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp.Message.Content)
			_, err := b.tgClient.Send(msg)
			log.Debug().Any("ollama", resp).Msg("bot: ollama response")
			b.storage.SaveMessage(update.Message.Chat.ID, resp.Message)
			duration = resp.TotalDuration
			return err
		}
		go func() {
			ctx := context.Background()
			err := b.ollama.Chat(ctx, req, respFunc)
			if err != nil {
				log.Error().Err(err).Msg("bot: problem get response from llm chat")
				conv.SendServiceMessage(msg.ErrorOccurred)
			}
		}()
	}
}

func (b *Bot) GetFile(url string) []byte {

	resp, err := b.httpClient.R().Get(url)
	if err != nil {
		log.Error().Err(err).Msg("file: problem download file")
		return nil
	}
	return resp.Body()
}

func (b *Bot) RateLimited(chatID int64) bool {

	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:chat:%d:1m", chatID)
	res, err := b.limiter.Allow(ctx, key, redis_rate.PerMinute(b.cfg.Limits.PerMinute))
	if err != nil {
		log.Error().Err(err).Msg("limit: problem check limit")
		return true
	}
	log.Info().Msgf("limit: allowed %d remaining %d", res.Allowed, res.Remaining)
	return res.Allowed == 0
}
