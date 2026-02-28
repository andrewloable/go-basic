// Package lexer implements Phase 1 of compilation: Lexical Analysis (also called
// lexing, scanning, or tokenizing).
//
// # What the lexer does
//
// The lexer transforms raw source text (a flat stream of characters) into a
// structured stream of Token objects. Each Token carries three pieces of data:
//
//   - Type     – a symbolic category (e.g. TOKEN_INTEGER, TOKEN_IDENTIFIER, TOKEN_PLUS)
//   - Literal  – the exact text slice from the source (e.g. "42", "myVar%", "+")
//   - Position – the (line, column) coordinates for error reporting
//
// The parser (Phase 2) never sees raw characters; it only sees this token stream.
// This clean separation of concerns keeps the grammar rules in the parser simple.
//
// # Design goal: single-pass scanning
//
// The lexer reads the input exactly once, left-to-right, one byte at a time.
// Two cursor fields – pos (current byte) and readPos (next byte) – implement a
// one-character look-ahead window via peekChar(). This is sufficient to resolve
// all ambiguities in Turbo BASIC syntax without backtracking.
//
// # Turbo BASIC peculiarities handled here
//
//   - Type suffixes on literals: 42& (long), 3.14! (single), 1.0# (double), 99% (integer)
//   - Type suffixes on identifiers: count% (integer var), name$ (string var)
//   - Alternate-radix literals: &H1F (hex), &O77 (octal), &B1010 (binary)
//   - Two comment styles: REM ... and ' ...
//   - Metacommands: $INCLUDE, $DYNAMIC, $STATIC, etc. (only at line start)
//   - Line-start label detection: an identifier followed by ':' becomes TOKEN_LABEL
//   - Exponent suffixes: 1.5E3 (single-precision), 1.5D3 (double-precision)
//   - Multi-character operators: <>, <=, >=
//   - Platform line endings: CR, LF, and CR+LF are all normalised to TOKEN_EOL
package lexer

import (
	"strings"
	"unicode"
)

// # Lexer state
//
// The Lexer struct is essentially a cursor into a string. All state required to
// produce the next token is stored here; no global state is used, so multiple
// Lexer instances can run concurrently.
//
// Key fields:
//
//   - input        – the entire source file held in memory as an immutable string
//   - pos/readPos  – two indices that implement the sliding one-char look-ahead;
//                    pos points at the byte currently being examined (l.ch),
//                    readPos points at the byte that will be consumed next
//   - ch           – the byte at pos, cached to avoid repeated index expressions
//   - line/col     – source coordinates updated on every readChar() call; used
//                    to stamp every token with its origin for error messages
//   - atLineStart  – tracks whether we are at the first non-whitespace position
//                    of a logical line; needed to distinguish labels and metacommands
//                    from mid-line identifiers and '$' operators
//   - prevTokenEOL – true at the very start of input and immediately after each
//                    TOKEN_EOL; used together with atLineStart for metacommand detection

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
//
// The constructor primes the cursor by calling readChar() once. After New()
// returns, l.ch holds the very first byte of the input and l.pos == 0, so the
// first call to NextToken() can begin scanning immediately without any special
// "start of input" check inside the main loop.
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

// readChar advances the lexer cursor by exactly one byte.
//
// The pattern of maintaining two indices (pos and readPos) is called a
// "two-pointer sliding window" and is a classic technique for implementing
// one-character lookahead without a separate buffer:
//
//   pos      — points at the byte currently stored in l.ch (already read)
//   readPos  — points at the NEXT byte to be read
//
// After readChar() returns:
//   old readPos → new pos
//   new readPos → old readPos + 1
//   l.ch        → the byte at new pos (or 0 if past end)
//
// Setting l.ch = 0 at end-of-input is a sentinel value (NUL byte). All
// character-class predicates (isLetter, isDigit) return false for 0, so the
// lexer naturally stops consuming characters when it hits the end of input.
//
// Byte counting (l.col++) keeps the column coordinate accurate for error
// messages. Line counting is handled separately in readNewline().
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // sentinel: NUL byte signals end of input
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++
}

// peekChar returns the next character without advancing the cursor.
//
// This is the one-character lookahead that lets the lexer resolve ambiguities
// involving two-character tokens:
//
//   '<'  followed by '>'  → TOKEN_NE  ("<>")
//   '<'  followed by '='  → TOKEN_LE  ("<=")
//   '&'  followed by 'H'  → start of hex literal ("&H...")
//   first digit of float  → peek after '.' to distinguish "1.5" from ".."
//
// peekChar() reads directly from l.input[l.readPos] without modifying pos,
// readPos, or ch, so the cursor position is unchanged after the call. The
// character returned is valid only until the next readChar() call.
//
// If there is no next character (end of input), peekChar returns 0 (NUL),
// the same sentinel used by readChar().
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// skipWhitespace advances past spaces and tabs only.
//
// Critically, newlines (\r and \n) are NOT skipped here. In BASIC, a newline
// ends a statement — it is semantically significant. The lexer emits TOKEN_EOL
// for newlines so the parser can detect statement boundaries. Treating newlines
// as ignorable whitespace (like C does) would break BASIC's line-oriented syntax.
//
// Tab characters (\t) ARE skipped because they are purely visual indentation;
// BASIC does not use tabs for semantic purposes (unlike Python).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
}

// NextToken is the main dispatch loop of the lexer.
//
// Each call returns exactly one token and advances the cursor past the
// characters that make up that token. The caller (parser or test harness)
// drives the lexer by repeatedly calling NextToken() until TOKEN_EOF.
//
// Dispatch strategy – ordered by decreasing specificity:
//
//  1. Whitespace is discarded first (spaces and tabs are not tokens in BASIC).
//  2. The NUL byte (ch == 0) signals end-of-input → TOKEN_EOF.
//  3. Newline characters (\r, \n) → TOKEN_EOL; BASIC is line-oriented so
//     line endings are semantically significant (they terminate statements).
//  4. Context-sensitive cases are tested before general ones:
//     - '&' followed by H/O/B → radix prefix literal (must beat plain '&' operator)
//     - '$' at line-start → metacommand (must beat '$' as a type-suffix operator)
//  5. Digit or '.' followed by digit → numeric literal.
//  6. Letter or '_' → identifier or keyword.
//  7. Anything else → operator or punctuation (or TOKEN_ILLEGAL).
//
// After each token, atLineStart and prevTokenEOL are updated so that the next
// call can make the same context-sensitive decisions correctly.
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

// readString scans a double-quoted string literal and applies Turbo BASIC's
// escape convention.
//
// String literal parsing is straightforward: consume characters between the
// opening and closing '"'. The only subtlety is the escape mechanism. Unlike C
// (which uses backslash escapes), Turbo BASIC represents a literal double-quote
// inside a string by doubling it: PRINT "say ""hello""" prints  say "hello".
// This is the classic Pascal/SQL convention.
//
// Error handling: if the closing quote is never found before a newline or EOF,
// the token type is set to TOKEN_ILLEGAL. The parser will then emit a
// "unterminated string literal" diagnostic using the token's position fields.
//
// Implementation note: a strings.Builder accumulates the decoded content so
// that doubled-quote pairs appear as a single '"' in the final Literal, matching
// what the runtime will actually use.
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

// readNumber recognises all numeric literal forms supported by Turbo BASIC and
// maps each to a typed token.
//
// Numeric literal recognition in a single pass works by tracking boolean flags
// as the cursor advances:
//
//   - hasDecimal   – set when a '.' is consumed mid-number; implies float
//   - hasExponent  – set when E/e/D/d is consumed; implies float
//
// The character sequence that can start a number (dispatched from NextToken) is:
// a decimal digit, OR a '.' immediately followed by a digit (e.g. ".5").
//
// Precision is determined in this priority order:
//
//  1. Exponent letter 'D'/'d' → TOKEN_DOUBLE (regardless of other flags)
//  2. Type suffix character after the digits:
//     '&' → TOKEN_LONG   (32-bit integer,  e.g. 100000&)
//     '!' → TOKEN_SINGLE (32-bit float,    e.g. 3.14!)
//     '#' → TOKEN_DOUBLE (64-bit float,    e.g. 3.14#)
//     '%' → TOKEN_INTEGER(16-bit integer,  e.g. 42%)
//  3. Decimal point or 'E'/'e' exponent → TOKEN_SINGLE (default float)
//  4. Plain digits with no suffix or point → TOKEN_INTEGER
//
// The Literal field always contains the raw digit characters WITHOUT the type
// suffix, because the suffix is purely a lexer-level hint; the parser and
// evaluator work with the token type, not the suffix character.
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

// readSpecialRadix handles alternate-base integer literals.
//
// Turbo BASIC uses an '&' prefix followed by a radix letter to write integer
// constants in bases other than 10:
//
//   &Hnn  – hexadecimal (base 16): digits 0-9, A-F
//   &Onn  – octal       (base  8): digits 0-7
//   &Bnn  – binary      (base  2): digits 0-1
//
// The dispatch in NextToken already confirmed that the character after '&' is
// one of H/h/O/o/B/b, so this function can unconditionally consume '&' and
// the radix letter, then read the appropriate digit alphabet in a tight loop.
//
// If no valid digits follow the radix letter the token is TOKEN_ILLEGAL, which
// lets the parser produce a meaningful "empty numeric literal" error.
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

// readIdentOrKeyword performs keyword disambiguation – the process of deciding
// whether a sequence of word characters is a reserved keyword or a user-defined
// identifier.
//
// # How keyword disambiguation works
//
// The lexer first greedily consumes all characters that can belong to a BASIC
// word: letters, digits, underscores, and dots (BASIC allows dots in names for
// record-field style identifiers like "player.score").
//
// It then checks, in order:
//
//  1. Type suffix immediately after the word (%, &, !, #, $).
//     If present, the token is unconditionally TOKEN_IDENTIFIER – no keyword
//     can carry a type suffix (PRINT% is not a keyword, it is a variable).
//     The suffix is folded into the Literal so the parser/evaluator can recover
//     the declared type (e.g. "count%" → integer variable).
//
//  2. Keyword table lookup via LookupIdent(upper).
//     The raw text is uppercased before lookup because Turbo BASIC is
//     case-insensitive: "print", "Print", and "PRINT" are all the same keyword.
//     LookupIdent returns TOKEN_IDENTIFIER for any word not in the table.
//
//  3. Label detection.
//     After a keyword lookup returns TOKEN_IDENTIFIER AND we are at the start
//     of a logical line, the lexer speculatively skips whitespace and peeks at
//     the next character. If it is ':', the word is a label definition and we
//     emit TOKEN_LABEL (consuming the colon). Otherwise the saved cursor state
//     is restored so scanning continues from where it left off. This is the one
//     place in the lexer where a small, bounded lookahead beyond one character
//     is needed.
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

// readOperatorOrPunctuation produces a Token for a single- or double-character
// operator or punctuation symbol.
//
// Most operators are a single character ('+', '-', '=', …). The three
// two-character operators in BASIC (<>, <=, >=) are handled by peeking at the
// next character: if it completes a two-character operator, both characters are
// consumed in one call; otherwise only the first is consumed.
//
// For example, when the current character is '<':
//
//   peekChar() == '>'  → consume both, emit TOKEN_NE (literal "<>")
//   peekChar() == '='  → consume both, emit TOKEN_LE (literal "<=")
//   anything else      → consume only '<', emit TOKEN_LT (literal "<")
//
// This eager (maximal-munch) strategy — always consume the longest valid token
// — is the standard rule for lexers and prevents ambiguity.
//
// Unrecognised characters fall to the default case and emit TOKEN_ILLEGAL.
// The parser will then report an "unexpected token" error and attempt recovery.
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

// AllTokens lexes the entire input eagerly and returns the full token slice.
//
// This is a convenience method primarily used in tests and debug tools. It is
// NOT used by the parser in normal compilation — the parser calls NextToken()
// lazily (one token at a time), which keeps memory use proportional to the
// two-token window rather than the entire file size.
//
// AllTokens is useful for:
//   - Unit tests that need to assert on the exact token sequence
//   - Debuggers/pretty-printers that want to display all tokens at once
//   - Fuzz testing that needs a deterministic token count
//
// The loop always terminates because NextToken() is guaranteed to advance the
// cursor by at least one byte per call, and TOKEN_EOF is emitted when pos
// reaches the end of the input.
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

// isLetter returns true if ch is a Unicode letter.
//
// Using unicode.IsLetter (rather than 'a' <= ch <= 'z') means the lexer
// correctly handles identifiers with accented or non-ASCII letters, which is
// rare in BASIC but occasionally found in localised programs. The cost is a
// rune conversion per character, which is negligible for typical identifier
// lengths.
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

// isDigit returns true if ch is an ASCII decimal digit (0–9).
//
// An ASCII range check (ch >= '0' && ch <= '9') is faster than calling
// unicode.IsDigit because it avoids the overhead of Unicode classification.
// For decimal digits this is correct: BASIC only uses ASCII numerals in
// numeric literals.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isHexDigit returns true if ch is a valid hexadecimal digit (0–9, A–F, a–f).
//
// Hex digits are used exclusively inside &H… literals. Both upper-case and
// lower-case letters are accepted so that "&H1a" and "&H1A" are both valid —
// a concession to programmer convenience that costs nothing at this level.
func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
}
