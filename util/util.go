package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/gruntwork-io/git-xargs/types"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ConvertStringToAllowedRepo accepts a user-supplied repo in the format of <github-organization>/<repo-name>.
// It trims out stray characters that we might expect in a repos file that was copy-pasted from json or an array,
// and it only returns an AllowedRepo if the user-supplied input looks valid. Note this does not actually look
// up the repo via the GitHub API because that's slow, and we do it later when converting repo names to GitHub response structs.
func ConvertStringToAllowedRepo(repoInput string) *types.AllowedRepo {

	logger := logging.GetLogger("git-xargs")

	// The regex for all common special characters to remove from the repo lines in the allowed repos file
	charRegex := regexp.MustCompile(`['",!]`)

	trimmedLine := strings.TrimSpace(repoInput)
	cleanedLine := charRegex.ReplaceAllString(trimmedLine, "")
	orgAndRepoSlice := strings.Split(cleanedLine, "/")
	// Guard against stray lines, extra dangling single quotes, etc
	if len(orgAndRepoSlice) < 2 {

		logger.WithFields(logrus.Fields{
			"Repo input": repoInput,
		}).Debug("Malformed repo input detected - skipping")

		return nil
	}

	// Validate both the org and name are not empty
	parsedOrg := orgAndRepoSlice[0]
	parsedName := orgAndRepoSlice[1]

	// If both org name and repo name are present, create a new allowed repo and add it to the list
	if parsedOrg != "" && parsedName != "" {
		repo := &types.AllowedRepo{
			Organization: parsedOrg,
			Name:         parsedName,
		}
		return repo
	}

	logger.WithFields(logrus.Fields{
		"Repo input": repoInput,
	}).Debug("Could not parse a valid repo from input. Repo must be specified in format <github-org>/<repo-name>, e.g., gruntwork-io/cloud-nuke")

	return nil
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func NewTestFileName() string {
	return fmt.Sprintf("test-file-%s", RandStringBytes(9))
}
