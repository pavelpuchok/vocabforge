package models

import (
	"fmt"
	"time"
)

type WordID string

func (id WordID) String() string {
	return string(id)
}

type LearnStatus int

func (s *LearnStatus) String() string {
	txt, err := s.MarshalText()
	if err != nil {
		return "unknown"
	}
	return txt
}

func (s *LearnStatus) MarshalText() (string, error) {
	switch *s {
	case Pending:
		return "pending", nil
	case InProgress:
		return "in_progress", nil
	case Learned:
		return "learned", nil
	default:
		return "", fmt.Errorf("%d is unknown LearnStatus", *s)
	}
}

func (s *LearnStatus) UnmarshalText(text string) error {
	switch text {
	case "pending":
		*s = Pending
	case "in_progress":
		*s = InProgress
	case "learned":
		*s = Learned
	default:
		return fmt.Errorf("%s is unknown LearnStatus representation", text)
	}

	return nil
}

const (
	Pending LearnStatus = iota
	InProgress
	Learned
)

type Word struct {
	ID              WordID
	UserID          UserID
	Spelling        string
	Definition      string
	LexicalCategory string
	Language        Language
	LearnStatus     LearnStatus
	AnsweredCount   uint
	AddedAt         time.Time
	LastAskedAt     time.Time
}
