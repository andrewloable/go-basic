REM Test SELECT CASE
FOR i = 1 TO 10
  SELECT CASE i
    CASE 1
      PRINT "one"
    CASE 2, 3
      PRINT "two or three"
    CASE 4 TO 6
      PRINT "four to six"
    CASE IS > 8
      PRINT "greater than eight"
    CASE ELSE
      PRINT "seven or eight"
  END SELECT
NEXT i
END
