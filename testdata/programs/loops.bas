REM Test all loop constructs

' FOR/NEXT
PRINT "FOR loop:"
FOR i = 1 TO 5
  PRINT i;
NEXT i
PRINT

' FOR with STEP
PRINT "FOR with STEP:"
FOR i = 10 TO 1 STEP -2
  PRINT i;
NEXT i
PRINT

' WHILE/WEND
PRINT "WHILE loop:"
x = 1
WHILE x <= 5
  PRINT x;
  x = x + 1
WEND
PRINT

' DO WHILE/LOOP
PRINT "DO WHILE loop:"
x = 1
DO WHILE x <= 5
  PRINT x;
  x = x + 1
LOOP
PRINT

' DO/LOOP UNTIL
PRINT "DO UNTIL loop:"
x = 1
DO
  PRINT x;
  x = x + 1
LOOP UNTIL x > 5
PRINT

END
