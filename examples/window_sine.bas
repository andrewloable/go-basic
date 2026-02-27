' WINDOW Statement - Sine Wave with Cartesian Coordinates
' Page: 382 (Chapter 5 - WINDOW statement)
SCREEN 1
WINDOW (-1, 1)-(1, -1)
LINE (-1, 0)-(1, 0)
LINE (0, -1)-(0, 1)
CIRCLE (0, 0), .5
FOR X = -1 TO 1 STEP .01
  PSET (X, SIN(X * 3.14159))
NEXT X
