// This file is intentionally left with only a package declaration.
// All AST node types have been distributed across focused files:
//
//   ast_base.go         - Position, NumType constants, Node/Expression/Statement interfaces, Program
//   expressions.go      - Expression node types
//   stmt_control.go     - Control flow statement types
//   stmt_declarations.go - Declaration and assignment statement types
//   stmt_io.go          - I/O statement types
//   stmt_graphics.go    - Graphics, sound, display, and miscellaneous statement types
package ast
