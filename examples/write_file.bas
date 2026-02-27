' WRITE# to File - Employee Data
' Page: 380-381 (Chapter 5 - WRITE# statement)
OPEN "DATA.TXT" FOR OUTPUT AS #1
FOR I = 1 TO 5
  READ N$, A%, S!
  WRITE #1, N$, A%, S!
NEXT I
CLOSE #1
DATA "Smith", 42, 35000.50
DATA "Jones", 35, 42000.00
DATA "Brown", 28, 28500.75
DATA "Davis", 51, 55000.00
DATA "Wilson", 33, 38000.25
