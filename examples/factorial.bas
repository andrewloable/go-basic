' Factorial Function (Multiline, Non-Recursive)
' Page: 99 (Chapter 4 - Multiline Functions)
DEF FNFactorial#(x%)
  LOCAL i%, total#
  IF x% < 0 OR x% > 170 THEN FNFactorial# = -1 : EXIT DEF
  total# = 1
  FOR i% = x% TO 2 STEP -1
    total# = total# * i%
  NEXT i%
  FNFactorial# = total#
END DEF
PRINT FNFactorial#(52)
