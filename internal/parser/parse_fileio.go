package parser

import (
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

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

func (p *Parser) parseGetStatement() ast.Statement {
	pos := p.curPos()
	p.nextToken() // skip GET

	// Graphics GET: GET (x1, y1)-(x2, y2), arrayVar
	// File GET:     GET [#]filenum [, recnum] [, variable]
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		// Graphics GET — parse (x1, y1)-(x2, y2), arrayVar
		p.nextToken() // skip (
		x1 := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		y1 := p.parseExpression(PREC_LOWEST)
		if p.curTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken() // skip )
		}
		if p.curTokenIs(lexer.TOKEN_MINUS) {
			p.nextToken() // skip -
		}
		var x2, y2 ast.Expression
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken() // skip (
			x2 = p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			y2 = p.parseExpression(PREC_LOWEST)
			if p.curTokenIs(lexer.TOKEN_RPAREN) {
				p.nextToken() // skip )
			}
		}
		var arrayVar ast.Expression
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			arrayVar = p.parseExpression(PREC_LOWEST)
		}
		return &ast.GraphicsGetStatement{BasePos: pos, X1: x1, Y1: y1, X2: x2, Y2: y2, ArrayVar: arrayVar}
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
		var arrayVar ast.Expression
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			arrayVar = p.parseExpression(PREC_LOWEST)
		}
		action := "XOR" // default action verb
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			switch p.curToken.Type {
			case lexer.TOKEN_PSET, lexer.TOKEN_PRESET, lexer.TOKEN_AND, lexer.TOKEN_OR, lexer.TOKEN_XOR:
				action = strings.ToUpper(p.curToken.Literal)
				p.nextToken()
			default:
				// Try to read as identifier
				expr := p.parseExpression(PREC_LOWEST)
				if id, ok := expr.(*ast.Identifier); ok {
					action = strings.ToUpper(id.Name)
				}
			}
		}
		return &ast.GraphicsPutStatement{BasePos: pos, X: x, Y: y, ArrayVar: arrayVar, Action: action}
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
