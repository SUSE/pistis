package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	directory string
	logger *slog.Logger
)

func convertLogLevel(levelStr string) slog.Level {
	switch levelStr {
		case "debug":
			return slog.LevelDebug
		case "info":
			return slog.LevelInfo
		case "warn":
			return slog.LevelWarn
		case "error":
			return slog.LevelError
		default:
			return slog.LevelInfo
	}
}

func Error(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

func handleError(action string, err error) {
	if err != nil {
		Error("%s failed: %s", action, err)
		os.Exit(1)
	}
}

func main() {
	var logLevelStr string

	flag.StringVar(&directory, "repository", ".", "Path to the Git repository")
	flag.StringVar(&logLevelStr, "loglevel", "info", "Logging level")

	flag.Parse()

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: convertLogLevel(logLevelStr)}))

	repository, err := git.PlainOpen(directory)
	handleError("Opening repository", err)

	ref, err := repository.Head()
	handleError("Reading HEAD", err)

	head := ref.Hash()
	Info("Head is at %s", head)

	cIter, err := repository.Log(&git.LogOptions{From: ref.Hash()})
	handleError("Reading history", err)

	err = cIter.ForEach(func(c *object.Commit) error {
		Info("Commit %s", c)
		return nil
	})
	handleError("Parsing commits", err)
}
