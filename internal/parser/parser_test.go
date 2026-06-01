package parser

import (
	"orion/internal/ast"
	"orion/internal/lexer"
	"testing"
)

func TestArrayDeclFromExpr(t *testing.T) {
	src := `integer num = 12
[integer] arr = num.toArray()`
	l := lexer.New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	p := New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Stmts) != 2 {
		t.Fatalf("expected 2 stmts, got %d", len(prog.Stmts))
	}
}

func TestIndexAssignment(t *testing.T) {
	src := `[integer] arr = [1, 2]
arr[0] = 3`
	l := lexer.New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	p := New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Stmts) != 2 {
		t.Fatalf("expected 2 stmts, got %d", len(prog.Stmts))
	}
	_, ok := prog.Stmts[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", prog.Stmts[1])
	}
}
