# Compiler Architecture Reference

Detailed phase-by-phase guide to the go-basic compiler internals.

## Phase 1: Lexer (`internal/lexer/`)

### Token Types

The lexer produces tokens classified into these groups:

**Literals**: `TOKEN_INTEGER`, `TOKEN_FLOAT`, `TOKEN_STRING`, `TOKEN_IDENTIFIER`

**Operators**: `TOKEN_PLUS`, `TOKEN_MINUS`, `TOKEN_ASTERISK`, `TOKEN_SLASH`, `TOKEN_BACKSLASH` (integer div), `TOKEN_CARET` (exponent), `TOKEN_MOD`

**Comparison**: `TOKEN_EQ`, `TOKEN_NE` (`<>`), `TOKEN_LT`, `TOKEN_GT`, `TOKEN_LE`, `TOKEN_GE`

**Logical**: `TOKEN_AND`, `TOKEN_OR`, `TOKEN_NOT`, `TOKEN_XOR`, `TOKEN_EQV`, `TOKEN_IMP`

**Punctuation**: `TOKEN_LPAREN`, `TOKEN_RPAREN`, `TOKEN_COMMA`, `TOKEN_SEMICOLON`, `TOKEN_COLON`, `TOKEN_HASH` (`#`)

**Type suffixes**: Consumed as part of the identifier token. The identifier `A$` is one TOKEN_IDENTIFIER with Literal="A$".

**Special**: `TOKEN_EOL` (end of line — significant in BASIC), `TOKEN_EOF`, `TOKEN_ILLEGAL`

**Keywords**: ~150 keyword tokens including `TOKEN_PRINT`, `TOKEN_FOR`, `TOKEN_IF`, `TOKEN_SUB`, `TOKEN_FUNCTION`, `TOKEN_DIM`, `TOKEN_GOTO`, `TOKEN_GOSUB`, etc. Each BASIC keyword has its own token type for fast dispatch in the parser.

**Metacommands**: `TOKEN_META_DYNAMIC` (`$DYNAMIC`), `TOKEN_META_STATIC` (`$STATIC`), `TOKEN_META_IF` (`$IF`), `TOKEN_META_ELSE` (`$ELSE`), `TOKEN_META_ENDIF` (`$ENDIF`), `TOKEN_META_INCLUDE` (`$INCLUDE`).

### Scanner Design

- Hand-written, single-pass, O(n)
- Case-insensitive keyword matching (all source uppercased for lookup)
- Newlines are significant tokens (TOKEN_EOL) — BASIC statements end at EOL
- The `.` character is emitted as TOKEN_ILLEGAL — used by the parser for struct field access
- Colon `:` separates multiple statements on one line
- `'` (apostrophe) starts a comment, equivalent to REM

### Key Functions

- `NextToken()` — main dispatch loop: skip whitespace, peek char, produce token
- `readNumber()` — handles integers, floats, hex (&H), octal (&O), scientific notation
- `readString()` — consumes "..." string literals
- `readIdentifier()` — reads identifier, checks keyword table, attaches type suffix

## Phase 2: Parser (`internal/parser/`)

### Architecture: Recursive Descent + Pratt (TDOP)

**Statement parsing** uses recursive descent: `parseStatement()` is a large switch on `curToken.Type` that dispatches to specialized parsers (`parseForStatement()`, `parseIfStatement()`, `parseSubStatement()`, etc.).

**Expression parsing** uses Pratt parsing (Top-Down Operator Precedence): `parseExpression(precedence)` calls a prefix handler for the current token, then loops calling infix handlers while the next operator's binding power exceeds the current precedence.

### Precedence Levels

```
PREC_LOWEST     = 0   // default
PREC_IMP        = 1   // IMP (logical implication)
PREC_EQV        = 2   // EQV (logical equivalence)
PREC_XOR        = 3   // XOR
PREC_OR         = 4   // OR
PREC_AND        = 5   // AND
PREC_NOT        = 6   // NOT (prefix)
PREC_COMPARISON = 7   // =, <>, <, >, <=, >=
PREC_CONCAT     = 8   // + (string concatenation, same as addition)
PREC_ADDITION   = 9   // +, -
PREC_MOD        = 10  // MOD
PREC_INTDIV     = 11  // \ (integer division)
PREC_MULTIPLY   = 12  // *, /
PREC_NEGATE     = 13  // - (unary prefix)
PREC_EXPONENT   = 14  // ^ (right-associative)
```

### Critical Function: parseIdentifierExpression()

This function resolves the ambiguity of `NAME(args)` — which could be:
1. **Builtin function call**: `CHR$(65)`, `LEFT$(s$, 5)`, `SIN(x)`
2. **User function call**: `FNHypot(a, b)` (DEF FN functions have "FN" prefix)
3. **Array access**: `A(3)`, `Matrix(i, j)`
4. **Simple variable**: `X`, `Name$`

Resolution order:
1. Strip type suffix from identifier name
2. If followed by `(`: check `isBuiltinFunction(name)` → FunctionCall node
3. If name starts with "FN": FunctionCall node (DEF FN call)
4. If followed by `(` but not builtin/FN: ArrayAccess node
5. If not followed by `(`: Identifier node
6. After creating the node, check for `.` (TOKEN_ILLEGAL) for struct field access

**Known bug**: Some $-suffixed builtins (CHR$, LEFT$, MID$) are incorrectly routed to the ArrayAccess branch because the suffix stripping and builtin check don't always align. This is tracked in task `v2j.1`.

### AST Node Catalog

#### Expression Nodes

| Node | Fields | Example |
|------|--------|---------|
| `NumberLiteral` | Value float64, NumType int | `42`, `3.14` |
| `StringLiteral` | Value string | `"Hello"` |
| `Identifier` | Name, TypeSuffix string | `X`, `Name$` |
| `InfixExpression` | Left, Right Expression, Operator string | `A + B` |
| `PrefixExpression` | Operator string, Right Expression | `-X`, `NOT flag` |
| `FunctionCall` | Name string, Args []Expression | `SIN(x)`, `LEFT$(s$, 5)` |
| `ArrayAccess` | Name, TypeSuffix string, Indices []Expression | `A(3)`, `M(i, j)` |
| `FieldAccessExpression` | Object Expression, Field string | `point.X` |

#### Statement Nodes (partial list of key ones)

| Node | Purpose |
|------|---------|
| `LetStatement` | Variable assignment: `LET X = expr` or `X = expr` |
| `PrintStatement` | PRINT with items (expressions, separators) |
| `ForStatement` | FOR/NEXT loop with Counter, Start, End, Step, Body |
| `IfStatement` | IF/THEN/ELSE with Condition, Consequence, Alternative |
| `WhileStatement` | WHILE/WEND loop |
| `DoLoopStatement` | DO/LOOP with optional WHILE/UNTIL |
| `SelectCaseStatement` | SELECT CASE with Cases list |
| `GotoStatement` | GOTO target (label or line number) |
| `GosubStatement` | GOSUB target |
| `ReturnStatement` | RETURN (from GOSUB) |
| `SubStatement` | SUB definition with Name, Params, Body |
| `FunctionStatement` | FUNCTION definition with return type |
| `DefFnStatement` | DEF FN inline function |
| `DimStatement` | DIM array declarations |
| `ReadStatement` | READ var1, var2 (from DATA pool) |
| `DataStatement` | DATA val1, val2, ... |
| `InputStatement` | INPUT "prompt"; var |
| `OpenStatement` | OPEN file FOR mode AS #n |
| `CloseStatement` | CLOSE #n |
| `OnComputedGotoStatement` | ON expr GOTO l1, l2, l3 |
| `OnComputedGosubStatement` | ON expr GOSUB l1, l2, l3 |
| `CallSubStatement` | CALL subname(args) |
| `RemStatement` | Comment (REM or ') |
| `FieldAssignStatement` | Struct field assignment: `p.X = 10` |

## Phase 3: Semantic Analysis (`internal/semantic/`)

Builds a symbol table with variable types inferred from:
1. Explicit type suffixes (A% → int16)
2. DIM declarations with AS clauses
3. Assignment context (type of RHS propagated to LHS)

The `SymbolTable` is passed to the codegen to resolve variable types.

## Phase 4a: Code Generation (`internal/codegen/`)

### Pre-pass: collectLabelsAndData()

Before emitting code, scans the entire AST to collect:
- **Labels**: line numbers and string labels that exist in the source
- **Referenced labels**: GOTO/GOSUB/ON GOTO/ON GOSUB targets (only these get emitted as Go labels)
- **DATA values**: all DATA statement values collected into a single pool

### Emission Order

1. Package declaration and imports
2. Import suppressor lines
3. `func main()` opening
4. Preamble: RNG init (if `needRng`), DATA pool, FileManager
5. Top-level statements in source order
6. `func main()` closing
7. SUB/FUNCTION definitions as separate Go functions

### Label Emission Strategy

Only labels referenced by GOTO/GOSUB are emitted in Go (Go compile error for unused labels).
Labels use the format `label_Name` or `label_123` for numeric line labels.

### The Default Catch-All

Any AST node without a dedicated case in `emitStatement()` falls through to:
```go
default:
    g.writeLinef("// TODO: %s", s.TokenLiteral())
```
This produces the `// TODO: CIRCLE`, `// TODO: SCREEN`, etc. stubs. Adding a new statement
to the codegen means adding its case ABOVE this default.

## Phase 4b: Bytecode VM (`internal/vm/`)

Secondary backend. Compiles AST to stack-based bytecode with opcodes like:
`OP_PUSH`, `OP_POP`, `OP_ADD`, `OP_SUB`, `OP_CALL`, `OP_JMP`, `OP_JZ`, etc.

The VM uses a value stack, a call stack, and a symbol table. Currently handles basic
arithmetic, control flow, and PRINT. Graphics and file I/O are not yet implemented in the VM.
