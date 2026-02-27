# Turbo BASIC Compiler in Go

A love letter to classic BASIC, rebuilt from scratch in Go.

This project is a working compiler that takes Turbo BASIC and QBasic source code and translates it into modern, compilable Go programs. It started as a curiosity — could we take those old .BAS files from the DOS era and bring them back to life on today's machines? Turns out, yes.

---

## What It Does

Feed it a `.bas` file. It reads through the BASIC source, breaks it into tokens, builds a syntax tree, and writes out a standalone `.go` file that does the same thing the original program did. The generated Go code uses a small runtime library for BASIC-specific features like `LEFT$()`, `SIN()`, and `PRINT USING`.

It handles the real stuff — not just toy programs:

- Control flow: `IF/THEN/ELSE`, `FOR/NEXT`, `WHILE/WEND`, `DO/LOOP`, `SELECT CASE`
- Procedures: `SUB`, `FUNCTION`, `DEF FN`, `GOSUB/RETURN`
- Data: `DIM` arrays, `DATA/READ`, user-defined `TYPE` structs
- I/O: `PRINT`, `INPUT`, `OPEN/CLOSE`, sequential and random file access
- String and math builtins: all the classics (`MID$`, `CHR$`, `SQR`, `RND`, etc.)
- Turbo BASIC extensions: `DO/LOOP`, `EXIT`, `SELECT CASE`, `INCR/DECR`, metacommands

There's also a bytecode VM backend that can interpret programs directly, though the Go transpiler is the primary output path.

---

## How It's Built

The compiler follows the classical pipeline that every compiler textbook describes, but implemented cleanly in Go without external dependencies:

```
Source (.bas)
    |  Lexer        Break source into tokens
    v
Token stream
    |  Parser       Build an abstract syntax tree
    v
AST
    |  Semantic     Infer types, build symbol table
    v
Annotated AST
    |  Codegen      Walk the tree, emit Go source
    v
Output (.go)
```

The parser uses recursive descent for statements and Pratt parsing (top-down operator precedence) for expressions — the same combination used in production compilers. Every BASIC keyword has its own token type so the parser can dispatch in O(1).

The code generator is a tree-walking transpiler. It visits each AST node and writes the equivalent Go code into a buffer. BASIC constructs that have no direct Go equivalent (like `GOSUB`, `DATA/READ`, or graphics commands) are handled by the runtime library or emitted as stubs.

---

## Project Layout

```
cmd/basicc/           Command-line tool
internal/
  lexer/              Scanner — characters to tokens
  parser/             Parser — tokens to syntax tree
  ast/                Syntax tree node definitions
  semantic/           Type analysis
  codegen/            Go code generator
  vm/                 Bytecode compiler and interpreter
  runtime/            Runtime library (math, strings, I/O, files)
  integration/        End-to-end tests
testdata/bas/         65 BASIC source programs (test suite)
examples/             Generated Go output
```

---

## Running It

```bash
# Build the compiler
go build ./cmd/basicc/

# Transpile a BASIC program to Go
./basicc transpile testdata/bas/hello.bas

# Compile and run the output
go run examples/hello.go
```

---

## Current State

The compiler successfully parses all 65 test programs, including the classic QBasic Gorillas game. 15 of the generated Go files compile and run cleanly. The remaining 50 have codegen issues being tracked and fixed — mostly around strict Go type casting, builtin function emission, and procedure handling.

The parser is solid with 24 regression tests covering every bug found during development. All 10 test packages pass.

---

## Why This Exists

Three reasons:

**Nostalgia.** Turbo BASIC was where a lot of us started. Those programs deserve to run again, not just sit in archives.

**Education.** Every stage of this compiler is written to be readable. The code includes comments explaining *how compilers work* — what Pratt parsing does, why ASTs matter, how code generation walks a tree. If you want to learn compiler construction by reading real code, this is a good place to start.

**Fun.** There's something deeply satisfying about typing `PRINT "HELLO WORLD"` and watching it come out the other side as working Go code.

---

## What's Next

Active development is focused on:

- Fixing numeric type casting across the codegen (the biggest remaining issue)
- Getting all 14 string builtins (`CHR$`, `LEFT$`, `MID$`, etc.) emitting correctly
- Proper `SUB`/`FUNCTION` emission as Go functions
- `INPUT` statement reading from stdin instead of DATA pools
- File I/O wiring (the runtime is ready, the codegen needs to call it)
- Graphics stubs for `SCREEN`, `CIRCLE`, `LINE`, `PSET`, `PAINT`
- Multi-dimensional array support

The goal: get all 65 test programs compiling and running.

---

## License

MIT

---

Built with Go. Inspired by Turbo BASIC. Made for anyone who remembers `10 PRINT "HELLO"`.
