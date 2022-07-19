package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessAllowedReposErrsOnBadFilepath(t *testing.T) {
	t.Parallel()

	intentionallyBadFilepath := "../data/test/i-am-not-really-here.sh"
	allowedRepos, err := ProcessAllowedRepos(intentionallyBadFilepath)

	assert.Error(t, err)
	assert.Equal(t, len(allowedRepos), 0)
}

func TestProcessAllowedReposCorrectlyParsesValidReposFile(t *testing.T) {
	t.Parallel()

	filepathToValidReposFile := "../data/test/test-file-parsing.txt"
	allowedRepos, err := ProcessAllowedRepos(filepathToValidReposFile)

	assert.NoError(t, err)
	assert.Equal(t, len(allowedRepos), 3)

	// Test that repo names are correctly parsed from the flat file by initially setting a map of each repo name
	// to false, and then updating each entry to true as we find them in the flat file. At the end, all map entries should be true / seen
	mapOfExpectedRepoNames := make(map[string]bool)
	mapOfExpectedRepoNames["fetch"] = false
	mapOfExpectedRepoNames["cloud-nuke"] = false
	mapOfExpectedRepoNames["bash-commons"] = false

	// ensure all test repos have the correct gruntwork-io org
	for _, repo := range allowedRepos {
		assert.Equal(t, repo.Organization, "gruntwork-io")
		// Update the map as having "seen" the repo
		mapOfExpectedRepoNames[repo.Name] = true
	}

	for _, v := range mapOfExpectedRepoNames {
		assert.True(t, v)
	}
}

func TestProcessAllowedReposCorrectlyFiltersMalformedInput(t *testing.T) {
	t.Parallel()

	filepathToReposFileWithSomeMalformedRepos := "../data/test/mixed-test-repos.txt"

	allowedRepos, err := ProcessAllowedRepos(filepathToReposFileWithSomeMalformedRepos)
	assert.NoError(t, err)

	// There are 3 valid repos defined in this test file, and 3 intentionally malformed repos, so only 3 should
	// be returned by the function as valid repos to operate on
	assert.Equal(t, len(allowedRepos), 3)

	// Test that repo names are correctly parsed from the flat file by initially setting a map of each repo name
	// to false, and then updating each entry to true as we find them in the flat file. At the end, all map entries should be true / seen
	mapOfExpectedRepoNames := make(map[string]bool)
	mapOfExpectedRepoNames["fetch"] = false
	mapOfExpectedRepoNames["cloud-nuke"] = false
	mapOfExpectedRepoNames["bash-commons"] = false

	// ensure all test repos have the correct gruntwork-io org
	for _, repo := range allowedRepos {
		assert.Equal(t, repo.Organization, "gruntwork-io")
		// Update the map as having "seen" the repo
		mapOfExpectedRepoNames[repo.Name] = true
	}

	for _, v := range mapOfExpectedRepoNames {
		assert.True(t, v)
	}
}
