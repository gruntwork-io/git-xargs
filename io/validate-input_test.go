package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureValidOptionsPassedRejectsEmptySelectors(t *testing.T) {
	t.Parallel()
	emptyTestConfig := &GitXargsConfig{}

	err := ensureValidOptionsPassed(emptyTestConfig)
	assert.Error(t, err)
}

func TestEnsureValidOptionsPassedAcceptsValidGithubOrg(t *testing.T) {
	t.Parallel()
	testConfigWithGithubOrg := &GitXargsConfig{
		BranchName: "test-branch",
		GithubOrg:  "gruntwork-io",
	}

	err := ensureValidOptionsPassed(testConfigWithGithubOrg)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptsValidReposFile(t *testing.T) {
	t.Parallel()
	testConfigWithReposFile := &GitXargsConfig{
		BranchName: "test-branch",
		ReposFile:  "./my-repos.txt",
	}

	err := ensureValidOptionsPassed(testConfigWithReposFile)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptedValidSingleRepo(t *testing.T) {
	t.Parallel()
	testConfigWithExplicitRepos := &GitXargsConfig{
		BranchName: "test-branch",
		RepoSlice:  []string{"gruntwork-io/cloud-nuke"},
	}

	err := ensureValidOptionsPassed(testConfigWithExplicitRepos)
	assert.NoError(t, err)
}

func TestEnsureValidOptionsPassedAcceptsAllFlagsSimultaneously(t *testing.T) {
	t.Parallel()
	testConfigWithAllSelectionCriteria := &GitXargsConfig{
		BranchName:    "test-branch",
		ReposFile:     "./my-repos.txt",
		RepoSlice:     []string{"gruntwork-io/cloud-nuke", "gruntwork-io/fetch"},
		GithubOrg:     "github-org",
		RepoFromStdIn: []string{"gruntwork-io/terragrunt"},
	}

	err := ensureValidOptionsPassed(testConfigWithAllSelectionCriteria)
	assert.NoError(t, err)
}
