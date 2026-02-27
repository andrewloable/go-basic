# go-basic — Turbo BASIC to Go Transpiler

## Project Overview

A compiler that translates Turbo BASIC / QBasic source code into compilable Go programs.
The pipeline: Lexer → Parser → Semantic Analysis → Code Generation (Go transpiler) + Bytecode VM.

## Codebase Structure

```
cmd/basicc/          CLI entry point (accepts .bas files)
internal/
  lexer/             Tokenizer: source text → token stream
  parser/            Recursive descent + Pratt parser: tokens → AST
  ast/               AST node definitions (Statement + Expression types)
  semantic/          Type inference, symbol table
  codegen/           Go transpiler: AST → .go source files
  vm/                Bytecode compiler + stack-based VM interpreter
  runtime/           Go runtime library (math, strings, I/O, files, errors)
  integration/       Integration tests
  testutil/          Shared test utilities
  benchmark/         Performance benchmarks
testdata/bas/        65 Turbo BASIC source programs
examples/            Generated Go output files
docs/                Project documentation
```

## Key Conventions

- **Name mangling**: BASIC `%→_pct`, `&→_lng`, `!→_sng`, `#→_dbl`, `$→_str`. Go reserved words get `b_` prefix.
- **Runtime prefix**: `rt` import alias for `internal/runtime` calls (e.g., `rt.Sin()`, `rt.Left()`).
- **Type mapping**: `%→int16`, `&→int32`, `!→float32`, `#→float64`, `$→string`, default `float32`.
- **Test**: `go test ./...` runs all 10 test packages. Parser has 24 regression tests (TestRegression*).
- **Build**: `go build ./cmd/basicc/` then `./basicc transpile testdata/bas/hello.bas`.
- **Task tracker**: `bd` (beads) for issue tracking. `bd list --status open` for current tasks.

## Expert Plugin

For deep compiler/language knowledge, load the plugin:
```
claude --plugin-dir .claude/plugins/go-basic-expert
```

This provides the "Turbo BASIC Compiler Expert" skill with detailed references on:
- Compiler architecture (lexer, parser, AST, codegen, VM internals)
- Turbo BASIC language (all types, statements, functions, graphics, file I/O)
- Go codegen patterns (type casting, name mangling, emission patterns, known pitfalls)
