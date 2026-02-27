' TIMER Function - Benchmarking
' Page: 374 (Chapter 5 - TIMER function)
Start! = TIMER
sum# = 0
FOR I& = 1 TO 100000
  sum# = sum# + SQR(I&)
NEXT I&
Elapsed! = TIMER - Start!
PRINT "Sum: "; sum#
PRINT "Time: "; Elapsed!; " seconds"
