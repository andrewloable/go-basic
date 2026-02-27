' ON/GOSUB - Computed GOSUB
' Page: 287 (Chapter 5 - ON/GOSUB statement)
INPUT "Enter choice (1-3): ", choice
ON choice GOSUB Sub1, Sub2, Sub3
END

Sub1:
  PRINT "You chose option 1"
  RETURN
Sub2:
  PRINT "You chose option 2"
  RETURN
Sub3:
  PRINT "You chose option 3"
  RETURN
