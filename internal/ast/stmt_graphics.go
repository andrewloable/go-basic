package ast

// ===========================================================================
// Graphics, sound, display, and miscellaneous statement nodes
// ===========================================================================

// ScreenStatement represents SCREEN.
type ScreenStatement struct {
	BasePos     Position
	Mode        Expression
	ColorSwitch Expression
	ActivePage  Expression
	VisualPage  Expression
}

func (n *ScreenStatement) statementNode()      {}
func (n *ScreenStatement) TokenLiteral() string { return "SCREEN" }
func (n *ScreenStatement) Pos() Position        { return n.BasePos }

// ColorStatement represents COLOR.
type ColorStatement struct {
	BasePos    Position
	Foreground Expression
	Background Expression
	Border     Expression
}

func (n *ColorStatement) statementNode()      {}
func (n *ColorStatement) TokenLiteral() string { return "COLOR" }
func (n *ColorStatement) Pos() Position        { return n.BasePos }

// PsetStatement represents PSET or PRESET.
type PsetStatement struct {
	BasePos  Position
	X        Expression
	Y        Expression
	Color    Expression
	IsStep   bool
	IsPreset bool
}

func (n *PsetStatement) statementNode()      {}
func (n *PsetStatement) TokenLiteral() string { return "PSET" }
func (n *PsetStatement) Pos() Position        { return n.BasePos }

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

func (n *LineStmt) statementNode()      {}
func (n *LineStmt) TokenLiteral() string { return "LINE" }
func (n *LineStmt) Pos() Position        { return n.BasePos }

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

func (n *CircleStmt) statementNode()      {}
func (n *CircleStmt) TokenLiteral() string { return "CIRCLE" }
func (n *CircleStmt) Pos() Position        { return n.BasePos }

// PaintStmt represents PAINT.
type PaintStmt struct {
	BasePos     Position
	X           Expression
	Y           Expression
	FillColor   Expression
	BorderColor Expression
}

func (n *PaintStmt) statementNode()      {}
func (n *PaintStmt) TokenLiteral() string { return "PAINT" }
func (n *PaintStmt) Pos() Position        { return n.BasePos }

// DrawStmt represents DRAW.
type DrawStmt struct {
	BasePos       Position
	CommandString Expression
}

func (n *DrawStmt) statementNode()      {}
func (n *DrawStmt) TokenLiteral() string { return "DRAW" }
func (n *DrawStmt) Pos() Position        { return n.BasePos }

// BeepStatement represents BEEP.
type BeepStatement struct {
	BasePos Position
}

func (n *BeepStatement) statementNode()      {}
func (n *BeepStatement) TokenLiteral() string { return "BEEP" }
func (n *BeepStatement) Pos() Position        { return n.BasePos }

// SoundStatement represents SOUND.
type SoundStatement struct {
	BasePos   Position
	Frequency Expression
	Duration  Expression
}

func (n *SoundStatement) statementNode()      {}
func (n *SoundStatement) TokenLiteral() string { return "SOUND" }
func (n *SoundStatement) Pos() Position        { return n.BasePos }

// PlayStatement represents PLAY.
type PlayStatement struct {
	BasePos       Position
	CommandString Expression
}

func (n *PlayStatement) statementNode()      {}
func (n *PlayStatement) TokenLiteral() string { return "PLAY" }
func (n *PlayStatement) Pos() Position        { return n.BasePos }

// LocateStatement represents LOCATE.
type LocateStatement struct {
	BasePos Position
	Row     Expression
	Col     Expression
}

func (n *LocateStatement) statementNode()      {}
func (n *LocateStatement) TokenLiteral() string { return "LOCATE" }
func (n *LocateStatement) Pos() Position        { return n.BasePos }

// ClsStatement represents CLS.
type ClsStatement struct {
	BasePos Position
	Mode    Expression // nil for default
}

func (n *ClsStatement) statementNode()      {}
func (n *ClsStatement) TokenLiteral() string { return "CLS" }
func (n *ClsStatement) Pos() Position        { return n.BasePos }

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

func (n *ViewStatement) statementNode()      {}
func (n *ViewStatement) TokenLiteral() string { return "VIEW" }
func (n *ViewStatement) Pos() Position        { return n.BasePos }

// SwapStatement represents SWAP.
type SwapStatement struct {
	BasePos Position
	Var1    Expression
	Var2    Expression
}

func (n *SwapStatement) statementNode()      {}
func (n *SwapStatement) TokenLiteral() string { return "SWAP" }
func (n *SwapStatement) Pos() Position        { return n.BasePos }

// IncrStatement represents INCR.
type IncrStatement struct {
	BasePos  Position
	Variable Expression
	Amount   Expression // nil means 1
}

func (n *IncrStatement) statementNode()      {}
func (n *IncrStatement) TokenLiteral() string { return "INCR" }
func (n *IncrStatement) Pos() Position        { return n.BasePos }

// DecrStatement represents DECR.
type DecrStatement struct {
	BasePos  Position
	Variable Expression
	Amount   Expression // nil means 1
}

func (n *DecrStatement) statementNode()      {}
func (n *DecrStatement) TokenLiteral() string { return "DECR" }
func (n *DecrStatement) Pos() Position        { return n.BasePos }

// RemStatement represents a REM (or ') comment.
type RemStatement struct {
	BasePos Position
	Text    string
}

func (n *RemStatement) statementNode()      {}
func (n *RemStatement) TokenLiteral() string { return "REM" }
func (n *RemStatement) Pos() Position        { return n.BasePos }

// LabelStatement represents a label (e.g. myLabel:).
type LabelStatement struct {
	BasePos Position
	Name    string
}

func (n *LabelStatement) statementNode()      {}
func (n *LabelStatement) TokenLiteral() string { return n.Name }
func (n *LabelStatement) Pos() Position        { return n.BasePos }

// LineNumberStatement represents a line number (e.g. 100).
type LineNumberStatement struct {
	BasePos Position
	Number  int
}

func (n *LineNumberStatement) statementNode()      {}
func (n *LineNumberStatement) TokenLiteral() string { return "LINE_NUMBER" }
func (n *LineNumberStatement) Pos() Position        { return n.BasePos }

// OnErrorGotoStatement represents ON ERROR GOTO.
type OnErrorGotoStatement struct {
	BasePos Position
	Target  string // label name, or "0" to disable
}

func (n *OnErrorGotoStatement) statementNode()      {}
func (n *OnErrorGotoStatement) TokenLiteral() string { return "ON ERROR GOTO" }
func (n *OnErrorGotoStatement) Pos() Position        { return n.BasePos }

// ResumeStatement represents RESUME, RESUME NEXT, or RESUME <label>.
type ResumeStatement struct {
	BasePos Position
	Type    string // "", "NEXT", or a label name
}

func (n *ResumeStatement) statementNode()      {}
func (n *ResumeStatement) TokenLiteral() string { return "RESUME" }
func (n *ResumeStatement) Pos() Position        { return n.BasePos }

// ErrorStatement represents ERROR <code> (trigger a runtime error).
type ErrorStatement struct {
	BasePos Position
	Code    Expression
}

func (n *ErrorStatement) statementNode()      {}
func (n *ErrorStatement) TokenLiteral() string { return "ERROR" }
func (n *ErrorStatement) Pos() Position        { return n.BasePos }

// OnEventGosubStatement represents ON <event> GOSUB (e.g. ON KEY GOSUB, ON TIMER GOSUB).
type OnEventGosubStatement struct {
	BasePos    Position
	EventType  string
	EventParam Expression
	Target     string
}

func (n *OnEventGosubStatement) statementNode()      {}
func (n *OnEventGosubStatement) TokenLiteral() string { return "ON" }
func (n *OnEventGosubStatement) Pos() Position        { return n.BasePos }

// PokeStatement represents POKE address, value (write byte to memory).
type PokeStatement struct {
	BasePos Position
	Address Expression
	Value   Expression
}

func (n *PokeStatement) statementNode()      {}
func (n *PokeStatement) TokenLiteral() string { return "POKE" }
func (n *PokeStatement) Pos() Position        { return n.BasePos }

// OnComputedGotoStatement represents ON expr GOTO t1, t2, t3 ...
type OnComputedGotoStatement struct {
	BasePos Position
	Expr    Expression
	Targets []string
}

func (n *OnComputedGotoStatement) statementNode()      {}
func (n *OnComputedGotoStatement) TokenLiteral() string { return "ON GOTO" }
func (n *OnComputedGotoStatement) Pos() Position        { return n.BasePos }

// OnComputedGosubStatement represents ON expr GOSUB t1, t2, t3 ...
type OnComputedGosubStatement struct {
	BasePos Position
	Expr    Expression
	Targets []string
}

func (n *OnComputedGosubStatement) statementNode()      {}
func (n *OnComputedGosubStatement) TokenLiteral() string { return "ON GOSUB" }
func (n *OnComputedGosubStatement) Pos() Position        { return n.BasePos }

// RandomizeStatement represents RANDOMIZE [seed]
type RandomizeStatement struct {
	BasePos Position
	Seed    Expression // nil means RANDOMIZE with no arg (use timer)
}

func (s *RandomizeStatement) statementNode()      {}
func (s *RandomizeStatement) TokenLiteral() string { return "RANDOMIZE" }
func (s *RandomizeStatement) Pos() Position        { return s.BasePos }
