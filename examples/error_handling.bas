' ON ERROR GOTO - Error Handling
' Page: 286 (Chapter 5 - ON ERROR statement)
ON ERROR GOTO ErrorHandler
OPEN "NONEXIST.FIL" FOR INPUT AS #1
PRINT "File opened successfully"
CLOSE #1
END

ErrorHandler:
  PRINT "Error"; ERR; "occurred"
  IF ERR = 53 THEN PRINT "File not found!"
  RESUME NEXT
