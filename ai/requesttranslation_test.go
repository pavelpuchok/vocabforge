package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateRequestTranslationPrompt(t *testing.T) {
	expected := `Target Language Code: "ru"
List of words:
1. Spelling: "foo", Definition: "foo def"
2. Spelling: "bar", Definition: "bar def"
3. Spelling: "baz", Definition: "baz def"
`

	actual := generateRequestTranslationPrompt([]WordWithDefinition{
		{
			Spelling:   "foo",
			Definition: "foo def",
		},
		{
			Spelling:   "bar",
			Definition: "bar def",
		},
		{
			Spelling:   "baz",
			Definition: "baz def",
		},
	}, "ru")

	assert.Equal(t, expected, actual)
}
