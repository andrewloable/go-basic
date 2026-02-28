DIM Name1$ AS STRING
DIM Name2$ AS STRING
DIM NumGames AS INTEGER

Name1$ = "Alice"
Name2$ = "Bob"
NumGames = 5

CALL ShowInfo

SUB ShowInfo
  SHARED Name1$, Name2$, NumGames
  PRINT Name1$
  PRINT Name2$
  PRINT NumGames
END SUB
