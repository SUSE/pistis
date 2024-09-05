/*
   Helper functions for pistis
   Copyright (C) 2024  SUSE LLC <georg.pfuetzenreuter@suse.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

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

func Debug(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

func handleInternalError(action string, err error) error {
	if err != nil {
		Debug("%s failed: %s", action, err)
	}

	return err
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

func fileToStrLines(filePath string) ([]string, error) {
	var lines []string

	file, err := os.Open(filePath)
	if err != nil {
		Debug("Reading file %s failed", filePath)
		return lines, err
	}

	defer file.Close()
	fileScanner := bufio.NewScanner(file)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	return lines, nil
}

func fileToStrMap(filePath string) (map[string]string, error) {
	lineMap := make(map[string]string)

	lines, err := fileToStrLines(filePath)
	if err != nil {
		return lineMap, err
	}

	for _, line := range lines {
		lineParts := strings.Fields(line)
		lineMap[lineParts[0]] = lineParts[1]
	}

	return lineMap, nil
}

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}

	return change.To.Name
}

func contains(values []string, value string) bool {
	for _, entry := range values {
		if entry == value {
			return true
		}
	}

	return false
}
