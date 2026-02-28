package integration

import (
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckExamplesBuild transpiles all .bas files in examples/ and runs
// "go build" on each, verifying the generated Go compiles without errors.
func TestCheckExamplesBuild(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.bas"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no .bas files found: %v", err)
	}

	// Resolve the module root (where go.mod lives) as an absolute path.
	// The replace directive in go.mod must use an absolute path because
	// the temp dir is not relative to the project root.
	moduleRoot, err2 := filepath.Abs(filepath.Join("..", ".."))
	if err2 != nil {
		t.Fatalf("failed to resolve module root: %v", err2)
	}

	ok, fail := 0, 0
	for _, f := range files {
		src, _ := os.ReadFile(f)
		goSrc, _, _, genErr := transpile(string(src))
		name := strings.TrimSuffix(filepath.Base(f), ".bas")
		if genErr != nil {
			t.Errorf("[%s] codegen error: %v", name, genErr)
			fail++
			continue
		}

		// Write the generated Go to a temp package INSIDE the project module.
		// This is required because the generated code imports internal/runtime,
		// and Go's internal package restriction only allows imports from within
		// the same module tree.
		tmpDir, merr := os.MkdirTemp(moduleRoot, "_build_"+name+"_*")
		if merr != nil {
			t.Errorf("[%s] mkdir error: %v", name, merr)
			fail++
			continue
		}
		defer os.RemoveAll(tmpDir)

		goFile := filepath.Join(tmpDir, "main.go")
		if werr := os.WriteFile(goFile, []byte(goSrc), 0644); werr != nil {
			t.Errorf("[%s] write error: %v", name, werr)
			fail++
			continue
		}
		cmd := exec.Command("go", "build", ".")
		cmd.Dir = tmpDir
		out, berr := cmd.CombinedOutput()
		if berr != nil {
			t.Errorf("[%s] go build failed:\n%s", name, strings.TrimSpace(string(out)))
			fail++
			continue
		}
		ok++
	}
	t.Logf("Build OK: %d/%d", ok, ok+fail)
	if fail > 0 {
		t.Errorf("%d files failed to build", fail)
	}
}

func TestCheckExamplesCompileSyntax(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.bas"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no .bas files found: %v", err)
	}
	ok, fail := 0, 0
	for _, f := range files {
		src, _ := os.ReadFile(f)
		goSrc, parseErrs, semErrs, genErr := transpile(string(src))
		name := strings.TrimSuffix(filepath.Base(f), ".bas")
		if genErr != nil {
			t.Errorf("[%s] codegen error: %v", name, genErr)
			fail++
			continue
		}
		if len(parseErrs) > 0 {
			t.Logf("[%s] parse warnings: %s", name, strings.Join(parseErrs, "; "))
		}
		if len(semErrs) > 0 {
			t.Logf("[%s] semantic warnings: %s", name, strings.Join(semErrs, "; "))
		}
		if _, fmtErr := format.Source([]byte(goSrc)); fmtErr != nil {
			t.Errorf("[%s] Go syntax error: %v", name, fmtErr)
			fail++
			continue
		}
		ok++
	}
	t.Logf("Syntax OK: %d/%d", ok, ok+fail)
	if fail > 0 {
		t.Errorf("%d files failed", fail)
	}
}
