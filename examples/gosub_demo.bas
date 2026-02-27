' GOSUB/RETURN Subroutine Demo
' Page: 233 (Chapter 5 - GOSUB statement)
PRINT "Before GOSUB"
GOSUB MySub
PRINT "After GOSUB"
END

MySub:
  PRINT "Inside the subroutine"
RETURN
