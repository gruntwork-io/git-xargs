package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

// This utility function accepts a path to the flatfile in which the user has defined their explicitly allowed repos
// It expects repos to be defined one per line in the following format: `gruntwork-io/cloud-nuke` with optional commas
// Stray single and double quotes are also handled and stripped out if they are encountered, and spacing is irrelevant
func processAllowedRepos(filepath string) ([]*AllowedRepo, error) {
	logger := logging.GetLogger("git-xargs")

	var allowedRepos []*AllowedRepo

	filepath = strings.TrimSpace(strings.Trim(filepath, "\n"))

	file, err := os.Open(filepath)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error":    err,
			"Filepath": filepath,
		}).Debug("Could not open")

		return allowedRepos, err
	}

	// By wrapping the file.Close in a deferred anonymous function, we are able to avoid a nasty edge-case where
	// an actual closeErr would not be checked or handled properly in the more common `defer file.Close()`
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			logger.WithFields(logrus.Fields{
				"Error": closeErr,
			}).Debug("Error closing allowed repos file")
		}
	}()

	// Read through the file line by line, extracting the repo organization and name by splitting on the / char
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		allowedRepo := convertStringToAllowedRepo(scanner.Text())

		if allowedRepo != nil {
			allowedRepos = append(allowedRepos, allowedRepo)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error parsing line from allowed repos file")
	}

	return allowedRepos, nil
}
