package main

import "github.com/go-git/go-git/v5"

type GitProvider interface {
	PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
}

type GitProductionProvider struct{}

func (g GitProductionProvider) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
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
