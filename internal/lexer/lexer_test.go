package lexer

import (
	"testing"
)

func TestNextToken_HelloWorld(t *testing.T) {
	input := `10 PRINT "HELLO WORLD"
20 END`

	l := New(input)
	expected := []struct {
		typ     TokenType
		literal string
	}{
		{TOKEN_INTEGER, "10"},
		{TOKEN_PRINT, "PRINT"},
		{TOKEN_STRING, "HELLO WORLD"},
		{TOKEN_EOL, "\n"},
		{TOKEN_INTEGER, "20"},
		{TOKEN_END, "END"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. got=%s, want=%s (literal=%q)",
				i, TokenName(tok.Type), TokenName(exp.typ), tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("test[%d]: literal wrong. got=%q, want=%q",
				i, tok.Literal, exp.literal)
		}
	}
}

func TestNextToken_Operators(t *testing.T) {
	input := `+ - * / \ ^ = <> < > <= >=`

	l := New(input)
	expected := []struct {
		typ     TokenType
		literal string
	}{
		{TOKEN_PLUS, "+"},
		{TOKEN_MINUS, "-"},
		{TOKEN_STAR, "*"},
		{TOKEN_SLASH, "/"},
		{TOKEN_BACKSLASH, "\\"},
		{TOKEN_CARET, "^"},
		{TOKEN_EQ, "="},
		{TOKEN_NE, "<>"},
		{TOKEN_LT, "<"},
		{TOKEN_GT, ">"},
		{TOKEN_LE, "<="},
		{TOKEN_GE, ">="},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("test[%d]: type wrong. got=%s, want=%s",
				i, TokenName(tok.Type), TokenName(exp.typ))
		}
		if tok.Literal != exp.literal {
			t.Errorf("test[%d]: literal wrong. got=%q, want=%q",
				i, tok.Literal, exp.literal)
		}
	}
}

func TestNextToken_Numbers(t *testing.T) {
	tests := []struct {
		input   string
		typ     TokenType
		literal string
	}{
		{"42", TOKEN_INTEGER, "42"},
		{"42&", TOKEN_LONG, "42"},
		{"3.14", TOKEN_SINGLE, "3.14"},
		{"3.14!", TOKEN_SINGLE, "3.14"},
		{"3.14#", TOKEN_DOUBLE, "3.14"},
		{"2.5E-3", TOKEN_SINGLE, "2.5E-3"},
		{"1.23D+5", TOKEN_DOUBLE, "1.23D+5"},
		{".5", TOKEN_SINGLE, ".5"},
		{"&HFF", TOKEN_HEX, "FF"},
		{"&H1a", TOKEN_HEX, "1a"},
		{"&O77", TOKEN_OCTAL, "77"},
		{"&B1010", TOKEN_BINARY_LIT, "1010"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Errorf("input %q: type wrong. got=%s, want=%s",
				tt.input, TokenName(tok.Type), TokenName(tt.typ))
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: literal wrong. got=%q, want=%q",
				tt.input, tok.Literal, tt.literal)
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
	}{
		{"PRINT", TOKEN_PRINT},
		{"print", TOKEN_PRINT},
		{"Print", TOKEN_PRINT},
		{"IF", TOKEN_IF},
		{"THEN", TOKEN_THEN},
		{"ELSE", TOKEN_ELSE},
		{"FOR", TOKEN_FOR},
		{"NEXT", TOKEN_NEXT},
		{"WHILE", TOKEN_WHILE},
		{"WEND", TOKEN_WEND},
		{"DO", TOKEN_DO},
		{"LOOP", TOKEN_LOOP},
		{"AND", TOKEN_AND},
		{"OR", TOKEN_OR},
		{"NOT", TOKEN_NOT},
		{"DIM", TOKEN_DIM},
		{"SUB", TOKEN_SUB},
		{"FUNCTION", TOKEN_FUNCTION},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Errorf("input %q: type wrong. got=%s, want=%s",
				tt.input, TokenName(tok.Type), TokenName(tt.typ))
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`"hello"`, "hello"},
		{`""`, ""},
		{`"hello ""world"""`, `hello "world"`},
		{`"spaces   here"`, "spaces   here"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TOKEN_STRING {
			t.Errorf("input %q: type wrong. got=%s, want=STRING",
				tt.input, TokenName(tok.Type))
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: literal wrong. got=%q, want=%q",
				tt.input, tok.Literal, tt.literal)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `PRINT "hi" ' this is a comment
REM this is also a comment`

	l := New(input)
	// PRINT
	tok := l.NextToken()
	if tok.Type != TOKEN_PRINT {
		t.Errorf("expected PRINT, got %s", TokenName(tok.Type))
	}
	// "hi"
	tok = l.NextToken()
	if tok.Type != TOKEN_STRING {
		t.Errorf("expected STRING, got %s", TokenName(tok.Type))
	}
	// ' comment
	tok = l.NextToken()
	if tok.Type != TOKEN_COMMENT {
		t.Errorf("expected COMMENT, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
	// EOL
	tok = l.NextToken()
	if tok.Type != TOKEN_EOL {
		t.Errorf("expected EOL, got %s", TokenName(tok.Type))
	}
	// REM comment
	tok = l.NextToken()
	if tok.Type != TOKEN_COMMENT {
		t.Errorf("expected COMMENT for REM, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

func TestNextToken_TypeSuffixes(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"count%", "count%"},
		{"total&", "total&"},
		{"pi!", "pi!"},
		{"value#", "value#"},
		{"name$", "name$"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TOKEN_IDENTIFIER {
			t.Errorf("input %q: type wrong. got=%s, want=IDENTIFIER",
				tt.input, TokenName(tok.Type))
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: literal wrong. got=%q, want=%q",
				tt.input, tok.Literal, tt.literal)
		}
	}
}

func TestNextToken_Labels(t *testing.T) {
	input := "Start:\n    PRINT \"hello\""
	l := New(input)

	tok := l.NextToken()
	if tok.Type != TOKEN_LABEL {
		t.Errorf("expected LABEL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
	if tok.Literal != "Start" {
		t.Errorf("expected literal 'Start', got %q", tok.Literal)
	}
}

func TestNextToken_MultiStatementLine(t *testing.T) {
	input := `x = 1 : y = 2`
	l := New(input)

	expected := []TokenType{
		TOKEN_IDENTIFIER, TOKEN_EQ, TOKEN_INTEGER,
		TOKEN_COLON,
		TOKEN_IDENTIFIER, TOKEN_EQ, TOKEN_INTEGER,
		TOKEN_EOF,
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("test[%d]: got=%s, want=%s", i, TokenName(tok.Type), TokenName(exp))
		}
	}
}

func TestNextToken_Metacommands(t *testing.T) {
	input := "$DYNAMIC\n$STATIC\n$INCLUDE"

	l := New(input)

	tok := l.NextToken()
	if tok.Type != TOKEN_META_DYNAMIC {
		t.Errorf("expected $DYNAMIC, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}

	l.NextToken() // EOL

	tok = l.NextToken()
	if tok.Type != TOKEN_META_STATIC {
		t.Errorf("expected $STATIC, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}

	l.NextToken() // EOL

	tok = l.NextToken()
	if tok.Type != TOKEN_META_INCLUDE {
		t.Errorf("expected $INCLUDE, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

func TestNextToken_ComplexProgram(t *testing.T) {
	input := `10 FOR i = 1 TO 10
20     PRINT i; " "; i * i
30 NEXT i
40 END`

	l := New(input)
	tokens := l.AllTokens()

	// Just verify it doesn't panic and produces tokens
	if len(tokens) < 10 {
		t.Errorf("expected at least 10 tokens, got %d", len(tokens))
	}

	// Last token should be EOF
	last := tokens[len(tokens)-1]
	if last.Type != TOKEN_EOF {
		t.Errorf("expected last token to be EOF, got %s", TokenName(last.Type))
	}
}

func TestNextToken_LineNumbers(t *testing.T) {
	input := "10 PRINT\n20 END"
	l := New(input)

	// Line numbers at start of line are just integers for now
	// (the parser will distinguish them from expressions)
	tok := l.NextToken()
	if tok.Type != TOKEN_INTEGER {
		t.Errorf("expected INTEGER for line number, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "10" {
		t.Errorf("expected '10', got %q", tok.Literal)
	}
}

func TestNextToken_Punctuation(t *testing.T) {
	input := "( ) , ; : #"
	l := New(input)

	expected := []TokenType{
		TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_COMMA,
		TOKEN_SEMICOLON, TOKEN_COLON, TOKEN_HASH,
		TOKEN_EOF,
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("test[%d]: got=%s, want=%s", i, TokenName(tok.Type), TokenName(exp))
		}
	}
}

func TestNextToken_EmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("expected EOF for empty input, got %s", TokenName(tok.Type))
	}
}

func TestNextToken_PositionTracking(t *testing.T) {
	input := "PRINT \"hello\"\nLET x = 5"
	l := New(input)

	tok := l.NextToken() // PRINT
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("PRINT: expected 1:1, got %d:%d", tok.Line, tok.Column)
	}

	tok = l.NextToken() // "hello"
	if tok.Line != 1 || tok.Column != 7 {
		t.Errorf("\"hello\": expected 1:7, got %d:%d", tok.Line, tok.Column)
	}

	tok = l.NextToken() // EOL
	tok = l.NextToken() // LET
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("LET: expected 2:1, got %d:%d", tok.Line, tok.Column)
	}
}

// TestReadNewlineCRLF verifies that a CR+LF sequence produces a single TOKEN_EOL.
func TestReadNewlineCRLF(t *testing.T) {
	l := New("\r\n")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOL {
		t.Errorf("CRLF: expected TOKEN_EOL, got %s", TokenName(tok.Type))
	}
	tok = l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("CRLF: expected TOKEN_EOF after EOL, got %s", TokenName(tok.Type))
	}
}

// TestReadNewlineCROnly verifies that a lone CR (no following LF) produces TOKEN_EOL.
func TestReadNewlineCROnly(t *testing.T) {
	l := New("\r")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOL {
		t.Errorf("CR-only: expected TOKEN_EOL, got %s", TokenName(tok.Type))
	}
	tok = l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("CR-only: expected TOKEN_EOF after EOL, got %s", TokenName(tok.Type))
	}
}

// TestReadNewlineCRLF_LineIncrement verifies that line number advances after CRLF.
func TestReadNewlineCRLF_LineIncrement(t *testing.T) {
	l := New("A\r\nB")
	l.NextToken()        // A (identifier)
	l.NextToken()        // EOL (CRLF)
	tok := l.NextToken() // B
	if tok.Line != 2 {
		t.Errorf("after CRLF: expected line 2, got line %d", tok.Line)
	}
}

// TestReadNewlineCR_LineIncrement verifies that line number advances after CR-only.
func TestReadNewlineCR_LineIncrement(t *testing.T) {
	l := New("A\rB")
	l.NextToken()        // A (identifier)
	l.NextToken()        // EOL (CR)
	tok := l.NextToken() // B
	if tok.Line != 2 {
		t.Errorf("after CR-only: expected line 2, got line %d", tok.Line)
	}
}

// TestReadStringEmpty verifies that an empty string literal ("") produces TOKEN_STRING.
func TestReadStringEmpty(t *testing.T) {
	l := New(`""`)
	tok := l.NextToken()
	if tok.Type != TOKEN_STRING {
		t.Errorf("empty string: expected TOKEN_STRING, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "" {
		t.Errorf("empty string: expected empty literal, got %q", tok.Literal)
	}
}

// TestReadStringUnterminated verifies that a string with no closing quote produces TOKEN_ILLEGAL.
func TestReadStringUnterminated(t *testing.T) {
	l := New(`"hello`)
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("unterminated string: expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
}

// TestReadStringUnterminated_AtNewline verifies that a string interrupted by newline produces TOKEN_ILLEGAL.
func TestReadStringUnterminated_AtNewline(t *testing.T) {
	l := New("\"hello\n")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("newline-terminated string: expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "hello" {
		t.Errorf("newline-terminated string: expected literal %q, got %q", "hello", tok.Literal)
	}
}

// TestReadStringUnterminated_AtCR verifies that a string interrupted by CR produces TOKEN_ILLEGAL.
func TestReadStringUnterminated_AtCR(t *testing.T) {
	l := New("\"world\r")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("CR-terminated string: expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
}

// TestReadNumber_IntegerPercent verifies that '42%' is lexed as TOKEN_INTEGER.
func TestReadNumber_IntegerPercent(t *testing.T) {
	l := New("42%")
	tok := l.NextToken()
	if tok.Type != TOKEN_INTEGER {
		t.Errorf("42%%: expected TOKEN_INTEGER, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "42" {
		t.Errorf("42%%: expected literal %q, got %q", "42", tok.Literal)
	}
}

// TestReadNumber_LongAmpersand verifies that '100&' is lexed as TOKEN_LONG.
func TestReadNumber_LongAmpersand(t *testing.T) {
	l := New("100&")
	tok := l.NextToken()
	if tok.Type != TOKEN_LONG {
		t.Errorf("100&: expected TOKEN_LONG, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "100" {
		t.Errorf("100&: expected literal %q, got %q", "100", tok.Literal)
	}
}

// TestReadNumber_SingleExclaim verifies that '1.5!' is lexed as TOKEN_SINGLE.
func TestReadNumber_SingleExclaim(t *testing.T) {
	l := New("1.5!")
	tok := l.NextToken()
	if tok.Type != TOKEN_SINGLE {
		t.Errorf("1.5!: expected TOKEN_SINGLE, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "1.5" {
		t.Errorf("1.5!: expected literal %q, got %q", "1.5", tok.Literal)
	}
}

// TestReadNumber_DoubleHash verifies that '3.14#' is lexed as TOKEN_DOUBLE.
func TestReadNumber_DoubleHash(t *testing.T) {
	l := New("3.14#")
	tok := l.NextToken()
	if tok.Type != TOKEN_DOUBLE {
		t.Errorf("3.14#: expected TOKEN_DOUBLE, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "3.14" {
		t.Errorf("3.14#: expected literal %q, got %q", "3.14", tok.Literal)
	}
}

// TestReadNumber_DoubleExponent verifies that 'D' exponent produces TOKEN_DOUBLE.
func TestReadNumber_DoubleExponent(t *testing.T) {
	l := New("1.23D+5")
	tok := l.NextToken()
	if tok.Type != TOKEN_DOUBLE {
		t.Errorf("1.23D+5: expected TOKEN_DOUBLE, got %s", TokenName(tok.Type))
	}
}

// TestReadNumber_LeadingDot verifies that '.5' is lexed as TOKEN_SINGLE.
func TestReadNumber_LeadingDot(t *testing.T) {
	l := New(".5")
	tok := l.NextToken()
	if tok.Type != TOKEN_SINGLE {
		t.Errorf(".5: expected TOKEN_SINGLE, got %s", TokenName(tok.Type))
	}
	if tok.Literal != ".5" {
		t.Errorf(".5: expected literal %q, got %q", ".5", tok.Literal)
	}
}

// TestReadSpecialRadix_HexUppercase verifies &HFF produces TOKEN_HEX with literal "FF".
func TestReadSpecialRadix_HexUppercase(t *testing.T) {
	l := New("&HFF")
	tok := l.NextToken()
	if tok.Type != TOKEN_HEX {
		t.Errorf("&HFF: expected TOKEN_HEX, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "FF" {
		t.Errorf("&HFF: expected literal %q, got %q", "FF", tok.Literal)
	}
}

// TestReadSpecialRadix_HexLowercase verifies &h1a produces TOKEN_HEX with literal "1a".
func TestReadSpecialRadix_HexLowercase(t *testing.T) {
	l := New("&h1a")
	tok := l.NextToken()
	if tok.Type != TOKEN_HEX {
		t.Errorf("&h1a: expected TOKEN_HEX, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "1a" {
		t.Errorf("&h1a: expected literal %q, got %q", "1a", tok.Literal)
	}
}

// TestReadSpecialRadix_OctalUppercase verifies &O77 produces TOKEN_OCTAL with literal "77".
func TestReadSpecialRadix_OctalUppercase(t *testing.T) {
	l := New("&O77")
	tok := l.NextToken()
	if tok.Type != TOKEN_OCTAL {
		t.Errorf("&O77: expected TOKEN_OCTAL, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "77" {
		t.Errorf("&O77: expected literal %q, got %q", "77", tok.Literal)
	}
}

// TestReadSpecialRadix_OctalLowercase verifies &o17 produces TOKEN_OCTAL with literal "17".
func TestReadSpecialRadix_OctalLowercase(t *testing.T) {
	l := New("&o17")
	tok := l.NextToken()
	if tok.Type != TOKEN_OCTAL {
		t.Errorf("&o17: expected TOKEN_OCTAL, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "17" {
		t.Errorf("&o17: expected literal %q, got %q", "17", tok.Literal)
	}
}

// TestReadSpecialRadix_BinaryUppercase verifies &B1010 produces TOKEN_BINARY_LIT.
func TestReadSpecialRadix_BinaryUppercase(t *testing.T) {
	l := New("&B1010")
	tok := l.NextToken()
	if tok.Type != TOKEN_BINARY_LIT {
		t.Errorf("&B1010: expected TOKEN_BINARY_LIT, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "1010" {
		t.Errorf("&B1010: expected literal %q, got %q", "1010", tok.Literal)
	}
}

// TestReadSpecialRadix_BinaryLowercase verifies &b101 produces TOKEN_BINARY_LIT.
func TestReadSpecialRadix_BinaryLowercase(t *testing.T) {
	l := New("&b101")
	tok := l.NextToken()
	if tok.Type != TOKEN_BINARY_LIT {
		t.Errorf("&b101: expected TOKEN_BINARY_LIT, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "101" {
		t.Errorf("&b101: expected literal %q, got %q", "101", tok.Literal)
	}
}

// TestReadSpecialRadix_HexNoDigits verifies &H with no following hex digits produces TOKEN_ILLEGAL.
func TestReadSpecialRadix_HexNoDigits(t *testing.T) {
	l := New("&H ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&H (no digits): expected TOKEN_ILLEGAL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestReadSpecialRadix_OctalNoDigits verifies &O with no following octal digits produces TOKEN_ILLEGAL.
func TestReadSpecialRadix_OctalNoDigits(t *testing.T) {
	l := New("&O ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&O (no digits): expected TOKEN_ILLEGAL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestReadSpecialRadix_BinaryNoDigits verifies &B with no following binary digits produces TOKEN_ILLEGAL.
func TestReadSpecialRadix_BinaryNoDigits(t *testing.T) {
	l := New("&B ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&B (no digits): expected TOKEN_ILLEGAL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestReadSpecialRadix_IllegalLiteralContent verifies that when no digits follow
// the radix letter, the Literal field contains the '&' prefix plus the radix char.
func TestReadSpecialRadix_IllegalLiteralContent(t *testing.T) {
	l := New("&H")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&H (EOF): expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
	if tok.Literal != "&H" {
		t.Errorf("&H (EOF): expected literal %q, got %q", "&H", tok.Literal)
	}
}
