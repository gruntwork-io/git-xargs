package repository

import (
	"context"
	"fmt"

	"github.com/gruntwork-io/git-xargs/auth"
	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/stats"
	"github.com/gruntwork-io/git-xargs/types"
	"github.com/gruntwork-io/go-commons/errors"

	"github.com/google/go-github/v43/github"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

// getFileDefinedRepos converts user-supplied repositories to GitHub API response objects that can be further processed
func getFileDefinedRepos(GithubClient auth.GithubClient, allowedRepos []*types.AllowedRepo, tracker *stats.RunStats) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	var allRepos []*github.Repository

	for _, allowedRepo := range allowedRepos {

		logger.WithFields(logrus.Fields{
			"Organization": allowedRepo.Organization,
			"Name":         allowedRepo.Name,
		}).Debug("Looking up filename provided repo")

		repo, resp, err := GithubClient.Repositories.Get(context.Background(), allowedRepo.Organization, allowedRepo.Name)

		if err != nil {
			logger.WithFields(logrus.Fields{
				"Error":                err,
				"Response Status Code": resp.StatusCode,
				"AllowedRepoOwner":     allowedRepo.Organization,
				"AllowedRepoName":      allowedRepo.Name,
			}).Debug("error getting single repo")

			if resp.StatusCode == 404 {
				// This repo does not exist / could not be fetched as named, so we won't include it in the list of repos to process

				// create an empty GitHub repo object to satisfy the stats tracking interface
				missingRepo := &github.Repository{
					Owner: &github.User{Login: github.String(allowedRepo.Organization)},
					Name:  github.String(allowedRepo.Name),
				}
				tracker.TrackSingle(stats.RepoNotExists, missingRepo)
				continue
			} else {
				return allRepos, errors.WithStackTrace(err)
			}
		}

		if resp.StatusCode == 200 {
			logger.WithFields(logrus.Fields{
				"Organization": allowedRepo.Organization,
				"Name":         allowedRepo.Name,
			}).Debug("Successfully fetched repo")

			allRepos = append(allRepos, repo)
		}
	}
	return allRepos, nil
}

// getReposByOrg takes the string name of a GitHub organization and pages through the API to fetch all of its repositories
func getReposByOrg(config *config.GitXargsConfig) ([]*github.Repository, error) {

	logger := logging.GetLogger("git-xargs")

	// Page through all of the organization's repos, collecting them in this slice
	var allRepos []*github.Repository

	if config.GithubOrg == "" {
		return allRepos, errors.WithStackTrace(types.NoGithubOrgSuppliedErr{})
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		var reposToAdd []*github.Repository
		repos, resp, err := config.GithubClient.Repositories.ListByOrg(context.Background(), config.GithubOrg, opt)
		if err != nil {
			return allRepos, errors.WithStackTrace(err)
		}

		// github.RepositoryListByOrgOptions doesn't seem to be able to filter out archived repos
		// So re-slice the repos list if --skip-archived-repos is passed and the repository is in archived/read-only state
		for i, repo := range repos {
			if config.SkipArchivedRepos && repo.GetArchived() {
				logger.WithFields(logrus.Fields{
					"Name": repo.GetFullName(),
				}).Debug("Skipping archived repository")

				// Track repos to skip because of archived status for our final run report
				config.Stats.TrackSingle(stats.ReposArchivedSkipped, repo)

				reposToAdd = append(repos[:i], repos[i+1:]...)
			} else {
				reposToAdd = repos
			}
		}

		allRepos = append(allRepos, reposToAdd...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	repoCount := len(allRepos)

	if repoCount == 0 {
		return nil, errors.WithStackTrace(types.NoReposFoundErr{GithubOrg: config.GithubOrg})
	}

	logger.WithFields(logrus.Fields{
		"Repo count": repoCount,
	}).Debug(fmt.Sprintf("Fetched repos from Github organization: %s", config.GithubOrg))

	config.Stats.TrackMultiple(stats.FetchedViaGithubAPI, allRepos)

	return allRepos, nil
}
