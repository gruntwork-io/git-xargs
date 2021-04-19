package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessAllowedReposErrsOnBadFilepath(t *testing.T) {
	t.Parallel()

	intentionallyBadFilepath := "_testdata/i-am-not-really-here.sh"
	allowedRepos, err := processAllowedRepos(intentionallyBadFilepath)

	assert.Error(t, err)
	assert.Equal(t, len(allowedRepos), 0)
}

func TestProcessAllowedReposCorrectlyParsesValidReposFile(t *testing.T) {
	t.Parallel()

	filepathToValidReposFile := "_testdata/good-test-repos.txt"
	allowedRepos, err := processAllowedRepos(filepathToValidReposFile)

	assert.NoError(t, err)
	assert.Equal(t, len(allowedRepos), 3)

	// Test that repo names are correctly parsed from the flat file by initiallly setting a map of each repo name
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

	filepathToReposFileWithSomeMalformedRepos := "_testdata/mixed-test-repos.txt"

	allowedRepos, err := processAllowedRepos(filepathToReposFileWithSomeMalformedRepos)
	assert.NoError(t, err)

	// There are 3 valid repos defined in this test file, and 3 intentionally malformed repos, so only 3 should
	// be returned by the function as valid repos to operate on
	assert.Equal(t, len(allowedRepos), 3)

	// Test that repo names are correctly parsed from the flat file by initiallly setting a map of each repo name
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
