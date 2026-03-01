package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// DIM
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitDim(s *ast.DimStatement) {
	for _, d := range s.Declarations {
		name := mangleName(d.Name + d.TypeSuffix)
		if len(d.Dimensions) == 0 {
			// Scalar declaration.
			goT := g.goTypeForDecl(d)
			if !g.declared[name] {
				g.writeLinef("var %s %s", name, goT)
			}
			g.declared[name] = true
		} else if len(d.Dimensions) == 1 {
			// 1-D array.
			goT := g.goTypeForDecl(d)
			upper := g.emitExpr(d.Dimensions[0].Upper)
			if g.declared[name] {
				// Variable already hoisted; use assignment instead of short declaration.
				g.writeLinef("%s = make([]%s, int(%s)+1)", name, goT, upper)
			} else {
				g.writeLinef("%s := make([]%s, int(%s)+1)", name, goT, upper)
			}
			g.declared[name] = true
			g.arrayDims[name] = 1
		} else {
			// Multi-dimensional: emit slice-of-slices with nested init loops.
			g.emitMultiDimArray(name, d)
		}
	}
}

// emitMultiDimArray emits a multi-dimensional array as nested slices.
// For example, DIM A(10, 20) becomes:
//
//	A := make([][]float32, 11)
//	for i_ := range A { A[i_] = make([]float32, 21) }
//
// DIM B(5, 10, 3) becomes:
//
//	B := make([][][]float32, 6)
//	for i_ := range B { B[i_] = make([][]float32, 11); for j_ := range B[i_] { B[i_][j_] = make([]float32, 4) } }
func (g *CodeGenerator) emitMultiDimArray(name string, d ast.DimDecl) {
	ndim := len(d.Dimensions)
	goT := g.goTypeForDecl(d)

	// Build the slice-of-slices type prefix: e.g. "[][]" for 2D, "[][][]" for 3D.
	slicePrefix := strings.Repeat("[]", ndim)

	// Outermost make: e.g. make([][]float32, 11)
	upper0 := g.emitExpr(d.Dimensions[0].Upper)
	outerType := slicePrefix + goT // e.g. "[][][]float32" for 3D — but outermost make uses "[]" of inner type
	// Actually the outermost make type is slicePrefix + goT but we need the full nesting.
	// For 2D: make([][]float32, N)  — the outer type is [][]float32
	// For 3D: make([][][]float32, N)
	assign := ":="
	if g.declared[name] {
		assign = "="
	}
	g.writeLinef("%s %s make(%s, int(%s)+1)", name, assign, outerType, upper0)
	g.declared[name] = true
	g.arrayDims[name] = ndim

	// Now emit the nested initialization loops.
	// For 2D (dims=[d0, d1]):
	//   for i_ := range A { A[i_] = make([]float32, int(d1)+1) }
	// For 3D (dims=[d0, d1, d2]):
	//   for i_ := range A { A[i_] = make([][]float32, int(d1)+1); for j_ := range A[i_] { A[i_][j_] = make([]float32, int(d2)+1) } }
	loopVars := []string{"i_", "j_", "k_", "l_", "m_"}
	if ndim-1 > len(loopVars) {
		// Fallback for very high dimensions (unlikely in BASIC).
		for extra := len(loopVars); extra < ndim-1; extra++ {
			loopVars = append(loopVars, fmt.Sprintf("idx%d_", extra))
		}
	}

	// Build a single line with nested for loops.
	var line strings.Builder
	for dim := 1; dim < ndim; dim++ {
		lv := loopVars[dim-1]

		// Build the accessor chain: A[i_][j_]...
		accessor := name
		for dd := 0; dd < dim; dd++ {
			accessor += fmt.Sprintf("[%s]", loopVars[dd])
		}

		// The range target is the parent: A (for dim=1), A[i_] (for dim=2), etc.
		rangeTarget := name
		for dd := 0; dd < dim-1; dd++ {
			rangeTarget += fmt.Sprintf("[%s]", loopVars[dd])
		}

		// The make type: remaining []'s + goT
		remainingSlices := strings.Repeat("[]", ndim-dim)
		makeType := remainingSlices + goT
		upper := g.emitExpr(d.Dimensions[dim].Upper)

		if dim > 1 {
			line.WriteString("; ")
		}
		line.WriteString(fmt.Sprintf("for %s := range %s { %s = make(%s, int(%s)+1)", lv, rangeTarget, accessor, makeType, upper))
	}
	// Close all the braces.
	for dim := ndim - 1; dim >= 1; dim-- {
		line.WriteString(" }")
	}

	g.writeLine(line.String())
}

func (g *CodeGenerator) emitRedim(s *ast.RedimStatement) {
	for _, d := range s.Declarations {
		name := mangleName(d.Name + d.TypeSuffix)
		if len(d.Dimensions) == 1 {
			goT := g.goTypeForDecl(d)
			upper := g.emitExpr(d.Dimensions[0].Upper)
			g.writeLinef("%s = make([]%s, int(%s)+1) // REDIM", name, goT, upper)
			g.arrayDims[name] = 1
		} else if len(d.Dimensions) > 1 {
			// Multi-dimensional REDIM: reuse the same slice-of-slices emitter.
			// Force assignment (not short decl) since REDIM implies the var exists.
			g.declared[name] = true
			g.emitMultiDimArray(name, d)
		}
	}
}

// ---------------------------------------------------------------------------
// READ / INPUT from stdin
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitRead(s *ast.ReadStatement) {
	// INPUT statement: read from stdin.
	if s.IsInput {
		g.emitInputFromStdin(s)
		return
	}

	// DATA READ: read from the compile-time data pool.
	for _, v := range s.Variables {
		varExpr := g.emitExpr(v)

		// Determine the Go type from the variable expression.
		goT := ""
		isStr := false
		switch vt := v.(type) {
		case *ast.Identifier:
			name := mangleName(vt.Name + vt.TypeSuffix)
			if !g.declared[name] {
				goT = g.goTypeForIdent(vt.Name + vt.TypeSuffix)
				g.writeLinef("var %s %s", name, goT)
				g.declared[name] = true
			}
			goT = g.goTypeForIdent(vt.Name + vt.TypeSuffix)
			isStr = isStringType(vt.Name + vt.TypeSuffix)
		case *ast.ArrayAccess:
			goT = g.goTypeForIdent(vt.Name + vt.TypeSuffix)
			isStr = isStringType(vt.Name + vt.TypeSuffix)
		}

		if goT == "" {
			// Unknown target — discard (should not happen in practice).
			g.writeLinef("_ = dataPool[dataIdx]; dataIdx++ // READ into %s", varExpr)
			continue
		}

		if isStr {
			g.writeLinef("%s = fmt.Sprint(dataPool[dataIdx]); dataIdx++", varExpr)
		} else {
			g.imports["fmt"] = true
			g.writeLinef("{ v_ := dataPool[dataIdx]; dataIdx++; switch tv_ := v_.(type) { case float64: %s = %s(tv_); case int: %s = %s(float64(tv_)); case string: %s = %s(rt.Val(tv_)); default: _ = tv_ } }", varExpr, goT, varExpr, goT, varExpr, goT)
		}
	}
}

// emitInputFromStdin emits stdin-reading code for INPUT and LINE INPUT statements.
// INPUT reads a comma-separated line from stdin; LINE INPUT reads the whole line.
func (g *CodeGenerator) emitInputFromStdin(s *ast.ReadStatement) {
	g.imports["fmt"] = true

	// Print the prompt if there is one (INPUT "Prompt: "; var).
	if s.Prompt != "" {
		g.writeLinef("fmt.Print(%q)", s.Prompt)
	}

	if s.IsLineInput {
		// LINE INPUT: read entire line into a single string variable.
		if len(s.Variables) > 0 {
			varExpr := g.emitExpr(s.Variables[0])
			if ident, ok := s.Variables[0].(*ast.Identifier); ok {
				name := mangleName(ident.Name + ident.TypeSuffix)
				if !g.declared[name] {
					g.writeLinef("var %s string", name)
					g.declared[name] = true
				}
			}
			g.writeLine("{ scanner_ := rt.NewScanner(); scanner_.Scan()")
			g.writeLinef("  %s = scanner_.Text() }", varExpr)
		}
		return
	}

	// Regular INPUT: read one comma-separated line, then assign parts to variables.
	if len(s.Variables) == 1 {
		// Single variable: read directly.
		v := s.Variables[0]
		varExpr := g.emitExpr(v)
		if ident, ok := v.(*ast.Identifier); ok {
			name := mangleName(ident.Name + ident.TypeSuffix)
			if !g.declared[name] {
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("var %s %s", name, goT)
				g.declared[name] = true
			}
			if isStringType(ident.Name + ident.TypeSuffix) {
				g.writeLine("{ scanner_ := rt.NewScanner(); scanner_.Scan()")
				g.writeLinef("  %s = scanner_.Text() }", varExpr)
			} else {
				g.writeLine("{ scanner_ := rt.NewScanner(); scanner_.Scan()")
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("  %s = %s(rt.Val(scanner_.Text())) }", varExpr, goT)
			}
		}
		return
	}

	// Multiple variables: read one line, split by comma, assign parts.
	// Ensure all variables are declared first.
	for _, v := range s.Variables {
		if ident, ok := v.(*ast.Identifier); ok {
			name := mangleName(ident.Name + ident.TypeSuffix)
			if !g.declared[name] {
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("var %s %s", name, goT)
				g.declared[name] = true
			}
		}
	}

	g.writeLine("{")
	g.writeLinef("  parts_ := rt.InputSplitLine()")
	for i, v := range s.Variables {
		varExpr := g.emitExpr(v)
		if ident, ok := v.(*ast.Identifier); ok {
			if isStringType(ident.Name + ident.TypeSuffix) {
				g.writeLinef("  if len(parts_) > %d { %s = parts_[%d] }", i, varExpr, i)
			} else {
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("  if len(parts_) > %d { %s = %s(rt.Val(parts_[%d])) }", i, varExpr, goT, i)
			}
		}
	}
	g.writeLine("}")
}

// ---------------------------------------------------------------------------
// SUB / FUNCTION / DEF FN – top-level declarations
//
// emitTopLevelDecl() is a second visitor entry point used exclusively for
// procedure-level constructs.  In BASIC, SUB and FUNCTION blocks can appear
// anywhere in the source file, but in Go they must be top-level declarations
// outside of any other function.
//
// The strategy is to redirect writes from the main buffer (g.buf) to a
// separate buffer (g.funcBuf) while walking the procedure body.  After the
// entire program has been processed, funcBuf is appended to the output after
// the closing brace of main(), producing valid Go.
//
// emitSubDecl / emitFuncDecl temporarily swap g.buf with a fresh buffer so
// that nested emitStatement() calls still write through the same pointer,
// then move the result into funcBuf.  The indent counter is also saved and
// restored so indentation inside the procedure body starts at level 1.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitTopLevelDecl(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.SubDeclaration:
		g.emitSubDecl(s)
	case *ast.FunctionDeclaration:
		g.emitFuncDecl(s)
	case *ast.DefFnDeclaration:
		g.emitDefFn(s)
	}
}

func (g *CodeGenerator) emitSubDecl(s *ast.SubDeclaration) {
	if s.IsForward {
		return
	}
	name := mangleName(s.Name)
	// Detect which parameters are assigned to in the body (need by-ref).
	byRef := g.findAssignedParams(s.Body, s.Params)
	params := g.emitParams(s.Params, byRef)
	g.funcWriteLinef(0, "func %s(%s) {", name, params)
	// Save and restore state — including the declared map so SUBs get their own scope.
	origBuf := g.buf
	origIndent := g.indent
	origDeclared := g.declared
	origInSub := g.inSubOrFunc
	origParamsByRef := g.paramsByRef
	g.buf = bytes.Buffer{}
	g.indent = 1
	g.inSubOrFunc = true
	g.paramsByRef = byRef
	// Fresh declared map for the SUB scope. Pre-populate with:
	// - shared/package-level variables (already declared at package level)
	// - parameters
	origHoistedVars := g.hoistedVars
	origHoistedSet := g.hoistedSet
	origHoistedTypes := g.hoistedTypes
	g.declared = make(map[string]bool)
	g.hoistedVars = nil
	g.hoistedSet = make(map[string]bool)
	g.hoistedTypes = make(map[string]string)
	for k := range g.sharedVars {
		g.declared[k] = true
		g.hoistedSet[k] = true
	}
	for k := range g.packageVarSet {
		g.declared[k] = true
		g.hoistedSet[k] = true
	}
	for _, p := range s.Params {
		pn := mangleName(p.Name)
		g.declared[pn] = true
		g.hoistedSet[pn] = true
	}
	// Hoist variables used in the SUB body to the top of the function.
	g.collectMainVariables(s.Body)
	for _, hv := range g.hoistedVars {
		g.declared[hv.name] = true
	}
	if len(g.hoistedVars) > 0 {
		for _, hv := range g.hoistedVars {
			g.writeLinef("var %s %s", hv.name, hv.typ)
		}
		// Suppress unused variable errors for hoisted variables.
		for _, hv := range g.hoistedVars {
			g.writeLinef("_ = %s", hv.name)
		}
	}
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.funcBuf.Write(g.buf.Bytes())
	g.buf = origBuf
	g.indent = origIndent
	g.declared = origDeclared
	g.hoistedVars = origHoistedVars
	g.hoistedSet = origHoistedSet
	g.hoistedTypes = origHoistedTypes
	g.inSubOrFunc = origInSub
	g.paramsByRef = origParamsByRef
	g.funcWriteLine(0, "}")
	g.funcWriteLine(0, "")
}

func (g *CodeGenerator) emitFuncDecl(s *ast.FunctionDeclaration) {
	if s.IsForward {
		return
	}
	name := mangleName(s.Name)
	// Detect which parameters are assigned to in the body (need by-ref).
	byRef := g.findAssignedParams(s.Body, s.Params)
	params := g.emitParams(s.Params, byRef)
	retType := g.goTypeForReturnType(s.ReturnType, s.Name)
	retVar := mangleName(s.Name)

	g.funcWriteLinef(0, "func %s(%s) %s {", name, params, retType)
	g.funcWriteLinef(1, "var %s %s", retVar, retType)

	// Save and restore state — including the declared map so FUNCTIONs get their own scope.
	origBuf := g.buf
	origIndent := g.indent
	origDeclared := g.declared
	origInSub := g.inSubOrFunc
	origParamsByRef := g.paramsByRef
	g.buf = bytes.Buffer{}
	g.indent = 1
	g.inSubOrFunc = true
	g.paramsByRef = byRef
	// Fresh declared map for the FUNCTION scope. Pre-populate with:
	// - shared/package-level variables (already declared at package level)
	// - parameters
	// - the return variable (just declared above)
	origHoistedVars := g.hoistedVars
	origHoistedSet := g.hoistedSet
	origHoistedTypes := g.hoistedTypes
	g.declared = make(map[string]bool)
	g.hoistedVars = nil
	g.hoistedSet = make(map[string]bool)
	g.hoistedTypes = make(map[string]string)
	for k := range g.sharedVars {
		g.declared[k] = true
		g.hoistedSet[k] = true
	}
	for k := range g.packageVarSet {
		g.declared[k] = true
		g.hoistedSet[k] = true
	}
	for _, p := range s.Params {
		pn := mangleName(p.Name)
		g.declared[pn] = true
		g.hoistedSet[pn] = true
	}
	g.declared[retVar] = true
	g.hoistedSet[retVar] = true
	// Hoist variables used in the FUNCTION body to the top of the function.
	g.collectMainVariables(s.Body)
	for _, hv := range g.hoistedVars {
		g.declared[hv.name] = true
	}
	if len(g.hoistedVars) > 0 {
		for _, hv := range g.hoistedVars {
			g.writeLinef("var %s %s", hv.name, hv.typ)
		}
		// Suppress unused variable errors for hoisted variables.
		for _, hv := range g.hoistedVars {
			g.writeLinef("_ = %s", hv.name)
		}
	}
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.funcBuf.Write(g.buf.Bytes())
	g.buf = origBuf
	g.indent = origIndent
	g.declared = origDeclared
	g.hoistedVars = origHoistedVars
	g.hoistedSet = origHoistedSet
	g.hoistedTypes = origHoistedTypes
	g.inSubOrFunc = origInSub
	g.paramsByRef = origParamsByRef

	g.funcWriteLinef(1, "return %s", retVar)
	g.funcWriteLine(0, "}")
	g.funcWriteLine(0, "")
}

func (g *CodeGenerator) emitDefFn(s *ast.DefFnDeclaration) {
	// When the source used the two-token form "DEF FN name(params)" (with a
	// space), the parser stores just "name" in s.Name.  The call-site emitter
	// (emitFnCallExpression) adds an "fn_" prefix so that the call becomes
	// fn_name(...).  We must use the same prefix in the definition so the
	// symbols match.
	//
	// When the source used the one-token form "DEF FNname(params)" (no space),
	// the parser stores "FNname" in s.Name.  FunctionCall (not FnCallExpression)
	// is used at the call site and emits mangleName("FNname") = "FNname", so
	// no prefix is needed.
	var name string
	if strings.HasPrefix(strings.ToUpper(s.Name), "FN") {
		// One-token form: FNfoo → function is called as FNfoo(...)
		name = mangleName(s.Name)
	} else {
		// Two-token form: FN foo → function is called as fn_foo(...)
		name = "fn_" + mangleName(s.Name)
	}
	params := g.emitParams(s.Params, nil)
	retType := g.goTypeForIdent(s.Name)

	if s.SingleLineExpr != nil {
		g.funcWriteLinef(0, "func %s(%s) %s {", name, params, retType)
		expr := g.emitExpr(s.SingleLineExpr)
		g.funcWriteLinef(1, "return %s(%s)", retType, expr)
		g.funcWriteLine(0, "}")
		g.funcWriteLine(0, "")
	} else {
		g.funcWriteLinef(0, "func %s(%s) %s {", name, params, retType)

		origBuf := g.buf
		origIndent := g.indent
		origInDefFn := g.inDefFn
		origDefFnReturnType := g.defFnReturnType
		origDefFnName := g.defFnName
		origDeclared := g.declared
		origHoistedVars := g.hoistedVars
		origHoistedSet := g.hoistedSet
		origHoistedTypes := g.hoistedTypes
		g.buf = bytes.Buffer{}
		g.indent = 1
		g.inDefFn = true
		g.defFnReturnType = retType
		g.defFnName = s.Name
		// Use a fresh declared map so that variables hoisted in main() are not
		// treated as already-declared inside this DEF FN body.  Each DEF FN
		// function has its own local scope in Go.
		g.declared = make(map[string]bool)
		g.hoistedVars = nil
		g.hoistedSet = make(map[string]bool)
		g.hoistedTypes = make(map[string]string)
		// Pre-seed with parameter names so we don't redeclare them.
		for _, p := range s.Params {
			pn := mangleName(p.Name)
			g.declared[pn] = true
			g.hoistedSet[pn] = true // don't hoist parameters
		}
		// Hoist variables used in the body to the top of the function so they
		// are visible across all branches (e.g. SELECT CASE).
		g.collectMainVariables(s.Body)
		for _, hv := range g.hoistedVars {
			g.declared[hv.name] = true
		}
		// Emit hoisted local variable declarations.
		if len(g.hoistedVars) > 0 {
			for _, hv := range g.hoistedVars {
				g.writeLinef("var %s %s", hv.name, hv.typ)
			}
		}
		for _, stmt := range s.Body {
			g.emitStatement(stmt)
		}
		g.funcBuf.Write(g.buf.Bytes())
		g.buf = origBuf
		g.indent = origIndent
		g.inDefFn = origInDefFn
		g.defFnReturnType = origDefFnReturnType
		g.defFnName = origDefFnName
		g.declared = origDeclared
		g.hoistedVars = origHoistedVars
		g.hoistedSet = origHoistedSet
		g.hoistedTypes = origHoistedTypes

		g.funcWriteLinef(1, "return %s", zeroValueForType(retType))
		g.funcWriteLine(0, "}")
		g.funcWriteLine(0, "")
	}
}

func (g *CodeGenerator) emitParams(params []ast.Parameter, byRef map[string]bool) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		name := mangleName(p.Name)
		goT := g.goTypeForParamType(p.Type, p.Name)
		if p.IsArray {
			goT = "[]" + goT
		} else if byRef[name] {
			goT = "*" + goT
		}
		parts = append(parts, fmt.Sprintf("%s %s", name, goT))
	}
	return strings.Join(parts, ", ")
}
