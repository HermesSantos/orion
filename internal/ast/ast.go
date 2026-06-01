package ast

// Node is the base interface for all AST nodes
type Node interface {
	nodeType() string
}

// --- Statements ---

type Program struct {
	Stmts []Node
}

func (p *Program) nodeType() string { return "Program" }

type VarDecl struct {
	TypeName string // "string", "integer", "bool", "float"
	Name     string
	Value    Node
}

func (v *VarDecl) nodeType() string { return "VarDecl" }

type ArrayDecl struct {
	ElemType string
	Name     string
	Elements []Node // literal [a, b, ...]
	Value    Node   // expression (e.g. x.toArray())
}

func (a *ArrayDecl) nodeType() string { return "ArrayDecl" }

type TupleDecl struct {
	Types []string
	Name  string
	Values []Node
}

func (t *TupleDecl) nodeType() string { return "TupleDecl" }

type FuncDecl struct {
	ReturnType string // "" means void
	Name       string
	Params     []Param
	Body       []Node
}

func (f *FuncDecl) nodeType() string { return "FuncDecl" }

type Param struct {
	TypeName string
	Name     string
}

type IfStmt struct {
	Condition Node   // nil for bare `or {}`
	Body      []Node
	ElseIfs   []ElseIf
	Else      []Node
}

type ElseIf struct {
	Condition Node
	Body      []Node
}

func (i *IfStmt) nodeType() string { return "IfStmt" }

type ForStmt struct {
	IndexVar string
	ValueVar string
	Target   string // array variable name
	Body     []Node
}

func (f *ForStmt) nodeType() string { return "ForStmt" }

type ReturnStmt struct {
	Value Node
}

func (r *ReturnStmt) nodeType() string { return "ReturnStmt" }

type ExprStmt struct {
	Expr Node
}

func (e *ExprStmt) nodeType() string { return "ExprStmt" }

// AssignStmt is assignment: target = value (variable or arr[i] = x)
type AssignStmt struct {
	Left  Node
	Right Node
}

func (a *AssignStmt) nodeType() string { return "AssignStmt" }

// --- Expressions ---

type Identifier struct {
	Name string
}

func (i *Identifier) nodeType() string { return "Identifier" }

type StringLit struct {
	Value string
}

func (s *StringLit) nodeType() string { return "StringLit" }

type IntLit struct {
	Value string
}

func (i *IntLit) nodeType() string { return "IntLit" }

type FloatLit struct {
	Value string
}

func (f *FloatLit) nodeType() string { return "FloatLit" }

type BoolLit struct {
	Value bool
}

func (b *BoolLit) nodeType() string { return "BoolLit" }

type ArrayLit struct {
	Elements []Node
}

func (a *ArrayLit) nodeType() string { return "ArrayLit" }

type TupleLit struct {
	Values []Node
}

func (t *TupleLit) nodeType() string { return "TupleLit" }

// io::write(...)
type IOWrite struct {
	Args []Node
}

func (io *IOWrite) nodeType() string { return "IOWrite" }

// functionCall(args...)
type FuncCall struct {
	Name string
	Args []Node
}

func (f *FuncCall) nodeType() string { return "FuncCall" }

// ::("hello, $name !")  — implicit return shorthand
type ImplicitReturn struct {
	Value Node
}

func (i *ImplicitReturn) nodeType() string { return "ImplicitReturn" }

// obj:method(args) or obj.method(args)
type MethodCall struct {
	Object Node
	Method string
	Args   []Node
}

func (m *MethodCall) nodeType() string { return "MethodCall" }

type BinaryExpr struct {
	Op    string
	Left  Node
	Right Node
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }

type UnaryExpr struct {
	Op      string
	Operand Node
}

func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }

// IndexExpr is array/slice access: arr[index]
type IndexExpr struct {
	Object Node
	Index  Node
}

func (i *IndexExpr) nodeType() string { return "IndexExpr" }
