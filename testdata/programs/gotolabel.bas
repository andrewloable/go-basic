REM Test GOTO and labels
PRINT "Start"
GOTO skip
PRINT "This should not print"
skip:
PRINT "After skip"

' Test GOSUB/RETURN
GOSUB greet
PRINT "Back from GOSUB"
GOTO done

greet:
PRINT "Hello from GOSUB!"
RETURN

done:
PRINT "Done"
END
