package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
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

func cloneLocalTestRepo(t *testing.T) (string, *git.Repository) {
	repo := getMockGithubRepo()

	config := NewGitXargsTestConfig()

	localPath, localRepo, err := cloneLocalRepository(config, repo)

	if err != nil {
		t.Logf("Could not clone local repo to localPath: %s\n", localPath)
		t.Fail()
	}

	return localPath, localRepo
}

func cleanupLocalTestRepo(t *testing.T, localPath string) error {
	removeErr := os.RemoveAll(localPath)
	if removeErr != nil {
		t.Logf("Error cleaning up test repo at path: %s err: %+v\n", localPath, removeErr)
	}
	return removeErr
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
