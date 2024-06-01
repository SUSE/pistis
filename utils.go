package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5/plumbing/object"
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

func fileToStr(file string) (content string) {
	contentBytes, readErr := os.ReadFile(file)
	handleError("Reading file", readErr)
	return string(contentBytes)
}

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}

	return change.To.Name
}
