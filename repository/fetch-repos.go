package repository

import (
	"context"
	"fmt"
	"strings"

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

	if config.GithubSearchQuery == "" {
		return nil, errors.WithStackTrace(types.NoGithubSearchQuerySuppliedErr{})
	}

	// Determine if this should be a code search or repository search
	if isCodeSearchQuery(config.GithubSearchQuery) {
		logger.WithFields(logrus.Fields{
			"Query": config.GithubSearchQuery,
		}).Debug("Detected code search query, using GitHub Code Search API")
		return getReposByCodeSearch(config)
	} else {
		logger.WithFields(logrus.Fields{
			"Query": config.GithubSearchQuery,
		}).Debug("Detected repository search query, using GitHub Repository Search API")
		return getReposByRepositorySearch(config)
	}
}

// getReposByRepositorySearch uses GitHub's repository search API to find repositories matching the given query
func getReposByRepositorySearch(config *config.GitXargsConfig) ([]*github.Repository, error) {
	logger := logging.GetLogger("git-xargs")

	var allRepos []*github.Repository

	// Build the search query
	searchQuery := config.GithubSearchQuery

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

	if config.GithubSearchQuery == "" {
		return allRepos, errors.WithStackTrace(types.NoGithubSearchQuerySuppliedErr{})
	}

	// Build the search query
	searchQuery := config.GithubSearchQuery

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

// isCodeSearchQuery determines if a query should use code search instead of repository search
// Code search queries typically contain file-specific qualifiers like path:, filename:, extension:
// or content search terms without repository-specific qualifiers
func isCodeSearchQuery(query string) bool {
	codeSearchIndicators := []string{
		"path:",
		"filename:",
		"extension:",
		"in:file",
		"in:path",
	}

	for _, indicator := range codeSearchIndicators {
		if strings.Contains(query, indicator) {
			return true
		}
	}

	// If the query doesn't contain typical repository search qualifiers and isn't obviously
	// a repository search, it's likely a code search
	repoSearchIndicators := []string{
		"language:",
		"topic:",
		"is:public",
		"is:private",
		"is:internal",
		"archived:",
		"fork:",
		"mirror:",
		"template:",
		"stars:",
		"forks:",
		"size:",
		"pushed:",
		"created:",
		"updated:",
	}

	hasRepoIndicator := false
	for _, indicator := range repoSearchIndicators {
		if strings.Contains(query, indicator) {
			hasRepoIndicator = true
			break
		}
	}

	// If it has no repository indicators and contains text that could be code content,
	// treat it as code search
	return !hasRepoIndicator
}
