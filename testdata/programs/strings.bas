REM Test string functions
s$ = "Hello, World!"

PRINT "Original: "; s$
PRINT "Length:   "; LEN(s$)
PRINT "Upper:    "; UCASE$(s$)
PRINT "Lower:    "; LCASE$(s$)
PRINT "Left 5:   "; LEFT$(s$, 5)
PRINT "Right 6:  "; RIGHT$(s$, 6)
PRINT "Mid 8,5:  "; MID$(s$, 8, 5)
PRINT "Trim:     "; LTRIM$("  padded  ")
PRINT "INSTR:    "; INSTR(s$, "World")
PRINT "CHR$(65): "; CHR$(65)
PRINT "ASC(A):   "; ASC("A")
PRINT "STR$(42): "; STR$(42)
PRINT "VAL:      "; VAL("3.14")
PRINT "HEX$(255):"; HEX$(255)
PRINT "SPACE$(5): ["; SPACE$(5); "]"
PRINT "STRING$(3,42): "; STRING$(3, 42)

END
