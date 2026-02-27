' Sequential File Output (Field-Delimited with WRITE#)
' Page: 116-117 (Chapter 4 - Field-Delimited Sequential Files)
OPEN "SEQUENTI.DTA" FOR OUTPUT AS #1
StringVariable$ = "This is a string of text"
Integer% = 1000
FloatingPoint! = 30000.1234
WRITE# 1, StringVariable$, Integer%, FloatingPoint!
StringVariable$ = "Another String"
Integer% = -32767
FloatingPoint! = 12345.54321
WRITE# 1, Integer%, StringVariable$, FloatingPoint!
CLOSE #1
END
