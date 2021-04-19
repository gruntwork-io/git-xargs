package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigureGithubClient performs a sanity check that you can configure a production Github API client
func TestConfigureGithubClient(t *testing.T) {
	t.Parallel()

	client := configureGithubClient()
	assert.NotNil(t, client)
}

// TestNoGithubOauthTokenPassed temporarily drops the existing GITHUB_OAUTH_TOKEN env var to ensure that the validation
// code throws an error when it is mising. It then replaces it. This is therefore the one test that cannot be run in
// parallel.
func TestNoGithubOAuthTokenPassed(t *testing.T) {
	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	defer os.Setenv("GITHUB_OAUTH_TOKEN", token)

	os.Setenv("GITHUB_OAUTH_TOKEN", "")

	err := ensureGithubOauthTokenSet()
	assert.Error(t, err)
}
