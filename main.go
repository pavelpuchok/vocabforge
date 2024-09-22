package main

import (
	"log/slog"
	"os"
)

const exitErrCode = 2

func main() {
	cfg, err := ParseConfig(os.Args)
	if err != nil {
		panic(err)
	}

	logger, err := initializeLogger(os.Stdout, cfg)
	if err != nil {
		logger.Error("main.main unable to initialize logger", slog.String("err", err.Error()))
		os.Exit(exitErrCode)
	}

	err = run(cfg, logger)
	if err != nil {
		logger.Error("failed", slog.String("err", err.Error()))
		os.Exit(exitErrCode)
	}
}
