# go-basic

A Turbo BASIC / QBasic to Go transpiler. Compiles classic `.bas` files into standalone Go programs.

```basic
10 PRINT "HELLO WORLD"
20 FOR I = 1 TO 5
30   PRINT "COUNT:"; I
40 NEXT I
50 END
```

becomes:

```go
package main

import "fmt"

func main() {
    fmt.Println("HELLO WORLD")
    var I float32
    for I = float32(1); I <= float32(5); I += float32(1) {
        fmt.Print("COUNT:", I)
        fmt.Println()
    }
}
```

## Features

- Full Turbo BASIC / QBasic parser — handles real-world programs including the classic QBasic Gorillas game
- Go source code output — generates standalone, compilable `.go` files
- Bytecode VM — alternative backend that interprets programs directly
- Runtime library — Go implementations of BASIC builtins (`LEFT$`, `MID$`, `SIN`, `PRINT USING`, etc.)
- Educational codebase — extensively commented to explain how compilers work

### Language Support

| Category | Supported |
|----------|-----------|
| Control flow | `IF/THEN/ELSE/ELSEIF`, `FOR/NEXT`, `WHILE/WEND`, `DO/LOOP`, `SELECT CASE`, `GOTO`, `GOSUB/RETURN` |
| Procedures | `SUB`, `FUNCTION`, `DEF FN`, `DECLARE`, `CALL` |
| Variables | Integer (`%`), Long (`&`), Single (`!`), Double (`#`), String (`$`) |
| Arrays | `DIM`, `REDIM`, `ERASE`, multi-dimensional |
| Data | `DATA`, `READ`, `RESTORE` |
| I/O | `PRINT`, `INPUT`, `LINE INPUT`, `PRINT USING`, `WRITE` |
| File I/O | `OPEN`, `CLOSE`, `PRINT#`, `INPUT#`, `WRITE#`, sequential/random/binary |
| String functions | `LEFT$`, `RIGHT$`, `MID$`, `CHR$`, `ASC`, `STR$`, `VAL`, `INSTR`, `UCASE$`, `LCASE$`, `LTRIM$`, `RTRIM$`, `HEX$`, `OCT$`, `BIN$`, `STRING$`, `SPACE$` |
| Math functions | `ABS`, `SGN`, `INT`, `FIX`, `SQR`, `SIN`, `COS`, `TAN`, `ATN`, `EXP`, `LOG`, `RND`, `CINT`, `CLNG`, `CSNG`, `CDBL` |
| User types | `TYPE...END TYPE` with field access |
| Graphics | `SCREEN`, `CIRCLE`, `LINE`, `PSET`, `PAINT`, `DRAW`, `VIEW`, `WINDOW` (stubs) |
| Sound | `SOUND`, `PLAY` (stubs) |
| Turbo BASIC extensions | `DO/LOOP`, `EXIT`, `SELECT CASE`, `INCR/DECR`, `$DYNAMIC`, `$IF/$ENDIF` |

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)

### Build

```bash
git clone https://github.com/loabletech/go-basic.git
cd go-basic
go build ./cmd/basicc/
```

### Transpile a BASIC Program

```bash
# Transpile to Go
./basicc transpile path/to/program.bas

# The output is written to examples/program.go
# Compile and run it
go run examples/program.go
```

### Run Tests

```bash
go test ./...
```

## How It Works

The compiler follows a classical multi-stage pipeline:

```
Source (.bas)  →  Lexer  →  Parser  →  Semantic Analysis  →  Code Generator  →  Output (.go)
```

**Lexer** (`internal/lexer/`) — Hand-written scanner. Converts source text into a stream of typed tokens. Handles BASIC keywords, type suffixes (`%`, `&`, `!`, `#`, `$`), string literals, numbers (decimal, hex `&H`, octal `&O`), and line structure.

**Parser** (`internal/parser/`) — Recursive descent for statements, Pratt parsing (TDOP) for expressions. Resolves the ambiguity between function calls, array accesses, and variable references. Produces a typed AST.

**AST** (`internal/ast/`) — Abstract syntax tree with Statement and Expression node types. Later phases walk the tree using Go type-switches (visitor pattern without the boilerplate).

**Semantic Analysis** (`internal/semantic/`) — Builds a symbol table with inferred types from variable suffixes, `DIM` declarations, and assignment context.

**Code Generator** (`internal/codegen/`) — Tree-walking transpiler that emits Go source code. Maps BASIC constructs to Go equivalents. Uses a runtime library for BASIC-specific features.

**Runtime** (`internal/runtime/`) — Pure Go implementations of BASIC builtins: math functions, string operations, file I/O, formatted output, and system functions.

**Bytecode VM** (`internal/vm/`) — Alternative backend. Compiles AST to stack-based bytecode and interprets it directly.

## Project Structure

```
cmd/basicc/           CLI tool
internal/
  lexer/              Tokenizer
  parser/             Recursive descent + Pratt parser
  ast/                AST node definitions
  semantic/           Type inference and symbol table
  codegen/            Go code generator (transpiler)
  vm/                 Bytecode compiler and interpreter
  runtime/            BASIC runtime library
    math.go           ABS, SIN, SQR, RND, etc.
    strings.go        LEFT$, MID$, CHR$, STR$, etc.
    io.go             PRINT helpers, formatting
    fileio.go         File I/O (OPEN, CLOSE, PRINT#, etc.)
    system.go         TIMER, DATE$, ENVIRON$, etc.
    errors.go         ON ERROR / RESUME support
  integration/        End-to-end tests
  benchmark/          Performance benchmarks
testdata/bas/         65 BASIC test programs
examples/             Generated Go output
docs/                 Documentation
```

## Current Status

| Metric | Count |
|--------|-------|
| Test programs | 65 (all parse successfully) |
| Generated Go files that compile | 15 / 65 |
| Test packages passing | 10 / 10 |
| Parser regression tests | 24 |

Active work is focused on fixing the code generator to handle strict Go typing, builtin function emission, and procedure translation. See the [project docs](docs/project.md) for details.

## Examples

The `testdata/bas/` directory contains 65 BASIC programs covering everything from hello world to the QBasic Gorillas game. Each has a corresponding generated `.go` file in `examples/`.

```bash
# Try these
./basicc transpile testdata/bas/hello.bas
./basicc transpile testdata/bas/fibonacci.bas
./basicc transpile testdata/bas/gorilla.bas
```

## Contributing

Contributions are welcome. The biggest areas where help is needed:

1. **Codegen type casting** — Making the Go output handle BASIC's loose typing correctly
2. **Builtin functions** — Getting all string/math builtins to emit proper `rt.*` calls
3. **File I/O** — Wiring the existing runtime FileManager into the code generator
4. **Graphics backend** — Replacing stubs with a real graphics library (ebiten, SDL2, etc.)

The codebase is extensively commented to explain compiler concepts. Start with `internal/lexer/token.go` and follow the pipeline.

## Learning Compiler Design

This project was built to be readable. Every major file includes comments explaining the compiler concepts it implements:

- **Lexical analysis** — `internal/lexer/lexer.go` explains scanning, tokens, and keywords
- **Parsing** — `internal/parser/parser.go` covers recursive descent, Pratt parsing, and precedence
- **AST design** — `internal/ast/ast.go` explains node hierarchies and the visitor pattern
- **Code generation** — `internal/codegen/codegen.go` walks through tree-walking transpilation
- **Virtual machines** — `internal/vm/` demonstrates bytecode compilation and stack-based execution

## License

MIT

## Acknowledgments

Inspired by Borland Turbo BASIC (1987) and Microsoft QBasic (1991). Built for anyone who remembers the joy of `RUN`.
