package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ollama/ollama/api"
	ollama "github.com/ollama/ollama/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/message"

	"github.com/mirwide/miaou/internal/bot/msg"
	"github.com/mirwide/miaou/internal/config"
	"github.com/mirwide/miaou/internal/storage"
	"github.com/mirwide/miaou/internal/tools"
	"github.com/mirwide/miaou/internal/tools/wiki"
	_ "github.com/mirwide/miaou/internal/translations"
)

type conversation struct {
	id         int64
	bot        *Bot
	model      config.Model
	translator *message.Printer
	log        zerolog.Logger
	toolsLimit int
	ready      chan bool
}

func NewConversation(chatID int64, lang string, bot *Bot) *conversation {
	var m config.Model
	c, _ := bot.storage.GetConversation(chatID)
	log := log.With().Int64("conversation", chatID).Logger()
	m, ok := bot.cfg.Models[c.Model]
	if !ok {
		log.Info().Msgf("model %s not found use default", c.Model)
		m = bot.cfg.Models[bot.cfg.DefaultModel]
	}

	translator := message.NewPrinter(ParseLang(lang))
	_ = bot.storage.SaveConversation(chatID, storage.Conversation{Model: m.Name})

	return &conversation{
		id:         chatID,
		bot:        bot,
		model:      m,
		translator: translator,
		log:        log,
		toolsLimit: 5,
		ready:      make(chan bool),
	}
}

func (c *conversation) SendServiceMessage(message string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(c.id, c.translator.Sprintf(message))
	result, err := c.bot.tgClient.Send(msg)
	if err != nil {
		c.log.Error().Err(err).Msg("convarsation: problem send service message")
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
	c.log.Debug().Any("ollama", resp).Msg("bot: ollama response")
	c.bot.storage.SaveMessage(c.id, resp.Message)
	if len(resp.Message.ToolCalls) > 0 && c.toolsLimit > 0 {
		c.toolsLimit--
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
			if call.Function.Arguments["callback"] != nil && c.model.Callback {
				c.bot.storage.SaveMessage(c.id, ollama.Message{
					Role:    "tool",
					Content: fmt.Sprintf("%v", call.Function.Arguments["callback"]),
				})
				log.Debug().Interface("data", call.Function.Arguments["callback"]).Msg("callback")
			}
		}
		c.SendOllama()
	} else {
		text := EscapeMarkdown(resp.Message.Content)
		if len(text) > 4096 {
			text = text[:4096]
			c.SendServiceMessage(msg.MsgToLong)
		}
		msg := tgbotapi.NewMessage(c.id, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = c.bot.tgClient.Send(msg)
		if err != nil && strings.Contains(err.Error(), "is reserved and must be escaped") {
			msg := tgbotapi.NewMessage(c.id, resp.Message.Content)
			_, err = c.bot.tgClient.Send(msg)
		}
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
					Parameters: tools.Parameters{
						Type:     "object",
						Required: []string{"callback"},
						Properties: tools.NewProperties(map[string]tools.Properties{
							"callback": {Type: api.PropertyType{"string"}, Description: "Опиши подробно на русском что ожидаешь получить"},
						})},
				},
			},
			ollama.Tool{
				Type: "function",
				Function: ollama.ToolFunction{
					Name:        "get_weather",
					Description: "Получить текущую погоду по городу",
					Parameters: tools.Parameters{
						Type:     "object",
						Required: []string{"city", "callback"},
						Properties: tools.NewProperties(map[string]tools.Properties{
							"city": {Type: api.PropertyType{"string"}, Description: "Название города в транслите"},
							"forecast_days": {Type: api.PropertyType{"int"},
								Description: "Количество дней прогноза, должно быть равно 1 если требуется прогноз на сегодняшний день. Максимальное значение 16"},
							"callback": {Type: api.PropertyType{"string"}, Description: "Опиши подробно на русском что ожидаешь получить"},
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
						Required: []string{"keyword", "lang", "callback"},
						Properties: tools.NewProperties(map[string]tools.Properties{
							"keyword":  {Type: api.PropertyType{"string"}, Description: "Ключевое слово по которому нужно получить информацию"},
							"lang":     {Type: api.PropertyType{"string"}, Description: "Язык результата", Enum: []any{"en", "ru"}},
							"callback": {Type: api.PropertyType{"string"}, Description: "Опиши подробно на русском что ожидаешь получить"},
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
	c.bot.wg.Add(1)
	go func() {
		defer c.bot.wg.Done()
		ctx := context.Background()
		err := c.bot.ollama.Chat(ctx, req, c.OllamaCallback)
		if err != nil {
			c.log.Error().Err(err).Msg("bot: problem get response from llm chat")
			c.SendServiceMessage(msg.ErrorOccurred)
		}
		c.ready <- true
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
		c.translator.Sprintf("Текущая модель: %s. Выберите другую:", c.model.Name))
	msg.ReplyMarkup = c.GenerateModelKeyboard()
	if _, err := c.bot.tgClient.Send(msg); err != nil {
		c.log.Error().Err(err).Msg("convarsation: problem send message")
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

func (c *conversation) SetChatCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: reset.command, Description: c.translator.Sprintf(reset.desc)},
		{Command: model.command, Description: c.translator.Sprintf(model.desc)},
	}
	scope := tgbotapi.NewBotCommandScopeChat(c.id)
	config := tgbotapi.SetMyCommandsConfig{
		Commands: commands,
		Scope:    &scope,
	}

	_, err := c.bot.tgClient.Request(config)
	if err != nil {
		log.Error().Err(err).Msg("problem set bot commands")
	}
	return err
}
