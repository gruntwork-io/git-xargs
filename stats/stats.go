package stats

import (
	"sync"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/gruntwork-io/git-xargs/printer"
	"github.com/gruntwork-io/git-xargs/types"
)

const (
	// DryRunSet denotes a repo will not have any file changes, branches made or PRs opened because the dry-run flag was set to true
	DryRunSet types.Event = "dry-run-set-no-changes-made"
	// ReposSelected denotes all the repositories that were targeted for processing by this tool AFTER filtering was applied to determine valid repos
	ReposSelected types.Event = "repos-selected-pre-processing"
	// ReposArchivedSkipped denotes all the repositories that were skipped from the list of repos to clone because the skip-archived-repos was set to true
	ReposArchivedSkipped types.Event = "repos-archived-skipped"
	// TargetBranchNotFound denotes the special branch used by this tool to make changes on was not found on lookup, suggesting it should be created
	TargetBranchNotFound types.Event = "target-branch-not-found"
	// TargetBranchAlreadyExists denotes the special branch used by this tool was already found (so it was likely already created by a previous run)
	TargetBranchAlreadyExists types.Event = "target-branch-already-exists"
	// TargetBranchLookupErr denotes an issue performing the lookup via GitHub API for the target branch - an API call failure
	TargetBranchLookupErr types.Event = "target-branch-lookup-err"
	// TargetBranchSuccessfullyCreated denotes a repo for which the target branch was created via GitHub API call
	TargetBranchSuccessfullyCreated types.Event = "target-branch-successfully-created"
	// FetchedViaGithubAPI denotes a repo was successfully listed by calling the GitHub API
	FetchedViaGithubAPI types.Event = "fetch-via-github-api"
	// RepoSuccessfullyCloned denotes a repo that was cloned to the local filesystem of the operator's machine
	RepoSuccessfullyCloned types.Event = "repo-successfully-cloned"
	// RepoFailedToClone denotes that for whatever reason we were unable to clone the repo to the local system
	RepoFailedToClone types.Event = "repo-failed-to-clone"
	// BranchCheckoutFailed denotes a failure to checkout a new tool specific branch in the given repo
	BranchCheckoutFailed types.Event = "branch-checkout-failed"
	// GetHeadRefFailed denotes a repo for which the HEAD git reference could not be obtained
	GetHeadRefFailed types.Event = "get-head-ref-failed"
	// CommandErrorOccurredDuringExecution denotes a repo for which the supplied command failed to be executed
	CommandErrorOccurredDuringExecution types.Event = "command-error-during-execution"
	// WorktreeStatusCheckFailed denotes a repo whose git status command failed post command execution
	WorktreeStatusCheckFailed types.Event = "worktree-status-check-failed"
	// WorktreeStatusCheckFailedCommand denotes a repo whose git status command failed following command execution
	WorktreeStatusCheckFailedCommand = "worktree-status-check-failed-command"
	// WorktreeStatusDirty denotes repos that had local file changes following execution of all their targeted
	WorktreeStatusDirty types.Event = "worktree-status-dirty"
	// WorktreeStatusClean denotes a repo that did not have any local file changes following command execution
	WorktreeStatusClean types.Event = "worktree-status-clean"
	// WorktreeAddFileFailed denotes a failure to add at least one file to the git stage following command execution
	WorktreeAddFileFailed types.Event = "worktree-add-file-failed"
	// CommitChangesFailed denotes an error git committing our file changes to the local repo
	CommitChangesFailed types.Event = "commit-changes-failed"
	// PushBranchFailed denotes a repo whose new tool-specific branch could not be pushed to remote origin
	PushBranchFailed types.Event = "push-branch-failed"
	// PushBranchSkipped denotes a repo whose local branch was not pushed due to the --dry-run flag being set
	PushBranchSkipped types.Event = "push-branch-skipped"
	// RepoNotExists denotes a repo + org combo that was supplied via file but could not be successfully looked up via the GitHub API (returned a 404)
	RepoNotExists types.Event = "repo-not-exists"
	// PullRequestOpenErr denotes a repo whose pull request containing config changes could not be made successfully
	PullRequestOpenErr types.Event = "pull-request-open-error"
	// PullRequestAlreadyExists denotes a repo where the pull request already exists for the requested branch, so we didn't open a new one
	PullRequestAlreadyExists types.Event = "pull-request-already-exists"
	// CommitsMadeDirectlyToBranch denotes a repo whose local worktree changes were committed directly to the specified branch because the --skip-pull-requests flag was passed
	CommitsMadeDirectlyToBranch types.Event = "commits-made-directly-to-branch"
	//DirectCommitsPushedToRemoteBranch denotes a repo whose changes were pushed to the remote specified branch because the --skip-pull-requests flag was passed
	DirectCommitsPushedToRemoteBranch types.Event = "direct-commits-pushed-to-remote"
	// BranchRemotePullFailed denotes a repo whose remote branch could not be fetched successfully
	BranchRemotePullFailed types.Event = "branch-remote-pull-failed"
	// BranchRemoteDidntExistYet denotes a repo whose specified branch didn't exist remotely yet and so was just created locally to begin with
	BranchRemoteDidntExistYet types.Event = "branch-remote-didnt-exist-yet"
	// RepoFlagSuppliedRepoMalformed denotes a repo passed via the --repo flag that was malformed (perhaps missing it's Github org prefix) and therefore unprocessable
	RepoFlagSuppliedRepoMalformed types.Event = "repo-flag-supplied-repo-malformed"
	// RepoDoesntSupportDraftPullRequestsErr denotes a repo that is incompatible with the submitted pull request configuration
	RepoDoesntSupportDraftPullRequestsErr types.Event = "repo-not-compatible-with-pull-config"
	// BaseBranchTargetInvalidErr denotes a repo that does not have the base branch specified by the user
	BaseBranchTargetInvalidErr types.Event = "base-branch-target-invalid"
	// PRFailedDueToRateLimits denotes a repo whose initial pull request failed as a result of being rate limited by GitHub
	PRFailedDueToRateLimitsErr types.Event = "pr-failed-due-to-rate-limits"
	//PRFailedAfterMaximumRetriesErr denotes a repo whose pull requests all failed to be created via GitHub following the maximum number of retries
	PRFailedAfterMaximumRetriesErr types.Event = "pr-failed-after-maximum-retries"
)

var allEvents = []types.AnnotatedEvent{
	{Event: FetchedViaGithubAPI, Description: "Repos successfully fetched via Github API"},
	{Event: DryRunSet, Description: "Repos that were not modified in any way because this was a dry-run"},
	{Event: ReposSelected, Description: "All repos that were targeted for processing AFTER filtering missing / malformed repos"},
	{Event: ReposArchivedSkipped, Description: "All repos that were filtered out with the --skip-archived-repos flag"},
	{Event: TargetBranchNotFound, Description: "Repos whose target branch was not found"},
	{Event: TargetBranchAlreadyExists, Description: "Repos whose target branch already existed"},
	{Event: TargetBranchLookupErr, Description: "Repos whose target branches could not be looked up due to an API error"},
	{Event: RepoSuccessfullyCloned, Description: "Repos that were successfully cloned to the local filesystem"},
	{Event: RepoFailedToClone, Description: "Repos that were unable to be cloned to the local filesystem"},
	{Event: BranchCheckoutFailed, Description: "Repos for which checking out a new tool-specific branch failed"},
	{Event: GetHeadRefFailed, Description: "Repos for which the HEAD git reference could not be obtained"},
	{Event: CommandErrorOccurredDuringExecution, Description: "Repos for which the supplied command raised an error during execution"},
	{Event: WorktreeStatusCheckFailed, Description: "Repos for which the git status command failed following command execution"},
	{Event: WorktreeStatusDirty, Description: "Repos that showed file changes to their working directory following command execution"},
	{Event: WorktreeStatusClean, Description: "Repos that showed NO file changes to their working directory following command execution"},
	{Event: CommitChangesFailed, Description: "Repos whose file changes failed to be committed for some reason"},
	{Event: PushBranchFailed, Description: "Repos whose tool-specific branch containing changes failed to push to remote origin"},
	{Event: PushBranchSkipped, Description: "Repos whose local branch was not pushed because the --dry-run flag was set"},
	{Event: RepoNotExists, Description: "Repos that were supplied by user but don't exist (404'd) via Github API"},
	{Event: PullRequestOpenErr, Description: "Repos against which pull requests failed to be opened"},
	{Event: PullRequestAlreadyExists, Description: "Repos where opening a pull request was skipped because a pull request was already open"},
	{Event: CommitsMadeDirectlyToBranch, Description: "Repos whose local changes were committed directly to the specified branch because --skip-pull-requests was passed"},
	{Event: DirectCommitsPushedToRemoteBranch, Description: "Repos whose changes were pushed directly to the remote branch because --skip-pull-requests was passed"},
	{Event: BranchRemotePullFailed, Description: "Repos whose remote branches could not be successfully pulled"},
	{Event: BranchRemoteDidntExistYet, Description: "Repos whose specified branches did not exist on the remote, and so were first created locally"},
	{Event: RepoFlagSuppliedRepoMalformed, Description: "Repos passed via the --repo flag that were malformed (missing their Github org prefix?) and therefore unprocessable"},
	{Event: RepoDoesntSupportDraftPullRequestsErr, Description: "Repos that do not support Draft PRs (--draft flag was passed)"},
	{Event: BaseBranchTargetInvalidErr, Description: "Repos that did not have the branch specified by --base-branch-name"},
	{Event: PRFailedDueToRateLimitsErr, Description: "Repos whose initial Pull Request failed to be created due to GitHub rate limits"},
	{Event: PRFailedAfterMaximumRetriesErr, Description: "Repos whose Pull Request failed to be created after the maximum number of retries"},
}

// RunStats will be a stats-tracker class that keeps score of which repos were touched, which were considered for update, which had branches made, PRs made, which were missing workflows or contexts, or had out of date workflows syntax values, etc
type RunStats struct {
	selectionMode         string
	repos                 map[types.Event][]*github.Repository
	skippedArchivedRepos  map[types.Event][]*github.Repository
	pulls                 map[string]string
	draftpulls            map[string]string
	command               []string
	fileProvidedRepos     []*types.AllowedRepo
	repoFlagProvidedRepos []*types.AllowedRepo
	startTime             time.Time
	skipPullRequests      bool
	mutex                 *sync.Mutex
}

// NewStatsTracker initializes a tracker struct that is capable of keeping tabs on which repos were handled and how
func NewStatsTracker() *RunStats {
	var fileProvidedRepos []*types.AllowedRepo
	var repoFlagProvidedRepos []*types.AllowedRepo

	t := &RunStats{
		repos:                 make(map[types.Event][]*github.Repository),
		skippedArchivedRepos:  make(map[types.Event][]*github.Repository),
		pulls:                 make(map[string]string),
		draftpulls:            make(map[string]string),
		command:               []string{},
		fileProvidedRepos:     fileProvidedRepos,
		repoFlagProvidedRepos: repoFlagProvidedRepos,
		startTime:             time.Now(),
		skipPullRequests:      false,
		mutex:                 &sync.Mutex{},
	}
	return t
}

// SetSelectionMode accepts a string representing the method by which repos were selected for this run - in order to print a human-legible description in the final report
func (r *RunStats) SetSelectionMode(mode string) {
	r.selectionMode = mode
}

// GetSelectionMode returns the currently set repo selection method
func (r *RunStats) GetSelectionMode() string {
	return r.selectionMode
}

// GetTotalRunSeconds returns the total time it took, in seconds, to run all the selected commands against all the targeted repos
func (r *RunStats) GetTotalRunSeconds() int {
	s := time.Since(r.startTime).Seconds()
	return int(s) % 60
}

// GetRepos returns the inner map of events to *github.Repositories that the RunStats maintains throughout the lifecycle of a given command run
func (r *RunStats) GetRepos() map[types.Event][]*github.Repository {
	return r.repos
}

// GetSkippedArchivedRepos returns the inner map of events to *github.Repositories that are excluded from the targeted repos list
func (r *RunStats) GetSkippedArchivedRepos() map[types.Event][]*github.Repository {
	return r.skippedArchivedRepos
}

// GetPullRequests returns the inner representation of the pull requests that were opened during the lifecycle of a given run
func (r *RunStats) GetPullRequests() map[string]string {
	return r.pulls
}

// GetDraftPullRequests returns the inner representation of the draft pull requests that were opened during the lifecycle of a given run
func (r *RunStats) GetDraftPullRequests() map[string]string {
	return r.draftpulls
}

// SetFileProvidedRepos sets the number of repos that were provided via file by the user on startup (as opposed to looked up via GitHub API via the --github-org flag)
func (r *RunStats) SetFileProvidedRepos(fileProvidedRepos []*types.AllowedRepo) {
	for _, ar := range fileProvidedRepos {
		r.fileProvidedRepos = append(r.fileProvidedRepos, ar)
	}
}

// GetFileProvidedRepos returns a slice of the repos that were provided via the --repos flag (as opposed to looked up via the GitHub API via the --github-org flag)
func (r *RunStats) GetFileProvidedRepos() []*types.AllowedRepo {
	return r.fileProvidedRepos
}

// SetRepoFlagProvidedRepos sets the number of repos that were provided via a single or multiple invocations of the --repo flag
func (r *RunStats) SetRepoFlagProvidedRepos(repoFlagProvidedRepos []*types.AllowedRepo) {
	for _, ar := range repoFlagProvidedRepos {
		r.repoFlagProvidedRepos = append(r.repoFlagProvidedRepos, ar)
	}
}

// SetSkipPullRequests tracks whether the user specified that pull requests should be skipped (in favor of committing and pushing directly to the specified branch)
func (r *RunStats) SetSkipPullRequests(skipPullRequests bool) {
	r.skipPullRequests = skipPullRequests
}

// SetCommand sets the user-supplied command to be run against the targeted repos
func (r *RunStats) SetCommand(c []string) {
	r.command = c
}

// GetMultiple returns the slice of pointers to GitHub repositories filed under the provided event's key
func (r *RunStats) GetMultiple(event types.Event) []*github.Repository {
	return r.repos[event]
}

// TrackSingle accepts a types.Event to associate with the supplied repo so that a final report can be generated at the end of each run
func (r *RunStats) TrackSingle(event types.Event, repo *github.Repository) {
	// TrackSingle is called from multiple concurrent writing goroutines, so we need to lock access to the underlying map
	defer r.mutex.Unlock()
	r.mutex.Lock()
	r.repos[event] = TrackEventIfMissing(r.repos[event], repo)
}

// TrackEventIfMissing prevents the addition of duplicates to the tracking slices. Repos may end up with file changes
// for example, from multiple command runs, so we don't need the same repo repeated multiple times in the final report
func TrackEventIfMissing(slice []*github.Repository, repo *github.Repository) []*github.Repository {
	for _, existingRepo := range slice {
		if existingRepo.GetName() == repo.GetName() {
			// We've already tracked this repo under this event, return the existing slice to avoid adding
			// it a second time
			return slice
		}
	}
	return append(slice, repo)
}

// TrackPullRequest stores the successful PR opening for the supplied Repo, at the supplied PR URL
// This function is safe to call from concurrent goroutines
func (r *RunStats) TrackPullRequest(repoName, prURL string) {
	defer r.mutex.Unlock()
	r.mutex.Lock()
	r.pulls[repoName] = prURL
}

// TrackDraftPullRequest stores the successful Draft PR opening for the supplied Repo, at the supplied PR URL
// This function is safe to call from concurrent goroutines
func (r *RunStats) TrackDraftPullRequest(repoName, prURL string) {
	defer r.mutex.Unlock()
	r.mutex.Lock()
	r.draftpulls[repoName] = prURL
}

// TrackMultiple accepts a types.Event and a slice of pointers to GitHub repos that will all be associated with that event
func (r *RunStats) TrackMultiple(event types.Event, repos []*github.Repository) {
	for _, repo := range repos {
		r.TrackSingle(event, repo)
	}
}

// GenerateRunReport creates a struct that contains all the information necessary to print a final summary report
func (r *RunStats) GenerateRunReport() *types.RunReport {
	return &types.RunReport{
		Repos:          r.GetRepos(),
		SkippedRepos:   r.GetSkippedArchivedRepos(),
		Command:        r.command,
		SelectionMode:  r.selectionMode,
		RuntimeSeconds: r.GetTotalRunSeconds(), FileProvidedRepos: r.GetFileProvidedRepos(),
		PullRequests:      r.GetPullRequests(),
		DraftPullRequests: r.GetDraftPullRequests(),
	}
}

// PrintReport renders to STDOUT a summary of each repo that was considered by this tool and what happened to it during processing
func (r *RunStats) PrintReport() {
	printer.PrintRepoReport(allEvents, r.GenerateRunReport())
}
