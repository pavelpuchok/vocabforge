package addword

import (
	"context"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
)

type UseCase struct {
	VocabularyService Service
}

type Service interface {
	AddWord(ctx context.Context, userID models.UserID, spell, definition, lang string, exercises []models.SentenceExercise) (models.Word, error)
}

func (u UseCase) Run(ctx context.Context, userID models.UserID, spell, definition, lang string) (models.Word, error) {
	word, err := u.VocabularyService.AddWord(ctx, userID, spell, definition, lang, nil)
	if err != nil {
		return word, fmt.Errorf("addword.UseCase.Run unable to add word. %w", err)
	}
	return word, nil
}
