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

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// Parser holds the state needed during a single-pass parse.
//
// The central design choice here is the two-token sliding window: curToken and
// peekToken. At any moment:
//
//   curToken  — the token the parser is currently acting on
//   peekToken — the token immediately after curToken (one lookahead)
//
// Two tokens are almost always sufficient for Turbo BASIC. The canonical
// example that requires peeking is distinguishing an assignment from a
// sub-call when seeing an IDENTIFIER:
//
//   MyVar = 5       → assignment (peek sees '=')
//   MySub arg1, arg2 → procedure call (peek sees something that is not '=')
//
// If more than one token of lookahead were ever needed, the standard fix is to
// add a third slot (peek2Token) rather than a full token buffer.
//
// errors accumulates every parse error as a formatted string. Collecting all
// errors (rather than stopping at the first) lets the user fix multiple
// mistakes in one compile cycle — a major quality-of-life improvement.
type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token // token currently being examined ("current" lookahead)
	peekToken lexer.Token // next token (one position ahead, "future" lookahead)
	errors    []string    // all parse errors collected so far; nil until an error occurs
}

// New creates a new Parser for the given lexer.
//
// Before returning, New "primes the pump" by calling nextToken() twice. After
// each call, the sliding window shifts by one:
//
//   Initial state:   curToken=zero, peekToken=zero
//   After 1st call:  curToken=zero, peekToken=first_real_token
//   After 2nd call:  curToken=first_real_token, peekToken=second_real_token
//
// After New() returns, the parser is ready to inspect the first real token
// via p.curToken without any special-casing in ParseProgram.
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
//
// The outer loop is the "statement driver" — the engine that keeps the parser
// moving forward. Its contract is simple:
//
//   1. Skip blank lines and colons (BASIC allows multiple statements per line
//      separated by colons, e.g.: FOR i = 1 TO 5 : PRINT i : NEXT i).
//   2. Call parseStatement() to consume and build one statement node.
//   3. Consume any trailing EOL or colon separator.
//   4. Repeat until TOKEN_EOF.
//
// Each parseXxx() method is responsible for consuming exactly the tokens that
// belong to its construct and leaving the cursor at the *next* unconsumed
// token. If every method upholds this contract, the driver loop composes them
// without any overlap or gap.
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

// nextToken advances the sliding window by one position.
//
// This is the single most-called method in the parser. The pattern:
//
//   p.curToken  = p.peekToken       // old "next" becomes "current"
//   p.peekToken = p.l.NextToken()   // fetch a fresh token from the lexer
//
// Think of it as sliding a two-character window along a string: the right
// character moves into the left slot, and the lexer fills the new right slot.
// The lexer is called lazily — it produces tokens one at a time on demand,
// rather than scanning the entire source up-front. This keeps memory usage
// proportional to the window size (two tokens), not the input size.
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

// expectPeek is the "assert and advance" helper used throughout grammar rules.
//
// Many grammar rules have mandatory tokens in fixed positions, for example:
//
//   FOR counter = start TO end
//                 ^           ^ these '=' and 'TO' are mandatory
//
// expectPeek checks that the next token is what the grammar requires. If it
// is, the window advances (consuming the expected token) and the method returns
// true so parsing can continue. If it is not, a descriptive error is recorded
// and the method returns false so the caller can decide how to recover.
//
// This pattern separates "is the next thing what I expect?" from "now what?" —
// the caller handles the false case, typically by returning nil or skipping to
// the end of the line.
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

// skipToEndOfLine implements "panic-mode error recovery".
//
// When a parse error is detected, the parser is in an uncertain state: it does
// not know what tokens follow the bad construct, so it cannot safely continue
// parsing the current statement. The simplest recovery strategy — and the one
// used by many production compilers — is to discard tokens until a known
// synchronisation point is reached.
//
// For BASIC, the natural synchronisation point is the end of the line (EOL)
// or the colon that separates statements on a single line. After recovery, the
// outer ParseProgram loop re-enters parseStatement() at a fresh statement
// boundary, maximising the chance of producing useful diagnostics for the rest
// of the file.
//
// This is called "panic mode" recovery because the parser "panics" (throws
// away tokens) until it finds solid ground. More sophisticated recovery
// strategies (like Fischer-LeBlanc) exist but are rarely worth the complexity
// cost for compiler projects.
func (p *Parser) skipToEndOfLine() {
	for p.curToken.Type != lexer.TOKEN_EOL && p.curToken.Type != lexer.TOKEN_EOF &&
		p.curToken.Type != lexer.TOKEN_COLON {
		p.nextToken()
	}
}
