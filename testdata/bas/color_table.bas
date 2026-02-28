DIM ColorName$ AS STRING
DIM ColorCode AS INTEGER

ColorName$ = "Red"
ColorCode = 1

CALL PrintColor

ColorName$ = "Blue"
ColorCode = 2

CALL PrintColor

SUB PrintColor
  SHARED ColorName$, ColorCode
  PRINT ColorName$; " = "; ColorCode
END SUB
