package main

import (
	"fmt"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func NewTestFileName() string {
	return fmt.Sprintf("test-file-%s", RandStringBytes(9))
}

func NewGitXargsTestConfig() *GitXargsConfig {

	config := NewGitXargsConfig()

	uniqueID := RandStringBytes(9)
	config.BranchName = fmt.Sprintf("test-branch-%s", uniqueID)
	config.CommitMessage = fmt.Sprintf("commit-message-%s", uniqueID)
	config.GitClient = NewGitClient(MockGitProvider{})

	return config

}
