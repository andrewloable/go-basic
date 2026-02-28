DIM Name1$ AS STRING
DIM Name2$ AS STRING
DIM NumGames AS INTEGER

Name1$ = "Alice"
Name2$ = "Bob"
NumGames = 5

CALL ShowInfo

' Test pass-by-reference: SUB modifies the caller's variable
DIM result AS SINGLE
result = 0
CALL Add(3, 4, result)
PRINT "3 + 4 ="; result

SUB ShowInfo
  SHARED Name1$, Name2$, NumGames
  PRINT Name1$
  PRINT Name2$
  PRINT NumGames
END SUB

SUB Add(a, b, res)
  res = a + b
END SUB
