package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type Subcommand string

const (
	CreateUser Subcommand = "create-user"
	AddWord    Subcommand = "add-word"
)

type Config struct {
	Subcommand Subcommand
	Mongo      struct {
		URI            string `koanf:"uri"`
		ConnectTimeout time.Duration
		DatabaseName   string `koanf:"database"`
	} `koanf:"mongo"`
	CLI struct {
		CommandTimeout time.Duration
	}
	Log struct {
		Type  LogType
		Level slog.Level
	}
	ChatGPT struct {
		APIToken string `koanf:"token"`
	} `koanf:"chatgpt"`
	Exercise struct {
		Sentences struct {
			DefaultCount int `koanf:"count"`
		} `koanf:"sentences,omitempty"`
	} `koanf:"exercise"`

	Spelling        string `koanf:"spelling"`
	Definition      string `koanf:"definition"`
	Language        string `koanf:"language"`
	LexicalCategory string `koanf:"lexical-category"`
	UserID          string `koanf:"user-id"`
}

type LogType int8

const (
	LogTypeText = iota
	LogTypeJSON
)

const EnvPrefix = "VOCABFORGE_"

func ParseConfig(args []string) (Config, error) {
	subcmd, flagSet, err := parseArgs(args)
	if err != nil {
		return Config{}, fmt.Errorf("main.ParseConfig unable to parse args. %w", err)
	}

	k := koanf.New(".")

	// Load default values
	err = k.Load(structs.Provider(configWithDefaults(subcmd), "koanf"), nil)
	if err != nil {
		return Config{}, fmt.Errorf("main.ParseConfig unable to load structs provider with default values. %w", err)
	}

	// Load values from env and overwrite defaults
	err = k.Load(env.Provider(EnvPrefix, ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, EnvPrefix)),
			"_", ".")
	}), nil)
	if err != nil {
		return Config{}, fmt.Errorf("main.ParseConfig unable to load env provider. %w", err)
	}

	// Load values from CLI and overwrite previous config
	err = k.Load(basicflag.Provider(flagSet, "."), nil)
	if err != nil {
		return Config{}, fmt.Errorf("main.ParseConfig unable to load basicflag provider. %w", err)
	}

	var cfg Config

	err = k.Unmarshal("", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("main.ParseConfig unable to unmarshal config. %w", err)
	}

	return cfg, nil
}

func parseArgs(args []string) (Subcommand, *flag.FlagSet, error) {
	//nolint:mnd
	if len(args) < 2 {
		return "", nil, errors.New("missing subcommand")
	}

	var sb Subcommand

	switch args[1] {
	case string(CreateUser):
		sb = CreateUser
	case string(AddWord):
		sb = AddWord
	default:
		return "", nil, fmt.Errorf("unknown subcommand %s", args[1])
	}

	fs := flag.NewFlagSet(string(sb), flag.ContinueOnError)

	//nolint:gocritic
	switch sb {
	case AddWord:
		fs.String("user-id", "", "user id")
		fs.String("spelling", "", "word's spelling")
		fs.String("definition", "", "word's definition")
		fs.String("language", "", "spelling and definition language, for ex: en_US")
		fs.String("lexical-category", "", "lexical category of word")
	}

	err := fs.Parse(args[2:])
	if err != nil {
		return "", nil, fmt.Errorf("unable to parse flagset. %w", err)
	}

	return sb, fs, nil
}

func configWithDefaults(s Subcommand) Config {
	cfg := Config{
		Subcommand: s,
	}
	//nolint:mnd
	cfg.Mongo.ConnectTimeout = 5 * time.Second
	cfg.Mongo.DatabaseName = "vocabforge"

	//nolint:mnd
	cfg.CLI.CommandTimeout = 15 * time.Second

	cfg.Log.Level = slog.LevelDebug

	cfg.Language = "en_US"

	cfg.Exercise.Sentences.DefaultCount = 16

	return cfg
}
