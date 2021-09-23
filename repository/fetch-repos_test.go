package repository

import (
	"testing"

	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/mocks"
	"github.com/gruntwork-io/git-xargs/types"
	"github.com/stretchr/testify/assert"
)

// TestGetFileDefinedRepos provides static allowedRepos input to the getFileDefined repos, ensuring that method returns
// all valid repos passed to it
func TestGetFileDefinedRepos(t *testing.T) {
	t.Parallel()

	config := config.NewGitXargsTestConfig()
	config.GithubClient = mocks.ConfigureMockGithubClient()

	allowedRepos := []*types.AllowedRepo{
		&types.AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "cloud-nuke",
		},
		&types.AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "fetch",
		},
		&types.AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "terratest",
		},
	}

	githubRepos, reposLookupErr := getFileDefinedRepos(config.GithubClient, allowedRepos, config.Stats)

	assert.Equal(t, len(githubRepos), len(allowedRepos))
	assert.NoError(t, reposLookupErr)
}

// TestGetReposByOrg ensures that you can pass a configuration specifying repo look up by GitHub Org to getReposByOrg
func TestGetReposByOrg(t *testing.T) {
	t.Parallel()

	config := config.NewGitXargsTestConfig()
	config.GithubOrg = "gruntwork-io"
	config.GithubClient = mocks.ConfigureMockGithubClient()

	githubRepos, reposByOrgLookupErr := getReposByOrg(config)

	assert.Equal(t, len(githubRepos), len(mocks.MockGithubRepositories))
	assert.NoError(t, reposByOrgLookupErr)
}

// TestSkipArchivedRepos ensures that you can filter out archived repositories
func TestSkipArchivedRepos(t *testing.T) {
	t.Parallel()

	config := config.NewGitXargsTestConfig()
	config.GithubOrg = "gruntwork-io"
	config.SkipArchivedRepos = true
	config.GithubClient = mocks.ConfigureMockGithubClient()

	githubRepos, reposByOrgLookupErr := getReposByOrg(config)

	assert.Equal(t, len(githubRepos), len(mocks.MockGithubRepositories)-1)
	assert.NoError(t, reposByOrgLookupErr)
}
