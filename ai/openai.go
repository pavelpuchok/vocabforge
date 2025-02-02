package ai

import "github.com/sashabaranov/go-openai"

type OpenAI struct {
	client *openai.Client
}

func NewOpenAI(client *openai.Client) *OpenAI {
	return &OpenAI{
		client: client,
	}
}
