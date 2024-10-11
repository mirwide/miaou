package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type conversation struct {
	bot *Bot
	id  int64
}

func NewConversation(chatID int64, bot *Bot) *conversation {
	return &conversation{
		bot: bot,
		id:  chatID,
	}
}

func (c *conversation) SendServiceMessage(message string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.id, c.bot.translator.Sprintf(message))
	return c.bot.tgClient.Send(msg)
}
