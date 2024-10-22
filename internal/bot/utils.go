package bot

import "strings"

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
	)
	return replacer.Replace(text)
}
