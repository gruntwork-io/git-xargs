package main

import "fmt"

// AllowedRepo represents a single repository under a Github organization that this tool may operate on
type AllowedRepo struct {
	Organization string `header:"Organization name"`
	Name         string `header:"URL"`
}

// ReducedRepo is a simplified form of the github.Repository struct
type ReducedRepo struct {
	Name string `header:"Repo name"`
	URL  string `header:"Repo url"`
}

// OpenedPullRequest is a simple two column representation of the repo name and its PR url
type PullRequest struct {
	Repo string `header:"Repo name"`
	URL  string `header:"PR URL"`
}

// AnnotatedEvent is used in printing the final report. It contains the info to print a section's table - both it's Event for looking up the tagged repos, and the human-legible decommandion for printing above the table
type AnnotatedEvent struct {
	Event       Event
	Description string
}

type NoArgumentsPassedErr struct{}

func (NoArgumentsPassedErr) Error() string {
	return fmt.Sprint("You must pass a valid command or script path to git-xargs")
}

type NoGithubOrgSuppliedErr struct{}

func (NoGithubOrgSuppliedErr) Error() string {
	return fmt.Sprint("You must pass a valid Github organization name")
}

type NoRepoSelectionsMadeErr struct{}

func (NoRepoSelectionsMadeErr) Error() string {
	return fmt.Sprint("You must target some repos for processing either via stdin or by providing one of the --github-org, --repos, or --repo flags")
}

type NoBranchNameErr struct{}

func (NoBranchNameErr) Error() string {
	return fmt.Sprint("You must pass a branch name to use via the --branch-name flag")
}

type NoReposFoundErr struct {
	GithubOrg string
}

func (err NoReposFoundErr) Error() string {
	return fmt.Sprintf("No repos found for the organization supplied via --github-org: %s", err.GithubOrg)
}

type NoValidReposFoundAfterFilteringErr struct{}

func (NoValidReposFoundAfterFilteringErr) Error() string {
	return fmt.Sprint("No valid repos were found after filtering out malformed input")
}

type NoCommandSuppliedErr struct{}

func (NoCommandSuppliedErr) Error() string {
	return fmt.Sprintf("You must supply a valid command or script to execute")
}

type NoGithubOauthTokenProvidedErr struct{}

func (NoGithubOauthTokenProvidedErr) Error() string {
	return fmt.Sprintf("You must export a valid Github personal access token as GITHUB_OAUTH_TOKEN")
}
