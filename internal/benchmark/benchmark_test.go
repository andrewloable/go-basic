package benchmark

import (
	"fmt"
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/codegen"
	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
)

// generateProgram creates a synthetic BASIC program of the given size.
func generateProgram(lines int) string {
	var sb strings.Builder
	sb.WriteString("DIM a(1000)\n")
	for i := 0; i < lines; i++ {
		switch i % 5 {
		case 0:
			fmt.Fprintf(&sb, "a(%d) = %d * %d + %d\n", i%1000, i, i+1, i+2)
		case 1:
			fmt.Fprintf(&sb, "IF a(%d) > %d THEN a(%d) = a(%d) - %d\n", i%1000, i, i%1000, i%1000, i)
		case 2:
			fmt.Fprintf(&sb, "PRINT a(%d);\n", i%1000)
		case 3:
			fmt.Fprintf(&sb, "x = SQR(ABS(a(%d)))\n", i%1000)
		case 4:
			fmt.Fprintf(&sb, "a(%d) = a(%d) + 1\n", i%1000, i%1000)
		}
	}
	sb.WriteString("END\n")
	return sb.String()
}

var sizes = []int{100, 500, 1000}

// BenchmarkLexer measures lexer throughput.
func BenchmarkLexer(b *testing.B) {
	for _, size := range sizes {
		prog := generateProgram(size)
		b.Run(fmt.Sprintf("lines_%d", size), func(b *testing.B) {
			b.SetBytes(int64(len(prog)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := lexer.New(prog)
				l.AllTokens()
			}
		})
	}
}

// BenchmarkParser measures parser throughput.
func BenchmarkParser(b *testing.B) {
	for _, size := range sizes {
		prog := generateProgram(size)
		b.Run(fmt.Sprintf("lines_%d", size), func(b *testing.B) {
			b.SetBytes(int64(len(prog)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := lexer.New(prog)
				p := parser.New(l)
				p.ParseProgram()
			}
		})
	}
}

// BenchmarkSemantic measures semantic analysis throughput.
func BenchmarkSemantic(b *testing.B) {
	for _, size := range sizes {
		prog := generateProgram(size)
		l := lexer.New(prog)
		p := parser.New(l)
		ast := p.ParseProgram()
		b.Run(fmt.Sprintf("lines_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := semantic.NewResolver(ast)
				r.Resolve()
			}
		})
	}
}

// BenchmarkCodegen measures code generation throughput.
func BenchmarkCodegen(b *testing.B) {
	for _, size := range sizes {
		prog := generateProgram(size)
		l := lexer.New(prog)
		p := parser.New(l)
		ast := p.ParseProgram()
		r := semantic.NewResolver(ast)
		table, _ := r.Resolve()
		b.Run(fmt.Sprintf("lines_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				gen := codegen.New()
				gen.Generate(ast, table)
			}
		})
	}
}

// BenchmarkFullPipeline measures the complete compilation pipeline.
func BenchmarkFullPipeline(b *testing.B) {
	for _, size := range sizes {
		prog := generateProgram(size)
		b.Run(fmt.Sprintf("lines_%d", size), func(b *testing.B) {
			b.SetBytes(int64(len(prog)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := lexer.New(prog)
				p := parser.New(l)
				ast := p.ParseProgram()
				r := semantic.NewResolver(ast)
				table, _ := r.Resolve()
				gen := codegen.New()
				gen.Generate(ast, table)
			}
		})
	}
}
