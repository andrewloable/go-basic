' CHR$ Function - Character Display
' Page: 163 (Chapter 5 - CHR$ function)
FOR I% = 32 TO 255
  PRINT USING "###  "; I%;
  PRINT CHR$(I%); "  ";
NEXT I%
END
