' BIN$ Function - Binary Display
' Page: 150 (Chapter 5 - BIN$ function)
FOR I% = -5 TO 5
  PRINT USING "The binary equivalent of -## = &";I%,BIN$(I%)
NEXT I%
END
