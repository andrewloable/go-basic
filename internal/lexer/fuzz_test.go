package lexer

import "testing"

// FuzzLexer tests that the lexer never panics on arbitrary input.
func FuzzLexer(f *testing.F) {
	// Seed corpus with valid BASIC fragments
	seeds := []string{
		`10 PRINT "Hello World"`,
		`PRINT "test"`,
		`LET x = 42`,
		`IF x > 0 THEN PRINT "yes"`,
		`FOR i = 1 TO 10 : NEXT i`,
		`WHILE x < 100 : WEND`,
		`DO : LOOP UNTIL x > 0`,
		`SELECT CASE x : CASE 1 : END SELECT`,
		`DIM a(10), b$(5,5)`,
		`GOSUB label : GOTO 100`,
		`REM this is a comment`,
		`' single quote comment`,
		`x% = 1 : y& = 2 : z! = 3.14 : w# = 2.718 : n$ = "test"`,
		`&HFF : &O77 : &B1010`,
		`3.14E+10 : 1.23D-5`,
		`$DYNAMIC`,
		`$INCLUDE`,
		`"embedded ""quotes"""`,
		`DATA 1, "two", 3.14`,
		`ON ERROR GOTO handler`,
		`PRINT USING "###.##"; x`,
		`SUB MySub (x, y) : END SUB`,
		`FUNCTION Add(a, b) : END FUNCTION`,
		`DEFINT A-Z`,
		`OPEN "file" FOR INPUT AS #1`,
		``,
		"\n\n\n",
		"+ - * / \\ ^ = <> < > <= >=",
		"( ) , ; : # $ % & !",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		l := New(input)
		tokens := l.AllTokens()

		// Must always end with EOF
		if len(tokens) == 0 {
			t.Fatal("AllTokens returned empty slice")
		}
		last := tokens[len(tokens)-1]
		if last.Type != TOKEN_EOF {
			t.Fatalf("last token is %s, want EOF", TokenName(last.Type))
		}

		// All tokens must have valid positions
		for i, tok := range tokens {
			if tok.Line < 0 {
				t.Fatalf("token[%d] has negative line: %d", i, tok.Line)
			}
		}
	})
}
