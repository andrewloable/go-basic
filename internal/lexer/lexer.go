package lexer

import (
	"strings"
	"unicode"
)

// Lexer tokenizes Turbo BASIC source code into a stream of tokens.
type Lexer struct {
	input        string
	pos          int  // current position in input (points to current char)
	readPos      int  // current reading position (after current char)
	ch           byte // current char under examination
	line         int  // current line number (1-based)
	col          int  // current column number (1-based)
	atLineStart  bool // true if we're at the beginning of a line (for line number/label detection)
	prevTokenEOL bool // true if previous token was EOL or start of input
}

// New creates a new Lexer for the given input source code.
func New(input string) *Lexer {
	l := &Lexer{
		input:        input,
		line:         1,
		col:          0,
		atLineStart:  true,
		prevTokenEOL: true,
	}
	l.readChar()
	return l
}

// readChar advances the lexer by one character.
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// skipWhitespace skips spaces and tabs (but not newlines).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := Token{Line: l.line, Column: l.col}

	switch {
	case l.ch == 0:
		tok.Type = TOKEN_EOF
		tok.Literal = ""
		return tok

	case l.ch == '\r' || l.ch == '\n':
		tok = l.readNewline()
		l.atLineStart = true
		l.prevTokenEOL = true
		return tok

	case l.ch == '\'':
		// Single-quote comment (Turbo BASIC extension)
		tok = l.readSingleQuoteComment()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	case l.ch == '"':
		tok = l.readString()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	case l.ch == '&' && (l.peekChar() == 'H' || l.peekChar() == 'h' ||
		l.peekChar() == 'O' || l.peekChar() == 'o' ||
		l.peekChar() == 'B' || l.peekChar() == 'b'):
		tok = l.readSpecialRadix()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	case l.ch == '$' && l.prevTokenEOL:
		// Metacommand at start of line or after REM
		tok = l.readMetacommand()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	case isDigit(l.ch) || (l.ch == '.' && isDigit(l.peekChar())):
		tok = l.readNumber()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	case isLetter(l.ch) || l.ch == '_':
		tok = l.readIdentOrKeyword()
		if tok.Type == TOKEN_REM {
			// REM keyword: rest of line is a comment
			tok = l.readRemComment(tok)
		}
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok

	default:
		tok = l.readOperatorOrPunctuation()
		l.atLineStart = false
		l.prevTokenEOL = false
		return tok
	}
}

// readNewline handles CR, LF, or CR+LF line endings.
func (l *Lexer) readNewline() Token {
	tok := Token{Type: TOKEN_EOL, Line: l.line, Column: l.col}
	tok.Literal = "\n"
	if l.ch == '\r' {
		l.readChar()
		if l.ch == '\n' {
			l.readChar()
		}
	} else {
		l.readChar()
	}
	l.line++
	l.col = 1 // next char read is already at col 1 of the new line
	return tok
}

// readString reads a double-quoted string literal.
// Embedded quotes are represented by two consecutive double-quotes.
func (l *Lexer) readString() Token {
	tok := Token{Type: TOKEN_STRING, Line: l.line, Column: l.col}
	l.readChar() // skip opening quote

	var sb strings.Builder
	for {
		if l.ch == 0 || l.ch == '\n' || l.ch == '\r' {
			// Unterminated string
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = sb.String()
			return tok
		}
		if l.ch == '"' {
			if l.peekChar() == '"' {
				// Embedded quote
				sb.WriteByte('"')
				l.readChar()
				l.readChar()
				continue
			}
			// End of string
			l.readChar() // skip closing quote
			break
		}
		sb.WriteByte(l.ch)
		l.readChar()
	}

	tok.Literal = sb.String()
	return tok
}

// readSingleQuoteComment reads a '-style comment.
func (l *Lexer) readSingleQuoteComment() Token {
	tok := Token{Type: TOKEN_COMMENT, Line: l.line, Column: l.col}
	l.readChar() // skip '
	start := l.pos
	for l.ch != 0 && l.ch != '\n' && l.ch != '\r' {
		l.readChar()
	}
	tok.Literal = l.input[start:l.pos]
	return tok
}

// readRemComment reads the rest of the line after a REM keyword.
func (l *Lexer) readRemComment(remTok Token) Token {
	tok := Token{Type: TOKEN_COMMENT, Line: remTok.Line, Column: remTok.Column}
	// Skip optional space after REM
	if l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
	start := l.pos
	for l.ch != 0 && l.ch != '\n' && l.ch != '\r' {
		l.readChar()
	}
	tok.Literal = l.input[start:l.pos]
	return tok
}

// readNumber reads an integer or floating-point number literal.
func (l *Lexer) readNumber() Token {
	tok := Token{Line: l.line, Column: l.col}
	start := l.pos
	hasDecimal := false
	hasExponent := false

	// Read digits before decimal point
	if l.ch == '.' {
		hasDecimal = true
		l.readChar()
	}

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if !hasDecimal && l.ch == '.' && l.peekChar() != '.' {
		hasDecimal = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for exponent (E/e for single, D/d for double)
	if l.ch == 'E' || l.ch == 'e' || l.ch == 'D' || l.ch == 'd' {
		isDouble := l.ch == 'D' || l.ch == 'd'
		hasExponent = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
		if isDouble {
			tok.Type = TOKEN_DOUBLE
			tok.Literal = l.input[start:l.pos]
			return tok
		}
	}

	// Check for type suffix
	switch l.ch {
	case '&':
		tok.Type = TOKEN_LONG
		tok.Literal = l.input[start:l.pos]
		l.readChar()
		return tok
	case '!':
		tok.Type = TOKEN_SINGLE
		tok.Literal = l.input[start:l.pos]
		l.readChar()
		return tok
	case '#':
		tok.Type = TOKEN_DOUBLE
		tok.Literal = l.input[start:l.pos]
		l.readChar()
		return tok
	case '%':
		tok.Type = TOKEN_INTEGER
		tok.Literal = l.input[start:l.pos]
		l.readChar()
		return tok
	}

	tok.Literal = l.input[start:l.pos]

	if hasDecimal || hasExponent {
		tok.Type = TOKEN_SINGLE
	} else {
		tok.Type = TOKEN_INTEGER
	}
	return tok
}

// readSpecialRadix reads &H (hex), &O (octal), or &B (binary) literals.
func (l *Lexer) readSpecialRadix() Token {
	tok := Token{Line: l.line, Column: l.col}
	l.readChar() // skip &
	radixChar := l.ch
	l.readChar() // skip H/O/B

	start := l.pos

	switch radixChar {
	case 'H', 'h':
		tok.Type = TOKEN_HEX
		for isHexDigit(l.ch) {
			l.readChar()
		}
	case 'O', 'o':
		tok.Type = TOKEN_OCTAL
		for l.ch >= '0' && l.ch <= '7' {
			l.readChar()
		}
	case 'B', 'b':
		tok.Type = TOKEN_BINARY_LIT
		for l.ch == '0' || l.ch == '1' {
			l.readChar()
		}
	}

	tok.Literal = l.input[start:l.pos]
	if tok.Literal == "" {
		tok.Type = TOKEN_ILLEGAL
		tok.Literal = string([]byte{'&', radixChar})
	}
	return tok
}

// readIdentOrKeyword reads an identifier or keyword.
func (l *Lexer) readIdentOrKeyword() Token {
	tok := Token{Line: l.line, Column: l.col}
	start := l.pos

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}

	literal := l.input[start:l.pos]
	upper := strings.ToUpper(literal)

	// Check for type suffix on identifiers
	switch l.ch {
	case '%':
		tok.Type = TOKEN_IDENTIFIER
		tok.Literal = literal + "%"
		l.readChar()
		return tok
	case '&':
		tok.Type = TOKEN_IDENTIFIER
		tok.Literal = literal + "&"
		l.readChar()
		return tok
	case '!':
		tok.Type = TOKEN_IDENTIFIER
		tok.Literal = literal + "!"
		l.readChar()
		return tok
	case '#':
		tok.Type = TOKEN_IDENTIFIER
		tok.Literal = literal + "#"
		l.readChar()
		return tok
	case '$':
		tok.Type = TOKEN_IDENTIFIER
		tok.Literal = literal + "$"
		l.readChar()
		return tok
	}

	// Check if it's a keyword
	tok.Type = LookupIdent(upper)
	tok.Literal = literal

	// Check if this identifier at the start of line followed by ':' is a label
	if tok.Type == TOKEN_IDENTIFIER && l.atLineStart {
		savedPos := l.pos
		savedReadPos := l.readPos
		savedCh := l.ch
		savedCol := l.col

		l.skipWhitespace()
		if l.ch == ':' {
			tok.Type = TOKEN_LABEL
			tok.Literal = literal
			l.readChar() // consume the colon
			return tok
		}
		// Restore position
		l.pos = savedPos
		l.readPos = savedReadPos
		l.ch = savedCh
		l.col = savedCol
	}

	return tok
}

// readMetacommand reads a $-prefixed metacommand.
func (l *Lexer) readMetacommand() Token {
	tok := Token{Line: l.line, Column: l.col}
	l.readChar() // skip $

	start := l.pos
	for isLetter(l.ch) {
		l.readChar()
	}

	word := strings.ToUpper(l.input[start:l.pos])

	switch word {
	case "DYNAMIC":
		tok.Type = TOKEN_META_DYNAMIC
	case "STATIC":
		tok.Type = TOKEN_META_STATIC
	case "INCLUDE":
		tok.Type = TOKEN_META_INCLUDE
	case "IF":
		tok.Type = TOKEN_META_IF
	case "ELSEIF":
		tok.Type = TOKEN_META_ELSEIF
	case "ELSE":
		tok.Type = TOKEN_META_ELSE
	case "ENDIF":
		tok.Type = TOKEN_META_ENDIF
	case "COM":
		tok.Type = TOKEN_META_COM
	case "SOUND":
		tok.Type = TOKEN_META_SOUND
	case "STACK":
		tok.Type = TOKEN_META_STACK
	case "SEGMENT":
		tok.Type = TOKEN_META_SEGMENT
	case "INLINE":
		tok.Type = TOKEN_META_INLINE
	case "EVENT":
		tok.Type = TOKEN_META_EVENT
	default:
		tok.Type = TOKEN_ILLEGAL
	}

	tok.Literal = "$" + word
	return tok
}

// readOperatorOrPunctuation reads operator and punctuation tokens.
func (l *Lexer) readOperatorOrPunctuation() Token {
	tok := Token{Line: l.line, Column: l.col}

	switch l.ch {
	case '+':
		tok.Type = TOKEN_PLUS
		tok.Literal = "+"
	case '-':
		tok.Type = TOKEN_MINUS
		tok.Literal = "-"
	case '*':
		tok.Type = TOKEN_STAR
		tok.Literal = "*"
	case '/':
		tok.Type = TOKEN_SLASH
		tok.Literal = "/"
	case '\\':
		tok.Type = TOKEN_BACKSLASH
		tok.Literal = "\\"
	case '^':
		tok.Type = TOKEN_CARET
		tok.Literal = "^"
	case '=':
		tok.Type = TOKEN_EQ
		tok.Literal = "="
	case '<':
		if l.peekChar() == '>' {
			tok.Type = TOKEN_NE
			tok.Literal = "<>"
			l.readChar()
		} else if l.peekChar() == '=' {
			tok.Type = TOKEN_LE
			tok.Literal = "<="
			l.readChar()
		} else {
			tok.Type = TOKEN_LT
			tok.Literal = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			tok.Type = TOKEN_GE
			tok.Literal = ">="
			l.readChar()
		} else {
			tok.Type = TOKEN_GT
			tok.Literal = ">"
		}
	case '(':
		tok.Type = TOKEN_LPAREN
		tok.Literal = "("
	case ')':
		tok.Type = TOKEN_RPAREN
		tok.Literal = ")"
	case ',':
		tok.Type = TOKEN_COMMA
		tok.Literal = ","
	case ';':
		tok.Type = TOKEN_SEMICOLON
		tok.Literal = ";"
	case ':':
		tok.Type = TOKEN_COLON
		tok.Literal = ":"
	case '#':
		tok.Type = TOKEN_HASH
		tok.Literal = "#"
	case '$':
		tok.Type = TOKEN_DOLLAR
		tok.Literal = "$"
	case '%':
		tok.Type = TOKEN_PERCENT
		tok.Literal = "%"
	case '&':
		tok.Type = TOKEN_AMPERSAND
		tok.Literal = "&"
	case '!':
		tok.Type = TOKEN_EXCLAIM
		tok.Literal = "!"
	default:
		tok.Type = TOKEN_ILLEGAL
		tok.Literal = string(l.ch)
	}

	l.readChar()
	return tok
}

// AllTokens lexes the entire input and returns all tokens.
func (l *Lexer) AllTokens() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
}
