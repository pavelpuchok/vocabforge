package sentences

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AIGenerator struct {
	client         *openai.Client
	promptProvider PromptProvider
	schema         *jsonschema.Definition
}

type aiResponse struct {
	Sentences []Sentence `json:"sentences"`
}

type Sentence struct {
	Text string `json:"text"`
}

type PromptProvider interface {
	Prompt(spelling, definition, lexicalCategory string, sentencesCount int) (string, error)
}

func NewAIGenerator(apiToken string, promptProvider PromptProvider) (AIGenerator, error) {
	// generate response schema
	schema, err := jsonschema.GenerateSchemaForType(aiResponse{})
	if err != nil {
		return AIGenerator{}, fmt.Errorf("sentences.NewAIGenerator unable to generate response schema. %w", err)
	}

	client := openai.NewClient(apiToken)

	return AIGenerator{
		client:         client,
		schema:         schema,
		promptProvider: promptProvider,
	}, nil
}

func (g AIGenerator) Generate(ctx context.Context, spelling, definition, lexicalCategory string, sentencesCount int) ([]Sentence, error) {
	prompt, err := g.promptProvider.Prompt(spelling, definition, lexicalCategory, sentencesCount)
	if err != nil {
		return nil, fmt.Errorf("sentences.AIGenerator.Generate unable to generate promt. %w", err)
	}

	response, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "word_learning",
				Schema: g.schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sentences.AIGenerator.Generate unable to make ChatGPT request. %w", err)
	}

	var result aiResponse
	err = g.schema.Unmarshal(response.Choices[0].Message.Content, &result)
	if err != nil {
		return nil, fmt.Errorf("sentences.AIGenerator.Generate unable to unmarshal response. %w", err)
	}

	return result.Sentences, nil
}
