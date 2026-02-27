' Cylinder Volume Function
' Page: 104 (Chapter 4 - Pass by Value/Reference)
DEF FNCylVol(radius, height) STATIC
  FNCylVol = radius * radius * 3.14159 * height
END DEF
r = 4.7 : h = 12.1
vol = FNCylVol(r,h)
PRINT vol
