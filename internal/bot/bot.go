package bot

import (
	"context"

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
	limiter    *redis_rate.Limiter
	translator *message.Printer
}

func NewBot(c *config.Config) (*Bot, error) {
	tgclient, err := tgbotapi.NewBotAPI(c.Telegram.Token)
	if err != nil {
		log.Error().Err(err).Msg("problem start telegram client")
		return nil, err
	}
	tgclient.Debug = true
	log.Info().Msgf("bot: authorized on account %s", tgclient.Self.UserName)

	rdb := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr,
	})

	limiter := redis_rate.NewLimiter(rdb)

	lang := language.MustParse("ru-RU")
	translator := message.NewPrinter(lang)

	return &Bot{
		tgclient:   tgclient,
		limiter:    limiter,
		translator: translator,
	}, nil
}

func (b *Bot) Run() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tgclient.GetUpdatesChan(u)

	client, err := ollama.ClientFromEnvironment()
	if err != nil {
		log.Fatal().Err(err).Msg("bot: problem start ollama client")
	}

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			var result tgbotapi.Message

			if b.RateLimited(update.Message.Chat.UserName) {
				b.Reject(update.Message.Chat.ID)
				continue
			}

			result, _ = b.Accept(update.Message.Chat.ID)

			ctx := context.Background()
			req := &ollama.ChatRequest{
				// Model: "gemma2:2b",
				// Model: "llama3.2:1b",
				Model: "llama3.2:3b",
				Messages: []ollama.Message{
					ollama.Message{
						Role:    "user",
						Content: update.Message.Text,
					},
				},
			}
			content := ""
			respFunc := func(resp ollama.ChatResponse) error {

				content += resp.Message.Content
				return nil
			}

			err = client.Chat(ctx, req, respFunc)
			if err != nil {
				log.Error().Err(err).Msg("bot: problem get response from llm chat")
			}
			delete := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, result.MessageID)
			b.tgclient.Send(delete)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, content)
			msg.ReplyToMessageID = update.Message.MessageID
			b.tgclient.Send(msg)
		}
	}
}

func (b *Bot) Accept(chatID int64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, b.translator.Sprintf("Запрос принят."))
	return b.tgclient.Send(msg)
}

func (b *Bot) Reject(chatID int64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, "Слишком много запросов, попробуй позже.")
	return b.tgclient.Send(msg)
}

func (b *Bot) RateLimited(username string) bool {

	ctx := context.Background()
	res, err := b.limiter.Allow(ctx, username, redis_rate.PerMinute(2))
	if err != nil {
		log.Error().Err(err).Msg("limit: problem check limit")
		return true
	}
	log.Info().Msgf("limit: allowed %d remaining %d", res.Allowed, res.Remaining)
	if res.Allowed == 0 {
		return true
	}
	return false
}
