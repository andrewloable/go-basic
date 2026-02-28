package vm

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: Opcode String()
// ---------------------------------------------------------------------------

func TestOpcodeString(t *testing.T) {
	if OpPush.String() != "PUSH" {
		t.Fatalf("expected PUSH, got %s", OpPush.String())
	}
	if OpHalt.String() != "HALT" {
		t.Fatalf("expected HALT, got %s", OpHalt.String())
	}
	// Unknown opcode.
	unknown := Opcode(255)
	s := unknown.String()
	if !strings.Contains(s, "UNKNOWN") {
		t.Fatalf("expected UNKNOWN in string, got %s", s)
	}
}

// ---------------------------------------------------------------------------
// Test: BuiltinID String()
// ---------------------------------------------------------------------------

func TestBuiltinIDString(t *testing.T) {
	if BuiltinAbs.String() != "ABS" {
		t.Fatalf("expected ABS, got %s", BuiltinAbs.String())
	}
	unknown := BuiltinID(255)
	s := unknown.String()
	if !strings.Contains(s, "UNKNOWN") {
		t.Fatalf("expected UNKNOWN in string, got %s", s)
	}
}

// ---------------------------------------------------------------------------
// Test: Chunk helpers
// ---------------------------------------------------------------------------

func TestChunkAddConstant(t *testing.T) {
	c := &Chunk{}
	idx0 := c.AddConstant(IntVal(1))
	idx1 := c.AddConstant(StringVal("foo"))
	if idx0 != 0 || idx1 != 1 {
		t.Fatalf("expected indices 0, 1 but got %d, %d", idx0, idx1)
	}
	if len(c.Constants) != 2 {
		t.Fatalf("expected 2 constants, got %d", len(c.Constants))
	}
}

func TestChunkEmit(t *testing.T) {
	c := &Chunk{}
	pos := c.Emit(OpPush, 0, 10)
	if pos != 0 {
		t.Fatalf("expected position 0, got %d", pos)
	}
	if len(c.Code) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(c.Code))
	}
	if c.Lines[0] != 10 {
		t.Fatalf("expected line 10, got %d", c.Lines[0])
	}
}

func TestChunkDisassemble(t *testing.T) {
	c := &Chunk{}
	c.Emit(OpPush, 0, 1)
	c.Emit(OpHalt, 0, 1)
	dis := c.Disassemble()
	if !strings.Contains(dis, "PUSH") || !strings.Contains(dis, "HALT") {
		t.Fatalf("disassembly missing expected instructions: %s", dis)
	}
}
