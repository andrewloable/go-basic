10 REM Fibonacci Sequence
20 DIM a AS LONG, b AS LONG, temp AS LONG
30 LET a = 0
40 LET b = 1
50 PRINT "Fibonacci Sequence:"
60 FOR i = 1 TO 20
70     PRINT a;
80     temp = a + b
90     a = b
100    b = temp
110 NEXT i
120 PRINT
130 END
