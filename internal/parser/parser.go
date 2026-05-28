package parser

import (
	"fmt"
	"orion/internal/ast"
	"orion/internal/lexer"
	"strings"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	// Filter out newlines for easier parsing
	var filtered []lexer.Token
	for _, t := range tokens {
		if t.Type != lexer.TOKEN_NEWLINE {
			filtered = append(filtered, t)
		}
	}
	return &Parser{tokens: filtered, pos: 0}
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAt(offset int) lexer.Token {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[idx]
}

func (p *Parser) advance() lexer.Token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	t := p.peek()
	if t.Type != tt {
		return t, fmt.Errorf("line %d: expected %s, got %s (%q)", t.Line, tt, t.Type, t.Literal)
	}
	return p.advance(), nil
}

func (p *Parser) skipSemis() {
	for p.peek().Type == lexer.TOKEN_SEMI {
		p.advance()
	}
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	for p.peek().Type != lexer.TOKEN_EOF {
		p.skipSemis()
		if p.peek().Type == lexer.TOKEN_EOF {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
		p.skipSemis()
	}
	return prog, nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	t := p.peek()

	// Array declaration: [type] name = [...]
	if t.Type == lexer.TOKEN_LBRACKET {
		return p.parseArrayDecl()
	}

	// Tuple declaration: (type, type) name = (val, val)
	if t.Type == lexer.TOKEN_LPAREN && p.isTupleTypeDecl() {
		return p.parseTupleDecl()
	}

	// if statement
	if t.Type == lexer.TOKEN_IF {
		return p.parseIfStmt()
	}

	// for statement
	if t.Type == lexer.TOKEN_FOR {
		return p.parseForStmt()
	}

	// io::write(...)
	if t.Type == lexer.TOKEN_IDENT && t.Literal == "io" && p.peekAt(1).Type == lexer.TOKEN_DCOLON {
		return p.parseIOCall()
	}

	// Implicit return: ::(...)
	if t.Type == lexer.TOKEN_DCOLON {
		return p.parseImplicitReturn()
	}

	// Function or variable declaration starting with a type keyword
	if isTypeToken(t) {
		return p.parseTypedDecl()
	}

	// void function declaration: name () { ... }  vs call: name(args)
	// Disambiguate by looking for { after the closing )
	if t.Type == lexer.TOKEN_IDENT && p.peekAt(1).Type == lexer.TOKEN_LPAREN {
		if p.isVoidFuncDecl() {
			return p.parseVoidFuncDecl()
		}
	}

	// Expression statement (func call, method call)
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.ExprStmt{Expr: expr}, nil
}

func (p *Parser) isTupleTypeDecl() bool {
	// Look ahead: (type, type) name =
	save := p.pos
	defer func() { p.pos = save }()
	p.advance() // consume (
	for p.peek().Type != lexer.TOKEN_RPAREN && p.peek().Type != lexer.TOKEN_EOF {
		p.advance()
	}
	if p.peek().Type != lexer.TOKEN_RPAREN {
		return false
	}
	p.advance() // consume )
	return p.peek().Type == lexer.TOKEN_IDENT
}

func isTypeToken(t lexer.Token) bool {
	return t.Type == lexer.TOKEN_TYPE_STRING ||
		t.Type == lexer.TOKEN_TYPE_INT ||
		t.Type == lexer.TOKEN_TYPE_BOOL ||
		t.Type == lexer.TOKEN_TYPE_FLOAT
}

// parseTypedDecl handles: type name = value  OR  type name(params) { body }
func (p *Parser) parseTypedDecl() (ast.Node, error) {
	typeTok := p.advance()
	typeName := typeTok.Literal

	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	name := nameTok.Literal

	// Function declaration: type name(params) { body }
	if p.peek().Type == lexer.TOKEN_LPAREN {
		return p.parseFuncDeclWith(typeName, name)
	}

	// Variable declaration: type name = value
	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.VarDecl{TypeName: typeName, Name: name, Value: val}, nil
}

func (p *Parser) parseFuncDeclWith(retType, name string) (ast.Node, error) {
	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FuncDecl{ReturnType: retType, Name: name, Params: params, Body: body}, nil
}

func (p *Parser) parseVoidFuncDecl() (ast.Node, error) {
	nameTok := p.advance()
	name := nameTok.Literal
	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FuncDecl{ReturnType: "", Name: name, Params: params, Body: body}, nil
}

func (p *Parser) parseParams() ([]ast.Param, error) {
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	var params []ast.Param
	for p.peek().Type != lexer.TOKEN_RPAREN && p.peek().Type != lexer.TOKEN_EOF {
		if !isTypeToken(p.peek()) {
			return nil, fmt.Errorf("line %d: expected type in param, got %q", p.peek().Line, p.peek().Literal)
		}
		typeTok := p.advance()
		nameTok, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		params = append(params, ast.Param{TypeName: typeTok.Literal, Name: nameTok.Literal})
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return params, nil
}

func (p *Parser) parseBlock() ([]ast.Node, error) {
	if _, err := p.expect(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}
	var stmts []ast.Node
	for p.peek().Type != lexer.TOKEN_RBRACE && p.peek().Type != lexer.TOKEN_EOF {
		p.skipSemis()
		if p.peek().Type == lexer.TOKEN_RBRACE {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipSemis()
	}
	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) parseIfStmt() (ast.Node, error) {
	p.advance() // consume 'if'
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	stmt := &ast.IfStmt{Condition: cond, Body: body}

	for p.peek().Type == lexer.TOKEN_OR_IF {
		p.advance() // consume 'or_if'
		eic, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		eib, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		stmt.ElseIfs = append(stmt.ElseIfs, ast.ElseIf{Condition: eic, Body: eib})
	}

	if p.peek().Type == lexer.TOKEN_OR {
		p.advance() // consume 'or'
		elsebody, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		stmt.Else = elsebody
	}

	return stmt, nil
}

func (p *Parser) parseForStmt() (ast.Node, error) {
	p.advance() // consume 'for'

	// for index, value :: target { body }
	indexTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_COMMA); err != nil {
		return nil, err
	}
	valueTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_DCOLON); err != nil {
		return nil, err
	}

	// Collection to iterate over
	target := ""
	if p.peek().Type == lexer.TOKEN_IDENT {
		target = p.advance().Literal
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.ForStmt{
		IndexVar: indexTok.Literal,
		ValueVar: valueTok.Literal,
		Target:   target,
		Body:     body,
	}, nil
}

func (p *Parser) parseArrayDecl() (ast.Node, error) {
	p.advance() // consume [
	elemTypeTok := p.advance()
	elemType := elemTypeTok.Literal
	if _, err := p.expect(lexer.TOKEN_RBRACKET); err != nil {
		return nil, err
	}
	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}
	// Parse array literal [val, val, ...]
	if _, err := p.expect(lexer.TOKEN_LBRACKET); err != nil {
		return nil, err
	}
	var elems []ast.Node
	for p.peek().Type != lexer.TOKEN_RBRACKET && p.peek().Type != lexer.TOKEN_EOF {
		elem, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elems = append(elems, elem)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TOKEN_RBRACKET); err != nil {
		return nil, err
	}
	return &ast.ArrayDecl{ElemType: elemType, Name: nameTok.Literal, Elements: elems}, nil
}

func (p *Parser) parseTupleDecl() (ast.Node, error) {
	p.advance() // consume (
	var types []string
	for p.peek().Type != lexer.TOKEN_RPAREN && p.peek().Type != lexer.TOKEN_EOF {
		types = append(types, p.advance().Literal)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}
	// Parse tuple literal (val, val)
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	var vals []ast.Node
	for p.peek().Type != lexer.TOKEN_RPAREN && p.peek().Type != lexer.TOKEN_EOF {
		v, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return &ast.TupleDecl{Types: types, Name: nameTok.Literal, Values: vals}, nil
}

func (p *Parser) parseIOCall() (ast.Node, error) {
	p.advance() // io
	p.advance() // ::
	p.advance() // write
	args, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}
	return &ast.IOWrite{Args: args}, nil
}

func (p *Parser) parseImplicitReturn() (ast.Node, error) {
	p.advance() // ::
	args, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("implicit return requires an argument")
	}
	return &ast.ImplicitReturn{Value: args[0]}, nil
}

func (p *Parser) parseCallArgs() ([]ast.Node, error) {
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	var args []ast.Node
	for p.peek().Type != lexer.TOKEN_RPAREN && p.peek().Type != lexer.TOKEN_EOF {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}

// parseExpr handles binary expressions and primary expressions
func (p *Parser) parseExpr() (ast.Node, error) {
	return p.parseBinary(0)
}

var precedences = map[string]int{
	"||": 1, "&&": 2,
	"==": 3, "!=": 3, "<": 4, ">": 4, "<=": 4, ">=": 4,
	"+": 5, "-": 5,
	"*": 6, "/": 6,
}

func (p *Parser) opPrec() int {
	t := p.peek()
	switch t.Type {
	case lexer.TOKEN_OR_OP:
		return 1
	case lexer.TOKEN_AND:
		return 2
	case lexer.TOKEN_EQ, lexer.TOKEN_NEQ:
		return 3
	case lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE:
		return 4
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS:
		return 5
	case lexer.TOKEN_STAR, lexer.TOKEN_SLASH:
		return 6
	}
	return 0
}

func (p *Parser) parseBinary(minPrec int) (ast.Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		prec := p.opPrec()
		if prec <= minPrec {
			break
		}
		op := p.advance().Literal
		right, err := p.parseBinary(prec)
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Node, error) {
	if p.peek().Type == lexer.TOKEN_NOT {
		op := p.advance().Literal
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, Operand: operand}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Node, error) {
	t := p.peek()

	switch t.Type {
	case lexer.TOKEN_STRING:
		p.advance()
		return &ast.StringLit{Value: t.Literal}, nil

	case lexer.TOKEN_INT:
		p.advance()
		return &ast.IntLit{Value: t.Literal}, nil

	case lexer.TOKEN_FLOAT:
		p.advance()
		return &ast.FloatLit{Value: t.Literal}, nil

	case lexer.TOKEN_BOOL:
		p.advance()
		return &ast.BoolLit{Value: t.Literal == "true"}, nil

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TOKEN_IDENT:
		// io::write
		if t.Literal == "io" && p.peekAt(1).Type == lexer.TOKEN_DCOLON {
			return p.parseIOCall()
		}

		name := p.advance().Literal

		// method call: name:method(args)
		if p.peek().Type == lexer.TOKEN_COLON {
			p.advance() // consume :
			methodTok, err := p.expect(lexer.TOKEN_IDENT)
			if err != nil {
				return nil, err
			}
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			return &ast.MethodCall{Object: name, Method: methodTok.Literal, Args: args}, nil
		}

		// function call: name(args)
		if p.peek().Type == lexer.TOKEN_LPAREN {
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			return &ast.FuncCall{Name: name, Args: args}, nil
		}

		return &ast.Identifier{Name: name}, nil

	default:
		return nil, fmt.Errorf("line %d: unexpected token %s (%q) in expression", t.Line, t.Type, t.Literal)
	}
}

// InterpolateString converts "$name" style interpolation into Go fmt.Sprintf format
func InterpolateString(s string) (string, []string, bool) {
	if !strings.Contains(s, "$") {
		return s, nil, false
	}
	var result strings.Builder
	var vars []string
	i := 0
	for i < len(s) {
		if s[i] == '$' {
			i++
			start := i
			for i < len(s) && (s[i] == '_' || (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= '0' && s[i] <= '9')) {
				i++
			}
			varName := s[start:i]
			if varName != "" {
				result.WriteString("%v")
				vars = append(vars, varName)
			} else {
				result.WriteByte('$')
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String(), vars, true
}

// isVoidFuncDecl looks ahead to see if this is: name ( params ) { body }
// rather than a regular function call name(args)
func (p *Parser) isVoidFuncDecl() bool {
	save := p.pos
	defer func() { p.pos = save }()
	p.advance() // consume name
	p.advance() // consume (
	// scan past params
	depth := 1
	for p.pos < len(p.tokens) && depth > 0 {
		tt := p.tokens[p.pos].Type
		if tt == lexer.TOKEN_LPAREN { depth++ }
		if tt == lexer.TOKEN_RPAREN { depth-- }
		p.pos++
	}
	// Next token after ) should be { for a function declaration
	return p.pos < len(p.tokens) && p.tokens[p.pos].Type == lexer.TOKEN_LBRACE
}
