package lexer

import "github.com/rielj/go-interpreter/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	// Operators
	case '=':
		// Check if the next character is an equal sign
		if l.peekChar() == '=' {
			// Save the current character
			ch := l.ch
			// Read the next character
			l.readChar()
			// Create a token
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			// Create a token
			tok = newToken(token.ASSIGN, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		// Check if the next character is an equal sign
		if l.peekChar() == '=' {
			// Save the current character
			ch := l.ch
			// Read the next character
			l.readChar()
			// Create a token
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			// Create a token
			tok = newToken(token.BANG, l.ch)
		}
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	// Delimiters
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '"':
		tok.Type = token.STRING
		// Read the string
		tok.Literal = l.readString()
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	// End of file
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		// If the character is a letter, we want to read the entire identifier
		if isLetter(l.ch) {
			// Read the identifier
			tok.Literal = l.readIdentifier()
			// Check if the identifier is a keyword
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			// Read the number
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// Read the next character
	l.readChar()
	return tok
}

// Peek at the next character
func (l *Lexer) peekChar() byte {
	// Check if the reading position is at the end of the input
	if l.readPosition >= len(l.input) {
		// ASCII code for "NUL"
		return 0
	} else {
		// ASCII code for the character at the reading position
		return l.input[l.readPosition]
	}
}

// Skip the whitespace
func (l *Lexer) skipWhitespace() {
	// Read the next character while the current character is a whitespace
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// Read the entire number
func (l *Lexer) readNumber() string {
	// Save the current position
	position := l.position
	// Read the next character
	for isDigit(l.ch) {
		l.readChar()
	}
	// Return the number
	return l.input[position:l.position]
}

// Read the entire string
func (l *Lexer) readString() string {
	// Save the current position
	position := l.position + 1
	// Read the next character
	for {
		l.readChar()
		// Check if the current character is a double quote or the end of the input
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	// Return the string
	return l.input[position:l.position]
}

func (l *Lexer) readChar() {
	// Check if the reading position is at the end of the input
	if l.readPosition >= len(l.input) {
		// ASCII code for "NUL"
		l.ch = 0
	} else {
		// ASCII code for the character at the current position
		l.ch = l.input[l.readPosition]
	}
	// Increment the current position and reading position by 1
	l.position = l.readPosition
	l.readPosition += 1
}

// Read the entire identifier
func (l *Lexer) readIdentifier() string {
	// Save the current position
	position := l.position
	// Read the next character
	for isLetter(l.ch) {
		l.readChar()
	}
	// Return the identifier
	return l.input[position:l.position]
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	// Read the first character
	l.readChar()
	return l
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	// Convert the byte to a string
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// Check if the character is a letter
func isLetter(ch byte) bool {
	// Check if the character is a letter or an underscore
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// Check if the character is a digit
func isDigit(ch byte) bool {
	// Check if the character is a digit
	return '0' <= ch && ch <= '9'
}
