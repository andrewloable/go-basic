' SEEK Statement - Binary File Position
' Page: 350 (Chapter 5 - SEEK statement)
OPEN "TEST.BIN" FOR BINARY AS #1
PUT$ 1, "Hello, World!"
SEEK 1, 1
GET$ 1, 13, A$
PRINT A$
SEEK 1, 8
GET$ 1, 6, A$
PRINT A$
CLOSE #1
