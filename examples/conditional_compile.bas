' $IF/$ELSE/$ENDIF - Conditional Compilation
' Page: 137 (Chapter 5 - $IF/$ELSE/$ENDIF metastatements)
%Debug = -1
%ScreenType = 1
$IF %Debug
  PRINT "Debug mode"
$ENDIF
$IF %ScreenType = 1
  SCREEN 1
$ELSE
  SCREEN 2
$ENDIF
