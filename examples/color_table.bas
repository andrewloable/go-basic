' COLOR Statement - Text Mode Attribute Table
' Page: 176 (Chapter 5 - COLOR statement, text mode)
FOR Back% = 0 TO 7
  FOR Fore% = 0 TO 15
    COLOR Fore%, Back%
    PRINT USING " ### ";Back%*16+Fore%;
  NEXT Fore%
  PRINT
NEXT Back%
PRINT
FOR Back% = 0 TO 7
  FOR Fore% = 16 TO 31
    COLOR Fore%, Back%
    PRINT USING " ### ";Back%*16+Fore%-16;
  NEXT Fore%
  PRINT
NEXT Back%
END
