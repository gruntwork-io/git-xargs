package io

import (
	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/types"
)

// EnsureValidOptionsPassed checks that user has provided one valid method for selecting repos to operate on
func EnsureValidOptionsPassed(config *config.GitXargsConfig) error {
	if len(config.RepoSlice) < 1 && config.ReposFile == "" && config.GithubOrg == "" && len(config.RepoFromStdIn) == 0 {
		return types.NoRepoSelectionsMadeErr{}
	}
	if config.BranchName == "" {
		return types.NoBranchNameErr{}
	}
	return nil
}
