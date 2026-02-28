package codegen

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// OPEN / CLOSE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitOpen(s *ast.OpenStatement) {
	g.needFileManager = true
	filename := g.emitExpr(s.Filename)
	fileNum := g.emitExpr(s.FileNum)
	mode := strings.ToUpper(s.Mode)
	modeConst := "rt.FileModeInput"
	switch mode {
	case "INPUT":
		modeConst = "rt.FileModeInput"
	case "OUTPUT":
		modeConst = "rt.FileModeOutput"
	case "APPEND":
		modeConst = "rt.FileModeAppend"
	case "RANDOM":
		modeConst = "rt.FileModeRandom"
	case "BINARY":
		modeConst = "rt.FileModeBinary"
	}
	recLen := "0"
	if s.RecLen != nil {
		recLen = g.emitExpr(s.RecLen)
	}
	g.writeLinef("fm.FileOpen(int(%s), string(%s), %s, int(%s))", fileNum, filename, modeConst, recLen)
}

func (g *CodeGenerator) emitClose(s *ast.CloseStatement) {
	g.needFileManager = true
	if len(s.FileNums) == 0 {
		g.writeLine("fm.FileCloseAll()")
	} else {
		for _, f := range s.FileNums {
			num := g.emitExpr(f)
			g.writeLinef("fm.FileClose(int(%s))", num)
		}
	}
}

// ---------------------------------------------------------------------------
// File I/O: PRINT#, INPUT#, WRITE#, FIELD, LSET, RSET, PUT, GET, SEEK
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitFilePrint(s *ast.FilePrintStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)

	// PUT$ filenum, data$ — binary file put
	if s.IsBinaryPut && len(s.Expressions) > 0 {
		data := g.emitExpr(s.Expressions[0])
		g.writeLinef("_ = fm.BinaryPutCur(int(%s), string(%s))", fileNum, data)
		return
	}

	if len(s.Expressions) == 0 {
		g.writeLinef("fm.FilePrint(int(%s))", fileNum)
		return
	}

	parts := make([]string, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		parts = append(parts, fmt.Sprintf("fmt.Sprint(%s)", g.emitExpr(expr)))
	}
	g.writeLinef("fm.FilePrint(int(%s), %s)", fileNum, strings.Join(parts, ", "))
}

func (g *CodeGenerator) emitFileInput(s *ast.FileInputStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)

	// GET$ filenum, length, var$ — binary file get
	if s.IsBinaryGet && len(s.Variables) >= 2 {
		length := g.emitExpr(s.Variables[0])
		varName := g.emitExpr(s.Variables[1])
		g.writeLinef("%s, _ = fm.BinaryGetCur(int(%s), int(%s))", varName, fileNum, length)
		return
	}

	for _, v := range s.Variables {
		varName := g.emitExpr(v)
		if s.IsLineInput {
			g.writeLine("{")
			g.indent++
			g.writeLinef("val_, err_ := fm.FileLineInput(int(%s))", fileNum)
			g.writeLine("_ = err_")
			// Determine if the variable is a string type
			if g.isStringExpr(v) {
				g.writeLinef("%s = val_", varName)
			} else {
				g.imports["strconv"] = true
				g.writeLinef("{ n_, err2_ := strconv.ParseFloat(val_, 64); _ = err2_; %s = %s(n_) }", varName, g.goTypeForExpr(v))
			}
			g.indent--
			g.writeLine("}")
		} else {
			g.writeLine("{")
			g.indent++
			g.writeLinef("val_, err_ := fm.FileInput(int(%s))", fileNum)
			g.writeLine("_ = err_")
			if g.isStringExpr(v) {
				g.writeLinef("%s = val_", varName)
			} else {
				g.imports["strconv"] = true
				g.imports["strings"] = true
				g.writeLinef("{ n_, err2_ := strconv.ParseFloat(strings.TrimSpace(val_), 64); _ = err2_; %s = %s(n_) }", varName, g.goTypeForExpr(v))
			}
			g.indent--
			g.writeLine("}")
		}
	}
}

func (g *CodeGenerator) emitFileWrite(s *ast.FileWriteStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)

	parts := make([]string, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		parts = append(parts, g.emitExpr(expr))
	}

	if len(parts) == 0 {
		g.writeLinef("fm.FileWrite(int(%s))", fileNum)
	} else {
		g.writeLinef("fm.FileWrite(int(%s), %s)", fileNum, strings.Join(parts, ", "))
	}
}

func (g *CodeGenerator) emitField(s *ast.FieldStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)

	g.writeLinef("fm.Field(int(%s), []rt.FieldDef{", fileNum)
	g.indent++
	for _, f := range s.Fields {
		length := g.emitExpr(f.Length)
		// Strip the $ suffix from the variable name for the field name
		fieldName := strings.TrimSuffix(f.VarName, "$")
		g.writeLinef("{Name: %q, Length: int(%s)},", fieldName, length)
		// Register the field definition so that after GET, codegen emits
		// fm.GetFieldValue() calls to populate the local variables.
		mangledName := mangleName(f.VarName)
		g.fieldDefs = append(g.fieldDefs, fieldDef{
			fileNum:   fileNum,
			fieldName: fieldName,
			varName:   mangledName,
		})
	}
	g.indent--
	g.writeLine("})")
}

func (g *CodeGenerator) emitLset(s *ast.LsetStatement) {
	g.needFileManager = true
	// LSET operates on a FIELD variable. We need to find which file number
	// the variable belongs to. For simplicity, we pass the variable name
	// and let the runtime search all open files.
	varName := strings.TrimSuffix(s.Variable, "$")
	value := g.emitExpr(s.Value)
	g.writeLinef("fm.Lset(0, %q, string(%s)) // LSET %s", varName, value, s.Variable)
}

func (g *CodeGenerator) emitRset(s *ast.RsetStatement) {
	g.needFileManager = true
	varName := strings.TrimSuffix(s.Variable, "$")
	value := g.emitExpr(s.Value)
	g.writeLinef("fm.Rset(0, %q, string(%s)) // RSET %s", varName, value, s.Variable)
}

func (g *CodeGenerator) emitPut(s *ast.PutStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)
	rec := "1"
	if s.RecordOrPos != nil {
		rec = g.emitExpr(s.RecordOrPos)
	}
	g.writeLinef("fm.RandomPut(int(%s), int(%s))", fileNum, rec)
}

func (g *CodeGenerator) emitGet(s *ast.GetStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)
	rec := "1"
	if s.RecordOrPos != nil {
		rec = g.emitExpr(s.RecordOrPos)
	}
	g.writeLinef("fm.RandomGet(int(%s), int(%s))", fileNum, rec)
	// After GET, populate local variables from the FIELD buffer.
	for _, fd := range g.fieldDefs {
		if fd.fileNum == fileNum {
			g.writeLinef("%s, _ = fm.GetFieldValue(int(%s), %q)", fd.varName, fd.fileNum, fd.fieldName)
		}
	}
}

func (g *CodeGenerator) emitSeek(s *ast.SeekStatement) {
	g.needFileManager = true
	fileNum := g.emitExpr(s.FileNum)
	pos := g.emitExpr(s.Position)
	g.writeLinef("fm.FileSeek(int(%s), int64(%s))", fileNum, pos)
}
