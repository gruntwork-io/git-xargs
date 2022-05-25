package main

import (
	"github.com/gruntwork-io/git-xargs/cmd"
	"github.com/gruntwork-io/git-xargs/common"
	"github.com/gruntwork-io/go-commons/entrypoint"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// VERSION is set at build time using -ldflags parameters. For example, we typically set this flag in circle.yml
// to the latest Git tag when building our Go apps:
//
// build-go-binaries --app-name my-app --dest-path bin --ld-flags "-X main.VERSION=$CIRCLE_TAG"
//
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string

var (
	LogLevelFlag = cli.StringFlag{
		Name:  "loglevel",
		Value: logrus.InfoLevel.String(),
	}
)

// initCli initializes the CLI app before any command is actually executed. This function will handle all the setup
// code, such as setting up the logger with the appropriate log level.
func initCli(cliContext *cli.Context) error {
	// Set logging level
	logLevel := cliContext.String(LogLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return errors.WithStackTrace(err)
	}
	logging.SetGlobalLogLevel(level)
	return nil
}

func setupApp() *cli.App {
	app := entrypoint.NewApp()
	entrypoint.HelpTextLineWidth = 120

	// Override the CLI FlagEnvHinter, so it only returns the Usage text of the Flag and doesn't append the envVar text. Original func https://github.com/urfave/cli/blob/master/flag.go#L652
	cli.FlagEnvHinter = func(envVar, str string) string {
		return str
	}

	app.Name = "git-xargs"
	app.Author = "Gruntwork <www.gruntwork.io>"

	app.Description = "git-xargs is a command-line tool (CLI) for making updates across multiple Github repositories with a single command."

	// Set the version number from your app from the VERSION variable that is passed in at build time
	app.Version = VERSION

	app.EnableBashCompletion = true

	app.Before = initCli

	app.Flags = []cli.Flag{
		LogLevelFlag,
		common.GenericGithubOrgFlag,
		common.GenericDraftPullRequestFlag,
		common.GenericDryRunFlag,
		common.GenericSkipPullRequestFlag,
		common.GenericSkipArchivedReposFlag,
		common.GenericRepoFlag,
		common.GenericRepoFileFlag,
		common.GenericBranchFlag,
		common.GenericBaseBranchFlag,
		common.GenericCommitMessageFlag,
		common.GenericPullRequestTitleFlag,
		common.GenericPullRequestDescriptionFlag,
		common.GenericMaxConcurrentReposFlag,
		common.GenericSecondsToWaitFlag,
		common.GenericMaxPullRequestRetriesFlag,
		common.GenericSecondsToWaitWhenRateLimitedFlag,
	}

	app.Action = cmd.RunGitXargs

	return app
}

// main should only setup the CLI flags and help texts.
func main() {
	app := setupApp()

	entrypoint.RunApp(app)
}
