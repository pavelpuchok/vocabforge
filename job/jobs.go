package job

import (
	"github.com/pavelpuchok/vocabforge/preply"
)

type TranslateWordsJob struct {
	UserID         int64
	Words          []preply.Node
	TargetLanguage string
}
