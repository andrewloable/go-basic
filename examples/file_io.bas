' OPEN Statement - File I/O Comprehensive Demo
' Pages: 297-300 (Chapter 5 - OPEN statement)
DEF FNPForKey$(Message$)
  PRINT Message$
  PRINT "Press a key to continue..."
  WHILE INKEY$ = ""
  WEND
  FNPForKey$ = ""
END DEF

SUB SequentialOutput
  KeyP$ = FNPForKey$("Creating a sequential output file")
  OPEN "OPEN.DTA" FOR OUTPUT AS #1
  FOR I% = 65 TO 90
    PRINT #1, CHR$(I%)
  NEXT I%
  CLOSE #1
END SUB

SUB SequentialAppend
  KeyP$ = FNPForKey$("Appending to the sequential file")
  OPEN "OPEN.DTA" FOR APPEND AS #1
  FOR I% = 97 TO 122
    PRINT #1, CHR$(I%)
  NEXT I%
  CLOSE #1
END SUB

SUB SequentialInput
  KeyP$ = FNPForKey$("Reading the sequential file")
  OPEN "OPEN.DTA" FOR INPUT AS #1
  WHILE NOT EOF(1)
    LINE INPUT #1, A$
    PRINT A$;
  WEND
  CLOSE #1
  PRINT
END SUB

CALL SequentialOutput
CALL SequentialAppend
CALL SequentialInput
END
