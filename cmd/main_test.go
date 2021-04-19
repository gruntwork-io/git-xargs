package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestSetupApp(t *testing.T) {
	app := setupApp()
	assert.NotNil(t, app)
}

func TestGitXargsRejectsEmptyArgs(t *testing.T) {
	app := setupApp()

	emptyFlagSet := flag.NewFlagSet("git-xargs-test", flag.ContinueOnError)
	emptyTestContext := cli.NewContext(app, emptyFlagSet, nil)

	err := runGitXargs(emptyTestContext)

	assert.Error(t, err)
}
