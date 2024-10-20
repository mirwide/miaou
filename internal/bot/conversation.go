package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mirwide/miaou/internal/bot/msg"
	"github.com/mirwide/miaou/internal/config"
	"github.com/mirwide/miaou/internal/storage"
	"github.com/mirwide/miaou/internal/tools"
	"github.com/mirwide/miaou/internal/tools/wiki"
	ollama "github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type conversation struct {
	id         int64
	bot        *Bot
	model      config.Model
	translator *message.Printer
}

func NewConversation(chatID int64, bot *Bot) *conversation {
	var l string
	var m config.Model
	c, _ := bot.storage.GetConversation(chatID)
	lang := "ru"
	switch lang {
	case "ru", "kz", "ua":
		l = "ru-RU"
	default:
		l = "en-US"
	}

	m, ok := bot.cfg.Models[c.Model]
	if !ok {
		log.Info().Msgf("model %s not found use default", c.Model)
		m = bot.cfg.Models[bot.cfg.DefaultModel]
	}

	translator := message.NewPrinter(language.MustParse(l))
	_ = bot.storage.SaveConversation(chatID, storage.Conversation{Model: m.Name})
	return &conversation{
		id:         chatID,
		bot:        bot,
		model:      m,
		translator: translator,
	}
}

func (c *conversation) SendServiceMessage(message string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(c.id, c.translator.Sprintf(message))
	result, err := c.bot.tgClient.Send(msg)
	if err != nil {
		log.Error().Err(err).Msg("convarsation: problem send service message")
	}
	return result
}

func (c *conversation) SendAction(action string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewChatAction(c.id, action)
	return c.bot.tgClient.Send(msg)
}

func (c *conversation) Reset() {
	c.bot.storage.Clear(c.id)
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
					Content: tools.GetTime(),
				}
			case "get_weather":
				city := call.Function.Arguments["city"]
				days, _ := call.Function.Arguments["forecast_days"].(int)
				msg = ollama.Message{
					Role:    "tool",
					Content: tools.GetWeather(fmt.Sprintf("%v", city), days),
				}
			case "get_wiki":
				keyword := call.Function.Arguments["keyword"]
				lang := call.Function.Arguments["lang"]
				msg = ollama.Message{
					Role:    "tool",
					Content: wiki.GetWiki(fmt.Sprintf("%v", keyword), fmt.Sprintf("%v", lang)),
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
		msg := tgbotapi.NewMessage(c.id, resp.Message.Content)
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err = c.bot.tgClient.Send(msg)
	}
	return err
}

func (c *conversation) SendOllama() {
	var f bool = false
	var t ollama.Tools
	messages := c.bot.storage.GetMessages(c.id)
	messages = append([]ollama.Message{{Role: "system",
		Content: "У тебя есть доступ к wikipedia. Запрашивай информацию с wikipedia если у тебя не достаточно данных. Возвращай только те данные с wikipedia о кторых спрашивали. Если ты не знаешь каких-то слов ищи их в wikipedia. Не упоминай что у тебя есть доступ к wikipedia."}}, messages...)
	if c.model.Tools && messages[len(messages)-1].Content != msg.Start {
		t = ollama.Tools{
			ollama.Tool{
				Type: "function",
				Function: ollama.ToolFunction{
					Name:        "get_time",
					Description: "Получить текущее время",
				},
			},
			ollama.Tool{
				Type: "function",
				Function: ollama.ToolFunction{
					Name:        "get_weather",
					Description: "Получить текущую погоду по городу",
					Parameters: tools.Parameters{
						Type:     "object",
						Required: []string{"city"},
						Properties: tools.NewProperties(map[string]tools.Properties{
							"city": {Type: "string", Description: "Название города в транслите"},
							"forecast_days": {Type: "int",
								Description: "Количество дней прогноза, должно быть равно 1 если требуется прогноз на сегодняшний день. Максимальное значение 16"},
						}),
					},
				},
			},
			ollama.Tool{
				Type: "function",
				Function: ollama.ToolFunction{
					Name:        "get_wiki",
					Description: "Получить информацию по ключевому слову",
					Parameters: tools.Parameters{
						Type:     "object",
						Required: []string{"keyword"},
						Properties: tools.NewProperties(map[string]tools.Properties{
							"keyword": {Type: "string", Description: "Ключевое слово по которому нужно получить информацию"},
							"lang":    {Type: "string", Description: "Язык результата", Enum: []string{"en", "ru"}},
						}),
					},
				},
			},
		}
	}
	req := &ollama.ChatRequest{
		Model:     c.model.Name,
		Messages:  messages,
		Stream:    &f,
		KeepAlive: &ollama.Duration{Duration: time.Hour * 12},
		Tools:     t,
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

func (c *conversation) SetModel(name string) error {
	c.model.Name = name
	conv, err := c.bot.storage.GetConversation(c.id)
	if err != nil {
		return err
	}
	conv.Model = name
	return c.bot.storage.SaveConversation(c.id, conv)
}

func (c *conversation) SendSelectModel() {
	msg := tgbotapi.NewMessage(c.id,
		c.translator.Sprintf("Текущая модель %s. Сменить:", c.model.Name))
	msg.ReplyMarkup = c.GenerateModelKeyboard()
	if _, err := c.bot.tgClient.Send(msg); err != nil {
		log.Error().Err(err).Msg("convarsation: problem send message")
	}
}

func (c *conversation) GenerateModelKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var buttons []tgbotapi.InlineKeyboardButton
	i := 0
	modelsCount := len(c.bot.cfg.Models)
	for name := range c.bot.cfg.Models {
		i++
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(name, name))

		if i%3 == 0 || i == modelsCount {
			rows = append(rows, buttons)
			buttons = nil
		}
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		rows...,
	)
}
