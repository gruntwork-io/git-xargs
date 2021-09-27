package config

import (
	"fmt"

	"github.com/gruntwork-io/git-xargs/auth"
	"github.com/gruntwork-io/git-xargs/common"
	"github.com/gruntwork-io/git-xargs/local"
	"github.com/gruntwork-io/git-xargs/stats"
	"github.com/gruntwork-io/git-xargs/util"
)

// GitXargsConfig is the internal representation of a given git-xargs run as specified by the user
type GitXargsConfig struct {
	DryRun                 bool
	SkipPullRequests       bool
	SkipArchivedRepos      bool
	MaxConcurrentRepos     int
	BranchName             string
	CommitMessage          string
	PullRequestTitle       string
	PullRequestDescription string
	ReposFile              string
	GithubOrg              string
	RepoSlice              []string
	RepoFromStdIn          []string
	Args                   []string
	GithubClient           auth.GithubClient
	GitClient              local.GitClient
	Stats                  *stats.RunStats
}

// NewGitXargsConfig sets reasonable defaults for a GitXargsConfig and returns a pointer to the config
func NewGitXargsConfig() *GitXargsConfig {
	return &GitXargsConfig{
		DryRun:                 false,
		SkipPullRequests:       false,
		SkipArchivedRepos:      false,
		MaxConcurrentRepos:     0,
		BranchName:             "",
		CommitMessage:          common.DefaultCommitMessage,
		PullRequestTitle:       common.DefaultPullRequestTitle,
		PullRequestDescription: common.DefaultPullRequestDescription,
		ReposFile:              "",
		GithubOrg:              "",
		RepoSlice:              []string{},
		RepoFromStdIn:          []string{},
		Args:                   []string{},
		GithubClient:           auth.ConfigureGithubClient(),
		GitClient:              local.NewGitClient(local.GitProductionProvider{}),
		Stats:                  stats.NewStatsTracker(),
	}
}

func NewGitXargsTestConfig() *GitXargsConfig {

	config := NewGitXargsConfig()

	uniqueID := util.RandStringBytes(9)
	config.BranchName = fmt.Sprintf("test-branch-%s", uniqueID)
	config.CommitMessage = fmt.Sprintf("commit-message-%s", uniqueID)
	config.GitClient = local.NewGitClient(local.MockGitProvider{})

	return config
}
