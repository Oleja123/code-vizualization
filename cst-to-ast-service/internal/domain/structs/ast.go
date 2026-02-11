package structs

import "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/interfaces"

// Location представляет позицию узла в исходном коде
type Location struct {
	Line      uint32 `json:"line"`
	Column    uint32 `json:"column"`
	EndLine   uint32 `json:"endLine"`
	EndColumn uint32 `json:"endColumn"`
}

// Type представляет тип данных в C (поддержка int, указателей и многомерных массивов)
type Type struct {
	BaseType     string `json:"baseType"`     // "int"
	PointerLevel int    `json:"pointerLevel"` // 0 = int, 1 = int*, 2 = int**
	ArraySizes   []int  `json:"arraySizes"`   // пустой если не массив, [10, 20] для int[10][20]
}

// Parameter представляет параметр функции
type Parameter struct {
	Type Type     `json:"type"`
	Name string   `json:"name"`
	Loc  Location `json:"location"`
}

// Program представляет корневой узел программы
type Program struct {
	Type         string            `json:"type"` // всегда "Program"
	Declarations []interfaces.Stmt `json:"declarations"`
	Loc          Location          `json:"location"`
}

func (p *Program) NodeType() string      { return "Program" }
func (p *Program) GetLocation() Location { return p.Loc }

// ============= Statements =============

// VariableDecl представляет объявление переменной
type VariableDecl struct {
	Type     string          `json:"type"`    // всегда "VariableDecl"
	VarType  Type            `json:"varType"` // тип переменной (int, int*, int[10])
	Name     string          `json:"name"`
	InitExpr interfaces.Expr `json:"initExpr,omitempty"` // может быть nil
	Loc      Location        `json:"location"`
}

func (v *VariableDecl) StmtNode()             {}
func (v *VariableDecl) NodeType() string      { return "VariableDecl" }
func (v *VariableDecl) GetLocation() Location { return v.Loc }

// FunctionDecl представляет объявление функции
type FunctionDecl struct {
	Type       string      `json:"type"` // всегда "FunctionDecl"
	Name       string      `json:"name"`
	ReturnType Type        `json:"returnType"`
	Parameters []Parameter `json:"parameters"`
	Body       *BlockStmt  `json:"body"`
	Loc        Location    `json:"location"`
}

func (f *FunctionDecl) StmtNode()             {}
func (f *FunctionDecl) NodeType() string      { return "FunctionDecl" }
func (f *FunctionDecl) GetLocation() Location { return f.Loc }

// ElseIfClause представляет else if блок в условном операторе
type ElseIfClause struct {
	Condition interfaces.Expr `json:"condition"`
	Block     interfaces.Stmt `json:"block"`
	Loc       Location        `json:"location"`
}

// IfStmt представляет условный оператор if/else if/else
type IfStmt struct {
	Type       string          `json:"type"` // всегда "IfStmt"
	Condition  interfaces.Expr `json:"condition"`
	ThenBlock  interfaces.Stmt `json:"thenBlock"`
	ElseIfList []ElseIfClause  `json:"elseIf,omitempty"`    // else if блоки
	ElseBlock  interfaces.Stmt `json:"elseBlock,omitempty"` // может быть nil
	Loc        Location        `json:"location"`
}

func (i *IfStmt) StmtNode()             {}
func (i *IfStmt) NodeType() string      { return "IfStmt" }
func (i *IfStmt) GetLocation() Location { return i.Loc }

// WhileStmt представляет цикл while
type WhileStmt struct {
	Type      string          `json:"type"` // всегда "WhileStmt"
	Condition interfaces.Expr `json:"condition"`
	Body      interfaces.Stmt `json:"body"`
	Loc       Location        `json:"location"`
}

func (w *WhileStmt) StmtNode()             {}
func (w *WhileStmt) NodeType() string      { return "WhileStmt" }
func (w *WhileStmt) GetLocation() Location { return w.Loc }

// ForStmt представляет цикл for
type ForStmt struct {
	Type      string          `json:"type"`                // всегда "ForStmt"
	Init      interfaces.Stmt `json:"init,omitempty"`      // может быть nil
	Condition interfaces.Expr `json:"condition,omitempty"` // может быть nil
	Post      interfaces.Stmt `json:"post,omitempty"`      // может быть nil (ExprStmt)
	Body      interfaces.Stmt `json:"body"`
	Loc       Location        `json:"location"`
}

func (f *ForStmt) StmtNode()             {}
func (f *ForStmt) NodeType() string      { return "ForStmt" }
func (f *ForStmt) GetLocation() Location { return f.Loc }

// ReturnStmt представляет оператор return
type ReturnStmt struct {
	Type  string          `json:"type"`            // всегда "ReturnStmt"
	Value interfaces.Expr `json:"value,omitempty"` // может быть nil
	Loc   Location        `json:"location"`
}

func (r *ReturnStmt) StmtNode()             {}
func (r *ReturnStmt) NodeType() string      { return "ReturnStmt" }
func (r *ReturnStmt) GetLocation() Location { return r.Loc }

// BlockStmt представляет блок операторов { ... }
type BlockStmt struct {
	Type       string            `json:"type"` // всегда "BlockStmt"
	Statements []interfaces.Stmt `json:"statements"`
	Loc        Location          `json:"location"`
}

func (b *BlockStmt) StmtNode()             {}
func (b *BlockStmt) NodeType() string      { return "BlockStmt" }
func (b *BlockStmt) GetLocation() Location { return b.Loc }

// ExprStmt представляет выражение как оператор
type ExprStmt struct {
	Type       string          `json:"type"` // всегда "ExprStmt"
	Expression interfaces.Expr `json:"expression"`
	Loc        Location        `json:"location"`
}

func (e *ExprStmt) StmtNode()             {}
func (e *ExprStmt) NodeType() string      { return "ExprStmt" }
func (e *ExprStmt) GetLocation() Location { return e.Loc }

// BreakStmt представляет оператор break
type BreakStmt struct {
	Type string   `json:"type"` // всегда "BreakStmt"
	Loc  Location `json:"location"`
}

func (b *BreakStmt) StmtNode()             {}
func (b *BreakStmt) NodeType() string      { return "BreakStmt" }
func (b *BreakStmt) GetLocation() Location { return b.Loc }

// ContinueStmt представляет оператор continue
type ContinueStmt struct {
	Type string   `json:"type"` // всегда "ContinueStmt"
	Loc  Location `json:"location"`
}

func (c *ContinueStmt) StmtNode()             {}
func (c *ContinueStmt) NodeType() string      { return "ContinueStmt" }
func (c *ContinueStmt) GetLocation() Location { return c.Loc }

// ============= Expressions =============

// Identifier представляет идентификатор (имя переменной/функции)
type Identifier struct {
	Type string   `json:"type"` // всегда "Identifier"
	Name string   `json:"name"`
	Loc  Location `json:"location"`
}

func (i *Identifier) ExprNode()             {}
func (i *Identifier) NodeType() string      { return "Identifier" }
func (i *Identifier) GetLocation() Location { return i.Loc }

// VariableExpr представляет обращение к переменной в выражении
type VariableExpr struct {
	Type string   `json:"type"` // всегда "VariableExpr"
	Name string   `json:"name"`
	Loc  Location `json:"location"`
}

func (v *VariableExpr) ExprNode()             {}
func (v *VariableExpr) NodeType() string      { return "VariableExpr" }
func (v *VariableExpr) GetLocation() Location { return v.Loc }

// IntLiteral представляет целочисленный литерал
type IntLiteral struct {
	Type  string   `json:"type"` // всегда "IntLiteral"
	Value int      `json:"value"`
	Loc   Location `json:"location"`
}

func (l *IntLiteral) ExprNode()             {}
func (l *IntLiteral) NodeType() string      { return "IntLiteral" }
func (l *IntLiteral) GetLocation() Location { return l.Loc }

// BinaryExpr представляет бинарное выражение
type BinaryExpr struct {
	Type     string          `json:"type"` // всегда "BinaryExpr"
	Left     interfaces.Expr `json:"left"`
	Operator string          `json:"operator"` // +, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||
	Right    interfaces.Expr `json:"right"`
	Loc      Location        `json:"location"`
}

func (b *BinaryExpr) ExprNode()             {}
func (b *BinaryExpr) NodeType() string      { return "BinaryExpr" }
func (b *BinaryExpr) GetLocation() Location { return b.Loc }

// UnaryExpr представляет унарное выражение
type UnaryExpr struct {
	Type      string          `json:"type"`     // всегда "UnaryExpr"
	Operator  string          `json:"operator"` // -, !, *, &, ++, --
	Operand   interfaces.Expr `json:"operand"`
	IsPostfix bool            `json:"isPostfix"` // true для постфиксных операторов (i++, i--)
	Loc       Location        `json:"location"`
}

func (u *UnaryExpr) ExprNode()             {}
func (u *UnaryExpr) NodeType() string      { return "UnaryExpr" }
func (u *UnaryExpr) GetLocation() Location { return u.Loc }

// AssignmentExpr представляет выражение присваивания
type AssignmentExpr struct {
	Type     string          `json:"type"` // всегда "AssignmentExpr"
	Left     interfaces.Expr `json:"left"` // Identifier или ArrayAccessExpr
	Operator string          `json:"operator"`
	Right    interfaces.Expr `json:"right"`
	Loc      Location        `json:"location"`
}

func (a *AssignmentExpr) ExprNode()             {}
func (a *AssignmentExpr) NodeType() string      { return "AssignmentExpr" }
func (a *AssignmentExpr) GetLocation() Location { return a.Loc }

// CallExpr представляет вызов функции
type CallExpr struct {
	Type         string            `json:"type"` // всегда "CallExpr"
	FunctionName string            `json:"functionName"`
	Arguments    []interfaces.Expr `json:"arguments"`
	Loc          Location          `json:"location"`
}

func (c *CallExpr) ExprNode()             {}
func (c *CallExpr) NodeType() string      { return "CallExpr" }
func (c *CallExpr) GetLocation() Location { return c.Loc }

// ArrayAccessExpr представляет доступ к элементу массива
type ArrayAccessExpr struct {
	Type  string          `json:"type"`  // всегда "ArrayAccessExpr"
	Array interfaces.Expr `json:"array"` // обычно Identifier
	Index interfaces.Expr `json:"index"`
	Loc   Location        `json:"location"`
}

func (a *ArrayAccessExpr) ExprNode()             {}
func (a *ArrayAccessExpr) NodeType() string      { return "ArrayAccessExpr" }
func (a *ArrayAccessExpr) GetLocation() Location { return a.Loc }

// ArrayInitExpr представляет инициализатор массива {1, 2, 3}
type ArrayInitExpr struct {
	Type     string            `json:"type"` // всегда "ArrayInitExpr"
	Elements []interfaces.Expr `json:"elements"`
	Loc      Location          `json:"location"`
}

func (a *ArrayInitExpr) ExprNode()             {}
func (a *ArrayInitExpr) NodeType() string      { return "ArrayInitExpr" }
func (a *ArrayInitExpr) GetLocation() Location { return a.Loc }
