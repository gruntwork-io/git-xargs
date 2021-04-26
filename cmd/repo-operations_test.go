package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/google/go-github/v32/github"
)

func getMockGithubRepo() *github.Repository {
	userLogin := "gruntwork-io"
	user := &github.User{
		Login: &userLogin,
	}

	repoName := "terragrunt"
	cloneURL := "https://github.com/gruntwork-io/terragrunt"

	repo := &github.Repository{
		Owner:    user,
		Name:     &repoName,
		CloneURL: &cloneURL,
	}

	return repo
}

// Test that we can execute a script and that the expected stdout and stderr get written to the logger
func TestExecuteCommandWithLogger(t *testing.T) {
	t.Parallel()

	cfg := NewGitXargsConfig()
	cfg.Args = []string{"./_testscripts/test-stdout-stderr.rb"}

	repo := getMockGithubRepo()

	var buffer bytes.Buffer
	logger := &logrus.Logger{
		Out:       &buffer,
		Level:     logrus.TraceLevel,
		Formatter: new(logrus.TextFormatter),
	}

	err := executeCommandWithLogger(cfg, ".", repo, logger)
	require.NoError(t, err)
	require.Contains(t, buffer.String(), "Hello, from STDOUT")
	require.Contains(t, buffer.String(), "Hello, from STDERR")
}
