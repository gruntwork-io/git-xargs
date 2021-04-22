package main

import "github.com/gruntwork-io/go-commons/errors"

// Sanity check that user has provided one valid method for selecting repos to operate on
func ensureValidOptionsPassed(config *GitXargsConfig) error {
	if len(config.RepoSlice) < 1 && config.ReposFile == "" && config.GithubOrg == "" && len(config.RepoFromStdIn) == 0 {
		return errors.WithStackTrace(NoRepoSelectionsMadeErr{})
	}
	if config.BranchName == "" {
		return errors.WithStackTrace(NoBranchNameErr{})
	}
	return nil
}
