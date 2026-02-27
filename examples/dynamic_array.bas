' $DYNAMIC Metastatement - Dynamic Array Demo
' Page: 135 (Chapter 5 - $DYNAMIC metastatement)
$DYNAMIC
ON ERROR GOTO ErrorHandler
PRINT FRE(-1)
DIM BigArray(10000)
BigArray(6666) = 66
PRINT FRE(-1)
ERASE BigArray
PRINT FRE(-1)
PRINT BigArray(6666)
END

ErrorHandler:
  PRINT "An error of type " ERR;
  PRINT " has occurred at address" ERADR
END
