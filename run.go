package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/pavelpuchok/vocabforge/models"
	"github.com/pavelpuchok/vocabforge/usecases/addword"
	"github.com/pavelpuchok/vocabforge/usecases/createuser"
	"github.com/pavelpuchok/vocabforge/users"
	"github.com/pavelpuchok/vocabforge/vocabulary"
	"github.com/pavelpuchok/vocabforge/vocabulary/sentences"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func run(cfg Config, logger *slog.Logger) error {
	if cfg.Mongo.URI == "" {
		return errors.New("main.run missing MongoDB URI")
	}

	db, err := initializeMongoDB(cfg)
	if err != nil {
		return fmt.Errorf("main.run unable to establish mongo database connection. %w", err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			logger.Error("main.run failed to disconnect MongoDB client", slog.String("err", err.Error()))
		}
	}()

	switch cfg.Subcommand {
	case CreateUser:
		err := processCreateUserCmd(logger, cfg, db)
		if err != nil {
			return fmt.Errorf("main.run create user command failed. %w", err)
		}
	case AddWord:
		err := processAddWordCmd(logger, cfg, db)
		if err != nil {
			return fmt.Errorf("main.run add word command failed. %w", err)
		}
	}

	return nil
}

func processCreateUserCmd(logger *slog.Logger, cfg Config, db *mongo.Database) error {
	createUser := createuser.UseCase{
		UsersService: users.NewService(users.NewMongoRepository(db)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CLI.CommandTimeout)
	defer cancel()

	usr, err := createUser.Run(ctx)
	if err != nil {
		return fmt.Errorf("main.processCreateUserCmd unable to create user. %w", err)
	}

	logger.InfoContext(ctx, "CreateUser: User created", slog.String("user_id", usr.ID.String()))
	return nil
}

func processAddWordCmd(logger *slog.Logger, cfg Config, db *mongo.Database) error {
	promptProvider, err := sentences.NewAIPromptProvider()
	if err != nil {
		return fmt.Errorf("main.processAddWordCmd unable to create prompt provider. %w", err)
	}

	aiGenerator, err := sentences.NewAIGenerator(cfg.ChatGPT.APIToken, promptProvider)
	if err != nil {
		return fmt.Errorf("main.processAddWordCmd unable to create AI generator. %w", err)
	}

	addWord := addword.UseCase{
		VocabularyService: vocabulary.NewService(vocabulary.NewMongoRepository(db), aiGenerator, cfg.Exercise.Sentences.DefaultCount),
	}

	userId, err := models.UserIDFromText(cfg.UserID)
	if err != nil {
		return fmt.Errorf("main.processAddWordCmd invalid user id received. %w", err)
	}

	lang, err := models.LanguageFromText(cfg.Language)
	if err != nil {
		return fmt.Errorf("main.processAddWordCmd invalid lang received. %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CLI.CommandTimeout)
	defer cancel()

	word, err := addWord.Run(ctx, userId, cfg.Spelling, cfg.Definition, cfg.LexicalCategory, lang)
	if err != nil {
		return fmt.Errorf("main.processAddWordCmd unable to add word. %w", err)
	}

	logger.InfoContext(ctx, "AddWord: word added", slog.String("word_id", word.ID.String()))
	return nil
}

func initializeMongoDB(cfg Config) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, fmt.Errorf("main.initializeMongoDB unable to connect to MongoDB. %w", err)
	}

	return client.Database(cfg.Mongo.DatabaseName), nil
}

func initializeLogger(w io.Writer, cfg Config) (*slog.Logger, error) {
	var h slog.Handler
	switch cfg.Log.Type {
	case LogTypeJSON:
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: cfg.Log.Level,
		})
	case LogTypeText:
		h = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: cfg.Log.Level,
		})
	default:
		return nil, fmt.Errorf("unknown log type %v", cfg.Log.Type)
	}
	return slog.New(h), nil
}
