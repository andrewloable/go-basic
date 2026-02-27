' Sequential File Input (Field-Delimited with INPUT#)
' Page: 117 (Chapter 4 - Sequential Files)
OPEN "SEQUENTI.DTA" FOR INPUT AS #1
StringVariable$ = ""
Integer% = 0
FloatingPoint! = 0
INPUT # 1, StringVariable$, Integer%, FloatingPoint!
PRINT StringVariable$, Integer%, FloatingPoint!
StringVariable$ = ""
Integer% = 0
FloatingPoint! = 0
INPUT # 1, Integer%, StringVariable$, FloatingPoint!
PRINT Integer%, StringVariable$, FloatingPoint!
CLOSE #1
END
