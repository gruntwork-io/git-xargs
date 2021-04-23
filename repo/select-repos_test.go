package main

import (
	"github.com/stretchr/testify/require"
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

	require.NoError(t, err)
	require.NotNil(t, repoSelection)
	assert.Equal(t, repoSelection.SelectionType, ExplicitReposOnCommandLine)

	configOrg := NewGitXargsTestConfig()
	configOrg.GithubOrg = "gruntwork-io"

	repoSelectionByOrg, orgErr := selectReposViaInput(configOrg)

	require.NoError(t, orgErr)
	require.NotNil(t, repoSelectionByOrg)
	assert.Equal(t, repoSelectionByOrg.SelectionType, GithubOrganization)

	configStdin := NewGitXargsTestConfig()
	configStdin.RepoFromStdIn = []string{"gruntwork-io/terratest", "gruntwork-io/cloud-nuke"}

	repoSelectionByStdin, stdInErr := selectReposViaInput(configStdin)

	require.NoError(t, stdInErr)
	require.NotNil(t, repoSelectionByStdin)
	assert.Equal(t, repoSelectionByStdin.SelectionType, ReposViaStdIn)
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

	config := NewGitXargsTestConfig()

	config.GithubOrg = "gruntwork-io"
	config.ReposFile = "repos.txt"
	config.RepoSlice = []string{"github.com/gruntwork-io/fetch", "github.com/gruntwork-io/cloud-nuke"}
	config.RepoFromStdIn = []string{"github.com/gruntwork-io/terragrunt", "github.com/gruntwork-io/terratest"}

	assert.Equal(t, GithubOrganization, getPreferredOrderOfRepoSelections(config))

	config.GithubOrg = ""

	assert.Equal(t, ReposFilePath, getPreferredOrderOfRepoSelections(config))

	config.ReposFile = ""

	assert.Equal(t, ExplicitReposOnCommandLine, getPreferredOrderOfRepoSelections(config))

	config.RepoSlice = []string{}

	assert.Equal(t, ReposViaStdIn, getPreferredOrderOfRepoSelections(config))
}
