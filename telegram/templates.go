package telegram

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pavelpuchok/vocabforge/db/sqlc"
)

const (
	wordDescriptionTemplateName      = "wordDescription"
	incorrectAnswerTemplateName      = "incorrectAnswer"
	correctAnswerTemplateName        = "correctAnswer"
	correctAnswerLearnedTemplateName = "correctAnswerLearned"
)

const templatesText = `
{{define "wordDescription"}}*{{escapeText .Spelling}}* _{{escapeText .LexicalCategory}}_

{{escapeText .Definition}}

RU: _{{escapeText .TranslationRu}}_
{{end}}

{{define "incorrectAnswer"}}{{escapeText "Oh, no! It's incorrect answer."}}

{{escapeText "Correct answer: "}}{{escapeText .ExpectedAnswer}}

{{escapeText "Word: "}}{{template "wordDescription" .Word}}

{{escapeText "Keep learning /learn_vocab"}}
{{end}}

{{define "correctAnswer"}}{{ escapeText "Good! You've correctly answered this word for "}}{{ .AnsweredCount }}{{escapeText " times in a row. Keep going! Learn next word /learn_vocab" }}{{end}}

{{define "correctAnswerLearned"}}{{ escapeText "Excelent! You've learned this word! Learn next word /learn_vocab" }}{{end}}
`

var templates *template.Template

func init() {
	templates = template.New("base").Funcs(template.FuncMap{
		"escapeText": escapeText,
	})
	template.Must(templates.Parse(templatesText))
}

func MustRenderIncorrectAnswer(word sqlc.VocabWord, expectedAnswer string) string {
	return mustRenderTemplate(incorrectAnswerTemplateName, struct {
		Word           sqlc.VocabWord
		ExpectedAnswer string
	}{
		Word:           word,
		ExpectedAnswer: expectedAnswer,
	})
}

func MustRenderCorrectAnswer(answeredCount int) string {
	return mustRenderTemplate(correctAnswerTemplateName, struct{ AnsweredCount int }{answeredCount})
}

func MustRenderCorrectAnswerLearned() string {
	return mustRenderTemplate(correctAnswerLearnedTemplateName, struct{}{})
}

func mustRenderTemplate(name string, data any) string {
	sb := &strings.Builder{}
	err := templates.ExecuteTemplate(sb, name, data)
	if err != nil {
		panic(fmt.Sprintf("unable to execute template %s. %s", name, err))
	}
	return sb.String()
}
