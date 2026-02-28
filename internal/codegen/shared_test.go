package codegen_test

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/codegen"
	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
)

func TestSharedFullPipeline(t *testing.T) {
	src := `DIM Name1$ AS STRING
DIM NumGames AS INTEGER

Name1$ = "Alice"
NumGames = 5

CALL ShowInfo

SUB ShowInfo
  SHARED Name1$, NumGames
  PRINT Name1$
  PRINT NumGames
END SUB
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("PARSE ERROR: %s", e)
		}
	}
	resolver := semantic.NewResolver(program)
	table, sErrs := resolver.Resolve()
	if len(sErrs) > 0 {
		for _, e := range sErrs {
			t.Logf("SEM ERROR: %s", e)
		}
	}
	gen := codegen.New()
	goSrc, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("CODEGEN ERROR: %v", err)
	}
	t.Logf("Generated:\n%s", goSrc)

	// Package-level vars should exist
	if !strings.Contains(goSrc, "var Name1_str string") {
		t.Error("missing package-level Name1_str")
	}
	if !strings.Contains(goSrc, "var NumGames float32") {
		t.Error("missing package-level NumGames")
	}

	// These should appear before func main()
	mainIdx := strings.Index(goSrc, "func main()")
	n1Idx := strings.Index(goSrc, "var Name1_str string")
	if n1Idx > mainIdx {
		t.Error("package-level var should be before main")
	}

	// SUB should not re-declare shared vars
	subIdx := strings.Index(goSrc, "func ShowInfo()")
	if subIdx < 0 {
		t.Fatal("missing ShowInfo function")
	}
	subBody := goSrc[subIdx:]
	if strings.Contains(subBody, "var Name1_str") {
		t.Error("SUB should not re-declare Name1_str")
	}
}
