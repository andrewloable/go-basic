REM Test error handling
ON ERROR GOTO handler

PRINT "Before error"
ERROR 5   ' Illegal function call
PRINT "This should not print"
END

handler:
PRINT "Error caught!"
PRINT "Error code:"; ERR
RESUME NEXT
