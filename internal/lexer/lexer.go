package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

func New(input string) *Lexer {
	return &Lexer{input: []rune(input), pos: 0, line: 1, col: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekAt(offset int) rune {
	p := l.pos + offset
	if p >= len(l.input) {
		return 0
	}
	return l.input[p]
}

func (l *Lexer) advance() rune {
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\r') {
		l.advance()
	}
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.input) && l.peek() != '\n' {
		l.advance()
	}
}

func (l *Lexer) readString() string {
	var sb strings.Builder
	l.advance() // consume opening "
	for l.pos < len(l.input) && l.peek() != '"' {
		ch := l.advance()
		if ch == '\\' && l.pos < len(l.input) {
			next := l.advance()
			switch next {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(next)
			}
		} else {
			sb.WriteRune(ch)
		}
	}
	l.advance() // consume closing "
	return sb.String()
}

func (l *Lexer) readNumber() (TokenType, string) {
	start := l.pos
	isFloat := false
	for l.pos < len(l.input) && (unicode.IsDigit(l.peek()) || l.peek() == '.') {
		if l.peek() == '.' {
			isFloat = true
		}
		l.advance()
	}
	lit := string(l.input[start:l.pos])
	if isFloat {
		return TOKEN_FLOAT, lit
	}
	return TOKEN_INT, lit
}

func (l *Lexer) readIdent() string {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	return string(l.input[start:l.pos])
}

var keywords = map[string]TokenType{
	"if":      TOKEN_IF,
	"or_if":   TOKEN_OR_IF,
	"or":      TOKEN_OR,
	"for":     TOKEN_FOR,
	"true":    TOKEN_BOOL,
	"false":   TOKEN_BOOL,
	"string":  TOKEN_TYPE_STRING,
	"integer": TOKEN_TYPE_INT,
	"int":     TOKEN_TYPE_INT,
	"bool":    TOKEN_TYPE_BOOL,
	"float":   TOKEN_TYPE_FLOAT,
	"return":  TOKEN_RETURN,
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for l.pos < len(l.input) {
		l.skipWhitespace()

		if l.pos >= len(l.input) {
			break
		}

		line, col := l.line, l.col
		ch := l.peek()

		// Newlines
		if ch == '\n' {
			l.advance()
			tokens = append(tokens, Token{TOKEN_NEWLINE, "\\n", line, col})
			continue
		}

		// Line comments
		if ch == '/' && l.peekAt(1) == '/' {
			l.skipLineComment()
			continue
		}

		// String literals
		if ch == '"' {
			s := l.readString()
			tokens = append(tokens, Token{TOKEN_STRING, s, line, col})
			continue
		}

		// Numbers
		if unicode.IsDigit(ch) {
			tt, lit := l.readNumber()
			tokens = append(tokens, Token{tt, lit, line, col})
			continue
		}

		// Identifiers and keywords
		if unicode.IsLetter(ch) || ch == '_' {
			ident := l.readIdent()
			tt, ok := keywords[ident]
			if !ok {
				tt = TOKEN_IDENT
			}
			tokens = append(tokens, Token{tt, ident, line, col})
			continue
		}

		// Two-char operators
		if ch == ':' && l.peekAt(1) == ':' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_DCOLON, "::", line, col})
			continue
		}
		if ch == ':' {
			l.advance()
			tokens = append(tokens, Token{TOKEN_COLON, ":", line, col})
			continue
		}
		if ch == '=' && l.peekAt(1) == '=' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_EQ, "==", line, col})
			continue
		}
		if ch == '!' && l.peekAt(1) == '=' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_NEQ, "!=", line, col})
			continue
		}
		if ch == '<' && l.peekAt(1) == '=' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_LTE, "<=", line, col})
			continue
		}
		if ch == '>' && l.peekAt(1) == '=' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_GTE, ">=", line, col})
			continue
		}
		if ch == '&' && l.peekAt(1) == '&' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_AND, "&&", line, col})
			continue
		}
		if ch == '|' && l.peekAt(1) == '|' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_OR_OP, "||", line, col})
			continue
		}
		if ch == '-' && l.peekAt(1) == '>' {
			l.advance()
			l.advance()
			tokens = append(tokens, Token{TOKEN_ARROW, "->", line, col})
			continue
		}

		// Single char tokens
		single := map[rune]TokenType{
			'(': TOKEN_LPAREN, ')': TOKEN_RPAREN,
			'{': TOKEN_LBRACE, '}': TOKEN_RBRACE,
			'[': TOKEN_LBRACKET, ']': TOKEN_RBRACKET,
			',': TOKEN_COMMA, ';': TOKEN_SEMI,
			'.': TOKEN_DOT, '=': TOKEN_ASSIGN,
			'+': TOKEN_PLUS, '-': TOKEN_MINUS,
			'*': TOKEN_STAR, '/': TOKEN_SLASH,
			'<': TOKEN_LT, '>': TOKEN_GT,
			'!': TOKEN_NOT,
		}
		if tt, ok := single[ch]; ok {
			l.advance()
			tokens = append(tokens, Token{tt, string(ch), line, col})
			continue
		}

		return nil, fmt.Errorf("line %d col %d: unexpected character '%c'", line, col, ch)
	}

	tokens = append(tokens, Token{TOKEN_EOF, "", l.line, l.col})
	return tokens, nil
}
