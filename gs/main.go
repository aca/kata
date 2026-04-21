package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <repo>\n  e.g. %s github.com/xtdlib/pgmap\n", os.Args[0], os.Args[0])
		os.Exit(2)
	}

	repo := strings.TrimSuffix(os.Args[1], ".git")
	repo = strings.TrimSuffix(repo, "/")
	if strings.HasPrefix(repo, "https://") {
		repo = strings.TrimPrefix(repo, "https://")
	} else if strings.HasPrefix(repo, "http://") {
		repo = strings.TrimPrefix(repo, "http://")
	} else if strings.HasPrefix(repo, "git@") {
		// git@github.com:foo/bar -> github.com/foo/bar
		rest := strings.TrimPrefix(repo, "git@")
		rest = strings.Replace(rest, ":", "/", 1)
		repo = rest
	}

	url := "https://" + repo
	branch, err := defaultBranch(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect default branch: %v\n", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "home dir: %v\n", err)
		os.Exit(1)
	}

	dest := filepath.Join(home, "src", repo, branch)
	if _, err := os.Stat(dest); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists\n", dest)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("git", "clone", "--branch", branch, url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	fmt.Println(dest)
}

func defaultBranch(url string) (string, error) {
	out, err := exec.Command("git", "ls-remote", "--symref", url, "HEAD").Output()
	if err != nil {
		return "", err
	}
	// first line: "ref: refs/heads/<branch>\tHEAD"
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "ref: ") {
			rest := strings.TrimPrefix(line, "ref: ")
			ref := strings.Fields(rest)[0]
			return strings.TrimPrefix(ref, "refs/heads/"), nil
		}
	}
	return "", fmt.Errorf("no symref in ls-remote output")
}
