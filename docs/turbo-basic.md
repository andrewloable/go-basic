# Turbo BASIC Language Reference

A comprehensive reference for Turbo BASIC (Borland, 1987) — a native-code compiler for MS-DOS, approximately 90% compatible with GW-BASIC/BASICA, and the predecessor to PowerBASIC. This document covers all commands, functions, statements, and language features.

---

## Table of Contents

1. [Language Architecture](#1-language-architecture)
2. [Data Types](#2-data-types)
3. [Variable Declaration and Scope](#3-variable-declaration-and-scope)
4. [Operators](#4-operators)
5. [Control Flow](#5-control-flow)
6. [Subroutines and Functions](#6-subroutines-and-functions)
7. [Math Functions](#7-math-functions)
8. [Type Conversion Functions](#8-type-conversion-functions)
9. [String Functions](#9-string-functions)
10. [I/O and Screen Statements](#10-io-and-screen-statements)
11. [File I/O](#11-file-io)
12. [Graphics](#12-graphics)
13. [Sound](#13-sound)
14. [Data Statements](#14-data-statements)
15. [Memory and System Access](#15-memory-and-system-access)
16. [Error Handling](#16-error-handling)
17. [Event Trapping](#17-event-trapping)
18. [Timing](#18-timing)
19. [System Variables](#19-system-variables)
20. [Input Devices](#20-input-devices)
21. [Program Control and Chaining](#21-program-control-and-chaining)
22. [Metacommands (Compiler Directives)](#22-metacommands-compiler-directives)
23. [Turbo BASIC Extensions vs. GW-BASIC](#23-turbo-basic-extensions-vs-gw-basic)
24. [Common Error Codes](#24-common-error-codes)

---

## 1. Language Architecture

Turbo BASIC is a **compiler** (not an interpreter) that generates native 8086/8088 machine code for MS-DOS.

- No 64K program size limitation (unlike interpretive BASIC)
- Full 8087 math coprocessor integration with software emulation fallback
- 80-bit extended precision floating point internally; IEEE format for storage
- Supports CGA, EGA, Hercules, and VGA graphics adapters
- Programs may use **line numbers** (optional) or **alphanumeric labels**
- Line numbers are not required for structured code

```basic
' Traditional style (line numbers)
10 PRINT "Hello, World!"
20 END

' Modern style (labels, no line numbers)
Start:
    PRINT "Hello, World!"
END
```

---

## 2. Data Types

| Type    | Suffix | Size         | Range                                   |
|---------|--------|--------------|-----------------------------------------|
| INTEGER | `%`    | 16-bit signed | -32,768 to 32,767                      |
| LONG    | `&`    | 32-bit signed | -2,147,483,648 to 2,147,483,647        |
| SINGLE  | `!`    | 32-bit float  | ~7 significant digits                  |
| DOUBLE  | `#`    | 64-bit float  | ~15–16 significant digits              |
| STRING  | `$`    | variable      | up to 32,767 characters                |

Type suffixes are appended directly to variable names:

```basic
Count%   ' INTEGER
Total&   ' LONG
Pi#      ' DOUBLE
Name$    ' STRING
```

### Literal Constants

```basic
42             ' integer
42&            ' long integer
3.14           ' single float
3.14159265#    ' double float
2.5E-3         ' scientific notation
"Hello"        ' string
&H1A           ' hexadecimal
&O77           ' octal
&B1010         ' binary (Turbo BASIC extension)
```

### Default Type Declarations (by letter range)

```basic
DEFINT A-Z     ' all undeclared vars are INTEGER
DEFLNG A-Z     ' all undeclared vars are LONG
DEFSNG A-Z     ' all undeclared vars are SINGLE
DEFDBL A-Z     ' all undeclared vars are DOUBLE
DEFSTR A-Z     ' all undeclared vars are STRING
```

---

## 3. Variable Declaration and Scope

### DIM — Declare arrays or fixed-length strings

```basic
DIM array(10)                    ' indices 0 to 10 (11 elements)
DIM array(1 TO 10)               ' indices 1 to 10
DIM matrix(5, 5)                 ' 2D array
DIM matrix(1 TO 5, 1 TO 5)      ' 2D with explicit bounds
DIM name AS STRING * 20          ' fixed-length string of 20 chars
DIM x AS INTEGER                 ' explicit type declaration
```

### REDIM — Resize a dynamic array

```basic
REDIM array(newSize)
```

### OPTION BASE — Set default lower bound for arrays

```basic
OPTION BASE 0    ' default; arrays start at index 0
OPTION BASE 1    ' arrays start at index 1
```

### ERASE — Reset or free arrays

```basic
ERASE array1, array2
```

### LBOUND / UBOUND — Query array bounds

```basic
low  = LBOUND(array)         ' lower bound of first dimension
high = UBOUND(array)         ' upper bound of first dimension
low2 = LBOUND(matrix, 2)    ' lower bound of second dimension
```

### Scope Modifiers

| Modifier   | Description |
|------------|-------------|
| `COMMON`   | Share variable between CHAINed programs |
| `SHARED`   | Share between SUB/FUNCTION and the main module |
| `LOCAL`    | Local to current SUB/FUNCTION (default inside procedures) |
| `STATIC`   | Retains value between procedure calls |

```basic
SUB MyProc()
    LOCAL tempVar AS INTEGER        ' local; lost after return
    STATIC callCount AS INTEGER     ' persists across calls
    SHARED globalData               ' visible in main program
END SUB
```

---

## 4. Operators

### Arithmetic

| Operator | Operation |
|----------|-----------|
| `+`      | Addition |
| `-`      | Subtraction / Unary negation |
| `*`      | Multiplication |
| `/`      | Floating-point division |
| `\`      | Integer division (truncates toward zero) |
| `MOD`    | Modulo (remainder) |
| `^`      | Exponentiation |

### Relational

| Operator | Meaning |
|----------|---------|
| `=`      | Equal to |
| `<>`     | Not equal to |
| `<`      | Less than |
| `>`      | Greater than |
| `<=`     | Less than or equal to |
| `>=`     | Greater than or equal to |

### Logical / Bitwise

| Operator | Operation |
|----------|-----------|
| `AND`    | Logical/bitwise AND |
| `OR`     | Logical/bitwise OR |
| `NOT`    | Logical/bitwise NOT |
| `XOR`    | Logical/bitwise XOR |
| `EQV`    | Logical equivalence |
| `IMP`    | Logical implication |

### String

| Operator | Operation |
|----------|-----------|
| `+`      | String concatenation |

### Operator Precedence (highest to lowest)

1. `^`
2. Unary `-`
3. `*`, `/`
4. `\`
5. `MOD`
6. `+`, `-`
7. Relational: `=`, `<>`, `<`, `>`, `<=`, `>=`
8. `NOT`
9. `AND`
10. `OR`
11. `XOR`
12. `EQV`
13. `IMP`

---

## 5. Control Flow

### IF / THEN / ELSEIF / ELSE / END IF

Single-line form:
```basic
IF condition THEN statement [ELSE statement]
```

Block form (Turbo BASIC extension):
```basic
IF condition THEN
    statements
ELSEIF condition THEN
    statements
ELSE
    statements
END IF
```

### SELECT CASE (Turbo BASIC extension)

```basic
SELECT CASE expression
    CASE value1
        statements
    CASE value2, value3         ' multiple values
        statements
    CASE value4 TO value5       ' range
        statements
    CASE IS > value6            ' comparison
        statements
    CASE ELSE
        statements
END SELECT
```

### FOR / NEXT

```basic
FOR counter = start TO end [STEP increment]
    statements
    [EXIT FOR]
NEXT [counter]
```

### WHILE / WEND

```basic
WHILE condition
    statements
WEND
```

### DO / LOOP (Turbo BASIC extension)

```basic
' Top-tested while
DO WHILE condition
    statements
    [EXIT DO]
LOOP

' Top-tested until
DO UNTIL condition
    statements
    [EXIT DO]
LOOP

' Bottom-tested while
DO
    statements
    [EXIT DO]
LOOP WHILE condition

' Bottom-tested until
DO
    statements
    [EXIT DO]
LOOP UNTIL condition

' Infinite loop
DO
    statements
    [EXIT DO]
LOOP
```

### GOTO

```basic
GOTO label
GOTO 1000          ' goto line number
```

### GOSUB / RETURN

```basic
GOSUB label
GOSUB 5000
...
label:
    statements
RETURN
```

### EXIT — Early exit from loops and procedures

```basic
EXIT FOR        ' exit FOR loop
EXIT DO         ' exit DO loop
EXIT WHILE      ' exit WHILE loop
EXIT SUB        ' exit subroutine
EXIT FUNCTION   ' exit function
EXIT DEF        ' exit DEF FN
```

### END / STOP / SYSTEM

```basic
END       ' terminate program normally
STOP      ' halt (resumable in IDE)
SYSTEM    ' exit to DOS
```

---

## 6. Subroutines and Functions

### SUB / END SUB (Turbo BASIC extension)

```basic
DECLARE SUB ProcName(param1 AS INTEGER, param2 AS STRING)

SUB ProcName(param1 AS INTEGER, param2 AS STRING)
    LOCAL localVar AS INTEGER
    STATIC persistVar AS LONG
    SHARED globalData
    ' statements
    EXIT SUB              ' optional early return
END SUB

CALL ProcName(arg1, arg2)
ProcName arg1, arg2       ' CALL keyword optional
```

Parameters are passed **by reference** by default. Use `BYVAL` for pass-by-value:

```basic
SUB Demo(BYVAL x AS INTEGER, y AS STRING)
```

### FUNCTION / END FUNCTION (Turbo BASIC extension)

```basic
DECLARE FUNCTION FuncName(param AS INTEGER) AS DOUBLE

FUNCTION FuncName(param AS INTEGER) AS DOUBLE
    LOCAL result AS DOUBLE
    result = param * 2.0
    FuncName = result       ' return value: assign to function name
    EXIT FUNCTION
END FUNCTION

result = FuncName(42)
```

### DEF FN / END DEF

Single-line (traditional BASIC):
```basic
DEF FNSquare(x) = x * x
result = FNSquare(5)
```

Multi-line (Turbo BASIC extension):
```basic
DEF FNCube(x)
    DEF FNCube = x * x * x
END DEF
```

### DECLARE — Forward reference

```basic
DECLARE SUB name(parameters)
DECLARE FUNCTION name(parameters) AS type
```

Required when calling a SUB or FUNCTION before its definition.

---

## 7. Math Functions

| Function     | Description |
|--------------|-------------|
| `ABS(x)`     | Absolute value |
| `SGN(x)`     | Sign: returns -1, 0, or 1 |
| `INT(x)`     | Floor (largest integer ≤ x) |
| `FIX(x)`     | Truncate fractional part toward zero |
| `CEIL(x)`    | Ceiling (smallest integer ≥ x) — Turbo BASIC extension |
| `SQR(x)`     | Square root |
| `EXP(x)`     | e raised to the power x |
| `EXP2(x)`    | 2 raised to the power x — Turbo BASIC extension |
| `EXP10(x)`   | 10 raised to the power x — Turbo BASIC extension |
| `LOG(x)`     | Natural logarithm (base e) |
| `LOG2(x)`    | Base-2 logarithm — Turbo BASIC extension |
| `LOG10(x)`   | Base-10 logarithm — Turbo BASIC extension |
| `SIN(x)`     | Sine (x in radians) |
| `COS(x)`     | Cosine (x in radians) |
| `TAN(x)`     | Tangent (x in radians) |
| `ATN(x)`     | Arctangent (returns radians) |
| `RND`        | Random number 0 ≤ n < 1 |
| `RANDOMIZE [seed]` | Seed the random number generator (statement) |

---

## 8. Type Conversion Functions

| Function   | Description |
|------------|-------------|
| `CINT(x)`  | Convert to INTEGER (rounds) |
| `CLNG(x)`  | Convert to LONG (rounds) |
| `CSNG(x)`  | Convert to SINGLE |
| `CDBL(x)`  | Convert to DOUBLE |
| `INT(x)`   | Convert to integer (floor) |
| `FIX(x)`   | Convert to integer (truncate toward zero) |

---

## 9. String Functions

| Function | Description |
|----------|-------------|
| `LEN(s$)` | Length of string |
| `LEFT$(s$, n)` | Leftmost n characters |
| `RIGHT$(s$, n)` | Rightmost n characters |
| `MID$(s$, start [, len])` | Substring starting at position `start` |
| `MID$(s$, start, len) = new$` | Replace substring (statement form) |
| `INSTR([start,] s$, find$)` | Position of `find$` in `s$`; 0 if not found |
| `ASC(s$)` | ASCII code of first character |
| `CHR$(n)` | Character with ASCII code n |
| `STR$(n)` | Convert number to string |
| `VAL(s$)` | Convert string to number |
| `HEX$(n)` | Hexadecimal string representation |
| `OCT$(n)` | Octal string representation |
| `BIN$(n)` | Binary string representation — Turbo BASIC extension |
| `UCASE$(s$)` | Convert to uppercase |
| `LCASE$(s$)` | Convert to lowercase |
| `LTRIM$(s$)` | Remove leading spaces |
| `RTRIM$(s$)` | Remove trailing spaces |
| `TRIM$(s$)` | Remove leading and trailing spaces — Turbo BASIC extension |
| `SPACE$(n)` | String of n spaces |
| `STRING$(n, c)` | String of n repetitions of character c |
| `LSET var$ = s$` | Left-justify string in fixed-length field |
| `RSET var$ = s$` | Right-justify string in fixed-length field |

### Random File Data Conversion

| Function   | Description |
|------------|-------------|
| `MKI$(n%)` | Convert INTEGER to 2-byte string |
| `MKL$(n&)` | Convert LONG to 4-byte string |
| `MKS$(n!)` | Convert SINGLE to 4-byte string |
| `MKD$(n#)` | Convert DOUBLE to 8-byte string |
| `CVI(s$)`  | Convert 2-byte string to INTEGER |
| `CVL(s$)`  | Convert 4-byte string to LONG |
| `CVS(s$)`  | Convert 4-byte string to SINGLE |
| `CVD(s$)`  | Convert 8-byte string to DOUBLE |

---

## 10. I/O and Screen Statements

### PRINT

```basic
PRINT [expr [, expr ...] [;]]
PRINT                    ' blank line
PRINT "Hello";           ' no newline
PRINT "A", "B"           ' tab-separated columns (zone width = 14)
PRINT TAB(20); "text"    ' move to column 20
PRINT SPC(5); "text"     ' insert 5 spaces
```

### PRINT USING

Format specifiers:

| Character | Meaning |
|-----------|---------|
| `#`       | Digit placeholder |
| `.`       | Decimal point |
| `,`       | Thousands separator |
| `+`       | Force sign (before or after `#`) |
| `-`       | Trailing minus for negatives |
| `^^^^`    | Scientific notation (4 carets = exponent field) |
| `$`       | Dollar sign |
| `$$`      | Floating dollar sign |
| `**`      | Fill leading spaces with asterisks |
| `**$`     | Fill with asterisks and floating dollar sign |
| `!`       | Single character of string |
| `\   \`   | String field (n+2 chars, n spaces between backslashes) |
| `&`       | Variable-length string field |
| `_`       | Output next character literally |

```basic
PRINT USING "###.##"; value
PRINT USING "$$###,###.##"; amount
PRINT USING "!\"; name$
```

### LPRINT / LPRINT USING

Same as `PRINT` / `PRINT USING` but outputs to the printer (LPT1).

### INPUT

```basic
INPUT [prompt;] var [, var ...]
INPUT "Enter name: "; name$
LINE INPUT [prompt;] var$     ' read entire line including commas
```

### INKEY$

```basic
k$ = INKEY$    ' empty string if no key pressed; otherwise key character
```

### GET$ (Turbo BASIC extension)

```basic
GET$ n, var$   ' read n characters from keyboard buffer
```

### INSTAT (Turbo BASIC extension)

```basic
IF INSTAT THEN ...   ' true (non-zero) if a key is waiting in the buffer
```

### LOCATE

```basic
LOCATE [row] [, col] [, cursor] [, start] [, stop]
```

### CLS

```basic
CLS      ' clear screen
CLS 0    ' clear both text and graphics viewport
CLS 1    ' clear text viewport
CLS 2    ' clear graphics viewport
```

### WRITE

```basic
WRITE [expr [, expr ...]]   ' output with commas; strings enclosed in quotes
```

### COLOR

```basic
COLOR [foreground] [, background] [, border]   ' text mode
COLOR [foreground] [, background]              ' graphics mode
```

### WIDTH

```basic
WIDTH [columns]           ' set screen width (40 or 80)
WIDTH "LPT1:", n          ' set printer width
```

### KEY

```basic
KEY n, string$   ' assign string to function key n (1-10, 15=F11, 16=F12)
KEY ON           ' display function key bar at bottom of screen
KEY OFF          ' hide function key bar
KEY LIST         ' list all key assignments
```

### Cursor Position Functions

```basic
row = CSRLIN        ' current cursor row (1-based)
col = POS(0)        ' current cursor column (1-based)
```

### SCREEN (read screen content)

```basic
charCode = SCREEN(row, col)      ' ASCII code at text position
attr     = SCREEN(row, col, 1)   ' attribute byte at text position
```

---

## 11. File I/O

### OPEN

```basic
OPEN filename$ FOR mode AS [#]filenum [LEN = reclen]

OPEN "data.txt"  FOR INPUT  AS #1
OPEN "out.txt"   FOR OUTPUT AS #1
OPEN "log.txt"   FOR APPEND AS #1
OPEN "data.dat"  FOR RANDOM AS #1 LEN = 64
OPEN "raw.bin"   FOR BINARY AS #1
```

### CLOSE

```basic
CLOSE [#filenum [, #filenum ...]]
CLOSE    ' close all open files
```

### Sequential File I/O

```basic
INPUT #n, var [, var ...]              ' read comma-delimited values
LINE INPUT #n, var$                    ' read entire line
PRINT #n, [expr [; expr ...]]          ' write to file
PRINT #n, USING fmt$; expr             ' formatted write
WRITE #n, [expr [, expr ...]]          ' write with delimiters and quotes
```

### Random Access File I/O

```basic
FIELD #n, len1 AS var1$, len2 AS var2$   ' define record layout
GET #n [, record]                         ' read record into FIELD variables
PUT #n [, record]                         ' write FIELD variables to record
LSET var$ = value$                        ' left-justify value into field
RSET var$ = value$                        ' right-justify value into field
```

### Binary File I/O

```basic
GET #n [, position], var    ' read variable from byte position
PUT #n [, position], var    ' write variable to byte position
```

### File Position and Status

```basic
SEEK #n, position    ' set file position (statement)
pos    = SEEK(n)     ' get current file position (function)
eof    = EOF(n)      ' -1 if at end of file, 0 otherwise
loc    = LOC(n)      ' current record/block position
length = LOF(n)      ' total length of file in bytes
```

### File System Operations

```basic
KILL filename$           ' delete file
NAME old$ AS new$        ' rename file
FILES [pattern$]         ' display directory listing
f$ = DIR$(pattern$)      ' return next filename matching pattern
CHDIR path$              ' change current directory
MKDIR path$              ' create directory
RMDIR path$              ' remove directory
```

### Binary Memory Load/Save

```basic
BLOAD filename$ [, offset]         ' load binary file into memory
BSAVE filename$, offset, length    ' save memory region to file
```

### Device I/O

```basic
IOCTL #n, string$    ' send control string to device driver
s$ = IOCTL$(n)       ' receive control string from device driver
```

---

## 12. Graphics

### SCREEN — Set video mode

```basic
SCREEN mode [, colorswitch] [, apage] [, vpage]

SCREEN 0    ' text mode (80×25 or 40×25)
SCREEN 1    ' CGA 320×200, 4 colors
SCREEN 2    ' CGA 640×200, 2 colors
SCREEN 3    ' Hercules 720×348
SCREEN 4    ' EGA 640×350
SCREEN 7    ' EGA 320×200, 16 colors
SCREEN 8    ' EGA 640×200, 16 colors
SCREEN 9    ' EGA 640×350, up to 64 colors
SCREEN 10   ' EGA monochrome 640×350
```

### PSET / PRESET — Plot a pixel

```basic
PSET (x, y) [, color]            ' plot pixel in color
PRESET (x, y) [, color]          ' plot pixel (default background color)
PSET STEP (dx, dy) [, color]     ' relative from current position
```

### LINE — Draw line or rectangle

```basic
LINE [(x1, y1)] - (x2, y2) [, color] [, B[F]]
' B  = draw box outline
' BF = draw filled box
LINE STEP (dx, dy) ...            ' endpoint relative to current position
```

### CIRCLE — Draw circle, arc, or ellipse

```basic
CIRCLE (x, y), radius [, color] [, start] [, end] [, aspect]
' start/end = arc angles in radians
' aspect    = y/x ratio (1 = circle; other values = ellipse)
```

### PAINT — Flood fill

```basic
PAINT (x, y) [, fillColor] [, borderColor]
```

### DRAW — Graphics macro language

```basic
DRAW string$
```

| Command   | Description |
|-----------|-------------|
| `Un`      | Up n steps |
| `Dn`      | Down n steps |
| `Ln`      | Left n steps |
| `Rn`      | Right n steps |
| `En`      | Diagonal up-right n steps |
| `Fn`      | Diagonal down-right n steps |
| `Gn`      | Diagonal down-left n steps |
| `Hn`      | Diagonal up-left n steps |
| `Mx,y`    | Move to absolute (x, y) |
| `M+x,+y`  | Move relative |
| `Bn`      | Blind prefix: move without drawing |
| `Nn`      | No-update prefix: move and return |
| `Cn`      | Set color n |
| `Sn`      | Scale factor (1–255; 4 = normal) |
| `An`      | Angle: 0=0°, 1=90°, 2=180°, 3=270° |
| `TAn`     | Turn angle: -360 to +360 degrees |
| `Ppaint,border` | Paint fill |
| `Xvar$`   | Execute substring |

```basic
DRAW "U20 R20 D20 L20"              ' draw a square
DRAW "C2 BM100,100 U50 R50 D50 L50" ' move to (100,100), draw colored square
```

### GET / PUT — Sprite capture and restore

```basic
GET (x1, y1)-(x2, y2), array%()          ' capture screen region to array
PUT (x1, y1), array%() [, action]         ' restore region from array
' action: PSET, PRESET, AND, OR, XOR (default)
```

### VIEW — Define viewport

```basic
VIEW [(x1, y1)-(x2, y2)] [, fillColor] [, borderColor]
VIEW PRINT [top TO bottom]    ' define text scroll region
```

### WINDOW — Logical coordinate system

```basic
WINDOW [(x1, y1)-(x2, y2)]    ' logical coordinates (y increases upward)
WINDOW SCREEN [(x1, y1)-(x2, y2)]  ' y increases downward
```

### PALETTE

```basic
PALETTE [attrib, color]        ' remap single palette entry
PALETTE USING array%()         ' set entire palette from array
```

### POINT — Read pixel color or position

```basic
c  = POINT(x, y)    ' color value at (x, y)
x  = POINT(0)       ' current graphics x (screen coords)
y  = POINT(1)       ' current graphics y (screen coords)
x  = POINT(2)       ' current graphics x (window coords)
y  = POINT(3)       ' current graphics y (window coords)
```

### PMAP — Coordinate mapping

```basic
xScreen = PMAP(xWin, 0)    ' window x to screen x
yScreen = PMAP(yWin, 1)    ' window y to screen y
xWin    = PMAP(xScreen, 2) ' screen x to window x
yWin    = PMAP(yScreen, 3) ' screen y to window y
```

---

## 13. Sound

### BEEP

```basic
BEEP    ' short 800 Hz beep
```

### SOUND

```basic
SOUND frequency, duration
' frequency: 37 to 32767 Hz
' duration:  timer ticks (18.2 ticks per second)

SOUND 440, 18.2    ' 440 Hz (middle A) for 1 second
SOUND 0, 9.1       ' silence for 0.5 seconds
```

### PLAY — Music Macro Language (MML)

```basic
PLAY string$
```

| Command      | Description |
|--------------|-------------|
| `A`–`G`      | Play note |
| `#` or `+`   | Sharp (follows note letter) |
| `-`          | Flat (follows note letter) |
| `On`         | Set octave 0–6 (O4 C = middle C) |
| `<`          | Down one octave |
| `>`          | Up one octave |
| `Ln`         | Note length: 1=whole, 2=half, 4=quarter, 8=eighth... |
| `Pn`         | Pause (rest) of length n |
| `Tn`         | Tempo in beats per minute (32–255; default 120) |
| `Nn`         | Play note by number 0–84 |
| `MS`         | Staccato style |
| `MN`         | Normal style |
| `ML`         | Legato style |
| `MF`         | Foreground: wait for all notes to finish |
| `MB`         | Background: continue program while music plays |

```basic
PLAY "T120 O4 L4 C D E F G A B C5"   ' C major scale at 120 BPM
PLAY "MB T180 L8 CDEFEDC"             ' background playback
```

---

## 14. Data Statements

```basic
DATA value1, value2, value3, ...   ' embed literal data in source
READ var1, var2, var3              ' read sequentially from DATA
RESTORE [label]                    ' reset DATA pointer to beginning or label
```

---

## 15. Memory and System Access

### PEEK / POKE

```basic
b = PEEK(address)       ' read byte from memory address
POKE address, value     ' write byte to memory address
```

### DEF SEG — Set memory segment

```basic
DEF SEG [= segment]    ' set segment register for PEEK/POKE/BLOAD/BSAVE
DEF SEG               ' reset to default segment
```

### VARPTR / VARSEG

```basic
offset  = VARPTR(var)    ' offset address of variable
segment = VARSEG(var)    ' segment address of variable
s$      = VARPTR$(var)   ' address packed as string (for CALL)
```

### INP / OUT — Hardware port access

```basic
b = INP(port)              ' read byte from I/O port
OUT port, value            ' write byte to I/O port
WAIT port, mask [, xormask]  ' wait until (port AND mask) = (xormask AND mask)
```

### CALL ABSOLUTE (Turbo BASIC extension)

```basic
CALL ABSOLUTE(arg1, arg2, ..., address)
' Call machine code subroutine at the given memory address
```

### CALL INTERRUPT / INTERRUPTX (Turbo BASIC extension)

```basic
DIM regs AS InterruptRegs
CALL INTERRUPT(intNum, regs)             ' invoke DOS/BIOS interrupt
CALL INTERRUPTX(intNum, inRegs, outRegs) ' separate in/out register sets
```

### Memory Info

```basic
bytes = FRE(0)     ' free string space
bytes = FRE(-1)    ' free far heap memory
bytes = FRE(n)     ' compact heap, return free memory
addr  = ENDMEM     ' highest address used by program (Turbo BASIC extension)
```

### MEMSET (Turbo BASIC extension)

```basic
MEMSET address, value, count    ' fill memory block with a single byte value
```

### Miscellaneous Statements

```basic
SWAP var1, var2              ' exchange values of two variables (same type)
INCR var [, amount]          ' increment by 1 or by amount (Turbo BASIC extension)
DECR var [, amount]          ' decrement by 1 or by amount (Turbo BASIC extension)
SHIFT variable, n            ' logical bit shift: left if n>0, right if n<0
ROTATE variable, n           ' bit rotate: left if n>0, right if n<0
TRON                         ' enable trace mode (print line numbers as executed)
TROFF                        ' disable trace mode
```

---

## 16. Error Handling

```basic
ON ERROR GOTO label    ' enable error trapping; jump to handler on error
ON ERROR GOTO 0        ' disable error trapping

' Inside error handler:
e     = ERR       ' error code
line  = ERL       ' line number where error occurred
dev   = ERDEV     ' device error code
dev$  = ERDEV$    ' device name string

RESUME            ' resume at statement that caused the error
RESUME NEXT       ' resume at statement after the error
RESUME label      ' resume at a specific label

ERROR n           ' simulate error n (for testing)
```

---

## 17. Event Trapping

```basic
ON KEY(n)    GOSUB label    ' trap function key n (1-10, 15=F11, 16=F12)
ON TIMER(n)  GOSUB label    ' trap timer every n seconds
ON COM(n)    GOSUB label    ' trap serial port n (1 or 2) receive
ON PEN       GOSUB label    ' trap light pen
ON STRIG(n)  GOSUB label    ' trap joystick button n (0, 2, 4, 6)
ON PLAY(n)   GOSUB label    ' trap when music queue drops below n notes

' Enable / disable / suspend trapping:
KEY(n)   ON / OFF / STOP
TIMER    ON / OFF / STOP
COM(n)   ON / OFF / STOP
PEN      ON / OFF / STOP
STRIG(n) ON / OFF / STOP
PLAY     ON / OFF / STOP
```

---

## 18. Timing

```basic
t  = TIMER       ' seconds since midnight (SINGLE precision)
t& = MTIMER      ' milliseconds elapsed (Turbo BASIC extension)
MTIMER           ' reset millisecond timer (statement form)
DELAY seconds    ' pause execution for given seconds (Turbo BASIC extension)
DELAY 0.5        ' pause 500 ms
```

---

## 19. System Variables

```basic
DATE$                   ' current date as "MM-DD-YYYY"
DATE$ = "MM-DD-YYYY"    ' set system date
TIME$                   ' current time as "HH:MM:SS"
TIME$ = "HH:MM:SS"      ' set system time
COMMAND$                ' command-line arguments string
ENVIRON$("name")        ' read environment variable
ENVIRON "name=value"    ' set environment variable (statement)
```

---

## 20. Input Devices

```basic
PEN(n)      ' light pen: n=0-9, returns various status values
STICK(n)    ' joystick position: n=0,1 (stick A x,y); n=2,3 (stick B x,y)
STRIG(n)    ' joystick trigger: n=0,2 (A,B current); n=1,3 (A,B last state)
```

---

## 21. Program Control and Chaining

```basic
RUN [filename$]                             ' run another program (or restart this one)
CHAIN filename$ [, line] [, ALL] [, DELETE range]
' ALL = pass all variables via COMMON
SHELL [command$]                            ' execute DOS command and return
SYSTEM                                      ' exit program, return to DOS
```

---

## 22. Metacommands (Compiler Directives)

Metacommands begin with `$` and are processed at compile time (not at runtime). They may appear as comments or as standalone lines.

| Metacommand | Description |
|-------------|-------------|
| `$DYNAMIC`  | Make subsequent arrays dynamic (heap-allocated) |
| `$STATIC`   | Make subsequent arrays static (compile-time allocated) |
| `$INCLUDE "filename"` | Insert source file at this point |
| `$IF expr`  | Begin conditional compilation |
| `$ELSEIF expr` | Conditional compilation alternative |
| `$ELSE`     | Conditional compilation else branch |
| `$ENDIF`    | End conditional compilation block |
| `$COM n`    | Set serial port receive buffer size (bytes) |
| `$SOUND n`  | Set background music buffer size (notes) |
| `$STACK n`  | Set program stack size (bytes) |
| `$SEGMENT`  | Define memory segment layout |
| `$INLINE data` | Embed raw machine code bytes inline |
| `$EVENT ON` | Enable background event checking |
| `$EVENT OFF` | Disable background event checking |

```basic
$INCLUDE "mylib.bas"

$IF Version > 1
    PRINT "Extended version"
$ELSE
    PRINT "Basic version"
$ENDIF

$DYNAMIC
DIM bigArray(10000)    ' dynamically allocated

$STACK 4096            ' set stack to 4 KB

$INLINE &H90, &H90     ' embed NOP NOP machine code
```

---

## 23. Turbo BASIC Extensions vs. GW-BASIC

The following features were **added** by Turbo BASIC beyond standard GW-BASIC / BASICA:

### New Data and Operations
- `LONG` integer type (`&` suffix, 32-bit signed)
- `BIN$()` — binary string representation
- `TRIM$()`, `LTRIM$()`, `RTRIM$()` — string trimming
- `LCASE$()`, `UCASE$()` — case conversion
- `CEIL()` — ceiling function
- `EXP2()`, `EXP10()`, `LOG2()`, `LOG10()` — extended math functions
- `INCR`, `DECR` — increment/decrement statements
- `SHIFT`, `ROTATE` — bit manipulation statements
- `MTIMER` — millisecond timer
- `DELAY` — sub-second pause
- `ENDMEM` — highest memory address used
- `MEMSET` — fill memory block
- `INSTAT` — keyboard status check
- `GET$` — buffered keyboard read

### Structured Programming
- `SUB` / `END SUB` — named procedures with local variables
- `FUNCTION` / `END FUNCTION` — typed functions
- `DECLARE` — forward reference for procedures
- `EXIT FOR`, `EXIT DO`, `EXIT WHILE`, `EXIT SUB`, `EXIT FUNCTION` — early exit
- `DO` / `LOOP` with `WHILE` or `UNTIL` at top or bottom
- Multi-line `IF` / `ELSEIF` / `ELSE` / `END IF`
- `SELECT CASE` / `END SELECT`
- Multi-line `DEF FN` / `END DEF`
- `LOCAL`, `STATIC`, `SHARED` variable scope modifiers
- `BYVAL` parameter modifier
- Recursion support
- Alphanumeric labels (line numbers not required)

### Compiler Features
- `$DYNAMIC` / `$STATIC` metacommands
- `$INCLUDE` source file inclusion
- `$IF` / `$ELSE` / `$ENDIF` conditional compilation
- `$COM`, `$SOUND`, `$STACK`, `$SEGMENT` configuration
- `$INLINE` inline machine code
- `$EVENT` event checking control
- `CALL ABSOLUTE` — machine code call
- `CALL INTERRUPT` / `CALL INTERRUPTX` — DOS/BIOS interrupt access

---

## 24. Common Error Codes

| Code | Meaning |
|------|---------|
| 1  | NEXT without FOR |
| 2  | Syntax error |
| 3  | RETURN without GOSUB |
| 4  | Out of DATA |
| 5  | Illegal function call |
| 6  | Overflow |
| 7  | Out of memory |
| 9  | Subscript out of range |
| 10 | Array already DIMensioned |
| 11 | Division by zero |
| 13 | Type mismatch |
| 14 | Out of string space |
| 20 | RESUME without error |
| 24 | Device timeout |
| 25 | Device fault |
| 27 | Out of paper |
| 52 | Bad file number |
| 53 | File not found |
| 54 | Bad file mode |
| 55 | File already open |
| 58 | File already exists |
| 61 | Disk full |
| 62 | Input past end of file |
| 63 | Bad record number |
| 64 | Bad filename |
| 67 | Too many files |
| 68 | Device unavailable |
| 71 | Disk not ready |
| 72 | Disk media error |
| 76 | Path not found |

---

## References

- Borland Turbo BASIC Owner's Handbook (1987) — [Internet Archive](https://archive.org/details/bitsavers_borlandBorsHandbook1987_15768512)
- PowerBASIC (successor to Turbo BASIC) — [Wikipedia](https://en.wikipedia.org/wiki/PowerBASIC)
- Music Macro Language — [Wikipedia](https://en.wikipedia.org/wiki/Music_Macro_Language)
