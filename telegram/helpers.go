package telegram

import (
	"fmt"
	"strings"
)

func formatMessage(tpl string, args ...any) string {
	return escapeText(fmt.Sprintf(tpl, args...))
}

func escapeText(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
		"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
		"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
		"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		"\\", "\\\\",
	)
	return replacer.Replace(text)
}
