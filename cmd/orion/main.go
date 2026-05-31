package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"orion/internal/codegen"
	"orion/internal/lexer"
	"orion/internal/parser"
)

func main() {
	var (
		run    = flag.Bool("run", false, "Compile and run the output immediately")
		output = flag.String("o", "", "Output file (default: <input>.go)")
		debug  = flag.Bool("debug", false, "Print tokens and AST for debugging")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Orion transpiler v0.1\n\nUsage:\n  orion [flags] <input.or>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	inputFile := args[0]
	if filepath.Ext(inputFile) != ".or" {
		fatalf("input file must have .or extension, got %q", filepath.Ext(inputFile))
	}

	src, err := os.ReadFile(inputFile)
	if err != nil {
		fatalf("error reading %s: %v", inputFile, err)
	}

	// Lexing
	l := lexer.New(string(src))
	tokens, err := l.Tokenize()
	if err != nil {
		fatalf("lexer error: %v", err)
	}

	if *debug {
		fmt.Println("=== TOKENS ===")
		for _, t := range tokens {
			fmt.Printf("  [%s] %q  (line %d)\n", t.Type, t.Literal, t.Line)
		}
	}

	// Parsing
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		fatalf("parse error: %v", err)
	}

	if *debug {
		fmt.Printf("\n=== AST: %d top-level statements ===\n", len(prog.Stmts))
		for i, s := range prog.Stmts {
			fmt.Printf("  [%d] %T\n", i, s)
		}
	}

	// Code generation
	gen := codegen.New()
	goCode, err := gen.Generate(prog)
	if err != nil {
		fatalf("codegen error: %v", err)
	}

	// Determine output path
	outPath := *output
	if outPath == "" {
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outPath = filepath.Join(filepath.Dir(inputFile), base+".go")
	}

	if err := os.WriteFile(outPath, []byte(goCode), 0644); err != nil {
		fatalf("error writing %s: %v", outPath, err)
	}

	fmt.Printf("transpiled from %s to %s\n", inputFile, outPath)

	// Optionally run
	if *run {
		cmd := exec.Command("go", "run", outPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fatalf("run error: %v", err)
		}
	}
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "orion: "+format+"\n", a...)
	os.Exit(1)
}
