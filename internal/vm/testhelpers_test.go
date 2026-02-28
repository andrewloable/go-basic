package vm

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
)

// ---------------------------------------------------------------------------
// Helper: compile BASIC source code to a Chunk
// ---------------------------------------------------------------------------

func compileSource(t *testing.T, src string) (*Chunk, *Compiler) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiler := NewCompiler(nil)
	chunk, err := compiler.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return chunk, compiler
}

// compileAndRun compiles source, runs it through the VM, returns captured output.
func compileAndRun(t *testing.T, src string) string {
	t.Helper()
	chunk, compiler := compileSource(t, src)
	v := NewVM(chunk)
	var buf strings.Builder
	v.SetOutput(&buf)
	if pool := compiler.DataPool(); len(pool) > 0 {
		v.SetDataPool(pool)
	}
	if err := v.Run(); err != nil {
		t.Fatalf("VM runtime error: %v", err)
	}
	return buf.String()
}

// makeChunk builds a chunk from instructions and constants.
func makeChunk(constants []Value, instructions []Instruction) *Chunk {
	lines := make([]int, len(instructions))
	for i := range lines {
		lines[i] = i + 1
	}
	return &Chunk{
		Code:      instructions,
		Constants: constants,
		Lines:     lines,
	}
}

func runVM(t *testing.T, chunk *Chunk) (*VM, string) {
	t.Helper()
	vm := NewVM(chunk)
	var buf strings.Builder
	vm.SetOutput(&buf)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run() error: %v", err)
	}
	return vm, buf.String()
}

// mustParse compiles BASIC source to a Chunk (fatals on parse/compile errors).
func mustParse(t *testing.T, src string) *Chunk {
	t.Helper()
	chunk, _ := compileSource(t, src)
	return chunk
}
