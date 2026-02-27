# Turbo BASIC Compiler (Go)

A Turbo BASIC–inspired compiler written in Go, designed to bring classic BASIC syntax into a modern, portable toolchain. This project aims to balance retro compatibility with clean compiler architecture and extensibility.

---

## 🚀 Overview

This compiler parses classic BASIC constructs such as:

* `PRINT`
* `LET`
* `IF / THEN`
* `FOR / NEXT`
* `GOTO`
* Line numbers
* Numeric and string variables

It transforms BASIC source code into one of the following targets:

* Native binaries (planned)
* Go source code (transpilation mode)
* Custom bytecode executed by a built-in virtual machine

The project is structured for clarity, maintainability, and educational value.

---

## 🧠 Goals

* Recreate the spirit of Turbo BASIC
* Maintain optional legacy compatibility (line numbers, GOTO)
* Provide a clean and modern compiler architecture
* Keep the codebase hackable and extensible
* Serve as a learning resource for compiler design

---

## 🏗 Architecture

The compiler follows a traditional multi-stage pipeline:

```
Source Code
   ↓
Lexer (Tokenizer)
   ↓
Parser
   ↓
Abstract Syntax Tree (AST)
   ↓
Code Generator
   ↓
Target Output (Go / Bytecode / Native)
```

### Project Structure

```
/cmd/basicc        → CLI entry point
/internal/lexer    → Tokenizer
/internal/parser   → Syntax parser
/internal/ast      → AST definitions
/internal/codegen  → Code generation
/internal/vm       → Bytecode virtual machine (optional)
```

---

## ✨ Example

### BASIC Input

```basic
10 PRINT "HELLO WORLD"
20 END
```

### Possible Output (Go transpile mode)

```go
package main
import "fmt"

func main() {
    fmt.Println("HELLO WORLD")
}
```

---

## 🔧 Features (Planned / In Progress)

* Line number support
* Expression parsing with operator precedence
* Variables (numeric and string)
* Control flow (`IF`, `FOR`, `GOTO`)
* Error reporting with line references
* Bytecode VM backend
* Cross-platform CLI binary

---

## 🎯 Design Philosophy

* Simple first, powerful later
* Clear separation of compiler stages
* No unnecessary magic
* Retro feel, modern implementation

---

## 📚 Educational Value

This project is ideal for:

* Learning compiler fundamentals
* Understanding parsing techniques
* Exploring AST design
* Experimenting with language extensions
* Studying VM implementation

---

## 🔮 Future Possibilities

* Native x86_64 code generation
* WebAssembly backend
* DOS-style runtime mode
* Graphics and sound extensions
* Full Turbo BASIC compatibility mode

---

## 📄 License

MIT (or your preferred open-source license)

---

Built with Go. Inspired by classic Turbo BASIC. Designed for modern systems.
