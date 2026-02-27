' SELECT CASE Statement
' Page: 352-353 (Chapter 5 - SELECT statement)
INPUT "Enter a number (1-100): ", n
SELECT CASE n
  CASE 1 TO 25
    PRINT "First quarter"
  CASE 26 TO 50
    PRINT "Second quarter"
  CASE 51 TO 75
    PRINT "Third quarter"
  CASE 76 TO 100
    PRINT "Fourth quarter"
  CASE IS < 1
    PRINT "Too small!"
  CASE IS > 100
    PRINT "Too big!"
  CASE ELSE
    PRINT "Something else"
END SELECT
