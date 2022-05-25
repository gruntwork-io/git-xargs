package cmd

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gruntwork-io/git-xargs/auth"
	"github.com/gruntwork-io/git-xargs/common"
	"github.com/gruntwork-io/git-xargs/config"
	gitxargs_io "github.com/gruntwork-io/git-xargs/io"
	"github.com/gruntwork-io/git-xargs/repository"
	"github.com/gruntwork-io/git-xargs/types"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/urfave/cli"
)

// parseGitXargsConfig accepts a urfave cli context and binds its values
// to an internal representation of the data supplied by the user
func parseGitXargsConfig(c *cli.Context) (*config.GitXargsConfig, error) {
	config := config.NewGitXargsConfig()
	config.Draft = c.Bool("draft")
	config.DryRun = c.Bool("dry-run")
	config.SkipPullRequests = c.Bool("skip-pull-requests")
	config.SkipArchivedRepos = c.Bool("skip-archived-repos")
	config.BranchName = c.String("branch-name")
	config.BaseBranchName = c.String("base-branch-name")
	config.CommitMessage = c.String("commit-message")
	config.PullRequestTitle = c.String("pull-request-title")
	config.PullRequestDescription = c.String("pull-request-description")
	config.ReposFile = c.String("repos")
	config.GithubOrg = c.String("github-org")
	config.RepoSlice = c.StringSlice("repo")
	config.MaxConcurrentRepos = c.Int("max-concurrent-repos")
	config.SecondsToSleepBetweenPRs = c.Int("seconds-between-prs")
	config.PullRequestRetries = c.Int("max-pr-retries")
	config.SecondsToSleepWhenRateLimited = c.Int("seconds-to-wait-when-rate-limited")

	// A non-positive ticker value won't work, so set to the default minimum if user passed a bad value
	tickerVal := c.Int("seconds-between-prs")
	if tickerVal < 1 {
		tickerVal = common.DefaultSecondsBetweenPRs
	}

	config.Ticker = time.NewTicker(time.Duration(tickerVal) * time.Second)
	config.Args = c.Args()

	shouldReadStdIn, err := dataBeingPipedToStdIn()
	if err != nil {
		return nil, err
	}
	if shouldReadStdIn {
		repos, err := parseSliceFromStdIn()
		if err != nil {
			return nil, err
		}
		config.RepoFromStdIn = repos
	}

	return config, nil
}

// Return true if there is data being piped to stdin and false otherwise
// Based on https://stackoverflow.com/a/26567513/483528.
func dataBeingPipedToStdIn() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	return stat.Mode()&os.ModeCharDevice == 0, nil
}

// Read the data being passed to stdin and parse it as a slice of strings, where we assume strings are separated by
// whitespace or newlines. All extra whitespace and empty lines are ignored.
func parseSliceFromStdIn() ([]string, error) {
	return parseSliceFromReader(os.Stdin)
}

// Read the data from the given reader and parse it as a slice of strings, where we assume strings are separated by
// whitespace or newlines. All extra whitespace and empty lines are ignored.
func parseSliceFromReader(reader io.Reader) ([]string, error) {
	out := []string{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		for _, word := range words {
			text := strings.TrimSpace(word)
			if text != "" {
				out = append(out, text)
			}
		}
	}

	err := scanner.Err()
	return out, errors.WithStackTrace(err)
}

// handleRepoProcessing encapsulates the main processing logic for the supplied repos and printing the run report that
// is built up throughout the processing
func handleRepoProcessing(config *config.GitXargsConfig) error {
	// Track whether pull requests were skipped
	config.Stats.SetSkipPullRequests(config.SkipPullRequests)

	// Update raw command supplied
	config.Stats.SetCommand(config.Args)

	if err := repository.OperateOnRepos(config); err != nil {
		return err
	}

	// Once all processing is complete, print out the summary of what was done
	config.Stats.PrintReport()

	return nil
}

// sanityCheckInputs performs validation on the user-supplied inputs to ensure we have everything we need:
// 1. An exported GITHUB_OAUTH_TOKEN
// 2. Arguments passed to the binary itself which should be executed against the targeted repos
// 3. At least one of the three valid methods for selecting repositories
func sanityCheckInputs(config *config.GitXargsConfig) error {
	if err := auth.EnsureGithubOauthTokenSet(); err != nil {
		return err
	}

	if len(config.Args) < 1 {
		return errors.WithStackTrace(types.NoArgumentsPassedErr{})
	}

	if err := gitxargs_io.EnsureValidOptionsPassed(config); err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}

// RunGitXargs is the urfave cli app's Action that is called when the user executes the binary
func RunGitXargs(c *cli.Context) error {
	// If someone calls us with no args at all, show the help text and exit
	if !c.Args().Present() {
		return cli.ShowAppHelp(c)
	}

	logger := logging.GetLogger("git-xargs")

	logger.Info("git-xargs running...")

	config, err := parseGitXargsConfig(c)
	if err != nil {
		return err
	}

	if err := sanityCheckInputs(config); err != nil {
		return err
	}

	// If DryRun is enabled, notify user that no file changes will be made
	if config.DryRun {
		logger.Info("Dry run setting enabled. No local branches will be pushed and no PRs will be opened in Github")
	}

	return handleRepoProcessing(config)
}
