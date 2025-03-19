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
	"github.com/pavelpuchok/vocabforge/server"
	"github.com/pavelpuchok/vocabforge/telegram"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/telebot.v4"
)

func main() {
	logLevel := os.Getenv("VF_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	logger := initLogger(logLevel)

	openaiAPIToken := readSecretsFile(os.Getenv("VF_OPENAI_TOKEN_FILE"))
	tgAPIToken := readSecretsFile(os.Getenv("VF_TELEGRAM_TOKEN_FILE"))
	httpToken := readSecretsFile(os.Getenv("VF_HTTP_TOKEN_FILE"))
	httpAddr := os.Getenv("VF_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	sqlDBPath := os.Getenv("VF_SQL_DB_PATH")

	sqlDB, err := db.NewSQLiteWithMigrations(sqlDBPath)
	if err != nil {
		panic(err)
	}

	openAI := ai.NewOpenAI(openai.NewClient(openaiAPIToken))
	queries := sqlc.New(sqlDB)

	rootCtx := context.Background()

	queue := job.Queue[job.TranslateWordsJob]{Storage: job.DBStorage{Queries: queries}, GroupName: "job.TranslateWordsJob"}

	worker := job.TranslateWordsJobWorker{
		Queries: queries,
		Queue:   queue,
		Logger:  logger.With(slog.String("Caller", "job.TranslateWordsJobWorker")),
		OpenAI:  openAI,
	}

	go worker.Run(rootCtx)

	httpServer := server.New(httpAddr, httpToken, queue, logger)
	go func() {
		if err := httpServer.Run(rootCtx); err != nil {
			logger.Error("unable to run HTTP server", slog.String("error", err.Error()))
		}
	}()

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
		Logger:  logger,
	}

	handlers.Register()

	bot.Start()
}

func initLogger(level string) *slog.Logger {
	var l slog.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		panic(fmt.Sprintf("unable to parse level: %s. err: %s", level, err))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}

func readSecretsFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("unable to read secrets file %s. Error: %s", path, err))
	}
	return string(data)
}
