package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigureGithubClient performs a sanity check that you can configure a production GitHub API client
func TestConfigureGithubClient(t *testing.T) {
	t.Parallel()

	client := ConfigureGithubClient()
	assert.NotNil(t, client)
}

// TestNoGithubOauthTokenPassed temporarily drops the existing GITHUB_OAUTH_TOKEN env var to ensure that the validation
// code throws an error when it is missing. It then replaces it. This is therefore the one test that cannot be run in
// parallel.
func TestNoGithubOAuthTokenPassed(t *testing.T) {
	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	defer os.Setenv("GITHUB_OAUTH_TOKEN", token)

	os.Setenv("GITHUB_OAUTH_TOKEN", "")

	err := EnsureGithubOauthTokenSet()
	assert.Error(t, err)
}
