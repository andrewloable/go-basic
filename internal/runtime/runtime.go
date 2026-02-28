/*
Package runtime is the runtime library for the go-basic transpiler.

# What Is a Runtime Library?

When a compiler translates source code from one language to another (or to
machine code), the generated output cannot be entirely self-contained. Programs
need to call built-in operations — printing text, computing a sine, opening a
file — that are too complex or platform-specific to inline at every call site.
A runtime library solves this by providing a pre-compiled set of functions that
the generated code can call by name.

This package plays exactly that role for go-basic:

  - The code generator (internal/codegen) emits Go source files that import this
    package using the alias "rt".
  - Instead of inlining an implementation of, say, LEFT$, the generator emits
    rt.Left(s, n) and relies on this package to do the work.
  - The compiler only needs to know each function's signature, not its body.

This clean separation is the same pattern used by mature compilers:

  - C — the C standard library (libc) provides printf, malloc, sin, etc.
  - Java — the JVM delegates to java.lang, java.io, and other standard packages.
  - Python — the interpreter calls into built-in C functions for len, print, etc.
  - Go itself — fmt, math, os, and strings are its own runtime support layer.

By separating "what the compiler knows about" from "what actually runs at
execution time," we gain several advantages:

 1. The compiler stays simple — it only emits call sites, not implementations.
 2. Behaviour can be fixed without recompiling user programs (just update this
    package and re-link/rebuild).
 3. Each function can be tested independently (see *_test.go files).

# Package Organisation

The runtime is split into focused files, mirroring the source-language domains:

  - math.go       — numeric built-ins (SIN, COS, LOG, RND, …)
  - strings.go    — string manipulation (LEFT$, MID$, INSTR, …)
  - io.go         — console I/O (PRINT USING, INPUT, LOCATE, …)
  - fileio.go     — file I/O (OPEN, CLOSE, GET, PUT, …)
  - graphics.go   — pixel graphics (SCREEN, PSET, LINE, CIRCLE, …)
  - system.go     — OS interaction (TIMER, SHELL, PEEK/POKE, …)
  - errors.go     — structured error handling (ON ERROR GOTO, ERR, ERL)
  - audio.go      — sound (PLAY, SOUND)
  - inkey.go      — keyboard polling (INKEY$)
*/
package runtime
