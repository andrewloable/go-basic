package parser

import (
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

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

// parseIdentifierStatement handles bare identifiers (assignments or sub-calls).
//
// When parseStatement() sees TOKEN_IDENTIFIER or TOKEN_INTEGER it cannot
// immediately know what construct it is looking at. It delegates here to
// disambiguate:
//
//   Integer token     → line number statement (old-style BASIC line numbering)
//   WRITE# / GET$ / PUT$ → special file I/O forms fused with a suffix token
//   identifier = expr → assignment (handled by parseAssignment)
//   identifier args   → implicit sub-call without CALL keyword (same as CALL)
//
// The file I/O special cases (WRITE#, GET$, PUT$) arise because the lexer
// attaches the '#' or '$' character to the keyword as a type suffix, producing
// a single TOKEN_IDENTIFIER with a compound literal. Catching these here avoids
// adding more token types to the main parseStatement dispatch.
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
		return &ast.FilePrintStatement{BasePos: pos, FileNum: fileNum, Expressions: []ast.Expression{data}, IsBinaryPut: true}
	case "GET$":
		// GET$ filenum, length, var$ — binary file get
		p.nextToken()
		fileNum := p.parseExpression(PREC_LOWEST)
		var args []ast.Expression
		for p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			args = append(args, p.parseExpression(PREC_LOWEST))
		}
		return &ast.FileInputStatement{BasePos: pos, FileNum: fileNum, Variables: args, IsBinaryGet: true}
	}

	// Identifier — could be assignment or sub call
	return p.parseAssignment(pos)
}

// isEndBlock returns true if the current END token is a block-closing keyword
// (END IF, END SELECT, END SUB, END FUNCTION, END DEF, END TYPE).
// A bare END on its own line is a program-termination statement, not a block closer.
//
// Tutorial note — why END is ambiguous
//
// "END" in BASIC has two completely different meanings:
//
//   1. Program terminator:   END          (stops the program, like os.Exit)
//   2. Block closer:         END IF       (closes an IF block)
//                            END SUB      (closes a SUB declaration)
//                            END FUNCTION (closes a FUNCTION declaration)
//                            … and so on
//
// The parser must look one token ahead (peekToken) to tell the two cases apart.
// isEndBlock() encapsulates this lookahead so that parseBlockUntil() can call
// it in one place rather than repeating the switch everywhere.
//
// This is an example of a context-sensitive grammar rule where the same token
// ("END") has different structural meanings depending on what follows it.
// A strictly context-free grammar cannot express this without an epsilon rule
// and a costly reorganisation; a recursive-descent parser handles it cleanly
// with a single lookahead check.
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
//
// This is the compiler's "read a block of code" workhorse. It is called by
// every block-structured construct — IF, FOR, WHILE, DO, SUB, FUNCTION, DEF —
// and in each case the caller passes the specific token(s) that mark the end
// of that block:
//
//   parseBlockUntil(TOKEN_NEXT)          → reads the body of a FOR loop
//   parseBlockUntil(TOKEN_WEND)          → reads the body of a WHILE loop
//   parseBlockUntil(TOKEN_END)           → reads the body of a SUB or FUNCTION
//   parseBlockUntil(TOKEN_ELSE, TOKEN_ELSEIF, TOKEN_END) → reads an IF branch
//
// The variadic terminators ...TokenType design allows different constructs to
// share a single generic block reader without creating separate specialised
// functions for each loop type.
//
// Special care is taken for TOKEN_END because "END" can appear both as a
// block terminator (END SUB) and as a program-end statement. The isEndBlock()
// lookahead check resolves this ambiguity.
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
