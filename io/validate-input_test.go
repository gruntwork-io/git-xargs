package io

import (
	"testing"

	"github.com/gruntwork-io/git-xargs/config"
	"github.com/stretchr/testify/assert"
)

func TestEnsureValidOptionsPassedRejectsEmptySelectors(t *testing.T) {
	t.Parallel()
	emptyTestConfig := &config.GitXargsConfig{}

	err := EnsureValidOptionsPassed(emptyTestConfig)
	assert.Error(t, err)
}

func TestEnsureValidOptionsPassedAcceptsValidGithubOrg(t *testing.T) {
	t.Parallel()
	testConfigWithGithubOrg := &config.GitXargsConfig{
		BranchName: "test-branch",
		GithubOrg:  "gruntwork-io",
	}

	err := EnsureValidOptionsPassed(testConfigWithGithubOrg)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptsValidReposFile(t *testing.T) {
	t.Parallel()
	testConfigWithReposFile := &config.GitXargsConfig{
		BranchName: "test-branch",
		ReposFile:  "./my-repos.txt",
	}

	err := EnsureValidOptionsPassed(testConfigWithReposFile)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptedValidSingleRepo(t *testing.T) {
	t.Parallel()
	testConfigWithExplicitRepos := &config.GitXargsConfig{
		BranchName: "test-branch",
		RepoSlice:  []string{"gruntwork-io/cloud-nuke"},
	}

	err := EnsureValidOptionsPassed(testConfigWithExplicitRepos)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptsAllFlagsSimultaneously(t *testing.T) {
	t.Parallel()
	testConfigWithAllSelectionCriteria := &config.GitXargsConfig{
		BranchName:    "test-branch",
		ReposFile:     "./my-repos.txt",
		RepoSlice:     []string{"gruntwork-io/cloud-nuke", "gruntwork-io/fetch"},
		GithubOrg:     "github-org",
		RepoFromStdIn: []string{"gruntwork-io/terragrunt"},
	}

	err := EnsureValidOptionsPassed(testConfigWithAllSelectionCriteria)
	assert.NoError(t, err)
}
