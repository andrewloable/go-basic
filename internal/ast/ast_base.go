// Package ast defines the Abstract Syntax Tree (AST) node types produced by
// the parser (Phase 2) and consumed by all later phases.
//
// # What is an Abstract Syntax Tree?
//
// An AST is a tree representation of the grammatical structure of source code.
// Unlike a Parse Tree (which mirrors every token and grammar rule), the AST is
// "abstract" — it omits syntactic noise (parentheses, commas, keywords that
// only provide structure) and keeps only the semantically meaningful pieces.
//
//	BASIC source:   FOR i = 1 TO 10 STEP 2
//	AST node:       ForStatement{ Var: "i", Start: 1, End: 10, Step: 2, Body: [...] }
//
// # Node hierarchy
//
// Every AST node implements the Node interface. Nodes fall into two categories:
//
//	Statement nodes  (statementNode marker method) — represent actions:
//	  LET, PRINT, FOR…NEXT, IF…THEN, GOSUB, SUB declarations, …
//
//	Expression nodes (expressionNode marker method) — represent values:
//	  NumberLiteral, StringLiteral, Identifier, BinaryExpr, FunctionCall, …
//
// Programs are sequences of statements; statements contain expressions.
//
// # Visitor pattern
//
// Later compiler phases (semantic analysis, code generation, VM compilation)
// walk the AST using Go type-switches. Each phase matches on concrete node
// types and implements the transformation it needs. This "visitor without
// visitors" pattern is idiomatic Go and avoids the boilerplate of explicit
// Visitor interfaces.
package ast

// ---------------------------------------------------------------------------
// Position
// ---------------------------------------------------------------------------

// Position represents a location in the source code.
//
// Every AST node carries a Position so that later compiler phases (semantic
// analysis, code generation) can emit precise error messages like
// "5:12: type mismatch — expected INTEGER, got STRING". Without position
// tracking you can only say "there is an error somewhere", which is nearly
// useless when debugging a large program.
//
// Position is stamped onto a node at parse time from the lexer token that
// triggered the node's creation. Once the AST is built, positions are
// read-only — the compiler never needs to update them.
type Position struct {
	Line   int // 1-based source line (the first line is 1, not 0)
	Column int // 1-based byte column within the line
}

// ---------------------------------------------------------------------------
// NumType constants
// ---------------------------------------------------------------------------

// NumType constants tag a numeric literal with its Turbo BASIC precision class.
// These mirror the four numeric types BASIC supports via type-suffix characters:
//
//	NumInt    — 16-bit signed integer (suffix %)
//	NumLong   — 32-bit signed integer (suffix &)
//	NumSingle — 32-bit IEEE-754 float (suffix ! or no suffix)
//	NumDouble — 64-bit IEEE-754 float (suffix #)
//
// The code generator uses these tags to emit the correct Go type cast so that
// the generated program preserves the numeric precision of the original BASIC.
// For example, NumInt → int16 cast, NumDouble → float64 cast.
const (
	NumInt = iota
	NumLong
	NumSingle
	NumDouble
)

// ---------------------------------------------------------------------------
// Core interfaces
// ---------------------------------------------------------------------------

// Node is the interface implemented by every AST node.
//
// Two methods are required of all nodes:
//
//   - TokenLiteral() — returns the source text of the primary token that
//     created the node (e.g. "FOR", "+", "42"). This is mainly for debugging:
//     pretty-printers and test assertions use it to reconstruct readable output.
//
//   - Pos() — returns the source position (line, column) for error reporting.
//     Compiler phases that detect problems call node.Pos() to include the
//     location in the diagnostic message.
//
// Concrete node types embed a BasePos field (of type Position) and implement
// Pos() by returning that field. This is the standard Go idiom for sharing
// interface method implementations via embedding.
type Node interface {
	TokenLiteral() string
	Pos() Position
}

// Expression is a node that produces a value.
//
// Examples: 42, "hello", a + b, SIN(x), arr(i).
//
// The expressionNode() method is a compile-time marker — it exists only so the
// Go type-checker can verify that a value placed where an Expression is expected
// actually implements the interface. The method body is always empty. This
// technique (sometimes called an "interface marker" or "tag method") prevents
// accidentally passing a Statement where an Expression is needed, which would
// be a semantic mistake the compiler could not catch without the marker.
type Expression interface {
	Node
	expressionNode() // marker — enforces type-safe use of expression nodes
}

// Statement is a node that performs an action.
//
// Examples: LET x = 5, PRINT "hello", FOR i = 1 TO 10, IF cond THEN ...
//
// Like expressionNode(), statementNode() is a marker method used purely for
// compile-time type safety. It ensures that a Statement cannot silently be used
// where an Expression is expected (and vice-versa), because the two interfaces
// have disjoint marker methods.
type Statement interface {
	Node
	statementNode() // marker — enforces type-safe use of statement nodes
}

// ---------------------------------------------------------------------------
// Program (root node)
// ---------------------------------------------------------------------------

// Program is the root AST node representing an entire source file.
//
// In a tree data structure there must be exactly one root — the node that
// contains everything else. Program plays that role. The parser creates one
// Program node and appends every top-level statement to its Statements slice.
//
// All later compiler phases begin their traversal here:
//
//	for _, stmt := range program.Statements {
//	    // visit or emit code for each top-level statement
//	}
//
// Because BASIC programs are flat (no nested modules or namespaces), a simple
// slice of statements is sufficient; no tree of scopes is required at this level.
type Program struct {
	BasePos    Position    // position of the very first token in the file
	Statements []Statement // all top-level statements in order
}

func (n *Program) statementNode()      {}
func (n *Program) TokenLiteral() string { return "PROGRAM" }
func (n *Program) Pos() Position        { return n.BasePos }
