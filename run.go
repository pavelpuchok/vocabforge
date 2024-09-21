package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/pavelpuchok/vocabforge/usecases/createuser"
	"github.com/pavelpuchok/vocabforge/users"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func run(cfg Config, logger *slog.Logger, args []string) error {
	//nolint:mnd
	if len(args) < 2 {
		return errors.New("atleast one subcommand (create-user) should be passed")
	}

	if cfg.Mongo.URI == "" {
		return errors.New("main.run missing MongoDB URI")
	}

	db, err := initializeMongoDB(cfg)
	if err != nil {
		return fmt.Errorf("main.run unable to establish mongo database connection. %w", err)
	}

	createUser := createuser.UseCase{
		UsersService: users.NewService(users.NewMongoRepository(db)),
	}

	switch args[1] {
	case "create-user":
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CLI.CommandTimeout)
		defer cancel()

		usr, err := createUser.Run(ctx)
		if err != nil {
			return fmt.Errorf("main.run unable to create user. %w", err)
		}

		logger.InfoContext(ctx, "CreateUser: User created", slog.Any("user", usr))
	default:
		return fmt.Errorf("unexpected subcommand %s", args[1])
	}

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

type App struct {
}

type CreateUserUseCase interface {
	Run(ctx context.Context)
}
