package bot_test

import (
	"testing"

	"github.com/mirwide/miaou/internal/bot"
	"github.com/stretchr/testify/assert"
)

func TestEscapeMarkdown(t *testing.T) {
	source := `Текст для проверки экранирования символов для telegram:
	* не поддерживаемые списки
	~ одиночная тильда
	! восклицательный знак
	**жирный**
	~~курсив~~
	`
	result := `Текст для проверки экранирования символов для telegram:
	\* не поддерживаемые списки
	\~ одиночная тильда
	\! восклицательный знак
	**жирный**
	~~курсив~~
	`
	actual := bot.EscapeMarkdown(source)
	assert.Equal(t, result, actual)

	source = "```yaml\nкод```"
	result = "```yaml\nкод```"
	actual = bot.EscapeMarkdown(source)
	assert.Equal(t, result, actual)

	source = "` одиночный гравис"
	result = "\\` одиночный гравис"
	actual = bot.EscapeMarkdown(source)
	assert.Equal(t, result, actual)
}
