package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repo wraps a go-git repository with helper methods.
type Repo struct {
	R    *gogit.Repository
	Path string
}

// Open opens the git repository at the given path (or cwd if empty).
func Open(path string) (*Repo, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repo{R: r, Path: path}, nil
}

// CurrentBranch returns the current branch name.
func (r *Repo) CurrentBranch() string {
	ref, err := r.R.Head()
	if err != nil {
		return ""
	}
	if ref.Name().IsBranch() {
		return ref.Name().Short()
	}
	return ""
}

// DetectBaseBranch returns the staging branch if it exists, else "main" or "master".
// If stagingBranch is empty, defaults to "staging".
func (r *Repo) DetectBaseBranch(stagingBranch ...string) string {
	sb := "staging"
	if len(stagingBranch) > 0 && stagingBranch[0] != "" {
		sb = stagingBranch[0]
	}

	// Check staging branch first (local and remote)
	for _, refName := range []string{"refs/heads/" + sb, "refs/remotes/origin/" + sb} {
		_, err := r.R.Reference(plumbing.ReferenceName(refName), false)
		if err == nil {
			return sb
		}
	}

	// Check main/master
	for _, pair := range []struct {
		ref  string
		name string
	}{
		{"refs/heads/main", "main"},
		{"refs/remotes/origin/main", "main"},
		{"refs/heads/master", "master"},
		{"refs/remotes/origin/master", "master"},
	} {
		_, err := r.R.Reference(plumbing.ReferenceName(pair.ref), false)
		if err == nil {
			return pair.name
		}
	}

	return "main"
}

// DetectBaseBranchSimple returns the staging branch if it exists, else "main" (no master fallback).
// If stagingBranch is empty, defaults to "staging".
func (r *Repo) DetectBaseBranchSimple(stagingBranch ...string) string {
	sb := "staging"
	if len(stagingBranch) > 0 && stagingBranch[0] != "" {
		sb = stagingBranch[0]
	}

	for _, refName := range []string{"refs/heads/" + sb, "refs/remotes/origin/" + sb} {
		_, err := r.R.Reference(plumbing.ReferenceName(refName), false)
		if err == nil {
			return sb
		}
	}
	return "main"
}

// RemoteURL returns the URL of the given remote (default "origin").
func (r *Repo) RemoteURL(name string) string {
	if name == "" {
		name = "origin"
	}
	remote, err := r.R.Remote(name)
	if err != nil {
		return ""
	}
	urls := remote.Config().URLs
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

// HasStagingBranch checks if the repo has a staging branch (local or remote).
func HasStagingBranch(cwd, branchName string) bool {
	cmd := exec.Command("git", "branch", "-a", "--list", "*"+branchName+"*")
	if cwd != "" {
		cmd.Dir = cwd
	}
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// ProjectName returns the current directory name (used as project name).
func ProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}
