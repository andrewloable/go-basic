' RANDOMIZE and RND - Random Numbers
' Page: 325-326 (Chapter 5 - RANDOMIZE/RND)
RANDOMIZE TIMER
FOR I = 1 TO 10
  PRINT INT(RND * 100) + 1;
NEXT I
PRINT

' Generate random numbers in a range
Low = 10 : High = 50
FOR I = 1 TO 10
  PRINT INT(RND * (High - Low + 1)) + Low;
NEXT I
