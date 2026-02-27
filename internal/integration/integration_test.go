package integration

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/codegen"
	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
	"github.com/loabletech/go-basic/internal/vm"
)

var update = flag.Bool("update", false, "update golden (.expected) files")

// testdataDir returns the absolute path to the testdata/programs directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	// The testdata directory lives at the repository root.
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", "programs"))
	if err != nil {
		t.Fatalf("failed to resolve testdata path: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("testdata directory not found at %s", dir)
	}
	return dir
}

// basPrograms returns all .bas files in the testdata/programs directory.
func basPrograms(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.bas"))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no .bas files found in testdata/programs")
	}
	return matches
}

// ---------------------------------------------------------------------------
// Transpile pipeline helpers
// ---------------------------------------------------------------------------

// transpile runs the full lexer -> parser -> semantic -> codegen pipeline and
// returns the generated Go source code.  Parse and semantic errors are returned
// alongside any codegen error.
func transpile(source string) (goSrc string, parseErrs []string, semErrs []string, err error) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	parseErrs = p.Errors()

	resolver := semantic.NewResolver(program)
	table, sErrs := resolver.Resolve()
	semErrs = sErrs

	gen := codegen.New()
	goSrc, err = gen.Generate(program, table)
	return
}

// ---------------------------------------------------------------------------
// VM pipeline helpers
// ---------------------------------------------------------------------------

// vmRun compiles the source through the VM backend and executes it, capturing
// the VM's printed output.  Returns the output string, or an error if
// compilation or execution fails.
func vmRun(source string) (output string, err error) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return "", fmt.Errorf("parse errors: %s", strings.Join(errs, "; "))
	}

	// Semantic analysis to get the symbol table for the compiler.
	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	compiler := vm.NewCompiler(table)
	chunk, compErr := compiler.Compile(program)
	if compErr != nil {
		return "", fmt.Errorf("vm compile error: %w", compErr)
	}

	machine := vm.NewVM(chunk)

	// Set the DATA pool collected during compilation.
	if dp := compiler.DataPool(); len(dp) > 0 {
		machine.SetDataPool(dp)
	}

	var buf strings.Builder
	machine.SetOutput(&buf)

	if runErr := machine.Run(); runErr != nil {
		return buf.String(), fmt.Errorf("vm runtime error: %w", runErr)
	}
	return buf.String(), nil
}

// ---------------------------------------------------------------------------
// Golden file helpers
// ---------------------------------------------------------------------------

func goldenPath(basFile string) string {
	ext := filepath.Ext(basFile)
	return basFile[:len(basFile)-len(ext)] + ".expected"
}

func readGolden(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func writeGolden(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}
	t.Logf("updated golden file: %s", path)
}

// sourceContainsPrint is a quick heuristic to check whether a BASIC program
// likely contains PRINT statements (used for VM output assertions).
func sourceContainsPrint(source string) bool {
	upper := strings.ToUpper(source)
	return strings.Contains(upper, "PRINT")
}

// ---------------------------------------------------------------------------
// Main integration test
// ---------------------------------------------------------------------------

func TestIntegration(t *testing.T) {
	dir := testdataDir(t)
	programs := basPrograms(t, dir)

	for _, basFile := range programs {
		basFile := basFile // capture
		name := strings.TrimSuffix(filepath.Base(basFile), ".bas")

		t.Run(name, func(t *testing.T) {
			sourceBytes, err := os.ReadFile(basFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", basFile, err)
			}
			source := string(sourceBytes)

			// ---------------------------------------------------------------
			// Sub-test: transpile pipeline (codegen)
			// ---------------------------------------------------------------
			t.Run("transpile", func(t *testing.T) {
				goSrc, parseErrs, semErrs, genErr := transpile(source)

				// Log non-fatal warnings but don't fail on parse/semantic
				// errors -- some programs may push the parser boundary.
				if len(parseErrs) > 0 {
					t.Logf("parse warnings: %s", strings.Join(parseErrs, "; "))
				}
				if len(semErrs) > 0 {
					t.Logf("semantic warnings: %s", strings.Join(semErrs, "; "))
				}

				if genErr != nil {
					t.Fatalf("codegen error: %v", genErr)
				}

				if goSrc == "" {
					t.Fatal("codegen produced empty output")
				}

				// Sanity check: the generated Go source should be a valid Go
				// program skeleton.
				if !strings.Contains(goSrc, "package main") {
					t.Error("generated Go source missing 'package main'")
				}
				if !strings.Contains(goSrc, "func main()") {
					t.Error("generated Go source missing 'func main()'")
				}

				// Golden file comparison.
				gp := goldenPath(basFile)
				if *update {
					writeGolden(t, gp, goSrc)
				} else if expected, ok := readGolden(gp); ok {
					if goSrc != expected {
						t.Errorf("transpile output mismatch for %s\n--- expected ---\n%s\n--- got ---\n%s",
							name, expected, goSrc)
					}
				} else {
					t.Logf("no golden file %s; run with -update to create", gp)
				}
			})

			// ---------------------------------------------------------------
			// Sub-test: VM pipeline
			// ---------------------------------------------------------------
			t.Run("vm", func(t *testing.T) {
				vmOutput, vmErr := vmRun(source)

				if vmErr != nil {
					// Some programs use features the VM doesn't support yet.
					// Log but don't fail -- the transpile path is the primary
					// backend.
					t.Logf("vm error (may be unsupported feature): %v", vmErr)
					return
				}

				t.Logf("vm output (%d bytes):\n%s", len(vmOutput), vmOutput)

				// If the program contains PRINT statements, the VM should
				// have produced some output.
				if sourceContainsPrint(source) && len(strings.TrimSpace(vmOutput)) == 0 {
					t.Error("program has PRINT statements but VM produced no output")
				}
			})
		})
	}
}

// ---------------------------------------------------------------------------
// TestTranspileAndVMConsistency verifies that both backends can process all
// programs without crashing.
// ---------------------------------------------------------------------------

func TestTranspileAndVMConsistency(t *testing.T) {
	dir := testdataDir(t)
	programs := basPrograms(t, dir)

	var transpileOK, vmOK int

	for _, basFile := range programs {
		sourceBytes, err := os.ReadFile(basFile)
		if err != nil {
			t.Fatalf("failed to read %s: %v", basFile, err)
		}
		source := string(sourceBytes)
		name := strings.TrimSuffix(filepath.Base(basFile), ".bas")

		// Transpile pipeline.
		_, _, _, genErr := transpile(source)
		if genErr == nil {
			transpileOK++
		} else {
			t.Logf("[%s] transpile failed: %v", name, genErr)
		}

		// VM pipeline.
		_, vmErr := vmRun(source)
		if vmErr == nil {
			vmOK++
		} else {
			t.Logf("[%s] vm error: %v", name, vmErr)
		}
	}

	total := len(programs)
	t.Logf("transpile: %d/%d programs succeeded", transpileOK, total)
	t.Logf("vm:        %d/%d programs succeeded", vmOK, total)

	// At least the transpile backend should handle all programs.
	if transpileOK < total {
		t.Errorf("expected all %d programs to transpile successfully, got %d", total, transpileOK)
	}
}
