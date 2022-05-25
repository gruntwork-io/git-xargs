package cmd

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/git-xargs/config"
	"github.com/gruntwork-io/git-xargs/mocks"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

// A smoke test that you can define a basic config and pass it all the way through the main processing routine without
// any errors
func TestHandleRepoProcessing(t *testing.T) {
	t.Parallel()

	testConfig := config.NewGitXargsTestConfig()
	testConfig.ReposFile = "../data/test/good-test-repos.txt"
	testConfig.BranchName = "test-branch-name"
	testConfig.CommitMessage = "test-commit-name"
	testConfig.Args = []string{"touch", "test.txt"}
	testConfig.GithubClient = mocks.ConfigureMockGithubClient()
	testConfig.PullRequestRetries = 0
	testConfig.SecondsToSleepBetweenPRs = 1

	err := handleRepoProcessing(testConfig)
	assert.NoError(t, err)
}

func TestParseSliceFromReader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"one string", "foo", []string{"foo"}},
		{"one string with whitespace", "    foo\t\t\t", []string{"foo"}},
		{"multiple strings separated by whitespace", "foo bar     baz\t\tblah", []string{"foo", "bar", "baz", "blah"}},
		{"multiple strings separated by newlines", "foo\nbar\nbaz\nblah", []string{"foo", "bar", "baz", "blah"}},
		{"multiple strings separated by newlines, with extra newlines", "\n\nfoo\nbar\n\nbaz\nblah\n\n\n", []string{"foo", "bar", "baz", "blah"}},
	}

	for _, testCase := range testCases {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseSliceFromReader(strings.NewReader(testCase.input))
			require.NoError(t, err)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
