package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// generate is an internal helper: builds an ast.Program from stmts,
// creates an empty SymbolTable, then runs Generate.
func generate(t *testing.T, stmts []ast.Statement) string {
	t.Helper()
	prog := &ast.Program{Statements: stmts}
	table := semantic.NewSymbolTable()
	gen := New()
	out, err := gen.Generate(prog, table)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	return out
}

// ---------------------------------------------------------------------------
// Tests for codegen.go — New(), Generate(), emitProgram(), oneArg, argN
// ---------------------------------------------------------------------------

func TestEmptyProgram(t *testing.T) {
	out := generate(t, nil)
	if !strings.Contains(out, "package main") {
		t.Error("output missing 'package main'")
	}
	if !strings.Contains(out, "func main()") {
		t.Error("output missing 'func main()'")
	}
	// Should import the runtime.
	if !strings.Contains(out, `rt "github.com/loabletech/go-basic/internal/runtime"`) {
		t.Error("output missing runtime import")
	}
}

func TestOutputContainsPackageMain(t *testing.T) {
	out := generate(t, nil)
	if !strings.HasPrefix(out, "package main\n") {
		t.Errorf("output should start with 'package main', got:\n%.100s", out)
	}
}

func TestOneArgFallback(t *testing.T) {
	// Call oneArg directly from the internal test package.
	g := New()
	got := g.oneArg(nil)
	if got != "0" {
		t.Errorf("oneArg(nil) = %q, want %q", got, "0")
	}
	got2 := g.oneArg([]string{})
	if got2 != "0" {
		t.Errorf("oneArg([]) = %q, want %q", got2, "0")
	}
}

func TestOneArgNonEmpty(t *testing.T) {
	g := New()
	got := g.oneArg([]string{"myArg"})
	if got != "myArg" {
		t.Errorf("oneArg([myArg]) = %q, want %q", got, "myArg")
	}
}

func TestArgNFallback(t *testing.T) {
	g := New()
	got := g.argN(nil, 0)
	if got != "0" {
		t.Errorf("argN(nil, 0) = %q, want %q", got, "0")
	}
	// Index past end.
	got2 := g.argN([]string{"a", "b"}, 5)
	if got2 != "0" {
		t.Errorf("argN([a,b], 5) = %q, want %q", got2, "0")
	}
}

func TestArgNNonEmpty(t *testing.T) {
	g := New()
	got := g.argN([]string{"first", "second", "third"}, 2)
	if got != "third" {
		t.Errorf("argN([...], 2) = %q, want %q", got, "third")
	}
}

func TestAddPackageVar(t *testing.T) {
	gen := New()
	gen.addPackageVar("myVar", "float32")
	if !gen.sharedVars["myVar"] {
		t.Error("expected myVar in sharedVars after addPackageVar")
	}
	// Adding again shouldn't duplicate
	gen.addPackageVar("myVar", "float32")
	count := 0
	for _, v := range gen.packageVars {
		if v.name == "myVar" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected myVar to appear once in packageVars, got %d", count)
	}
}

func TestAddHoistedVar(t *testing.T) {
	gen := New()
	gen.addHoistedVar("v1", "int16")
	gen.addHoistedVar("v1", "int16") // duplicate — should not appear twice
	count := 0
	for _, v := range gen.hoistedVars {
		if v.name == "v1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected v1 to appear once in hoistedVars, got %d", count)
	}
}

func TestAddHoistedVarSkipsShared(t *testing.T) {
	gen := New()
	gen.sharedVars["sharedV"] = true
	gen.addHoistedVar("sharedV", "float32")
	for _, v := range gen.hoistedVars {
		if v.name == "sharedV" {
			t.Error("expected sharedV to be skipped in hoistedVars (already package-level)")
		}
	}
}

func TestEmitExprNil(t *testing.T) {
	// nil expression returns "0"
	gen := New()
	got := gen.emitExpr(nil)
	if got != "0" {
		t.Errorf("emitExpr(nil) = %q, want 0", got)
	}
}
