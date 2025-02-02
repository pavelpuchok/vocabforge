package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type requestExerciseResponse struct {
	Exercise Exercise `json:"exercise"`
}

type Exercise struct {
	Sentence    string `json:"sentence"`
	MissingWord string `json:"missingWord"`
}

const placeholder = "<%placeholder%>"

const requestExerciseSystemMessage = `Objective: Generate a sentence using a given word (in its correct tense and case) based on its meaning. The sentence should align with the specified language proficiency level, providing an appropriate challenge for the learner. The sentence must be clear enough to help the learner infer the missing word.

Input Format:
1. A word (spelling and definition) provided separately.
2. A specified language proficiency level (e.g., A2, B1, C1, etc.).

Output Format:
1. A sentence where the given word is replaced with <%placeholder%>.
2. A given word in correct tense and case

Instructions:
1. Use the provided word meaning to create a meaningful sentence.
2. Ensure the sentence corresponds to the given language level and challenges the learner accordingly.
3. Make the sentence clear enough to help the learner deduce the missing word.
4. Adapt the tense and grammatical case of the word if necessary to match the sentence context.
5. Replace the word with <%placeholder%> in the final output.`

var requestExerciseSchema *jsonschema.Definition

func init() {
	var err error
	requestExerciseSchema, err = jsonschema.GenerateSchemaForType(requestExerciseResponse{})
	if err != nil {
		panic(fmt.Sprintf("unable to generate json schema for OpenAI request exercise response. %s", err))
	}
}

func (o *OpenAI) RequestExercise(ctx context.Context, w WordWithDefinition, profLevel string) (Exercise, error) {
	res, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: requestExerciseSystemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: generateRequestExercisePrompt(w, profLevel),
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "request_exercise",
				Schema: requestExerciseSchema,
				Strict: true,
			},
		},
	})

	if err != nil {
		return Exercise{}, fmt.Errorf("RequestExercise failed %w", err)
	}

	var result requestExerciseResponse
	err = requestExerciseSchema.Unmarshal(res.Choices[0].Message.Content, &result)
	if err != nil {
		return Exercise{}, fmt.Errorf("RequestExercise unable to parse response. %w", err)
	}

	result.Exercise.Sentence = strings.ReplaceAll(result.Exercise.Sentence, placeholder, "???")

	return result.Exercise, nil
}

func generateRequestExercisePrompt(w WordWithDefinition, languageLevel string) string {
	return fmt.Sprintf(`Spelling: "%s"
Definition: "%s"
Proficiency: "%s"
`, w.Spelling, w.Definition, languageLevel)
}
