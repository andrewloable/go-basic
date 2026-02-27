# Code Generation Patterns

Detailed Go emission patterns for translating BASIC constructs, plus known pitfalls.

## Type Mapping Rules

### Variable Type Resolution

The codegen determines a variable's Go type from its BASIC name suffix using `goTypeForIdent()`:

```
X%    → int16      (suffix _pct)
X&    → int32      (suffix _lng)
X!    → float32    (suffix _sng)
X#    → float64    (suffix _dbl)
X$    → string     (suffix _str)
X     → float32    (no suffix, BASIC default is SINGLE)
```

### Type Casting Rules

BASIC is loosely typed — any numeric type can be assigned to any other. Go requires explicit casts.

**Assignment**: Always cast RHS to LHS type:
```go
// BASIC: A% = 3.14
var A_pct int16 = int16(3.14)
```

**Binary operations**: Both operands must match. Cast narrower to wider:
```go
// BASIC: result = X% + Y!
result = float32(X_pct) + Y_sng
```

**Runtime function calls**: Math functions accept/return float64:
```go
// BASIC: Y! = SIN(X!)
Y_sng = float32(rt.Sin(float64(X_sng)))
```

**DATA/READ type-switch**: Must cast `tv_` (always float64) to target:
```go
// Reading into an int16 variable
case float64: A_pct = int16(tv_)
// Reading into a float32 variable
case float64: S_sng = float32(tv_)
```

## Statement Emission Patterns

### LET (Variable Assignment)

```basic
' BASIC:
LET X! = 42
Y% = X! + 1
```

```go
// First occurrence (declaration):
var X_sng float32 = float32(42)
// Subsequent (assignment):
Y_pct = int16(float32(X_sng) + float32(1))
```

Track declared variables in `g.declared` map. First use → `var`, subsequent → plain `=`.

### FOR/NEXT

```basic
FOR I = 1 TO 10 STEP 2
    PRINT I
NEXT I
```

```go
end_0 := float32(10)
step_0 := float32(2)
for I = float32(1); (step_0 > 0 && I <= end_0) || (step_0 < 0 && I >= end_0) || (step_0 == 0); I += step_0 {
    fmt.Println(I)
}
```

The complex condition handles positive, negative, and zero step values. The end/step vars
and all cast expressions must use the counter's Go type (not always float32).

### IF/THEN/ELSE

```basic
IF X > 0 THEN
    PRINT "yes"
ELSEIF X = 0 THEN
    PRINT "zero"
ELSE
    PRINT "no"
END IF
```

```go
if X > float32(0) {
    fmt.Println("yes")
} else if X == float32(0) {
    fmt.Println("zero")
} else {
    fmt.Println("no")
}
```

Note: BASIC `=` in comparisons → Go `==`.

### SELECT CASE

Emitted as a Go tagless switch (not if/else chain) so that `break` statements inside case bodies compile correctly:

```go
switch {
case Score_pct >= int16(90) && Score_pct <= int16(100):
    Grade_str = "A"
case Score_pct >= int16(80) && Score_pct <= int16(89):
    Grade_str = "B"
default:
    Grade_str = "F"
}
```

### DO/LOOP

```go
// DO WHILE condition ... LOOP
for condition {
    // body
}

// DO ... LOOP UNTIL condition
for {
    // body
    if condition { break }
}

// DO ... LOOP (infinite)
for {
    // body
}
```

### GOTO / Labels

```go
goto label_MyLabel
// ...
label_MyLabel:
```

Only emit labels that are referenced by GOTO/GOSUB (tracked in `g.referencedLabels`).
Unreferenced labels cause Go compilation errors.

### ON computed GOTO/GOSUB

```basic
ON N GOTO label1, label2, label3
```

```go
switch int(N) {
case 1:
    goto label_label1
case 2:
    goto label_label2
case 3:
    goto label_label3
}
```

Label names use `g.labelName(target)` not `mangleName(target)`.

### GOSUB / RETURN

GOSUB is emitted as `goto` to the label. RETURN is complex because Go doesn't have a
GOSUB return stack. Current approach: emit the label and hope the control flow works.
A proper fix would use a state machine or closure-based approach.

## Procedure Emission Patterns

### SUB

```basic
SUB Greet (Name$)
    PRINT "Hello "; Name$
END SUB
```

Should emit as a Go function (NOT a closure variable):
```go
func Greet(Name_str string) {
    fmt.Print("Hello ", Name_str)
}
```

**Known bug**: Currently emitted as closures (`Greet := func(...){}`) causing
declared-and-not-used errors. Tracked in task `rn3.1`.

### FUNCTION

```basic
FUNCTION Square# (X#)
    Square# = X# * X#
END FUNCTION
```

```go
func Square_dbl(X_dbl float64) float64 {
    var Square_dbl float64
    Square_dbl = X_dbl * X_dbl
    return Square_dbl
}
```

Assignment to the function name inside the body = setting the return value.

### DEF FN

```basic
DEF FNCtoF(C) = C * 9 / 5 + 32
```

```go
FNCtoF := func(C float32) float32 { return C * 9.0 / 5.0 + 32.0 }
```

Return type must match the function name's suffix. Recursive DEF FN: assignments to the
function name inside the body must become `return` statements.

## Builtin Function Emission

### Standard Pattern (emitFunctionCall switch)

```go
case "SIN":
    return fmt.Sprintf("rt.Sin(float64(%s))", args[0])
case "LEFT$":
    return fmt.Sprintf("rt.Left(%s, int(%s))", args[0], args[1])
case "CHR$":
    return fmt.Sprintf("func() string { v, _ := rt.Chr(int(%s)); return v }()", args[0])
```

Functions returning `(value, error)` must be wrapped in a closure to discard the error in
expression context.

### No-Argument Builtins (emitIdentifier)

These are Identifier nodes, not FunctionCall nodes. Handle in emitIdentifier():

```go
case "INKEY$": return "rt.Inkey()"
case "DATE$":  return "rt.DateStr()"
case "TIME$":  return "rt.TimeStr()"
case "RND":    g.needRng = true; return "rng.Rnd(1)"
case "ERR":    return "float64(errState.Err())"
```

### The $-Suffix ArrayAccess Bug

When $-suffixed builtins end up as ArrayAccess nodes, the codegen produces:
```go
CHR_str[int(65)]     // WRONG — should be rt.Chr(int(65))
LEFT_str[int(s)]     // WRONG — should be rt.Left(s, int(5))
```

Fix: either fix the parser to always create FunctionCall nodes for builtins, or add a
guard in emitArrayAccess() to detect builtin names and reroute.

## DATA / READ Pattern

DATA values are collected into a flat `[]interface{}` pool at compile time:

```go
var dataPool []interface{}
_ = dataPool
dataPool = []interface{}{
    "Smith",
    float64(42),
    35000.5,
}
dataIdx := 0
_ = dataIdx

// READ Name$
Name_str = fmt.Sprint(dataPool[dataIdx]); dataIdx++
// READ Age%
{ v_ := dataPool[dataIdx]; dataIdx++; switch tv_ := v_.(type) {
    case float64: Age_pct = int16(tv_)  // MUST cast to target type
    case string: Age_pct = int16(rt.Val(tv_))
}}
```

## PRINT Emission

```basic
PRINT "Hello"; X; TAB(20); Y
PRINT USING "##.##"; value
```

PRINT items separated by `;` have no space between them. Items separated by `,` use tab zones.
A trailing `;` suppresses the newline.

```go
fmt.Print("Hello")
fmt.Print(X)
fmt.Print(rt.Tab(col, 20))
fmt.Print(Y)
fmt.Println()  // no trailing ; → newline
```

## File I/O Pattern

Requires a FileManager instance:

```go
fm := rt.NewFileManager()
defer fm.FileCloseAll()

// OPEN "data.txt" FOR OUTPUT AS #1
fm.FileOpen(1, "data.txt", rt.FileModeOutput, 0)

// PRINT #1, Name$, Age%
fm.FilePrint(1, false, Name_str, Age_pct)

// CLOSE #1
fm.FileClose(1)
```

## Common Pitfalls

### 1. Untyped vs Typed Constants

Go integer literals are untyped and can be assigned to any numeric type:
```go
var x float32 = 42      // OK — untyped constant
var x float32 = float64(42)  // ERROR — float64 is a TYPED constant
```

`formatGoNumber()` should emit `42` not `float64(42)` for integer constants.

### 2. Goto Over Declaration

Go forbids `goto` jumping over a `var` declaration. If SUBs are emitted as closure variables
(`Greet := func(){}`) and a GOTO jumps past them, compilation fails. Fix: emit SUBs as
top-level Go functions, not closures.

### 3. Break Not In Loop

When SELECT CASE was emitted as if/else chains, `break` (from EXIT SELECT or CASE
separation) had no enclosing loop/switch. Fix: emit as `switch {}` (tagless switch).

### 4. Unused Labels

Go treats unused labels as compilation errors. Only emit labels that appear in
`g.referencedLabels`. ON ERROR GOTO targets should NOT be added to referencedLabels
if the ON ERROR statement itself is emitted as a TODO comment.

### 5. Runtime Function Signatures

All `rt.*` math functions take and return `float64`. Always cast args:
```go
rt.Sin(float64(x_sng))     // NOT rt.Sin(x_sng)
```

String functions take specific types — check the signature:
```go
rt.Left(s, int(n))          // second arg is int, not float64
rt.Mid(s, int(start), int(length))
rt.Chr(int(n))              // returns (string, error)
```

### 6. Multi-Dimensional Arrays

Currently only 1-D arrays are fully supported. 2-D arrays need slice-of-slices:
```go
// DIM A(10, 20)
A := make([][]float32, 11)
for i := range A { A[i] = make([]float32, 21) }

// A(3, 5) = 42
A[int(3)][int(5)] = float32(42)
```
