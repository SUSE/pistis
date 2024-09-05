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
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"

	"github.com/hairyhenderson/go-codeowners"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

var (
	keyring            string
	keyringFile        string
	directory          string
	logger             *slog.Logger
	featureIgnoreMerge bool
	gitlab             string
	gitlab_token       string
)

func main() {
	var logLevelStr string

	flag.BoolVar(&featureIgnoreMerge, "ignore-merge", false, "Do not try to validate merge commits")
	flag.StringVar(&directory, "repository", ".", "Path to the Git repository")
	flag.StringVar(&gitlab, "gitlab", "", "URL to a GitLab instance for building the PGP keyring")
	flag.StringVar(&keyringFile, "keyring", "", "Alternatively, path to file containing an existing armored keyring")
	flag.StringVar(&logLevelStr, "loglevel", "info", "Logging level")

	flag.Parse()

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: convertLogLevel(logLevelStr)}))

	gitlab_token = os.Getenv("GITLAB_TOKEN")

	if keyringFile == "" && gitlab == "" {
		Error("Specify -gitlab OR -keyring.")
		os.Exit(1)
	}

	if keyringFile == "" && gitlab != "" && gitlab_token == "" {
		Error("Pass GITLAB_TOKEN or -keyring.")
		os.Exit(1)
	}

	logic()
}

func buildKeyring(coUserPath string, gitlabUserNames []string, gitlab string) string {
	Debug("buildKeyring()")

	ring, err := crypto.NewKeyRing(nil)
	handleError("Creating keyring", err)

	codeownerUserNames, err := fileToStrMap(coUserPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		handleError("Reading users file", err)
	}

	for _, username := range codeownerUserNames {
		if contains(gitlabUserNames, username) {
			Debug("Redundant user entry for %s", username)
		} else {
			gitlabUserNames = append(gitlabUserNames, username)
		}
	}

	for _, username := range gitlabUserNames {
		response, err := http.Get(fmt.Sprintf("%s/%s.gpg", gitlab, username))
		msg := fmt.Sprintf(" for %s", username)
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

func getGitLabUsernames(history object.CommitIter, exclusions []string) ([]string, error) {
	Debug("getGitLabUsernames()")

	if gitlab == "" || gitlab_token == "" {
		Debug("No GitLab present, skipping.")
		return nil, nil
	}

	client := &http.Client{}

	var gitlabUserEmails []string
	var gitlabUserNames []string

	type GitLabUser struct {
		Username string `json:"username"`
	}

	err := history.ForEach(func(commit *object.Commit) error {
		hash := commit.Hash.String()
		if contains(exclusions, hash) {
			Debug("Returning early at commit %s", hash)
			return storer.ErrStop
		}

		committer := commit.Committer
		email := committer.Email

		if contains(gitlabUserEmails, email) {
			return nil
		}

		gitlabUserEmails = append(gitlabUserEmails, email)

		url := fmt.Sprintf("%s/api/v4/users?search=%s", gitlab, email)
		Debug(url)

		request, err := http.NewRequest("GET", url, nil)
		handleError(fmt.Sprintf("Constructing request to %s", url), err)

		request.Header.Set("PRIVATE-TOKEN", gitlab_token)

		response, err := client.Do(request)
		handleError(fmt.Sprintf("Requesting %s", url), err)

		body, err := ioutil.ReadAll(response.Body)
		handleError(fmt.Sprintf("Parsing response from %s", url), err)

		if response.StatusCode != http.StatusOK {
			handleError(fmt.Sprintf("Requesting %s", url), errors.New(response.Status))
		}

		var FoundUsers []GitLabUser

		jerr := json.Unmarshal(body, &FoundUsers)
		handleError(fmt.Sprintf("Parsing response from %s as JSON", url), jerr)

		if len(FoundUsers) == 0 {
			Debug("No username found for %s", email)
			return nil
		}

		if len(FoundUsers) > 1 {
			// TODO: whilst it is unlikely to have more than one user returned for a given email address, it would still be good to handle the scenario better
			Debug("More than one user found, result might be inaccurate: %s", FoundUsers)
		}

		username := FoundUsers[0].Username
		Debug("Found user in GitLab: %s", username)
		gitlabUserNames = append(gitlabUserNames, username)

		return nil

	})
	handleError("Parsing commits", err)

	return gitlabUserNames, nil
}

func logic() {
	Debug("logic()")

	repository, err := git.PlainOpen(directory)
	handleError("Opening repository", err)

	co, err := codeowners.FromFile(directory)
	handleError("Reading CODEOWNERS", err)

	coFpMap, err := fileToStrMap(directory + "/CODEOWNERS_FINGERPRINTS")
	handleError("Reading fingerprints file", err)

	exclusions, err := fileToStrLines(directory + "/TRUSTED_COMMITS")
	handleError("Reading exclusion file", err)

	ref, err := repository.Head()
	handleError("Reading HEAD", err)

	head := ref.Hash()
	Info("Head is at %s", head)

	history, err := repository.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderBSF})
	handleError("Reading history", err)

	gitlabUserNames, err := getGitLabUsernames(history, exclusions)
	handleError("Getting usernames from GitLab", err)

	if keyringFile != "" || gitlab == "" || gitlab_token == "" {
		Debug("Using keyring from %s", keyringFile)
		keyring = fileToStr(keyringFile)
	} else {
		keyring = buildKeyring(directory+"/CODEOWNERS_USERNAMES", gitlabUserNames, gitlab)
	}

	// why is the history empty if consumed a second time?
	history, err = repository.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderBSF})
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
