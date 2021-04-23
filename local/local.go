package local

import "github.com/go-git/go-git/v5"

type GitProvider interface {
	PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

type GitProductionProvider struct{}

func (g GitProductionProvider) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}

type MockGitProvider struct{}

func (g MockGitProvider) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {

	// Intercept the provided clone options and point to the locally checked out copy of github.com/gruntwork-io/fetch
	// to prevent any actual cloning or pushing being done to a real remote repo during testing
	o.URL = "../data/test/test-repo"

	return git.PlainClone(path, isBare, o)
}

type GitClient struct {
	GitProvider
}

func NewGitClient(provider GitProvider) GitClient {
	return GitClient{
		provider,
	}
}
