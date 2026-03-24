package cmd

import (
	"testing"
)

func TestGetNestedKey_TopLevel(t *testing.T) {
	m := map[string]interface{}{
		"staging_branch": "staging",
		"language":       "auto",
	}

	val, ok := getNestedKey(m, "staging_branch")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "staging" {
		t.Errorf("expected 'staging', got '%v'", val)
	}
}

func TestGetNestedKey_DotNotation(t *testing.T) {
	m := map[string]interface{}{
		"stack": map[string]interface{}{
			"language":  "go",
			"framework": "none",
		},
	}

	val, ok := getNestedKey(m, "stack.language")
	if !ok {
		t.Fatal("expected nested key to be found")
	}
	if val != "go" {
		t.Errorf("expected 'go', got '%v'", val)
	}
}

func TestGetNestedKey_DeepNesting(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep",
			},
		},
	}

	val, ok := getNestedKey(m, "a.b.c")
	if !ok {
		t.Fatal("expected deep nested key to be found")
	}
	if val != "deep" {
		t.Errorf("expected 'deep', got '%v'", val)
	}
}

func TestGetNestedKey_MissingKey(t *testing.T) {
	m := map[string]interface{}{
		"staging_branch": "staging",
	}

	_, ok := getNestedKey(m, "nonexistent")
	if ok {
		t.Error("expected key not found")
	}
}

func TestGetNestedKey_MissingNestedKey(t *testing.T) {
	m := map[string]interface{}{
		"stack": map[string]interface{}{
			"language": "go",
		},
	}

	_, ok := getNestedKey(m, "stack.missing")
	if ok {
		t.Error("expected nested key not found")
	}
}

func TestGetNestedKey_NonMapNesting(t *testing.T) {
	m := map[string]interface{}{
		"staging_branch": "staging",
	}

	_, ok := getNestedKey(m, "staging_branch.sub")
	if ok {
		t.Error("expected failure when trying to descend into non-map value")
	}
}
