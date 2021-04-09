package ast

type Module struct {
	File     *File
	Modules  map[string]*File
	BasePath string
}

type Comment struct {
	MultiLine bool
	Str       string
	Pos       Position
}

func (i *Comment) Position() Position {
	return i.Pos
}

type ImportStmt struct {
	Pos     Position
	Alias   string
	Path    string
	AbsPath string
}

func (i *ImportStmt) Position() Position {
	return i.Pos
}

func (i *ImportStmt) stmtNode() {}

type File struct {
	Path       string
	Stms       []Stmt
	Global     []Stmt
	Comments   []*Comment
	Imports    []*ImportStmt
	Attributes []string
}

func (f *File) Import(alias string) *ImportStmt {
	for _, v := range f.Imports {
		if v.Alias == alias {
			return v
		}
	}
	return nil
}

type TryStmt struct {
	Pos        Position
	Body       *BlockStmt
	CatchIdent *VarDeclStmt
	Catch      *BlockStmt
	Finally    *BlockStmt
}

func (i *TryStmt) Position() Position {
	return i.Pos
}

func (i *TryStmt) stmtNode() {}

type SwitchStmt struct {
	Pos        Position
	Expression Expr
	Blocks     []*CaseBlock
	Default    *CaseBlock
	label      string
	continuePC int
	breakPC    int
}

func (i *SwitchStmt) SetLabel(s string) {
	i.label = s
}

func (i *SwitchStmt) SetContinuePC(pc int) {
	i.continuePC = pc
}

func (i *SwitchStmt) SetBreakPC(pc int) {
	i.breakPC = pc
}

func (i *SwitchStmt) Label() string {
	return i.label
}

func (i *SwitchStmt) BreakPC() int {
	return i.breakPC
}

func (i *SwitchStmt) Position() Position {
	return i.Pos
}

func (i *SwitchStmt) stmtNode() {}

type CaseBlock struct {
	Pos        Position
	Expression Expr
	Stmts      []Stmt
}

func (i *CaseBlock) Position() Position {
	return i.Pos
}

type IfStmt struct {
	Pos      Position
	IfBlocks []*IfBlock
	Else     *BlockStmt
}

type IfBlock struct {
	Condition Expr
	Body      *BlockStmt
}

func (i *IfStmt) Position() Position {
	return i.Pos
}

func (i *IfStmt) stmtNode() {}

type WhileStmt struct {
	Pos        Position
	Expression Expr
	Body       *BlockStmt

	label      string
	continuePC int
	breakPC    int
}

func (i *WhileStmt) SetLabel(s string) {
	i.label = s
}

func (i *WhileStmt) SetContinuePC(pc int) {
	i.continuePC = pc
}

func (i *WhileStmt) SetBreakPC(pc int) {
	i.breakPC = pc
}

func (i *WhileStmt) Label() string {
	return i.label
}

func (i *WhileStmt) ContinuePC() int {
	return i.continuePC
}

func (i *WhileStmt) BreakPC() int {
	return i.breakPC
}
func (i *WhileStmt) Position() Position {
	return i.Pos
}

func (i *WhileStmt) stmtNode() {}

type ForStmt struct {
	Pos          Position
	Declaration  []Stmt
	Expression   Expr
	Step         Stmt
	InExpression Expr
	OfExpression Expr
	Body         *BlockStmt

	label      string
	continuePC int
	breakPC    int
}

func (i *ForStmt) SetLabel(s string) {
	i.label = s
}

func (i *ForStmt) SetContinuePC(pc int) {
	i.continuePC = pc
}

func (i *ForStmt) SetBreakPC(pc int) {
	i.breakPC = pc
}

func (i *ForStmt) Label() string {
	return i.label
}

func (i *ForStmt) ContinuePC() int {
	return i.continuePC
}

func (i *ForStmt) BreakPC() int {
	return i.breakPC
}

func (i *ForStmt) Position() Position {
	return i.Pos
}

func (i *ForStmt) stmtNode() {}

type AsignStmt struct {
	Left  Expr
	Value Expr
}

func (i *AsignStmt) Position() Position {
	return i.Left.Position()
}

func (i *AsignStmt) stmtNode() {}

type IndexAsignStmt struct {
	Pos       Position
	Name      string
	IndexExpr Expr
	Value     Expr
}

func (i *IndexAsignStmt) Position() Position {
	return i.Pos
}

func (i *IndexAsignStmt) stmtNode() {}

type IncStmt struct {
	Operator Type
	Left     Expr
}

func (i *IncStmt) Position() Position {
	return i.Left.Position()
}

func (i *IncStmt) stmtNode() {}

type CallStmt struct {
	*CallExpr
}

func (i *CallStmt) stmtNode() {}

type TailCallStmt struct {
	*CallExpr
}

func (i *TailCallStmt) stmtNode() {}

// A BlockStmt node represents a braced statement list.
type BlockStmt struct {
	Lbrace Position // Position of "{"
	List   []Stmt
	Rbrace Position // Position of "}"
}

func (i *BlockStmt) Position() Position {
	return i.Lbrace
}
func (i *BlockStmt) stmtNode() {}

type ClassDeclStmt struct {
	Pos        Position
	Name       string
	Exported   bool
	Fields     []*VarDeclStmt
	Functions  []*FuncDeclStmt
	Attributes []string
}

func (c *ClassDeclStmt) Position() Position {
	return c.Pos
}
func (*ClassDeclStmt) declNode() {}
func (*ClassDeclStmt) stmtNode() {}

type EnumDeclStmt struct {
	Pos      Position
	Name     string
	Values   []EnumValue
	Exported bool
}

type EnumValue struct {
	Pos   Position
	Name  string
	Kind  Type
	Value *ConstantExpr
}

func (i *EnumDeclStmt) Position() Position {
	return i.Pos
}
func (i *EnumDeclStmt) declNode() {}
func (i *EnumDeclStmt) stmtNode() {}

type FuncDeclStmt struct {
	Pos        Position
	Args       *Arguments
	Variadic   bool
	Body       *BlockStmt
	Name       string
	Exported   bool
	Anonymous  bool
	Attributes []string
	Comment    *Comment

	// a Object value means that it is a method of that object
	ReceiverType string
}

func (i *FuncDeclStmt) Position() Position {
	return i.Pos
}
func (i *FuncDeclStmt) declNode() {}
func (i *FuncDeclStmt) stmtNode() {}

type VarDeclStmt struct {
	Pos      Position
	Name     string
	Value    Expr
	Exported bool
	Const    bool
}

func (i *VarDeclStmt) Position() Position {
	return i.Pos
}
func (i *VarDeclStmt) declNode() {}
func (i *VarDeclStmt) stmtNode() {}

type DeleteStmt struct {
	Object   string
	Property string
	Pos      Position
}

func (i *DeleteStmt) Position() Position {
	return i.Pos
}
func (i *DeleteStmt) stmtNode() {}

type BreakStmt struct {
	Pos   Position
	Label string
}

func (i *BreakStmt) Position() Position {
	return i.Pos
}

func (i *BreakStmt) stmtNode() {}

type ContinueStmt struct {
	Pos   Position
	Label string
}

func (i *ContinueStmt) Position() Position {
	return i.Pos
}

func (i *ContinueStmt) stmtNode() {}

type ReturnStmt struct {
	Pos   Position
	Value Expr
}

func (i *ReturnStmt) Position() Position {
	return i.Pos
}

func (i *ReturnStmt) stmtNode() {}

type ThrowStmt struct {
	Pos   Position
	Value Expr
}

func (i *ThrowStmt) Position() Position {
	return i.Pos
}

func (i *ThrowStmt) stmtNode() {}

// FuncDeclExpr is a function as a value expression
type FuncDeclExpr struct {
	Pos      Position
	Args     *Arguments
	Variadic bool
	Body     *BlockStmt
}

func (i *FuncDeclExpr) Position() Position {
	return i.Pos
}
func (i *FuncDeclExpr) exprNode() {}

// RegisterExpr is just the index of a local register.
// Is for internal use only, when you need to pass directly
// the result of a compiled expression of any kind.
type RegisterExpr struct {
	Pos Position
	Reg int
}

func (i *RegisterExpr) Position() Position {
	return i.Pos
}
func (i *RegisterExpr) exprNode() {}

type IdentExpr struct {
	Pos  Position
	Name string
}

func (i *IdentExpr) Position() Position {
	return i.Pos
}
func (i *IdentExpr) exprNode() {}

type ConstantExpr struct {
	Pos   Position
	Kind  Type
	Value string
}

func (i *ConstantExpr) Position() Position {
	return i.Pos
}
func (i *ConstantExpr) exprNode() {}

type UnaryExpr struct {
	Pos      Position
	Operator Type
	Operand  Expr
}

func (i *UnaryExpr) Position() Position {
	return i.Pos
}
func (i *UnaryExpr) exprNode() {}

type BinaryExpr struct {
	Operator Type
	Left     Expr
	Right    Expr
}

func (i *BinaryExpr) Position() Position {
	return i.Left.Position()
}
func (i *BinaryExpr) exprNode() {}

type TernaryExpr struct {
	Condition Expr
	Left      Expr
	Right     Expr
}

func (i *TernaryExpr) Position() Position {
	return i.Left.Position()
}
func (i *TernaryExpr) exprNode() {}

type NewInstanceExpr struct {
	Name   Expr
	Lparen Position
	Args   []Expr
	Rparen Position
	Spread bool // if the last argument has a spread operator
}

func (i *NewInstanceExpr) Position() Position {
	return i.Name.Position()
}
func (*NewInstanceExpr) exprNode() {}

type CallExpr struct {
	Ident    Expr
	Lparen   Position
	Args     []Expr
	Rparen   Position
	Spread   bool // if the last argument has a spread operator
	Optional bool // Optional chaining ?.
	First    bool // if it is the first part of the expression
}

func (i *CallExpr) Position() Position {
	return i.Ident.Position()
}
func (i *CallExpr) exprNode() {}

type TypeofExpr struct {
	Expr Expr
}

func (i *TypeofExpr) Position() Position {
	return i.Expr.Position()
}
func (i *TypeofExpr) exprNode() {}

type SelectorExpr struct {
	X        Expr       // expression
	Sel      *IdentExpr // field selector
	Optional bool       // Optional chaining ?.
	First    bool       // if it is the first part of the expression
}

func (i *SelectorExpr) Position() Position {
	return i.X.Position()
}
func (i *SelectorExpr) exprNode() {}

type IndexExpr struct {
	Left     Expr
	Lbrack   Position
	Index    Expr
	Rbrack   Position
	Optional bool // Optional chaining ?.
	First    bool // if it is the first part of the expression
}

func (i *IndexExpr) Position() Position {
	return i.Left.Position()
}
func (i *IndexExpr) exprNode() {}

type KeyValue struct {
	Key     string
	KeyType Type
	Value   Expr
}

type MapDeclExpr struct {
	Pos  Position
	List []KeyValue
}

func (i *MapDeclExpr) Position() Position {
	return i.Pos
}
func (i *MapDeclExpr) exprNode() {}

type ArrayDeclExpr struct {
	Pos  Position
	List []Expr
}

func (i *ArrayDeclExpr) Position() Position {
	return i.Pos
}
func (i *ArrayDeclExpr) exprNode() {}

type Node interface {
	Position() Position
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Decl interface {
	Node
	declNode()
}

type Target interface {
	Label() string
}

type ContinueTarget interface {
	Label() string
	ContinuePC() int
}

type BreakTarget interface {
	Label() string
	BreakPC() int
}

type Arguments struct {
	Opening Position
	List    []*Field
}

type Field struct {
	Pos      Position
	Name     string
	Optional bool
}
