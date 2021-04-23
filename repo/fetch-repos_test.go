package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetFileDefinedRepos provides static allowedRepos input to the getFileDefined repos, ensuring that method returns
// all valid repos passed to it
func TestGetFileDefinedRepos(t *testing.T) {
	t.Parallel()

	config := NewGitXargsTestConfig()
	config.GithubClient = configureMockGithubClient()

	allowedRepos := []*AllowedRepo{
		&AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "cloud-nuke",
		},
		&AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "fetch",
		},
		&AllowedRepo{
			Organization: "gruntwork-io",
			Name:         "terratest",
		},
	}

	githubRepos, reposLookupErr := getFileDefinedRepos(config.GithubClient, allowedRepos, config.Stats)

	assert.Equal(t, len(githubRepos), len(mockGithubRepositories))
	assert.NoError(t, reposLookupErr)
}

// TestGetReposByOrg ensures that you can pass a configuration specifying repo look up by Github Org to getReposByOrg
func TestGetReposByOrg(t *testing.T) {
	t.Parallel()

	config := NewGitXargsTestConfig()
	config.GithubOrg = "gruntwork-io"
	config.GithubClient = configureMockGithubClient()

	githubRepos, reposByOrgLookupErr := getReposByOrg(config)

	assert.Equal(t, len(githubRepos), len(mockGithubRepositories))
	assert.NoError(t, reposByOrgLookupErr)
}
