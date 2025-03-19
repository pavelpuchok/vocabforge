package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pavelpuchok/vocabforge/ai"
	"github.com/pavelpuchok/vocabforge/db/sqlc"
	"gopkg.in/telebot.v4"
)

type Handlers struct {
	RootCtx context.Context
	OpenAI  *ai.OpenAI
	Bot     *telebot.Bot
	Logger  *slog.Logger
}

type NonAuthorizedHandler func(telebot.Context, *sqlc.Queries) error
type AuthorizedHandler func(telebot.Context, *sqlc.Queries, sqlc.User) error

func (h Handlers) Register() {
	h.registerNonAuthorized("/start", h.handleStart)
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

	err = ctx.Reply(formatMessage("Hello %s! Your ID: %d", usr.Name, usr.ID))
	if err != nil {
		return fmt.Errorf("failed to send response. %w", err)
	}

	return nil
}

func (h Handlers) handleLearnVocab(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	h.Logger.Debug("handleLearnVocab about to start", slog.Int64("userID", user.ID))

	learningCount, err := queries.CountLearningWords(h.RootCtx, user.ID)
	if err != nil {
		return fmt.Errorf("unable to count learning words. %w", err)
	}
	h.Logger.Debug(fmt.Sprintf("handleLearnVocab total learning words: %d", learningCount), slog.Int64("userID", user.ID))

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
		h.Logger.Debug(fmt.Sprintf("handleLearnVocab using the oldest unseen word"), slog.Int64("userID", user.ID), slog.Int64("wordID", w.ID))
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
		h.Logger.Debug(fmt.Sprintf("handleLearnVocab using new word as no learning ones left"), slog.Int64("userID", user.ID), slog.Int64("wordID", w.ID))
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

func (h Handlers) handleReply(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	h.Logger.Debug(fmt.Sprintf("handleReply is about to start"), slog.Int64("userID", user.ID))
	telegramMessageID := int64(ctx.Message().ReplyTo.ID)
	rootCtx := h.RootCtx

	ex, err := queries.GetExerciseByTelegramIDAndUserID(rootCtx, sqlc.GetExerciseByTelegramIDAndUserIDParams{
		UserID:        user.ID,
		TelegramMsgID: telegramMessageID,
	})
	if err != nil {
		return fmt.Errorf("unable to find exercise (UserID: %d, TelegramMsgID: %d). %w", user.ID, telegramMessageID, err)
	}

	answerText := strings.TrimSpace(strings.ToLower(ctx.Text()))
	expectedText := strings.TrimSpace(strings.ToLower(ex.Answer))
	isCorrectAnswer := answerText == expectedText

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
		h.Logger.Debug(fmt.Sprintf("handleReply answer is incorrect. Expected: %s, Got: %s", expectedText, answerText), slog.Int64("userID", user.ID), slog.Int64("wordID", w.ID), slog.Int64("exerciseID", ex.ID))
		err = queries.ResetWordAnsweredCount(h.RootCtx, w.ID)
		if err != nil {
			return fmt.Errorf("failed to reset word answered count. %w", err)
		}
		err = ctx.Reply(MustRenderIncorrectAnswer(w, ex.Answer))
		if err != nil {
			return fmt.Errorf("failed to send incorrect answer reply. %w", err)
		}
		return nil
	}

	answeredCount, err := queries.IncrementWordAnsweredCount(rootCtx, w.ID)
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

		h.Logger.Debug("handleReply word is learned", slog.Int64("userID", user.ID), slog.Int64("wordID", w.ID), slog.Int64("exerciseID", ex.ID))
		return nil
	}

	err = ctx.Reply(MustRenderCorrectAnswer(int(answeredCount + 1)))
	if err != nil {
		return fmt.Errorf("failed to send correct answer reply. %w", err)
	}
	h.Logger.Debug(fmt.Sprintf("handleReply answer is correct. Answered Count: %d", answeredCount), slog.Int64("userID", user.ID), slog.Int64("wordID", w.ID), slog.Int64("exerciseID", ex.ID))

	return nil
}

func (h Handlers) handleCallback(ctx telebot.Context, queries *sqlc.Queries, user sqlc.User) error {
	h.Logger.Debug(fmt.Sprintf("handleCallback is about to start"), slog.Int64("userID", user.ID))
	return nil
}
