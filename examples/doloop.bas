' DO/LOOP Statement - Various Loop Forms
' Page: 202-203 (Chapter 5 - DO/LOOP statement)

' DO WHILE...LOOP (test at top)
x = 1
DO WHILE x <= 10
  PRINT x;
  x = x + 1
LOOP
PRINT

' DO UNTIL...LOOP (test at top)
x = 1
DO UNTIL x > 10
  PRINT x;
  x = x + 1
LOOP
PRINT

' DO...LOOP WHILE (test at bottom)
x = 1
DO
  PRINT x;
  x = x + 1
LOOP WHILE x <= 10
PRINT

' DO...LOOP UNTIL (test at bottom)
x = 1
DO
  PRINT x;
  x = x + 1
LOOP UNTIL x > 10
