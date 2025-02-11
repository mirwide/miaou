package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ollama "github.com/ollama/ollama/api"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/mirwide/miaou/internal/bot/msg"
	"github.com/mirwide/miaou/internal/config"
	"github.com/mirwide/miaou/internal/llm"
	"github.com/mirwide/miaou/internal/storage"
	_ "github.com/mirwide/miaou/internal/translations"
)

type Bot struct {
	cfg        *config.Config
	tgClient   *tgbotapi.BotAPI
	tgChannel  *tgbotapi.UpdatesChannel
	httpClient *resty.Client
	ollama     *llm.LLM
	limiter    *redis_rate.Limiter
	storage    *storage.Storage
	wg         sync.WaitGroup
}

func NewBot(cfg *config.Config, llm *llm.LLM, st *storage.Storage) (*Bot, error) {
	tgClient, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Error().Err(err).Msg("bot: problem start telegram client")
		return nil, err
	}
	if cfg.LogLevel == "DEBUG" {
		tgClient.Debug = true
	}
	log.Info().Msgf("bot: authorized on account %s", tgClient.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	tgChannel := tgClient.GetUpdatesChan(u)

	httpClient := resty.New()

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
		DB:   cfg.Limits.Db,
	})

	limiter := redis_rate.NewLimiter(rdb)

	return &Bot{
		cfg:        cfg,
		tgChannel:  &tgChannel,
		tgClient:   tgClient,
		httpClient: httpClient,
		ollama:     llm,
		limiter:    limiter,
		storage:    st,
	}, nil
}

func (b *Bot) Run(ctx context.Context) {
	for {
		// non-locking read from channel
		select {
		case <-ctx.Done():
			log.Info().Msg("stop receive message from telegram")
			b.tgClient.StopReceivingUpdates()
			// wait for all response send
			b.wg.Wait()
			return
		case update := <-*b.tgChannel:
			if update.Message == nil && update.CallbackQuery == nil {
				continue
			}
			if update.CallbackQuery != nil {
				log.Info().Msg(update.CallbackQuery.Data)
				conv := NewConversation(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.LanguageCode, b)
				err := conv.SetModel(update.CallbackQuery.Data)
				if err != nil {
					conv.SendServiceMessage(msg.ErrorOccurred)
					continue
				}
				msg := tgbotapi.NewEditMessageTextAndMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					conv.translator.Sprintf("Текущая модель: %s.", update.CallbackQuery.Data),
					tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}})
				_, err = b.tgClient.Send(msg)
				if err != nil {
					log.Error().Err(err).Msg("bot: problem send markup")
				}
				conv.Reset()
				continue
			}
			conv := NewConversation(update.Message.Chat.ID, update.Message.From.LanguageCode, b)
			text := update.Message.Text
			if b.RateLimited(update.Message.Chat.ID) {
				conv.SendServiceMessage(msg.ToManyRequests)
				continue
			}

			switch update.Message.Command() {
			case "reset":
				conv.Reset()
				text = conv.StartMsg()
			case "start", "model":
				conv.SendSelectModel()
				continue
			}

			log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			ticker := time.NewTicker(5 * time.Second) // telegram reset status after 5 seconds
			defer ticker.Stop()
			go func() {
				for {
					select {
					case <-conv.ready:
						return
					case <-ticker.C:
						if _, err := conv.SendAction(tgbotapi.ChatTyping); err != nil {
							log.Error().Err(err).Msg("problem send action")
						}
					}
				}
			}()
			var images []ollama.ImageData

			if update.Message.Photo != nil {
				if conv.model.Vision {
					url, err := b.tgClient.GetFileDirectURL(update.Message.Photo[0].FileID)
					if err != nil {
						log.Error().Err(err).Msgf("problem get file %s", update.Message.Chat.Photo.BigFileID)
						continue
					}
					images = append(images, ollama.ImageData(b.GetFile(url)))
					text = text + " " + update.Message.Caption
				} else {
					conv.SendServiceMessage(msg.ImagesNotAllowed)
					continue
				}
			}
			b.storage.SaveMessage(update.Message.Chat.ID, ollama.Message{
				Role:    "user",
				Content: text,
				Images:  images,
			})
			conv.SendOllama()
		}
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
