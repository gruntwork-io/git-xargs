package repository

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gruntwork-io/git-xargs/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v43/github"
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

// Test that we can execute a script and that the expected stdout and stderr get written to the logger, even if that
// script exits with an error (exit status 1).
func TestExecuteCommandWithLogger(t *testing.T) {
	t.Parallel()

	cfg := config.NewGitXargsConfig()
	cfg.Args = []string{"../data/test/_testscripts/test-stdout-stderr.sh"}
	repo := getMockGithubRepo()

	var buffer bytes.Buffer
	logger := &logrus.Logger{
		Out:       &buffer,
		Level:     logrus.TraceLevel,
		Formatter: new(logrus.TextFormatter),
	}

	err := executeCommandWithLogger(cfg, ".", repo, logger)
	assert.Errorf(t, err, "exit status 1")
	assert.Contains(t, buffer.String(), "Hello, from STDOUT")
	assert.Contains(t, buffer.String(), "Hello, from STDERR")
}

// Test that we can execute a script and that the environment variables are set correctly.
func TestExecuteCommandWithLoggerWithEnvVars(t *testing.T) {
	t.Parallel()

	cfg := config.NewGitXargsConfig()
	cfg.Args = []string{"../data/test/_testscripts/test-env-vars.sh"}
	repo := getMockGithubRepo()

	var buffer bytes.Buffer

	// Test whether the lack of --dry-run sets environment variable correctly
	cfg.DryRun = false

	logger := &logrus.Logger{
		Out:       &buffer,
		Level:     logrus.TraceLevel,
		Formatter: new(logrus.TextFormatter),
	}

	err := executeCommandWithLogger(cfg, ".", repo, logger)
	assert.NoError(t, err)
	assert.Contains(t, buffer.String(), "XARGS_DRY_RUN=false")
	assert.Contains(t, buffer.String(), fmt.Sprintf("XARGS_REPO_NAME=%s", *repo.Name))
	assert.Contains(t, buffer.String(), fmt.Sprintf("XARGS_REPO_OWNER=%s", *repo.Owner.Login))

	// Test whether --dry-run sets environment variable correctly
	cfg.DryRun = true

	logger = &logrus.Logger{
		Out:       &buffer,
		Level:     logrus.TraceLevel,
		Formatter: new(logrus.TextFormatter),
	}

	err = executeCommandWithLogger(cfg, ".", repo, logger)
	assert.NoError(t, err)
	assert.Contains(t, buffer.String(), "XARGS_DRY_RUN=true")
	assert.Contains(t, buffer.String(), fmt.Sprintf("XARGS_REPO_NAME=%s", *repo.Name))
	assert.Contains(t, buffer.String(), fmt.Sprintf("XARGS_REPO_OWNER=%s", *repo.Owner.Login))
}
