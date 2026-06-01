package codegen

import (
	"fmt"
	"orion/internal/ast"
	"orion/internal/parser"
	"strings"
)

type Generator struct {
	sb        strings.Builder
	indent    int
	useFmt    bool
	useFmtStr bool // for fmt.Sprintf in string interpolation
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) ind() string {
	return strings.Repeat("\t", g.indent)
}

func (g *Generator) line(s string) {
	g.sb.WriteString(g.ind() + s + "\n")
}

func (g *Generator) mapType(orionType string) string {
	switch orionType {
	case "string":
		return "string"
	case "integer", "int":
		return "int"
	case "bool":
		return "bool"
	case "float":
		return "float64"
	default:
		return orionType
	}
}

func (g *Generator) Generate(prog *ast.Program) (string, error) {
	// First pass: check if fmt is needed
	for _, stmt := range prog.Stmts {
		if g.needsFmt(stmt) {
			g.useFmt = true
			break
		}
	}

	bodyGen := &Generator{indent: 1, useFmt: g.useFmt}
	for _, stmt := range prog.Stmts {
		err := bodyGen.genStmt(stmt)
		if err != nil {
			return "", err
		}
	}
	bodyStr := bodyGen.sb.String()
	g.useFmtStr = bodyGen.useFmtStr

	// Build final file
	g.sb.WriteString("package main\n\n")
	if g.useFmt || g.useFmtStr {
		g.sb.WriteString("import \"fmt\"\n\n")
	}
	g.sb.WriteString("func main() {\n")
	g.sb.WriteString(bodyStr)
	g.sb.WriteString("}\n")

	return g.sb.String(), nil
}

func (g *Generator) needsFmt(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IOWrite:
		return true
	case *ast.FuncDecl:
		for _, s := range n.Body {
			if g.needsFmt(s) {
				return true
			}
		}
	case *ast.IfStmt:
		for _, s := range n.Body {
			if g.needsFmt(s) {
				return true
			}
		}
		for _, ei := range n.ElseIfs {
			for _, s := range ei.Body {
				if g.needsFmt(s) {
					return true
				}
			}
		}
		for _, s := range n.Else {
			if g.needsFmt(s) {
				return true
			}
		}
	case *ast.ForStmt:
		for _, s := range n.Body {
			if g.needsFmt(s) {
				return true
			}
		}
	case *ast.ExprStmt:
		return g.needsFmt(n.Expr)
	}
	return false
}

func (g *Generator) genStmt(node ast.Node) error {
	switch n := node.(type) {
	case *ast.VarDecl:
		return g.genVarDecl(n)
	case *ast.ArrayDecl:
		return g.genArrayDecl(n)
	case *ast.TupleDecl:
		return g.genTupleDecl(n)
	case *ast.FuncDecl:
		return g.genFuncDecl(n)
	case *ast.IfStmt:
		return g.genIfStmt(n)
	case *ast.ForStmt:
		return g.genForStmt(n)
	case *ast.IOWrite:
		return g.genIOWrite(n)
	case *ast.ImplicitReturn:
		expr, err := g.genExpr(n.Value)
		if err != nil {
			return err
		}
		g.line("return " + expr)
	case *ast.ExprStmt:
		expr, err := g.genExpr(n.Expr)
		if err != nil {
			return err
		}
		g.line(expr)
	case *ast.ReturnStmt:
		if n.Value != nil {
			expr, err := g.genExpr(n.Value)
			if err != nil {
				return err
			}
			g.line("return " + expr)
		} else {
			g.line("return")
		}
	default:
		return fmt.Errorf("unknown statement node: %T", node)
	}
	return nil
}

func (g *Generator) genVarDecl(n *ast.VarDecl) error {
	goType := g.mapType(n.TypeName)
	val, err := g.genExpr(n.Value)
	if err != nil {
		return err
	}
	g.line(fmt.Sprintf("var %s %s = %s", n.Name, goType, val))
	g.line(fmt.Sprintf("_ = %s", n.Name))
	return nil
}

func (g *Generator) genArrayDecl(n *ast.ArrayDecl) error {
	goType := g.mapType(n.ElemType)
	var elems []string
	for _, e := range n.Elements {
		s, err := g.genExpr(e)
		if err != nil {
			return err
		}
		elems = append(elems, s)
	}
	g.line(fmt.Sprintf("%s := []%s{%s}", n.Name, goType, strings.Join(elems, ", ")))
	return nil
}

func (g *Generator) genTupleDecl(n *ast.TupleDecl) error {
	// Tuples become a struct or a named pair; we use a map[string]interface{} for simplicity
	// and generate a helper struct inline or just use individual vars
	// We'll represent tuples as []interface{} and print as JSON-like
	var vals []string
	for _, v := range n.Values {
		s, err := g.genExpr(v)
		if err != nil {
			return err
		}
		vals = append(vals, s)
	}
	g.line(fmt.Sprintf("%s := []interface{}{%s}", n.Name, strings.Join(vals, ", ")))
	return nil
}

func (g *Generator) genFuncDecl(n *ast.FuncDecl) error {
	var params []string
	for _, p := range n.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, g.mapType(p.TypeName)))
	}
	retType := ""
	if n.ReturnType != "" {
		retType = " " + g.mapType(n.ReturnType)
	}

	// Emit as closure variable so it works inside main()
	g.line("")
	g.line(fmt.Sprintf("%s := func(%s)%s {", n.Name, strings.Join(params, ", "), retType))
	g.indent++
	for _, stmt := range n.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}
	g.indent--
	g.line("}")
	g.line("")
	return nil
}

func (g *Generator) genIfStmt(n *ast.IfStmt) error {
	cond, err := g.genExpr(n.Condition)
	if err != nil {
		return err
	}
	g.line(fmt.Sprintf("if %s {", cond))
	g.indent++
	for _, stmt := range n.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}
	g.indent--

	for _, ei := range n.ElseIfs {
		eic, err := g.genExpr(ei.Condition)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("} else if %s {", eic))
		g.indent++
		for _, stmt := range ei.Body {
			if err := g.genStmt(stmt); err != nil {
				return err
			}
		}
		g.indent--
	}

	if len(n.Else) > 0 {
		g.line("} else {")
		g.indent++
		for _, stmt := range n.Else {
			if err := g.genStmt(stmt); err != nil {
				return err
			}
		}
		g.indent--
	}
	g.line("}")
	return nil
}

func (g *Generator) genForStmt(n *ast.ForStmt) error {
	target := n.Target
	if target == "" {
		target = "_orionArray"
	}
	g.line(fmt.Sprintf("for %s, %s := range %s {", n.IndexVar, n.ValueVar, target))
	g.indent++
	// suppress unused index warning
	g.line(fmt.Sprintf("_ = %s", n.IndexVar))
	for _, stmt := range n.Body {
		if err := g.genStmt(stmt); err != nil {
			return err
		}
	}
	g.indent--
	g.line("}")
	return nil
}

func (g *Generator) genIOWrite(n *ast.IOWrite) error {
	if len(n.Args) == 1 {
		expr, err := g.genExpr(n.Args[0])
		if err != nil {
			return err
		}
		// Check if it's a tuple/slice → print as JSON-like
		if isIdentifier(n.Args[0]) {
			g.line(fmt.Sprintf("fmt.Println(%s)", expr))
			return nil
		}
		g.line(fmt.Sprintf("fmt.Println(%s)", expr))
		return nil
	}
	var exprs []string
	for _, arg := range n.Args {
		e, err := g.genExpr(arg)
		if err != nil {
			return err
		}
		exprs = append(exprs, e)
	}
	g.line(fmt.Sprintf("fmt.Println(%s)", strings.Join(exprs, ", ")))
	return nil
}

func isIdentifier(n ast.Node) bool {
	_, ok := n.(*ast.Identifier)
	return ok
}

func (g *Generator) genExpr(node ast.Node) (string, error) {
	switch n := node.(type) {
	case *ast.StringLit:
		// Check for interpolation
		format, vars, hasInterp := parser.InterpolateString(n.Value)
		if hasInterp {
			g.useFmtStr = true
			varList := strings.Join(vars, ", ")
			return fmt.Sprintf(`fmt.Sprintf("%s", %s)`, format, varList), nil
		}
		return fmt.Sprintf("%q", n.Value), nil

	case *ast.IntLit:
		return n.Value, nil

	case *ast.FloatLit:
		return n.Value, nil

	case *ast.BoolLit:
		if n.Value {
			return "true", nil
		}
		return "false", nil

	case *ast.Identifier:
		return n.Name, nil

	case *ast.BinaryExpr:
		left, err := g.genExpr(n.Left)
		if err != nil {
			return "", err
		}
		right, err := g.genExpr(n.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", left, n.Op, right), nil

	case *ast.UnaryExpr:
		operand, err := g.genExpr(n.Operand)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s%s)", n.Op, operand), nil

	case *ast.FuncCall:
		var args []string
		for _, arg := range n.Args {
			s, err := g.genExpr(arg)
			if err != nil {
				return "", err
			}
			args = append(args, s)
		}
		return fmt.Sprintf("%s(%s)", n.Name, strings.Join(args, ", ")), nil

	case *ast.IOWrite:
		// When used as expression (shouldn't happen but handle gracefully)
		var args []string
		for _, arg := range n.Args {
			s, err := g.genExpr(arg)
			if err != nil {
				return "", err
			}
			args = append(args, s)
		}
		return fmt.Sprintf("fmt.Println(%s)", strings.Join(args, ", ")), nil

	case *ast.MethodCall:
		return g.genMethodCall(n)

	case *ast.ArrayLit:
		var elems []string
		for _, e := range n.Elements {
			s, err := g.genExpr(e)
			if err != nil {
				return "", err
			}
			elems = append(elems, s)
		}
		return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", ")), nil

	case *ast.TupleLit:
		var vals []string
		for _, v := range n.Values {
			s, err := g.genExpr(v)
			if err != nil {
				return "", err
			}
			vals = append(vals, s)
		}
		return fmt.Sprintf("[]interface{}{%s}", strings.Join(vals, ", ")), nil

	case *ast.IndexExpr:
		obj, err := g.genExpr(n.Object)
		if err != nil {
			return "", err
		}
		idx, err := g.genExpr(n.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s[%s]", obj, idx), nil

	default:
		return "", fmt.Errorf("unknown expression node: %T", node)
	}
}

func (g *Generator) genMethodCall(n *ast.MethodCall) (string, error) {
	switch n.Method {
	case "push":
		if len(n.Args) != 1 {
			return "", fmt.Errorf("push() requires 1 argument")
		}
		arg, err := g.genExpr(n.Args[0])
		if err != nil {
			return "", err
		}
		// emit as statement via side-effect: we return an assignment expression
		// Since this runs inside genExpr, we need to emit a statement.
		// We'll generate it as a special string that gets emitted directly.
		return fmt.Sprintf("func() { %s = append(%s, %s) }()", n.Object, n.Object, arg), nil

	case "pop":
		return fmt.Sprintf("func() interface{} { val := %s[len(%s)-1]; %s = %s[:len(%s)-1]; return val }()", n.Object, n.Object, n.Object, n.Object, n.Object), nil

	case "first":
		return fmt.Sprintf("%s[0]", n.Object), nil

	case "last":
		return fmt.Sprintf("%s[len(%s)-1]", n.Object, n.Object), nil

	case "remove":
		if len(n.Args) != 1 {
			return "", fmt.Errorf("remove() requires 1 argument (index)")
		}
		arg, err := g.genExpr(n.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() { %s = append(%s[:%s], %s[%s+1:]...) }()", n.Object, n.Object, arg, n.Object, arg), nil

	case "removeValue":
		if len(n.Args) != 1 {
			return "", fmt.Errorf("removeValue() requires 1 argument")
		}
		arg, err := g.genExpr(n.Args[0])
		if err != nil {
			return "", err
		}
		indent := strings.Repeat("\t", g.indent)
		var sb2 strings.Builder
		sb2.WriteString("func() {\n")
		sb2.WriteString(indent + "\tfor _i, _v := range " + n.Object + " {\n")
		sb2.WriteString(indent + "\t\tif fmt.Sprintf(\"%v\", _v) == fmt.Sprintf(\"%v\", " + arg + ") {\n")
		sb2.WriteString(indent + "\t\t\t" + n.Object + " = append(" + n.Object + "[:_i], " + n.Object + "[_i+1:]...)\n")
		sb2.WriteString(indent + "\t\t\tbreak\n")
		sb2.WriteString(indent + "\t\t}\n")
		sb2.WriteString(indent + "\t}\n")
		sb2.WriteString(indent + "}()")
		return sb2.String(), nil

	default:
		return "", fmt.Errorf("unknown method: %s", n.Method)
	}
}
