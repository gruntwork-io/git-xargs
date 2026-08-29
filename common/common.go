package common

import "github.com/urfave/cli"

const (
	GithubOrgFlagName                    = "github-org"
	DraftPullRequestFlagName             = "draft"
	DryRunFlagName                       = "dry-run"
	SkipPullRequestsFlagName             = "skip-pull-requests"
	SkipArchivedReposFlagName            = "skip-archived-repos"
	RepoFlagName                         = "repo"
	ReposFileFlagName                    = "repos"
	CommitMessageFlagName                = "commit-message"
	BranchFlagName                       = "branch-name"
	BaseBranchFlagName                   = "base-branch-name"
	PullRequestTitleFlagName             = "pull-request-title"
	PullRequestDescriptionFlagName       = "pull-request-description"
	PullRequestReviewersFlagName         = "reviewers"
	PullRequestTeamReviewersFlagName     = "team-reviewers"
	SecondsToWaitBetweenPrsFlagName      = "seconds-between-prs"
	DefaultCommitMessage                 = "git-xargs programmatic commit"
	DefaultPullRequestTitle              = "git-xargs programmatic pull request"
	DefaultPullRequestDescription        = "git-xargs programmatic pull request"
	MaxPullRequestRetriesFlagName        = "max-pr-retries"
	SecondsToWaitWhenRateLimitedFlagName = "seconds-to-wait-when-rate-limited"
	MaxConcurrentClonesFlagName          = "max-concurrent-clones"
	NoSkipCIFlagName                     = "no-skip-ci"
	KeepClonedRepositoriesFlagName       = "keep-cloned-repositories"
	DefaultMaxConcurrentClones           = 4
	DefaultSecondsBetweenPRs             = 1
	DefaultMaxPullRequestRetries         = 3
	DefaultSecondsToWaitWhenRateLimited  = 60
	GithubRepositorySearchFlagName       = "github-repository-search"
	GithubCodeSearchFlagName             = "github-code-search"
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
	GenericBaseBranchFlag = cli.StringFlag{
		Name:  BaseBranchFlagName,
		Usage: "The base branch that changes should be merged into",
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
	GenericPullRequestReviewersFlag = cli.StringSliceFlag{
		Name:  PullRequestReviewersFlagName,
		Usage: "A list of GitHub usernames to request reviews from",
	}
	GenericPullRequestTeamReviewersFlag = cli.StringSliceFlag{
		Name:  PullRequestTeamReviewersFlagName,
		Usage: "A list of GitHub team names to request reviews from",
	}
	GenericSecondsToWaitFlag = cli.IntFlag{
		Name:  SecondsToWaitBetweenPrsFlagName,
		Usage: "The number of seconds to sleep between pull requests in order to respect GitHub API rate limits. Increase this number if you are being rate limited regularly. Defaults to 12 seconds.",
		Value: DefaultSecondsBetweenPRs,
	}
	GenericMaxPullRequestRetriesFlag = cli.IntFlag{
		Name:  MaxPullRequestRetriesFlagName,
		Usage: "The number of times to retry a pull request that failed due to rate limiting. Defaults to 3.",
		Value: DefaultMaxPullRequestRetries,
	}
	GenericSecondsToWaitWhenRateLimitedFlag = cli.IntFlag{
		Name:  SecondsToWaitWhenRateLimitedFlagName,
		Usage: "The number of additional seconds to sleep before attempting to open a PR again, when rate limited by GitHub. Defaults to 60.",
		Value: DefaultSecondsToWaitWhenRateLimited,
	}
	GenericMaxConcurrentClonesFlag = cli.IntFlag{
		Name:  MaxConcurrentClonesFlagName,
		Usage: "The maximum number of concurrent clones to run at once. Defaults to 4. If set to 0 no limit will be applied.",
		Value: DefaultMaxConcurrentClones,
	}
	GenericNoSkipCIFlag = cli.BoolFlag{
		Name:  NoSkipCIFlagName,
		Usage: "By default, git-xargs prepends \"[skip ci]\" to its commit messages. Pass this flag to prevent \"[skip ci]\" from being prepending to commit messages.",
	}
	GenericKeepClonedRepositoriesFlag = cli.BoolFlag{
		Name:  KeepClonedRepositoriesFlagName,
		Usage: "By default, git-xargs deletes the cloned repositories from the temp directory after the command has finished running, to save space on your machine. Pass this flag to prevent git-xargs from deleting the cloned repositories.",
	}
	GenericGithubRepositorySearchFlag = cli.StringFlag{
		Name:  GithubRepositorySearchFlagName,
		Usage: "GitHub repository search query to find repositories (e.g., 'language:go', 'is:private', 'topic:docker'). See GitHub repository search syntax for more options.",
	}
	GenericGithubCodeSearchFlag = cli.StringFlag{
		Name:  GithubCodeSearchFlagName,
		Usage: "GitHub code search query to find repositories containing matching code (e.g., 'path:Dockerfile', 'filename:package.json', 'extension:py print'). Repositories will be extracted from code search results. See GitHub code search syntax for more options.",
	}
)
