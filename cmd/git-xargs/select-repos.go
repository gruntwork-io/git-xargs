package main

import (
	"github.com/google/go-github/v32/github"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"

	"github.com/sirupsen/logrus"
)

type RepoSelectionCriteria string

const (
	ReposViaStdIn              RepoSelectionCriteria = "repo-stdin"
	ExplicitReposOnCommandLine RepoSelectionCriteria = "repo-flag"
	ReposFilePath              RepoSelectionCriteria = "repos-file"
	GithubOrganization         RepoSelectionCriteria = "github-org"
)

// getPreferredOrderOfRepoSelections codifies the order in which flags will be preferred when the user supplied more
// than one:
// 1. --github-org is a string representing the Github org to page through via API for all repos.
// 2. --repos is a string representing a filepath to a repos file
// 3. --repo is a string slice flag that can be called multiple times
// 4. stdin allows you to pipe repos in from other CLI tools
func getPreferredOrderOfRepoSelections(config *GitXargsConfig) RepoSelectionCriteria {
	if config.GithubOrg != "" {
		return GithubOrganization
	}
	if config.ReposFile != "" {
		return ReposFilePath
	}
	if len(config.RepoSlice) > 0 {
		return ExplicitReposOnCommandLine
	}
	return ReposViaStdIn
}

// RepoSelection is a struct that presents a uniform interface to present to OperateRepos that converts
// user-supplied repos in the format of <org-name>/<repo-name> to Github API response objects that we actually
// pass into processRepos which does the git cloning, command execution, comitting and pull request opening
type RepoSelection struct {
	SelectionType          RepoSelectionCriteria
	AllowedRepos           []*AllowedRepo
	GithubOrganizationName string
}

func (r RepoSelection) GetCriteria() RepoSelectionCriteria {
	return r.SelectionType
}

func (r RepoSelection) GetAllowedRepos() []*AllowedRepo {
	return r.AllowedRepos
}

func (r RepoSelection) GetGithubOrg() string {
	return r.GithubOrganizationName
}

// selectReposViaInput will examine the various repo and github-org flags to determine which should be selected and processed (only one at a time is used)
func selectReposViaInput(config *GitXargsConfig) (*RepoSelection, error) {

	def := &RepoSelection{
		SelectionType:          GithubOrganization,
		AllowedRepos:           []*AllowedRepo{},
		GithubOrganizationName: config.GithubOrg,
	}

	switch getPreferredOrderOfRepoSelections(config) {
	case ExplicitReposOnCommandLine:
		allowedRepos, err := selectReposViaRepoFlag(config.RepoSlice)
		if err != nil {
			return def, err
		}
		return &RepoSelection{
			SelectionType:          ExplicitReposOnCommandLine,
			AllowedRepos:           allowedRepos,
			GithubOrganizationName: "",
		}, nil

	case ReposFilePath:
		allowedRepos, err := processAllowedRepos(config.ReposFile)
		if err != nil {
			return def, err
		}
		return &RepoSelection{
			SelectionType:          ReposFilePath,
			AllowedRepos:           allowedRepos,
			GithubOrganizationName: "",
		}, nil

	case GithubOrganization:
		return def, nil

	case ReposViaStdIn:
		allowedRepos, err := selectReposViaRepoFlag(config.RepoFromStdIn)
		if err != nil {
			return def, err
		}
		return &RepoSelection{
			SelectionType:          ReposViaStdIn,
			AllowedRepos:           allowedRepos,
			GithubOrganizationName: "",
		}, nil

	default:
		return def, nil
	}
}

// selectReposViaRepoFlag converts the string slice of repo flags provided via stdin or by invocations of the --repo
// flag into the internal representation of AllowedRepo that we we use prior to fetching the corresponding repo from
// Github
func selectReposViaRepoFlag(inputRepos []string) ([]*AllowedRepo, error) {
	var allowedRepos []*AllowedRepo

	for _, repoInput := range inputRepos {
		allowedRepo := convertStringToAllowedRepo(repoInput)
		if allowedRepo != nil {
			allowedRepos = append(allowedRepos, allowedRepo)
		}
	}
	if len(allowedRepos) < 1 {
		return allowedRepos, errors.WithStackTrace(NoRepoSelectionsMadeErr{})
	}

	return allowedRepos, nil
}

// fetchUserProvidedReposViaGithub converts repos provided as strings, already validated as being well-formed, into Github API repo objects that can be further processed
func fetchUserProvidedReposViaGithubAPI(githubClient GithubClient, rs RepoSelection, stats *RunStats) ([]*github.Repository, error) {
	ar := rs.GetAllowedRepos()
	return getFileDefinedRepos(githubClient, ar, stats)

}

// There are three ways to select repos to operate on via this tool:
// 1. the --repo flag, which specifies a single repo, and which can be passed multiple times, e.g., --repo gruntwork-io/fetch --repo gruntwork-io/cloud-nuke, etc
// 2. the --repos flag which specifies the path to the user-defined flatfile of repos in the format of 'gruntwork-io/cloud-nuke', one repo per line
// 3. the --github-org flag which specifies the Github organization that should have all its repos fetched via API
//
// This function acts as a switch, depending upon whether or not the user provided an explicit list of repos to operate
//
// However, even though there are two methods for users to select repos, we still only want a single uniform interface
// for dealing with a repo throughout this tool, and that is the *github.Repository type provided by the go-github
// library. Therefore, this function serves the purpose of creating that uniform interface, by looking up flatfile-provided
// repos via go-github, so that we're only ever dealing with pointers to github.Repositories going forward
func OperateOnRepos(config *GitXargsConfig) error {

	logger := logging.GetLogger("git-xargs")

	// The set of github repositories the tool will actually process
	var reposToIterate []*github.Repository

	// repoSelection is a representations of the user-supplied input, containing the repo organization and name
	repoSelection, err := selectReposViaInput(config)

	if err != nil {
		return err
	}

	switch repoSelection.GetCriteria() {

	case GithubOrganization:
		// If githubOrganization is set, the user did not provide a flatfile or explicit repos via the -repo(s) flags, so we're just looking up all the Github
		// repos via their Organization name via the Github API
		reposFetchedFromGithubAPI, err := getReposByOrg(config)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"Error":        err,
				"Organization": config.GithubOrg,
			}).Debug("Failure looking up repos for organization")
			return err
		}
		// We gather all the repos by fetching them from the Github API, paging through the results of the supplied organization
		reposToIterate = reposFetchedFromGithubAPI

		logger.Debugf("Using Github org: %s as source of repositories. Paging through Github API for repos.", config.GithubOrg)

	case ReposFilePath:
		githubRepos, err := fetchUserProvidedReposViaGithubAPI(config.GithubClient, *repoSelection, config.Stats)
		if err != nil {
			return err
		}

		reposToIterate = githubRepos

		// Update count of number of repos the the tool read in from the provided file
		config.Stats.SetFileProvidedRepos(repoSelection.GetAllowedRepos())

	case ExplicitReposOnCommandLine, ReposViaStdIn:
		githubRepos, err := fetchUserProvidedReposViaGithubAPI(config.GithubClient, *repoSelection, config.Stats)
		if err != nil {
			return err
		}

		reposToIterate = githubRepos // Update the count of number of repos the tool read in from explicit --repo flags
		config.Stats.SetRepoFlagProvidedRepos(repoSelection.GetAllowedRepos())

	default:
		// We've got no repos to iterate on, so return an error
		return errors.WithStackTrace(NoValidReposFoundAfterFilteringErr{})
	}

	// Track the repos selected for processing
	config.Stats.TrackMultiple(ReposSelected, reposToIterate)

	// Print out the repos that we've filtered for processing in debug mode
	for _, repo := range reposToIterate {
		logger.WithFields(logrus.Fields{
			"Repository": repo.GetName(),
		}).Debug("Repo will have all targeted scripts run against it")
	}
	// Now that we've gathered up the repos we're going to operate on, do the actual processing by running the
	// user-defined scripts against each repo and handling the resulting git operations that follow
	if err := processRepos(config, reposToIterate); err != nil {
		return err
	}

	return nil
}
