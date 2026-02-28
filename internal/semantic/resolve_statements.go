package semantic

import (
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Statement-specific resolvers
//
// TUTORIAL — Statement resolution vs. expression resolution
//
// Like the code generator, the resolver distinguishes between statements
// (which produce side effects on the symbol table) and expressions (which
// are resolved for the names they reference).
//
// Statement resolvers:
//   - May define new symbols (resolveLet defines a variable on first use)
//   - May open/close scopes (resolveSubDecl calls EnterScope/ExitScope)
//   - Recursively resolve their sub-expressions and sub-statements
//
// Expression resolvers:
//   - Only mark symbols as Used (they never define new ones, except for
//     the implicit-declaration rule in resolveIdentifier)
//   - Recursively resolve nested sub-expressions
//
// This clear separation keeps the logic manageable. Adding a new statement
// type means adding a case to resolveStatement() and a new resolveXxx()
// method that follows the same pattern as the existing ones.
// ---------------------------------------------------------------------------

func (r *Resolver) resolveLet(s *ast.LetStatement) {
	// Resolve the right-hand side first.
	r.resolveExpression(s.Value)

	// Implicitly declare the variable if it is not yet known.
	// TUTORIAL: BASIC uses implicit variable declaration — you can use a variable
	// without a prior DIM or explicit declaration. The resolver handles this by
	// checking if the variable is already in the symbol table; if not, it defines
	// it now. This mirrors BASIC's runtime behavior where the first assignment
	// to a name brings the variable into existence.
	name := s.Name.Name
	suffix := s.Name.TypeSuffix
	fullName := name + suffix
	dt := r.Table.ResolveType(fullName)

	if sym := r.Table.Lookup(fullName); sym != nil {
		sym.Defined = true
		sym.Used = true
	} else {
		_ = r.Table.Define(fullName, &Symbol{
			Name:     fullName,
			Type:     SymVariable,
			DataType: dt,
			Scope:    r.Table.CurrentScope.Name,
			Defined:  true,
			Line:     s.Pos().Line,
			Column:   s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveDim(s *ast.DimStatement) {
	for _, d := range s.Declarations {
		name := d.Name + d.TypeSuffix
		dims := len(d.Dimensions)
		dt := r.resolveElementType(d.ElementType, d.TypeSuffix, d.Name)

		symType := SymVariable
		if dims > 0 {
			symType = SymArray
		}

		if existing := r.Table.LookupLocal(name); existing != nil {
			r.addError(d.BasePos.Line, d.BasePos.Column,
				"duplicate DIM of %q in scope %s", name, r.Table.CurrentScope.Name)
			continue
		}

		_ = r.Table.Define(name, &Symbol{
			Name:      name,
			Type:      symType,
			DataType:  dt,
			Scope:     r.Table.CurrentScope.Name,
			Defined:   true,
			Line:      d.BasePos.Line,
			Column:    d.BasePos.Column,
			ArrayDims: dims,
		})

		// Resolve dimension expressions.
		for _, dr := range d.Dimensions {
			r.resolveExpressionIfNotNil(dr.Lower)
			r.resolveExpressionIfNotNil(dr.Upper)
		}
		r.resolveExpressionIfNotNil(d.StringLength)
	}
}

func (r *Resolver) resolveRedim(s *ast.RedimStatement) {
	for _, d := range s.Declarations {
		name := d.Name + d.TypeSuffix
		dt := r.resolveElementType(d.ElementType, d.TypeSuffix, d.Name)
		dims := len(d.Dimensions)

		// REDIM replaces existing dimension info.
		if existing := r.Table.Lookup(name); existing != nil {
			existing.ArrayDims = dims
			existing.DataType = dt
		} else {
			_ = r.Table.Define(name, &Symbol{
				Name:      name,
				Type:      SymArray,
				DataType:  dt,
				Scope:     r.Table.CurrentScope.Name,
				Defined:   true,
				Line:      d.BasePos.Line,
				Column:    d.BasePos.Column,
				ArrayDims: dims,
			})
		}
		for _, dr := range d.Dimensions {
			r.resolveExpressionIfNotNil(dr.Lower)
			r.resolveExpressionIfNotNil(dr.Upper)
		}
	}
}

func (r *Resolver) resolveFor(s *ast.ForStatement) {
	// The counter variable is implicitly declared.
	name := s.Counter.Name + s.Counter.TypeSuffix
	dt := r.Table.ResolveType(name)
	if sym := r.Table.Lookup(name); sym != nil {
		sym.Defined = true
		sym.Used = true
	} else {
		_ = r.Table.Define(name, &Symbol{
			Name:     name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    r.Table.CurrentScope.Name,
			Defined:  true,
			Line:     s.Counter.Pos().Line,
			Column:   s.Counter.Pos().Column,
		})
	}
	r.resolveExpression(s.Start)
	r.resolveExpression(s.End)
	r.resolveExpressionIfNotNil(s.Step)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveWhile(s *ast.WhileStatement) {
	r.resolveExpression(s.Condition)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveDoLoop(s *ast.DoLoopStatement) {
	r.resolveExpressionIfNotNil(s.Condition)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveIf(s *ast.IfStatement) {
	r.resolveExpression(s.Condition)
	for _, stmt := range s.ThenBlock {
		r.resolveStatement(stmt)
	}
	for _, clause := range s.ElseIfClauses {
		r.resolveExpression(clause.Condition)
		for _, stmt := range clause.Body {
			r.resolveStatement(stmt)
		}
	}
	for _, stmt := range s.ElseBlock {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveSelectCase(s *ast.SelectCaseStatement) {
	r.resolveExpression(s.TestExpr)
	for _, c := range s.Cases {
		for _, cv := range c.Values {
			r.resolveExpressionIfNotNil(cv.Value)
			r.resolveExpressionIfNotNil(cv.EndValue)
		}
		for _, stmt := range c.Body {
			r.resolveStatement(stmt)
		}
	}
	for _, stmt := range s.ElseBlock {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolvePrint(s *ast.PrintStatement) {
	for _, expr := range s.Expressions {
		r.resolveExpression(expr)
	}
	r.resolveExpressionIfNotNil(s.Format)
}

func (r *Resolver) resolveSubDecl(s *ast.SubDeclaration) {
	if s.IsForward {
		return // forward declaration already registered in pass 1
	}
	// TUTORIAL — Scope management for SUB declarations
	//
	// When we enter a SUB body, we open a new scope. This implements BASIC's
	// local variable rules: variables declared (or implicitly created) inside
	// a SUB are local to that SUB and do not pollute the global scope.
	//
	// The EnterScope/ExitScope pair acts like a stack push/pop:
	//   - EnterScope makes a new scope that is a child of the current one.
	//   - All symbol definitions during the walk go into this child scope.
	//   - ExitScope returns to the parent, making the child scope inaccessible
	//     to subsequent code outside the SUB.
	//
	// Parameters are explicitly defined in the new scope before the body is
	// walked, so that they are visible as local variables inside the SUB.
	r.Table.EnterScope(s.Name)
	// Register parameters as local variables.
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveFunctionDecl(s *ast.FunctionDeclaration) {
	if s.IsForward {
		return
	}
	r.Table.EnterScope(s.Name)
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	// The function name itself is a local variable that holds the return value.
	// TUTORIAL — BASIC's return-value mechanism
	//
	// In BASIC, a FUNCTION returns a value by assigning to the function's own
	// name inside the body:
	//
	//   FUNCTION Square(x AS SINGLE) AS SINGLE
	//     Square = x * x      ' assign to function name to set return value
	//   END FUNCTION
	//
	// This is implemented by registering the function name as a local variable
	// inside the function's scope. The codegen phase later translates the
	// assignment "Square = x * x" into a Go return statement. Defining the
	// function name as a symbol here allows the assignment to be resolved
	// normally as "variable assignment" rather than requiring special handling.
	rt := resolveReturnType(s.ReturnType, s.Name, r.Table)
	_ = r.Table.Define(s.Name, &Symbol{
		Name:       s.Name,
		Type:       SymVariable,
		DataType:   rt,
		Scope:      strings.ToUpper(s.Name),
		Defined:    true,
		Line:       s.Pos().Line,
		Column:     s.Pos().Column,
		ReturnType: rt,
	})
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveDefFn(s *ast.DefFnDeclaration) {
	r.Table.EnterScope(s.Name)
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	r.resolveExpressionIfNotNil(s.SingleLineExpr)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveGoto(s *ast.GotoStatement) {
	// TUTORIAL — Deferred label validation
	//
	// We cannot validate GOTO targets immediately during Pass 2 because the
	// target label may appear later in the source (a forward jump). Instead,
	// we record the reference in r.labelRefs and validate all references after
	// Pass 2 completes (in the post-pass of Resolve()). This is a standard
	// technique called "fixup" or "deferred resolution": collect references now,
	// validate them later when you are certain all definitions are known.
	r.labelRefs = append(r.labelRefs, labelRef{
		name:   s.Target,
		line:   s.Pos().Line,
		column: s.Pos().Column,
	})
}

func (r *Resolver) resolveGosub(s *ast.GosubStatement) {
	r.labelRefs = append(r.labelRefs, labelRef{
		name:   s.Target,
		line:   s.Pos().Line,
		column: s.Pos().Column,
	})
}

func (r *Resolver) resolveOnErrorGoto(s *ast.OnErrorGotoStatement) {
	if s.Target != "0" {
		r.labelRefs = append(r.labelRefs, labelRef{
			name:   s.Target,
			line:   s.Pos().Line,
			column: s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveOnEventGosub(s *ast.OnEventGosubStatement) {
	r.resolveExpressionIfNotNil(s.EventParam)
	if s.Target != "" {
		r.labelRefs = append(r.labelRefs, labelRef{
			name:   s.Target,
			line:   s.Pos().Line,
			column: s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveDefType(s *ast.DefTypeStatement) {
	var dt DataType
	switch strings.ToUpper(s.Type) {
	case "DEFINT":
		dt = TypeInteger
	case "DEFLNG":
		dt = TypeLong
	case "DEFSNG":
		dt = TypeSingle
	case "DEFDBL":
		dt = TypeDouble
	case "DEFSTR":
		dt = TypeString
	default:
		return
	}
	for _, lr := range s.LetterRanges {
		r.Table.SetDefType(lr.Start, lr.End, dt)
	}
}

func (r *Resolver) resolveOptionBase(s *ast.OptionBaseStatement) {
	if r.Table.OptionBase != 0 && r.Table.OptionBase != s.Value {
		r.addError(s.Pos().Line, s.Pos().Column,
			"OPTION BASE conflict: previously set to %d, now %d", r.Table.OptionBase, s.Value)
	}
	r.Table.OptionBase = s.Value
}

func (r *Resolver) resolveScopeStmt(s *ast.ScopeStatement) {
	mod := strings.ToUpper(s.Modifier)
	for _, varName := range s.Variables {
		key := strings.ToUpper(varName)
		switch mod {
		case "SHARED":
			// SHARED inside a SUB/FUNCTION makes a global variable visible.
			if gsym := r.Table.GlobalScope.Symbols[key]; gsym != nil {
				gsym.IsShared = true
				// Also add an alias in the current scope.
				r.Table.CurrentScope.Symbols[key] = gsym
			} else {
				// Implicit global declaration.
				dt := r.Table.ResolveType(varName)
				sym := &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    "global",
					Defined:  true,
					IsShared: true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				}
				r.Table.GlobalScope.Symbols[key] = sym
				r.Table.CurrentScope.Symbols[key] = sym
			}
		case "STATIC":
			if sym := r.Table.LookupLocal(varName); sym != nil {
				sym.IsStatic = true
			} else {
				dt := r.Table.ResolveType(varName)
				_ = r.Table.Define(varName, &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    r.Table.CurrentScope.Name,
					Defined:  true,
					IsStatic: true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				})
			}
		default:
			// LOCAL, COMMON – just ensure a symbol exists.
			if r.Table.LookupLocal(varName) == nil {
				dt := r.Table.ResolveType(varName)
				_ = r.Table.Define(varName, &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    r.Table.CurrentScope.Name,
					Defined:  true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				})
			}
		}
	}
}

func (r *Resolver) resolveArrayAssignment(s *ast.ArrayAssignment) {
	r.resolveExpression(s.Value)
	// Resolve index expressions.
	for _, idx := range s.Array.Indices {
		r.resolveExpression(idx)
	}
	// Mark the array as used/defined.
	name := s.Array.Name + s.Array.TypeSuffix
	if sym := r.Table.Lookup(name); sym != nil {
		sym.Defined = true
		sym.Used = true
	}
}

func (r *Resolver) resolveRead(s *ast.ReadStatement) {
	for _, v := range s.Variables {
		r.resolveExpression(v)
	}
}

func (r *Resolver) resolveSwap(s *ast.SwapStatement) {
	r.resolveExpression(s.Var1)
	r.resolveExpression(s.Var2)
}

func (r *Resolver) resolveIncr(s *ast.IncrStatement) {
	r.resolveExpression(s.Variable)
	r.resolveExpressionIfNotNil(s.Amount)
}

func (r *Resolver) resolveDecr(s *ast.DecrStatement) {
	r.resolveExpression(s.Variable)
	r.resolveExpressionIfNotNil(s.Amount)
}

func (r *Resolver) resolveOpen(s *ast.OpenStatement) {
	r.resolveExpressionIfNotNil(s.Filename)
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecLen)
}

func (r *Resolver) resolveClose(s *ast.CloseStatement) {
	for _, f := range s.FileNums {
		r.resolveExpression(f)
	}
}

func (r *Resolver) resolveFileInput(s *ast.FileInputStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, v := range s.Variables {
		r.resolveExpression(v)
	}
}

func (r *Resolver) resolveFilePrint(s *ast.FilePrintStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, e := range s.Expressions {
		r.resolveExpression(e)
	}
	r.resolveExpressionIfNotNil(s.Format)
}

func (r *Resolver) resolveFileWrite(s *ast.FileWriteStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, e := range s.Expressions {
		r.resolveExpression(e)
	}
}

func (r *Resolver) resolveGet(s *ast.GetStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecordOrPos)
	r.resolveExpressionIfNotNil(s.Variable)
}

func (r *Resolver) resolvePut(s *ast.PutStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecordOrPos)
	r.resolveExpressionIfNotNil(s.Variable)
}

func (r *Resolver) resolveSeek(s *ast.SeekStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.Position)
}
