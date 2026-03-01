package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/loabletech/go-basic/internal/codegen"
	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
)

const shortUsage = `go-basic — Turbo BASIC / QBasic to Go Transpiler

Usage:
  go-basic transpile <file.bas>                Transpile to Go source (stdout)
  go-basic transpile <file.bas> -o <file.go>   Transpile and write to a file
  go-basic compile <file.bas>                  Compile to executable binary
  go-basic compile <file.bas> -o <output>      Compile with custom output name
  go-basic help                                Show detailed help

Description:
  Reads a Turbo BASIC or QBasic source file (.bas), parses it, performs
  semantic analysis, and generates equivalent Go source code.

  The generated Go code depends on the runtime library at:
    github.com/loabletech/go-basic/internal/runtime
`

const detailedHelp = `Supported BASIC Features:
  Statements   PRINT, INPUT, LET, IF/THEN/ELSE, FOR/NEXT, WHILE/WEND,
               DO/LOOP, SELECT CASE, GOTO, GOSUB/RETURN, DEF FN,
               SUB, FUNCTION, DIM, REDIM, CONST, TYPE, SWAP, END, STOP
  I/O          OPEN, CLOSE, PRINT#, INPUT#, LINE INPUT, WRITE#,
               GET, PUT, FIELD, LSET, RSET, LOC, LOF, EOF, SEEK
  Graphics     SCREEN, PSET, PRESET, LINE, CIRCLE, PAINT, DRAW,
               GET, PUT, PALETTE, COLOR, VIEW, WINDOW, CLS
  Audio        SOUND, PLAY
  Data         DATA, READ, RESTORE
  Types        INTEGER (%), LONG (&), SINGLE (!), DOUBLE (#), STRING ($)
  Operators    +, -, *, /, \, MOD, ^, AND, OR, XOR, NOT, EQV, IMP

Screen Modes:
  SCREEN 0     Text mode (80x25)
  SCREEN 1     CGA 320x200, 4 colors
  SCREEN 2     CGA 640x200, 2 colors
  SCREEN 7     EGA 320x200, 16 colors
  SCREEN 9     EGA 640x350, 16 colors
  SCREEN 12    VGA 640x480, 16 colors
  SCREEN 13    VGA 320x200, 256 colors

Examples:
  go-basic transpile hello.bas                 Print generated Go to stdout
  go-basic transpile hello.bas -o hello.go     Write generated Go to hello.go
  go-basic transpile hello.bas > hello.go      Redirect output to a file
  go-basic compile hello.bas                   Compile to ./hello binary
  go-basic compile hello.bas -o myapp          Compile with custom output name

  # Manual build workflow:
  mkdir out && cd out && go mod init myapp
  go-basic transpile ../hello.bas -o main.go
  go mod edit -require github.com/loabletech/go-basic@latest
  go build -o myapp . && ./myapp
`

func transpile(inputFile string) string {
	src, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "PARSE:", e)
		}
	}

	resolver := semantic.NewResolver(program)
	resolver.Resolve()
	gen := codegen.New()
	goSrc, genErr := gen.Generate(program, resolver.Table)
	if genErr != nil {
		fmt.Fprintln(os.Stderr, "CODEGEN:", genErr)
		os.Exit(1)
	}
	return goSrc
}

func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(shortUsage)
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "help", "--help", "-h":
		fmt.Print(shortUsage)
		fmt.Print(detailedHelp)
		os.Exit(0)

	case "transpile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "error: missing input file")
			fmt.Fprintln(os.Stderr, "usage: go-basic transpile <file.bas> [-o <file.go>]")
			os.Exit(1)
		}
		goSrc := transpile(os.Args[2])
		if out := flagValue(os.Args[3:], "-o"); out != "" {
			if err := os.WriteFile(out, []byte(goSrc), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "wrote %s\n", out)
		} else {
			fmt.Print(goSrc)
		}

	case "compile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "error: missing input file")
			fmt.Fprintln(os.Stderr, "usage: go-basic compile <file.bas> [-o <output>]")
			os.Exit(1)
		}
		inputFile := os.Args[2]
		goSrc := transpile(inputFile)

		// Determine output binary name.
		outputBin := flagValue(os.Args[3:], "-o")
		if outputBin == "" {
			base := filepath.Base(inputFile)
			outputBin = strings.TrimSuffix(base, filepath.Ext(base))
			if runtime.GOOS == "windows" {
				outputBin += ".exe"
			}
		}

		// Write main.go inside the project tree (temporarily) so it can
		// import internal packages without module boundary issues.
		modRoot, _ := os.Getwd()
		buildDir := filepath.Join(modRoot, ".go-basic-build")
		os.MkdirAll(buildDir, 0755)
		defer os.RemoveAll(buildDir)

		buildMain := filepath.Join(buildDir, "main.go")
		if err := os.WriteFile(buildMain, []byte(goSrc), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		absOut, _ := filepath.Abs(outputBin)
		buildCmd := exec.Command("go", "build", "-o", absOut, ".")
		buildCmd.Dir = buildDir
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: go build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "compiled %s -> %s\n", inputFile, outputBin)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		fmt.Print(shortUsage)
		os.Exit(1)
	}
}
