package main

import (
	"flag"
	"log/slog"
	"os"

//	"github.com/hairyhenderson/go-codeowners"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
//	"github.com/go-git/go-git/v5/utils/merkletrie"
)

var (
	keyring string
	directory string
	logger *slog.Logger
)

func main() {
	var keyringFile string
	var logLevelStr string

	flag.StringVar(&keyringFile, "keyring", "", "Path to file containing an armored keyring")
	flag.StringVar(&directory, "repository", ".", "Path to the Git repository")
	flag.StringVar(&logLevelStr, "loglevel", "info", "Logging level")

	flag.Parse()

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: convertLogLevel(logLevelStr)}))

	keyring = fileToStr(keyringFile)

	logic()
}

func logic() {
	repository, err := git.PlainOpen(directory)
	handleError("Opening repository", err)

	ref, err := repository.Head()
	handleError("Reading HEAD", err)

	head := ref.Hash()
	Info("Head is at %s", head)

	history, err := repository.Log(&git.LogOptions{From: ref.Hash()})
	handleError("Reading history", err)

	var previousTree *object.Tree

	err = history.ForEach(func(commit *object.Commit) error {
		Info("%s", commit.Hash)

		tree, err := commit.Tree()
		handleError("Reading commit tree", err)

		if previousTree != nil {

			patch, err := tree.Patch(previousTree)
			handleError("Reading patch", err)

			var changedFiles []string

			for _, fileStat := range patch.Stats() {
				changedFiles = append(changedFiles,fileStat.Name)
			}

			for _, file := range changedFiles {
				Info(file)
			}

			//changes, err := tree.Diff(previousTree)
			//handleError("Reading diff", err)

			//for _, change := range changes {
			//	action, err := change.Action()
			//	handleError("Reading action", err)
			//	if action != merkletrie.Delete {
			//		name := getChangeName(change)
			//		Info(name)
			//	}
			//}
		}

		_, verifyErr := commit.Verify(keyring)
		handleError("Verifying commit", verifyErr)

		previousTree = tree

		return nil
	})
	handleError("Parsing commits", err)
}