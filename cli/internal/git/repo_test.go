package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temp dir with a git repo and at least one commit so HEAD
// is valid. It returns the repo path and a cleanup function.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		// Provide minimal git identity so commits don't fail in CI.
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", args, out)
		}
	}

	run("git", "init", dir)
	run("git", "-C", dir, "checkout", "-b", "main")

	// Create a file and commit so HEAD exists.
	readmeFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("init"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	run("git", "-C", dir, "add", ".")
	run("git", "-C", dir, "commit", "-m", "init")

	return dir
}

// ── ProjectName ──────────────────────────────────────────────────────────────

func TestProjectName_ReturnsCwdBaseName(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got := ProjectName()
	want := filepath.Base(dir)
	if got != want {
		t.Errorf("ProjectName() = %q, want %q", got, want)
	}
}

// ── Section ───────────────────────────────────────────────────────────────────

func TestSection_FormatsHeader(t *testing.T) {
	title := "hello"
	result := Section(title)

	if !strings.HasPrefix(result, "\n── ") {
		t.Errorf("Section() should start with newline and ── prefix, got: %q", result)
	}
	if !strings.Contains(result, title) {
		t.Errorf("Section() should contain title %q, got: %q", title, result)
	}
}

func TestSection_ShortTitleGetsPadding(t *testing.T) {
	short := "hi"       // len = 2  → padding = 52
	long := "hello"     // len = 5  → padding = 49

	shortResult := Section(short)
	longResult := Section(long)

	// Short title → more trailing dashes than long title.
	shortDashes := strings.Count(shortResult, "─") - 2 // subtract the leading ──
	longDashes := strings.Count(longResult, "─") - 2

	if shortDashes <= longDashes {
		t.Errorf("shorter title should produce more padding dashes: short=%d long=%d", shortDashes, longDashes)
	}
}

func TestSection_VeryLongTitleGetsZeroPadding(t *testing.T) {
	// 55 chars > 54, so padding = 0 and no trailing dashes.
	title := strings.Repeat("x", 55)
	result := Section(title)

	// After "── <title> " there should be no trailing ─ characters.
	suffix := strings.TrimPrefix(result, "\n── "+title+" ")
	if strings.Contains(suffix, "─") {
		t.Errorf("very long title should produce no trailing dashes, got suffix: %q", suffix)
	}
}

// ── Open ──────────────────────────────────────────────────────────────────────

func TestOpen_ValidRepo(t *testing.T) {
	dir := initRepo(t)

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%q) unexpected error: %v", dir, err)
	}
	if repo == nil {
		t.Fatal("Open() returned nil repo")
	}
}

func TestOpen_NonGitDirectory(t *testing.T) {
	dir := t.TempDir() // plain directory, no git

	_, err := Open(dir)
	if err == nil {
		t.Fatal("Open() should return error for non-git directory")
	}
}

func TestOpen_EmptyPathUsesCwd(t *testing.T) {
	dir := initRepo(t)

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	repo, err := Open("")
	if err != nil {
		t.Fatalf("Open(\"\") unexpected error: %v", err)
	}
	if repo == nil {
		t.Fatal("Open(\"\") returned nil repo")
	}
}

// ── CurrentBranch ─────────────────────────────────────────────────────────────

func TestCurrentBranch_ReturnsBranchName(t *testing.T) {
	dir := initRepo(t) // already on main with one commit

	// Create and switch to a feature branch.
	cmd := exec.Command("git", "checkout", "-b", "feat/my-feature")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout -b: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.CurrentBranch()
	if got != "feat/my-feature" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "feat/my-feature")
	}
}

func TestCurrentBranch_OnMain(t *testing.T) {
	dir := initRepo(t) // initRepo leaves us on main

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.CurrentBranch()
	if got != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "main")
	}
}

// ── DetectBaseBranch ──────────────────────────────────────────────────────────

func TestDetectBaseBranch_NoStagingHasMain(t *testing.T) {
	dir := initRepo(t) // has main, no staging

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranch()
	if got != "main" {
		t.Errorf("DetectBaseBranch() = %q, want %q", got, "main")
	}
}

func TestDetectBaseBranch_DefaultStagingBranchIsStaging(t *testing.T) {
	dir := initRepo(t)

	// Create a "staging" branch.
	cmd := exec.Command("git", "branch", "staging")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch staging: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranch()
	if got != "staging" {
		t.Errorf("DetectBaseBranch() = %q, want %q", got, "staging")
	}
}

func TestDetectBaseBranch_CustomStagingBranch(t *testing.T) {
	dir := initRepo(t)

	// Create a "develop" branch to use as the custom staging branch.
	cmd := exec.Command("git", "branch", "develop")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch develop: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranch("develop")
	if got != "develop" {
		t.Errorf("DetectBaseBranch(\"develop\") = %q, want %q", got, "develop")
	}
}

// ── HasStagingBranch ──────────────────────────────────────────────────────────

func TestHasStagingBranch_ReturnsFalseForNonExistent(t *testing.T) {
	dir := initRepo(t)

	if HasStagingBranch(dir, "staging") {
		t.Error("HasStagingBranch() should return false when branch does not exist")
	}
}

func TestHasStagingBranch_ReturnsTrueForExistingBranch(t *testing.T) {
	dir := initRepo(t)

	cmd := exec.Command("git", "branch", "staging")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch staging: %s", out)
	}

	if !HasStagingBranch(dir, "staging") {
		t.Error("HasStagingBranch() should return true when branch exists")
	}
}

// ── Run ───────────────────────────────────────────────────────────────────────

func TestRun_SuccessReturnsStdout(t *testing.T) {
	// "git version" is always available and produces non-empty output.
	stdout, _, err := Run("git", "version")
	if err != nil {
		t.Fatalf("Run(\"git\", \"version\") unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "git version") {
		t.Errorf("Run() stdout = %q, want it to contain \"git version\"", stdout)
	}
}

func TestRun_NonZeroExitReturnsError(t *testing.T) {
	_, _, err := Run("git", "rev-parse", "refs/heads/branch-that-does-not-exist-xyz")
	if err == nil {
		t.Fatal("Run() should return error for non-zero exit command")
	}
}

func TestRun_NonExistentCommandReturnsError(t *testing.T) {
	_, _, err := Run("nonexistent-command-xyz-abc")
	if err == nil {
		t.Fatal("Run() should return error for nonexistent command")
	}
}

// ── CommitsSince ──────────────────────────────────────────────────────────────

// makeCommit creates a new file and commits it in the given repo dir.
func makeCommit(t *testing.T, dir, filename, content, message string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	for _, args := range [][]string{
		{"git", "-C", dir, "add", filename},
		{"git", "-C", dir, "commit", "-m", message},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = gitEnv
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", args, out)
		}
	}
}

func TestCommitsSince_ReturnsLogSinceBase(t *testing.T) {
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@test.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@test.com")

	dir := initRepo(t)
	makeCommit(t, dir, "second.txt", "second", "second commit")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()

	out, err := CommitsSince("HEAD~1")
	if err != nil {
		t.Fatalf("CommitsSince(\"HEAD~1\") unexpected error: %v", err)
	}
	if !strings.Contains(out, "second commit") {
		t.Errorf("CommitsSince() = %q, want it to contain \"second commit\"", out)
	}
}

// ── ChangedFiles ──────────────────────────────────────────────────────────────

func TestChangedFiles_ReturnsFilesChangedSinceBase(t *testing.T) {
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@test.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@test.com")

	dir := initRepo(t)
	makeCommit(t, dir, "feature.go", "package main", "add feature.go")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()

	files, err := ChangedFiles("HEAD~1")
	if err != nil {
		t.Fatalf("ChangedFiles(\"HEAD~1\") unexpected error: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("ChangedFiles() returned empty slice, want at least one file")
	}
	found := false
	for _, f := range files {
		if f == "feature.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ChangedFiles() = %v, want it to contain \"feature.go\"", files)
	}
}

// ── DiffStat ──────────────────────────────────────────────────────────────────

func TestDiffStat_ReturnsNonEmptyStat(t *testing.T) {
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@test.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@test.com")

	dir := initRepo(t)
	makeCommit(t, dir, "stats.go", "package main\n\nfunc Foo() {}", "add stats.go")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()

	stat, err := DiffStat("HEAD~1")
	if err != nil {
		t.Fatalf("DiffStat(\"HEAD~1\") unexpected error: %v", err)
	}
	if stat == "" {
		t.Error("DiffStat() returned empty string, want non-empty stat")
	}
}

// ── DetectBaseBranchSimple ────────────────────────────────────────────────────

func TestDetectBaseBranchSimple_NoStagingReturnsMain(t *testing.T) {
	dir := initRepo(t) // has main, no staging branch

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranchSimple()
	if got != "main" {
		t.Errorf("DetectBaseBranchSimple() = %q, want \"main\"", got)
	}
}

func TestDetectBaseBranchSimple_WithStagingBranchReturnsStagingName(t *testing.T) {
	dir := initRepo(t)

	// Create a "staging" branch so it exists.
	cmd := exec.Command("git", "branch", "staging")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch staging: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranchSimple()
	if got != "staging" {
		t.Errorf("DetectBaseBranchSimple() = %q, want \"staging\"", got)
	}
}

func TestDetectBaseBranchSimple_CustomStagingBranchName(t *testing.T) {
	dir := initRepo(t)

	// Create a "release" branch to use as a custom staging branch.
	cmd := exec.Command("git", "branch", "release")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch release: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.DetectBaseBranchSimple("release")
	if got != "release" {
		t.Errorf("DetectBaseBranchSimple(\"release\") = %q, want \"release\"", got)
	}
}

// ── RemoteURL ─────────────────────────────────────────────────────────────────

func TestRemoteURL_NoRemoteReturnsEmpty(t *testing.T) {
	dir := initRepo(t) // fresh repo, no remotes

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.RemoteURL("origin")
	if got != "" {
		t.Errorf("RemoteURL(\"origin\") = %q, want \"\" (no remote configured)", got)
	}
}

func TestRemoteURL_ReturnsConfiguredURL(t *testing.T) {
	dir := initRepo(t)

	const wantURL = "https://github.com/test/test.git"
	cmd := exec.Command("git", "remote", "add", "origin", wantURL)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.RemoteURL("origin")
	if got != wantURL {
		t.Errorf("RemoteURL(\"origin\") = %q, want %q", got, wantURL)
	}
}

func TestRemoteURL_EmptyNameDefaultsToOrigin(t *testing.T) {
	dir := initRepo(t)

	const wantURL = "https://github.com/test/repo.git"
	cmd := exec.Command("git", "remote", "add", "origin", wantURL)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add: %s", out)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got := repo.RemoteURL("") // empty name → should default to "origin"
	if got != wantURL {
		t.Errorf("RemoteURL(\"\") = %q, want %q", got, wantURL)
	}
}

// ── ProjectName (error path) ──────────────────────────────────────────────────

func TestProjectName_ReturnsBaseNameOfArbitraryDir(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got := ProjectName()
	if got == "" {
		t.Error("ProjectName() returned empty string, want directory base name")
	}
	if got != filepath.Base(dir) {
		t.Errorf("ProjectName() = %q, want %q", got, filepath.Base(dir))
	}
}
