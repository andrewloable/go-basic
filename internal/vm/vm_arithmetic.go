package vm

import (
	"math"
	"strings"
)

// ---------------------------------------------------------------------------
// Arithmetic opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execAdd() {
	b := vm.pop()
	a := vm.pop()
	if a.Type == ValString && b.Type == ValString {
		// String addition is concatenation.
		vm.push(StringVal(a.Str + b.Str))
	} else if a.Type == ValInt && b.Type == ValInt {
		vm.push(IntVal(a.Int + b.Int))
	} else {
		vm.push(FloatVal(a.asFloat() + b.asFloat()))
	}
}

func (vm *VM) execSub() {
	b := vm.pop()
	a := vm.pop()
	if a.Type == ValInt && b.Type == ValInt {
		vm.push(IntVal(a.Int - b.Int))
	} else {
		vm.push(FloatVal(a.asFloat() - b.asFloat()))
	}
}

func (vm *VM) execMul() {
	b := vm.pop()
	a := vm.pop()
	if a.Type == ValInt && b.Type == ValInt {
		vm.push(IntVal(a.Int * b.Int))
	} else {
		vm.push(FloatVal(a.asFloat() * b.asFloat()))
	}
}

func (vm *VM) execDiv() error {
	b := vm.pop()
	a := vm.pop()
	bf := b.asFloat()
	if bf == 0 {
		return vm.runtimeError("division by zero")
	}
	vm.push(FloatVal(a.asFloat() / bf))
	return nil
}

func (vm *VM) execIDiv() error {
	b := vm.pop()
	a := vm.pop()
	bi := b.asInt()
	if bi == 0 {
		return vm.runtimeError("division by zero")
	}
	vm.push(IntVal(a.asInt() / bi))
	return nil
}

func (vm *VM) execMod() error {
	b := vm.pop()
	a := vm.pop()
	bi := b.asInt()
	if bi == 0 {
		return vm.runtimeError("division by zero (MOD)")
	}
	vm.push(IntVal(a.asInt() % bi))
	return nil
}

func (vm *VM) execPow() {
	b := vm.pop()
	a := vm.pop()
	vm.push(FloatVal(math.Pow(a.asFloat(), b.asFloat())))
}

func (vm *VM) execNeg() {
	a := vm.pop()
	if a.Type == ValInt {
		vm.push(IntVal(-a.Int))
	} else {
		vm.push(FloatVal(-a.asFloat()))
	}
}

// ---------------------------------------------------------------------------
// Comparison opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execEq() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) == 0))
}

func (vm *VM) execNe() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) != 0))
}

func (vm *VM) execLt() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) < 0))
}

func (vm *VM) execGt() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) > 0))
}

func (vm *VM) execLe() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) <= 0))
}

func (vm *VM) execGe() {
	b := vm.pop()
	a := vm.pop()
	vm.push(boolVal(vm.compareValues(a, b) >= 0))
}

// ---------------------------------------------------------------------------
// Logical opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execAnd() {
	b := vm.pop()
	a := vm.pop()
	vm.push(IntVal(a.asInt() & b.asInt()))
}

func (vm *VM) execOr() {
	b := vm.pop()
	a := vm.pop()
	vm.push(IntVal(a.asInt() | b.asInt()))
}

func (vm *VM) execXor() {
	b := vm.pop()
	a := vm.pop()
	vm.push(IntVal(a.asInt() ^ b.asInt()))
}

func (vm *VM) execNot() {
	a := vm.pop()
	vm.push(IntVal(^a.asInt()))
}

func (vm *VM) execEqv() {
	b := vm.pop()
	a := vm.pop()
	vm.push(IntVal(^(a.asInt() ^ b.asInt())))
}

func (vm *VM) execImp() {
	b := vm.pop()
	a := vm.pop()
	vm.push(IntVal((^a.asInt()) | b.asInt()))
}

// ---------------------------------------------------------------------------
// String concat opcode handler
// ---------------------------------------------------------------------------

func (vm *VM) execConcat() {
	b := vm.pop()
	a := vm.pop()
	vm.push(StringVal(a.Str + b.Str))
}

// ---------------------------------------------------------------------------
// Comparison helper
// ---------------------------------------------------------------------------

// compareValues returns <0, 0, or >0 analogous to strcmp semantics.
func (vm *VM) compareValues(a, b Value) int {
	// String comparison.
	if a.Type == ValString && b.Type == ValString {
		return strings.Compare(a.Str, b.Str)
	}
	// Numeric comparison.
	af := a.asFloat()
	bf := b.asFloat()
	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}

// boolVal converts a Go bool to the BASIC true/false Value.
func boolVal(b bool) Value {
	if b {
		return TrueValue
	}
	return FalseValue
}
