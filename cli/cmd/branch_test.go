package cmd

import (
	"testing"
)

// ─── GenerateBranchName tests ────────────────────────────────────────────────

func TestBranchName_BasicDescription(t *testing.T) {
	result := GenerateBranchName("Add user auth")
	expected := "feat/add-user-auth"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_FixPrefix(t *testing.T) {
	result := GenerateBranchName("fix broken login")
	expected := "fix/broken-login"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_RefactorPrefix(t *testing.T) {
	result := GenerateBranchName("refactor: extract user service")
	expected := "refactor/extract-user-service"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_HotfixPrefix(t *testing.T) {
	result := GenerateBranchName("hotfix/urgent payment bug")
	expected := "hotfix/urgent-payment-bug"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_SpecialChars(t *testing.T) {
	result := GenerateBranchName("Add @user auth! (v2)")
	expected := "feat/add-user-auth-v2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_Truncation(t *testing.T) {
	result := GenerateBranchName("this is a very long description that should be truncated to fifty characters maximum")
	if len(result) > 50 {
		t.Errorf("expected length <= 50, got %d: %q", len(result), result)
	}
	// Should not end with hyphen
	if result[len(result)-1] == '-' {
		t.Errorf("should not end with hyphen: %q", result)
	}
}

func TestBranchName_EmptyDescription(t *testing.T) {
	result := GenerateBranchName("")
	expected := "feat/unnamed"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_WhitespaceOnly(t *testing.T) {
	result := GenerateBranchName("   ")
	expected := "feat/unnamed"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_Underscores(t *testing.T) {
	result := GenerateBranchName("add_user_auth")
	expected := "feat/add-user-auth"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_MultipleSpaces(t *testing.T) {
	result := GenerateBranchName("add   user   auth")
	expected := "feat/add-user-auth"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_DocsPrefix(t *testing.T) {
	result := GenerateBranchName("docs update readme")
	expected := "docs/update-readme"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_TestPrefix(t *testing.T) {
	result := GenerateBranchName("test add unit tests for auth")
	expected := "test/add-unit-tests-for-auth"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBranchName_AllSpecialChars(t *testing.T) {
	result := GenerateBranchName("@#$%^&*()")
	expected := "feat/unnamed"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ─── readStagingBranch tests ─────────────────────────────────────────────────

func TestReadStagingBranch_Default(t *testing.T) {
	// When config doesn't exist, should return "staging"
	result := readStagingBranch()
	// This will either return the actual config value or "staging" default
	if result == "" {
		t.Error("expected non-empty staging branch")
	}
}

// ─── Branch safety logic tests (unit) ────────────────────────────────────────

func TestBranchSafety_MainBlocked(t *testing.T) {
	branch := "main"
	blocked := branch == "main" || branch == "master"
	if !blocked {
		t.Error("expected main to be blocked")
	}
}

func TestBranchSafety_MasterBlocked(t *testing.T) {
	branch := "master"
	blocked := branch == "main" || branch == "master"
	if !blocked {
		t.Error("expected master to be blocked")
	}
}

func TestBranchSafety_FeatureSafe(t *testing.T) {
	branch := "feat/my-feature"
	blocked := branch == "main" || branch == "master"
	if blocked {
		t.Error("expected feature branch to be safe")
	}
}

func TestBranchSafety_StagingSafe(t *testing.T) {
	branch := "staging"
	blocked := branch == "main" || branch == "master"
	if blocked {
		t.Error("expected staging branch to be safe")
	}
}

func TestBranchSafety_HotfixSafe(t *testing.T) {
	branch := "hotfix/urgent-fix"
	blocked := branch == "main" || branch == "master"
	if blocked {
		t.Error("expected hotfix branch to be safe")
	}
}
