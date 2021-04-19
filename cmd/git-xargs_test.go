package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A smoke test that you can define a basic config and pass it all the way through the main processing routine without
// any errors
func TestHandleRepoProcessing(t *testing.T) {
	t.Parallel()

	config := NewGitXargsTestConfig()
	config.ReposFile = "./_testdata/good-test-repos.txt"
	config.BranchName = "test-branch-name"
	config.CommitMessage = "test-commit-name"
	config.Args = []string{"touch", "test.txt"}
	config.GithubClient = configureMockGithubClient()
	err := handleRepoProcessing(config)

	assert.NoError(t, err)
}
