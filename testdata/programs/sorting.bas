REM Bubble sort
DIM a(10)

' Initialize with random-ish values
a(1) = 64 : a(2) = 34 : a(3) = 25
a(4) = 12 : a(5) = 22 : a(6) = 11
a(7) = 90 : a(8) = 1  : a(9) = 45
a(10) = 7

PRINT "Before sort:"
FOR i = 1 TO 10
  PRINT a(i);
NEXT i
PRINT

' Bubble sort
FOR i = 1 TO 9
  FOR j = 1 TO 10 - i
    IF a(j) > a(j + 1) THEN
      SWAP a(j), a(j + 1)
    END IF
  NEXT j
NEXT i

PRINT "After sort:"
FOR i = 1 TO 10
  PRINT a(i);
NEXT i
PRINT

END
