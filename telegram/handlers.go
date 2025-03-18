package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pavelpuchok/vocabforge/ai"
	"github.com/pavelpuchok/vocabforge/db/sqlc"
	"github.com/pavelpuchok/vocabforge/job"
	"github.com/pavelpuchok/vocabforge/preply"
	"gopkg.in/telebot.v4"
)

type Handlers struct {
	RootCtx context.Context
	OpenAI  *ai.OpenAI
	Bot     *telebot.Bot
}

type NonAuthorizedHandler func(telebot.Context, *sqlc.Queries) error
type AuthorizedHandler func(telebot.Context, *sqlc.Queries, sqlc.User) error

func (h Handlers) Register() {
	h.registerNonAuthorized("/start", h.handleStart)
	h.registerAuthorized("/preply_sync", h.handlePreplySync)
	h.registerAuthorized("/learn_vocab", h.handleLearnVocab)
	h.registerAuthorized(telebot.OnCallback, h.handleCallback)
	h.registerAuthorized(telebot.OnReply, h.handleReply)
}

func (h Handlers) registerNonAuthorized(cmd string, handler NonAuthorizedHandler) {
	h.Bot.Handle(cmd, func(ctx telebot.Context) error {
		db, err := GetSQLTx(ctx)
		if err != nil {
			return fmt.Errorf("%s failed to get SQLTx. %w", cmd, err)
		}

		queries := sqlc.New(db)

		if err := handler(ctx, queries); err != nil {
			return fmt.Errorf("%s handler failed. %w", cmd, err)
		}

		return nil
	})
}

func (h Handlers) registerAuthorized(cmd string, handler AuthorizedHandler) {
	h.Bot.Handle(cmd, func(ctx telebot.Context) error {
		db, err := GetSQLTx(ctx)
		if err != nil {
			return fmt.Errorf("%s failed to get SQLTx. %w", cmd, err)
		}

		queries := sqlc.New(db)

		user, err := queries.GetUserByTelegramID(h.RootCtx, sqlc.GetUserByTelegramIDParams{
			TelegramID:     ctx.Sender().ID,
			TelegramChatID: ctx.Chat().ID,
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ctx.Reply(formatMessage("Not authorized. Try /start command first"))
			}

			return fmt.Errorf("%s failed to get user. %w", cmd, err)
		}

		if err := handler(ctx, queries, user); err != nil {
			return fmt.Errorf("%s handler failed. %w", cmd, err)
		}

		return nil
	})
}

func (h Handlers) handleStart(ctx telebot.Context, queries *sqlc.Queries) error {
	usr, err := queries.CreateOrUpdateUser(h.RootCtx, sqlc.CreateOrUpdateUserParams{
		TelegramID:     ctx.Sender().ID,
		TelegramChatID: ctx.Chat().ID,
		Name:           ctx.Sender().FirstName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create an user. %w", err)
	}

	err = ctx.Reply(formatMessage("Hello %s!", usr.Name))
	if err != nil {
		return fmt.Errorf("failed to send response. %w", err)
	}

	return nil
}

func (h Handlers) handlePreplySync(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	cookies := strings.Join(ctx.Args(), " ")

	fetcher := preply.VocabFetcher{}

	q := job.Queue[job.TranslateWordsJob]{
		GroupName: "job.TranslateWordsJob",
		Storage:   job.DBStorage{Queries: queries},
	}

	limit := 100
	offset := 0
	count := 0
	for {
		v, err := fetcher.Fetch(cookies, limit, offset)
		if err != nil {
			ctx.Reply(fmt.Sprintf("Fetch failed. %s", err))
			return fmt.Errorf("failed to fetch vocab. %w", err)
		}

		err = q.Enqueue(h.RootCtx, job.TranslateWordsJob{
			Words:          v.Words.Nodes,
			UserID:         user.ID,
			TargetLanguage: "ru",
		})

		if err != nil {
			ctx.Reply(fmt.Sprintf("Unable to enqueu words translation. %s", err))
			return fmt.Errorf("failed to enqueue translation. %w", err)
		}

		count += len(v.Words.Nodes)

		offset += limit
		if offset >= v.Words.TotalCount {
			break
		}
	}

	return ctx.Reply(formatMessage("%d new words were added to the translation queue.", count))
}

func (h Handlers) handleLearnVocab(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	learningCount, err := queries.CountLearningWords(h.RootCtx, user.ID)
	if err != nil {
		return fmt.Errorf("unable to count learning words. %w", err)
	}

	var w sqlc.VocabWord
	var wordFound bool
	if learningCount < 16 {
		wordFound = true
		w, err = queries.GetOldestUnseenWord(h.RootCtx, user.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unable to count learning words. %w", err)
			}
			wordFound = false
		}
	}

	if !wordFound {
		w, err = queries.GetLearningWord(h.RootCtx, user.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ctx.Reply(formatMessage("No new words to learn."))
			}

			ctx.Reply(formatMessage("Failed to find word %s", err))
			return fmt.Errorf("unable to get word. %w", err)
		}
	}

	err = queries.IncrementWordViewedCountByID(h.RootCtx, sqlc.IncrementWordViewedCountByIDParams{
		LastShowedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:           w.ID,
	})
	if err != nil {
		ctx.Reply(formatMessage("Something failed"))
		return fmt.Errorf("unable to update word views count. %w", err)
	}

	ex, err := h.OpenAI.RequestExercise(h.RootCtx, ai.WordWithDefinition{
		Spelling:   w.Spelling,
		Definition: w.Definition,
	}, "C1")
	if err != nil {
		ctx.Reply(formatMessage("Failed to generate exercise. %s", err))
		return fmt.Errorf("unable to generate exercise. %w", err)
	}

	msg, err := h.Bot.Reply(ctx.Message(), formatMessage("%s", ex.Sentence), &telebot.ReplyMarkup{
		ForceReply: true,
		Selective:  true,
	})
	if err != nil {
		ctx.Reply(fmt.Sprintf("Failed to generate exercise. %s", err))
		return fmt.Errorf("unable to generate exercise. %w", err)
	}

	_, err = queries.CreateExercise(h.RootCtx, sqlc.CreateExerciseParams{
		WordID:        w.ID,
		UserID:        user.ID,
		Question:      ex.Sentence,
		Answer:        ex.MissingWord,
		TelegramMsgID: int64(msg.ID),
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return fmt.Errorf("unable to create exercise record. %w", err)
	}

	return nil
}

func (h Handlers) handleCallback(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	fmt.Printf("callback: %s\n", ctx.Callback().MessageID)
	return nil
}

func (h Handlers) handleReply(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	telegramMessageID := int64(ctx.Message().ReplyTo.ID)
	rootCtx := h.RootCtx

	ex, err := queries.GetExerciseByTelegramIDAndUserID(rootCtx, sqlc.GetExerciseByTelegramIDAndUserIDParams{
		UserID:        user.ID,
		TelegramMsgID: telegramMessageID,
	})
	if err != nil {
		return fmt.Errorf("unable to find exercise (UserID: %d, TelegramMsgID: %d). %w", user.ID, telegramMessageID, err)
	}

	answerText := ctx.Text()
	isCorrectAnswer := strings.TrimSpace(answerText) == ex.Answer

	_, err = queries.SetExerciseAnswer(rootCtx, sqlc.SetExerciseAnswerParams{
		Answered:          true,
		AnsweredCorrectly: sql.NullBool{Bool: isCorrectAnswer, Valid: true},
		AnsweredAt:        sql.NullTime{Time: time.Now(), Valid: true},
		UserID:            user.ID,
		TelegramMsgID:     ex.TelegramMsgID,
	})
	if err != nil {
		return fmt.Errorf("failed to update exercise. %w", err)
	}

	w, err := queries.GetWordByID(rootCtx, ex.WordID)
	if err != nil {
		return fmt.Errorf("failed to get word. %w", err)
	}

	if !isCorrectAnswer {
		fmt.Printf("Expected: '%s', Got: '%s'\n", ex.Answer, strings.TrimSpace(answerText))
		err = queries.ResetWordAnsweredCount(h.RootCtx, sqlc.ResetWordAnsweredCountParams{
			LastShowedAt: sql.NullTime{Time: time.Now(), Valid: true},
			ID:           w.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to reset word answered count. %w", err)
		}
		err = ctx.Reply(MustRenderIncorrectAnswer(w, ex.Answer))
		if err != nil {
			return fmt.Errorf("failed to send incorrect answer reply. %w", err)
		}
		return nil
	}

	answeredCount, err := queries.IncrementWordAnsweredCount(rootCtx, sqlc.IncrementWordAnsweredCountParams{
		LastShowedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:           w.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update word answered count. %w", err)
	}

	if answeredCount >= 8 {
		err = queries.SetWordLearned(rootCtx, sqlc.SetWordLearnedParams{
			LearnedAt: sql.NullTime{Time: time.Now(), Valid: true},
			ID:        w.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to update word answered count. %w", err)
		}

		err = ctx.Reply(MustRenderCorrectAnswerLearned())
		if err != nil {
			return fmt.Errorf("failed to send correct answer reply for a learned word. %w", err)
		}
		return nil
	}

	err = ctx.Reply(MustRenderCorrectAnswer(int(answeredCount)))
	if err != nil {
		return fmt.Errorf("failed to send correct answer reply. %w", err)
	}

	return nil
}
