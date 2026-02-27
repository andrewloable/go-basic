// Package lexer implements the lexical analysis (scanning) phase of the BASIC
// compiler pipeline.
//
// # Compiler pipeline overview
//
// A compiler typically has these front-end phases:
//
//	Source text  →  [Lexer]  →  token stream  →  [Parser]  →  AST  →  ...
//
// The lexer (also called a scanner or tokeniser) is the very first phase. Its
// sole job is to convert a flat sequence of characters into a flat sequence of
// tokens so that the parser can work at a higher level of abstraction without
// caring about whitespace, case folding, or individual characters.
//
// # What is a token?
//
// A token is the smallest meaningful unit of a language — the "atom of
// meaning". Just as words are the atoms of a sentence, tokens are the atoms of
// a program. Every token has:
//
//   - A type  — an integer tag that classifies the token (e.g. TOKEN_PLUS,
//     TOKEN_INTEGER, TOKEN_PRINT). Because it is just an int, the parser can
//     compare token types in O(1) with a simple == or switch, rather than
//     doing a string comparison on every character.
//   - A literal — the exact source text that was consumed to produce the token
//     (e.g. "3.14", "PRINT", "+"). This is kept so that error messages can
//     quote the original source text and so that literal values (numbers,
//     strings) can be parsed later.
//   - A position — (line, column) so that diagnostics point at the right place.
//
// # How the lexer works (character stream → token stream)
//
// The lexer keeps a read cursor into the source string and repeatedly applies
// the following algorithm:
//
//  1. Skip whitespace (spaces, tabs — but NOT newlines, which are significant
//     in BASIC because statements end at end-of-line).
//  2. Peek at the current character to decide which kind of token to produce.
//  3. Consume as many characters as belong to that token (e.g. all digits of a
//     number, all letters of an identifier or keyword).
//  4. Wrap the consumed text in a Token struct and emit it.
//  5. Repeat until the end of the source is reached, then emit TOKEN_EOF.
//
// This is a hand-written, single-pass, O(n) scanner — the simplest and most
// common design for production compilers.
//
// # BASIC type suffixes (%  &  !  #  $)
//
// BASIC uses one-character type sigils that are written directly after an
// identifier or a numeric literal to declare its type without a separate "AS"
// keyword:
//
//	A%   → integer variable   (16-bit signed)
//	A&   → long variable      (32-bit signed)
//	A!   → single-precision float
//	A#   → double-precision float
//	A$   → string variable
//
// The same characters can also appear as standalone punctuation (e.g. # is
// used for file-handle numbers, $ appears in built-in function names). The
// lexer resolves the ambiguity contextually:
//
//   - When one of these characters follows an identifier or numeric literal
//     without any intervening whitespace, it is scanned as a TOKEN_SUFFIX_*
//     token (attached meaning).
//   - When it appears on its own, it is scanned as the matching punctuation
//     token (TOKEN_HASH, TOKEN_DOLLAR, etc.).
//
// Keeping suffix tokens distinct from punctuation tokens lets the parser apply
// different grammar rules for each case without look-back.
//
// # TokenType as an integer enum
//
// TokenType is declared as "type TokenType int" and all constants are produced
// with iota, which assigns consecutive integers starting at 0. This gives:
//
//   - Compact representation — one int per token, no heap allocation.
//   - O(1) comparison — the parser uses token.Type == TOKEN_PRINT, not string
//     equality, so hot-path checks are a single integer compare.
//   - Switch-friendly dispatch — Go's switch on an integer is compiled to a
//     jump table by the compiler for dense ranges, giving O(1) dispatch.
//   - Debuggability — the tokenNames map translates the integer back to a
//     human-readable string for error messages and pretty-printers.
//
// # Keyword recognition (LookupIdent)
//
// When the lexer finishes scanning a sequence of letters and digits, it does
// not yet know whether it has read a keyword (PRINT, FOR, IF …) or a
// user-defined name (myVar, counter …). LookupIdent resolves this:
//
//  1. The raw text is upper-cased (BASIC is case-insensitive).
//  2. The upper-cased text is looked up in the "keywords" map
//     (string → TokenType).
//  3. If found, the keyword's token type is returned; otherwise TOKEN_IDENTIFIER
//     is returned.
//
// Using a hash map makes keyword recognition O(1) on average regardless of
// how many keywords the language has. The alternative — a long chain of string
// comparisons or a trie — is slower or more complex to maintain.
//
// Metacommands ($DYNAMIC, $INCLUDE, …) are NOT in this map; they begin with
// '$' so the lexer recognises them by their leading character and handles them
// as a separate token family.
package lexer

import "fmt"

// TokenType represents the type of a lexical token.
// It is an integer rather than a string so that the parser can compare token
// types with a single integer operation instead of a string comparison.
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

	// ---- Literal tokens: punctuation and operators ----
	// Each single- or double-character symbol that has syntactic meaning gets
	// its own token type. Giving every symbol a distinct integer tag means the
	// parser never needs to inspect the Literal field when making grammar
	// decisions — it only checks the Type, which is a single integer compare.
	// Multi-character operators (<=, >=, <>) are recognised by the lexer during
	// a single peek-ahead step and emitted as a single token.

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
	// Keywords are reserved words that have a fixed meaning in the language
	// grammar. They are detected by LookupIdent (see below): after the lexer
	// reads a run of identifier characters it upper-cases the text and probes
	// the keywords map. If there is a hit, the keyword's token type is used
	// instead of TOKEN_IDENTIFIER. This strategy keeps the scanner simple —
	// it always reads identifiers the same way — while still giving the parser
	// a distinct token type per keyword so grammar rules can be written as
	// plain integer comparisons.
	//
	// BASIC is case-insensitive ("print" and "PRINT" are the same keyword), so
	// the map stores only upper-case keys; the lexer normalises before lookup.

	// I/O keywords
	TOKEN_PRINT
	TOKEN_LPRINT
	TOKEN_INPUT
	TOKEN_LINE // LINE (as in LINE INPUT, LINE graphics)
	TOKEN_WRITE
	TOKEN_USING
	TOKEN_LOCATE
	TOKEN_CLS
	TOKEN_CLEAR
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
	TOKEN_TYPE
	TOKEN_CONST
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

	// ---- Metacompiler tokens ($-directives) ----
	// Metacommands are compiler directives embedded inside comment lines
	// (lines beginning with REM or '). They control the compiler itself rather
	// than the runtime behaviour of the program. In QB/GW-BASIC they appear as
	// "'$INCLUDE:'file.bi'" or "'$DYNAMIC".
	//
	// Because metacommands begin with '$', they cannot be ordinary identifiers
	// (BASIC identifiers start with a letter). The lexer detects them by
	// peeking at the first character of a comment body: if it is '$' the rest
	// of the word is scanned and mapped to one of the TOKEN_META_* types below.
	// This gives the parser a way to act on compiler directives at parse time
	// (e.g. to include another file or switch memory-model modes) without
	// treating them as user code.

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
	"CLEAR":     TOKEN_CLEAR,
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
	"TYPE":      TOKEN_TYPE,
	"CONST":     TOKEN_CONST,
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
	TOKEN_CLEAR:            "CLEAR",
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
	TOKEN_TYPE:             "TYPE",
	TOKEN_CONST:            "CONST",
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
