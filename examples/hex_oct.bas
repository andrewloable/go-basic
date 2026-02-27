' OCT$ and HEX$ Functions
' Page: 281-235 (Chapter 5 - OCT$/HEX$ functions)
FOR I% = 0 TO 15
  PRINT USING "## = & = &"; I%, HEX$(I%), OCT$(I%)
NEXT I%
