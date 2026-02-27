package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LoadFixture reads a test fixture file from testdata relative to the project root.
func LoadFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", path, err)
	}
	return string(data)
}

// AssertEqual compares two values and fails the test if they differ.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNoError fails the test if err is non-nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// GoldenFile compares got against a golden file. If the UPDATE_GOLDEN
// environment variable is set, it writes the golden file instead.
func GoldenFile(t *testing.T, goldenPath string, got string) {
	t.Helper()

	if os.Getenv("UPDATE_GOLDEN") != "" {
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create golden dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s (run with UPDATE_GOLDEN=1 to create): %v", goldenPath, err)
	}

	if got != string(want) {
		gotLines := strings.Split(got, "\n")
		wantLines := strings.Split(string(want), "\n")
		t.Errorf("output does not match golden file %s", goldenPath)
		maxLines := len(gotLines)
		if len(wantLines) > maxLines {
			maxLines = len(wantLines)
		}
		for i := 0; i < maxLines; i++ {
			g := ""
			w := ""
			if i < len(gotLines) {
				g = gotLines[i]
			}
			if i < len(wantLines) {
				w = wantLines[i]
			}
			if g != w {
				t.Errorf("  line %d:\n    got:  %q\n    want: %q", i+1, g, w)
			}
		}
	}
}
