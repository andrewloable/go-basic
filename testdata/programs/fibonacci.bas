REM Fibonacci sequence
DIM fib(20)
fib(0) = 0
fib(1) = 1
FOR i = 2 TO 20
  fib(i) = fib(i - 1) + fib(i - 2)
NEXT i
PRINT "Fibonacci sequence:"
FOR i = 0 TO 20
  PRINT fib(i);
NEXT i
PRINT
END
