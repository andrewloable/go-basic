package vm

import (
	"github.com/loabletech/go-basic/internal/runtime"
)

// ---------------------------------------------------------------------------
// I/O opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execPrint() {
	v := vm.pop()
	vm.printStr(v.asString())
}

func (vm *VM) execPrintNewline() {
	vm.printStr("\n")
}

func (vm *VM) execPrintTab() {
	vm.printStr("\t")
}

func (vm *VM) execPrintSemicolon() {
	// No-op spacer — prevents newline in PRINT but doesn't emit output.
}

func (vm *VM) execInput(inst Instruction) {
	prompt := ""
	if inst.Operand >= 0 && int(inst.Operand) < len(vm.chunk.Constants) {
		prompt = vm.chunk.Constants[inst.Operand].Str
	}
	line := runtime.InputPrompt(prompt)
	vm.push(StringVal(line))
}
