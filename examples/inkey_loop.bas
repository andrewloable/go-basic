' INKEY$ Function - Key Reading
' Page: 241 (Chapter 5 - INKEY$ function)
PRINT "Press keys (ESC to quit)..."
DO
  K$ = INKEY$
  IF LEN(K$) = 1 THEN
    PRINT "Key: "; K$; " ASCII: "; ASC(K$)
  ELSEIF LEN(K$) = 2 THEN
    PRINT "Extended key, scan code: "; ASC(MID$(K$,2,1))
  END IF
LOOP UNTIL K$ = CHR$(27)
