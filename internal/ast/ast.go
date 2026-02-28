// Package ast defines the Abstract Syntax Tree (AST) node types produced by
// the parser (Phase 2) and consumed by all later phases.
//
// # What is an Abstract Syntax Tree?
//
// An AST is a tree representation of the grammatical structure of source code.
// Unlike a Parse Tree (which mirrors every token and grammar rule), the AST is
// "abstract" — it omits syntactic noise (parentheses, commas, keywords that
// only provide structure) and keeps only the semantically meaningful pieces.
//
//   BASIC source:   FOR i = 1 TO 10 STEP 2
//   AST node:       ForStatement{ Var: "i", Start: 1, End: 10, Step: 2, Body: [...] }
//
// # Node hierarchy
//
// Every AST node implements the Node interface. Nodes fall into two categories:
//
//   Statement nodes  (statementNode marker method) — represent actions:
//     LET, PRINT, FOR…NEXT, IF…THEN, GOSUB, SUB declarations, …
//
//   Expression nodes (expressionNode marker method) — represent values:
//     NumberLiteral, StringLiteral, Identifier, BinaryExpr, FunctionCall, …
//
// Programs are sequences of statements; statements contain expressions.
//
// # Visitor pattern
//
// Later compiler phases (semantic analysis, code generation, VM compilation)
// walk the AST using Go type-switches. Each phase matches on concrete node
// types and implements the transformation it needs. This "visitor without
// visitors" pattern is idiomatic Go and avoids the boilerplate of explicit
// Visitor interfaces.
package ast

// ---------------------------------------------------------------------------
// Position
// ---------------------------------------------------------------------------

// Position represents a location in the source code.
type Position struct {
	Line   int
	Column int
}

// ---------------------------------------------------------------------------
// NumType constants
// ---------------------------------------------------------------------------

const (
	NumInt = iota
	NumLong
	NumSingle
	NumDouble
)

// ---------------------------------------------------------------------------
// Core interfaces
// ---------------------------------------------------------------------------

// Node is the interface implemented by every AST node.
type Node interface {
	TokenLiteral() string
	Pos() Position
}

// Expression is a node that produces a value.
type Expression interface {
	Node
	expressionNode() // marker
}

// Statement is a node that performs an action.
type Statement interface {
	Node
	statementNode() // marker
}

// ===========================================================================
// Expression nodes
// ===========================================================================

// NumberLiteral represents a numeric constant (e.g. 42, 3.14, 1&, 2.5#).
type NumberLiteral struct {
	BasePos      Position
	Value        float64
	OriginalText string
	NumType      int
}

func (n *NumberLiteral) expressionNode()      {}
func (n *NumberLiteral) TokenLiteral() string  { return n.OriginalText }
func (n *NumberLiteral) Pos() Position         { return n.BasePos }

// StringLiteral represents a quoted string constant.
type StringLiteral struct {
	BasePos Position
	Value   string
}

func (n *StringLiteral) expressionNode()      {}
func (n *StringLiteral) TokenLiteral() string  { return n.Value }
func (n *StringLiteral) Pos() Position         { return n.BasePos }

// Identifier represents a variable or named reference (e.g. x, count%, name$).
type Identifier struct {
	BasePos    Position
	Name       string
	TypeSuffix string // "%", "&", "!", "#", "$", or ""
}

func (n *Identifier) expressionNode()      {}
func (n *Identifier) TokenLiteral() string  { return n.Name }
func (n *Identifier) Pos() Position         { return n.BasePos }

// BinaryExpr represents an infix expression (e.g. a + b, x AND y).
type BinaryExpr struct {
	BasePos  Position
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryExpr) expressionNode()      {}
func (n *BinaryExpr) TokenLiteral() string  { return n.Operator }
func (n *BinaryExpr) Pos() Position         { return n.BasePos }

// UnaryExpr represents a prefix expression (e.g. -x, NOT flag).
type UnaryExpr struct {
	BasePos  Position
	Operator string
	Operand  Expression
}

func (n *UnaryExpr) expressionNode()      {}
func (n *UnaryExpr) TokenLiteral() string  { return n.Operator }
func (n *UnaryExpr) Pos() Position         { return n.BasePos }

// FunctionCall represents a function invocation (e.g. ABS(x), MID$(s,1,3)).
type FunctionCall struct {
	BasePos Position
	Name    string
	Args    []Expression
}

func (n *FunctionCall) expressionNode()      {}
func (n *FunctionCall) TokenLiteral() string  { return n.Name }
func (n *FunctionCall) Pos() Position         { return n.BasePos }

// ArrayAccess represents an array element reference (e.g. a%(1), grid(r,c)).
type ArrayAccess struct {
	BasePos    Position
	Name       string
	TypeSuffix string
	Indices    []Expression
}

func (n *ArrayAccess) expressionNode()      {}
func (n *ArrayAccess) TokenLiteral() string  { return n.Name }
func (n *ArrayAccess) Pos() Position         { return n.BasePos }

// GroupExpr represents a parenthesized expression (e.g. (a + b)).
type GroupExpr struct {
	BasePos Position
	Inner   Expression
}

func (n *GroupExpr) expressionNode()      {}
func (n *GroupExpr) TokenLiteral() string  { return "(" }
func (n *GroupExpr) Pos() Position         { return n.BasePos }

// ===========================================================================
// Core statement nodes
// ===========================================================================

// Program is the root AST node representing an entire source file.
type Program struct {
	BasePos    Position
	Statements []Statement
}

func (n *Program) statementNode()       {}
func (n *Program) TokenLiteral() string  { return "PROGRAM" }
func (n *Program) Pos() Position         { return n.BasePos }

// PrintStatement represents a PRINT (or ?) statement.
type PrintStatement struct {
	BasePos        Position
	Expressions    []Expression
	Separators     []string // each is ",", ";", or ""
	HasTrailingSep bool
	Format         Expression // non-nil when PRINT USING
}

func (n *PrintStatement) statementNode()       {}
func (n *PrintStatement) TokenLiteral() string  { return "PRINT" }
func (n *PrintStatement) Pos() Position         { return n.BasePos }

// LetStatement represents a LET (or implicit) assignment.
type LetStatement struct {
	BasePos Position
	Name    *Identifier
	Value   Expression
}

func (n *LetStatement) statementNode()       {}
func (n *LetStatement) TokenLiteral() string  { return "LET" }
func (n *LetStatement) Pos() Position         { return n.BasePos }

// ArrayAssignment represents assignment to an array element.
type ArrayAssignment struct {
	BasePos Position
	Array   *ArrayAccess
	Value   Expression
}

func (n *ArrayAssignment) statementNode()       {}
func (n *ArrayAssignment) TokenLiteral() string  { return "LET" }
func (n *ArrayAssignment) Pos() Position         { return n.BasePos }

// IfStatement represents IF / THEN / ELSEIF / ELSE / END IF.
type IfStatement struct {
	BasePos       Position
	Condition     Expression
	ThenBlock     []Statement
	ElseIfClauses []ElseIfClause
	ElseBlock     []Statement
	IsSingleLine  bool
}

func (n *IfStatement) statementNode()       {}
func (n *IfStatement) TokenLiteral() string  { return "IF" }
func (n *IfStatement) Pos() Position         { return n.BasePos }

// ElseIfClause represents a single ELSEIF branch within an IfStatement.
type ElseIfClause struct {
	BasePos   Position
	Condition Expression
	Body      []Statement
}

// ForStatement represents FOR / TO / STEP / NEXT.
type ForStatement struct {
	BasePos Position
	Counter *Identifier
	Start   Expression
	End     Expression
	Step    Expression // nil when no STEP specified
	Body    []Statement
}

func (n *ForStatement) statementNode()       {}
func (n *ForStatement) TokenLiteral() string  { return "FOR" }
func (n *ForStatement) Pos() Position         { return n.BasePos }

// WhileStatement represents WHILE / WEND.
type WhileStatement struct {
	BasePos   Position
	Condition Expression
	Body      []Statement
}

func (n *WhileStatement) statementNode()       {}
func (n *WhileStatement) TokenLiteral() string  { return "WHILE" }
func (n *WhileStatement) Pos() Position         { return n.BasePos }

// DoLoopStatement represents DO / LOOP with optional WHILE or UNTIL.
type DoLoopStatement struct {
	BasePos   Position
	Condition Expression
	Body      []Statement
	TestAtTop bool
	IsUntil   bool
}

func (n *DoLoopStatement) statementNode()       {}
func (n *DoLoopStatement) TokenLiteral() string  { return "DO" }
func (n *DoLoopStatement) Pos() Position         { return n.BasePos }

// SelectCaseStatement represents SELECT CASE.
type SelectCaseStatement struct {
	BasePos   Position
	TestExpr  Expression
	Cases     []CaseClause
	ElseBlock []Statement
}

func (n *SelectCaseStatement) statementNode()       {}
func (n *SelectCaseStatement) TokenLiteral() string  { return "SELECT CASE" }
func (n *SelectCaseStatement) Pos() Position         { return n.BasePos }

// CaseClause represents a single CASE branch.
type CaseClause struct {
	BasePos Position
	Values  []CaseValue
	Body    []Statement
}

// CaseValue represents a single value, range, or comparison in a CASE clause.
type CaseValue struct {
	BasePos      Position
	Value        Expression
	EndValue     Expression // for TO ranges
	Comparison   string     // for IS > (the operator)
	IsRange      bool
	IsComparison bool
}

// GotoStatement represents GOTO.
type GotoStatement struct {
	BasePos Position
	Target  string
}

func (n *GotoStatement) statementNode()       {}
func (n *GotoStatement) TokenLiteral() string  { return "GOTO" }
func (n *GotoStatement) Pos() Position         { return n.BasePos }

// GosubStatement represents GOSUB.
type GosubStatement struct {
	BasePos Position
	Target  string
}

func (n *GosubStatement) statementNode()       {}
func (n *GosubStatement) TokenLiteral() string  { return "GOSUB" }
func (n *GosubStatement) Pos() Position         { return n.BasePos }

// ReturnStatement represents RETURN.
type ReturnStatement struct {
	BasePos Position
}

func (n *ReturnStatement) statementNode()       {}
func (n *ReturnStatement) TokenLiteral() string  { return "RETURN" }
func (n *ReturnStatement) Pos() Position         { return n.BasePos }

// ExitStatement represents EXIT FOR, EXIT DO, EXIT WHILE, etc.
type ExitStatement struct {
	BasePos  Position
	ExitType string // "FOR", "DO", "WHILE", "SUB", "FUNCTION", "DEF"
}

func (n *ExitStatement) statementNode()       {}
func (n *ExitStatement) TokenLiteral() string  { return "EXIT" }
func (n *ExitStatement) Pos() Position         { return n.BasePos }

// EndStatement represents END (terminate program).
type EndStatement struct {
	BasePos Position
}

func (n *EndStatement) statementNode()       {}
func (n *EndStatement) TokenLiteral() string  { return "END" }
func (n *EndStatement) Pos() Position         { return n.BasePos }

// StopStatement represents STOP.
type StopStatement struct {
	BasePos Position
}

func (n *StopStatement) statementNode()       {}
func (n *StopStatement) TokenLiteral() string  { return "STOP" }
func (n *StopStatement) Pos() Position         { return n.BasePos }

// SystemStatement represents SYSTEM.
type SystemStatement struct {
	BasePos Position
}

func (n *SystemStatement) statementNode()       {}
func (n *SystemStatement) TokenLiteral() string  { return "SYSTEM" }
func (n *SystemStatement) Pos() Position         { return n.BasePos }

// DataStatement represents DATA.
type DataStatement struct {
	BasePos Position
	Values  []Expression
}

func (n *DataStatement) statementNode()       {}
func (n *DataStatement) TokenLiteral() string  { return "DATA" }
func (n *DataStatement) Pos() Position         { return n.BasePos }

// ReadStatement represents READ (from DATA pool) or INPUT (from stdin).
// IsInput=true means this came from an INPUT/LINE INPUT statement and should
// read from stdin rather than the DATA pool.
type ReadStatement struct {
	BasePos    Position
	Variables  []Expression
	IsInput    bool   // true for INPUT/LINE INPUT, false for DATA READ
	Prompt     string // prompt string for INPUT (if any), already decoded
	IsLineInput bool  // true for LINE INPUT (read whole line)
}

func (n *ReadStatement) statementNode()       {}
func (n *ReadStatement) TokenLiteral() string  { return "READ" }
func (n *ReadStatement) Pos() Position         { return n.BasePos }

// RestoreStatement represents RESTORE.
type RestoreStatement struct {
	BasePos Position
	Target  string // label or empty
}

func (n *RestoreStatement) statementNode()       {}
func (n *RestoreStatement) TokenLiteral() string  { return "RESTORE" }
func (n *RestoreStatement) Pos() Position         { return n.BasePos }

// SwapStatement represents SWAP.
type SwapStatement struct {
	BasePos Position
	Var1    Expression
	Var2    Expression
}

func (n *SwapStatement) statementNode()       {}
func (n *SwapStatement) TokenLiteral() string  { return "SWAP" }
func (n *SwapStatement) Pos() Position         { return n.BasePos }

// IncrStatement represents INCR.
type IncrStatement struct {
	BasePos  Position
	Variable Expression
	Amount   Expression // nil means 1
}

func (n *IncrStatement) statementNode()       {}
func (n *IncrStatement) TokenLiteral() string  { return "INCR" }
func (n *IncrStatement) Pos() Position         { return n.BasePos }

// DecrStatement represents DECR.
type DecrStatement struct {
	BasePos  Position
	Variable Expression
	Amount   Expression // nil means 1
}

func (n *DecrStatement) statementNode()       {}
func (n *DecrStatement) TokenLiteral() string  { return "DECR" }
func (n *DecrStatement) Pos() Position         { return n.BasePos }

// RemStatement represents a REM (or ') comment.
type RemStatement struct {
	BasePos Position
	Text    string
}

func (n *RemStatement) statementNode()       {}
func (n *RemStatement) TokenLiteral() string  { return "REM" }
func (n *RemStatement) Pos() Position         { return n.BasePos }

// LabelStatement represents a label (e.g. myLabel:).
type LabelStatement struct {
	BasePos Position
	Name    string
}

func (n *LabelStatement) statementNode()       {}
func (n *LabelStatement) TokenLiteral() string  { return n.Name }
func (n *LabelStatement) Pos() Position         { return n.BasePos }

// LineNumberStatement represents a line number (e.g. 100).
type LineNumberStatement struct {
	BasePos Position
	Number  int
}

func (n *LineNumberStatement) statementNode()       {}
func (n *LineNumberStatement) TokenLiteral() string  { return "LINE_NUMBER" }
func (n *LineNumberStatement) Pos() Position         { return n.BasePos }

// ===========================================================================
// Declaration nodes
// ===========================================================================

// DimStatement represents DIM.
type DimStatement struct {
	BasePos      Position
	Declarations []DimDecl
}

func (n *DimStatement) statementNode()       {}
func (n *DimStatement) TokenLiteral() string  { return "DIM" }
func (n *DimStatement) Pos() Position         { return n.BasePos }

// DimDecl describes a single variable/array declaration inside DIM.
type DimDecl struct {
	BasePos      Position
	Name         string
	TypeSuffix   string
	Dimensions   []DimRange
	ElementType  string
	StringLength Expression
}

// DimRange describes the lower..upper bounds of one array dimension.
type DimRange struct {
	BasePos Position
	Lower   Expression
	Upper   Expression
}

// RedimStatement represents REDIM.
type RedimStatement struct {
	BasePos      Position
	Declarations []DimDecl
}

func (n *RedimStatement) statementNode()       {}
func (n *RedimStatement) TokenLiteral() string  { return "REDIM" }
func (n *RedimStatement) Pos() Position         { return n.BasePos }

// OptionBaseStatement represents OPTION BASE 0 or OPTION BASE 1.
type OptionBaseStatement struct {
	BasePos Position
	Value   int // 0 or 1
}

func (n *OptionBaseStatement) statementNode()       {}
func (n *OptionBaseStatement) TokenLiteral() string  { return "OPTION BASE" }
func (n *OptionBaseStatement) Pos() Position         { return n.BasePos }

// EraseStatement represents ERASE.
type EraseStatement struct {
	BasePos Position
	Names   []string
}

func (n *EraseStatement) statementNode()       {}
func (n *EraseStatement) TokenLiteral() string  { return "ERASE" }
func (n *EraseStatement) Pos() Position         { return n.BasePos }

// SubDeclaration represents SUB / END SUB.
type SubDeclaration struct {
	BasePos   Position
	Name      string
	Params    []Parameter
	Body      []Statement
	IsForward bool
}

func (n *SubDeclaration) statementNode()       {}
func (n *SubDeclaration) TokenLiteral() string  { return "SUB" }
func (n *SubDeclaration) Pos() Position         { return n.BasePos }

// FunctionDeclaration represents FUNCTION / END FUNCTION.
type FunctionDeclaration struct {
	BasePos    Position
	Name       string
	Params     []Parameter
	ReturnType string
	Body       []Statement
	IsForward  bool
}

func (n *FunctionDeclaration) statementNode()       {}
func (n *FunctionDeclaration) TokenLiteral() string  { return "FUNCTION" }
func (n *FunctionDeclaration) Pos() Position         { return n.BasePos }

// DefFnDeclaration represents DEF FN (single-line or multi-line).
type DefFnDeclaration struct {
	BasePos        Position
	Name           string
	Params         []Parameter
	Body           []Statement
	SingleLineExpr Expression
}

func (n *DefFnDeclaration) statementNode()       {}
func (n *DefFnDeclaration) TokenLiteral() string  { return "DEF FN" }
func (n *DefFnDeclaration) Pos() Position         { return n.BasePos }

// Parameter describes a formal parameter of a SUB, FUNCTION, or DEF FN.
type Parameter struct {
	BasePos Position
	Name    string
	Type    string
	IsByVal bool
}

// DefTypeStatement represents DEFINT, DEFLNG, DEFSNG, DEFDBL, DEFSTR.
type DefTypeStatement struct {
	BasePos      Position
	LetterRanges []LetterRange
	Type         string // "DEFINT", "DEFLNG", "DEFSNG", "DEFDBL", "DEFSTR"
}

func (n *DefTypeStatement) statementNode()       {}
func (n *DefTypeStatement) TokenLiteral() string  { return n.Type }
func (n *DefTypeStatement) Pos() Position         { return n.BasePos }

// LetterRange represents a range of letters (e.g. A-Z) in a DEFTYPE statement.
type LetterRange struct {
	Start byte
	End   byte
}

// ScopeStatement represents SHARED, LOCAL, STATIC, or COMMON.
type ScopeStatement struct {
	BasePos   Position
	Modifier  string // "SHARED", "LOCAL", "STATIC", "COMMON"
	Variables []string
}

func (n *ScopeStatement) statementNode()       {}
func (n *ScopeStatement) TokenLiteral() string  { return n.Modifier }
func (n *ScopeStatement) Pos() Position         { return n.BasePos }

// ===========================================================================
// File I/O nodes
// ===========================================================================

// OpenStatement represents OPEN.
type OpenStatement struct {
	BasePos  Position
	Filename Expression
	Mode     string
	FileNum  Expression
	RecLen   Expression
}

func (n *OpenStatement) statementNode()       {}
func (n *OpenStatement) TokenLiteral() string  { return "OPEN" }
func (n *OpenStatement) Pos() Position         { return n.BasePos }

// CloseStatement represents CLOSE.
type CloseStatement struct {
	BasePos  Position
	FileNums []Expression // nil = close all
}

func (n *CloseStatement) statementNode()       {}
func (n *CloseStatement) TokenLiteral() string  { return "CLOSE" }
func (n *CloseStatement) Pos() Position         { return n.BasePos }

// FileInputStatement represents INPUT# or LINE INPUT#.
type FileInputStatement struct {
	BasePos     Position
	FileNum     Expression
	Variables   []Expression
	IsLineInput bool
}

func (n *FileInputStatement) statementNode()       {}
func (n *FileInputStatement) TokenLiteral() string  { return "INPUT#" }
func (n *FileInputStatement) Pos() Position         { return n.BasePos }

// FilePrintStatement represents PRINT#.
type FilePrintStatement struct {
	BasePos        Position
	FileNum        Expression
	Expressions    []Expression
	Separators     []string
	Format         Expression
	HasTrailingSep bool
}

func (n *FilePrintStatement) statementNode()       {}
func (n *FilePrintStatement) TokenLiteral() string  { return "PRINT#" }
func (n *FilePrintStatement) Pos() Position         { return n.BasePos }

// FileWriteStatement represents WRITE#.
type FileWriteStatement struct {
	BasePos     Position
	FileNum     Expression
	Expressions []Expression
}

func (n *FileWriteStatement) statementNode()       {}
func (n *FileWriteStatement) TokenLiteral() string  { return "WRITE#" }
func (n *FileWriteStatement) Pos() Position         { return n.BasePos }

// GetStatement represents GET (binary/random file read).
type GetStatement struct {
	BasePos     Position
	FileNum     Expression
	RecordOrPos Expression
	Variable    Expression
}

func (n *GetStatement) statementNode()       {}
func (n *GetStatement) TokenLiteral() string  { return "GET" }
func (n *GetStatement) Pos() Position         { return n.BasePos }

// PutStatement represents PUT (binary/random file write).
type PutStatement struct {
	BasePos     Position
	FileNum     Expression
	RecordOrPos Expression
	Variable    Expression
}

func (n *PutStatement) statementNode()       {}
func (n *PutStatement) TokenLiteral() string  { return "PUT" }
func (n *PutStatement) Pos() Position         { return n.BasePos }

// SeekStatement represents SEEK.
type SeekStatement struct {
	BasePos  Position
	FileNum  Expression
	Position Expression
}

func (n *SeekStatement) statementNode()       {}
func (n *SeekStatement) TokenLiteral() string  { return "SEEK" }
func (n *SeekStatement) Pos() Position         { return n.BasePos }

// FieldStatement represents FIELD.
type FieldStatement struct {
	BasePos Position
	FileNum Expression
	Fields  []FieldDef
}

func (n *FieldStatement) statementNode()       {}
func (n *FieldStatement) TokenLiteral() string  { return "FIELD" }
func (n *FieldStatement) Pos() Position         { return n.BasePos }

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

func (n *LsetStatement) statementNode()       {}
func (n *LsetStatement) TokenLiteral() string  { return "LSET" }
func (n *LsetStatement) Pos() Position         { return n.BasePos }

// RsetStatement represents RSET.
type RsetStatement struct {
	BasePos  Position
	Variable string
	Value    Expression
}

func (n *RsetStatement) statementNode()       {}
func (n *RsetStatement) TokenLiteral() string  { return "RSET" }
func (n *RsetStatement) Pos() Position         { return n.BasePos }

// KillStatement represents KILL (delete a file).
type KillStatement struct {
	BasePos  Position
	Filename Expression
}

func (n *KillStatement) statementNode()       {}
func (n *KillStatement) TokenLiteral() string  { return "KILL" }
func (n *KillStatement) Pos() Position         { return n.BasePos }

// NameStatement represents NAME ... AS ... (rename a file).
type NameStatement struct {
	BasePos Position
	OldName Expression
	NewName Expression
}

func (n *NameStatement) statementNode()       {}
func (n *NameStatement) TokenLiteral() string  { return "NAME" }
func (n *NameStatement) Pos() Position         { return n.BasePos }

// ChdirStatement represents CHDIR.
type ChdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *ChdirStatement) statementNode()       {}
func (n *ChdirStatement) TokenLiteral() string  { return "CHDIR" }
func (n *ChdirStatement) Pos() Position         { return n.BasePos }

// MkdirStatement represents MKDIR.
type MkdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *MkdirStatement) statementNode()       {}
func (n *MkdirStatement) TokenLiteral() string  { return "MKDIR" }
func (n *MkdirStatement) Pos() Position         { return n.BasePos }

// RmdirStatement represents RMDIR.
type RmdirStatement struct {
	BasePos Position
	Path    Expression
}

func (n *RmdirStatement) statementNode()       {}
func (n *RmdirStatement) TokenLiteral() string  { return "RMDIR" }
func (n *RmdirStatement) Pos() Position         { return n.BasePos }

// ===========================================================================
// Graphics / Sound nodes
// ===========================================================================

// ScreenStatement represents SCREEN.
type ScreenStatement struct {
	BasePos     Position
	Mode        Expression
	ColorSwitch Expression
	ActivePage  Expression
	VisualPage  Expression
}

func (n *ScreenStatement) statementNode()       {}
func (n *ScreenStatement) TokenLiteral() string  { return "SCREEN" }
func (n *ScreenStatement) Pos() Position         { return n.BasePos }

// ColorStatement represents COLOR.
type ColorStatement struct {
	BasePos    Position
	Foreground Expression
	Background Expression
	Border     Expression
}

func (n *ColorStatement) statementNode()       {}
func (n *ColorStatement) TokenLiteral() string  { return "COLOR" }
func (n *ColorStatement) Pos() Position         { return n.BasePos }

// PsetStatement represents PSET or PRESET.
type PsetStatement struct {
	BasePos  Position
	X        Expression
	Y        Expression
	Color    Expression
	IsStep   bool
	IsPreset bool
}

func (n *PsetStatement) statementNode()       {}
func (n *PsetStatement) TokenLiteral() string  { return "PSET" }
func (n *PsetStatement) Pos() Position         { return n.BasePos }

// LineStmt represents the LINE graphics command.
type LineStmt struct {
	BasePos Position
	X1      Expression
	Y1      Expression
	X2      Expression
	Y2      Expression
	Color   Expression
	BoxMode string // "", "B", "BF"
	IsStep  bool
}

func (n *LineStmt) statementNode()       {}
func (n *LineStmt) TokenLiteral() string  { return "LINE" }
func (n *LineStmt) Pos() Position         { return n.BasePos }

// CircleStmt represents the CIRCLE graphics command.
type CircleStmt struct {
	BasePos Position
	X       Expression
	Y       Expression
	Radius  Expression
	Color   Expression
	Start   Expression
	End     Expression
	Aspect  Expression
}

func (n *CircleStmt) statementNode()       {}
func (n *CircleStmt) TokenLiteral() string  { return "CIRCLE" }
func (n *CircleStmt) Pos() Position         { return n.BasePos }

// PaintStmt represents PAINT.
type PaintStmt struct {
	BasePos     Position
	X           Expression
	Y           Expression
	FillColor   Expression
	BorderColor Expression
}

func (n *PaintStmt) statementNode()       {}
func (n *PaintStmt) TokenLiteral() string  { return "PAINT" }
func (n *PaintStmt) Pos() Position         { return n.BasePos }

// DrawStmt represents DRAW.
type DrawStmt struct {
	BasePos       Position
	CommandString Expression
}

func (n *DrawStmt) statementNode()       {}
func (n *DrawStmt) TokenLiteral() string  { return "DRAW" }
func (n *DrawStmt) Pos() Position         { return n.BasePos }

// BeepStatement represents BEEP.
type BeepStatement struct {
	BasePos Position
}

func (n *BeepStatement) statementNode()       {}
func (n *BeepStatement) TokenLiteral() string  { return "BEEP" }
func (n *BeepStatement) Pos() Position         { return n.BasePos }

// SoundStatement represents SOUND.
type SoundStatement struct {
	BasePos   Position
	Frequency Expression
	Duration  Expression
}

func (n *SoundStatement) statementNode()       {}
func (n *SoundStatement) TokenLiteral() string  { return "SOUND" }
func (n *SoundStatement) Pos() Position         { return n.BasePos }

// PlayStatement represents PLAY.
type PlayStatement struct {
	BasePos       Position
	CommandString Expression
}

func (n *PlayStatement) statementNode()       {}
func (n *PlayStatement) TokenLiteral() string  { return "PLAY" }
func (n *PlayStatement) Pos() Position         { return n.BasePos }

// LocateStatement represents LOCATE.
type LocateStatement struct {
	BasePos Position
	Row     Expression
	Col     Expression
}

func (n *LocateStatement) statementNode()       {}
func (n *LocateStatement) TokenLiteral() string  { return "LOCATE" }
func (n *LocateStatement) Pos() Position         { return n.BasePos }

// ClsStatement represents CLS.
type ClsStatement struct {
	BasePos Position
	Mode    Expression // nil for default
}

func (n *ClsStatement) statementNode()       {}
func (n *ClsStatement) TokenLiteral() string  { return "CLS" }
func (n *ClsStatement) Pos() Position         { return n.BasePos }

// ViewStatement represents VIEW or VIEW PRINT.
type ViewStatement struct {
	BasePos     Position
	X1          Expression
	Y1          Expression
	X2          Expression
	Y2          Expression
	FillColor   Expression
	BorderColor Expression
	IsPrint     bool
	Top         Expression
	Bottom      Expression
}

func (n *ViewStatement) statementNode()       {}
func (n *ViewStatement) TokenLiteral() string  { return "VIEW" }
func (n *ViewStatement) Pos() Position         { return n.BasePos }

// ===========================================================================
// Error handling and event nodes
// ===========================================================================

// OnErrorGotoStatement represents ON ERROR GOTO.
type OnErrorGotoStatement struct {
	BasePos Position
	Target  string // label name, or "0" to disable
}

func (n *OnErrorGotoStatement) statementNode()       {}
func (n *OnErrorGotoStatement) TokenLiteral() string  { return "ON ERROR GOTO" }
func (n *OnErrorGotoStatement) Pos() Position         { return n.BasePos }

// ResumeStatement represents RESUME, RESUME NEXT, or RESUME <label>.
type ResumeStatement struct {
	BasePos Position
	Type    string // "", "NEXT", or a label name
}

func (n *ResumeStatement) statementNode()       {}
func (n *ResumeStatement) TokenLiteral() string  { return "RESUME" }
func (n *ResumeStatement) Pos() Position         { return n.BasePos }

// ErrorStatement represents ERROR <code> (trigger a runtime error).
type ErrorStatement struct {
	BasePos Position
	Code    Expression
}

func (n *ErrorStatement) statementNode()       {}
func (n *ErrorStatement) TokenLiteral() string  { return "ERROR" }
func (n *ErrorStatement) Pos() Position         { return n.BasePos }

// OnEventGosubStatement represents ON <event> GOSUB (e.g. ON KEY GOSUB, ON TIMER GOSUB).
type OnEventGosubStatement struct {
	BasePos    Position
	EventType  string
	EventParam Expression
	Target     string
}

func (n *OnEventGosubStatement) statementNode()       {}
func (n *OnEventGosubStatement) TokenLiteral() string  { return "ON" }
func (n *OnEventGosubStatement) Pos() Position         { return n.BasePos }

// PokeStatement represents POKE address, value (write byte to memory).
type PokeStatement struct {
	BasePos  Position
	Address  Expression
	Value    Expression
}

func (n *PokeStatement) statementNode()       {}
func (n *PokeStatement) TokenLiteral() string  { return "POKE" }
func (n *PokeStatement) Pos() Position         { return n.BasePos }

// ConstStatement represents CONST name = expr (compile-time constant).
type ConstStatement struct {
	BasePos Position
	Name    string
	Value   Expression
}

func (n *ConstStatement) statementNode()       {}
func (n *ConstStatement) TokenLiteral() string  { return "CONST" }
func (n *ConstStatement) Pos() Position         { return n.BasePos }

// ClearStatement represents the CLEAR statement (resets all variables).
type ClearStatement struct {
	BasePos Position
}

func (n *ClearStatement) statementNode()       {}
func (n *ClearStatement) TokenLiteral() string  { return "CLEAR" }
func (n *ClearStatement) Pos() Position         { return n.BasePos }

// OnComputedGotoStatement represents ON expr GOTO t1, t2, t3 ...
type OnComputedGotoStatement struct {
	BasePos Position
	Expr    Expression
	Targets []string
}

func (n *OnComputedGotoStatement) statementNode()       {}
func (n *OnComputedGotoStatement) TokenLiteral() string  { return "ON GOTO" }
func (n *OnComputedGotoStatement) Pos() Position         { return n.BasePos }

// OnComputedGosubStatement represents ON expr GOSUB t1, t2, t3 ...
type OnComputedGosubStatement struct {
	BasePos Position
	Expr    Expression
	Targets []string
}

func (n *OnComputedGosubStatement) statementNode()       {}
func (n *OnComputedGosubStatement) TokenLiteral() string  { return "ON GOSUB" }
func (n *OnComputedGosubStatement) Pos() Position         { return n.BasePos }

// TypeBlockStatement represents a TYPE name ... END TYPE user-defined type.
type TypeBlockStatement struct {
	BasePos Position
	Name    string
	Fields  []TypeField
}

// TypeField represents a single field in a TYPE block.
type TypeField struct {
	Name     string
	TypeName string
}

func (n *TypeBlockStatement) statementNode()       {}
func (n *TypeBlockStatement) TokenLiteral() string  { return "TYPE" }
func (n *TypeBlockStatement) Pos() Position         { return n.BasePos }

// FnAssignStatement represents FN name = expr inside a DEF FN block (return value assignment).
type FnAssignStatement struct {
	BasePos Position
	Name    string
	Value   Expression
}

func (n *FnAssignStatement) statementNode()       {}
func (n *FnAssignStatement) TokenLiteral() string  { return "FN" }
func (n *FnAssignStatement) Pos() Position         { return n.BasePos }

// RandomizeStatement represents RANDOMIZE [seed]
type RandomizeStatement struct {
	BasePos Position
	Seed    Expression // nil means RANDOMIZE with no arg (use timer)
}

func (s *RandomizeStatement) statementNode()       {}
func (s *RandomizeStatement) TokenLiteral() string  { return "RANDOMIZE" }
func (s *RandomizeStatement) Pos() Position         { return s.BasePos }

// FnCallExpression represents FN name(args) in expression context.
type FnCallExpression struct {
	BasePos Position
	Name    string
	Args    []Expression
}

func (n *FnCallExpression) expressionNode()       {}
func (n *FnCallExpression) TokenLiteral() string  { return "FN" }
func (n *FnCallExpression) Pos() Position         { return n.BasePos }

// FieldAccessExpression represents struct/TYPE member access: expr.field
// In Turbo BASIC, the TYPE keyword defines record types whose fields are
// accessed with dot notation — e.g., Point.X, BCoor(i).XCoor.
type FieldAccessExpression struct {
	BasePos Position
	Object  Expression // the base expression (identifier or array access)
	Field   string     // the field name
}

func (n *FieldAccessExpression) expressionNode()       {}
func (n *FieldAccessExpression) TokenLiteral() string  { return "." }
func (n *FieldAccessExpression) Pos() Position         { return n.BasePos }

// FieldAssignStatement represents struct/TYPE member assignment: expr.field = value
type FieldAssignStatement struct {
	BasePos Position
	Object  Expression // the base expression
	Field   string     // the field name
	Value   Expression // the value to assign
}

func (n *FieldAssignStatement) statementNode()       {}
func (n *FieldAssignStatement) TokenLiteral() string  { return "." }
func (n *FieldAssignStatement) Pos() Position         { return n.BasePos }
