package lexer

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_EOL
	TOKEN_COMMENT

	// Identifiers and literals
	TOKEN_IDENTIFIER
	TOKEN_INTEGER    // 42
	TOKEN_LONG       // 42&
	TOKEN_SINGLE     // 3.14 or 3.14!
	TOKEN_DOUBLE     // 3.14159265#
	TOKEN_STRING     // "hello"
	TOKEN_HEX        // &H1A
	TOKEN_OCTAL      // &O77
	TOKEN_BINARY_LIT // &B1010

	// Line structure
	TOKEN_LINENUMBER // 10, 20, 100 at start of line
	TOKEN_LABEL      // MyLabel:

	// Arithmetic operators
	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_STAR      // *
	TOKEN_SLASH     // /
	TOKEN_BACKSLASH // \
	TOKEN_CARET     // ^

	// Relational operators
	TOKEN_EQ // =
	TOKEN_NE // <>
	TOKEN_LT // <
	TOKEN_GT // >
	TOKEN_LE // <=
	TOKEN_GE // >=

	// Punctuation
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_COMMA     // ,
	TOKEN_SEMICOLON // ;
	TOKEN_COLON     // :
	TOKEN_HASH      // #
	TOKEN_DOLLAR    // $ (standalone)
	TOKEN_PERCENT   // % (standalone)
	TOKEN_AMPERSAND // & (standalone)
	TOKEN_EXCLAIM   // ! (standalone)

	// Type suffix tokens (attached to identifiers)
	TOKEN_SUFFIX_PERCENT   // % on identifier
	TOKEN_SUFFIX_AMPERSAND // & on identifier
	TOKEN_SUFFIX_EXCLAIM   // ! on identifier
	TOKEN_SUFFIX_HASH      // # on identifier
	TOKEN_SUFFIX_DOLLAR    // $ on identifier

	// ---- Keywords ----

	// I/O keywords
	TOKEN_PRINT
	TOKEN_LPRINT
	TOKEN_INPUT
	TOKEN_LINE // LINE (as in LINE INPUT, LINE graphics)
	TOKEN_WRITE
	TOKEN_USING
	TOKEN_LOCATE
	TOKEN_CLS
	TOKEN_COLOR
	TOKEN_WIDTH
	TOKEN_KEY
	TOKEN_BEEP
	TOKEN_SOUND
	TOKEN_PLAY

	// Assignment
	TOKEN_LET

	// Control flow
	TOKEN_IF
	TOKEN_THEN
	TOKEN_ELSE
	TOKEN_ELSEIF
	TOKEN_END
	TOKEN_FOR
	TOKEN_TO
	TOKEN_STEP
	TOKEN_NEXT
	TOKEN_WHILE
	TOKEN_WEND
	TOKEN_DO
	TOKEN_LOOP
	TOKEN_UNTIL
	TOKEN_SELECT
	TOKEN_CASE
	TOKEN_IS
	TOKEN_GOTO
	TOKEN_GOSUB
	TOKEN_RETURN
	TOKEN_EXIT
	TOKEN_STOP
	TOKEN_SYSTEM
	TOKEN_ON
	TOKEN_RUN
	TOKEN_CHAIN
	TOKEN_SHELL

	// Logical/bitwise keyword operators
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_XOR
	TOKEN_EQV
	TOKEN_IMP
	TOKEN_MOD

	// Declaration keywords
	TOKEN_DIM
	TOKEN_REDIM
	TOKEN_AS
	TOKEN_SUB
	TOKEN_FUNCTION
	TOKEN_DECLARE
	TOKEN_CALL
	TOKEN_DEF
	TOKEN_FN
	TOKEN_BYVAL
	TOKEN_LOCAL
	TOKEN_SHARED
	TOKEN_STATIC
	TOKEN_COMMON
	TOKEN_OPTION
	TOKEN_BASE

	// Type keywords
	TOKEN_INTEGER_KW // INTEGER
	TOKEN_LONG_KW    // LONG
	TOKEN_SINGLE_KW  // SINGLE
	TOKEN_DOUBLE_KW  // DOUBLE
	TOKEN_STRING_KW  // STRING

	// Default type keywords
	TOKEN_DEFINT
	TOKEN_DEFLNG
	TOKEN_DEFSNG
	TOKEN_DEFDBL
	TOKEN_DEFSTR

	// Array/variable keywords
	TOKEN_ERASE
	TOKEN_SWAP
	TOKEN_INCR
	TOKEN_DECR
	TOKEN_LBOUND
	TOKEN_UBOUND

	// Data keywords
	TOKEN_DATA
	TOKEN_READ
	TOKEN_RESTORE
	TOKEN_REM

	// Error handling
	TOKEN_ERROR
	TOKEN_RESUME

	// File I/O keywords
	TOKEN_OPEN
	TOKEN_CLOSE
	TOKEN_OUTPUT_KW
	TOKEN_APPEND
	TOKEN_RANDOM_KW
	TOKEN_BINARY_KW
	TOKEN_FIELD
	TOKEN_GET
	TOKEN_PUT
	TOKEN_SEEK
	TOKEN_LSET
	TOKEN_RSET
	TOKEN_KILL
	TOKEN_NAME
	TOKEN_FILES
	TOKEN_CHDIR
	TOKEN_MKDIR
	TOKEN_RMDIR
	TOKEN_BLOAD
	TOKEN_BSAVE
	TOKEN_IOCTL
	TOKEN_LEN_KW // LEN keyword (as in LEN = reclen)
	TOKEN_EOF_KW // EOF keyword/function

	// Graphics keywords
	TOKEN_SCREEN
	TOKEN_PSET
	TOKEN_PRESET
	TOKEN_CIRCLE
	TOKEN_PAINT
	TOKEN_DRAW
	TOKEN_VIEW
	TOKEN_WINDOW
	TOKEN_PALETTE

	// Memory/system keywords
	TOKEN_PEEK
	TOKEN_POKE
	TOKEN_INP
	TOKEN_OUT
	TOKEN_WAIT
	TOKEN_VARPTR
	TOKEN_VARSEG

	// Tracing
	TOKEN_TRON
	TOKEN_TROFF

	// Timing
	TOKEN_TIMER
	TOKEN_DELAY
	TOKEN_MTIMER

	// Misc keywords
	TOKEN_TAB
	TOKEN_SPC
	TOKEN_RANDOMIZE
	TOKEN_INSTAT

	// Metacommands
	TOKEN_META_DYNAMIC  // $DYNAMIC
	TOKEN_META_STATIC   // $STATIC
	TOKEN_META_INCLUDE  // $INCLUDE
	TOKEN_META_IF       // $IF
	TOKEN_META_ELSEIF   // $ELSEIF
	TOKEN_META_ELSE     // $ELSE
	TOKEN_META_ENDIF    // $ENDIF
	TOKEN_META_COM      // $COM
	TOKEN_META_SOUND    // $SOUND
	TOKEN_META_STACK    // $STACK
	TOKEN_META_SEGMENT  // $SEGMENT
	TOKEN_META_INLINE   // $INLINE
	TOKEN_META_EVENT    // $EVENT
)

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Position returns a human-readable position string.
func (t Token) Position() string {
	return fmt.Sprintf("%d:%d", t.Line, t.Column)
}

// String returns a human-readable token representation.
func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %s", tokenNames[t.Type], t.Literal, t.Position())
}

// keywords maps uppercase keyword strings to their token types.
var keywords = map[string]TokenType{
	"PRINT":     TOKEN_PRINT,
	"LPRINT":    TOKEN_LPRINT,
	"INPUT":     TOKEN_INPUT,
	"LINE":      TOKEN_LINE,
	"WRITE":     TOKEN_WRITE,
	"USING":     TOKEN_USING,
	"LOCATE":    TOKEN_LOCATE,
	"CLS":       TOKEN_CLS,
	"COLOR":     TOKEN_COLOR,
	"WIDTH":     TOKEN_WIDTH,
	"KEY":       TOKEN_KEY,
	"BEEP":      TOKEN_BEEP,
	"SOUND":     TOKEN_SOUND,
	"PLAY":      TOKEN_PLAY,
	"LET":       TOKEN_LET,
	"IF":        TOKEN_IF,
	"THEN":      TOKEN_THEN,
	"ELSE":      TOKEN_ELSE,
	"ELSEIF":    TOKEN_ELSEIF,
	"END":       TOKEN_END,
	"FOR":       TOKEN_FOR,
	"TO":        TOKEN_TO,
	"STEP":      TOKEN_STEP,
	"NEXT":      TOKEN_NEXT,
	"WHILE":     TOKEN_WHILE,
	"WEND":      TOKEN_WEND,
	"DO":        TOKEN_DO,
	"LOOP":      TOKEN_LOOP,
	"UNTIL":     TOKEN_UNTIL,
	"SELECT":    TOKEN_SELECT,
	"CASE":      TOKEN_CASE,
	"IS":        TOKEN_IS,
	"GOTO":      TOKEN_GOTO,
	"GOSUB":     TOKEN_GOSUB,
	"RETURN":    TOKEN_RETURN,
	"EXIT":      TOKEN_EXIT,
	"STOP":      TOKEN_STOP,
	"SYSTEM":    TOKEN_SYSTEM,
	"ON":        TOKEN_ON,
	"RUN":       TOKEN_RUN,
	"CHAIN":     TOKEN_CHAIN,
	"SHELL":     TOKEN_SHELL,
	"AND":       TOKEN_AND,
	"OR":        TOKEN_OR,
	"NOT":       TOKEN_NOT,
	"XOR":       TOKEN_XOR,
	"EQV":       TOKEN_EQV,
	"IMP":       TOKEN_IMP,
	"MOD":       TOKEN_MOD,
	"DIM":       TOKEN_DIM,
	"REDIM":     TOKEN_REDIM,
	"AS":        TOKEN_AS,
	"SUB":       TOKEN_SUB,
	"FUNCTION":  TOKEN_FUNCTION,
	"DECLARE":   TOKEN_DECLARE,
	"CALL":      TOKEN_CALL,
	"DEF":       TOKEN_DEF,
	"FN":        TOKEN_FN,
	"BYVAL":     TOKEN_BYVAL,
	"LOCAL":     TOKEN_LOCAL,
	"SHARED":    TOKEN_SHARED,
	"STATIC":    TOKEN_STATIC,
	"COMMON":    TOKEN_COMMON,
	"OPTION":    TOKEN_OPTION,
	"BASE":      TOKEN_BASE,
	"INTEGER":   TOKEN_INTEGER_KW,
	"LONG":      TOKEN_LONG_KW,
	"SINGLE":    TOKEN_SINGLE_KW,
	"DOUBLE":    TOKEN_DOUBLE_KW,
	"STRING":    TOKEN_STRING_KW,
	"DEFINT":    TOKEN_DEFINT,
	"DEFLNG":    TOKEN_DEFLNG,
	"DEFSNG":    TOKEN_DEFSNG,
	"DEFDBL":    TOKEN_DEFDBL,
	"DEFSTR":    TOKEN_DEFSTR,
	"ERASE":     TOKEN_ERASE,
	"SWAP":      TOKEN_SWAP,
	"INCR":      TOKEN_INCR,
	"DECR":      TOKEN_DECR,
	"LBOUND":    TOKEN_LBOUND,
	"UBOUND":    TOKEN_UBOUND,
	"DATA":      TOKEN_DATA,
	"READ":      TOKEN_READ,
	"RESTORE":   TOKEN_RESTORE,
	"REM":       TOKEN_REM,
	"ERROR":     TOKEN_ERROR,
	"RESUME":    TOKEN_RESUME,
	"OPEN":      TOKEN_OPEN,
	"CLOSE":     TOKEN_CLOSE,
	"OUTPUT":    TOKEN_OUTPUT_KW,
	"APPEND":    TOKEN_APPEND,
	"RANDOM":    TOKEN_RANDOM_KW,
	"BINARY":    TOKEN_BINARY_KW,
	"FIELD":     TOKEN_FIELD,
	"GET":       TOKEN_GET,
	"PUT":       TOKEN_PUT,
	"SEEK":      TOKEN_SEEK,
	"LSET":      TOKEN_LSET,
	"RSET":      TOKEN_RSET,
	"KILL":      TOKEN_KILL,
	"NAME":      TOKEN_NAME,
	"FILES":     TOKEN_FILES,
	"CHDIR":     TOKEN_CHDIR,
	"MKDIR":     TOKEN_MKDIR,
	"RMDIR":     TOKEN_RMDIR,
	"BLOAD":     TOKEN_BLOAD,
	"BSAVE":     TOKEN_BSAVE,
	"IOCTL":     TOKEN_IOCTL,
	"LEN":       TOKEN_LEN_KW,
	"EOF":       TOKEN_EOF_KW,
	"SCREEN":    TOKEN_SCREEN,
	"PSET":      TOKEN_PSET,
	"PRESET":    TOKEN_PRESET,
	"CIRCLE":    TOKEN_CIRCLE,
	"PAINT":     TOKEN_PAINT,
	"DRAW":      TOKEN_DRAW,
	"VIEW":      TOKEN_VIEW,
	"WINDOW":    TOKEN_WINDOW,
	"PALETTE":   TOKEN_PALETTE,
	"PEEK":      TOKEN_PEEK,
	"POKE":      TOKEN_POKE,
	"INP":       TOKEN_INP,
	"OUT":       TOKEN_OUT,
	"WAIT":      TOKEN_WAIT,
	"VARPTR":    TOKEN_VARPTR,
	"VARSEG":    TOKEN_VARSEG,
	"TRON":      TOKEN_TRON,
	"TROFF":     TOKEN_TROFF,
	"TIMER":     TOKEN_TIMER,
	"DELAY":     TOKEN_DELAY,
	"MTIMER":    TOKEN_MTIMER,
	"TAB":       TOKEN_TAB,
	"SPC":       TOKEN_SPC,
	"RANDOMIZE": TOKEN_RANDOMIZE,
	"INSTAT":    TOKEN_INSTAT,
}

// LookupIdent returns the token type for an identifier.
// If the identifier is a keyword, it returns the keyword token type.
// Otherwise, it returns TOKEN_IDENTIFIER.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENTIFIER
}

// tokenNames maps token types to human-readable names for debugging.
var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:          "ILLEGAL",
	TOKEN_EOF:              "EOF",
	TOKEN_EOL:              "EOL",
	TOKEN_COMMENT:          "COMMENT",
	TOKEN_IDENTIFIER:       "IDENTIFIER",
	TOKEN_INTEGER:          "INTEGER",
	TOKEN_LONG:             "LONG",
	TOKEN_SINGLE:           "SINGLE",
	TOKEN_DOUBLE:           "DOUBLE",
	TOKEN_STRING:           "STRING",
	TOKEN_HEX:              "HEX",
	TOKEN_OCTAL:            "OCTAL",
	TOKEN_BINARY_LIT:       "BINARY",
	TOKEN_LINENUMBER:       "LINENUMBER",
	TOKEN_LABEL:            "LABEL",
	TOKEN_PLUS:             "PLUS",
	TOKEN_MINUS:            "MINUS",
	TOKEN_STAR:             "STAR",
	TOKEN_SLASH:            "SLASH",
	TOKEN_BACKSLASH:        "BACKSLASH",
	TOKEN_CARET:            "CARET",
	TOKEN_EQ:               "EQ",
	TOKEN_NE:               "NE",
	TOKEN_LT:               "LT",
	TOKEN_GT:               "GT",
	TOKEN_LE:               "LE",
	TOKEN_GE:               "GE",
	TOKEN_LPAREN:           "LPAREN",
	TOKEN_RPAREN:           "RPAREN",
	TOKEN_COMMA:            "COMMA",
	TOKEN_SEMICOLON:        "SEMICOLON",
	TOKEN_COLON:            "COLON",
	TOKEN_HASH:             "HASH",
	TOKEN_DOLLAR:           "DOLLAR",
	TOKEN_PERCENT:          "PERCENT",
	TOKEN_AMPERSAND:        "AMPERSAND",
	TOKEN_EXCLAIM:          "EXCLAIM",
	TOKEN_SUFFIX_PERCENT:   "SUFFIX_%",
	TOKEN_SUFFIX_AMPERSAND: "SUFFIX_&",
	TOKEN_SUFFIX_EXCLAIM:   "SUFFIX_!",
	TOKEN_SUFFIX_HASH:      "SUFFIX_#",
	TOKEN_SUFFIX_DOLLAR:    "SUFFIX_$",
	TOKEN_PRINT:            "PRINT",
	TOKEN_LPRINT:           "LPRINT",
	TOKEN_INPUT:            "INPUT",
	TOKEN_LINE:             "LINE",
	TOKEN_WRITE:            "WRITE",
	TOKEN_USING:            "USING",
	TOKEN_LOCATE:           "LOCATE",
	TOKEN_CLS:              "CLS",
	TOKEN_COLOR:            "COLOR",
	TOKEN_WIDTH:            "WIDTH",
	TOKEN_KEY:              "KEY",
	TOKEN_BEEP:             "BEEP",
	TOKEN_SOUND:            "SOUND",
	TOKEN_PLAY:             "PLAY",
	TOKEN_LET:              "LET",
	TOKEN_IF:               "IF",
	TOKEN_THEN:             "THEN",
	TOKEN_ELSE:             "ELSE",
	TOKEN_ELSEIF:           "ELSEIF",
	TOKEN_END:              "END",
	TOKEN_FOR:              "FOR",
	TOKEN_TO:               "TO",
	TOKEN_STEP:             "STEP",
	TOKEN_NEXT:             "NEXT",
	TOKEN_WHILE:            "WHILE",
	TOKEN_WEND:             "WEND",
	TOKEN_DO:               "DO",
	TOKEN_LOOP:             "LOOP",
	TOKEN_UNTIL:            "UNTIL",
	TOKEN_SELECT:           "SELECT",
	TOKEN_CASE:             "CASE",
	TOKEN_IS:               "IS",
	TOKEN_GOTO:             "GOTO",
	TOKEN_GOSUB:            "GOSUB",
	TOKEN_RETURN:           "RETURN",
	TOKEN_EXIT:             "EXIT",
	TOKEN_STOP:             "STOP",
	TOKEN_SYSTEM:           "SYSTEM",
	TOKEN_ON:               "ON",
	TOKEN_AND:              "AND",
	TOKEN_OR:               "OR",
	TOKEN_NOT:              "NOT",
	TOKEN_XOR:              "XOR",
	TOKEN_EQV:              "EQV",
	TOKEN_IMP:              "IMP",
	TOKEN_MOD:              "MOD",
	TOKEN_DIM:              "DIM",
	TOKEN_REDIM:            "REDIM",
	TOKEN_AS:               "AS",
	TOKEN_SUB:              "SUB",
	TOKEN_FUNCTION:         "FUNCTION",
	TOKEN_DECLARE:          "DECLARE",
	TOKEN_CALL:             "CALL",
	TOKEN_DEF:              "DEF",
	TOKEN_FN:               "FN",
	TOKEN_BYVAL:            "BYVAL",
	TOKEN_LOCAL:            "LOCAL",
	TOKEN_SHARED:           "SHARED",
	TOKEN_STATIC:           "STATIC",
	TOKEN_COMMON:           "COMMON",
	TOKEN_OPTION:           "OPTION",
	TOKEN_BASE:             "BASE",
	TOKEN_INTEGER_KW:       "INTEGER_KW",
	TOKEN_LONG_KW:          "LONG_KW",
	TOKEN_SINGLE_KW:        "SINGLE_KW",
	TOKEN_DOUBLE_KW:        "DOUBLE_KW",
	TOKEN_STRING_KW:        "STRING_KW",
	TOKEN_DEFINT:           "DEFINT",
	TOKEN_DEFLNG:           "DEFLNG",
	TOKEN_DEFSNG:           "DEFSNG",
	TOKEN_DEFDBL:           "DEFDBL",
	TOKEN_DEFSTR:           "DEFSTR",
	TOKEN_ERASE:            "ERASE",
	TOKEN_SWAP:             "SWAP",
	TOKEN_INCR:             "INCR",
	TOKEN_DECR:             "DECR",
	TOKEN_LBOUND:           "LBOUND",
	TOKEN_UBOUND:           "UBOUND",
	TOKEN_DATA:             "DATA",
	TOKEN_READ:             "READ",
	TOKEN_RESTORE:          "RESTORE",
	TOKEN_REM:              "REM",
	TOKEN_ERROR:            "ERROR",
	TOKEN_RESUME:           "RESUME",
	TOKEN_OPEN:             "OPEN",
	TOKEN_CLOSE:            "CLOSE",
	TOKEN_OUTPUT_KW:        "OUTPUT",
	TOKEN_APPEND:           "APPEND",
	TOKEN_RANDOM_KW:        "RANDOM",
	TOKEN_BINARY_KW:        "BINARY_KW",
	TOKEN_FIELD:            "FIELD",
	TOKEN_GET:              "GET",
	TOKEN_PUT:              "PUT",
	TOKEN_SEEK:             "SEEK",
	TOKEN_LSET:             "LSET",
	TOKEN_RSET:             "RSET",
	TOKEN_KILL:             "KILL",
	TOKEN_NAME:             "NAME",
	TOKEN_FILES:            "FILES",
	TOKEN_CHDIR:            "CHDIR",
	TOKEN_MKDIR:            "MKDIR",
	TOKEN_RMDIR:            "RMDIR",
	TOKEN_BLOAD:            "BLOAD",
	TOKEN_BSAVE:            "BSAVE",
	TOKEN_IOCTL:            "IOCTL",
	TOKEN_LEN_KW:           "LEN",
	TOKEN_EOF_KW:           "EOF_KW",
	TOKEN_SCREEN:           "SCREEN",
	TOKEN_PSET:             "PSET",
	TOKEN_PRESET:           "PRESET",
	TOKEN_CIRCLE:           "CIRCLE",
	TOKEN_PAINT:            "PAINT",
	TOKEN_DRAW:             "DRAW",
	TOKEN_VIEW:             "VIEW",
	TOKEN_WINDOW:           "WINDOW",
	TOKEN_PALETTE:          "PALETTE",
	TOKEN_PEEK:             "PEEK",
	TOKEN_POKE:             "POKE",
	TOKEN_INP:              "INP",
	TOKEN_OUT:              "OUT",
	TOKEN_WAIT:             "WAIT",
	TOKEN_VARPTR:           "VARPTR",
	TOKEN_VARSEG:           "VARSEG",
	TOKEN_TRON:             "TRON",
	TOKEN_TROFF:            "TROFF",
	TOKEN_TIMER:            "TIMER",
	TOKEN_DELAY:            "DELAY",
	TOKEN_MTIMER:           "MTIMER",
	TOKEN_TAB:              "TAB",
	TOKEN_SPC:              "SPC",
	TOKEN_RANDOMIZE:        "RANDOMIZE",
	TOKEN_INSTAT:           "INSTAT",
	TOKEN_META_DYNAMIC:     "$DYNAMIC",
	TOKEN_META_STATIC:      "$STATIC",
	TOKEN_META_INCLUDE:     "$INCLUDE",
	TOKEN_META_IF:          "$IF",
	TOKEN_META_ELSEIF:      "$ELSEIF",
	TOKEN_META_ELSE:        "$ELSE",
	TOKEN_META_ENDIF:       "$ENDIF",
	TOKEN_META_COM:         "$COM",
	TOKEN_META_SOUND:       "$SOUND",
	TOKEN_META_STACK:       "$STACK",
	TOKEN_META_SEGMENT:     "$SEGMENT",
	TOKEN_META_INLINE:      "$INLINE",
	TOKEN_META_EVENT:       "$EVENT",
	TOKEN_RUN:              "RUN",
	TOKEN_CHAIN:            "CHAIN",
	TOKEN_SHELL:            "SHELL",
}

// TokenName returns the human-readable name of a token type.
func TokenName(t TokenType) string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

// IsKeywordOperator returns true if the token is a keyword-style operator.
func IsKeywordOperator(t TokenType) bool {
	switch t {
	case TOKEN_AND, TOKEN_OR, TOKEN_NOT, TOKEN_XOR, TOKEN_EQV, TOKEN_IMP, TOKEN_MOD:
		return true
	}
	return false
}
