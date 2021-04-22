package main

import (
	"context"
	"fmt"

	"github.com/gruntwork-io/go-commons/errors"

	"github.com/google/go-github/v32/github"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

// getFileDefinedRepos converts user-supplied repositories to Github API response objects that can be further processed
func getFileDefinedRepos(GithubClient GithubClient, allowedRepos []*AllowedRepo, stats *RunStats) ([]*github.Repository, error) {
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

				// create an empty github repo object to satisfy the stats tracking interface
				missingRepo := &github.Repository{
					Owner: &github.User{Login: github.String(allowedRepo.Organization)},
					Name:  github.String(allowedRepo.Name),
				}
				stats.TrackSingle(RepoNotExists, missingRepo)
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

// getReposByOrg takes the string name of a Github organization and pages through the API to fetch all of its repositories
func getReposByOrg(config *GitXargsConfig) ([]*github.Repository, error) {

	logger := logging.GetLogger("git-xargs")

	// Page through all of the organization's repos, collecting them in this slice
	var allRepos []*github.Repository

	if config.GithubOrg == "" {
		return allRepos, errors.WithStackTrace(NoGithubOrgSuppliedErr{})
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := config.GithubClient.Repositories.ListByOrg(context.Background(), config.GithubOrg, opt)
		if err != nil {
			return allRepos, errors.WithStackTrace(err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	repoCount := len(allRepos)

	if repoCount == 0 {
		return nil, errors.WithStackTrace(NoReposFoundErr{GithubOrg: config.GithubOrg})
	}

	logger.WithFields(logrus.Fields{
		"Repo count": repoCount,
	}).Debug(fmt.Sprintf("Fetched repos from Github organization: %s", config.GithubOrg))

	config.Stats.TrackMultiple(FetchedViaGithubAPI, allRepos)

	return allRepos, nil
}
