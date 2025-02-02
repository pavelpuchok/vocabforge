package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatTgMessage(t *testing.T) {
	assert.Equal(t, "abc\\.\\.\\. def\\!\\!\\!", formatMessage("abc%s def%s", "...", "!!!"))
}
