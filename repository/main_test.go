package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestMain gives the tests their own git identity, so that the commits they create don't depend on the git config of
// the machine running them.
func TestMain(m *testing.M) {
	os.Exit(runWithTestGitConfig(m))
}

func runWithTestGitConfig(m *testing.M) int {
	dir, err := os.MkdirTemp("", "git-xargs-test-git-config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating temp dir for git config: %v\n", err)
		return 1
	}
	defer os.RemoveAll(dir)

	configPath := filepath.Join(dir, "git", "config")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "creating git config dir: %v\n", err)
		return 1
	}
	if err := os.WriteFile(configPath, []byte("[user]\n\tname = Grunty\n\temail = grunty@gruntwork.io\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "writing git config: %v\n", err)
		return 1
	}

	// go-git reads $XDG_CONFIG_HOME/git/config before ~/.gitconfig, and the git CLI reads GIT_CONFIG_GLOBAL instead
	// of ~/.gitconfig.
	os.Setenv("XDG_CONFIG_HOME", dir)
	os.Setenv("GIT_CONFIG_GLOBAL", configPath)

	return m.Run()
}
