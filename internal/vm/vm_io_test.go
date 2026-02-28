package vm

import (
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: Hello World bytecode program
// ---------------------------------------------------------------------------

func TestHelloWorld(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello, World!")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintNewline},
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	if output != "Hello, World!\n" {
		t.Fatalf("expected \"Hello, World!\\n\", got %q", output)
	}
}

func TestPrintNumbers(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintNewline},
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	// BASIC prints positive numbers with a leading and trailing space.
	if output != " 42 \n" {
		t.Fatalf("expected \" 42 \\n\", got %q", output)
	}
}

func TestPrintTabAndSemicolon(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("A"), StringVal("B")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintTab},
			{Op: OpPush, Operand: 1},
			{Op: OpPrint},
			{Op: OpPrintSemicolon}, // should suppress newline
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	expected := "A\tB"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// ---------------------------------------------------------------------------
// Test: DATA / READ / RESTORE
// ---------------------------------------------------------------------------

func TestDataReadRestore(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("a"), StringVal("b")},
		[]Instruction{
			{Op: OpRead, Operand: 0},    // read into "a"
			{Op: OpRead, Operand: 1},    // read into "b"
			{Op: OpRestore, Operand: -1}, // restore to beginning
			{Op: OpRead, Operand: 0},    // read into "a" again
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	vm.SetDataPool([]Value{IntVal(10), IntVal(20)})
	var buf strings.Builder
	vm.SetOutput(&buf)
	err := vm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After: a=10 (re-read after restore), b=20
	if vm.globals["a"].Int != 10 {
		t.Fatalf("expected a=10, got %s", vm.globals["a"])
	}
	if vm.globals["b"].Int != 20 {
		t.Fatalf("expected b=20, got %s", vm.globals["b"])
	}
}

func TestReadOutOfData(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("x")},
		[]Instruction{
			{Op: OpRead, Operand: 0},
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	// No data pool set — should error.
	err := vm.Run()
	if err == nil {
		t.Fatal("expected 'out of DATA' error")
	}
	if !strings.Contains(err.Error(), "out of DATA") {
		t.Fatalf("expected 'out of DATA' in error, got: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Test: execInput — via stdin pipe
// ---------------------------------------------------------------------------

func TestExecInput(t *testing.T) {
	// execInput is never emitted by the compiler (INPUT uses DATA/READ path, known issue 6is.1).
	// Test it directly via Chunk construction, bypassing the compiler.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := w.WriteString("hello\n"); err != nil {
		t.Fatalf("pipe write: %v", err)
	}
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	// Operand -1 means no prompt string.
	chunk := makeChunk(
		[]Value{},
		[]Instruction{
			{Op: OpInput, Operand: -1},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if len(vm.stack) == 0 {
		t.Fatal("execInput: expected value on stack")
	}
	if vm.stack[0].asString() != "hello" {
		t.Errorf("execInput: expected 'hello', got %q", vm.stack[0].asString())
	}
}

func TestExecInputWithPrompt(t *testing.T) {
	// Test execInput with a prompt string constant (Operand >= 0).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := w.WriteString("world\n"); err != nil {
		t.Fatalf("pipe write: %v", err)
	}
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	// Operand 0 points to constants[0] which is the prompt string.
	chunk := makeChunk(
		[]Value{StringVal("Enter: ")},
		[]Instruction{
			{Op: OpInput, Operand: 0},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if len(vm.stack) == 0 {
		t.Fatal("execInputWithPrompt: expected value on stack")
	}
	if vm.stack[0].asString() != "world" {
		t.Errorf("execInputWithPrompt: expected 'world', got %q", vm.stack[0].asString())
	}
}

// ---------------------------------------------------------------------------
// Test: execPrintSemicolon — no-op function, verify no crash via program
// ---------------------------------------------------------------------------

func TestExecPrintSemicolonNoCrash(t *testing.T) {
	// PRINT a; b emits OpPrintSemicolon — verify it runs without error.
	output := compileAndRun(t, `PRINT 1; 2; 3`)
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Errorf("PRINT with semicolons = %q, want 1 2 3", output)
	}
}
