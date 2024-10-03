package getexercise

import (
	"context"
	"errors"
	"fmt"

	"github.com/pavelpuchok/vocabforge/models"
)

var ErrNoWordsToLearn = errors.New("no words to learn")

type UseCase struct {
	Exercises ExercisesService
}

type ExercisesService interface {
	StatsByUser(ctx context.Context, userId models.UserID) (struct {
		Pending    uint
		InProgress uint
	}, error)
	OldestByUser(ctx context.Context, userId models.UserID, status models.LearnStatus) (models.SentenceExercise, error)
}

const maxLearningWords = 64

// TODO: fix return tyeps
func (u UseCase) Run(ctx context.Context, userID models.UserID) (models.SentenceExercise, error) {
	s, err := u.Exercises.StatsByUser(ctx, userID)
	if err != nil {
		return models.SentenceExercise{}, fmt.Errorf("getexercise.UseCase.Run unable to get user's vocab stats. %w", err)
	}

	if s.InProgress == 0 && s.Pending == 0 {
		return models.SentenceExercise{}, ErrNoWordsToLearn
	}

	// check if we can start learning new words
	if s.InProgress < maxLearningWords && s.Pending > 0 {
		// We can start learning new words
		// Find word in Pending state and take any sentence for it
		ex, err := u.Exercises.OldestByUser(ctx, userID, models.Pending)
		if err != nil {
			return models.SentenceExercise{}, fmt.Errorf("getexercise.UseCase.Run unable to get exercise for new word. %w", err)
		}
		return ex, nil
	}
	// Return sentences from words in InProgress state
	ex, err := u.Exercises.OldestByUser(ctx, userID, models.InProgress)
	if err != nil {
		return models.SentenceExercise{}, fmt.Errorf("getexercise.UseCase.Run unable to get exercise for learning words. %w", err)
	}
	return ex, nil
}
