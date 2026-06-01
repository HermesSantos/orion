package parser

import (
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
