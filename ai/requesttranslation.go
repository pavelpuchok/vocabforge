package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type WordWithDefinition struct {
	Definition string
	Spelling   string
}

const requestTranslationSystemMessage = `Objective: Translate the provided list of words with spelling and definitions into the target language while preserving the order.

Input Format:
A list of words, each with spelling and definition.
Target language code (e.g., 'ru' for Russian, 'es' for Spanish).

Output Format:
A list of translated words in the same order.
Each item should contain the translated spelling.

Instructions:
1. Translate each word and definition accurately.
2. Maintain the original order of the list.
3. Ensure the translated definitions preserve the intended meaning.
4. Use formal and commonly accepted translations.`

var requestTranslationSchema *jsonschema.Definition

type requestTranslationResponse struct {
	TranslatedWords []struct {
		Spelling string `json:"spelling"`
	} `json:"translatedWords"`
}

func init() {
	var err error
	requestTranslationSchema, err = jsonschema.GenerateSchemaForType(requestTranslationResponse{})
	if err != nil {
		panic(fmt.Sprintf("unable to generate json schema for OpenAI request translation response. %s", err))
	}
}

func (o *OpenAI) RequestTranslation(ctx context.Context, words []WordWithDefinition, targetLanguage string) ([]string, error) {
	res, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: requestTranslationSystemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: generateRequestTranslationPrompt(words, targetLanguage),
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "request_translation",
				Schema: requestTranslationSchema,
				Strict: true,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("RequestTranslation failed %w", err)
	}

	var result requestTranslationResponse
	err = requestTranslationSchema.Unmarshal(res.Choices[0].Message.Content, &result)
	if err != nil {
		return nil, fmt.Errorf("RequestTranslation unable to parse response. %w", err)
	}

	if len(result.TranslatedWords) != len(words) {
		return nil, fmt.Errorf("unexpected length of translated words list. Expected %d, Got %d", len(words), len(result.TranslatedWords))
	}

	translatedSpellings := make([]string, len(result.TranslatedWords))
	for i, w := range result.TranslatedWords {
		translatedSpellings[i] = w.Spelling
	}

	return translatedSpellings, nil
}

func generateRequestTranslationPrompt(words []WordWithDefinition, targetLanguage string) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("Target Language Code: \"%s\"\n", targetLanguage))
	b.WriteString("List of words:\n")

	for i, w := range words {
		b.WriteString(fmt.Sprintf("%d. Spelling: \"%s\", Definition: \"%s\"\n", i+1, w.Spelling, w.Definition))
	}

	return b.String()
}
