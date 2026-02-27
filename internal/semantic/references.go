package semantic

import (
	"fmt"
	"strconv"

	"github.com/loabletech/go-basic/internal/ast"
)

// ReferenceResolver resolves GOTO/GOSUB targets, labels, line numbers, and DATA pools.
type ReferenceResolver struct {
	Program    *ast.Program
	LineMap    map[int]int    // line number -> index in Program.Statements
	LabelMap   map[string]int // label name (upper) -> index in Program.Statements
	DataPool   []ast.Expression
	Errors     []string
}

// NewReferenceResolver creates a new ReferenceResolver.
func NewReferenceResolver(prog *ast.Program) *ReferenceResolver {
	return &ReferenceResolver{
		Program:  prog,
		LineMap:  make(map[int]int),
		LabelMap: make(map[string]int),
	}
}

// Resolve performs all reference resolution passes.
func (rr *ReferenceResolver) Resolve() []string {
	// Pass 1: collect all labels, line numbers, and DATA statements
	rr.collectTargets(rr.Program.Statements)

	// Pass 2: validate all GOTO/GOSUB/ON ERROR/RESTORE references
	rr.validateReferences(rr.Program.Statements)

	return rr.Errors
}

func (rr *ReferenceResolver) errorf(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	rr.Errors = append(rr.Errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
}

// collectTargets walks statements to find labels, line numbers, and DATA.
func (rr *ReferenceResolver) collectTargets(stmts []ast.Statement) {
	for i, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.LabelStatement:
			upper := toUpper(s.Name)
			if _, exists := rr.LabelMap[upper]; exists {
				rr.errorf(s.BasePos, "duplicate label: %s", s.Name)
			} else {
				rr.LabelMap[upper] = i
			}

		case *ast.LineNumberStatement:
			if _, exists := rr.LineMap[s.Number]; exists {
				rr.errorf(s.BasePos, "duplicate line number: %d", s.Number)
			} else {
				rr.LineMap[s.Number] = i
			}

		case *ast.DataStatement:
			rr.DataPool = append(rr.DataPool, s.Values...)

		case *ast.SubDeclaration:
			rr.collectTargets(s.Body)

		case *ast.FunctionDeclaration:
			rr.collectTargets(s.Body)

		case *ast.IfStatement:
			rr.collectTargets(s.ThenBlock)
			for _, elif := range s.ElseIfClauses {
				rr.collectTargets(elif.Body)
			}
			rr.collectTargets(s.ElseBlock)

		case *ast.ForStatement:
			rr.collectTargets(s.Body)

		case *ast.WhileStatement:
			rr.collectTargets(s.Body)

		case *ast.DoLoopStatement:
			rr.collectTargets(s.Body)

		case *ast.SelectCaseStatement:
			for _, c := range s.Cases {
				rr.collectTargets(c.Body)
			}
			rr.collectTargets(s.ElseBlock)
		}
	}
}

// validateReferences checks that all jump targets exist.
func (rr *ReferenceResolver) validateReferences(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.GotoStatement:
			rr.validateTarget(s.Target, s.BasePos, "GOTO")

		case *ast.GosubStatement:
			rr.validateTarget(s.Target, s.BasePos, "GOSUB")

		case *ast.OnErrorGotoStatement:
			if s.Target != "0" && s.Target != "" {
				rr.validateTarget(s.Target, s.BasePos, "ON ERROR GOTO")
			}

		case *ast.OnEventGosubStatement:
			rr.validateTarget(s.Target, s.BasePos, "ON "+s.EventType+" GOSUB")

		case *ast.RestoreStatement:
			if s.Target != "" {
				rr.validateTarget(s.Target, s.BasePos, "RESTORE")
			}

		case *ast.ResumeStatement:
			if s.Type != "" && s.Type != "NEXT" {
				rr.validateTarget(s.Type, s.BasePos, "RESUME")
			}

		case *ast.SubDeclaration:
			rr.validateReferences(s.Body)

		case *ast.FunctionDeclaration:
			rr.validateReferences(s.Body)

		case *ast.IfStatement:
			rr.validateReferences(s.ThenBlock)
			for _, elif := range s.ElseIfClauses {
				rr.validateReferences(elif.Body)
			}
			rr.validateReferences(s.ElseBlock)

		case *ast.ForStatement:
			rr.validateReferences(s.Body)

		case *ast.WhileStatement:
			rr.validateReferences(s.Body)

		case *ast.DoLoopStatement:
			rr.validateReferences(s.Body)

		case *ast.SelectCaseStatement:
			for _, c := range s.Cases {
				rr.validateReferences(c.Body)
			}
			rr.validateReferences(s.ElseBlock)
		}
	}
}

// validateTarget checks whether a target (label name or line number string) is valid.
func (rr *ReferenceResolver) validateTarget(target string, pos ast.Position, context string) {
	// Try as line number first
	if num, err := strconv.Atoi(target); err == nil {
		if _, exists := rr.LineMap[num]; !exists {
			rr.errorf(pos, "%s target: undefined line number %d", context, num)
		}
		return
	}

	// Try as label
	upper := toUpper(target)
	if _, exists := rr.LabelMap[upper]; !exists {
		rr.errorf(pos, "%s target: undefined label '%s'", context, target)
	}
}

// GetDataPool returns the collected DATA pool for use by the code generator.
func (rr *ReferenceResolver) GetDataPool() []ast.Expression {
	return rr.DataPool
}

// GetLineMap returns the line number -> statement index mapping.
func (rr *ReferenceResolver) GetLineMap() map[int]int {
	return rr.LineMap
}

// GetLabelMap returns the label -> statement index mapping.
func (rr *ReferenceResolver) GetLabelMap() map[string]int {
	return rr.LabelMap
}

// toUpper is a helper for case-insensitive comparison.
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
