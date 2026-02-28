// Package vm implements the bytecode virtual machine for the Turbo BASIC compiler.
// The VM source is organized across several focused files:
//   - vm_types.go     — Value types, Instruction, Chunk, CallFrame, arrayValue
//   - vm_core.go      — VM struct, NewVM, Run() dispatch loop
//   - vm_arithmetic.go — Arithmetic, comparison, and logical opcode handlers
//   - vm_control.go   — Control flow, variable, array, data, and type-conversion handlers
//   - vm_io.go        — I/O opcode handlers
//   - vm_builtins.go  — Built-in function dispatch and implementations
package vm
