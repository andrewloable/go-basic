' IF Block Statement - Grade Calculator
' Pages: 238-239 (Chapter 5 - IF block statement)
INPUT "Enter your score (0-100): ", score
IF score >= 90 THEN
  PRINT "Grade: A"
ELSEIF score >= 80 THEN
  PRINT "Grade: B"
ELSEIF score >= 70 THEN
  PRINT "Grade: C"
ELSEIF score >= 60 THEN
  PRINT "Grade: D"
ELSE
  PRINT "Grade: F"
END IF
