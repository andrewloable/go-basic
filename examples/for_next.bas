' FOR/NEXT Statement - Various Forms
' Pages: 225-226 (Chapter 5 - FOR/NEXT statement)

' Count up
FOR I = 1 TO 10
  PRINT I;
NEXT I
PRINT

' Count down
FOR I = 10 TO 1 STEP -1
  PRINT I;
NEXT I
PRINT

' Step by 2
FOR I = 0 TO 20 STEP 2
  PRINT I;
NEXT I
PRINT

' Floating-point step
FOR X = 0 TO 3.14159 STEP .1
  PRINT USING "##.#  ##.####";X, SIN(X)
NEXT X
