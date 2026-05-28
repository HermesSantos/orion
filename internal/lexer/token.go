package lexer

type TokenType string

const (
	// Literals
	TOKEN_IDENT   TokenType = "IDENT"
	TOKEN_STRING  TokenType = "STRING"
	TOKEN_INT     TokenType = "INT"
	TOKEN_FLOAT   TokenType = "FLOAT"
	TOKEN_BOOL    TokenType = "BOOL"

	// Types
	TOKEN_TYPE_STRING  TokenType = "TYPE_STRING"
	TOKEN_TYPE_INT     TokenType = "TYPE_INT"
	TOKEN_TYPE_BOOL    TokenType = "TYPE_BOOL"
	TOKEN_TYPE_FLOAT   TokenType = "TYPE_FLOAT"

	// Keywords
	TOKEN_IF      TokenType = "if"
	TOKEN_OR_IF   TokenType = "or_if"
	TOKEN_OR      TokenType = "or"
	TOKEN_FOR     TokenType = "for"
	TOKEN_RETURN  TokenType = "return"

	// Operators
	TOKEN_ASSIGN  TokenType = "="
	TOKEN_PLUS    TokenType = "+"
	TOKEN_MINUS   TokenType = "-"
	TOKEN_STAR    TokenType = "*"
	TOKEN_SLASH   TokenType = "/"
	TOKEN_EQ      TokenType = "=="
	TOKEN_NEQ     TokenType = "!="
	TOKEN_LT      TokenType = "<"
	TOKEN_GT      TokenType = ">"
	TOKEN_LTE     TokenType = "<="
	TOKEN_GTE     TokenType = ">="
	TOKEN_AND     TokenType = "&&"
	TOKEN_OR_OP   TokenType = "||"
	TOKEN_NOT     TokenType = "!"

	// Delimiters
	TOKEN_LPAREN   TokenType = "("
	TOKEN_RPAREN   TokenType = ")"
	TOKEN_LBRACE   TokenType = "{"
	TOKEN_RBRACE   TokenType = "}"
	TOKEN_LBRACKET TokenType = "["
	TOKEN_RBRACKET TokenType = "]"
	TOKEN_COMMA    TokenType = ","
	TOKEN_SEMI     TokenType = ";"
	TOKEN_COLON    TokenType = ":"
	TOKEN_DCOLON   TokenType = "::"
	TOKEN_DOT      TokenType = "."

	// Special
	TOKEN_ARROW   TokenType = "->"
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_COMMENT TokenType = "COMMENT"
	TOKEN_NEWLINE TokenType = "NEWLINE"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}
