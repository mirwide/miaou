package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mirwide/miaou/internal/bot/msg"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type conversation struct {
	id         int64
	bot        *Bot
	translator *message.Printer
}

func NewConversation(chatID int64, bot *Bot, lang string) *conversation {
	var l string
	switch lang {
	case "ru", "kz", "ua":
		l = "ru-RU"
	default:
		l = "en-US"
	}
	translator := message.NewPrinter(language.MustParse(l))
	return &conversation{
		id:         chatID,
		bot:        bot,
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
