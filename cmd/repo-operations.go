package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"

	"github.com/google/go-github/v32/github"

	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
)

// cloneLocalRepository clones a remote Github repo via SSH to a local temporary directory so that the supplied command
// can be run against the repo locally and any git changes handled thereafter. The local directory has
// git-xargs-<repo-name> appended to it to make it easier to find when you are looking for it while debugging
func cloneLocalRepository(config *GitXargsConfig, repo *github.Repository) (string, *git.Repository, error) {
	logger := logging.GetLogger("git-xargs")

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

	localRepository, err := config.GitClient.PlainClone(repositoryDir, false, &git.CloneOptions{
		URL:      repo.GetCloneURL(),
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: repo.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	})

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
			"Repo":  repo.GetName(),
		}).Debug("Error cloning repository")

		// Track failure to clone for our final run report
		config.Stats.TrackSingle(RepoFailedToClone, repo)

		return repositoryDir, nil, errors.WithStackTrace(err)
	}

	config.Stats.TrackSingle(RepoSuccessfullyCloned, repo)

	return repositoryDir, localRepository, nil
}

// getLocalRepoHeadRef looks up the HEAD reference of the locally cloned git repository, which is required by
// downstream operations such as branching
func getLocalRepoHeadRef(config *GitXargsConfig, localRepository *git.Repository, repo *github.Repository) (*plumbing.Reference, error) {
	logger := logging.GetLogger("git-xargs")

	ref, headErr := localRepository.Head()
	if headErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": headErr,
			"Repo":  repo.GetName(),
		}).Debug("Error getting HEAD ref from local repo")

		config.Stats.TrackSingle(GetHeadRefFailed, repo)

		return nil, errors.WithStackTrace(headErr)
	}
	return ref, nil
}

// executeCommand runs the user-supplied command and runs it against the given repository, adding any git changes that may occur
func executeCommand(config *GitXargsConfig, repositoryDir string, repo *github.Repository, worktree *git.Worktree) error {

	logger := logging.GetLogger("git-xargs")

	if len(config.Args) < 1 {
		return errors.WithStackTrace(NoCommandSuppliedErr{})
	}

	cmdArgs := config.Args

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = repositoryDir

	logger.WithFields(logrus.Fields{
		"Repo":      repo.GetName(),
		"Directory": repositoryDir,
		"Command":   config.Args,
	}).Debug("Executing command against local clone of repo...")

	stdoutStdErr, err := cmd.CombinedOutput()

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error getting output of command execution")
		// Track the command error against the repo
		config.Stats.TrackSingle(CommandErrorOccurredDuringExecution, repo)
		return errors.WithStackTrace(err)
	}

	logger.WithFields(logrus.Fields{
		"CombinedOutput": string(stdoutStdErr),
	}).Debug("Received output of command run")

	status, statusErr := worktree.Status()

	if statusErr != nil {
		logger.WithFields(logrus.Fields{
			"Error": statusErr,
			"Repo":  repo.GetName(),
			"Dir":   repositoryDir,
		}).Debug("Error looking up worktree status")

		// Track the status check failure
		config.Stats.TrackSingle(WorktreeStatusCheckFailedCommand, repo)
		return errors.WithStackTrace(statusErr)
	}

	// If the supplied command resulted in any changes, we need to stage, add and commit them
	if !status.IsClean() {
		logger.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("Local repository worktree no longer clean, will stage and add new files and commit changes")

		// Track the fact that worktree changes were made following execution
		config.Stats.TrackSingle(WorktreeStatusDirty, repo)

		for filepath := range status {
			if status.IsUntracked(filepath) {
				fmt.Printf("Found untracked file %s. Adding to stage", filepath)
				_, addErr := worktree.Add(filepath)
				if addErr != nil {
					logger.WithFields(logrus.Fields{
						"Error":    addErr,
						"Filepath": filepath,
					}).Debug("Error adding file to git stage")
					// Track the file staging failure
					config.Stats.TrackSingle(WorktreeAddFileFailed, repo)
					return errors.WithStackTrace(addErr)
				}
			}
		}

	} else {
		logger.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("Local repository status is clean - nothing to stage or commit")

		// Track the fact that repo had no file changes post command execution
		config.Stats.TrackSingle(WorktreeStatusClean, repo)
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
func checkoutLocalBranch(config *GitXargsConfig, ref *plumbing.Reference, worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository) (plumbing.ReferenceName, error) {
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
		logger.WithFields(logrus.Fields{
			"Error": checkoutErr,
			"Repo":  remoteRepository.GetName(),
		}).Debug("Error creating new branch")

		// Track the error checking out the branch
		config.Stats.TrackSingle(BranchCheckoutFailed, remoteRepository)

		return branchName, errors.WithStackTrace(checkoutErr)
	}

	// Pull latest code from remote branch if it exists to avoid fast-forwarding errors
	po := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: branchName,
		Auth: &http.BasicAuth{
			Username: remoteRepository.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
		Progress: os.Stdout,
	}

	pullErr := worktree.Pull(po)

	if pullErr != nil {

		if pullErr == plumbing.ErrReferenceNotFound {
			// The suppled branch just doesn't exist yet on the remote - this is not a fatal error and will
			// allow the new branch to be pushed in pushLocalBranch
			config.Stats.TrackSingle(BranchRemoteDidntExistYet, remoteRepository)
			return branchName, nil
		}

		// Track the error pulling the latest from the remote branch
		config.Stats.TrackSingle(BranchRemotePullFailed, remoteRepository)

		return branchName, errors.WithStackTrace(pullErr)
	}

	return branchName, nil
}

// commitLocalChanges will create a commit using the supplied or default commit message and will add any untracked, deleted
// or modified files that resulted from script execution
func commitLocalChanges(config *GitXargsConfig, worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository) error {

	logger := logging.GetLogger("git-xargs")

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
		config.Stats.TrackSingle(CommitChangesFailed, remoteRepository)
		return errors.WithStackTrace(commitErr)
	}

	// If --skip-pull-requests was passed, track the repos whose changes were committed directly to the main branch
	if config.SkipPullRequests {
		config.Stats.TrackSingle(CommitsMadeDirectlyToBranch, remoteRepository)
	}

	return nil
}

// pushLocalBranch pushes the branch in the local clone of the /tmp/ directory repository to the Github remote origin
// so that a pull request can be opened against it via the Github API
func pushLocalBranch(config *GitXargsConfig, remoteRepository *github.Repository, localRepository *git.Repository) error {
	logger := logging.GetLogger("git-xargs")

	if config.DryRun {
		logger.WithFields(logrus.Fields{
			"Repo": remoteRepository.GetName(),
		}).Debug("Skipping branch push to remote origin because --dry-run flag is set")

		config.Stats.TrackSingle(PushBranchSkipped, remoteRepository)
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
		config.Stats.TrackSingle(PushBranchFailed, remoteRepository)
		return errors.WithStackTrace(pushErr)
	}

	logger.WithFields(logrus.Fields{
		"Repo": remoteRepository.GetName(),
	}).Debug("Successfully pushed local branch to remote origin")

	// If --skip-pull-requests was passed, track the fact that these changes were pushed directly to the main branch
	if config.SkipPullRequests {
		config.Stats.TrackSingle(DirectCommitsPushedToRemoteBranch, remoteRepository)
	}

	return nil
}

// Attempt to open a pull request via the Github API, of the supplied branch specific to this tool, against the main
// branch for the remote origin
func openPullRequest(config *GitXargsConfig, repo *github.Repository, branch string) error {

	logger := logging.GetLogger("git-xargs")

	if config.DryRun || config.SkipPullRequests {
		logger.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("--dry-run and / or --skip-pull-requests is set to true, so skipping opening a pull request!")
		return nil
	}

	// If the user only supplies a commit message, use that for both the pull request title and descriptions,
	// unless they are provided separately
	titleToUse := config.PullRequestTitle
	descriptionToUse := config.PullRequestDescription

	commitMessage := config.CommitMessage

	if commitMessage != DefaultCommitMessage {
		if titleToUse == DefaultPullRequestTitle {
			titleToUse = commitMessage
		}

		if descriptionToUse == DefaultPullRequestDescription {
			descriptionToUse = commitMessage
		}
	}

	// Configure pull request options that the Github client accepts when making calls to open new pull requests
	newPR := &github.NewPullRequest{
		Title:               github.String(titleToUse),
		Head:                github.String(branch),
		Base:                github.String("master"),
		Body:                github.String(descriptionToUse),
		MaintainerCanModify: github.Bool(true),
	}

	// Make a pull request via the Github API
	pr, _, err := config.GithubClient.PullRequests.Create(context.Background(), *repo.GetOwner().Login, repo.GetName(), newPR)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"Error": err,
			"Head":  branch,
			"Base":  "master",
			"Body":  descriptionToUse,
		}).Debug("Error opening Pull request")

		// Track pull request open failure
		config.Stats.TrackSingle(PullRequestOpenErr, repo)
		return errors.WithStackTrace(err)
	}

	logger.WithFields(logrus.Fields{
		"Pull Request URL": pr.GetHTMLURL(),
	}).Debug("Successfully opened pull request")

	// Track successful opening of the pull request, extracting the HTML url to the PR itself for easier review
	config.Stats.TrackPullRequest(repo.GetName(), pr.GetHTMLURL())
	return nil
}
