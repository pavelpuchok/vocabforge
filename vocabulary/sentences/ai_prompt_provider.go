package sentences

import (
	"fmt"
	"strings"
	"text/template"
)

const promptTemplateText = `Generate {{.SentencesCount}} exercises for learning the word '{{.Spelling}}'.
Word: '{{.Spelling}}'. Definition: '{{.Definition}}'. Lexical Category: {{.LexicalCategory}}.

Instructions:
- Each sentence should use the word '{{.Spelling}}'.
- Format each sentence with the word '{{.Spelling}}' prefixed with <% and postfixed with %>.
- Ensure the sentences are varied and cover different tenses if applicable.`

type promptTemplateCtx struct {
	SentencesCount  int
	Spelling        string
	Definition      string
	LexicalCategory string
}

type AIPromptProvider struct {
	tpl *template.Template
}

func NewAIPromptProvider() (AIPromptProvider, error) {
	tpl, err := template.New("AIPromptProviderTemplate").Parse(promptTemplateText)
	if err != nil {
		return AIPromptProvider{}, fmt.Errorf("sentences.NewAIPromptProvider unable to create prompt template. %w", err)
	}

	return AIPromptProvider{
		tpl: tpl,
	}, nil
}

func (p AIPromptProvider) Prompt(spelling, definition, lexicalCategory string, sentencesCount int) (string, error) {
	sb := strings.Builder{}
	err := p.tpl.Execute(&sb, promptTemplateCtx{
		SentencesCount:  sentencesCount,
		Spelling:        spelling,
		Definition:      definition,
		LexicalCategory: lexicalCategory,
	})
	if err != nil {
		return "", fmt.Errorf("vocabulary.AIPromptProvider.Prompt unable to render template. %w", err)
	}
	return sb.String(), nil
}
