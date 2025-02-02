package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/pavelpuchok/vocabforge/ai"
	"github.com/pavelpuchok/vocabforge/db"
	"github.com/pavelpuchok/vocabforge/telegram"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/telebot.v4"
)

func main() {
	openaiAPIToken := os.Getenv("VF_OPENAI_TOKEN")
	tgAPIToken := os.Getenv("VF_TELEGRAM_TOKEN")
	sqlDBPath := os.Getenv("VF_SQL_DB_PATH")

	sqlDB, err := db.NewSQLiteWithMigrations(sqlDBPath)
	if err != nil {
		panic(err)
	}

	settings := telebot.Settings{
		Token:     tgAPIToken,
		Poller:    &telebot.LongPoller{Timeout: 2 * time.Second},
		ParseMode: telebot.ModeMarkdownV2,
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		panic(err)
	}

	bot.Use(telegram.NewSQLTxMiddleware(sqlDB, slog.Default()))

	handlers := telegram.Handlers{
		OpenAI:  ai.NewOpenAI(openai.NewClient(openaiAPIToken)),
		RootCtx: context.Background(),
		Bot:     bot,
	}

	handlers.Register()

	bot.Start()
}
