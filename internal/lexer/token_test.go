package lexer

import (
	"strings"
	"testing"
)

// TestTokenPosition verifies that Token.Position() returns "line:col".
func TestTokenPosition(t *testing.T) {
	tok := Token{Type: TOKEN_PRINT, Literal: "PRINT", Line: 5, Column: 10}
	got := tok.Position()
	want := "5:10"
	if got != want {
		t.Errorf("Position(): got %q, want %q", got, want)
	}
}

// TestTokenPosition_FirstColumn verifies position at line 1, column 1.
func TestTokenPosition_FirstColumn(t *testing.T) {
	tok := Token{Type: TOKEN_EOF, Literal: "", Line: 1, Column: 1}
	got := tok.Position()
	want := "1:1"
	if got != want {
		t.Errorf("Position(): got %q, want %q", got, want)
	}
}

// TestTokenString verifies that Token.String() includes the token name, literal, and position.
func TestTokenString(t *testing.T) {
	tok := Token{Type: TOKEN_PRINT, Literal: "PRINT", Line: 1, Column: 1}
	s := tok.String()
	if !strings.Contains(s, "PRINT") {
		t.Errorf("String() %q does not contain token name PRINT", s)
	}
	if !strings.Contains(s, "1:1") {
		t.Errorf("String() %q does not contain position 1:1", s)
	}
}

// TestTokenString_WithLiteral verifies the quoted literal appears in Token.String().
func TestTokenString_WithLiteral(t *testing.T) {
	tok := Token{Type: TOKEN_STRING, Literal: "hello", Line: 3, Column: 7}
	s := tok.String()
	if !strings.Contains(s, `"hello"`) {
		t.Errorf("String() %q does not contain quoted literal \"hello\"", s)
	}
	if !strings.Contains(s, "3:7") {
		t.Errorf("String() %q does not contain position 3:7", s)
	}
}

// TestTokenName_KnownType verifies that TokenName returns a known name.
func TestTokenName_KnownType(t *testing.T) {
	got := TokenName(TOKEN_AND)
	want := "AND"
	if got != want {
		t.Errorf("TokenName(TOKEN_AND): got %q, want %q", got, want)
	}
}

// TestTokenName_UnknownType verifies that TokenName returns "UNKNOWN(n)" for unknown types.
func TestTokenName_UnknownType(t *testing.T) {
	unknownType := TokenType(99999)
	got := TokenName(unknownType)
	if !strings.HasPrefix(got, "UNKNOWN(") {
		t.Errorf("TokenName(99999): got %q, want prefix UNKNOWN(", got)
	}
}

// TestTokenName_AllMetaTokens verifies names for all metacommand token types.
func TestTokenName_AllMetaTokens(t *testing.T) {
	tests := []struct {
		typ  TokenType
		want string
	}{
		{TOKEN_META_DYNAMIC, "$DYNAMIC"},
		{TOKEN_META_STATIC, "$STATIC"},
		{TOKEN_META_INCLUDE, "$INCLUDE"},
		{TOKEN_META_IF, "$IF"},
		{TOKEN_META_ELSEIF, "$ELSEIF"},
		{TOKEN_META_ELSE, "$ELSE"},
		{TOKEN_META_ENDIF, "$ENDIF"},
		{TOKEN_META_COM, "$COM"},
		{TOKEN_META_SOUND, "$SOUND"},
		{TOKEN_META_STACK, "$STACK"},
		{TOKEN_META_SEGMENT, "$SEGMENT"},
		{TOKEN_META_INLINE, "$INLINE"},
		{TOKEN_META_EVENT, "$EVENT"},
	}
	for _, tt := range tests {
		got := TokenName(tt.typ)
		if got != tt.want {
			t.Errorf("TokenName(%d): got %q, want %q", tt.typ, got, tt.want)
		}
	}
}

// TestIsKeywordOperator_True verifies that AND/OR/NOT/XOR/EQV/IMP/MOD return true.
func TestIsKeywordOperator_True(t *testing.T) {
	keywordOps := []struct {
		typ  TokenType
		name string
	}{
		{TOKEN_AND, "AND"},
		{TOKEN_OR, "OR"},
		{TOKEN_NOT, "NOT"},
		{TOKEN_XOR, "XOR"},
		{TOKEN_EQV, "EQV"},
		{TOKEN_IMP, "IMP"},
		{TOKEN_MOD, "MOD"},
	}
	for _, tt := range keywordOps {
		if !IsKeywordOperator(tt.typ) {
			t.Errorf("IsKeywordOperator(%s): got false, want true", tt.name)
		}
	}
}

// TestIsKeywordOperator_False verifies that non-keyword-operator types return false.
func TestIsKeywordOperator_False(t *testing.T) {
	nonOps := []struct {
		typ  TokenType
		name string
	}{
		{TOKEN_PRINT, "PRINT"},
		{TOKEN_IF, "IF"},
		{TOKEN_PLUS, "PLUS"},
		{TOKEN_INTEGER, "INTEGER"},
		{TOKEN_IDENTIFIER, "IDENTIFIER"},
		{TOKEN_EOF, "EOF"},
	}
	for _, tt := range nonOps {
		if IsKeywordOperator(tt.typ) {
			t.Errorf("IsKeywordOperator(%s): got true, want false", tt.name)
		}
	}
}

// TestMetacommandDynamic verifies that '$DYNAMIC' at line start is lexed as TOKEN_META_DYNAMIC.
func TestMetacommandDynamic(t *testing.T) {
	l := New("$DYNAMIC")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_DYNAMIC {
		t.Errorf("expected TOKEN_META_DYNAMIC, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandStatic verifies that '$STATIC' at line start is lexed as TOKEN_META_STATIC.
func TestMetacommandStatic(t *testing.T) {
	l := New("$STATIC")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_STATIC {
		t.Errorf("expected TOKEN_META_STATIC, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandInclude verifies that '$INCLUDE' at line start is lexed as TOKEN_META_INCLUDE.
func TestMetacommandInclude(t *testing.T) {
	l := New("$INCLUDE")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_INCLUDE {
		t.Errorf("expected TOKEN_META_INCLUDE, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandIf verifies that '$IF' at line start is lexed as TOKEN_META_IF.
func TestMetacommandIf(t *testing.T) {
	l := New("$IF")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_IF {
		t.Errorf("expected TOKEN_META_IF, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandElseif verifies that '$ELSEIF' at line start is lexed as TOKEN_META_ELSEIF.
func TestMetacommandElseif(t *testing.T) {
	l := New("$ELSEIF")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_ELSEIF {
		t.Errorf("expected TOKEN_META_ELSEIF, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandElse verifies that '$ELSE' at line start is lexed as TOKEN_META_ELSE.
func TestMetacommandElse(t *testing.T) {
	l := New("$ELSE")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_ELSE {
		t.Errorf("expected TOKEN_META_ELSE, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandEndif verifies that '$ENDIF' at line start is lexed as TOKEN_META_ENDIF.
func TestMetacommandEndif(t *testing.T) {
	l := New("$ENDIF")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_ENDIF {
		t.Errorf("expected TOKEN_META_ENDIF, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandCom verifies that '$COM' at line start is lexed as TOKEN_META_COM.
func TestMetacommandCom(t *testing.T) {
	l := New("$COM")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_COM {
		t.Errorf("expected TOKEN_META_COM, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandSound verifies that '$SOUND' at line start is lexed as TOKEN_META_SOUND.
func TestMetacommandSound(t *testing.T) {
	l := New("$SOUND")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_SOUND {
		t.Errorf("expected TOKEN_META_SOUND, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandStack verifies that '$STACK' at line start is lexed as TOKEN_META_STACK.
func TestMetacommandStack(t *testing.T) {
	l := New("$STACK")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_STACK {
		t.Errorf("expected TOKEN_META_STACK, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandSegment verifies that '$SEGMENT' at line start is lexed as TOKEN_META_SEGMENT.
func TestMetacommandSegment(t *testing.T) {
	l := New("$SEGMENT")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_SEGMENT {
		t.Errorf("expected TOKEN_META_SEGMENT, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandInline verifies that '$INLINE' at line start is lexed as TOKEN_META_INLINE.
func TestMetacommandInline(t *testing.T) {
	l := New("$INLINE")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_INLINE {
		t.Errorf("expected TOKEN_META_INLINE, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandEvent verifies that '$EVENT' at line start is lexed as TOKEN_META_EVENT.
func TestMetacommandEvent(t *testing.T) {
	l := New("$EVENT")
	tok := l.NextToken()
	if tok.Type != TOKEN_META_EVENT {
		t.Errorf("expected TOKEN_META_EVENT, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandUnknown verifies that an unrecognised metacommand like '$XYZ'
// produces TOKEN_ILLEGAL.
func TestMetacommandUnknown(t *testing.T) {
	l := New("$XYZ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("expected TOKEN_ILLEGAL for unknown metacommand, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestMetacommandLiteral verifies that the Literal field of a metacommand token
// contains the '$'-prefixed uppercase word.
func TestMetacommandLiteral(t *testing.T) {
	l := New("$DYNAMIC")
	tok := l.NextToken()
	if tok.Literal != "$DYNAMIC" {
		t.Errorf("metacommand literal: got %q, want %q", tok.Literal, "$DYNAMIC")
	}
}

// TestMetacommandAfterNewline verifies that a metacommand on the second line is
// still recognised (prevTokenEOL is set after TOKEN_EOL).
func TestMetacommandAfterNewline(t *testing.T) {
	l := New("\n$STATIC")
	l.NextToken() // EOL
	tok := l.NextToken()
	if tok.Type != TOKEN_META_STATIC {
		t.Errorf("expected TOKEN_META_STATIC after newline, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}
