package ast

// ===========================================================================
// I/O statement nodes
// ===========================================================================

// PrintStatement represents a PRINT (or ?) statement.
type PrintStatement struct {
	BasePos        Position
	Expressions    []Expression
	Separators     []string // each is ",", ";", or ""
	HasTrailingSep bool
	Format         Expression // non-nil when PRINT USING
}

func (n *PrintStatement) statementNode()      {}
func (n *PrintStatement) TokenLiteral() string { return "PRINT" }
func (n *PrintStatement) Pos() Position        { return n.BasePos }

// DataStatement represents DATA.
type DataStatement struct {
	BasePos Position
	Values  []Expression
}

func (n *DataStatement) statementNode()      {}
func (n *DataStatement) TokenLiteral() string { return "DATA" }
func (n *DataStatement) Pos() Position        { return n.BasePos }

// ReadStatement represents READ (from DATA pool) or INPUT (from stdin).
// IsInput=true means this came from an INPUT/LINE INPUT statement and should
// read from stdin rather than the DATA pool.
type ReadStatement struct {
	BasePos     Position
	Variables   []Expression
	IsInput     bool   // true for INPUT/LINE INPUT, false for DATA READ
	Prompt      string // prompt string for INPUT (if any), already decoded
	IsLineInput bool   // true for LINE INPUT (read whole line)
}

func (n *ReadStatement) statementNode()      {}
func (n *ReadStatement) TokenLiteral() string { return "READ" }
func (n *ReadStatement) Pos() Position        { return n.BasePos }

// RestoreStatement represents RESTORE.
type RestoreStatement struct {
	BasePos Position
	Target  string // label or empty
}

func (n *RestoreStatement) statementNode()      {}
func (n *RestoreStatement) TokenLiteral() string { return "RESTORE" }
func (n *RestoreStatement) Pos() Position        { return n.BasePos }

// OpenStatement represents OPEN.
type OpenStatement struct {
	BasePos  Position
	Filename Expression
	Mode     string
	FileNum  Expression
	RecLen   Expression
}

func (n *OpenStatement) statementNode()      {}
func (n *OpenStatement) TokenLiteral() string { return "OPEN" }
func (n *OpenStatement) Pos() Position        { return n.BasePos }

// CloseStatement represents CLOSE.
type CloseStatement struct {
	BasePos  Position
	FileNums []Expression // nil = close all
}

func (n *CloseStatement) statementNode()      {}
func (n *CloseStatement) TokenLiteral() string { return "CLOSE" }
func (n *CloseStatement) Pos() Position        { return n.BasePos }

// FileInputStatement represents INPUT# or LINE INPUT#.
type FileInputStatement struct {
	BasePos     Position
	FileNum     Expression
	Variables   []Expression
	IsLineInput bool
}

func (n *FileInputStatement) statementNode()      {}
func (n *FileInputStatement) TokenLiteral() string { return "INPUT#" }
func (n *FileInputStatement) Pos() Position        { return n.BasePos }

// FilePrintStatement represents PRINT#.
type FilePrintStatement struct {
	BasePos        Position
	FileNum        Expression
	Expressions    []Expression
	Separators     []string
	Format         Expression
	HasTrailingSep bool
}

func (n *FilePrintStatement) statementNode()      {}
func (n *FilePrintStatement) TokenLiteral() string { return "PRINT#" }
func (n *FilePrintStatement) Pos() Position        { return n.BasePos }

// FileWriteStatement represents WRITE#.
type FileWriteStatement struct {
	BasePos     Position
	FileNum     Expression
	Expressions []Expression
}

func (n *FileWriteStatement) statementNode()      {}
func (n *FileWriteStatement) TokenLiteral() string { return "WRITE#" }
func (n *FileWriteStatement) Pos() Position        { return n.BasePos }

// GetStatement represents GET (binary/random file read).
type GetStatement struct {
	BasePos     Position
	FileNum     Expression
	RecordOrPos Expression
	Variable    Expression
}

func (n *GetStatement) statementNode()      {}
func (n *GetStatement) TokenLiteral() string { return "GET" }
func (n *GetStatement) Pos() Position        { return n.BasePos }

// PutStatement represents PUT (binary/random file write).
type PutStatement struct {
	BasePos     Position
	FileNum     Expression
	RecordOrPos Expression
	Variable    Expression
}

func (n *PutStatement) statementNode()      {}
func (n *PutStatement) TokenLiteral() string { return "PUT" }
func (n *PutStatement) Pos() Position        { return n.BasePos }

// SeekStatement represents SEEK.
type SeekStatement struct {
	BasePos  Position
	FileNum  Expression
	Position Expression
}

func (n *SeekStatement) statementNode()      {}
func (n *SeekStatement) TokenLiteral() string { return "SEEK" }
func (n *SeekStatement) Pos() Position        { return n.BasePos }

// FieldStatement represents FIELD.
type FieldStatement struct {
	BasePos Position
	FileNum Expression
	Fields  []FieldDef
}

func (n *FieldStatement) statementNode()      {}
func (n *FieldStatement) TokenLiteral() string { return "FIELD" }
func (n *FieldStatement) Pos() Position        { return n.BasePos }

// FieldDef describes one field within a FIELD statement.
type FieldDef struct {
	BasePos Position
	Length  Expression
	VarName string
}

// LsetStatement represents LSET.
type LsetStatement struct {
	BasePos  Position
	Variable string
	Value    Expression
}

func (n *LsetStatement) statementNode()      {}
func (n *LsetStatement) TokenLiteral() string { return "LSET" }
func (n *LsetStatement) Pos() Position        { return n.BasePos }

// RsetStatement represents RSET.
type RsetStatement struct {
	BasePos  Position
	Variable string
	Value    Expression
}

func (n *RsetStatement) statementNode()      {}
func (n *RsetStatement) TokenLiteral() string { return "RSET" }
func (n *RsetStatement) Pos() Position        { return n.BasePos }

// KillStatement represents KILL (delete a file).
type KillStatement struct {
	BasePos  Position
	Filename Expression
}

func (n *KillStatement) statementNode()      {}
func (n *KillStatement) TokenLiteral() string { return "KILL" }
func (n *KillStatement) Pos() Position        { return n.BasePos }

// NameStatement represents NAME ... AS ... (rename a file).
type NameStatement struct {
	BasePos Position
	OldName Expression
	NewName Expression
}

func (n *NameStatement) statementNode()      {}
func (n *NameStatement) TokenLiteral() string { return "NAME" }
func (n *NameStatement) Pos() Position        { return n.BasePos }

// ChdirStatement represents CHDIR.
type ChdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *ChdirStatement) statementNode()      {}
func (n *ChdirStatement) TokenLiteral() string { return "CHDIR" }
func (n *ChdirStatement) Pos() Position        { return n.BasePos }

// MkdirStatement represents MKDIR.
type MkdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *MkdirStatement) statementNode()      {}
func (n *MkdirStatement) TokenLiteral() string { return "MKDIR" }
func (n *MkdirStatement) Pos() Position        { return n.BasePos }

// RmdirStatement represents RMDIR.
type RmdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *RmdirStatement) statementNode()      {}
func (n *RmdirStatement) TokenLiteral() string { return "RMDIR" }
func (n *RmdirStatement) Pos() Position        { return n.BasePos }
