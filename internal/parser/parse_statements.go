package parser

import (
	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// ---------------------------------------------------------------------------
// Statement parsers
//
// Tutorial note — statement parsing vs. expression parsing
//
// Statement parsing uses straightforward recursive descent: parseStatement()
// inspects curToken.Type and dispatches to the specific handler for that
// keyword. Each handler is responsible for:
//
//   1. Consuming all tokens that belong to the statement (including the
//      leading keyword already in curToken).
//   2. Building and returning the corresponding AST node.
//   3. Leaving curToken pointing at the first token AFTER the statement
//      (i.e., the EOL, colon, or next statement's keyword).
//
// This contract is what allows ParseProgram's outer loop to sequence
// statements without any coordination between handlers.
//
// Expression parsing is different (see parse_expressions.go): it uses the
// Pratt algorithm because expressions can nest arbitrarily and their
// structure depends on numeric precedence, not on a fixed keyword.
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
//
// BASIC's PRINT statement uses separators to control spacing:
//
//   PRINT a, b     → comma moves to the next print zone (~14-char column)
//   PRINT a; b     → semicolon prints with no space between values
//   PRINT a b      → juxtaposition (no separator) treated the same as semicolon
//
// The parser collects expressions and separators in parallel slices so that
// the code generator can emit the correct Go formatting calls for each pair.
// HasTrailingSep is set when a separator appears after the last expression,
// indicating that the cursor should stay on the same line (no newline output).
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

// parseAssignment handles LET x = expr, arr(i) = expr, and struct.field = expr.
//
// BASIC allows three assignment targets:
//
//   - Simple variable:    x = 5        → LetStatement
//   - Array element:      arr(2) = 10  → ArrayAssignment
//   - TYPE record field:  p.x = 1.0    → FieldAssignStatement
//
// The parser re-uses parseIdentifierExpression() (an expression parser) to
// parse the left-hand side, then inspects the concrete type of the returned
// node to determine which AST statement node to create. This avoids duplicating
// the identifier/array/field parsing logic in both expression and statement contexts.
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
		// WIDTH col [, row] — set text dimensions; currently a no-op
		cols := p.parseExpression(PREC_LOWEST)
		var rows ast.Expression
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			rows = p.parseExpression(PREC_LOWEST)
		}
		_ = cols
		_ = rows
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

