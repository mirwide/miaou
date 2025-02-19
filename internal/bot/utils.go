package bot

import (
	"strings"

	"golang.org/x/text/language"
)

func EscapeMarkdown(text string) string {

	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
		"\\(", ")", "\\)", "~", "\\~", ">", "\\>", "`", "\\`",
		"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
		"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
	)
	text = replacer.Replace(text)
	replacer = strings.NewReplacer(
		"\\`\\`\\`", "```", "\\*\\*", "**", "\\~\\~", "~~",
		"<think\\>", "```", "</think\\>", "```",
	)
	return replacer.Replace(text)
}

func ParseLang(lang string) language.Tag {
	switch lang {
	case "ru", "kz", "ua":
		return language.Russian
	case "es":
		return language.Spanish
	default:
		return language.English
	}
}
