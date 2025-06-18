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
		// So filter the repos list if --skip-archived-repos is passed and the repository is in archived/read-only state
		if config.SkipArchivedRepos {
			for _, repo := range repos {
				if repo.GetArchived() {
					logger.WithFields(logrus.Fields{
						"Name": repo.GetFullName(),
					}).Debug("Skipping archived repository")

					// Track repos to skip because of archived status for our final run report
					config.Stats.TrackSingle(stats.ReposArchivedSkipped, repo)
				} else {
					reposToAdd = append(reposToAdd, repo)
				}
			}
		} else {
			reposToAdd = repos
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

// getReposBySearch uses GitHub's search API to find repositories matching the given query
func getReposBySearch(config *config.GitXargsConfig) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	// Handle different search scenarios
	if config.GithubRepositorySearch != "" && config.GithubCodeSearch != "" {
		// Both searches provided - return intersection
		logger.Debug("Both repository and code search queries provided, finding intersection")
		return getReposByIntersection(config)
	} else if config.GithubRepositorySearch != "" {
		// Only repository search
		return getReposByRepositorySearch(config)
	} else if config.GithubCodeSearch != "" {
		// Only code search
		return getReposByCodeSearch(config)
	}

	return nil, errors.WithStackTrace(types.NoGithubSearchQuerySuppliedErr{})
}

// getReposByIntersection finds repositories that match both repository and code search queries
func getReposByIntersection(config *config.GitXargsConfig) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	// Get repositories from repository search
	repoSearchRepos, err := getReposByRepositorySearch(config)
	if err != nil {
		return nil, err
	}

	// Get repositories from code search
	codeSearchRepos, err := getReposByCodeSearch(config)
	if err != nil {
		return nil, err
	}

	// Find intersection
	repoMap := make(map[string]*github.Repository)
	for _, repo := range repoSearchRepos {
		repoMap[repo.GetFullName()] = repo
	}

	var intersectionRepos []*github.Repository
	for _, repo := range codeSearchRepos {
		if _, found := repoMap[repo.GetFullName()]; found {
			intersectionRepos = append(intersectionRepos, repo)
		}
	}

	repoCount := len(intersectionRepos)
	if repoCount == 0 {
		return nil, errors.WithStackTrace(types.NoReposFoundFromSearchErr{
			Query: fmt.Sprintf("intersection of repository search '%s' and code search '%s'",
				config.GithubRepositorySearch, config.GithubCodeSearch),
		})
	}

	logger.WithFields(logrus.Fields{
		"Repo count":       repoCount,
		"Repository Query": config.GithubRepositorySearch,
		"Code Query":       config.GithubCodeSearch,
	}).Debug("Found intersection of repository and code search results")

	config.Stats.TrackMultiple(stats.FetchedViaGithubAPI, intersectionRepos)

	return intersectionRepos, nil
}

// getReposByRepositorySearch uses GitHub's repository search API to find repositories matching the given query
func getReposByRepositorySearch(config *config.GitXargsConfig) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	var allRepos []*github.Repository

	if config.GithubRepositorySearch == "" {
		return nil, errors.WithStackTrace(types.NoGithubSearchQuerySuppliedErr{})
	}

	// Build the search query
	searchQuery := config.GithubRepositorySearch

	// If a specific organization is provided, add it to the query
	if config.GithubOrg != "" {
		searchQuery = fmt.Sprintf("%s org:%s", searchQuery, config.GithubOrg)
	}

	logger.WithFields(logrus.Fields{
		"Query": searchQuery,
	}).Debug("Searching for repositories using GitHub Repository Search API")

	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		var reposToAdd []*github.Repository
		result, resp, err := config.GithubClient.Search.Repositories(context.Background(), searchQuery, opt)
		if err != nil {
			return allRepos, errors.WithStackTrace(err)
		}

		repos := result.Repositories

		// Filter out archived repos if --skip-archived-repos is passed
		if config.SkipArchivedRepos {
			for _, repo := range repos {
				if repo.GetArchived() {
					logger.WithFields(logrus.Fields{
						"Name": repo.GetFullName(),
					}).Debug("Skipping archived repository from search results")

					// Track repos to skip because of archived status for our final run report
					config.Stats.TrackSingle(stats.ReposArchivedSkipped, repo)
				} else {
					reposToAdd = append(reposToAdd, repo)
				}
			}
		} else {
			reposToAdd = repos
		}

		allRepos = append(allRepos, reposToAdd...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	repoCount := len(allRepos)

	if repoCount == 0 {
		return nil, errors.WithStackTrace(types.NoReposFoundFromSearchErr{Query: searchQuery})
	}

	logger.WithFields(logrus.Fields{
		"Repo count": repoCount,
		"Query":      searchQuery,
	}).Debug("Fetched repos from GitHub Repository Search API")

	config.Stats.TrackMultiple(stats.FetchedViaGithubAPI, allRepos)

	return allRepos, nil
}

// getReposByCodeSearch uses GitHub's code search API to find repositories containing matching code
func getReposByCodeSearch(config *config.GitXargsConfig) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	var allRepos []*github.Repository
	repoMap := make(map[string]*github.Repository) // To avoid duplicates

	if config.GithubCodeSearch == "" {
		return allRepos, errors.WithStackTrace(types.NoGithubSearchQuerySuppliedErr{})
	}

	// Build the search query
	searchQuery := config.GithubCodeSearch

	// If a specific organization is provided, add it to the query
	if config.GithubOrg != "" {
		searchQuery = fmt.Sprintf("%s org:%s", searchQuery, config.GithubOrg)
	}

	logger.WithFields(logrus.Fields{
		"Query": searchQuery,
	}).Debug("Searching for code using GitHub Code Search API")

	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		result, resp, err := config.GithubClient.Search.Code(context.Background(), searchQuery, opt)
		if err != nil {
			return allRepos, errors.WithStackTrace(err)
		}

		// Extract unique repositories from code search results
		for _, codeResult := range result.CodeResults {
			repo := codeResult.Repository
			if repo != nil {
				repoKey := repo.GetFullName()

				// Skip archived repos if --skip-archived-repos is passed
				if config.SkipArchivedRepos && repo.GetArchived() {
					logger.WithFields(logrus.Fields{
						"Name": repo.GetFullName(),
					}).Debug("Skipping archived repository from code search results")

					// Track repos to skip because of archived status for our final run report
					config.Stats.TrackSingle(stats.ReposArchivedSkipped, repo)
					continue
				}

				// Add to map to avoid duplicates
				repoMap[repoKey] = repo
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Convert map to slice
	for _, repo := range repoMap {
		allRepos = append(allRepos, repo)
	}

	repoCount := len(allRepos)

	if repoCount == 0 {
		return nil, errors.WithStackTrace(types.NoReposFoundFromSearchErr{Query: searchQuery})
	}

	logger.WithFields(logrus.Fields{
		"Repo count": repoCount,
		"Query":      searchQuery,
	}).Debug("Fetched repos from GitHub Code Search API")

	config.Stats.TrackMultiple(stats.FetchedViaGithubAPI, allRepos)

	return allRepos, nil
}
