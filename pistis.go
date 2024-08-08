/*
   Git commit verification tool
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
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/hairyhenderson/go-codeowners"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

var (
	keyring            string
	directory          string
	logger             *slog.Logger
	featureIgnoreMerge bool
)

func main() {
	var gitlab string
	var keyringFile string
	var logLevelStr string

	flag.BoolVar(&featureIgnoreMerge, "ignore-merge", false, "Do not try to validate merge commits")
	flag.StringVar(&directory, "repository", ".", "Path to the Git repository")
	flag.StringVar(&gitlab, "gitlab", "", "URL to a GitLab instance for building the PGP keyring")
	flag.StringVar(&keyringFile, "keyring", "", "Alternatively, path to file containing an existing armored keyring")
	flag.StringVar(&logLevelStr, "loglevel", "info", "Logging level")

	flag.Parse()

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: convertLogLevel(logLevelStr)}))

	if keyringFile == "" && gitlab != "" {
		keyring = buildKeyring(directory+"/CODEOWNERS_USERNAMES", gitlab)
	} else if keyringFile != "" && gitlab == "" {
		keyring = fileToStr(keyringFile)
	} else {
		Error("Specify -gitlab OR -keyring.")
		os.Exit(1)
	}

	logic()
}

func getCodeOwnerFingerprints(coFpPath string) map[string]string {
	coFpFile, err := os.Open(coFpPath)
	handleError("Reading fingerprints file", err)
	defer coFpFile.Close()
	coFpScanner := bufio.NewScanner(coFpFile)

	coFpMap := make(map[string]string)

	for coFpScanner.Scan() {
		coFpParts := strings.Split(coFpScanner.Text(), " ")
		coFpMap[coFpParts[0]] = coFpParts[1]
	}

	return coFpMap
}

func getCodeOwnerUsernames(coUserPath string) map[string]string {
	coUserFile, err := os.Open(coUserPath)
	handleError("Reading users file", err)
	defer coUserFile.Close()
	coUserScanner := bufio.NewScanner(coUserFile)

	coUserMap := make(map[string]string)

	for coUserScanner.Scan() {
		coUserParts := strings.Split(coUserScanner.Text(), " ")
		coUserMap[coUserParts[0]] = coUserParts[1]
	}

	return coUserMap
}

func getExclusions(exclPath string) []string {
	exclFile, err := os.Open(exclPath)
	handleError("Reading exclusion file", err)
	defer exclFile.Close()
	exclScanner := bufio.NewScanner(exclFile)

	exclusions := make([]string, 0)

	for exclScanner.Scan() {
		exclusions = append(exclusions, exclScanner.Text())
	}

	return exclusions
}

func buildKeyring(coUserPath string, gitlab string) string {
	ring, err := crypto.NewKeyRing(nil)
	handleError("Creating keyring", err)

	for email, username := range getCodeOwnerUsernames(coUserPath) {
		response, err := http.Get(fmt.Sprintf("%s/%s.gpg", gitlab, username))
		msg := fmt.Sprintf(" for %s (%s)", email, username)
		handleError("Reading key"+msg, err)
		defer response.Body.Close()
		handleError("Reading response"+msg, err)
		// TODO: validate response ?
		key, err := crypto.NewKeyFromArmoredReader(response.Body)
		handleError("Constructing key"+msg, err)
		handleError("Adding key"+msg, ring.AddKey(key))
	}

	armoredRing, err := ring.Armor()
	handleError("Encoding ring", err)

	return armoredRing
}

func logic() {
	repository, err := git.PlainOpen(directory)
	handleError("Opening repository", err)

	co, err := codeowners.FromFile(directory)
	handleError("Reading CODEOWNERS", err)

	coFpMap := getCodeOwnerFingerprints(directory + "/CODEOWNERS_FINGERPRINTS")

	exclusions := getExclusions(directory + "/TRUSTED_COMMITS")

	ref, err := repository.Head()
	handleError("Reading HEAD", err)

	head := ref.Hash()
	Info("Head is at %s", head)

	history, err := repository.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderBSF})
	handleError("Reading history", err)

	var previousTree *object.Tree

	err = history.ForEach(func(commit *object.Commit) error {
		if featureIgnoreMerge && len(commit.ParentHashes) > 1 {
			Debug("Ignoring merge commit %s (%d) parents: %s", commit.Hash.String()[0:9], commit.NumParents(), commit.ParentHashes)
			return nil
		}
		hash := commit.Hash

		if contains(exclusions, hash.String()) {
			Info("Returning early at commit %s", hash)
			return storer.ErrStop
		}

		Info("Reading commit %s", hash)

		pgpObj, verifyErr := commit.Verify(keyring)
		Debug("Author %s", commit.Author)
		Debug("Committer %s", commit.Committer)
		Debug("Signature %s", commit.PGPSignature)

		handleError("Verifying commit", verifyErr)
		cFp := hex.EncodeToString(pgpObj.PrimaryKey.Fingerprint[:])

		tree, err := commit.Tree()
		handleError("Reading commit tree", err)

		if previousTree != nil {

			patch, err := tree.Patch(previousTree)
			handleError("Reading patch", err)

			var changedFiles []string

			for _, fileStat := range patch.Stats() {
				changedFiles = append(changedFiles, fileStat.Name)
			}

			for _, file := range changedFiles {
				Info(file)
				owners := co.Owners(file)
				for i, owner := range owners {
					ownerFp, haveOwnerFp := coFpMap[owner]
					foundValidOwner := false
					if haveOwnerFp {
						Info("Owner #%d is %s with fingerprint %s", i, owner, ownerFp)
						Info("Commit is signed by fingerprint %s", cFp)
						if cFp == ownerFp {
							Info("Matches")
							foundValidOwner = true
						}

					} else {
						// all CODEOWNERS must have an associated fingerprint
						Error("Owner #%d is %s with no fingerprint", i, owner)
						os.Exit(1)
					}
					if !foundValidOwner {
						Error("File is covered by CODEOWNERS, but commit modifying it was not signed by a valid owner.")
						os.Exit(1)
					}
				}
			}
		}

		previousTree = tree

		return nil
	})
	handleError("Parsing commits", err)
}
