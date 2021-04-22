package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProcessRepo smoke tests the processRepo function with a basic test config - however, the MockGitProvider implemented
// in git_test.go intercepts the call to git.PlainClone to modify the repo URL to the local checkout of gruntwork-io/fetch
// which is bundled in _testdata to allow tests to run against an actual repository without making any network calls or pushes to actual remote repositories
func TestProcessRepo(t *testing.T) {
	t.Parallel()

	// Hackily create a simple git repo at ./testdata/test-repo if it doesn't already exist
	cmd := exec.Command("bash", "-c", "mkdir -p test-repo && cd test-repo && git init && touch README.md && git add README.md && git commit -m \"Add README.md\"")
	cmd.Dir = "./_testdata/"
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Error creating test git repo at ./_testdata/test-repo: +%v\n", err)
		t.Log(string(cmdOut))
	} else {
		t.Log("TestProcessRepo Successfully created test git repo at ./_testdata/test-repo")
	}

	config := NewGitXargsTestConfig()
	config.Args = []string{"touch", NewTestFileName()}
	config.GithubClient = configureMockGithubClient()

	// Run a command to delete all local branches in the "./_testdata/test-repo" repo to avoid the git-xargs repo
	// growing in size over time with test data
	defer cleanupLocalTestRepoChanges(t, config)

	processErr := processRepo(config, getMockGithubRepo())
	assert.NoError(t, processErr)
}

func cleanupLocalTestRepoChanges(t *testing.T, config *GitXargsConfig) {
	t.Log("cleanupLocalTestRepoChanges deleting branches in local test repo to avoid bloat...")
	// Force delete all of the branches that are not either "master" or "main"
	cmd := exec.Command("bash", "-c", "git branch | grep -v 'master' | grep -v '*' | xargs -r  git branch -D")
	cmd.Dir = "./_testdata/test-repo"
	cmdOut, err := cmd.CombinedOutput()
	t.Log(string(cmdOut))
	if err != nil {
		t.Logf("cleanupLocalTestRepoChanges error deleting test branches: %+v\n", err)
	} else {
		t.Log("cleanupLocalTestRepoChanges succesfully deleted branches in local test repo")
	}
}
