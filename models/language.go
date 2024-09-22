package models

import "fmt"

type Language string

func (l *Language) String() string {
	s, err := l.MarshalText()
	if err != nil {
		return fmt.Sprintf("unknown(%s)", *l)
	}
	return s
}

func (l *Language) MarshalText() (string, error) {
	return string(*l), nil
}

func (l *Language) UnmarshalText(s string) error {
	*l = Language(s)
	return nil
}

func LanguageFromText(s string) (Language, error) {
	var lang Language
	err := lang.UnmarshalText(s)
	if err != nil {
		return "", fmt.Errorf("models.LanguageFromText invalid language string %s. %w", s, err)
	}
	return lang, nil
}
