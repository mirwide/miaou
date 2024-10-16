package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mirwide/miaou/internal/bot/model"
	"github.com/mirwide/miaou/internal/bot/msg"
	ollama "github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type conversation struct {
	id         int64
	bot        *Bot
	model      model.Model
	translator *message.Printer
}

func NewConversation(chatID int64, bot *Bot, modelName string, lang string) *conversation {
	var l string
	switch lang {
	case "ru", "kz", "ua":
		l = "ru-RU"
	default:
		l = "en-US"
	}
	m, err := model.NewModel(modelName)
	if err != nil {
		m, _ = model.NewModel(bot.cfg.DefaultModel)
	}

	translator := message.NewPrinter(language.MustParse(l))
	return &conversation{
		id:         chatID,
		bot:        bot,
		model:      m,
		translator: translator,
	}
}

func (c *conversation) SendServiceMessage(message string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.id, c.translator.Sprintf(message))
	return c.bot.tgClient.Send(msg)
}

func (c *conversation) SendAction(action string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewChatAction(c.id, action)
	return c.bot.tgClient.Send(msg)
}

func (c *conversation) Reset() error {
	return c.bot.storage.Clear(c.id)
}

func (c *conversation) StartMsg() string {
	return c.translator.Sprintf(msg.Start)
}

func (c *conversation) OllamaCallback(resp ollama.ChatResponse) error {
	var err error
	log.Debug().Any("ollama", resp).Msg("bot: ollama response")
	c.bot.storage.SaveMessage(c.id, resp.Message)
	if len(resp.Message.ToolCalls) > 0 {
		var msg ollama.Message
		for _, call := range resp.Message.ToolCalls {
			switch call.Function.Name {
			case "get_time":
				msg = ollama.Message{
					Role:    "tool",
					Content: c.GetTime(),
				}
			default:
				msg = ollama.Message{
					Role:    "tool",
					Content: fmt.Sprintf("Функция %s не поддерживается.", call.Function.Name),
				}
			}
			c.bot.storage.SaveMessage(c.id, msg)
		}
		c.SendOllama()
	} else {

		//duration = resp.TotalDuration
		msg := tgbotapi.NewMessage(c.id, resp.Message.Content)
		_, err = c.bot.tgClient.Send(msg)
	}
	return err
}

func (c *conversation) SendOllama() {
	var f bool = false
	var tools ollama.Tools
	messages := c.bot.storage.GetMessages(c.id)
	if c.model.SupportTools {
		tools = ollama.Tools{
			ollama.Tool{
				Type: "function",
				Function: ollama.ToolFunction{
					Name:        "get_time",
					Description: "Получить текущее время",
				},
			}}
	}
	req := &ollama.ChatRequest{
		Model:     c.model.Name,
		Messages:  messages,
		Stream:    &f,
		KeepAlive: &ollama.Duration{Duration: time.Hour * 12},
		Tools:     tools,
	}
	go func() {
		ctx := context.Background()
		err := c.bot.ollama.Chat(ctx, req, c.OllamaCallback)
		if err != nil {
			log.Error().Err(err).Msg("bot: problem get response from llm chat")
			c.SendServiceMessage(msg.ErrorOccurred)
		}
	}()
}

func (c *conversation) GetTime() string {
	now := time.Now()
	return fmt.Sprintf("Текущее время.\nГод: %d\nМесяц: %d\nДень: %d\nЧас: %d\nМинута: %d\nСекунда: %d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}
