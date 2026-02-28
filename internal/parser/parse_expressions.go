package parser

import (
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// ---------------------------------------------------------------------------
// Expression parsing — Pratt / top-down operator precedence (TDOP)
// ---------------------------------------------------------------------------
//
// Classic recursive descent builds one parsing function per precedence level
// (parseMulDiv calls parseUnary which calls parseAtom, etc.). That works, but
// adding a new operator requires restructuring several functions.
//
// The Pratt technique (invented by Vaughan Pratt, 1973) is cleaner:
//
//  1. Assign each token type a *binding power* (precedence number).
//  2. A single parseExpression(minPrec) loop keeps consuming infix operators
//     as long as the next operator's precedence is higher than minPrec.
//  3. Prefix handlers (numbers, identifiers, unary minus) are called once at
//     the start; infix handlers (binary +, -, *, …) chain onto the left side.
//
// Example: parsing "2 + 3 * 4"
//   - parseExpression(PREC_LOWEST) calls parsePrefixExpression → NumberLiteral(2)
//   - curToken is '+' (PREC_ADD=8 > PREC_LOWEST=1) → enter loop
//   - parseInfixExpression calls parseExpression(PREC_ADD) for the right side
//     - parsePrefixExpression → NumberLiteral(3)
//     - curToken is '*' (PREC_MUL=11 > PREC_ADD=8) → enter inner loop
//     - parseInfixExpression calls parseExpression(PREC_MUL) → NumberLiteral(4)
//     - returns BinaryExpr(3 * 4)
//   - returns BinaryExpr(2 + (3*4))  ← correct precedence, no extra functions!

// Precedence levels for Turbo BASIC operators (higher = tighter binding).
//
// Tutorial note — how precedence levels work in a Pratt parser
//
// Each constant is just an integer. The relative ordering is what matters:
// a higher number means "bind more tightly" (evaluated first). The Pratt
// parseExpression loop keeps consuming infix operators as long as:
//
//   current operator's precedence > the floor passed to parseExpression
//
// Example walkthrough for "2 + 3 * 4":
//
//   parseExpression(PREC_LOWEST=1) called
//     prefix → NumberLiteral(2), curToken now '+'
//     '+' has PREC_ADD=8  > 1 → enter loop
//       parseInfixExpression(left=2)
//         op='+', prec=PREC_ADD=8
//         calls parseExpression(PREC_ADD=8) for right side
//           prefix → NumberLiteral(3), curToken now '*'
//           '*' has PREC_MUL=11 > 8 → enter inner loop
//             parseInfixExpression(left=3)
//               op='*', calls parseExpression(PREC_MUL=11)
//                 prefix → NumberLiteral(4), curToken EOL
//                 EOL's precedence = PREC_LOWEST=1, not > 11 → exit
//               returns NumberLiteral(4)
//             returns BinaryExpr(3 * 4)
//           '*' consumed, curToken EOL, PREC_LOWEST not > 8 → exit
//           returns BinaryExpr(3 * 4)
//       returns BinaryExpr(2 + (3*4))   ← correct!
//
// PREC_LOWEST (1) is the "no preference" floor: any binary operator has
// higher precedence, so the loop will always run at least once when there
// is an operator. PREC_UNARY (12) is passed when parsing the right side of
// a unary operator so that subsequent unary chains associate correctly.
const (
	_ int = iota
	PREC_LOWEST // 1 — floor value: always less than any real operator
	PREC_IMP    // 2 — IMP  (logical implication — lowest binary operator)
	PREC_EQV    // 3 — EQV  (logical equivalence)
	PREC_XOR    // 4 — XOR  (bitwise/logical exclusive or)
	PREC_OR     // 5 — OR   (bitwise/logical or)
	PREC_AND    // 6 — AND  (bitwise/logical and)
	PREC_NOT    // 7 — NOT  (unary, but placed here so AND/OR are below it)
	PREC_REL    // 8 — = <> < > <= >=  (relational comparisons)
	PREC_ADD    // 9 — + -  (additive)
	PREC_MOD    // 10 — MOD (modulo — between additive and multiplicative)
	PREC_IDIV   // 11 — \   (integer division — tighter than MOD)
	PREC_MUL    // 12 — * / (multiplicative — tighter than additive)
	PREC_UNARY  // 13 — unary - (prefix minus; used as the floor when parsing operands of unary ops)
	PREC_POWER  // 14 — ^   (exponentiation — tightest binary operator; right-associative)
)

// tokenPrecedence returns the binding power of an infix operator token.
//
// This is the Pratt parser's lookup table. A token not listed here returns
// PREC_LOWEST, which effectively acts as "this token is not a binary operator"
// and causes the parseExpression loop to stop.
//
// Note that NOT is also listed here even though it is a prefix (unary) operator.
// In BASIC "NOT" can appear in a chain like "x AND NOT y" where the parser
// needs to know that NOT binds more tightly than AND. The unary handling itself
// is done in parsePrefixExpression; this table controls what happens when NOT
// appears to the right of another operator.
func tokenPrecedence(t lexer.TokenType) int {
	switch t {
	case lexer.TOKEN_IMP:
		return PREC_IMP
	case lexer.TOKEN_EQV:
		return PREC_EQV
	case lexer.TOKEN_XOR:
		return PREC_XOR
	case lexer.TOKEN_OR:
		return PREC_OR
	case lexer.TOKEN_AND:
		return PREC_AND
	case lexer.TOKEN_EQ, lexer.TOKEN_NE, lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LE, lexer.TOKEN_GE:
		return PREC_REL
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS:
		return PREC_ADD
	case lexer.TOKEN_MOD:
		return PREC_MOD
	case lexer.TOKEN_BACKSLASH:
		return PREC_IDIV
	case lexer.TOKEN_STAR, lexer.TOKEN_SLASH:
		return PREC_MUL
	case lexer.TOKEN_CARET:
		return PREC_POWER
	}
	return PREC_LOWEST
}

// parseExpression is the heart of the Pratt parser.
// It parses an expression whose root operator has precedence > the given floor.
//
// The loop invariant: curToken is always the *next* operator (or end-of-expr).
// Each iteration "absorbs" one infix operator and its right operand, building
// a left-leaning expression tree that respects operator precedence.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	left := p.parsePrefixExpression() // Parse the leftmost primary/unary
	if left == nil {
		return nil
	}

	// Keep consuming infix operators as long as they bind tighter than our floor.
	// EOL and COLON are statement terminators — they always stop expression parsing.
	for p.curToken.Type != lexer.TOKEN_EOF &&
		p.curToken.Type != lexer.TOKEN_EOL &&
		p.curToken.Type != lexer.TOKEN_COLON &&
		precedence < tokenPrecedence(p.curToken.Type) {

		left = p.parseInfixExpression(left)
		if left == nil {
			return nil
		}
	}

	return left
}

// parsePrefixExpression handles the *start* of an expression — either a primary
// value (number, string, identifier) or a prefix/unary operator (-, NOT).
// In Pratt terminology, these are the "null-denotation" (nud) handlers.
func (p *Parser) parsePrefixExpression() ast.Expression {
	switch p.curToken.Type {
	case lexer.TOKEN_INTEGER, lexer.TOKEN_LONG, lexer.TOKEN_SINGLE, lexer.TOKEN_DOUBLE:
		return p.parseNumberLiteral()
	case lexer.TOKEN_HEX, lexer.TOKEN_OCTAL, lexer.TOKEN_BINARY_LIT:
		return p.parseRadixLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_IDENTIFIER:
		return p.parseIdentifierExpression()
	case lexer.TOKEN_LPAREN:
		return p.parseGroupExpression()
	case lexer.TOKEN_MINUS:
		return p.parseUnaryExpression()
	case lexer.TOKEN_NOT:
		return p.parseUnaryExpression()
	// Built-in functions that look like keywords
	case lexer.TOKEN_LEN_KW, lexer.TOKEN_EOF_KW, lexer.TOKEN_TAB, lexer.TOKEN_SPC,
		lexer.TOKEN_PEEK, lexer.TOKEN_INP, lexer.TOKEN_TIMER, lexer.TOKEN_INSTAT,
		lexer.TOKEN_LBOUND, lexer.TOKEN_UBOUND, lexer.TOKEN_SCREEN,
		lexer.TOKEN_VARPTR, lexer.TOKEN_VARSEG:
		return p.parseBuiltinFunction()
	// FN name(args) call in expression context
	case lexer.TOKEN_FN:
		return p.parseFnCallExpression()
	default:
		p.addError("unexpected token in expression: %s (%q)", lexer.TokenName(p.curToken.Type), p.curToken.Literal)
		p.nextToken()
		return nil
	}
}

// parseInfixExpression handles binary operators ("+", "-", "AND", …).
// In Pratt terminology these are the "left-denotation" (led) handlers.
// The left operand has already been parsed and is passed in as `left`.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	pos := p.curPos()
	operator := p.curToken.Literal
	prec := tokenPrecedence(p.curToken.Type)

	// Right-associative for exponentiation
	if p.curToken.Type == lexer.TOKEN_CARET {
		p.nextToken()
		right := p.parseExpression(prec - 1)
		return &ast.BinaryExpr{BasePos: pos, Left: left, Operator: operator, Right: right}
	}

	p.nextToken()
	right := p.parseExpression(prec)
	return &ast.BinaryExpr{BasePos: pos, Left: left, Operator: operator, Right: right}
}

// parseNumberLiteral parses an integer, long, single, or double literal.
func (p *Parser) parseNumberLiteral() ast.Expression {
	pos := p.curPos()
	lit := p.curToken.Literal
	var numType int
	switch p.curToken.Type {
	case lexer.TOKEN_INTEGER:
		numType = ast.NumInt
	case lexer.TOKEN_LONG:
		numType = ast.NumLong
	case lexer.TOKEN_SINGLE:
		numType = ast.NumSingle
	case lexer.TOKEN_DOUBLE:
		numType = ast.NumDouble
	}

	val, err := strconv.ParseFloat(strings.Replace(lit, "D", "E", 1), 64)
	if err != nil {
		p.addError("could not parse number: %q", lit)
		p.nextToken()
		return nil
	}

	p.nextToken()
	return &ast.NumberLiteral{BasePos: pos, Value: val, OriginalText: lit, NumType: numType}
}

// parseRadixLiteral parses &H, &O, or &B literals.
func (p *Parser) parseRadixLiteral() ast.Expression {
	pos := p.curPos()
	lit := p.curToken.Literal
	var base int
	switch p.curToken.Type {
	case lexer.TOKEN_HEX:
		base = 16
	case lexer.TOKEN_OCTAL:
		base = 8
	case lexer.TOKEN_BINARY_LIT:
		base = 2
	}
	val, err := strconv.ParseInt(lit, base, 64)
	if err != nil {
		p.addError("could not parse radix literal: %q", lit)
		p.nextToken()
		return nil
	}
	p.nextToken()
	return &ast.NumberLiteral{BasePos: pos, Value: float64(val), OriginalText: lit, NumType: ast.NumInt}
}

// parseStringLiteral parses a string literal.
func (p *Parser) parseStringLiteral() ast.Expression {
	pos := p.curPos()
	val := p.curToken.Literal
	p.nextToken()
	return &ast.StringLiteral{BasePos: pos, Value: val}
}

// parseIdentifierExpression handles the ambiguity between variable references,
// array accesses, and function calls — all of which start with an identifier:
//
//	x           → Identifier (variable read)
//	arr(i)      → ArrayAccess (array element) or FunctionCall (user/built-in fn)
//	arr(i).fld  → FieldAccessExpression (TYPE struct field)
//
// BASIC uses the same syntax for array indexing and function calls (both use
// parentheses), so the distinction is deferred to semantic analysis. Here we
// return an ArrayAccess node for all identifier(...) forms; the semantic pass
// will reclassify them if the name is a known function.
//
// It also handles struct/TYPE member access (dot notation): expr.field or expr.field(idx).
func (p *Parser) parseIdentifierExpression() ast.Expression {
	pos := p.curPos()
	name := p.curToken.Literal

	// Split type suffix from name
	typeSuffix := ""
	if len(name) > 0 {
		last := name[len(name)-1]
		if last == '%' || last == '&' || last == '!' || last == '#' || last == '$' {
			typeSuffix = string(last)
			name = name[:len(name)-1]
		}
	}

	p.nextToken()

	// Check for array access or function call: name(args)
	var expr ast.Expression
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken() // skip (
		args := p.parseExpressionList()
		if !p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.addError("expected ), got %s", lexer.TokenName(p.curToken.Type))
		} else {
			p.nextToken() // skip )
		}
		// If name is uppercase and looks like a built-in, treat as function call.
		// Also treat DEF FN functions (names starting with "FN") as function calls.
		// IMPORTANT: include the typeSuffix in the lookup — many BASIC built-ins
		// end in "$" (e.g., CHR$, LEFT$, MID$) which was stripped earlier.
		upper := strings.ToUpper(name)
		fullName := upper + typeSuffix
		if isBuiltinFunction(fullName) {
			expr = &ast.FunctionCall{BasePos: pos, Name: fullName, Args: args}
		} else if isBuiltinFunction(upper) {
			expr = &ast.FunctionCall{BasePos: pos, Name: upper + typeSuffix, Args: args}
		} else if strings.HasPrefix(upper, "FN") {
			// Turbo BASIC convention: DEF FN functions are named with "FN" prefix.
			// FNFoo(args) calls the user-defined function FNFoo.
			expr = &ast.FunctionCall{BasePos: pos, Name: name + typeSuffix, Args: args}
		} else {
			// Could be array access or user function — use ArrayAccess for now
			// (semantic analysis will distinguish)
			expr = &ast.ArrayAccess{BasePos: pos, Name: name, TypeSuffix: typeSuffix, Indices: args}
		}
	} else {
		expr = &ast.Identifier{BasePos: pos, Name: name, TypeSuffix: typeSuffix}
	}

	// Handle struct/TYPE member access: expr.field or expr.field(idx).
	// The "." character is lexed as TOKEN_ILLEGAL in Turbo BASIC since BASIC
	// does not normally use dot notation — it is only valid for TYPE fields.
	for p.curToken.Type == lexer.TOKEN_ILLEGAL && p.curToken.Literal == "." {
		p.nextToken() // skip .
		if !p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
			break
		}
		fieldName := p.curToken.Literal
		// Strip type suffix from field name (e.g., XCoor%)
		if len(fieldName) > 0 {
			last := fieldName[len(fieldName)-1]
			if last == '%' || last == '&' || last == '!' || last == '#' || last == '$' {
				fieldName = fieldName[:len(fieldName)-1]
			}
		}
		p.nextToken()
		fa := &ast.FieldAccessExpression{BasePos: pos, Object: expr, Field: fieldName}
		// Check for subscript on the field: .field(idx)
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken() // skip (
			args := p.parseExpressionList()
			if p.curTokenIs(lexer.TOKEN_RPAREN) {
				p.nextToken()
			}
			expr = &ast.ArrayAccess{BasePos: pos, Name: fa.Field, Indices: args}
		} else {
			expr = fa
		}
	}

	return expr
}

// parseGroupExpression parses a parenthesized expression.
func (p *Parser) parseGroupExpression() ast.Expression {
	pos := p.curPos()
	p.nextToken() // skip (
	inner := p.parseExpression(PREC_LOWEST)
	if !p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.addError("expected ), got %s", lexer.TokenName(p.curToken.Type))
		return inner
	}
	p.nextToken() // skip )
	return &ast.GroupExpr{BasePos: pos, Inner: inner}
}

// parseUnaryExpression parses NOT or unary minus.
func (p *Parser) parseUnaryExpression() ast.Expression {
	pos := p.curPos()
	op := p.curToken.Literal
	p.nextToken()
	operand := p.parseExpression(PREC_UNARY)
	return &ast.UnaryExpr{BasePos: pos, Operator: op, Operand: operand}
}

// parseBuiltinFunction parses a keyword function call like LEN(s), EOF(n).
func (p *Parser) parseBuiltinFunction() ast.Expression {
	pos := p.curPos()
	name := strings.ToUpper(p.curToken.Literal)
	p.nextToken()

	// Some builtins don't require parens (e.g., TIMER, INSTAT)
	if !p.curTokenIs(lexer.TOKEN_LPAREN) {
		return &ast.FunctionCall{BasePos: pos, Name: name, Args: nil}
	}

	p.nextToken() // skip (
	args := p.parseExpressionList()
	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken() // skip )
	}
	return &ast.FunctionCall{BasePos: pos, Name: name, Args: args}
}

// parseExpressionList parses a comma-separated list of expressions.
func (p *Parser) parseExpressionList() []ast.Expression {
	var exprs []ast.Expression

	if p.curTokenIs(lexer.TOKEN_RPAREN) || p.curTokenIs(lexer.TOKEN_EOL) || p.curTokenIs(lexer.TOKEN_EOF) {
		return exprs
	}

	expr := p.parseExpression(PREC_LOWEST)
	if expr != nil {
		exprs = append(exprs, expr)
	}

	for p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // skip comma
		expr = p.parseExpression(PREC_LOWEST)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}

	return exprs
}

// isBuiltinFunction returns true for known built-in function names.
func isBuiltinFunction(name string) bool {
	switch name {
	case "ABS", "SGN", "INT", "FIX", "CEIL", "SQR", "EXP", "EXP2", "EXP10",
		"LOG", "LOG2", "LOG10", "SIN", "COS", "TAN", "ATN",
		"LEFT$", "RIGHT$", "MID$", "LEN", "INSTR", "ASC", "CHR$",
		"STR$", "VAL", "HEX$", "OCT$", "BIN$",
		"UCASE$", "LCASE$", "LTRIM$", "RTRIM$", "TRIM$",
		"SPACE$", "STRING$",
		"CINT", "CLNG", "CSNG", "CDBL",
		"MKI$", "MKL$", "MKS$", "MKD$",
		"CVI", "CVL", "CVS", "CVD",
		"RND", "FRE", "EOF", "LOC", "LOF", "SEEK",
		"POS", "CSRLIN", "POINT", "PMAP", "DIR$",
		"PEEK", "INP", "TIMER", "MTIMER", "INSTAT",
		"LBOUND", "UBOUND", "SCREEN", "VARPTR", "VARSEG",
		"TAB", "SPC", "INKEY$", "COMMAND$", "ENVIRON$",
		"DATE$", "TIME$", "ENDMEM", "ERDEV", "ERDEV$",
		"ERR", "ERL", "IOCTL$":
		return true
	}
	return false
}
