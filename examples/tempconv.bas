' Temperature Conversion Functions
' Page: 98 (Chapter 4 - Functions)
DEF FNCtoF(degreesC) = (1.8 * degreesC) + 32
DEF FNFtoC(degreesF) = (degreesF - 32) * .555555
temp = 100
PRINT FNCtoF(temp)
INPUT "Enter today's high: ", th
temp = FNFtoC(th)
PRINT temp
