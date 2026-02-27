' FIELD/GET/PUT - Random File with Presidents
' Pages: 222-224 (Chapter 5 - FIELD statement)
OPEN "TEST.DTA" AS #1 LEN = 27
FIELD #1, 25 AS Rone$, 2 AS Rtwo$
FOR I% = 1 TO 5
  READ Rone$, Rtwo%
  LSET Rone$ = Rone$
  LSET Rtwo$ = MKI$(Rtwo%)
  PUT 1, I%
NEXT I%
FOR I% = 1 TO 5
  GET 1, I%
  PRINT Rone$, CVI(Rtwo$)
NEXT I%
CLOSE #1
END
DATA "George Washington", 57
DATA "John Adams", 61
DATA "Thomas Jefferson", 57
DATA "James Madison", 57
DATA "James Monroe", 59
