REM Test arrays
DIM a(10)
DIM matrix(3, 3)

' Fill 1D array
FOR i = 1 TO 10
  a(i) = i * i
NEXT i

' Print 1D array
PRINT "Squares:"
FOR i = 1 TO 10
  PRINT a(i);
NEXT i
PRINT

' Fill 2D array (multiplication table)
FOR i = 1 TO 3
  FOR j = 1 TO 3
    matrix(i, j) = i * j
  NEXT j
NEXT i

' Print 2D array
PRINT "Multiplication table:"
FOR i = 1 TO 3
  FOR j = 1 TO 3
    PRINT matrix(i, j);
  NEXT j
  PRINT
NEXT i

END
