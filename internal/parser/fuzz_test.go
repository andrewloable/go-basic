package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
)

// FuzzParser tests that the parser never panics on arbitrary input.
func FuzzParser(f *testing.F) {
	// Seed corpus with valid and semi-valid BASIC programs
	seeds := []string{
		`PRINT "Hello World"`,
		`10 PRINT "Hello"`,
		`LET x = 42`,
		`x = 1 + 2 * 3`,
		`IF x > 0 THEN PRINT "yes" ELSE PRINT "no"`,
		`IF x > 0 THEN
PRINT "yes"
ELSE
PRINT "no"
END IF`,
		`FOR i = 1 TO 10
PRINT i
NEXT i`,
		`FOR i = 10 TO 1 STEP -1
PRINT i
NEXT i`,
		`WHILE x < 10
x = x + 1
WEND`,
		`DO WHILE x < 10
x = x + 1
LOOP`,
		`DO
x = x + 1
LOOP UNTIL x > 10`,
		`SELECT CASE x
CASE 1
PRINT "one"
CASE 2, 3
PRINT "two or three"
CASE 4 TO 6
PRINT "four to six"
CASE ELSE
PRINT "other"
END SELECT`,
		`DIM a(10)
DIM b$(5, 5)`,
		`SUB MySub (x, y)
PRINT x + y
END SUB`,
		`FUNCTION Add(a, b)
Add = a + b
END FUNCTION`,
		`GOTO 100
GOSUB label
RETURN`,
		`DATA 1, 2, 3
READ a, b, c
RESTORE`,
		`ON ERROR GOTO handler
RESUME NEXT`,
		`SWAP a, b`,
		`INCR x
DECR y, 5`,
		`PRINT ABS(-5); SQR(16); LEFT$("hello", 3)`,
		`OPEN "test.dat" FOR OUTPUT AS #1
PRINT #1, "hello"
CLOSE #1`,
		`END`,
		// Edge cases
		``,
		`::::::`,
		`PRINT`,
		`IF THEN`,
		`FOR TO NEXT`,
		`((((x))))`,
		`x = -(-(-1))`,
		`x = NOT NOT NOT 1`,
		`x = 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10`,
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()

		// Must always return a non-nil Program
		if program == nil {
			t.Fatal("ParseProgram returned nil")
		}

		// Errors are OK (invalid input), but no panics
		_ = p.Errors()
	})
}
