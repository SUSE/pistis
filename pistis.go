package main

import (
	"bufio"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/hairyhenderson/go-codeowners"

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

func getCodeOwnerFingerprints(coFpPath string) map[string]string {
	coFpFile, err := os.Open(coFpPath)
	handleError("Reading fingerprints file", err)
	defer coFpFile.Close()
	coFpScanner := bufio.NewScanner(coFpFile)

	coFpMap := make(map[string]string)

	for coFpScanner.Scan () {
		coFpParts := strings.Split(coFpScanner.Text(), " ")
		coFpMap[coFpParts[0]] = coFpParts[1]
	}

	return coFpMap
}

func logic() {
	repository, err := git.PlainOpen(directory)
	handleError("Opening repository", err)

	co, err := codeowners.FromFile(directory)
	handleError("Reading CODEOWNERS", err)

	coFpMap := getCodeOwnerFingerprints(directory + "/CODEOWNERS_FINGERPRINTS")

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
				owners := co.Owners(file)
				for i, owner := range owners {
					ownerFp, haveOwnerFp := coFpMap[owner]
					if haveOwnerFp {
						Info("Owner #%d is %s with fingerprint %s", i, owner, ownerFp)
					} else {
						Info("Owner #%d is %s with no fingerprint", i, owner)
					}
				}
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
