package sentences

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAIPromptProvider_Prompt(t *testing.T) {
	t.Parallel()

	p, err := NewAIPromptProvider()
	if err != nil {
		t.Fatal(err)
	}

	actual, err := p.Prompt("foo", "bar", "adverb", 123)
	if err != nil {
		t.Error(err)
	}

	expected := `Generate 123 exercises for learning the word 'foo'.
Word: 'foo'. Definition: 'bar'. Lexical Category: adverb.

Instructions:
- Each sentence should use the word 'foo'.
- Format each sentence with the word 'foo' prefixed with <% and postfixed with %>.
- Ensure the sentences are varied and cover different tenses if applicable.`

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected prompt (-want +got):\n%s", diff)
	}
}
