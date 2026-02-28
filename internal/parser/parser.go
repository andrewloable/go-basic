// Package parser implements Phase 2 of the Turbo BASIC compiler pipeline:
// syntactic analysis (parsing). It consumes the token stream produced by the
// lexer and builds an Abstract Syntax Tree (AST).
//
// # How a Parser Works
//
// A parser imposes grammatical structure on the flat token stream. Where the
// lexer answers "what words are here?", the parser answers "how do those words
// form sentences (statements) and phrases (expressions)?".
//
// This parser uses two well-known techniques:
//
//  1. RECURSIVE DESCENT for statements — each BASIC keyword gets its own
//     parseXxx() method that handles the specific grammar rule. The top-level
//     parseStatement() is a big switch on the current token type that
//     dispatches to the right handler.
//
//  2. PRATT PARSING (top-down operator precedence) for expressions — rather
//     than writing one recursive descent function per precedence level, we
//     assign each token a numeric precedence and drive a single loop. This
//     elegantly handles left-associativity, right-associativity, and mixed
//     precedence without deeply nested functions.
//
// # Two-Token Lookahead
//
// The parser always holds two tokens: curToken (current) and peekToken (next).
// Two tokens are enough for Turbo BASIC because most grammatical decisions can
// be made by looking one token ahead — e.g., seeing PRINT tells us this is a
// print statement; seeing an identifier followed by '(' means it is a function
// call or array access rather than a plain variable reference.
//
// # Error Recovery
//
// When a parse error is detected, the parser records the error via addError()
// and calls skipToEndOfLine() to discard the rest of the malformed statement.
// This "panic-mode recovery" strategy allows the parser to continue past errors
// and report multiple problems in a single pass.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// Parser holds the state needed during a single-pass parse.
// The two-token window (curToken + peekToken) is the only look-ahead needed.
type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token // token currently being examined
	peekToken lexer.Token // next token (one position ahead)
	errors    []string
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Prime the two-token window: call nextToken twice so both curToken and
	// peekToken hold real tokens before ParseProgram starts.
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns all parse errors accumulated during parsing.
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses the entire program and returns the root AST node.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		BasePos: ast.Position{Line: 1, Column: 1},
	}

	for p.curToken.Type != lexer.TOKEN_EOF {
		// Skip blank lines
		if p.curToken.Type == lexer.TOKEN_EOL {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// Consume EOL or colon (multi-statement separator)
		for p.curToken.Type == lexer.TOKEN_EOL || p.curToken.Type == lexer.TOKEN_COLON {
			p.nextToken()
		}
	}

	return program
}

// nextToken advances to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token has the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the peek token has the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek advances if the peek token matches, otherwise records an error.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError records an error about unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	p.addError("expected %s, got %s (%q)",
		lexer.TokenName(t), lexer.TokenName(p.peekToken.Type), p.peekToken.Literal)
}

// addError records a parse error with position information.
func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, fmt.Sprintf("%d:%d: %s", p.curToken.Line, p.curToken.Column, msg))
}

// curPos returns the current token's position as an AST position.
func (p *Parser) curPos() ast.Position {
	return ast.Position{Line: p.curToken.Line, Column: p.curToken.Column}
}

// skipToEndOfLine advances past tokens until EOL or EOF.
func (p *Parser) skipToEndOfLine() {
	for p.curToken.Type != lexer.TOKEN_EOL && p.curToken.Type != lexer.TOKEN_EOF &&
		p.curToken.Type != lexer.TOKEN_COLON {
		p.nextToken()
	}
}

// parseStatement is the top-level statement dispatcher. It examines the
// current token type and routes to the appropriate parseXxx() method.
//
// This is the core of recursive-descent parsing for statements: each BASIC
// keyword maps to exactly one grammar rule, and each rule has its own method.
// New statement types are added here (and nowhere else in the top-level flow).
//
// For identifiers that are not keywords (variables, user sub calls, etc.) we
// fall through to parseIdentifierStatement() which tries both assignment and
// procedure-call forms.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.TOKEN_COMMENT:
		return p.parseComment()
	case lexer.TOKEN_LABEL:
		return p.parseLabel()
	case lexer.TOKEN_PRINT, lexer.TOKEN_LPRINT:
		return p.parsePrintStatement()
	case lexer.TOKEN_LET:
		return p.parseLetStatement()
	case lexer.TOKEN_IF:
		return p.parseIfStatement()
	case lexer.TOKEN_FOR:
		return p.parseForStatement()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStatement()
	case lexer.TOKEN_DO:
		return p.parseDoLoopStatement()
	case lexer.TOKEN_SELECT:
		return p.parseSelectCaseStatement()
	case lexer.TOKEN_GOTO:
		return p.parseGotoStatement()
	case lexer.TOKEN_GOSUB:
		return p.parseGosubStatement()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStatement()
	case lexer.TOKEN_EXIT:
		return p.parseExitStatement()
	case lexer.TOKEN_END:
		return p.parseEndStatement()
	case lexer.TOKEN_STOP:
		pos := p.curPos()
		p.nextToken()
		return &ast.StopStatement{BasePos: pos}
	case lexer.TOKEN_SYSTEM:
		pos := p.curPos()
		p.nextToken()
		return &ast.SystemStatement{BasePos: pos}
	case lexer.TOKEN_DIM:
		return p.parseDimStatement(false)
	case lexer.TOKEN_REDIM:
		return p.parseDimStatement(true)
	case lexer.TOKEN_SUB:
		return p.parseSubDeclaration()
	case lexer.TOKEN_FUNCTION:
		return p.parseFunctionDeclaration()
	case lexer.TOKEN_DECLARE:
		return p.parseDeclare()
	case lexer.TOKEN_DEF:
		return p.parseDefStatement()
	case lexer.TOKEN_DATA:
		return p.parseDataStatement()
	case lexer.TOKEN_READ:
		return p.parseReadStatement()
	case lexer.TOKEN_RESTORE:
		return p.parseRestoreStatement()
	case lexer.TOKEN_SWAP:
		return p.parseSwapStatement()
	case lexer.TOKEN_INCR:
		return p.parseIncrDecrStatement(true)
	case lexer.TOKEN_DECR:
		return p.parseIncrDecrStatement(false)
	case lexer.TOKEN_ERASE:
		return p.parseEraseStatement()
	case lexer.TOKEN_ON:
		return p.parseOnStatement()
	case lexer.TOKEN_RESUME:
		return p.parseResumeStatement()
	case lexer.TOKEN_ERROR:
		return p.parseErrorStatement()
	case lexer.TOKEN_OPEN:
		return p.parseOpenStatement()
	case lexer.TOKEN_CLOSE:
		return p.parseCloseStatement()
	case lexer.TOKEN_INPUT:
		return p.parseInputStatement()
	case lexer.TOKEN_LINE:
		return p.parseLineStatement()
	case lexer.TOKEN_WRITE:
		return p.parseWriteStatement()
	case lexer.TOKEN_CLS:
		return p.parseClsStatement()
	case lexer.TOKEN_CLEAR:
		pos := p.curPos()
		p.nextToken()
		// CLEAR may have optional stack/segment arguments — skip them
		p.skipToEndOfLine()
		return &ast.ClearStatement{BasePos: pos}
	case lexer.TOKEN_TYPE:
		return p.parseTypeBlock()
	case lexer.TOKEN_CONST:
		return p.parseConstStatement()
	case lexer.TOKEN_FN:
		return p.parseFnAssign()
	case lexer.TOKEN_DOLLAR, lexer.TOKEN_PERCENT,
		lexer.TOKEN_META_DYNAMIC, lexer.TOKEN_META_STATIC, lexer.TOKEN_META_INCLUDE,
		lexer.TOKEN_META_IF, lexer.TOKEN_META_ELSEIF, lexer.TOKEN_META_ELSE,
		lexer.TOKEN_META_ENDIF, lexer.TOKEN_META_COM, lexer.TOKEN_META_SOUND,
		lexer.TOKEN_META_STACK, lexer.TOKEN_META_SEGMENT, lexer.TOKEN_META_INLINE,
		lexer.TOKEN_META_EVENT:
		// Metacompiler directives ($IF, $DYNAMIC, %DEFINE, etc.) — skip line.
		// These are compile-time conditionals with no runtime equivalent.
		p.skipToEndOfLine()
		return nil
	case lexer.TOKEN_LOCATE:
		return p.parseLocateStatement()
	case lexer.TOKEN_COLOR:
		return p.parseColorStatement()
	case lexer.TOKEN_BEEP:
		pos := p.curPos()
		p.nextToken()
		return &ast.BeepStatement{BasePos: pos}
	case lexer.TOKEN_SOUND:
		return p.parseSoundStatement()
	case lexer.TOKEN_PLAY:
		return p.parsePlayDrawStatement("PLAY")
	case lexer.TOKEN_DRAW:
		return p.parsePlayDrawStatement("DRAW")
	case lexer.TOKEN_SCREEN:
		return p.parseScreenStatement()
	case lexer.TOKEN_OPTION:
		return p.parseOptionStatement()
	case lexer.TOKEN_DEFINT, lexer.TOKEN_DEFLNG, lexer.TOKEN_DEFSNG, lexer.TOKEN_DEFDBL, lexer.TOKEN_DEFSTR:
		return p.parseDefTypeStatement()
	case lexer.TOKEN_SHARED, lexer.TOKEN_LOCAL, lexer.TOKEN_STATIC, lexer.TOKEN_COMMON:
		return p.parseScopeStatement()
	case lexer.TOKEN_CALL:
		return p.parseCallStatement()
	case lexer.TOKEN_RANDOMIZE:
		return p.parseRandomizeStatement()
	case lexer.TOKEN_PSET, lexer.TOKEN_PRESET:
		return p.parsePsetStatement()
	case lexer.TOKEN_CIRCLE:
		return p.parseCircleStatement()
	case lexer.TOKEN_PAINT:
		return p.parsePaintStatement()
	case lexer.TOKEN_GET:
		return p.parseGetStatement()
	case lexer.TOKEN_PUT:
		return p.parsePutStatement()
	case lexer.TOKEN_SEEK:
		return p.parseSeekStatement()
	case lexer.TOKEN_FIELD:
		return p.parseFieldStatement()
	case lexer.TOKEN_LSET, lexer.TOKEN_RSET:
		return p.parseLsetRsetStatement()
	case lexer.TOKEN_KILL:
		return p.parseSingleExprStatement("KILL")
	case lexer.TOKEN_CHDIR:
		return p.parseSingleExprStatement("CHDIR")
	case lexer.TOKEN_MKDIR:
		return p.parseSingleExprStatement("MKDIR")
	case lexer.TOKEN_RMDIR:
		return p.parseSingleExprStatement("RMDIR")
	case lexer.TOKEN_NAME:
		return p.parseNameStatement()
	case lexer.TOKEN_TRON:
		pos := p.curPos()
		p.nextToken()
		return &ast.RemStatement{BasePos: pos, Text: "TRON"}
	case lexer.TOKEN_TROFF:
		pos := p.curPos()
		p.nextToken()
		return &ast.RemStatement{BasePos: pos, Text: "TROFF"}
	case lexer.TOKEN_POKE:
		return p.parsePokeStatement()
	case lexer.TOKEN_WIDTH:
		pos := p.curPos()
		p.nextToken() // skip WIDTH
		// WIDTH col [, row] — set console/screen width; consumed but emitted as no-op
		p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.parseExpression(PREC_LOWEST)
		}
		return &ast.RemStatement{BasePos: pos, Text: "WIDTH"}
	case lexer.TOKEN_DELAY:
		return p.parseDelayStatement()
	case lexer.TOKEN_VIEW:
		return p.parseViewStatement()
	case lexer.TOKEN_WINDOW:
		return p.parseWindowStatement()
	case lexer.TOKEN_PALETTE:
		return p.parsePaletteStatement()
	case lexer.TOKEN_IDENTIFIER, lexer.TOKEN_INTEGER:
		return p.parseIdentifierStatement()
	default:
		p.addError("unexpected token: %s (%q)", lexer.TokenName(p.curToken.Type), p.curToken.Literal)
		p.nextToken()
		return nil
	}
}

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
const (
	_ int = iota
	PREC_LOWEST
	PREC_IMP    // IMP  (logical implication — lowest operator)
	PREC_EQV    // EQV  (logical equivalence)
	PREC_XOR    // XOR  (bitwise/logical exclusive or)
	PREC_OR     // OR   (bitwise/logical or)
	PREC_AND    // AND  (bitwise/logical and)
	PREC_NOT    // NOT  (unary, but binds tighter than AND)
	PREC_REL    // = <> < > <= >=  (relational comparisons)
	PREC_ADD    // + -  (additive)
	PREC_MOD    // MOD  (modulo)
	PREC_IDIV   // \    (integer division)
	PREC_MUL    // * /  (multiplicative — tighter than additive)
	PREC_UNARY  // unary - (prefix minus)
	PREC_POWER  // ^    (exponentiation — tightest binary operator)
)

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
//   x           → Identifier (variable read)
//   arr(i)      → ArrayAccess (array element) or FunctionCall (user/built-in fn)
//   arr(i).fld  → FieldAccessExpression (TYPE struct field)
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

// ---------------------------------------------------------------------------
// Statement parsers
// ---------------------------------------------------------------------------

func (p *Parser) parseComment() ast.Statement {
	pos := p.curPos()
	text := p.curToken.Literal
	p.nextToken()
	return &ast.RemStatement{BasePos: pos, Text: text}
}

func (p *Parser) parseLabel() ast.Statement {
	pos := p.curPos()
	name := p.curToken.Literal
	p.nextToken()
	return &ast.LabelStatement{BasePos: pos, Name: name}
}

func (p *Parser) parsePrintStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip PRINT/LPRINT

	// PRINT #n, expr ... — file output
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken() // skip #
		fileNum := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		exprs := p.parsePrintExprList()
		hasSep := false
		if len(exprs.Separators) > 0 && len(exprs.Separators) >= len(exprs.Expressions) {
			hasSep = true
		}
		return &ast.FilePrintStatement{
			BasePos:        pos,
			FileNum:        fileNum,
			Expressions:    exprs.Expressions,
			Separators:     exprs.Separators,
			HasTrailingSep: hasSep,
		}
	}

	stmt := &ast.PrintStatement{BasePos: pos}

	// Check for PRINT USING
	if p.curTokenIs(lexer.TOKEN_USING) {
		p.nextToken() // skip USING
		stmt.Format = p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken() // skip ; after format string
		}
	}

	exprs := p.parsePrintExprList()
	stmt.Expressions = exprs.Expressions
	stmt.Separators = exprs.Separators

	// Check for trailing separator (suppress newline)
	if len(stmt.Separators) > 0 && len(stmt.Separators) >= len(stmt.Expressions) {
		stmt.HasTrailingSep = true
	}

	return stmt
}

// printExprResult holds the parsed expressions and separators for a PRINT statement.
type printExprResult struct {
	Expressions []ast.Expression
	Separators  []string
}

// parsePrintExprList parses a PRINT expression list, handling `,`, `;`, and implicit
// concatenation (adjacent expressions without a separator, treated as `;`).
func (p *Parser) parsePrintExprList() printExprResult {
	var r printExprResult
	for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) &&
		!p.curTokenIs(lexer.TOKEN_COLON) && !p.curTokenIs(lexer.TOKEN_COMMENT) {
		expr := p.parseExpression(PREC_LOWEST)
		if expr != nil {
			r.Expressions = append(r.Expressions, expr)
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			r.Separators = append(r.Separators, ",")
			p.nextToken()
		} else if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
			r.Separators = append(r.Separators, ";")
			p.nextToken()
		} else if canStartExpression(p.curToken.Type) {
			// Implicit concatenation: adjacent expressions treated as `;`
			r.Separators = append(r.Separators, ";")
		} else {
			break
		}
	}
	return r
}

// canStartExpression returns true if the token type can begin an expression.
func canStartExpression(t lexer.TokenType) bool {
	switch t {
	case lexer.TOKEN_INTEGER, lexer.TOKEN_LONG, lexer.TOKEN_SINGLE, lexer.TOKEN_DOUBLE,
		lexer.TOKEN_STRING, lexer.TOKEN_IDENTIFIER, lexer.TOKEN_LPAREN,
		lexer.TOKEN_MINUS, lexer.TOKEN_NOT,
		lexer.TOKEN_HEX, lexer.TOKEN_OCTAL, lexer.TOKEN_BINARY_LIT,
		lexer.TOKEN_LEN_KW, lexer.TOKEN_EOF_KW, lexer.TOKEN_TAB, lexer.TOKEN_SPC,
		lexer.TOKEN_PEEK, lexer.TOKEN_INP, lexer.TOKEN_TIMER, lexer.TOKEN_INSTAT,
		lexer.TOKEN_LBOUND, lexer.TOKEN_UBOUND, lexer.TOKEN_SCREEN,
		lexer.TOKEN_VARPTR, lexer.TOKEN_VARSEG, lexer.TOKEN_FN:
		return true
	}
	return false
}

func (p *Parser) parseLetStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip LET

	return p.parseAssignment(pos)
}

func (p *Parser) parseAssignment(pos ast.Position) ast.Statement {
	if !p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
		p.addError("expected identifier in assignment, got %s", lexer.TokenName(p.curToken.Type))
		p.skipToEndOfLine()
		return nil
	}

	nameExpr := p.parseIdentifierExpression()

	if !p.curTokenIs(lexer.TOKEN_EQ) {
		p.addError("expected = in assignment, got %s", lexer.TokenName(p.curToken.Type))
		p.skipToEndOfLine()
		return nil
	}
	p.nextToken() // skip =

	value := p.parseExpression(PREC_LOWEST)

	// Check if it's an array assignment
	if arr, ok := nameExpr.(*ast.ArrayAccess); ok {
		return &ast.ArrayAssignment{BasePos: pos, Array: arr, Value: value}
	}
	if ident, ok := nameExpr.(*ast.Identifier); ok {
		return &ast.LetStatement{BasePos: pos, Name: ident, Value: value}
	}
	// Struct/TYPE field assignment: expr.field = value
	if fa, ok := nameExpr.(*ast.FieldAccessExpression); ok {
		return &ast.FieldAssignStatement{BasePos: pos, Object: fa.Object, Field: fa.Field, Value: value}
	}

	p.addError("invalid assignment target")
	return nil
}

func (p *Parser) parseIfStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip IF

	condition := p.parseExpression(PREC_LOWEST)

	if !p.curTokenIs(lexer.TOKEN_THEN) {
		p.addError("expected THEN, got %s", lexer.TokenName(p.curToken.Type))
		p.skipToEndOfLine()
		return nil
	}
	p.nextToken() // skip THEN

	stmt := &ast.IfStatement{BasePos: pos, Condition: condition}

	// Determine if single-line or block IF
	if p.curTokenIs(lexer.TOKEN_EOL) || p.curTokenIs(lexer.TOKEN_EOF) {
		// Block IF
		p.nextToken() // skip EOL
		stmt.ThenBlock = p.parseBlockUntil(lexer.TOKEN_ELSE, lexer.TOKEN_ELSEIF, lexer.TOKEN_END)

		// Parse ELSEIF clauses
		for p.curTokenIs(lexer.TOKEN_ELSEIF) {
			elseIfPos := p.curPos()
			p.nextToken() // skip ELSEIF
			elseIfCond := p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_THEN) {
				p.nextToken() // skip THEN
			}
			if p.curTokenIs(lexer.TOKEN_EOL) {
				p.nextToken()
			}
			elseIfBody := p.parseBlockUntil(lexer.TOKEN_ELSE, lexer.TOKEN_ELSEIF, lexer.TOKEN_END)
			stmt.ElseIfClauses = append(stmt.ElseIfClauses, ast.ElseIfClause{
				BasePos: elseIfPos, Condition: elseIfCond, Body: elseIfBody,
			})
		}

		// Parse ELSE block
		if p.curTokenIs(lexer.TOKEN_ELSE) {
			p.nextToken() // skip ELSE
			if p.curTokenIs(lexer.TOKEN_EOL) {
				p.nextToken()
			}
			stmt.ElseBlock = p.parseBlockUntil(lexer.TOKEN_END)
		}

		// Expect END IF
		if p.curTokenIs(lexer.TOKEN_END) {
			p.nextToken() // skip END
			if p.curTokenIs(lexer.TOKEN_IF) {
				p.nextToken() // skip IF
			}
		}
	} else {
		// Single-line IF
		stmt.IsSingleLine = true
		thenStmt := p.parseStatement()
		if thenStmt != nil {
			stmt.ThenBlock = []ast.Statement{thenStmt}
		}

		if p.curTokenIs(lexer.TOKEN_ELSE) {
			p.nextToken() // skip ELSE
			elseStmt := p.parseStatement()
			if elseStmt != nil {
				stmt.ElseBlock = []ast.Statement{elseStmt}
			}
		}
	}

	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip FOR

	if !p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
		p.addError("expected identifier after FOR")
		p.skipToEndOfLine()
		return nil
	}

	counter := p.parseIdentifierAsIdent()
	if !p.curTokenIs(lexer.TOKEN_EQ) {
		p.addError("expected = after FOR variable")
		p.skipToEndOfLine()
		return nil
	}
	p.nextToken() // skip =

	start := p.parseExpression(PREC_LOWEST)

	if !p.curTokenIs(lexer.TOKEN_TO) {
		p.addError("expected TO in FOR statement")
		p.skipToEndOfLine()
		return nil
	}
	p.nextToken() // skip TO

	end := p.parseExpression(PREC_LOWEST)

	var step ast.Expression
	if p.curTokenIs(lexer.TOKEN_STEP) {
		p.nextToken() // skip STEP
		step = p.parseExpression(PREC_LOWEST)
	}

	// Skip EOL
	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	body := p.parseBlockUntil(lexer.TOKEN_NEXT)

	// Consume NEXT [counter]
	if p.curTokenIs(lexer.TOKEN_NEXT) {
		p.nextToken() // skip NEXT
		if p.curTokenIs(lexer.TOKEN_IDENTIFIER) {
			p.nextToken() // skip optional counter name
		}
	}

	return &ast.ForStatement{
		BasePos: pos, Counter: counter, Start: start, End: end, Step: step, Body: body,
	}
}

func (p *Parser) parseWhileStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip WHILE

	condition := p.parseExpression(PREC_LOWEST)

	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	body := p.parseBlockUntil(lexer.TOKEN_WEND)

	if p.curTokenIs(lexer.TOKEN_WEND) {
		p.nextToken() // skip WEND
	}

	return &ast.WhileStatement{BasePos: pos, Condition: condition, Body: body}
}

func (p *Parser) parseDoLoopStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DO

	stmt := &ast.DoLoopStatement{BasePos: pos}

	// Check for DO WHILE / DO UNTIL (top-tested)
	if p.curTokenIs(lexer.TOKEN_WHILE) {
		stmt.TestAtTop = true
		p.nextToken()
		stmt.Condition = p.parseExpression(PREC_LOWEST)
	} else if p.curTokenIs(lexer.TOKEN_UNTIL) {
		stmt.TestAtTop = true
		stmt.IsUntil = true
		p.nextToken()
		stmt.Condition = p.parseExpression(PREC_LOWEST)
	}

	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	stmt.Body = p.parseBlockUntil(lexer.TOKEN_LOOP)

	if p.curTokenIs(lexer.TOKEN_LOOP) {
		p.nextToken() // skip LOOP

		// Check for LOOP WHILE / LOOP UNTIL (bottom-tested)
		if !stmt.TestAtTop {
			if p.curTokenIs(lexer.TOKEN_WHILE) {
				p.nextToken()
				stmt.Condition = p.parseExpression(PREC_LOWEST)
			} else if p.curTokenIs(lexer.TOKEN_UNTIL) {
				stmt.IsUntil = true
				p.nextToken()
				stmt.Condition = p.parseExpression(PREC_LOWEST)
			}
		}
	}

	return stmt
}

func (p *Parser) parseSelectCaseStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SELECT

	if p.curTokenIs(lexer.TOKEN_CASE) {
		p.nextToken() // skip CASE
	}

	testExpr := p.parseExpression(PREC_LOWEST)

	if p.curTokenIs(lexer.TOKEN_EOL) {
		p.nextToken()
	}

	stmt := &ast.SelectCaseStatement{BasePos: pos, TestExpr: testExpr}

	for p.curTokenIs(lexer.TOKEN_CASE) {
		p.nextToken() // skip CASE

		if p.curTokenIs(lexer.TOKEN_ELSE) {
			p.nextToken() // skip ELSE
			if p.curTokenIs(lexer.TOKEN_EOL) {
				p.nextToken()
			}
			stmt.ElseBlock = p.parseBlockUntil(lexer.TOKEN_CASE, lexer.TOKEN_END)
			continue
		}

		clause := ast.CaseClause{BasePos: p.curPos()}
		clause.Values = p.parseCaseValues()

		if p.curTokenIs(lexer.TOKEN_EOL) {
			p.nextToken()
		}

		clause.Body = p.parseBlockUntil(lexer.TOKEN_CASE, lexer.TOKEN_END)
		stmt.Cases = append(stmt.Cases, clause)
	}

	// END SELECT
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_SELECT) {
			p.nextToken()
		}
	}

	return stmt
}

func (p *Parser) parseCaseValues() []ast.CaseValue {
	var values []ast.CaseValue

	for {
		cv := ast.CaseValue{BasePos: p.curPos()}

		if p.curTokenIs(lexer.TOKEN_IS) {
			// CASE IS > value
			p.nextToken() // skip IS
			cv.IsComparison = true
			cv.Comparison = p.curToken.Literal
			p.nextToken() // skip operator
			cv.Value = p.parseExpression(PREC_LOWEST)
		} else {
			cv.Value = p.parseExpression(PREC_LOWEST)
			// Check for TO (range)
			if p.curTokenIs(lexer.TOKEN_TO) {
				cv.IsRange = true
				p.nextToken() // skip TO
				cv.EndValue = p.parseExpression(PREC_LOWEST)
			}
		}

		values = append(values, cv)

		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // skip comma
	}

	return values
}

func (p *Parser) parseGotoStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip GOTO
	target := p.curToken.Literal
	p.nextToken()
	return &ast.GotoStatement{BasePos: pos, Target: target}
}

func (p *Parser) parseGosubStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip GOSUB
	target := p.curToken.Literal
	p.nextToken()
	return &ast.GosubStatement{BasePos: pos, Target: target}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip RETURN
	return &ast.ReturnStatement{BasePos: pos}
}

func (p *Parser) parseExitStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip EXIT
	exitType := strings.ToUpper(p.curToken.Literal)
	p.nextToken()
	return &ast.ExitStatement{BasePos: pos, ExitType: exitType}
}

func (p *Parser) parseEndStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip END

	// Check for END SUB, END FUNCTION, END IF, END SELECT, END DEF
	// These are handled by their parent parsers, so bare END = program end
	if p.curTokenIs(lexer.TOKEN_EOL) || p.curTokenIs(lexer.TOKEN_EOF) || p.curTokenIs(lexer.TOKEN_COLON) {
		return &ast.EndStatement{BasePos: pos}
	}

	// Might be END IF, END SELECT, etc. — let the caller handle this
	return &ast.EndStatement{BasePos: pos}
}

func (p *Parser) parseDimStatement(isRedim bool) ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DIM/REDIM

	// Skip optional scope modifiers: SHARED, LOCAL, STATIC, COMMON, DYNAMIC
	for p.curTokenIs(lexer.TOKEN_SHARED) || p.curTokenIs(lexer.TOKEN_LOCAL) ||
		p.curTokenIs(lexer.TOKEN_STATIC) || p.curTokenIs(lexer.TOKEN_COMMON) ||
		(p.curTokenIs(lexer.TOKEN_IDENTIFIER) && strings.ToUpper(p.curToken.Literal) == "DYNAMIC") {
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
	return &ast.DimStatement{BasePos: pos, Declarations: decls}
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

func (p *Parser) parseOnStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip ON

	// ON ERROR GOTO target
	if p.curTokenIs(lexer.TOKEN_ERROR) {
		p.nextToken() // skip ERROR
		if p.curTokenIs(lexer.TOKEN_GOTO) {
			p.nextToken() // skip GOTO
			target := p.curToken.Literal
			p.nextToken()
			return &ast.OnErrorGotoStatement{BasePos: pos, Target: target}
		}
	}

	// Peek ahead: if a known event keyword (KEY, TIMER, STRIG, etc.) is followed
	// by LPAREN or directly GOSUB/GOTO with a single target, it is an event handler.
	// Otherwise treat as computed ON expr GOTO/GOSUB target1[, target2, ...].
	eventKeywords := map[string]bool{
		"KEY": true, "TIMER": true, "STRIG": true, "PLAY": true, "PEN": true,
		"COM": true, "UEVENT": true,
	}
	candidate := strings.ToUpper(p.curToken.Literal)
	if eventKeywords[candidate] {
		eventType := candidate
		p.nextToken()
		var eventParam ast.Expression
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken()
			eventParam = p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_RPAREN) {
				p.nextToken()
			}
		}
		if p.curTokenIs(lexer.TOKEN_GOSUB) {
			p.nextToken()
			target := p.curToken.Literal
			p.nextToken()
			return &ast.OnEventGosubStatement{
				BasePos: pos, EventType: eventType, EventParam: eventParam, Target: target,
			}
		}
	}

	// Computed ON expr GOTO/GOSUB target1, target2, ...
	expr := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_GOTO) {
		p.nextToken() // skip GOTO
		targets := p.parseTargetList()
		return &ast.OnComputedGotoStatement{BasePos: pos, Expr: expr, Targets: targets}
	}
	if p.curTokenIs(lexer.TOKEN_GOSUB) {
		p.nextToken() // skip GOSUB
		targets := p.parseTargetList()
		return &ast.OnComputedGosubStatement{BasePos: pos, Expr: expr, Targets: targets}
	}

	p.addError("expected GOTO or GOSUB after ON expression")
	p.skipToEndOfLine()
	return nil
}

// parseTargetList parses a comma-separated list of GOTO/GOSUB targets (labels or line numbers).
func (p *Parser) parseTargetList() []string {
	var targets []string
	for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		targets = append(targets, p.curToken.Literal)
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		} else {
			break
		}
	}
	return targets
}

func (p *Parser) parseResumeStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip RESUME

	resumeType := ""
	if p.curTokenIs(lexer.TOKEN_NEXT) {
		resumeType = "NEXT"
		p.nextToken()
	} else if p.curTokenIs(lexer.TOKEN_IDENTIFIER) || p.curTokenIs(lexer.TOKEN_INTEGER) {
		resumeType = p.curToken.Literal
		p.nextToken()
	}

	return &ast.ResumeStatement{BasePos: pos, Type: resumeType}
}

func (p *Parser) parseErrorStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip ERROR
	code := p.parseExpression(PREC_LOWEST)
	return &ast.ErrorStatement{BasePos: pos, Code: code}
}

func (p *Parser) parseOpenStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip OPEN

	filename := p.parseExpression(PREC_LOWEST)

	mode := ""
	if p.curTokenIs(lexer.TOKEN_FOR) {
		p.nextToken() // skip FOR
		mode = strings.ToUpper(p.curToken.Literal)
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOKEN_AS) {
		p.nextToken() // skip AS
	}

	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken() // skip optional #
	}

	fileNum := p.parseExpression(PREC_LOWEST)

	var recLen ast.Expression
	if p.curTokenIs(lexer.TOKEN_LEN_KW) {
		p.nextToken() // skip LEN
		if p.curTokenIs(lexer.TOKEN_EQ) {
			p.nextToken() // skip =
		}
		recLen = p.parseExpression(PREC_LOWEST)
	}

	return &ast.OpenStatement{
		BasePos: pos, Filename: filename, Mode: mode, FileNum: fileNum, RecLen: recLen,
	}
}

func (p *Parser) parseCloseStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip CLOSE

	var fileNums []ast.Expression
	for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		if p.curTokenIs(lexer.TOKEN_HASH) {
			p.nextToken() // skip #
		}
		expr := p.parseExpression(PREC_LOWEST)
		if expr != nil {
			fileNums = append(fileNums, expr)
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return &ast.CloseStatement{BasePos: pos, FileNums: fileNums}
}

func (p *Parser) parseInputStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip INPUT

	// Check for INPUT #n (file input)
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken() // skip #
		fileNum := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		vars := p.parseExpressionList()
		return &ast.FileInputStatement{BasePos: pos, FileNum: fileNum, Variables: vars}
	}

	// Regular INPUT: optional prompt string followed by `;` or `,`
	// e.g.: INPUT "Enter value: ";x  or  INPUT "Enter value: ", x  or  INPUT x
	prompt := ""
	if p.curTokenIs(lexer.TOKEN_STRING) {
		prompt = p.curToken.Literal
		p.nextToken() // consume the string literal
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) || p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken() // skip prompt separator
		}
	}
	vars := p.parseExpressionList()
	return &ast.ReadStatement{BasePos: pos, Variables: vars, IsInput: true, Prompt: prompt}
}

func (p *Parser) parseLineStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip LINE

	if p.curTokenIs(lexer.TOKEN_INPUT) {
		p.nextToken() // skip INPUT
		// LINE INPUT [#n,] [prompt;] var$
		if p.curTokenIs(lexer.TOKEN_HASH) {
			p.nextToken()
			fileNum := p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			vars := p.parseExpressionList()
			return &ast.FileInputStatement{BasePos: pos, FileNum: fileNum, Variables: vars, IsLineInput: true}
		}
		// Optional prompt string before `;` separator
		prompt := ""
		if p.curTokenIs(lexer.TOKEN_STRING) {
			prompt = p.curToken.Literal
			p.nextToken()
			if p.curTokenIs(lexer.TOKEN_SEMICOLON) || p.curTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
		}
		vars := p.parseExpressionList()
		return &ast.ReadStatement{BasePos: pos, Variables: vars, IsInput: true, Prompt: prompt, IsLineInput: true}
	}

	// LINE graphics: LINE [(x1,y1)]-(x2,y2) [, color] [, B[F]]
	stmt := &ast.LineStmt{BasePos: pos}
	// Parse coordinate pairs
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		stmt.X1 = p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		stmt.Y1 = p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
		}
	}
	if p.curTokenIs(lexer.TOKEN_MINUS) {
		p.nextToken() // skip -
	}
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		stmt.X2 = p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		stmt.Y2 = p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
		}
	}
	// Optional color
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			stmt.Color = p.parseExpression(PREC_LOWEST)
		}
	}
	// Optional B or BF
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		mode := strings.ToUpper(p.curToken.Literal)
		if mode == "B" || mode == "BF" {
			stmt.BoxMode = mode
			p.nextToken()
		}
	}

	return stmt
}

func (p *Parser) parseWriteStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip WRITE

	// WRITE #n
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
		fileNum := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		exprs := p.parseExpressionList()
		return &ast.FileWriteStatement{BasePos: pos, FileNum: fileNum, Expressions: exprs}
	}

	exprs := p.parseExpressionList()
	return &ast.PrintStatement{BasePos: pos, Expressions: exprs}
}

func (p *Parser) parseClsStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip CLS
	var mode ast.Expression
	if p.curTokenIs(lexer.TOKEN_INTEGER) {
		mode = p.parseExpression(PREC_LOWEST)
	}
	return &ast.ClsStatement{BasePos: pos, Mode: mode}
}

func (p *Parser) parseLocateStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip LOCATE
	row := p.parseExpression(PREC_LOWEST)
	var col ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		col = p.parseExpression(PREC_LOWEST)
	}
	return &ast.LocateStatement{BasePos: pos, Row: row, Col: col}
}

func (p *Parser) parseColorStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip COLOR
	fg := p.parseExpression(PREC_LOWEST)
	var bg, border ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		bg = p.parseExpression(PREC_LOWEST)
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		border = p.parseExpression(PREC_LOWEST)
	}
	return &ast.ColorStatement{BasePos: pos, Foreground: fg, Background: bg, Border: border}
}

func (p *Parser) parseSoundStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SOUND
	freq := p.parseExpression(PREC_LOWEST)
	if !p.curTokenIs(lexer.TOKEN_COMMA) {
		p.addError("expected comma in SOUND")
	} else {
		p.nextToken()
	}
	dur := p.parseExpression(PREC_LOWEST)
	return &ast.SoundStatement{BasePos: pos, Frequency: freq, Duration: dur}
}

func (p *Parser) parsePlayDrawStatement(keyword string) ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip PLAY/DRAW
	cmdStr := p.parseExpression(PREC_LOWEST)
	if keyword == "PLAY" {
		return &ast.PlayStatement{BasePos: pos, CommandString: cmdStr}
	}
	return &ast.DrawStmt{BasePos: pos, CommandString: cmdStr}
}

func (p *Parser) parseScreenStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SCREEN
	mode := p.parseExpression(PREC_LOWEST)
	var cs, ap, vp ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			cs = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			ap = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		vp = p.parseExpression(PREC_LOWEST)
	}
	return &ast.ScreenStatement{BasePos: pos, Mode: mode, ColorSwitch: cs, ActivePage: ap, VisualPage: vp}
}

func (p *Parser) parseOptionStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip OPTION
	if p.curTokenIs(lexer.TOKEN_BASE) {
		p.nextToken() // skip BASE
		val, _ := strconv.Atoi(p.curToken.Literal)
		p.nextToken()
		return &ast.OptionBaseStatement{BasePos: pos, Value: val}
	}
	p.addError("expected BASE after OPTION")
	p.skipToEndOfLine()
	return nil
}

func (p *Parser) parseDefTypeStatement() ast.Statement {
	pos := p.curPos()
	typeName := strings.ToUpper(p.curToken.Literal)
	p.nextToken()

	var ranges []ast.LetterRange
	for {
		start := strings.ToUpper(p.curToken.Literal)[0]
		p.nextToken()
		end := start
		if p.curTokenIs(lexer.TOKEN_MINUS) {
			p.nextToken()
			end = strings.ToUpper(p.curToken.Literal)[0]
			p.nextToken()
		}
		ranges = append(ranges, ast.LetterRange{Start: start, End: end})
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken()
	}

	return &ast.DefTypeStatement{BasePos: pos, LetterRanges: ranges, Type: typeName}
}

func (p *Parser) parseScopeStatement() ast.Statement {
	pos := p.curPos()
	modifier := strings.ToUpper(p.curToken.Literal)
	p.nextToken()

	var vars []string
	for {
		vars = append(vars, p.curToken.Literal)
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken()
	}

	return &ast.ScopeStatement{BasePos: pos, Modifier: modifier, Variables: vars}
}

func (p *Parser) parseCallStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip CALL

	name := p.curToken.Literal
	p.nextToken()

	var args []ast.Expression
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		args = p.parseExpressionList()
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
		}
	}

	return &ast.LetStatement{
		BasePos: pos,
		Name:    &ast.Identifier{BasePos: pos, Name: name},
		Value:   &ast.FunctionCall{BasePos: pos, Name: name, Args: args},
	}
}

func (p *Parser) parseRandomizeStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip RANDOMIZE
	var seed ast.Expression
	if !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		seed = p.parseExpression(PREC_LOWEST)
	}
	return &ast.RandomizeStatement{
		BasePos: pos,
		Seed:    seed,
	}
}

func (p *Parser) parsePsetStatement() ast.Statement {
	pos := p.curPos()
	isPreset := p.curTokenIs(lexer.TOKEN_PRESET)
	p.nextToken() // skip PSET/PRESET

	isStep := false
	if p.curTokenIs(lexer.TOKEN_STEP) {
		isStep = true
		p.nextToken()
	}

	if !p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.addError("expected ( after PSET/PRESET")
		return nil
	}
	p.nextToken()
	x := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	y := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
	}

	var color ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		color = p.parseExpression(PREC_LOWEST)
	}

	return &ast.PsetStatement{BasePos: pos, X: x, Y: y, Color: color, IsStep: isStep, IsPreset: isPreset}
}

func (p *Parser) parseCircleStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip CIRCLE

	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
	}
	x := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	y := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	radius := p.parseExpression(PREC_LOWEST)

	var color, start, end, aspect ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			color = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			start = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			end = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		aspect = p.parseExpression(PREC_LOWEST)
	}

	return &ast.CircleStmt{BasePos: pos, X: x, Y: y, Radius: radius, Color: color, Start: start, End: end, Aspect: aspect}
}

func (p *Parser) parsePaintStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip PAINT
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
	}
	x := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	y := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
	}
	var fill, border ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) {
			fill = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		border = p.parseExpression(PREC_LOWEST)
	}
	return &ast.PaintStmt{BasePos: pos, X: x, Y: y, FillColor: fill, BorderColor: border}
}

func (p *Parser) parseGetStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip GET

	// Graphics GET: GET (x1, y1)-(x2, y2), arrayVar
	// File GET:     GET [#]filenum [, recnum] [, variable]
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		// Graphics GET — consume (x1, y1)-(x2, y2), array and emit as no-op
		p.nextToken() // skip (
		p.parseExpression(PREC_LOWEST) // x1
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		p.parseExpression(PREC_LOWEST) // y1
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken() // skip )
		}
		if p.curTokenIs(lexer.TOKEN_MINUS) {
			p.nextToken() // skip - (step separator)
		}
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken() // skip (
			p.parseExpression(PREC_LOWEST) // x2
			if p.curTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			p.parseExpression(PREC_LOWEST) // y2
			if p.curTokenIs(lexer.TOKEN_RPAREN) {
				p.nextToken() // skip )
			}
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.parseExpression(PREC_LOWEST) // arrayVar
		}
		return &ast.RemStatement{BasePos: pos, Text: "GET (graphics)"}
	}

	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
	}
	fileNum := p.parseExpression(PREC_LOWEST)
	var recOrPos, variable ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) && !p.curTokenIs(lexer.TOKEN_EOL) {
			recOrPos = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		variable = p.parseExpression(PREC_LOWEST)
	}
	return &ast.GetStatement{BasePos: pos, FileNum: fileNum, RecordOrPos: recOrPos, Variable: variable}
}

func (p *Parser) parsePutStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip PUT

	// Graphics PUT: PUT (x, y), arrayVar [, mode]
	// File PUT:     PUT [#]filenum [, recnum] [, variable]
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		// Graphics PUT — parse (x, y) coordinate pair
		p.nextToken() // skip (
		x := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		y := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken() // skip )
		}
		var arrayVar, mode ast.Expression
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			arrayVar = p.parseExpression(PREC_LOWEST)
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			// Graphics mode is a keyword (PSET, PRESET, AND, OR, XOR) — not a normal
			// expression, so consume the token directly rather than calling parseExpression.
			switch p.curToken.Type {
			case lexer.TOKEN_PSET, lexer.TOKEN_PRESET, lexer.TOKEN_AND, lexer.TOKEN_OR, lexer.TOKEN_XOR:
				modeIdent := &ast.Identifier{BasePos: p.curPos(), Name: p.curToken.Literal}
				mode = modeIdent
				p.nextToken()
			default:
				mode = p.parseExpression(PREC_LOWEST)
			}
		}
		// Emit as a no-op graphics statement (not yet implemented in transpiler)
		_ = x; _ = y; _ = arrayVar; _ = mode
		return &ast.RemStatement{BasePos: pos, Text: "PUT (graphics)"}
	}

	// File PUT
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
	}
	fileNum := p.parseExpression(PREC_LOWEST)
	var recOrPos, variable ast.Expression
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_COMMA) && !p.curTokenIs(lexer.TOKEN_EOL) {
			recOrPos = p.parseExpression(PREC_LOWEST)
		}
	}
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		variable = p.parseExpression(PREC_LOWEST)
	}
	return &ast.PutStatement{BasePos: pos, FileNum: fileNum, RecordOrPos: recOrPos, Variable: variable}
}

func (p *Parser) parseSeekStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip SEEK
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
	}
	fileNum := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	position := p.parseExpression(PREC_LOWEST)
	return &ast.SeekStatement{BasePos: pos, FileNum: fileNum, Position: position}
}

func (p *Parser) parseFieldStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip FIELD
	if p.curTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
	}
	fileNum := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}

	var fields []ast.FieldDef
	for !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) {
		length := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_AS) {
			p.nextToken()
		}
		varName := p.curToken.Literal
		p.nextToken()
		fields = append(fields, ast.FieldDef{BasePos: p.curPos(), Length: length, VarName: varName})
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return &ast.FieldStatement{BasePos: pos, FileNum: fileNum, Fields: fields}
}

func (p *Parser) parseLsetRsetStatement() ast.Statement {
	pos := p.curPos()
	isLset := p.curTokenIs(lexer.TOKEN_LSET)
	p.nextToken() // skip LSET/RSET
	varName := p.curToken.Literal
	p.nextToken()
	if p.curTokenIs(lexer.TOKEN_EQ) {
		p.nextToken()
	}
	value := p.parseExpression(PREC_LOWEST)
	if isLset {
		return &ast.LsetStatement{BasePos: pos, Variable: varName, Value: value}
	}
	return &ast.RsetStatement{BasePos: pos, Variable: varName, Value: value}
}

func (p *Parser) parseSingleExprStatement(keyword string) ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip keyword
	expr := p.parseExpression(PREC_LOWEST)
	switch keyword {
	case "KILL":
		return &ast.KillStatement{BasePos: pos, Filename: expr}
	case "CHDIR":
		return &ast.ChdirStatement{BasePos: pos, Path: expr}
	case "MKDIR":
		return &ast.MkdirStatement{BasePos: pos, Path: expr}
	case "RMDIR":
		return &ast.RmdirStatement{BasePos: pos, Path: expr}
	}
	return nil
}

func (p *Parser) parseNameStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip NAME
	oldName := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_AS) {
		p.nextToken()
	}
	newName := p.parseExpression(PREC_LOWEST)
	return &ast.NameStatement{BasePos: pos, OldName: oldName, NewName: newName}
}

func (p *Parser) parseDelayStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip DELAY
	expr := p.parseExpression(PREC_LOWEST)
	// Represent as a function call in the AST
	return &ast.LetStatement{
		BasePos: pos,
		Name:    &ast.Identifier{BasePos: pos, Name: "DELAY"},
		Value:   expr,
	}
}

func (p *Parser) parseViewStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip VIEW
	stmt := &ast.ViewStatement{BasePos: pos}
	if p.curTokenIs(lexer.TOKEN_PRINT) {
		stmt.IsPrint = true
		p.nextToken()
		if !p.curTokenIs(lexer.TOKEN_EOL) && !p.curTokenIs(lexer.TOKEN_EOF) {
			stmt.Top = p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_TO) {
				p.nextToken()
				stmt.Bottom = p.parseExpression(PREC_LOWEST)
			}
		}
		return stmt
	}
	// VIEW graphics viewport — skip for now, parse coordinates
	p.skipToEndOfLine()
	return stmt
}

func (p *Parser) parseWindowStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip WINDOW
	p.skipToEndOfLine()
	return &ast.RemStatement{BasePos: pos, Text: "WINDOW"}
}

func (p *Parser) parsePaletteStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip PALETTE
	p.skipToEndOfLine()
	return &ast.RemStatement{BasePos: pos, Text: "PALETTE"}
}

// parseIdentifierStatement handles bare identifiers (assignments or calls).
func (p *Parser) parseIdentifierStatement() ast.Statement {
	pos := p.curPos()

	// Check if this is a line number
	if p.curTokenIs(lexer.TOKEN_INTEGER) {
		num, _ := strconv.Atoi(p.curToken.Literal)
		p.nextToken()
		return &ast.LineNumberStatement{BasePos: pos, Number: num}
	}

	// Handle identifiers that are tokenized with type-suffix characters but
	// are actually special statement keywords.
	upper := strings.ToUpper(p.curToken.Literal)
	switch upper {
	case "WRITE#":
		// WRITE# n, expr, ... — file write (lexer fuses WRITE and # into WRITE#)
		p.nextToken() // skip WRITE#
		fileNum := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		exprs := p.parseExpressionList()
		return &ast.FileWriteStatement{BasePos: pos, FileNum: fileNum, Expressions: exprs}
	case "PUT$":
		// PUT$ filenum, data$ — binary file put
		p.nextToken()
		fileNum := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		data := p.parseExpression(PREC_LOWEST)
		return &ast.FilePrintStatement{BasePos: pos, FileNum: fileNum, Expressions: []ast.Expression{data}}
	case "GET$":
		// GET$ filenum, length, var$ — binary file get
		p.nextToken()
		fileNum := p.parseExpression(PREC_LOWEST)
		var args []ast.Expression
		for p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			args = append(args, p.parseExpression(PREC_LOWEST))
		}
		return &ast.FileInputStatement{BasePos: pos, FileNum: fileNum, Variables: args}
	}

	// Identifier — could be assignment or sub call
	return p.parseAssignment(pos)
}

// isEndBlock returns true if the current END token is a block-closing keyword
// (END IF, END SELECT, END SUB, END FUNCTION, END DEF, END TYPE).
// A bare END on its own line is a program-termination statement, not a block closer.
func (p *Parser) isEndBlock() bool {
	next := p.peekToken
	return next.Type == lexer.TOKEN_IF ||
		next.Type == lexer.TOKEN_SELECT ||
		next.Type == lexer.TOKEN_SUB ||
		next.Type == lexer.TOKEN_FUNCTION ||
		next.Type == lexer.TOKEN_DEF ||
		next.Type == lexer.TOKEN_TYPE
}

// parseBlockUntil parses statements until one of the terminating tokens is found.
func (p *Parser) parseBlockUntil(terminators ...lexer.TokenType) []ast.Statement {
	var stmts []ast.Statement

	for {
		if p.curToken.Type == lexer.TOKEN_EOF {
			break
		}

		// Check for terminators.
		// Special case: TOKEN_END is only a block terminator if followed by
		// a block-closing keyword (IF, SELECT, SUB, FUNCTION, DEF).
		// A bare END (program-end statement) should be parsed normally.
		for _, t := range terminators {
			if p.curToken.Type == t {
				if t == lexer.TOKEN_END && !p.isEndBlock() {
					break // Not a block closer — fall through to parseStatement
				}
				return stmts
			}
		}

		if p.curToken.Type == lexer.TOKEN_EOL {
			p.nextToken()
			continue
		}

		if p.curToken.Type == lexer.TOKEN_COLON {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	return stmts
}

// parseIdentifierAsIdent parses the current token as a simple Identifier.
func (p *Parser) parseIdentifierAsIdent() *ast.Identifier {
	pos := p.curPos()
	name := p.curToken.Literal
	typeSuffix := ""
	if len(name) > 0 {
		last := name[len(name)-1]
		if last == '%' || last == '&' || last == '!' || last == '#' || last == '$' {
			typeSuffix = string(last)
			name = name[:len(name)-1]
		}
	}
	p.nextToken()
	return &ast.Identifier{BasePos: pos, Name: name, TypeSuffix: typeSuffix}
}

// parseTypeBlock parses a TYPE name ... END TYPE user-defined type declaration.
// Fields have the form: fieldname AS typename
func (p *Parser) parseTypeBlock() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip TYPE
	name := p.curToken.Literal
	p.nextToken()
	// skip to end of line after type name
	p.skipToEndOfLine()

	var fields []ast.TypeField
	for !p.curTokenIs(lexer.TOKEN_EOF) {
		if p.curTokenIs(lexer.TOKEN_EOL) {
			p.nextToken()
			continue
		}
		if p.curTokenIs(lexer.TOKEN_END) {
			p.nextToken() // skip END
			// skip TYPE
			if strings.ToUpper(p.curToken.Literal) == "TYPE" || p.curTokenIs(lexer.TOKEN_TYPE) {
				p.nextToken()
			}
			break
		}
		fieldName := p.curToken.Literal
		p.nextToken()
		typeName := ""
		if p.curTokenIs(lexer.TOKEN_AS) {
			p.nextToken() // skip AS
			typeName = strings.ToUpper(p.curToken.Literal)
			p.nextToken()
		}
		if fieldName != "" {
			fields = append(fields, ast.TypeField{Name: fieldName, TypeName: typeName})
		}
		p.skipToEndOfLine()
	}
	return &ast.TypeBlockStatement{BasePos: pos, Name: name, Fields: fields}
}

// parseFnAssign parses FN name = expr (return value assignment inside a DEF FN block)
// or FN name(args) as a standalone call (discard return value).
func (p *Parser) parseFnAssign() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip FN
	name := p.curToken.Literal
	// Strip type suffix if present (FNname$ etc.)
	if len(name) > 0 {
		last := name[len(name)-1]
		if last == '$' || last == '%' || last == '!' || last == '#' || last == '&' {
			name = name[:len(name)-1]
		}
	}
	p.nextToken() // skip name
	if p.curTokenIs(lexer.TOKEN_EQ) {
		p.nextToken() // skip =
		val := p.parseExpression(PREC_LOWEST)
		return &ast.FnAssignStatement{BasePos: pos, Name: name, Value: val}
	}
	// FN name(args) as a standalone call — use LetStatement with a dummy target
	var args []ast.Expression
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		args = p.parseExpressionList()
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
		}
	}
	call := &ast.FnCallExpression{BasePos: pos, Name: name, Args: args}
	return &ast.LetStatement{
		BasePos: pos,
		Name:    &ast.Identifier{BasePos: pos, Name: "_fn_" + name},
		Value:   call,
	}
}

// parsePokeStatement parses POKE address, value.
func (p *Parser) parsePokeStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip POKE
	addr := p.parseExpression(PREC_LOWEST)
	if p.curTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
	}
	val := p.parseExpression(PREC_LOWEST)
	return &ast.PokeStatement{BasePos: pos, Address: addr, Value: val}
}

// parseConstStatement parses CONST name = expr (compile-time constant declaration).
// In the transpiled Go output this becomes a regular variable assignment; the
// semantic distinction is preserved in the AST for potential optimisation.
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

// parseFnCallExpression parses FN name(args) in an expression context.
func (p *Parser) parseFnCallExpression() ast.Expression {
	pos := p.curPos()
	p.nextToken() // skip FN
	name := p.curToken.Literal
	// Strip type suffix if present (FNname$ etc.)
	if len(name) > 0 {
		last := name[len(name)-1]
		if last == '$' || last == '%' || last == '!' || last == '#' || last == '&' {
			name = name[:len(name)-1]
		}
	}
	p.nextToken()
	var args []ast.Expression
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		args = p.parseExpressionList()
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
		}
	}
	return &ast.FnCallExpression{BasePos: pos, Name: name, Args: args}
}
