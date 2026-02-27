' ON/GOTO - Computed GOTO
' Page: 288 (Chapter 5 - ON/GOTO statement)
INPUT "Enter choice (1-3): ", choice
ON choice GOTO Label1, Label2, Label3
PRINT "Invalid choice"
END

Label1:
  PRINT "Branch 1"
  END
Label2:
  PRINT "Branch 2"
  END
Label3:
  PRINT "Branch 3"
  END
