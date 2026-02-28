package parser

import (
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// parseDimStatement handles DIM (declare) and REDIM (re-dimension) statements.
//
// DIM allocates memory for variables and arrays. Without DIM, BASIC creates
// variables implicitly on first use. With DIM the programmer can:
//
//   1. Declare the type of a scalar variable:  DIM x AS DOUBLE
//   2. Allocate a fixed-size array:            DIM arr(10) AS INTEGER
//   3. Declare a multi-dimensional array:      DIM grid(5, 5)
//   4. Specify lower and upper bounds:         DIM a(1 TO 10)
//
// REDIM is DIM's runtime variant: it can change array dimensions after
// the initial allocation, which is useful for dynamic data structures.
//
// The isRedim parameter distinguishes the two so the caller can pass the
// right keyword token type; both are otherwise parsed identically.
func (p *Parser) parseDimStatement(isRedim bool) ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DIM/REDIM

	// Check for optional scope modifiers: SHARED, LOCAL, STATIC, COMMON, DYNAMIC
	isShared := false
	for p.curTokenIs(lexer.TOKEN_SHARED) || p.curTokenIs(lexer.TOKEN_LOCAL) ||
		p.curTokenIs(lexer.TOKEN_STATIC) || p.curTokenIs(lexer.TOKEN_COMMON) ||
		(p.curTokenIs(lexer.TOKEN_IDENTIFIER) && strings.ToUpper(p.curToken.Literal) == "DYNAMIC") {
		if p.curTokenIs(lexer.TOKEN_SHARED) {
			isShared = true
		}
		p.nextToken()
	}

	var decls []ast.DimDecl

	for {
		decl := p.parseDimDecl()
		decls = append(decls, decl)
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // skip comma
	}

	if isRedim {
		return &ast.RedimStatement{BasePos: pos, Declarations: decls}
	}
	return &ast.DimStatement{BasePos: pos, Declarations: decls, IsShared: isShared}
}

func (p *Parser) parseDimDecl() ast.DimDecl {
	decl := ast.DimDecl{BasePos: p.curPos()}

	name := p.curToken.Literal
	// Split type suffix
	if len(name) > 0 {
		last := name[len(name)-1]
		if last == '%' || last == '&' || last == '!' || last == '#' || last == '$' {
			decl.TypeSuffix = string(last)
			name = name[:len(name)-1]
		}
	}
	decl.Name = name
	p.nextToken()

	// Parse dimensions
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken() // skip (
		for {
			dimRange := ast.DimRange{BasePos: p.curPos()}
			first := p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_TO) {
				p.nextToken() // skip TO
				dimRange.Lower = first
				dimRange.Upper = p.parseExpression(PREC_LOWEST)
			} else {
				dimRange.Upper = first
			}
			decl.Dimensions = append(decl.Dimensions, dimRange)
			if !p.curTokenIs(lexer.TOKEN_COMMA) {
				break
			}
			p.nextToken()
		}
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken() // skip )
		}
	}

	// AS type
	if p.curTokenIs(lexer.TOKEN_AS) {
		p.nextToken() // skip AS
		if p.curTokenIs(lexer.TOKEN_STRING_KW) {
			decl.ElementType = "STRING"
			p.nextToken()
			if p.curTokenIs(lexer.TOKEN_STAR) {
				p.nextToken() // skip *
				decl.StringLength = p.parseExpression(PREC_LOWEST)
			}
		} else {
			decl.ElementType = strings.ToUpper(p.curToken.Literal)
			p.nextToken()
		}
	}

	return decl
}

// parseSubDeclaration parses SUB name(params) … END SUB.
//
// SUBs are the primary code-organisation mechanism in Turbo BASIC. Unlike
// functions, they do not return a value and are called as statements (not
// inside expressions).
//
// The parsing strategy mirrors the general recursive-descent pattern:
//   1. Consume the SUB keyword and read the name.
//   2. Parse the parameter list (delegated to parseParameterList).
//   3. Call parseBlockUntil(TOKEN_END) to collect all body statements.
//      The sentinel is TOKEN_END; parseBlockUntil knows to stop only when
//      TOKEN_END is followed by TOKEN_SUB (via isEndBlock()), distinguishing
//      "END SUB" from a bare "END" program-terminator.
//   4. Consume "END SUB".
func (p *Parser) parseSubDeclaration() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SUB
	name := p.curToken.Literal
	p.nextToken() // skip name

	params := p.parseParameterList()

	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	body := p.parseBlockUntil(lexer.TOKEN_END)

	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken() // skip END
		if p.curTokenIs(lexer.TOKEN_SUB) {
			p.nextToken() // skip SUB
		}
	}

	return &ast.SubDeclaration{BasePos: pos, Name: name, Params: params, Body: body}
}

// parseFunctionDeclaration parses FUNCTION name(params) [AS type] … END FUNCTION.
//
// The structure is almost identical to parseSubDeclaration with one addition:
// the optional "AS type" return-type annotation. In Go output, this annotation
// becomes the function's return type. Without it, the return type is inferred
// from the function name's suffix character or defaults to float32.
//
// A student exercise: notice that both parseSubDeclaration and
// parseFunctionDeclaration call parseParameterList and parseBlockUntil with
// the same logic. A future refactor could extract a shared "parseCallable"
// helper that handles both cases — but only if the shared logic grows large
// enough to justify the indirection.
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip FUNCTION
	name := p.curToken.Literal
	p.nextToken() // skip name

	params := p.parseParameterList()

	retType := ""
	if p.curTokenIs(lexer.TOKEN_AS) {
		p.nextToken() // skip AS
		retType = strings.ToUpper(p.curToken.Literal)
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	body := p.parseBlockUntil(lexer.TOKEN_END)

	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken() // skip END
		if p.curTokenIs(lexer.TOKEN_FUNCTION) {
			p.nextToken() // skip FUNCTION
		}
	}

	return &ast.FunctionDeclaration{
		BasePos: pos, Name: name, Params: params, ReturnType: retType, Body: body,
	}
}

func (p *Parser) parseDeclare() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DECLARE

	if p.curTokenIs(lexer.TOKEN_SUB) {
		p.nextToken() // skip SUB
		name := p.curToken.Literal
		p.nextToken()
		params := p.parseParameterList()
		return &ast.SubDeclaration{BasePos: pos, Name: name, Params: params, IsForward: true}
	}

	if p.curTokenIs(lexer.TOKEN_FUNCTION) {
		p.nextToken() // skip FUNCTION
		name := p.curToken.Literal
		p.nextToken()
		params := p.parseParameterList()
		retType := ""
		if p.curTokenIs(lexer.TOKEN_AS) {
			p.nextToken()
			retType = strings.ToUpper(p.curToken.Literal)
			p.nextToken()
		}
		return &ast.FunctionDeclaration{
			BasePos: pos, Name: name, Params: params, ReturnType: retType, IsForward: true,
		}
	}

	p.addError("expected SUB or FUNCTION after DECLARE")
	p.skipToEndOfLine()
	return nil
}

// parseParameterList parses a parenthesised formal-parameter list.
//
// The list is optional: if the current token is not '(' there are no
// parameters and an empty slice is returned. This handles:
//
//   SUB Greet            ← no parameter list at all
//   SUB Greet()          ← explicit empty list
//   SUB Greet(name$)     ← one parameter
//   SUB Fill(arr() AS INTEGER, n AS INTEGER) ← array parameter + scalar
//
// The "()" after a parameter name signals that the parameter is an array
// (passed by reference). This is syntactic sugar; the parameter is still
// stored as a regular Parameter with a name like "arr".
//
// BYVAL is the only modifier handled here. BASIC's default passing convention
// is by reference (a pointer to the caller's variable), so BYVAL is the
// exception, not the rule. The IsByVal flag in the resulting Parameter lets
// the code generator emit the correct Go pointer/value semantics.
func (p *Parser) parseParameterList() []ast.Parameter {
	var params []ast.Parameter

	if !p.curTokenIs(lexer.TOKEN_LPAREN) {
		return params
	}
	p.nextToken() // skip (

	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
		return params
	}

	for {
		param := ast.Parameter{BasePos: p.curPos()}

		if p.curTokenIs(lexer.TOKEN_BYVAL) {
			param.IsByVal = true
			p.nextToken()
		}

		param.Name = p.curToken.Literal
		p.nextToken()

		// Handle array parameter marker: paramName() — the () indicates the param
		// is passed as an array reference (e.g., SUB Foo (arr() AS Integer))
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			param.IsArray = true
			p.nextToken() // skip (
			if p.curTokenIs(lexer.TOKEN_RPAREN) {
				p.nextToken() // skip )
			}
		}

		if p.curTokenIs(lexer.TOKEN_AS) {
			p.nextToken()
			param.Type = strings.ToUpper(p.curToken.Literal)
			p.nextToken()
		}

		params = append(params, param)

		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
	}

	return params
}

func (p *Parser) parseDefStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DEF

	if !p.curTokenIs(lexer.TOKEN_FN) && !p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
		p.addError("expected FN after DEF")
		p.skipToEndOfLine()
		return nil
	}

	// Handle DEF SEG [= segment]
	if strings.ToUpper(p.curToken.Literal) == "SEG" {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_EQ) {
			p.nextToken() // skip =
			p.parseExpression(PREC_LOWEST) // consume but discard segment address
		}
		return &ast.RemStatement{BasePos: pos, Text: "DEF SEG"}
	}

	// DEF FN
	if p.curTokenIs(lexer.TOKEN_FN) {
		p.nextToken() // skip FN
	}

	name := p.curToken.Literal
	p.nextToken()

	params := p.parseParameterList()

	// Single-line form: DEF FNname(x) = expression
	if p.curTokenIs(lexer.TOKEN_EQ) {
		p.nextToken() // skip =
		expr := p.parseExpression(PREC_LOWEST)
		return &ast.DefFnDeclaration{BasePos: pos, Name: name, Params: params, SingleLineExpr: expr}
	}

	// Multi-line form
	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}
	body := p.parseBlockUntil(lexer.TOKEN_END)
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_DEF) {
			p.nextToken()
		}
	}

	return &ast.DefFnDeclaration{BasePos: pos, Name: name, Params: params, Body: body}
}

func (p *Parser) parseDataStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DATA

	var values []ast.Expression
	for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) {
		val := p.parseDataValue()
		if val != nil {
			values = append(values, val)
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return &ast.DataStatement{BasePos: pos, Values: values}
}

// parseDataValue reads a single DATA item. Quoted strings and numeric literals
// are parsed normally. Unquoted strings (identifiers, operators, etc.) are
// collected as a raw string value up to the next comma or end of line.
func (p *Parser) parseDataValue() ast.Expression {
	valPos := p.curPos()
	switch p.curToken.Type {
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_INTEGER, lexer.TOKEN_LONG, lexer.TOKEN_SINGLE, lexer.TOKEN_DOUBLE:
		return p.parseNumberLiteral()
	case lexer.TOKEN_MINUS:
		// Negative numeric literal
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_INTEGER) || p.curTokenIs(lexer.TOKEN_LONG) ||
			p.curTokenIs(lexer.TOKEN_SINGLE) || p.curTokenIs(lexer.TOKEN_DOUBLE) {
			num := p.parseNumberLiteral()
			if nl, ok := num.(*ast.NumberLiteral); ok {
				nl.Value = -nl.Value
				return nl
			}
		}
		return &ast.StringLiteral{BasePos: valPos, Value: "-"}
	default:
		// Unquoted DATA value: collect all tokens up to comma/EOL as a raw string
		var sb strings.Builder
		for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) &&
			!p.curTokenIs(lexer.TOKEN_COMMA) {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(p.curToken.Literal)
			p.nextToken()
		}
		raw := strings.TrimSpace(sb.String())
		if raw == "" {
			return nil
		}
		return &ast.StringLiteral{BasePos: valPos, Value: raw}
	}
}

func (p *Parser) parseReadStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip READ

	vars := p.parseExpressionList()
	return &ast.ReadStatement{BasePos: pos, Variables: vars}
}

func (p *Parser) parseRestoreStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip RESTORE

	target := ""
	if p.curTokenIs(lexer.TOKEN_IDENTIFIER) || p.curTokenIs(lexer.TOKEN_INTEGER) {
		target = p.curToken.Literal
		p.nextToken()
	}

	return &ast.RestoreStatement{BasePos: pos, Target: target}
}

func (p *Parser) parseSwapStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SWAP

	var1 := p.parseExpression(PREC_LOWEST)
	if !p.curTokenIs(lexer.TOKEN_COMMA) {
		p.addError("expected comma in SWAP")
	} else {
		p.nextToken()
	}
	var2 := p.parseExpression(PREC_LOWEST)

	return &ast.SwapStatement{BasePos: pos, Var1: var1, Var2: var2}
}

func (p *Parser) parseIncrDecrStatement(isIncr bool) ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip INCR/DECR

	variable := p.parseExpression(PREC_LOWEST)
	var amount ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		amount = p.parseExpression(PREC_LOWEST)
	}

	if isIncr {
		return &ast.IncrStatement{BasePos: pos, Variable: variable, Amount: amount}
	}
	return &ast.DecrStatement{BasePos: pos, Variable: variable, Amount: amount}
}

func (p *Parser) parseEraseStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip ERASE

	var names []string
	for {
		names = append(names, p.curToken.Literal)
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken()
	}

	return &ast.EraseStatement{BasePos: pos, Names: names}
}

func (p *Parser) parseConstStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip CONST
	name := p.curToken.Literal
	p.nextToken() // skip name
	if p.curTokenIs(lexer.TOKEN_EQ) {
		p.nextToken() // skip =
	}
	val := p.parseExpression(PREC_LOWEST)
	return &ast.ConstStatement{BasePos: pos, Name: name, Value: val}
}
