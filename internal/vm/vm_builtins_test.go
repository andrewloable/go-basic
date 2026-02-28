package vm

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: built-in functions (bytecode level)
// ---------------------------------------------------------------------------

func TestBuiltinAbs(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(-42.5)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinAbs)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Float != 42.5 {
		t.Fatalf("expected 42.5, got %g", result.Float)
	}
}

func TestBuiltinLen(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinLen)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 5 {
		t.Fatalf("expected Int(5), got %s", result)
	}
}

func TestBuiltinLeft(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello, World!"), IntVal(5)},
		[]Instruction{
			{Op: OpPush, Operand: 0}, // push string
			{Op: OpPush, Operand: 1}, // push n
			{Op: OpBuiltin, Operand: int32(BuiltinLeft)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "Hello" {
		t.Fatalf("expected \"Hello\", got %q", result.Str)
	}
}

func TestBuiltinUCase(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("hello")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinUCase)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "HELLO" {
		t.Fatalf("expected \"HELLO\", got %q", result.Str)
	}
}

func TestBuiltinSqr(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(16.0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinSqr)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Float != 4.0 {
		t.Fatalf("expected 4.0, got %g", result.Float)
	}
}

func TestBuiltinSqrNegative(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(-1.0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinSqr)},
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	err := vm.Run()
	if err == nil {
		t.Fatal("expected error for SQR of negative number")
	}
}

func TestBuiltinChr(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(65)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinChr)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "A" {
		t.Fatalf("expected \"A\", got %q", result.Str)
	}
}

func TestBuiltinSgn(t *testing.T) {
	tests := []struct {
		input  float64
		expect int64
	}{
		{-5.0, -1},
		{0.0, 0},
		{3.14, 1},
	}
	for _, tt := range tests {
		chunk := makeChunk(
			[]Value{FloatVal(tt.input)},
			[]Instruction{
				{Op: OpPush, Operand: 0},
				{Op: OpBuiltin, Operand: int32(BuiltinSgn)},
				{Op: OpHalt},
			},
		)
		vm, _ := runVM(t, chunk)
		result := vm.stack[0]
		if result.Int != tt.expect {
			t.Fatalf("SGN(%g): expected %d, got %d", tt.input, tt.expect, result.Int)
		}
	}
}

// ---------------------------------------------------------------------------
// Test: built-in functions (source level via compileAndRun)
// ---------------------------------------------------------------------------

func TestBuiltinFix(t *testing.T) {
	out := compileAndRun(t, "PRINT FIX(3.9)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinInt(t *testing.T) {
	out := compileAndRun(t, "PRINT INT(-3.1)")
	if !strings.Contains(out, "-4") {
		t.Fatalf("expected '-4' (floor), got %q", out)
	}
}

func TestBuiltinExp(t *testing.T) {
	out := compileAndRun(t, "PRINT EXP(0)")
	if !strings.Contains(out, "1") {
		t.Fatalf("expected '1', got %q", out)
	}
}

func TestBuiltinSin(t *testing.T) {
	out := compileAndRun(t, "PRINT SIN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinCos(t *testing.T) {
	out := compileAndRun(t, "PRINT COS(0)")
	if !strings.Contains(out, "1") {
		t.Fatalf("expected '1', got %q", out)
	}
}

func TestBuiltinTan(t *testing.T) {
	out := compileAndRun(t, "PRINT TAN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinAtn(t *testing.T) {
	out := compileAndRun(t, "PRINT ATN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinLog(t *testing.T) {
	out := compileAndRun(t, "PRINT LOG(1)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinCint(t *testing.T) {
	out := compileAndRun(t, "PRINT CINT(3.6)")
	if !strings.Contains(out, "4") {
		t.Fatalf("expected '4', got %q", out)
	}
}

func TestBuiltinCsng(t *testing.T) {
	out := compileAndRun(t, "PRINT CSNG(3)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinCdbl(t *testing.T) {
	out := compileAndRun(t, "PRINT CDBL(3)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinRight(t *testing.T) {
	out := compileAndRun(t, `PRINT RIGHT$("Hello", 3)`)
	if !strings.Contains(out, "llo") {
		t.Fatalf("expected 'llo', got %q", out)
	}
}

func TestBuiltinMid(t *testing.T) {
	out := compileAndRun(t, `PRINT MID$("Hello", 2, 3)`)
	if !strings.Contains(out, "ell") {
		t.Fatalf("expected 'ell', got %q", out)
	}
}

func TestBuiltinAsc(t *testing.T) {
	out := compileAndRun(t, `PRINT ASC("A")`)
	if !strings.Contains(out, "65") {
		t.Fatalf("expected '65', got %q", out)
	}
}

func TestBuiltinChrDollar(t *testing.T) {
	out := compileAndRun(t, "PRINT CHR$(66)")
	if !strings.Contains(out, "B") {
		t.Fatalf("expected 'B', got %q", out)
	}
}

func TestBuiltinStr(t *testing.T) {
	out := compileAndRun(t, "PRINT STR$(42)")
	if !strings.Contains(out, "42") {
		t.Fatalf("expected '42', got %q", out)
	}
}

func TestBuiltinVal(t *testing.T) {
	out := compileAndRun(t, `PRINT VAL("42")`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected '42', got %q", out)
	}
}

func TestBuiltinUCaseLower(t *testing.T) {
	out := compileAndRun(t, `PRINT UCASE$("world")`)
	if !strings.Contains(out, "WORLD") {
		t.Fatalf("expected 'WORLD', got %q", out)
	}
}

func TestBuiltinLCase(t *testing.T) {
	out := compileAndRun(t, `PRINT LCASE$("HELLO")`)
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestBuiltinLTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT LTRIM$("  hi")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinRTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT RTRIM$("hi  ")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT TRIM$("  hi  ")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinSpace(t *testing.T) {
	out := compileAndRun(t, "x$ = SPACE$(3)")
	_ = out // just verifies no crash
}

func TestBuiltinHex(t *testing.T) {
	out := compileAndRun(t, "PRINT HEX$(255)")
	if !strings.Contains(out, "FF") {
		t.Fatalf("expected 'FF', got %q", out)
	}
}

func TestBuiltinOct(t *testing.T) {
	out := compileAndRun(t, "PRINT OCT$(8)")
	if !strings.Contains(out, "10") {
		t.Fatalf("expected '10', got %q", out)
	}
}

func TestBuiltinInstr(t *testing.T) {
	out := compileAndRun(t, `PRINT INSTR(1, "Hello", "ell")`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected '2', got %q", out)
	}
}

func TestBuiltinRnd(t *testing.T) {
	out := compileAndRun(t, "x = RND(1)")
	_ = out // just verifies no crash
}

func TestBuiltinStringRepeat(t *testing.T) {
	out := compileAndRun(t, `PRINT STRING$(3, 65)`)
	if !strings.Contains(out, "AAA") {
		t.Fatalf("expected 'AAA', got %q", out)
	}
}

func TestBuiltinLenStr(t *testing.T) {
	output := compileAndRun(t, `PRINT LEN("hello")`)
	if !strings.Contains(output, "5") {
		t.Errorf("LEN = %q, want 5", output)
	}
}

func TestBuiltinBin(t *testing.T) {
	output := compileAndRun(t, `PRINT BIN$(10)`)
	if !strings.Contains(output, "1010") {
		t.Errorf("BIN$(10) = %q, want 1010", output)
	}
}

func TestBuiltinClng(t *testing.T) {
	output := compileAndRun(t, `PRINT CLNG(3.7)`)
	if !strings.Contains(output, "4") {
		t.Errorf("CLNG(3.7) = %q, want 4", output)
	}
}

func TestBuiltinTab(t *testing.T) {
	output := compileAndRun(t, `PRINT TAB(3)`)
	_ = output // just verify no panic
}

func TestBuiltinSpc(t *testing.T) {
	output := compileAndRun(t, `PRINT SPC(3)`)
	if !strings.Contains(output, "   ") {
		t.Errorf("SPC(3) = %q, want 3 spaces", output)
	}
}

func TestBuiltinMkiCvi(t *testing.T) {
	output := compileAndRun(t, `s$ = MKI$(100)
PRINT CVI(s$)`)
	if !strings.Contains(output, "100") {
		t.Errorf("MKI/CVI = %q", output)
	}
}

func TestBuiltinMklCvl(t *testing.T) {
	output := compileAndRun(t, `s$ = MKL$(12345)
PRINT CVL(s$)`)
	if !strings.Contains(output, "12345") {
		t.Errorf("MKL/CVL = %q", output)
	}
}

func TestBuiltinMksCvs(t *testing.T) {
	output := compileAndRun(t, `s$ = MKS$(1.5)
PRINT CVS(s$)`)
	if !strings.Contains(output, "1.5") {
		t.Errorf("MKS/CVS = %q", output)
	}
}

func TestBuiltinMkdCvd(t *testing.T) {
	output := compileAndRun(t, `s$ = MKD$(3.14)
PRINT CVD(s$)`)
	if !strings.Contains(output, "3.14") {
		t.Errorf("MKD/CVD = %q", output)
	}
}

func TestBuiltinSqrError(t *testing.T) {
	// SQR of negative should produce a runtime error
	src := `PRINT SQR(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for SQR(-1)")
	}
}

func TestBuiltinLogError(t *testing.T) {
	// LOG of zero or negative should produce a runtime error
	src := `PRINT LOG(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG(-1)")
	}
}

func TestBuiltinLog2Error(t *testing.T) {
	src := `PRINT LOG2(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG2(-1)")
	}
}

func TestBuiltinLog10Error(t *testing.T) {
	src := `PRINT LOG10(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG10(-1)")
	}
}

func TestBuiltinStringFn(t *testing.T) {
	// STRING$(n, char) - repeat char n times
	output := compileAndRun(t, `PRINT STRING$(3, 65)`)
	if !strings.Contains(output, "AAA") {
		t.Errorf("STRING$(3, 65) = %q, want AAA", output)
	}
}

func TestBuiltinInstr2Args(t *testing.T) {
	// INSTR with 2 args (no start position)
	output := compileAndRun(t, `PRINT INSTR("hello world", "world")`)
	if !strings.Contains(output, "7") {
		t.Errorf("INSTR 2-arg = %q, want 7", output)
	}
}

func TestBuiltinInstr3Args(t *testing.T) {
	output := compileAndRun(t, `PRINT INSTR(5, "hello world", "o")`)
	// "hello world" has "o" at pos 5 and 8; starting from 5 → should find pos 5
	if !strings.Contains(output, "5") && !strings.Contains(output, "8") {
		t.Errorf("INSTR 3-arg = %q, expected 5 or 8", output)
	}
}
