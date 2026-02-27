' SUB/END SUB - Procedure Definition
' Page: 367-369 (Chapter 5 - SUB statement)
SUB Greet(Name$)
  PRINT "Hello, "; Name$; "!"
END SUB

SUB Add(a, b, result)
  result = a + b
END SUB

CALL Greet("World")
CALL Add(3, 4, sum)
PRINT "3 + 4 = "; sum
