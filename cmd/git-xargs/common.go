package main

import "github.com/urfave/cli"

const (
	GithubOrgFlagName              = "github-org"
	DryRunFlagName                 = "dry-run"
	SkipPullRequestsFlagName       = "skip-pull-requests"
	RepoFlagName                   = "repo"
	ReposFileFlagName              = "repos"
	CommitMessageFlagName          = "commit-message"
	BranchFlagName                 = "branch-name"
	PullRequestTitleFlagName       = "pull-request-title"
	PullRequestDescriptionFlagName = "pull-request-description"
	DefaultCommitMessage           = "git-xargs programmatic commit"
	DefaultPullRequestTitle        = "git-xargs programmatic pull request"
	DefaultPullRequestDescription  = "git-xargs programmatic pull request"
)

var (
	genericGithubOrgFlag = cli.StringFlag{
		Name:  GithubOrgFlagName,
		Usage: "The Github organization to fetch all repositories from.",
	}
	genericDryRunFlag = cli.BoolFlag{
		Name:  DryRunFlagName,
		Usage: "When dry-run is set to true, no local branch changes will pushed and no pull requests will be opened.",
	}
	genericSkipPullRequestFlag = cli.BoolFlag{
		Name:  SkipPullRequestsFlagName,
		Usage: "When skip-pull-requests is set to true, no pull requests will be opened. All changes will be committed and pushed to the specified branch directly.",
	}
	genericRepoFlag = cli.StringSliceFlag{
		Name:  RepoFlagName,
		Usage: "A single repo name to run the command on in the format of <github-organization/repo-name>. Can be invoked multiple times with different repo names",
	}
	genericRepoFileFlag = cli.StringFlag{
		Name:  ReposFileFlagName,
		Usage: "The path to a file containing repos, one per line in the format of <github-organization/repo-name>",
	}
	genericBranchFlag = cli.StringFlag{
		Name:     BranchFlagName,
		Usage:    "The name of the branch on which changes will be made",
	}
	genericCommitMessageFlag = cli.StringFlag{
		Name:  CommitMessageFlagName,
		Usage: "The commit message to use when creating commits from changes introduced by your command or script",
		Value: DefaultCommitMessage,
	}
	genericPullRequestTitleFlag = cli.StringFlag{
		Name:  PullRequestTitleFlagName,
		Usage: "The title to add to pull requests opened by git-xargs",
		Value: DefaultPullRequestTitle,
	}
	genericPullRequestDescriptionFlag = cli.StringFlag{
		Name:  PullRequestDescriptionFlagName,
		Usage: "The description to add to pull requests opened by git-xargs",
		Value: DefaultPullRequestDescription,
	}
)
