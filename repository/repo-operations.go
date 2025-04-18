package repository

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"

	"github.com/google/go-github/v43/github"

	"github.com/gruntwork-io/git-xargs/common"
	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/stats"
	"github.com/gruntwork-io/git-xargs/types"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
)

// cloneLocalRepository clones a remote GitHub repo via SSH to a local temporary directory so that the supplied command
// can be run against the repo locally and any git changes handled thereafter. The local directory has
// git-xargs-<repo-name> appended to it to make it easier to find when you are looking for it while debugging
func cloneLocalRepository(config *config.GitXargsConfig, repo *github.Repository) (string, *git.Repository, error) {
	logger := logging.GetLogger("git-xargs")
	config.CloneJobsLimiter <- struct{}{}

	logger.WithFields(logrus.Fields{
		"Repo": repo.GetName(),
	}).Debug("Attempting to clone repository using GITHUB_OAUTH_TOKEN")

	repositoryDir, tmpDirErr := ioutil.TempDir("", fmt.Sprintf("git-xargs-%s", repo.GetName()))
	if tmpDirErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": tmpDirErr,
			"Repo":  repo.GetName(),
		}).Debug("Failed to create temporary directory to hold repo")
		return repositoryDir, nil, errors.WithStackTrace(tmpDirErr)
	}

	gitProgressBuffer := bytes.NewBuffer(nil)
	localRepository, err := config.GitClient.PlainClone(repositoryDir, false, &git.CloneOptions{
		URL:      repo.GetCloneURL(),
		Progress: gitProgressBuffer,
		Auth: &http.BasicAuth{
			Username: repo.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	})

	logger.WithFields(logrus.Fields{
		"Repo": repo.GetName(),
	}).Debug(gitProgressBuffer)

	<-config.CloneJobsLimiter

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
			"Repo":  repo.GetName(),
		}).Debug("Error cloning repository")

		// Track failure to clone for our final run report
		config.Stats.TrackSingle(stats.RepoFailedToClone, repo)

		return repositoryDir, nil, errors.WithStackTrace(err)
	}

	config.Stats.TrackSingle(stats.RepoSuccessfullyCloned, repo)

	return repositoryDir, localRepository, nil
}

// getLocalRepoHeadRef looks up the HEAD reference of the locally cloned git repository, which is required by
// downstream operations such as branching
func getLocalRepoHeadRef(config *config.GitXargsConfig, localRepository *git.Repository, repo *github.Repository) (*plumbing.Reference, error) {
	logger := logging.GetLogger("git-xargs")

	ref, headErr := localRepository.Head()
	if headErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": headErr,
			"Repo":  repo.GetName(),
		}).Debug("Error getting HEAD ref from local repo")

		config.Stats.TrackSingle(stats.GetHeadRefFailed, repo)

		return nil, errors.WithStackTrace(headErr)
	}
	return ref, nil
}

// executeCommand runs the user-supplied command against the given repository
func executeCommand(config *config.GitXargsConfig, repositoryDir string, repo *github.Repository) error {
	return executeCommandWithLogger(config, repositoryDir, repo, logging.GetLogger("git-xargs"))
}

// executeCommandWithLogger runs the user-supplied command against the given repository, and sends the log output
// to the given logger
func executeCommandWithLogger(config *config.GitXargsConfig, repositoryDir string, repo *github.Repository, logger *logrus.Logger) error {
	if len(config.Args) < 1 {
		return errors.WithStackTrace(types.NoCommandSuppliedErr{})
	}

	cmdArgs := config.Args

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = repositoryDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("XARGS_DRY_RUN=%t", config.DryRun))
	cmd.Env = append(cmd.Env, fmt.Sprintf("XARGS_REPO_NAME=%s", repo.GetName()))
	cmd.Env = append(cmd.Env, fmt.Sprintf("XARGS_REPO_OWNER=%s", repo.GetOwner().GetLogin()))

	logger.WithFields(logrus.Fields{
		"Repo":      repo.GetName(),
		"Directory": repositoryDir,
		"Command":   config.Args,
	}).Debug("Executing command against local clone of repo...")

	stdoutStdErr, err := cmd.CombinedOutput()

	logger.Debugf("Output of command %v for repo %s in directory %s:\n%s", config.Args, repo.GetName(), repositoryDir, string(stdoutStdErr))

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error getting output of command execution")
		// Track the command error against the repo
		config.Stats.TrackSingle(stats.CommandErrorOccurredDuringExecution, repo)
		return errors.WithStackTrace(err)
	}

	return nil
}

// getLocalWorkTree looks up the working tree of the locally cloned repository and returns it if possible, or an error
func getLocalWorkTree(repositoryDir string, localRepository *git.Repository, repo *github.Repository) (*git.Worktree, error) {
	logger := logging.GetLogger("git-xargs")

	worktree, worktreeErr := localRepository.Worktree()

	if worktreeErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": worktreeErr,
			"Repo":  repo.GetName(),
			"Dir":   repositoryDir,
		}).Debug("Error looking up local repository's worktree")

		return nil, errors.WithStackTrace(worktreeErr)
	}
	return worktree, nil
}

// checkoutLocalBranch creates a local branch specific to this tool in the locally checked out copy of the repo in the /tmp folder
func checkoutLocalBranch(config *config.GitXargsConfig, ref *plumbing.Reference, worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository) (plumbing.ReferenceName, error) {
	logger := logging.GetLogger("git-xargs")

	// BranchName is a global variable that is set in cmd/root.go. It is override-able by the operator via the --branch-name or -b flag. It defaults to "git-xargs"

	branchName := plumbing.NewBranchReferenceName(config.BranchName)
	logger.WithFields(logrus.Fields{
		"Branch Name": branchName,
		"Repo":        remoteRepository.GetName(),
	}).Debug("Created branch")

	// Create a branch specific to the multi repo script runner
	co := &git.CheckoutOptions{
		Hash:   ref.Hash(),
		Branch: branchName,
		Create: true,
	}

	// Attempt to checkout the new tool-specific branch on which the supplied command will be executed
	checkoutErr := worktree.Checkout(co)

	if checkoutErr != nil {
		if config.SkipPullRequests &&
			remoteRepository.GetDefaultBranch() == config.BranchName &&
			strings.Contains(checkoutErr.Error(), "already exists") {
			// User has requested pull requess be skipped, meaning they want their commits pushed on their target branch
			// If the target branch is also the repo's default branch and therefore already exists, we don't have an error
		} else {
			logger.WithFields(logrus.Fields{
				"Error": checkoutErr,
				"Repo":  remoteRepository.GetName(),
			}).Debug("Error creating new branch")

			// Track the error checking out the branch
			config.Stats.TrackSingle(stats.BranchCheckoutFailed, remoteRepository)

			return branchName, errors.WithStackTrace(checkoutErr)
		}
	}

	// Pull latest code from remote branch if it exists to avoid fast-forwarding errors
	gitProgressBuffer := bytes.NewBuffer(nil)
	po := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: branchName,
		Auth: &http.BasicAuth{
			Username: remoteRepository.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
		Progress: gitProgressBuffer,
	}

	logger.WithFields(logrus.Fields{
		"Repo": remoteRepository.GetName(),
	}).Debug(gitProgressBuffer)

	pullErr := worktree.Pull(po)

	if pullErr != nil {

		if pullErr == plumbing.ErrReferenceNotFound {
			// The supplied branch just doesn't exist yet on the remote - this is not a fatal error and will
			// allow the new branch to be pushed in pushLocalBranch
			config.Stats.TrackSingle(stats.BranchRemoteDidntExistYet, remoteRepository)
			return branchName, nil
		}

		if pullErr == git.NoErrAlreadyUpToDate {
			// The local branch is already up to date, which is not a fatal error
			return branchName, nil
		}

		// Track the error pulling the latest from the remote branch
		config.Stats.TrackSingle(stats.BranchRemotePullFailed, remoteRepository)

		return branchName, errors.WithStackTrace(pullErr)
	}

	return branchName, nil
}

// updateRepo will check for any changes in worktree as a result of script execution, and if any are present,
// add any untracked, deleted or modified files, create a commit using the supplied or default commit message,
// push the code to the remote repo, and open a pull request.
func updateRepo(config *config.GitXargsConfig,
	repositoryDir string,
	worktree *git.Worktree,
	remoteRepository *github.Repository,
	localRepository *git.Repository,
	branchName string,
) error {
	logger := logging.GetLogger("git-xargs")

	status, statusErr := worktree.Status()

	if statusErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": statusErr,
			"Repo":  remoteRepository.GetName(),
			"Dir":   repositoryDir,
		}).Debug("Error looking up worktree status")

		// Track the status check failure
		config.Stats.TrackSingle(stats.WorktreeStatusCheckFailedCommand, remoteRepository)
		return errors.WithStackTrace(statusErr)
	}

	// If there are no changes, we log it, track it, and return
	if status.IsClean() {
		logger.WithFields(logrus.Fields{
			"Repo": remoteRepository.GetName(),
		}).Debug("Local repository status is clean - nothing to stage or commit")

		// Track the fact that repo had no file changes post command execution
		config.Stats.TrackSingle(stats.WorktreeStatusClean, remoteRepository)
		return nil
	}

	// Commit any untracked files, modified or deleted files that resulted from script execution
	commitErr := commitLocalChanges(status, config, repositoryDir, worktree, remoteRepository, localRepository)
	if commitErr != nil {
		return commitErr
	}

	// Push the local branch containing all of our changes from executing the supplied command
	pushBranchErr := pushLocalBranch(config, remoteRepository, localRepository)
	if pushBranchErr != nil {
		return pushBranchErr
	}

	// Create an OpenPrRequest that tracks retries
	opr := types.OpenPrRequest{
		Repo:    remoteRepository,
		Branch:  branchName,
		Retries: 0,
	}

	return openPullRequestsWithThrottling(config, opr)
}

// commitLocalChanges will check for any changes in worktree as a result of script execution, and if any are present,
// add any untracked, deleted or modified files and create a commit using the supplied or default commit message.
func commitLocalChanges(status git.Status, config *config.GitXargsConfig, repositoryDir string, worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository) error {
	logger := logging.GetLogger("git-xargs")

	// If there are changes, we need to stage, add and commit them
	logger.WithFields(logrus.Fields{
		"Repo": remoteRepository.GetName(),
	}).Debug("Local repository worktree no longer clean, will stage and add new files and commit changes")

	// Track the fact that worktree changes were made following execution
	config.Stats.TrackSingle(stats.WorktreeStatusDirty, remoteRepository)

	for filepath := range status {
		if status.IsUntracked(filepath) {
			logger.WithFields(logrus.Fields{
				"Filepath": filepath,
			}).Debug("Found untracked file. Adding to stage")

			_, addErr := worktree.Add(filepath)
			if addErr != nil {
				logger.WithFields(logrus.Fields{
					"Error":    addErr,
					"Filepath": filepath,
				}).Debug("Error adding file to git stage")
				// Track the file staging failure
				config.Stats.TrackSingle(stats.WorktreeAddFileFailed, remoteRepository)
				return errors.WithStackTrace(addErr)
			}
		}
	}

	// With all our untracked files staged, we can now create a commit, passing the All
	// option when configuring our commit option so that all modified and deleted files
	// will have their changes committed
	commitOps := &git.CommitOptions{
		All: true,
	}

	_, commitErr := worktree.Commit(config.CommitMessage, commitOps)

	if commitErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": commitErr,
			"Repo":  remoteRepository.GetName(),
		})

		// If we reach this point, we were unable to commit our changes, so we'll
		// continue rather than attempt to push an empty branch and open an empty PR
		config.Stats.TrackSingle(stats.CommitChangesFailed, remoteRepository)
		return errors.WithStackTrace(commitErr)
	}

	// If --skip-pull-requests was passed, track the repos whose changes were committed directly to the main branch
	if config.SkipPullRequests {
		config.Stats.TrackSingle(stats.CommitsMadeDirectlyToBranch, remoteRepository)
	}

	return nil
}

// pushLocalBranch pushes the branch in the local clone of the /tmp/ directory repository to the GitHub remote origin
// so that a pull request can be opened against it via the GitHub API
func pushLocalBranch(config *config.GitXargsConfig, remoteRepository *github.Repository, localRepository *git.Repository) error {
	logger := logging.GetLogger("git-xargs")

	if config.DryRun {
		logger.WithFields(logrus.Fields{
			"Repo": remoteRepository.GetName(),
		}).Debug("Skipping branch push to remote origin because --dry-run flag is set")

		config.Stats.TrackSingle(stats.PushBranchSkipped, remoteRepository)
		return nil
	}
	// Push the changes to the remote repo
	po := &git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: remoteRepository.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	}
	pushErr := localRepository.Push(po)

	if pushErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": pushErr,
			"Repo":  remoteRepository.GetName(),
		}).Debug("Error pushing new branch to remote origin")

		// Track the push failure
		config.Stats.TrackSingle(stats.PushBranchFailed, remoteRepository)
		return errors.WithStackTrace(pushErr)
	}

	logger.WithFields(logrus.Fields{
		"Repo": remoteRepository.GetName(),
	}).Debug("Successfully pushed local branch to remote origin")

	// If --skip-pull-requests was passed, track the fact that these changes were pushed directly to the main branch
	if config.SkipPullRequests {
		config.Stats.TrackSingle(stats.DirectCommitsPushedToRemoteBranch, remoteRepository)
	}

	return nil
}

// Attempt to open a pull request via the GitHub API, of the supplied branch specific to this tool, against the main
// branch for the remote origin
func openPullRequest(config *config.GitXargsConfig, pr types.OpenPrRequest) error {
	logger := logging.GetLogger("git-xargs")

	// If the current request has already exhausted the configured number of PR retries, short-circuit
	if pr.Retries > config.PullRequestRetries {
		config.Stats.TrackSingle(stats.PRFailedAfterMaximumRetriesErr, pr.Repo)
		return nil
	}

	if config.DryRun || config.SkipPullRequests {
		logger.WithFields(logrus.Fields{
			"Repo": pr.Repo.GetName(),
		}).Debug("--dry-run and / or --skip-pull-requests is set to true, so skipping opening a pull request!")
		return nil
	}

	logger.Debugf("openPullRequest received job with retries: %d. Config max retries for this run: %d", pr.Retries, config.PullRequestRetries)

	repoDefaultBranch := config.BaseBranchName
	if repoDefaultBranch == "" {
		repoDefaultBranch = pr.Repo.GetDefaultBranch()
	}

	pullRequestAlreadyExists, err := pullRequestAlreadyExistsForBranch(config, pr.Repo, pr.Branch, repoDefaultBranch)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
			"Head":  pr.Branch,
			"Base":  repoDefaultBranch,
		}).Debug("Error listing pull requests")

		// Track pull request open failure
		config.Stats.TrackSingle(stats.PullRequestOpenErr, pr.Repo)
		return errors.WithStackTrace(err)
	}

	if pullRequestAlreadyExists {
		logger.WithFields(logrus.Fields{
			"Repo": pr.Repo.GetName(),
			"Head": pr.Branch,
			"Base": repoDefaultBranch,
		}).Debug("Pull request already exists for this branch, so skipping opening a pull request!")

		// Track that we skipped opening a pull request
		config.Stats.TrackSingle(stats.PullRequestAlreadyExists, pr.Repo)
		return nil
	}

	// If the user only supplies a commit message, use that for both the pull request title and descriptions,
	// unless they are provided separately
	titleToUse := config.PullRequestTitle
	descriptionToUse := config.PullRequestDescription

	commitMessage := config.CommitMessage

	if commitMessage != common.DefaultCommitMessage {
		if titleToUse == common.DefaultPullRequestTitle {
			titleToUse = commitMessage
		}

		if descriptionToUse == common.DefaultPullRequestDescription {
			descriptionToUse = commitMessage
		}
	}

	// Configure pull request options that the GitHub client accepts when making calls to open new pull requests
	newPR := &github.NewPullRequest{
		Title:               github.String(titleToUse),
		Head:                github.String(pr.Branch),
		Base:                github.String(repoDefaultBranch),
		Body:                github.String(descriptionToUse),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(config.Draft),
	}

	// Make a pull request via the Github API
	githubPR, resp, err := config.GithubClient.PullRequests.Create(context.Background(), *pr.Repo.GetOwner().Login, pr.Repo.GetName(), newPR)

	// The go-github library's CheckResponse method can return two different types of rate limiting error:
	// 1. AbuseRateLimitError which may contain a Retry-After header whose value we can use to slow down, or
	// 2. RateLimitError which may contain information about when the rate limit will be removed, that we can also use to slow down
	// Therefore, we need to use type assertions to test for each type of error response, and accordingly fetch the data it may contain
	// about how long git-xargs should wait before its next attempt to open a pull request
	githubErr := github.CheckResponse(resp.Response)

	if githubErr != nil {

		isRateLimited := false

		// Create a new open pull request struct that we'll eventually send on the PRChan
		opr := types.OpenPrRequest{
			Repo:    pr.Repo,
			Branch:  pr.Branch,
			Retries: 1,
		}

		// If this request has been seen before, increment its retries count, taking into account previous iterations
		opr.Retries = (pr.Retries + 1)

		// If GitHub returned an error of type RateLimitError, we can attempt to compute the next time to try the request again
		// by reading its rate limit information
		if rateLimitError, ok := githubErr.(*github.RateLimitError); ok {
			isRateLimited = true
			retryAfter := time.Until(rateLimitError.Rate.Reset.Time)
			opr.Delay = retryAfter
			logger.Debugf("git-xargs parsed retryAfter %d from GitHub rate limit error's reset time", retryAfter)
		}

		// If GitHub returned a Retry-After header, use its value, otherwise use the default
		if abuseRateLimitError, ok := githubErr.(*github.AbuseRateLimitError); ok {
			isRateLimited = true
			if abuseRateLimitError.RetryAfter != nil {
				if abuseRateLimitError.RetryAfter.Seconds() > 0 {
					opr.Delay = *abuseRateLimitError.RetryAfter
				}
			}
		}

		if isRateLimited {
			// If we couldn't determine a more accurate delay from GitHub API response headers, then fall back to our user-configurable default
			if opr.Delay == 0 {
				opr.Delay = time.Duration(config.SecondsToSleepWhenRateLimited)
			}

			logger.Debugf("Retrying PR for repo: %s again later with %d second delay due to secondary rate limiting.", pr.Repo.GetName(), opr.Delay)
			// Put another pull request on the channel so this can effectively be retried after a cooldown

			// Keep track of the repo's PR initially failing due to rate limiting
			config.Stats.TrackSingle(stats.PRFailedDueToRateLimitsErr, pr.Repo)
			return openPullRequestsWithThrottling(config, opr)
		}
	}

	// Otherwise, if we reach this point, we can assume we are not rate limited, and hence must do some
	// further inspection on the error values returned to us to determine what went wrong
	prErrorMessage := "Error opening pull request"

	// Github's API will return HTTP status code 422 for several different errors
	// Currently, there are two such errors that git-xargs is concerned with:
	// 1. User passes the --draft flag, but the targeted repo does not support draft pull requests
	// 2. User passes the --base-branch-name flag, specifying a branch that does not exist in the repo
	if err != nil {
		if resp.StatusCode == 422 {
			switch {
			case strings.Contains(err.Error(), "Draft pull requests are not supported"):
				prErrorMessage = "Error opening pull request: draft PRs not supported for this repo. See https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests"
				config.Stats.TrackSingle(stats.RepoDoesntSupportDraftPullRequestsErr, pr.Repo)

			case strings.Contains(err.Error(), "Field:base Code:invalid"):
				prErrorMessage = fmt.Sprintf("Error opening pull request: Base branch name: %s is invalid", config.BaseBranchName)
				config.Stats.TrackSingle(stats.BaseBranchTargetInvalidErr, pr.Repo)

			default:
				config.Stats.TrackSingle(stats.PullRequestOpenErr, pr.Repo)
			}
		}

		// If the Github reponse's status code is not 422, fallback to logging and tracking a generic pull request error
		config.Stats.TrackSingle(stats.PullRequestOpenErr, pr.Repo)

		logger.WithFields(logrus.Fields{
			"Error": err,
			"Head":  pr.Branch,
			"Base":  repoDefaultBranch,
			"Body":  descriptionToUse,
		}).Debug(prErrorMessage)
		return errors.WithStackTrace(err)
	}

	// There was no error opening the pull request
	logger.WithFields(logrus.Fields{
		"Pull Request URL": githubPR.GetHTMLURL(),
	}).Debug("Successfully opened pull request")

	reviewersRequest := github.ReviewersRequest{
		NodeID:        githubPR.NodeID,
		Reviewers:     config.Reviewers,
		TeamReviewers: config.TeamReviewers,
	}

	// If the user supplied reviewer information on the pull request, initiate a separate request to ask for reviews
	if config.HasReviewers() {
		_, _, reviewRequestErr := config.GithubClient.PullRequests.RequestReviewers(context.Background(), *pr.Repo.GetOwner().Login, pr.Repo.GetName(), githubPR.GetNumber(), reviewersRequest)
		if reviewRequestErr != nil {
			config.Stats.TrackSingle(stats.RequestReviewersErr, pr.Repo)
		}

	}

	if config.HasAssignees() {
		assignees := config.Assignees
		_, _, assigneeErr := config.GithubClient.Issues.AddAssignees(context.Background(), *pr.Repo.GetOwner().Login, pr.Repo.GetName(), githubPR.GetNumber(), assignees)
		if assigneeErr != nil {
			config.Stats.TrackSingle(stats.AddAssigneesErr, pr.Repo)
		}
	}

	if config.Draft {
		config.Stats.TrackDraftPullRequest(pr.Repo.GetName(), githubPR.GetHTMLURL())
	} else {
		// Track successful opening of the pull request, extracting the HTML url to the PR itself for easier review
		config.Stats.TrackPullRequest(pr.Repo.GetName(), githubPR.GetHTMLURL())
	}
	return nil
}

// Returns true if a pull request already exists in the given repo for the given branch
func pullRequestAlreadyExistsForBranch(config *config.GitXargsConfig, repo *github.Repository, branch string, repoDefaultBranch string) (bool, error) {
	opts := &github.PullRequestListOptions{
		// Filter pulls by head user or head organization and branch name in the format of user:ref-name or organization:ref-name
		// https://docs.github.com/en/rest/reference/pulls#list-pull-requests
		Head: fmt.Sprintf("%s:%s", *repo.GetOwner().Login, branch),
		Base: repoDefaultBranch,
	}

	prs, _, err := config.GithubClient.PullRequests.List(context.Background(), *repo.GetOwner().Login, repo.GetName(), opts)
	if err != nil {
		return false, errors.WithStackTrace(err)
	}

	return len(prs) > 0, nil
}
