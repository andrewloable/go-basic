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
