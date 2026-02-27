---
name: Turbo BASIC Compiler Expert
description: >
  This skill should be used when the user asks to "fix codegen", "add a BASIC statement",
  "implement a builtin function", "fix type errors", "fix compilation errors", "parse a
  BASIC construct", "add an AST node", "emit Go code", "fix the transpiler", "add runtime
  function", "implement a BASIC keyword", "fix gorilla.bas", "debug the compiler", or works
  on any task related to the go-basic Turbo BASIC to Go transpiler. It provides expert-level
  knowledge of the Turbo BASIC language, compiler architecture, Go codegen patterns, and the
  specific codebase conventions used in this project.
---

# Turbo BASIC Compiler Expert

Expert knowledge for developing the `go-basic` Turbo BASIC to Go transpiler. Combines deep
understanding of the Turbo BASIC language, classical compiler architecture, and idiomatic Go
code generation patterns.

## Compiler Pipeline

The compiler has four sequential phases plus a parallel VM backend:

```
Source (.bas)
    │  Phase 1: Lexer      internal/lexer/
    ▼
Token stream
    │  Phase 2: Parser     internal/parser/
    ▼
AST
    │  Phase 3: Semantic   internal/semantic/
    ▼
Annotated AST
    ├─ Phase 4a: Codegen   internal/codegen/   → Go source (.go)
    └─ Phase 4b: VM        internal/vm/        → Bytecode (stack-based VM)
```

The primary backend is the **Go transpiler** (codegen). The bytecode VM is secondary.

## Codebase Map

| Directory | Purpose | Key File(s) |
|-----------|---------|-------------|
| `cmd/basicc/` | CLI entry point | `main.go` — accepts `.bas`, runs pipeline |
| `internal/lexer/` | Tokenizer | `lexer.go` (scanner), `token.go` (token types + keywords) |
| `internal/parser/` | Recursive descent + Pratt parser | `parser.go` (~2500 lines) |
| `internal/ast/` | AST node definitions | `ast.go` — all Statement/Expression types |
| `internal/semantic/` | Type inference, symbol table | `analyzer.go` |
| `internal/codegen/` | Go transpiler | `codegen.go` (~1700 lines) |
| `internal/vm/` | Bytecode compiler + interpreter | `compiler.go`, `vm.go`, `opcodes.go` |
| `internal/runtime/` | Go runtime library | `math.go`, `strings.go`, `system.go`, `io.go`, `fileio.go`, `errors.go` |
| `examples/` | Generated `.go` output files | 65 files from `.bas` sources |
| `testdata/bas/` | Source `.bas` test files | 65 Turbo BASIC programs |

## Key Architectural Patterns

### Name Mangling (codegen)

BASIC type suffixes are illegal in Go identifiers. The `mangleName()` function translates:

| BASIC Suffix | Meaning | Go Suffix | Go Type |
|---|---|---|---|
| `%` | Integer (16-bit) | `_pct` | `int16` |
| `&` | Long (32-bit) | `_lng` | `int32` |
| `!` | Single float | `_sng` | `float32` |
| `#` | Double float | `_dbl` | `float64` |
| `$` | String | `_str` | `string` |
| (none) | Default single | (none) | `float32` |

Go reserved words get a `b_` prefix: `return` → `b_return`, `type` → `b_type`.

### Parser Architecture

The parser uses **recursive descent** for statements and **Pratt parsing (TDOP)** for
expressions. Key dispatch points:

- `parseStatement()` — keyword-dispatch switch on `curToken.Type` for all BASIC statements
- `parseExpression(precedence)` — Pratt loop: prefix handler, then infix loop while binding power holds
- `parseIdentifierExpression()` — critical ambiguity resolver: is `FOO(x)` a function call, array access, or builtin? Checks `isBuiltinFunction()`, `FN` prefix, and context.

### Codegen Structure

The `CodeGenerator` walks the AST and emits Go source via `emitStatement()` (type-switch on
Statement nodes) and `emitExpr()` (type-switch on Expression nodes). Output sections:

1. `package main` + imports
2. Unused-import suppressors (`var _ = fmt.Sprintf`)
3. `func main()` — all top-level BASIC statements
4. SUB/FUNCTION/DEF FN — emitted after main as Go functions

### Runtime Library Pattern

BASIC builtins that have no direct Go equivalent are implemented in `internal/runtime/` and
called with the `rt` import alias: `rt.Left(s, n)`, `rt.Sin(x)`, `rt.Timer()`.

## Development Workflow

### Adding a New BASIC Statement

1. **Lexer**: If new keyword needed, add token constant to `token.go` and keyword map
2. **AST**: Add Statement node struct to `ast.go` with `BasePos`, fields, and marker methods
3. **Parser**: Add case in `parseStatement()`, implement `parseXxxStatement()` function
4. **Codegen**: Add case in `emitStatement()`, implement `emitXxx()` function
5. **VM** (optional): Add case in `compileStatement()`
6. **Test**: Add parser test in `parser_test.go`, integration test in `internal/integration/`

### Adding a New Builtin Function

1. **Runtime**: Implement in appropriate `runtime/*.go` file (math.go, strings.go, etc.)
2. **Parser**: Ensure name is in `isBuiltinFunction()` list (parser.go)
3. **Codegen**: Add case in `emitFunctionCall()` switch — map BASIC name to `rt.Xxx()` call
4. **Handle special cases**:
   - `$`-suffixed builtins (CHR$, LEFT$, MID$): must be recognized BEFORE the ArrayAccess branch
   - No-argument builtins (INKEY$, DATE$, RND): handle in `emitIdentifier()` since parser creates Identifier nodes
   - Functions with error returns (CHR, SQR, LOG): wrap in inline closure or use `_` discard

### Fixing Type Errors in Generated Code

The most common codegen bug. Root cause: BASIC is loosely typed, Go is strict.

1. **Constant assignments**: Ensure `emitLet()` casts RHS to LHS variable type
2. **Binary operations**: Both operands must have the same Go type — cast the narrower one
3. **Runtime function args**: Math functions take `float64` — wrap args in `float64()`
4. **Runtime function returns**: Cast return values to target variable type
5. **DATA/READ**: The type-switch on `tv_` must cast to the target variable's Go type

Use `goTypeForIdent(name)` to determine a variable's Go type from its BASIC suffix.

## Critical Known Issues

These are documented in the bd task tracker. Consult before starting work:

- **$-suffixed builtins as arrays** (`v2j.1`): CHR$, LEFT$, MID$ etc. emitted as array accesses
- **No-arg builtins as variables** (`v2j.2`): INKEY$, DATE$, RND emitted as bare identifiers
- **SUBs as closures** (`rn3.1`): causes declared-and-not-used + goto-over-declaration
- **Type mismatches** (`27p.*`): float64/float32/int16 casting throughout codegen
- **INPUT as DATA/READ** (`6is.1`): INPUT should use stdin, not dataPool

## Testing

```bash
go test ./...                              # all tests (10 packages)
go build ./cmd/basicc/                     # build compiler
./basicc testdata/bas/hello.bas            # transpile single file
go build examples/hello.go                 # verify generated Go compiles
```

Regression tests in `parser_test.go` cover 24 specific parser bugs (TestRegression* functions).

## Additional Resources

### Reference Files

For detailed language and codegen patterns, consult:
- **`references/compiler-architecture.md`** — Detailed phase-by-phase architecture, AST node catalog, parser internals
- **`references/turbo-basic-reference.md`** — Complete Turbo BASIC language reference: types, statements, functions, graphics, file I/O
- **`references/codegen-patterns.md`** — Go emission patterns for every BASIC construct, known pitfalls, type mapping rules
