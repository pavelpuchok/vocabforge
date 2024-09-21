package main

import (
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Mongo struct {
		URI            string
		ConnectTimeout time.Duration
		DatabaseName   string
	}
	CLI struct {
		CommandTimeout time.Duration
	}
	Log struct {
		Type  LogType
		Level slog.Level
	}
}

type LogType int8

const (
	LogTypeText = iota
	LogTypeJSON
)

type Env interface {
	Getenv(key string) string
}

type EnvOs struct{}

func (_ EnvOs) Getenv(key string) string {
	return os.Getenv(key)
}

func ParseConfigFromEnv(env Env) (Config, error) {
	cfg := configWithDefaults()

	// Read values from environment variables
	cfg.Mongo.URI = env.Getenv("MONGO_URI")

	return cfg, nil
}

func configWithDefaults() Config {
	cfg := Config{}
	//nolint:mnd
	cfg.Mongo.ConnectTimeout = 5 * time.Second
	cfg.Mongo.DatabaseName = "vocabforge"

	//nolint:mnd
	cfg.CLI.CommandTimeout = 3 * time.Second

	cfg.Log.Level = slog.LevelDebug

	return cfg
}
