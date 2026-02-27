REM Test SUB and FUNCTION

DECLARE SUB PrintLine (msg$)
DECLARE FUNCTION Factorial (n)

CALL PrintLine("Testing subroutines")

FOR i = 1 TO 10
  PRINT i; "! ="; Factorial(i)
NEXT i

CALL PrintLine("Done!")
END

SUB PrintLine (msg$)
  PRINT "--- "; msg$; " ---"
END SUB

FUNCTION Factorial (n)
  IF n <= 1 THEN
    Factorial = 1
  ELSE
    Factorial = n * Factorial(n - 1)
  END IF
END FUNCTION
