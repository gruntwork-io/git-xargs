package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCustomErrorStatements performs quick sanity checks on the custom error types to ensure they return the expected messages
func TestCustomErrorStatements(t *testing.T) {
	t.Parallel()

	err := &NoArgumentsPassedErr{}
	assert.Equal(t, "You must pass a valid command or script path to git-xargs", err.Error())

	errNoGithubOrg := &NoGithubOrgSuppliedErr{}
	assert.Equal(t, "You must pass a valid Github organization name", errNoGithubOrg.Error())

	errNoRepoSelected := &NoRepoSelectionsMadeErr{}
	assert.Equal(t, "You must target some repos for processing either via stdin or by providing one of the --github-org, --repos, or --repo flags", errNoRepoSelected.Error())

	errNoReposFound := &NoReposFoundErr{GithubOrg: "gruntwork-io"}
	assert.Equal(t, "No repos found for the organization supplied via --github-org: gruntwork-io", errNoReposFound.Error())

	errNoValidReposFoundAfterFiltering := NoValidReposFoundAfterFilteringErr{}
	assert.Equal(t, "No valid repos were found after filtering out malformed input", errNoValidReposFoundAfterFiltering.Error())

	errNoCommandSupplied := NoCommandSuppliedErr{}
	assert.Equal(t, "You must supply a valid command or script to execute", errNoCommandSupplied.Error())

	errNoGithubOauthTokenProvided := NoGithubOauthTokenProvidedErr{}
	assert.Equal(t, "You must export a valid Github personal access token as GITHUB_OAUTH_TOKEN", errNoGithubOauthTokenProvided.Error())

}
