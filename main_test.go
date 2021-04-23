package main

import (
	"flag"
	"strings"
	"testing"

	"github.com/gruntwork-io/git-xargs/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestSetupApp(t *testing.T) {
	app := setupApp()
	assert.NotNil(t, app)
}

func TestGitXargsShowsHelpTextForEmptyArgs(t *testing.T) {
	app := setupApp()

	// Capture the app's stdout
	var stdout strings.Builder
	app.Writer = &stdout

	emptyFlagSet := flag.NewFlagSet("git-xargs-test", flag.ContinueOnError)
	emptyTestContext := cli.NewContext(app, emptyFlagSet, nil)

	err := cmd.RunGitXargs(emptyTestContext)

	// Make sure we see the help text
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), app.Description)
}
