package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mirwide/tgbot/internal/bot/model"
	"github.com/mirwide/tgbot/internal/bot/msg"
	"github.com/mirwide/tgbot/internal/config"
	"github.com/redis/go-redis/v9"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/go-redis/redis_rate/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ollama "github.com/ollama/ollama/api"

	"github.com/rs/zerolog/log"
)

type Bot struct {
	tgclient   *tgbotapi.BotAPI
	httpClient *resty.Client
	ollama     *ollama.Client
	limiter    *redis_rate.Limiter
	translator *message.Printer
}

func NewBot(c *config.Config) (*Bot, error) {
	tgclient, err := tgbotapi.NewBotAPI(c.Telegram.Token)
	if err != nil {
		log.Error().Err(err).Msg("bot: problem start telegram client")
		return nil, err
	}
	tgclient.Debug = true
	log.Info().Msgf("bot: authorized on account %s", tgclient.Self.UserName)

	httpClient := resty.New()

	ollama, err := ollama.ClientFromEnvironment()
	if err != nil {
		log.Error().Err(err).Msg("bot: problem start ollama client")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr,
	})

	limiter := redis_rate.NewLimiter(rdb)

	lang := language.MustParse("ru-RU")
	translator := message.NewPrinter(lang)

	return &Bot{
		tgclient:   tgclient,
		httpClient: httpClient,
		ollama:     ollama,
		limiter:    limiter,
		translator: translator,
	}, nil
}

func (b *Bot) Run() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tgclient.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			var result tgbotapi.Message

			if b.RateLimited(update.Message.Chat.ID) {
				b.SendServiceMessage(update.Message.Chat.ID, msg.ToManyRequests)
				continue
			}

			result, _ = b.SendServiceMessage(update.Message.Chat.ID, msg.Accepted)

			ctx := context.Background()
			var f bool = false
			var images []ollama.ImageData
			text := update.Message.Text
			if update.Message.Photo != nil {
				url, err := b.tgclient.GetFileDirectURL(update.Message.Photo[0].FileID)
				if err != nil {
					log.Error().Err(err).Msgf("bot: problem get file %s", update.Message.Chat.Photo.BigFileID)
					continue
				}
				images = append(images, ollama.ImageData(b.GetFile(url)))
				text = text + " " + update.Message.Caption
			}
			req := &ollama.ChatRequest{
				Model: model.Gemma2_2b,
				Messages: []ollama.Message{
					{
						Role:    "user",
						Content: text,
						Images:  images,
					},
				},
				Stream:    &f,
				KeepAlive: &ollama.Duration{Duration: time.Minute * 60},
			}
			respFunc := func(resp ollama.ChatResponse) error {

				delete := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, result.MessageID)
				b.tgclient.Send(delete)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp.Message.Content)
				msg.ReplyToMessageID = update.Message.MessageID
				_, err := b.tgclient.Send(msg)
				return err
			}

			err := b.ollama.Chat(ctx, req, respFunc)
			if err != nil {
				log.Error().Err(err).Msg("bot: problem get response from llm chat")
				b.SendServiceMessage(update.Message.Chat.ID, msg.ErrorOccurred)
				continue
			}
		}
	}
}

func (b *Bot) SendServiceMessage(chatID int64, message string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, b.translator.Sprintf(message))
	return b.tgclient.Send(msg)
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
	key := fmt.Sprintf("chart: %d", chatID)
	res, err := b.limiter.Allow(ctx, key, redis_rate.PerMinute(2))
	if err != nil {
		log.Error().Err(err).Msg("limit: problem check limit")
		return true
	}
	log.Info().Msgf("limit: allowed %d remaining %d", res.Allowed, res.Remaining)
	return res.Allowed == 0
}
