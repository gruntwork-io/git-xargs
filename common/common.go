package common

import "github.com/urfave/cli"

const (
	GithubOrgFlagName              = "github-org"
	DraftPullRequestFlagName       = "draft"
	DryRunFlagName                 = "dry-run"
	SkipPullRequestsFlagName       = "skip-pull-requests"
	SkipArchivedReposFlagName      = "skip-archived-repos"
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
	GenericGithubOrgFlag = cli.StringFlag{
		Name:  GithubOrgFlagName,
		Usage: "The Github organization to fetch all repositories from.",
	}
	GenericDraftPullRequestFlag = cli.BoolFlag{
		Name:  DraftPullRequestFlagName,
		Usage: "Whether to open pull requests in draft mode",
	}
	GenericDryRunFlag = cli.BoolFlag{
		Name:  DryRunFlagName,
		Usage: "When dry-run is set to true, no local branch changes will pushed and no pull requests will be opened.",
	}
	GenericSkipPullRequestFlag = cli.BoolFlag{
		Name:  SkipPullRequestsFlagName,
		Usage: "When skip-pull-requests is set to true, no pull requests will be opened. All changes will be committed and pushed to the specified branch directly.",
	}
	GenericSkipArchivedReposFlag = cli.BoolFlag{
		Name:  SkipArchivedReposFlagName,
		Usage: "Used in conjunction with github-org, will exclude archived repositories.",
	}
	GenericRepoFlag = cli.StringSliceFlag{
		Name:  RepoFlagName,
		Usage: "A single repo name to run the command on in the format of <github-organization/repo-name>. Can be invoked multiple times with different repo names",
	}
	GenericRepoFileFlag = cli.StringFlag{
		Name:  ReposFileFlagName,
		Usage: "The path to a file containing repos, one per line in the format of <github-organization/repo-name>",
	}
	GenericBranchFlag = cli.StringFlag{
		Name:  BranchFlagName,
		Usage: "The name of the branch on which changes will be made",
	}
	GenericCommitMessageFlag = cli.StringFlag{
		Name:  CommitMessageFlagName,
		Usage: "The commit message to use when creating commits from changes introduced by your command or script",
		Value: DefaultCommitMessage,
	}
	GenericPullRequestTitleFlag = cli.StringFlag{
		Name:  PullRequestTitleFlagName,
		Usage: "The title to add to pull requests opened by git-xargs",
		Value: DefaultPullRequestTitle,
	}
	GenericPullRequestDescriptionFlag = cli.StringFlag{
		Name:  PullRequestDescriptionFlagName,
		Usage: "The description to add to pull requests opened by git-xargs",
		Value: DefaultPullRequestDescription,
	}
)
