# Turbo BASIC Language Reference

Complete reference for Turbo BASIC / QBasic language features relevant to the go-basic transpiler.

## Type System

### Variable Types

| Suffix | Type | Go Equivalent | Range |
|--------|------|---------------|-------|
| `%` | Integer | `int16` | -32,768 to 32,767 |
| `&` | Long | `int32` | -2,147,483,648 to 2,147,483,647 |
| `!` | Single | `float32` | ~7 digits precision |
| `#` | Double | `float64` | ~15 digits precision |
| `$` | String | `string` | 0-32,767 characters |
| (none) | Single (default) | `float32` | Default numeric type |

### Type Declaration Alternatives

```basic
DIM X AS INTEGER      ' Same as X%
DIM Y AS LONG         ' Same as Y&
DIM Z AS SINGLE       ' Same as Z!
DIM D AS DOUBLE       ' Same as D#
DIM S AS STRING       ' Same as S$
```

### DEFtype Statements

Set default type for variables starting with specific letters:
```basic
DEFINT A-Z      ' All vars default to INTEGER
DEFSNG X-Z      ' X, Y, Z default to SINGLE
DEFSTR S        ' S* vars default to STRING
```

### Implicit Type Conversion

BASIC automatically converts between numeric types. The transpiler must emit explicit Go casts:
- Widening (int16 → float64): always safe
- Narrowing (float64 → int16): may lose precision, but BASIC allows it
- String ↔ Numeric: use VAL() and STR$()

## Control Flow Statements

### IF/THEN/ELSE

```basic
' Single-line:
IF x > 0 THEN PRINT "positive" ELSE PRINT "non-positive"

' Multi-line block:
IF x > 0 THEN
    PRINT "positive"
ELSEIF x = 0 THEN
    PRINT "zero"
ELSE
    PRINT "negative"
END IF
```

### FOR/NEXT

```basic
FOR i = 1 TO 10 STEP 2
    PRINT i
NEXT i
```

Step can be negative. Loop terminates when counter passes the end value in the step direction.

### WHILE/WEND

```basic
WHILE x < 100
    x = x * 2
WEND
```

### DO/LOOP (Turbo BASIC extension)

```basic
DO WHILE condition    DO UNTIL condition    DO
    ...                   ...                   ...
LOOP                  LOOP                  LOOP WHILE condition

DO                    DO
    ...                   ...
LOOP UNTIL condition  LOOP          ' infinite loop, use EXIT DO
```

### SELECT CASE

```basic
SELECT CASE score
    CASE 90 TO 100
        grade$ = "A"
    CASE 80 TO 89
        grade$ = "B"
    CASE IS >= 70
        grade$ = "C"
    CASE ELSE
        grade$ = "F"
END SELECT
```

### GOTO / GOSUB / RETURN

```basic
GOTO label
GOSUB subroutine
RETURN              ' return from GOSUB

ON n GOTO label1, label2, label3      ' computed GOTO
ON n GOSUB sub1, sub2, sub3           ' computed GOSUB
```

### EXIT

```basic
EXIT FOR        ' break out of FOR loop
EXIT DO         ' break out of DO loop
EXIT WHILE      ' break out of WHILE loop (Turbo BASIC)
EXIT LOOP       ' same as EXIT DO (Turbo BASIC)
EXIT SUB        ' return from SUB
EXIT FUNCTION   ' return from FUNCTION
EXIT DEF        ' return from DEF FN
```

## Procedures

### SUB (Subroutine)

```basic
SUB CalculateArea (length!, width!)
    SHARED result!           ' access global variable
    result! = length! * width!
END SUB

' Calling:
CALL CalculateArea(10, 20)
CalculateArea 10, 20         ' CALL keyword optional
```

Parameters are passed by reference by default. Use `BYVAL` for value semantics.
`SHARED` makes global variables accessible inside the SUB.

### FUNCTION

```basic
FUNCTION Square# (x#)
    Square# = x# * x#       ' assign to function name = set return value
END FUNCTION

' Calling:
result# = Square#(5)
```

### DEF FN (Inline Function)

```basic
DEF FNCtoF(c) = c * 9 / 5 + 32     ' single-expression
temp = FNCtoF(100)

' Multi-line:
DEF FNFactorial#(n)
    IF n <= 1 THEN
        FNFactorial# = 1
    ELSE
        FNFactorial# = n * FNFactorial#(n - 1)
    END IF
END DEF
```

Assigning to the function name inside the body sets the return value.

### DECLARE

```basic
DECLARE SUB MySub (x AS INTEGER)
DECLARE FUNCTION MyFunc# (x#)
```

Forward declaration. The parser records these but they don't generate code.

## Arrays

### DIM

```basic
DIM A(100)              ' 1-D array, indices 0-100 (101 elements)
DIM Matrix(10, 20)      ' 2-D array
DIM Names$(50)          ' String array
DIM SHARED Flags%(100)  ' Shared (global) array
```

BASIC arrays are 0-based by default (can be changed with OPTION BASE 1).

### REDIM

```basic
REDIM A(200)            ' Resize dynamic array (clears contents)
REDIM PRESERVE A(200)   ' Resize but keep existing data
```

### ERASE

```basic
ERASE A, B$             ' Deallocate arrays
```

## DATA / READ / RESTORE

```basic
DATA "Smith", 42, 35000.50
DATA "Jones", 35, 42000

READ Name$, Age%, Salary!
READ Name$, Age%, Salary!

RESTORE                 ' Reset read pointer to first DATA
```

DATA values are collected into a single pool at compile time. READ consumes them sequentially.

## String Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `LEFT$(s$, n)` | `string, int → string` | First n characters |
| `RIGHT$(s$, n)` | `string, int → string` | Last n characters |
| `MID$(s$, start, len)` | `string, int, int → string` | Substring from start |
| `LEN(s$)` | `string → int` | String length |
| `INSTR([start,] s$, find$)` | `[int,] string, string → int` | Find substring |
| `CHR$(n)` | `int → string` | Character from ASCII code |
| `ASC(s$)` | `string → int` | ASCII code of first char |
| `STR$(n)` | `number → string` | Number to string |
| `VAL(s$)` | `string → number` | String to number |
| `HEX$(n)` | `int → string` | Hex representation |
| `OCT$(n)` | `int → string` | Octal representation |
| `BIN$(n)` | `int → string` | Binary representation (Turbo BASIC) |
| `UCASE$(s$)` | `string → string` | Uppercase |
| `LCASE$(s$)` | `string → string` | Lowercase |
| `LTRIM$(s$)` | `string → string` | Remove leading spaces |
| `RTRIM$(s$)` | `string → string` | Remove trailing spaces |
| `SPACE$(n)` | `int → string` | String of n spaces |
| `STRING$(n, char)` | `int, int/string → string` | Repeat character n times |

## Math Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `ABS(x)` | `number → number` | Absolute value |
| `SGN(x)` | `number → int` | Sign (-1, 0, 1) |
| `INT(x)` | `number → number` | Floor (round toward -infinity) |
| `FIX(x)` | `number → number` | Truncate (round toward zero) |
| `SQR(x)` | `number → number` | Square root |
| `SIN(x)` | `number → number` | Sine (radians) |
| `COS(x)` | `number → number` | Cosine |
| `TAN(x)` | `number → number` | Tangent |
| `ATN(x)` | `number → number` | Arctangent |
| `EXP(x)` | `number → number` | e^x |
| `LOG(x)` | `number → number` | Natural logarithm |
| `CINT(x)` | `number → int16` | Convert to integer |
| `CLNG(x)` | `number → int32` | Convert to long |
| `CSNG(x)` | `number → float32` | Convert to single |
| `CDBL(x)` | `number → float64` | Convert to double |
| `RND[(x)]` | `[number] → number` | Random 0-1. Can use without parens. |
| `RANDOMIZE seed` | Statement | Seed the RNG |

## I/O Functions (no-argument forms)

These builtins are used WITHOUT parentheses — they look like variables:

| Name | Returns | Description |
|------|---------|-------------|
| `INKEY$` | `string` | Non-blocking keyboard read, "" if no key |
| `DATE$` | `string` | Current date "MM-DD-YYYY" |
| `TIME$` | `string` | Current time "HH:MM:SS" |
| `TIMER` | `number` | Seconds since midnight (float) |

## System Functions

| Function | Description |
|----------|-------------|
| `FRE(x)` | Free memory. `FRE("")` string space, `FRE(-1)` array, `FRE(-2)` stack |
| `PEEK(addr)` | Read byte from memory address (DOS-specific) |
| `POKE addr, val` | Write byte to memory address (statement, not function) |
| `ERR` | Last error code (no-arg, looks like variable) |
| `ERL` | Last error line number |
| `ERADR` | Error address (Turbo BASIC extension) |
| `ENVIRON$(name$)` | Get environment variable |
| `ENVIRON "NAME=VALUE"` | Set environment variable (statement form) |
| `COMMAND$` | Command-line arguments |

## Error Handling

```basic
ON ERROR GOTO handler    ' Enable error trapping
...
ON ERROR GOTO 0          ' Disable error trapping

handler:
    PRINT "Error"; ERR; "at line"; ERL
    RESUME NEXT          ' Continue after the failing statement
    ' or: RESUME         ' Retry the failing statement
    ' or: RESUME label   ' Jump to label
```

## File I/O

### Sequential Files

```basic
OPEN "data.txt" FOR OUTPUT AS #1    ' Write mode
PRINT #1, "Hello"                    ' Write line
WRITE #1, Name$, Age%, Salary!       ' Write comma-delimited
CLOSE #1

OPEN "data.txt" FOR INPUT AS #1     ' Read mode
INPUT #1, line$                      ' Read line
LINE INPUT #1, fullLine$             ' Read entire line including commas
CLOSE #1

OPEN "data.txt" FOR APPEND AS #1    ' Append mode
```

### Random Access Files

```basic
OPEN "data.dat" FOR RANDOM AS #1 LEN = 128
FIELD #1, 20 AS Name$, 4 AS Age$, 8 AS Salary$
LSET Name$ = "Smith"
LSET Age$ = MKI$(42)
PUT #1, 1                           ' Write record 1
GET #1, 1                           ' Read record 1
CLOSE #1
```

### Binary Files

```basic
OPEN "data.bin" FOR BINARY AS #1
SEEK #1, 100                        ' Move to byte 100
PUT #1, , value%                     ' Write at current position
GET #1, , value%                     ' Read at current position
CLOSE #1
```

### File Functions

| Function | Description |
|----------|-------------|
| `EOF(n)` | End of file? Returns -1 (true) or 0 |
| `LOC(n)` | Current position in file |
| `LOF(n)` | Length of file |
| `SEEK(n)` | Current byte position (function form) |

## Graphics (Screen Mode Dependent)

### SCREEN

```basic
SCREEN 0     ' Text mode (80x25)
SCREEN 7     ' 320x200, 16 colors (EGA)
SCREEN 9     ' 640x350, 16 colors (EGA)
SCREEN 12    ' 640x480, 16 colors (VGA)
SCREEN 13    ' 320x200, 256 colors (VGA)
```

### Drawing Commands

```basic
PSET (x, y) [, color]                            ' Set pixel
LINE (x1,y1)-(x2,y2) [, color [, B[F]]]          ' Line or box (B=box, BF=filled)
CIRCLE (x,y), radius [, color [, start, end [, aspect]]]  ' Circle/ellipse/arc
PAINT (x,y) [, fillColor [, borderColor]]         ' Flood fill
DRAW "command string"                              ' Turtle graphics
VIEW [[SCREEN] (x1,y1)-(x2,y2) [, color [, border]]]  ' Set viewport
WINDOW [[SCREEN] (x1,y1)-(x2,y2)]                ' Set coordinate system
```

### Graphics GET/PUT

```basic
GET (x1,y1)-(x2,y2), arrayName    ' Capture screen region to array
PUT (x,y), arrayName [, mode]     ' Display array on screen
' Modes: PSET, PRESET, AND, OR, XOR
```

## Sound

```basic
SOUND frequency, duration    ' Tone in Hz, duration in clock ticks (18.2/sec)
PLAY "command string"        ' MML music macro language
' PLAY commands: T=tempo, O=octave, L=length, CDEFGAB=notes, P=pause
' MB=music background, MF=music foreground (Turbo BASIC)
```

## TYPE (User-Defined Types / Structs)

```basic
TYPE XYPoint
    XCoor AS INTEGER
    YCoor AS INTEGER
END TYPE

DIM p AS XYPoint
p.XCoor = 100
p.YCoor = 200

DIM pts(10) AS XYPoint
pts(1).XCoor = 50
```

## Metacompiler Directives

```basic
'$DYNAMIC              ' Arrays are dynamic (can REDIM)
'$STATIC               ' Arrays are static (fixed size)
'$INCLUDE: 'filename'  ' Include source file

'$IF flag = 1
    ' conditional compilation
'$ELSE
    ' alternative
'$ENDIF
```

Note: metacommands start with `'$` (apostrophe-dollar) — they look like comments to other BASIC compilers.
