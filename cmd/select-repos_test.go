package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSelectReposViaInput ensures the selectReposViaInput function correctly returns the correct repo target type
// given the 3 different ways to target repos for processing
func TestSelectReposViaInput(t *testing.T) {
	t.Parallel()

	config := NewGitXargsTestConfig()
	config.RepoSlice = []string{"gruntwork-io/terratest", "gruntwork-io/cloud-nuke"}

	repoSelection, err := selectReposViaInput(config)

	assert.NotNil(t, repoSelection)
	assert.Equal(t, repoSelection.SelectionType, ExplicitReposOnCommandLine)
	assert.NoError(t, err)

	configOrg := NewGitXargsTestConfig()
	configOrg.GithubOrg = "gruntwork-io"

	repoSelectionByOrg, orgErr := selectReposViaInput(configOrg)

	assert.NotNil(t, repoSelectionByOrg)
	assert.Equal(t, repoSelectionByOrg.SelectionType, GithubOrganization)
	assert.NoError(t, orgErr)
}

// TestOperateOnRepos smoke tests the OperateOnRepos method
func TestOperateOnRepos(t *testing.T) {
	t.Parallel()

	config := NewGitXargsTestConfig()
	config.GithubOrg = "gruntwork-io"
	config.GithubClient = configureMockGithubClient()

	err := OperateOnRepos(config)
	assert.NoError(t, err)

	configReposOnCommandLine := NewGitXargsTestConfig()
	configReposOnCommandLine.GithubClient = configureMockGithubClient()

	configReposOnCommandLine.RepoSlice = []string{"gruntwork-io/fetch", "gruntwork-io/cloud-nuke"}

	cmdLineErr := OperateOnRepos(configReposOnCommandLine)
	assert.NoError(t, cmdLineErr)
}

// TestGetpreferredOrderOfRepoSelections ensures the getPreferredOrderOfRepoSelections returns the expected method
// for fetching repos given the three possible means of targeting repositories for processing
func TestGetPreferredOrderOfRepoSelections(t *testing.T) {
	t.Parallel()

	configReposOnCommandLine := NewGitXargsTestConfig()

	// Ensure that, for a config with one or more repos defined on the command
	// line via --repo,
	// ExplicitReposOnCommandLine is returned
	configReposOnCommandLine.RepoSlice = []string{"github.com/gruntwork-io/fetch", "github.com/gruntwork-io/cloud-nuke"}

	selectionCriteria := getPreferredOrderOfRepoSelections(configReposOnCommandLine)
	assert.Equal(t, selectionCriteria, ExplicitReposOnCommandLine)

	configReposFile := NewGitXargsTestConfig()
	configReposFile.ReposFile = "./my-test-repos.txt"

	selectionCriteria2 := getPreferredOrderOfRepoSelections(configReposFile)
	assert.Equal(t, selectionCriteria2, ReposFilePath)

	configGithubOrg := NewGitXargsTestConfig()
	configGithubOrg.GithubOrg = "gruntwork-io"

	selectionCriteria3 := getPreferredOrderOfRepoSelections(configGithubOrg)
	assert.Equal(t, selectionCriteria3, GithubOrganization)
}
