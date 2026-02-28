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
	// Use a TokenType value that is definitely not in tokenNames.
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

// TestReadNewlineCRLF verifies that a CR+LF sequence produces a single TOKEN_EOL.
func TestReadNewlineCRLF(t *testing.T) {
	l := New("\r\n")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOL {
		t.Errorf("CRLF: expected TOKEN_EOL, got %s", TokenName(tok.Type))
	}
	// After consuming CR+LF, next token should be EOF.
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
	l.NextToken() // A (identifier)
	l.NextToken() // EOL (CRLF)
	tok := l.NextToken() // B
	if tok.Line != 2 {
		t.Errorf("after CRLF: expected line 2, got line %d", tok.Line)
	}
}

// TestReadNewlineCR_LineIncrement verifies that line number advances after CR-only.
func TestReadNewlineCR_LineIncrement(t *testing.T) {
	l := New("A\rB")
	l.NextToken() // A (identifier)
	l.NextToken() // EOL (CR)
	tok := l.NextToken() // B
	if tok.Line != 2 {
		t.Errorf("after CR-only: expected line 2, got line %d", tok.Line)
	}
}

// TestReadStringEmpty verifies that an empty string literal ("") produces
// TOKEN_STRING with an empty Literal.
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

// TestReadStringUnterminated verifies that a string with no closing quote
// before EOF produces TOKEN_ILLEGAL.
func TestReadStringUnterminated(t *testing.T) {
	l := New(`"hello`)
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("unterminated string: expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
}

// TestReadStringUnterminated_AtNewline verifies that a string interrupted by a
// newline produces TOKEN_ILLEGAL.
func TestReadStringUnterminated_AtNewline(t *testing.T) {
	l := New("\"hello\n")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("newline-terminated string: expected TOKEN_ILLEGAL, got %s", TokenName(tok.Type))
	}
	// The partial literal should be preserved.
	if tok.Literal != "hello" {
		t.Errorf("newline-terminated string: expected literal %q, got %q", "hello", tok.Literal)
	}
}

// TestReadStringUnterminated_AtCR verifies that a string interrupted by CR
// produces TOKEN_ILLEGAL.
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

// TestReadSpecialRadix_HexNoDigits verifies &H with no following hex digits
// produces TOKEN_ILLEGAL.
func TestReadSpecialRadix_HexNoDigits(t *testing.T) {
	l := New("&H ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&H (no digits): expected TOKEN_ILLEGAL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestReadSpecialRadix_OctalNoDigits verifies &O with no following octal digits
// produces TOKEN_ILLEGAL.
func TestReadSpecialRadix_OctalNoDigits(t *testing.T) {
	l := New("&O ")
	tok := l.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("&O (no digits): expected TOKEN_ILLEGAL, got %s (%q)", TokenName(tok.Type), tok.Literal)
	}
}

// TestReadSpecialRadix_BinaryNoDigits verifies &B with no following binary digits
// produces TOKEN_ILLEGAL.
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
