' STR$ and VAL Functions - String/Number Conversion
' Page: 361-374 (Chapter 5 - STR$/VAL functions)
A = 123.456
A$ = STR$(A)
PRINT "String: '"; A$; "'"
PRINT "Length: "; LEN(A$)

B$ = "789.012"
B = VAL(B$)
PRINT "Number: "; B
PRINT "Doubled: "; B * 2
