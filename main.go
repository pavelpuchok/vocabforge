package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pavelpuchok/vocabforge/ai"
	"github.com/pavelpuchok/vocabforge/db"
	"github.com/pavelpuchok/vocabforge/db/sqlc"
	"github.com/pavelpuchok/vocabforge/job"
	"github.com/pavelpuchok/vocabforge/telegram"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/telebot.v4"
)

func main() {
	openaiAPIToken := readSecretsFile(os.Getenv("VF_OPENAI_TOKEN_FILE"))
	tgAPIToken := readSecretsFile(os.Getenv("VF_TELEGRAM_TOKEN_FILE"))
	sqlDBPath := os.Getenv("VF_SQL_DB_PATH")

	sqlDB, err := db.NewSQLiteWithMigrations(sqlDBPath)
	if err != nil {
		panic(err)
	}

	logger := slog.Default()
	openAI := ai.NewOpenAI(openai.NewClient(openaiAPIToken))
	queries := sqlc.New(sqlDB)

	rootCtx := context.Background()

	worker := job.TranslateWordsJobWorker{
		Queries: queries,
		Queue:   job.Queue[job.TranslateWordsJob]{Storage: job.DBStorage{Queries: queries}, GroupName: "job.TranslateWordsJob"},
		Logger:  logger.With(slog.String("Caller", "job.TranslateWordsJobWorker")),
		OpenAI:  openAI,
	}

	go worker.Run(rootCtx)

	settings := telebot.Settings{
		Token:     tgAPIToken,
		Poller:    &telebot.LongPoller{Timeout: 2 * time.Second},
		ParseMode: telebot.ModeMarkdownV2,
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		panic(err)
	}

	bot.Use(telegram.NewSQLTxMiddleware(sqlDB, logger))

	handlers := telegram.Handlers{
		OpenAI:  openAI,
		RootCtx: rootCtx,
		Bot:     bot,
	}

	handlers.Register()

	bot.Start()
}

func readSecretsFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("unable to read secrets file %s. Error: %s", path, err))
	}
	return string(data)
}
