' READ/DATA/RESTORE - Student Scores
' Page: 328-329 (Chapter 5 - READ statement)
FOR I = 1 TO 5
  READ Name$, Score%
  PRINT Name$; TAB(20); Score%
NEXT I
RESTORE
READ FirstName$
PRINT "First name is: "; FirstName$
DATA "Alice", 95, "Bob", 87, "Carol", 92, "Dave", 78, "Eve", 88
