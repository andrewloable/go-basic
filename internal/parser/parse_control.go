package parser

import (
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// parseIfStatement handles both the single-line and block forms of IF.
//
// The disambiguation between the two forms is made by looking at the token
// immediately after THEN:
//
//   IF cond THEN<EOL>   → block form  (THEN is the last token on the line)
//   IF cond THEN stmt   → single-line (THEN is followed by something else)
//
// A single token of lookahead (curToken after advancing past THEN) is enough
// to make this decision — a classic example of LL(1) parsing where one token
// of lookahead suffices for the grammar.
//
// The block form calls the general parseBlockUntil() helper to consume
// statements until ELSEIF, ELSE, or END is seen. parseBlockUntil() is the
// compiler's "read statements until you see a sentinel" workhorse and is
// used by all block-structured constructs (FOR, WHILE, DO, SUB, etc.).
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

// parseForStatement handles FOR counter = start TO end [STEP step] … NEXT.
//
// The grammar is rigid — the tokens appear in a fixed order — making this a
// straightforward example of recursive-descent parsing where each token is
// consumed with a specific expectation. Any deviation from the expected order
// triggers an error and bails out via skipToEndOfLine().
//
// Step is parsed only when the optional STEP keyword is present. When absent,
// Step is left as nil in the ForStatement node; the code generator substitutes
// the default value of 1 at emit time.
//
// After the body is parsed, NEXT [counter] is consumed. The optional counter
// name after NEXT (e.g. "NEXT i") is a BASIC readability convention; the
// parser validates only its syntactic presence, not whether it matches the
// FOR variable (that would be semantic analysis).
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
