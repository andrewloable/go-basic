' Recursive Factorial Function
' Page: 109-110 (Chapter 4 - Recursion)
DEF FNFactorial#(n%)
  IF n% > 1 AND n% <= 170 THEN
    FNFactorial# = n% * FNFactorial#(n%-1)
  ELSEIF n% = 0 OR n% = 1 THEN
    FNFactorial# = 1
  ELSE
    FNFactorial# = -1
  END IF
END DEF
PRINT FNFactorial#(3)
