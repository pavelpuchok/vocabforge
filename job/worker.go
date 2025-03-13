package job

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/pavelpuchok/vocabforge/ai"
	"github.com/pavelpuchok/vocabforge/db/sqlc"
	"github.com/pavelpuchok/vocabforge/preply"
)

type TranslateWordsJobWorker struct {
	Queries *sqlc.Queries
	Queue   Queue[TranslateWordsJob]
	Logger  *slog.Logger
	OpenAI  *ai.OpenAI
}

func (w TranslateWordsJobWorker) Run(ctx context.Context) {
	for {
		job, err := w.Queue.Dequeue(ctx)
		if err != nil {
			//TODO: enqeue failed job
			if !errors.Is(err, ErrJobNotFound) {
				w.Logger.Error("unexpected dequeu job error", slog.String("error", err.Error()))
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(10 * time.Second)
		}

		// TODO: add transactions
		w.processTick(ctx, job)
	}
}

func (w TranslateWordsJobWorker) processTick(ctx context.Context, job TranslateWordsJob) error {
	newWords, err := w.filterWords(ctx, job)
	if err != nil {
		return err
	}

	for wordsChunk := range slices.Chunk(newWords, 5) {
		w.translateWords(ctx, wordsChunk, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w TranslateWordsJobWorker) filterWords(ctx context.Context, job TranslateWordsJob) ([]preply.Node, error) {
	ids := make([]sql.NullString, 0, len(job.Words))
	for _, w := range job.Words {
		ids = append(ids, sql.NullString{
			String: w.ID,
			Valid:  true,
		})
	}

	res, err := w.Queries.ListExistingUserWordsByPreplyIDs(ctx, sqlc.ListExistingUserWordsByPreplyIDsParams{
		UserID:    job.UserID,
		PreplyIds: ids,
	})

	if err != nil {
		return nil, err
	}

	existing := make(map[string]bool, len(res))

	for _, w := range res {
		if !w.Valid {
			continue
		}
		existing[w.String] = true
	}

	filtered := make([]preply.Node, 0, len(job.Words))
	for _, w := range job.Words {
		if existing[w.ID] {
			continue
		}

		filtered = append(filtered, w)
	}

	return filtered, nil
}

func (w TranslateWordsJobWorker) translateWords(ctx context.Context, wordsChunk []preply.Node, job TranslateWordsJob) error {
	defs := make([]ai.WordWithDefinition, len(wordsChunk))
	for i, w := range wordsChunk {
		defs[i] = ai.WordWithDefinition{
			Definition: w.Definition,
			Spelling:   w.Spelling,
		}
	}

	translatedWords, err := w.OpenAI.RequestTranslation(ctx, defs, job.TargetLanguage)
	if err != nil {
		w.Logger.Warn("unexpected translation error. Retry one more time", slog.String("error", err.Error()))
		translatedWords, err = w.OpenAI.RequestTranslation(ctx, defs, job.TargetLanguage)
		if err != nil {
			w.Logger.Warn("unexpected translation error %s after retry", slog.String("error", err.Error()))
			return err
		}
	}

	for i, node := range wordsChunk {
		_, err = w.Queries.AddWord(ctx, sqlc.AddWordParams{
			PreplyID: sql.NullString{
				String: node.ID,
				Valid:  true,
			},
			Spelling:        node.Spelling,
			Definition:      node.Definition,
			LexicalCategory: node.LexicalCategory,
			TranslationRu:   translatedWords[i],
			Lang:            "en",
			UserID:          job.UserID,
			AddedAt:         time.Now(),
		})

		if err != nil {
			if err.Error() != "constraint failed: UNIQUE constraint failed: vocab_words.user_id, vocab_words.preply_id (2067)" {
				w.Logger.Error("failed to add a word", slog.String("error", err.Error()))
				return err
			}

			w.Logger.Warn("failed to add a word because it's already added. Possible there is a concurrent process that add words too", slog.Int64("UserID", job.UserID))
			continue
		}

		w.Logger.Debug("added new word", slog.Int64("UserID", job.UserID))
	}

	return nil
}
