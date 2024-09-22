package vocabulary

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
)

type Service struct {
	repository Repository
}

func NewService(repo Repository) Service {
	return Service{
		repo,
	}
}

type Repository interface {
	AddWord(ctx context.Context, userID models.UserID, spell, definition, lexicalCategory string, lang models.Language, exercises []models.SentenceExercise) (models.Word, error)
}

func (s Service) AddWord(ctx context.Context, userID models.UserID, spell, definition, lexicalCategory string, lang models.Language, exercises []models.SentenceExercise) (models.Word, error) {
	word, err := s.repository.AddWord(ctx, userID, spell, definition, lexicalCategory, lang, exercises)
	if err != nil {
		return word, fmt.Errorf("vocabulary.Service.AddWord unable to add word. %w", err)
	}
	return word, nil
}
