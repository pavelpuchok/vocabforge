package vocabulary

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
	"github.com/pavelpuchok/vocabforge/vocabulary/sentences"
)

type Service struct {
	repository            Repository
	sentences             SentencesGenerator
	defaultSentencesCount int
}

type SentencesGenerator interface {
	Generate(ctx context.Context, spell, definition, lexicalCategory string, sentencesCount int) ([]sentences.Sentence, error)
}

func NewService(repo Repository, sentences SentencesGenerator, sentencesCount int) Service {
	return Service{
		repo,
		sentences,
		sentencesCount,
	}
}

type Repository interface {
	AddWord(ctx context.Context, userID models.UserID, spell, definition, lexicalCategory string, lang models.Language, exercises []string) (models.Word, error)
}

func (s Service) AddWord(ctx context.Context, userID models.UserID, spell, definition, lexicalCategory string, lang models.Language, exercises []string) (models.Word, error) {
	if len(exercises) == 0 {
		sentences, err := s.sentences.Generate(ctx, spell, definition, lexicalCategory, s.defaultSentencesCount)
		if err != nil {
			return models.Word{}, fmt.Errorf("vocabulary.Service.AddWord unable to generate exercises. %w", err)
		}

		exercises = make([]string, len(sentences))

		for i, ss := range sentences {
			exercises[i] = ss.Text
		}
	}

	word, err := s.repository.AddWord(ctx, userID, spell, definition, lexicalCategory, lang, exercises)
	if err != nil {
		return word, fmt.Errorf("vocabulary.Service.AddWord unable to add word. %w", err)
	}
	return word, nil
}
