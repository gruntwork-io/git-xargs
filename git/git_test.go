package main

import (
	"github.com/go-git/go-git/v5"
)

type MockGitProvider struct{}

func (g MockGitProvider) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {

	// Intercept the provided clone options and point to the locally checked out copy of github.com/gruntwork-io/fetch
	// to prevent any actual cloning or pushing being done to a real remote repo during testing
	o.URL = "./_testdata/test-repo"

	return git.PlainClone(path, isBare, o)
}
