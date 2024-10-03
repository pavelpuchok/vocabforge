package models

import "fmt"

type SentenceExerciseID string

func SentenceExerciseIDFromText(s string) (SentenceExerciseID, error) {
	if len(s) != objectIdHexLen {
		return "", fmt.Errorf("models.SentenceExerciseIDFromText invalid sentence exercise ID string %s", s)
	}

	return SentenceExerciseID(s), nil
}

type SentenceExercise struct {
	ID       SentenceExerciseID
	Word     Word
	Sentence string
	Answered bool
}
