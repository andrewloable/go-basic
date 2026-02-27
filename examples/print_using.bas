' PRINT USING - Formatted Output
' Pages: 318-322 (Chapter 5 - PRINT USING statement)
PRINT USING "##.## "; 3.14159
PRINT USING "###,###.## "; 12345.678
PRINT USING "$##,###.##"; 1234.56
PRINT USING "+###.## "; 123.45, -67.89
PRINT USING "##.####^^^^"; 12345678.9
PRINT USING "!"; "Hello"
PRINT USING "\ \"; "Hello World"
PRINT USING "&"; "Hello World"

' Formatted table
FOR I = 1 TO 10
  PRINT USING "Item ### costs $##,###.## and weighs ###.# lbs"; I, I*99.95, I*2.3
NEXT I
