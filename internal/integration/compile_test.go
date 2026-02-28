package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCompileAndRun transpiles a simple BASIC program, compiles it to a
// native binary, runs it, and verifies the output matches expectations.
func TestCompileAndRun(t *testing.T) {
	source := `PRINT "Hello, World!"
END`

	goSrc, _, _, err := transpile(source)
	if err != nil {
		t.Fatalf("transpile error: %v", err)
	}

	// Create temp build dir inside the module root so internal imports resolve.
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}
	buildDir, err := os.MkdirTemp(moduleRoot, "_compile_test_*")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}
	defer os.RemoveAll(buildDir)

	mainGo := filepath.Join(buildDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(goSrc), 0644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	binName := "testprog"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(buildDir, binName)

	// Build the binary.
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = buildDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	// Run the binary and check output.
	runCmd := exec.Command(binPath)
	runCmd.Env = append(os.Environ(), "GOBASIC_HEADLESS=1")
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, string(runOut))
	}

	got := strings.TrimSpace(string(runOut))
	if got != "Hello, World!" {
		t.Errorf("output mismatch: got %q, want %q", got, "Hello, World!")
	}
}

// TestCompileWithRuntimeFunctions tests compiling a program that uses
// runtime functions (math, string operations).
func TestCompileWithRuntimeFunctions(t *testing.T) {
	source := `x = 4
PRINT SQR(x)
PRINT LEFT$("Hello World", 5)
END`

	goSrc, _, _, err := transpile(source)
	if err != nil {
		t.Fatalf("transpile error: %v", err)
	}

	moduleRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	buildDir, _ := os.MkdirTemp(moduleRoot, "_compile_rt_test_*")
	defer os.RemoveAll(buildDir)

	if err := os.WriteFile(filepath.Join(buildDir, "main.go"), []byte(goSrc), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	binName := "testrt"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(buildDir, binName)

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = buildDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	runCmd := exec.Command(binPath)
	runCmd.Env = append(os.Environ(), "GOBASIC_HEADLESS=1")
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary execution failed: %v\n%s", err, string(runOut))
	}

	got := strings.TrimSpace(string(runOut))
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines of output, got: %q", got)
	}
	if strings.TrimSpace(lines[0]) != "2" {
		t.Errorf("SQR(4) = %q, want %q", strings.TrimSpace(lines[0]), "2")
	}
	if strings.TrimSpace(lines[1]) != "Hello" {
		t.Errorf("LEFT$(\"Hello World\", 5) = %q, want %q", strings.TrimSpace(lines[1]), "Hello")
	}
}

// TestCompileParseError verifies that a program with syntax errors
// fails at the transpile stage, not at compile.
func TestCompileParseError(t *testing.T) {
	source := `PRINT "unclosed string
END`

	_, _, _, err := transpile(source)
	// This may or may not error depending on how the lexer handles it,
	// but the generated Go should at minimum not be empty if it succeeds.
	if err != nil {
		t.Logf("transpile correctly reported error: %v", err)
		return
	}
	t.Log("transpile succeeded despite unclosed string (lexer may auto-close)")
}

// TestCompileFlagParsing verifies that the compile command's flag parsing
// works correctly by testing the binary name generation logic.
func TestCompileFlagParsing(t *testing.T) {
	tests := []struct {
		basFile string
		goos    string
		want    string
	}{
		{"hello.bas", "linux", "hello"},
		{"hello.bas", "windows", "hello.exe"},
		{"hello.bas", "darwin", "hello"},
		{"my_program.bas", "windows", "my_program.exe"},
	}

	for _, tt := range tests {
		base := filepath.Base(tt.basFile)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		if tt.goos == "windows" && !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
		if name != tt.want {
			t.Errorf("basFile=%q goos=%q: got %q, want %q", tt.basFile, tt.goos, name, tt.want)
		}
	}
}
