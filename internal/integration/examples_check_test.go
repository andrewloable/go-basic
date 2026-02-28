package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// examplesDir returns the absolute path to the examples/ directory at the repo root.
func examplesDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("failed to resolve examples path: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("examples directory not found at %s", dir)
	}
	return dir
}

// examplesBasPrograms returns all .bas files in the examples/ directory.
func examplesBasPrograms(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.bas"))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no .bas files found in examples/")
	}
	return matches
}

// TestCheckExamples transpiles every .bas file in examples/, writes the
// generated Go to a temp dir, then runs `go build` on each to check for
// compile errors. Results are reported per-file with failure category.
func TestCheckExamples(t *testing.T) {
	dir := examplesDir(t)
	programs := examplesBasPrograms(t, dir)

	// Temp directory for generated Go files.
	tmpBase, err := os.MkdirTemp("", "go-basic-examples-check-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpBase)

	// Determine the module root so `go build` can resolve the runtime package.
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("failed to resolve module root: %v", err)
	}

	type result struct {
		name      string
		parseErrs []string
		semErrs   []string
		codegenErr error
		buildErr   string
		ok         bool
	}

	var results []result

	for _, basFile := range programs {
		name := strings.TrimSuffix(filepath.Base(basFile), ".bas")

		sourceBytes, err := os.ReadFile(basFile)
		if err != nil {
			t.Logf("[%s] SKIP: cannot read file: %v", name, err)
			continue
		}
		source := string(sourceBytes)

		// --- Transpile ---
		goSrc, parseErrs, semErrs, codegenErr := transpile(source)

		r := result{
			name:       name,
			parseErrs:  parseErrs,
			semErrs:    semErrs,
			codegenErr: codegenErr,
		}

		if codegenErr != nil {
			results = append(results, r)
			continue
		}

		if goSrc == "" {
			r.buildErr = "codegen produced empty output"
			results = append(results, r)
			continue
		}

		// --- Write generated Go to its own package directory ---
		pkgDir := filepath.Join(tmpBase, name)
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			t.Logf("[%s] SKIP: mkdir failed: %v", name, err)
			continue
		}

		goFile := filepath.Join(pkgDir, "main.go")
		if err := os.WriteFile(goFile, []byte(goSrc), 0644); err != nil {
			t.Logf("[%s] SKIP: write failed: %v", name, err)
			continue
		}

		// Write a go.mod that references the local module so the runtime
		// package import resolves correctly.
		goModContent := fmt.Sprintf(`module example/%s

go 1.23.1

require github.com/loabletech/go-basic v0.0.0

replace github.com/loabletech/go-basic => %s
`, name, moduleRoot)
		if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goModContent), 0644); err != nil {
			t.Logf("[%s] SKIP: write go.mod failed: %v", name, err)
			continue
		}

		// --- go mod tidy then go build ---
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = pkgDir
		tidyCmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
		tidyOut, tidyErr := tidyCmd.CombinedOutput()
		if tidyErr != nil {
			r.buildErr = fmt.Sprintf("go mod tidy failed: %v\n%s", tidyErr, string(tidyOut))
			results = append(results, r)
			continue
		}

		buildCmd := exec.Command("go", "build", ".")
		buildCmd.Dir = pkgDir
		buildOut, buildErr := buildCmd.CombinedOutput()
		if buildErr != nil {
			r.buildErr = fmt.Sprintf("%s", strings.TrimSpace(string(buildOut)))
		} else {
			r.ok = true
		}

		results = append(results, r)
	}

	// --- Report ---
	var (
		okCount        int
		parseErrFiles  []string
		semErrFiles    []string
		codegenFails   []string
		buildFails     []result
	)

	for _, r := range results {
		switch {
		case r.codegenErr != nil:
			codegenFails = append(codegenFails, fmt.Sprintf("  %s: %v", r.name, r.codegenErr))
			if len(r.parseErrs) > 0 {
				parseErrFiles = append(parseErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.parseErrs, "; ")))
			}
			if len(r.semErrs) > 0 {
				semErrFiles = append(semErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.semErrs, "; ")))
			}
		case r.buildErr != "":
			buildFails = append(buildFails, r)
			if len(r.parseErrs) > 0 {
				parseErrFiles = append(parseErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.parseErrs, "; ")))
			}
			if len(r.semErrs) > 0 {
				semErrFiles = append(semErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.semErrs, "; ")))
			}
		case r.ok:
			okCount++
			if len(r.parseErrs) > 0 {
				parseErrFiles = append(parseErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.parseErrs, "; ")))
			}
			if len(r.semErrs) > 0 {
				semErrFiles = append(semErrFiles, fmt.Sprintf("  %s: %s", r.name, strings.Join(r.semErrs, "; ")))
			}
		}
	}

	total := len(results)
	t.Logf("=== SUMMARY ===")
	t.Logf("Total .bas files processed: %d", total)
	t.Logf("Go build OK: %d", okCount)
	t.Logf("Codegen failures: %d", len(codegenFails))
	t.Logf("Go build failures: %d", len(buildFails))
	t.Logf("")

	if len(parseErrFiles) > 0 {
		t.Logf("=== PARSE WARNINGS (non-fatal) ===")
		for _, s := range parseErrFiles {
			t.Logf("%s", s)
		}
		t.Logf("")
	}

	if len(semErrFiles) > 0 {
		t.Logf("=== SEMANTIC WARNINGS (non-fatal) ===")
		for _, s := range semErrFiles {
			t.Logf("%s", s)
		}
		t.Logf("")
	}

	if len(codegenFails) > 0 {
		t.Logf("=== CODEGEN FAILURES ===")
		for _, s := range codegenFails {
			t.Logf("%s", s)
		}
		t.Logf("")
	}

	if len(buildFails) > 0 {
		t.Logf("=== GO BUILD FAILURES ===")
		for _, r := range buildFails {
			t.Logf("  --- %s ---", r.name)
			t.Logf("  %s", r.buildErr)
			t.Logf("")
		}
	}

	// Fail the test if anything went wrong so CI catches it.
	if len(codegenFails) > 0 || len(buildFails) > 0 {
		t.Errorf("%d codegen failure(s) and %d Go build failure(s) found (see logs above)",
			len(codegenFails), len(buildFails))
	}
}
