package main

import (
	"bufio"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/urfave/cli"
	"io"
	"os"
	"strings"
)

// GitXargsConfig is the internal representation of a given git-xargs run as specified by the user
type GitXargsConfig struct {
	DryRun                 bool
	SkipPullRequests       bool
	BranchName             string
	CommitMessage          string
	PullRequestTitle       string
	PullRequestDescription string
	ReposFile              string
	GithubOrg              string
	RepoSlice              []string
	RepoFromStdIn          []string
	Args                   []string
	GithubClient           GithubClient
	GitClient              GitClient
	Stats                  *RunStats
}

// NewGitXargsConfig sets reasonable defaults for a GitXargsConfig and returns a pointer to a the config
func NewGitXargsConfig() *GitXargsConfig {
	return &GitXargsConfig{
		DryRun:                 false,
		SkipPullRequests:       false,
		BranchName:             "",
		CommitMessage:          DefaultCommitMessage,
		PullRequestTitle:       DefaultPullRequestTitle,
		PullRequestDescription: DefaultPullRequestDescription,
		ReposFile:              "",
		GithubOrg:              "",
		RepoSlice:              []string{},
		RepoFromStdIn:          []string{},
		Args:                   []string{},
		GithubClient:           configureGithubClient(),
		GitClient:              NewGitClient(GitProductionProvider{}),
		Stats:                  NewStatsTracker(),
	}
}

// parseGitXargsConfig accepts a urfave cli context and binds its values
// to an internal representation of the data supplied by the user
func parseGitXargsConfig(c *cli.Context) (*GitXargsConfig, error) {
	config := NewGitXargsConfig()
	config.DryRun = c.Bool("dry-run")
	config.SkipPullRequests = c.Bool("skip-pull-requests")
	config.BranchName = c.String("branch-name")
	config.CommitMessage = c.String("commit-message")
	config.PullRequestTitle = c.String("pull-request-title")
	config.PullRequestDescription = c.String("pull-request-description")
	config.ReposFile = c.String("repos")
	config.GithubOrg = c.String("github-org")
	config.RepoSlice = c.StringSlice("repo")
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
func handleRepoProcessing(config *GitXargsConfig) error {
	// Track whether or not pull requests were skipped
	config.Stats.SetSkipPullRequests(config.SkipPullRequests)

	// Update raw command supplied
	config.Stats.SetCommand(config.Args)

	if err := OperateOnRepos(config); err != nil {
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
func sanityCheckInputs(config *GitXargsConfig) error {
	if err := ensureGithubOauthTokenSet(); err != nil {
		return err
	}

	if len(config.Args) < 1 {
		return errors.WithStackTrace(NoArgumentsPassedErr{})
	}

	if err := ensureValidOptionsPassed(config); err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}

// runGitXargs is the urfave cli app's Action that is called when the user executes the binary
func runGitXargs(c *cli.Context) error {
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
