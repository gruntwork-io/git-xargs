package repository

import (
	"testing"

	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/mocks"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

// TestSelectReposViaInput ensures the selectReposViaInput function correctly returns the correct repo target type
// given the 3 different ways to target repos for processing
func TestSelectReposViaInput(t *testing.T) {
	t.Parallel()

	testConfig := config.NewGitXargsTestConfig()
	testConfig.RepoSlice = []string{"gruntwork-io/terratest", "gruntwork-io/cloud-nuke"}

	repoSelection, err := selectReposViaInput(testConfig)

	require.NoError(t, err)
	require.NotNil(t, repoSelection)
	assert.Equal(t, repoSelection.SelectionType, ExplicitReposOnCommandLine)

	configOrg := config.NewGitXargsTestConfig()
	configOrg.GithubOrg = "gruntwork-io"

	repoSelectionByOrg, orgErr := selectReposViaInput(configOrg)

	require.NoError(t, orgErr)
	require.NotNil(t, repoSelectionByOrg)
	assert.Equal(t, repoSelectionByOrg.SelectionType, GithubOrganization)

	configStdin := config.NewGitXargsTestConfig()
	configStdin.RepoFromStdIn = []string{"gruntwork-io/terratest", "gruntwork-io/cloud-nuke"}

	repoSelectionByStdin, stdInErr := selectReposViaInput(configStdin)

	require.NoError(t, stdInErr)
	require.NotNil(t, repoSelectionByStdin)
	assert.Equal(t, repoSelectionByStdin.SelectionType, ReposViaStdIn)
}

// TestOperateOnRepos smoke tests the OperateOnRepos method
func TestOperateOnRepos(t *testing.T) {
	t.Parallel()

	testConfig := config.NewGitXargsTestConfig()
	testConfig.GithubOrg = "gruntwork-io"
	testConfig.GithubClient = mocks.ConfigureMockGithubClient()

	err := OperateOnRepos(testConfig)
	assert.NoError(t, err)

	configReposOnCommandLine := config.NewGitXargsTestConfig()
	configReposOnCommandLine.GithubClient = mocks.ConfigureMockGithubClient()

	configReposOnCommandLine.RepoSlice = []string{"gruntwork-io/fetch", "gruntwork-io/cloud-nuke"}

	cmdLineErr := OperateOnRepos(configReposOnCommandLine)
	assert.NoError(t, cmdLineErr)
}

// TestGetPreferredOrderOfRepoSelections ensures the getPreferredOrderOfRepoSelections returns the expected method
// for fetching repos given the three possible means of targeting repositories for processing
func TestGetPreferredOrderOfRepoSelections(t *testing.T) {
	t.Parallel()

	testConfig := config.NewGitXargsTestConfig()

	testConfig.GithubOrg = "gruntwork-io"
	testConfig.ReposFile = "repos.txt"
	testConfig.RepoSlice = []string{"github.com/gruntwork-io/fetch", "github.com/gruntwork-io/cloud-nuke"}
	testConfig.RepoFromStdIn = []string{"github.com/gruntwork-io/terragrunt", "github.com/gruntwork-io/terratest"}

	assert.Equal(t, GithubOrganization, getPreferredOrderOfRepoSelections(testConfig))

	testConfig.GithubOrg = ""

	assert.Equal(t, ReposFilePath, getPreferredOrderOfRepoSelections(testConfig))

	testConfig.ReposFile = ""

	assert.Equal(t, ExplicitReposOnCommandLine, getPreferredOrderOfRepoSelections(testConfig))

	testConfig.RepoSlice = []string{}

	assert.Equal(t, ReposViaStdIn, getPreferredOrderOfRepoSelections(testConfig))
}
