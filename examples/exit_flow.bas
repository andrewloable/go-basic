' EXIT Statement - Comprehensive Flow Control Demo
' Pages: 218-220 (Chapter 5 - EXIT statement)
DEF FN Loops(Selection%)
  SELECT CASE Selection%
  CASE 1
    PRINT "You executed selection 1"
    FOR Count% = 1 TO 10
      PRINT Count%
      IF Count% = 5 THEN EXIT FOR
    NEXT Count%
  CASE 2
    PRINT "You executed selection 2"
    Count% = 0
    DO
      INCR Count%
      PRINT Count%
      IF Count% = 5 THEN EXIT LOOP
    LOOP
  CASE 3
    PRINT "You executed selection 3"
    Count% = 0
    WHILE -1
      INCR Count%
      PRINT Count%
      IF Count% = 5 THEN EXIT LOOP
    WEND
  END SELECT
  FN Loops = Count%
END DEF

SUB Controls(Selection%, Dummy%)
  SELECT CASE Selection%
  CASE 1
    PRINT "You executed selection 1"
    IF Dummy% > 100 THEN
      PRINT "Greater than 100"
      EXIT SUB
    END IF
    PRINT "Less than or equal to 100"
  CASE 2
    PRINT "You executed selection 2"
    PRINT RND(Dummy%)
    EXIT LOOP
  END SELECT
  PRINT "You executed selection 3"
END SUB

PRINT FN Loops(1)
PRINT FN Loops(2)
PRINT FN Loops(3)
INPUT "Enter a number: ";Dummy%
FOR Count% = 1 TO 2
  CALL Controls(Count%, Dummy%)
NEXT Count%
END
