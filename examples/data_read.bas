' DATA/READ/RESTORE Demo
' Page: 188-189 (Chapter 5 - DATA statement)
FOR I% = 1 TO 3
  READ A$, B%
  PRINT A$, B%
NEXT I%
RESTORE
READ X$
PRINT "After RESTORE, first item is: "; X$
END
DATA "Alpha", 1, "Beta", 2, "Gamma", 3
