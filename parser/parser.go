package parser

import (
	"fmt"
	"strconv"

	"github.com/rielj/go-interpreter/ast"
	"github.com/rielj/go-interpreter/lexer"
	"github.com/rielj/go-interpreter/token"
)

const (
	_ int = iota
	// LOWEST is the lowest precedence
	LOWEST
	// EQUALS is the equals precedence
	EQUALS // ==
	// LESSGREATER is the less/greater precedence
	LESSGREATER // > or <
	// SUM is the sum precedence
	SUM // +
	// PRODUCT is the product precedence
	PRODUCT // *
	// PREFIX is the prefix precedence
	PREFIX // -X or !X
	// CALL is the call precedence
	CALL // myFunction(X)
	// INDEX is the index precedence
	INDEX // array[index]
)

// precedences
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

// Parser is a type that represents a parser
type Parser struct {
	l *lexer.Lexer

	errors []string

	curToken  token.Token
	peekToken token.Token

	// Prefix and infix parsing functions
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// Register prefix parsing functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)

	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.registerPrefix(token.IF, p.parseIfExpression)

	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.registerPrefix(token.STRING, p.parseStringLiteral)

	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	// Register infix parsing functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens to set curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses a program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Loop through all the tokens until we reach the end of the file
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// Read the next token
		p.nextToken()
	}

	return program
}

// nextToken reads the next token from the lexer and sets curToken and peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// RegisterPrefix registers a prefix parsing function
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// RegisterInfix registers an infix parsing function
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// parseIdentifier parses an identifier
func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseBoolean parses a boolean
func (p *Parser) parseBoolean() ast.Expression {
	defer untrace(trace("parseBoolean"))
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression parses a grouped expression
func (p *Parser) parseGroupedExpression() ast.Expression {
	defer untrace(trace("parseGroupedExpression"))
	// Read the next token
	p.nextToken()

	// Parse the expression
	exp := p.parseExpression(LOWEST)

	// Check if the next token is a closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression parses an if expression
func (p *Parser) parseIfExpression() ast.Expression {
	defer untrace(trace("parseIfExpression"))
	expression := &ast.IfExpression{Token: p.curToken}

	// Check if the next token is an opening parenthesis
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Read the next token
	p.nextToken()

	// Parse the condition
	expression.Condition = p.parseExpression(LOWEST)

	// Check if the next token is a closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Check if the next token is an opening brace
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the consequence
	expression.Consequence = p.parseBlockStatement()

	// Check if the next token is an else
	if p.peekTokenIs(token.ELSE) {
		// Read the next token
		p.nextToken()

		// Check if the next token is an opening brace
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// Parse the alternative
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseHashLiteral parses a hash literal
func (p *Parser) parseHashLiteral() ast.Expression {
	defer untrace(trace("parseHashLiteral"))
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	// Loop through all the pairs until we reach a closing brace
	for !p.peekTokenIs(token.RBRACE) {
		// Read the next token
		p.nextToken()

		// Parse the key
		key := p.parseExpression(LOWEST)

		// Check if the next token is a colon
		if !p.expectPeek(token.COLON) {
			return nil
		}

		// Read the next token
		p.nextToken()

		// Parse the value
		value := p.parseExpression(LOWEST)

		// Add the pair
		hash.Pairs[key] = value

		// Check if the next token is a comma
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	// Check if the next token is a closing brace
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

// parseArrayLiteral parses an array literal
func (p *Parser) parseArrayLiteral() ast.Expression {
	defer untrace(trace("parseArrayLiteral"))
	array := &ast.ArrayLiteral{Token: p.curToken}

	// Parse the elements
	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

// parseExpressionList parses an expression list
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	defer untrace(trace("parseExpressionList"))
	list := []ast.Expression{}

	// Check if the next token is the end token
	if p.peekTokenIs(end) {
		// Read the next token
		p.nextToken()
		return list
	}

	// Read the next token
	p.nextToken()

	// Parse the first expression
	list = append(list, p.parseExpression(LOWEST))

	// Loop through all the expressions
	for p.peekTokenIs(token.COMMA) {
		// Read the next token
		p.nextToken()
		p.nextToken()

		// Parse the expression
		list = append(list, p.parseExpression(LOWEST))
	}

	// Check if the next token is the end token
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseStringLiteral parses a string literal
func (p *Parser) parseStringLiteral() ast.Expression {
	defer untrace(trace("parseStringLiteral"))
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseFunctionLiteral parses a function literal
func (p *Parser) parseFunctionLiteral() ast.Expression {
	defer untrace(trace("parseFunctionLiteral"))
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// Check if the next token is an opening parenthesis
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Parse the parameters
	lit.Parameters = p.parseFunctionParameters()

	// Check if the next token is an opening brace
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the body
	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses function parameters
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	defer untrace(trace("parseFunctionParameters"))
	// Check if the next token is a closing parenthesis
	if p.peekTokenIs(token.RPAREN) {
		// Read the next token
		p.nextToken()
		return []*ast.Identifier{}
	}

	// Read the next token
	p.nextToken()

	// Create a list of identifiers
	identifiers := []*ast.Identifier{{Token: p.curToken, Value: p.curToken.Literal}}

	// Loop through all the identifiers
	for p.peekTokenIs(token.COMMA) {
		// Read the next token
		p.nextToken()
		p.nextToken()

		// Add the identifier
		identifiers = append(
			identifiers,
			&ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		)
	}

	// Check if the next token is a closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// parseBlockStatement parses a block statement
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	defer untrace(trace("parseBlockStatement"))
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	// Read the next token
	p.nextToken()

	// Loop through all the statements until we reach a closing brace
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Parse the statement
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// Read the next token
		p.nextToken()
	}

	return block
}

// parseIntegerLiteral parses an integer literal
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// Try to parse the integer
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(
			p.errors,
			msg,
		)
		return nil
	}

	// Set the value
	lit.Value = value

	return lit
}

// parseInfixExpression parses an infix expression
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// Get the precedence of the current token
	precedence := p.curPrecedence()

	// Read the next token
	p.nextToken()

	// Parse the right expression
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseIndexExpression parses an index expression
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseIndexExpression"))
	expression := &ast.IndexExpression{Token: p.curToken, Left: left}

	// Read the next token
	p.nextToken()

	// Parse the index
	expression.Index = p.parseExpression(LOWEST)

	// Check if the next token is a closing bracket
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return expression
}

// parseCallExpression parses a call expression
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	defer untrace(trace("parseCallExpression"))
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// curPrecedence returns the precedence of the current token
func (p *Parser) curPrecedence() int {
	// Check if there is a precedence for the current token
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	// Return the lowest precedence if there is none
	return LOWEST
}

// peekPrecedence returns the precedence of the next token
func (p *Parser) peekPrecedence() int {
	// Check if there is a precedence for the next token
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	// Return the lowest precedence if there is none
	return LOWEST
}

// parsePrefixExpression parses a prefix expression
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	// Read the next token
	p.nextToken()

	// Parse the right expression
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		// Parse a let statement
		return p.parseLetStatement()
	case token.RETURN:
		// Parse a return statement
		return p.parseReturnStatement()
	default:
		// Parse an expression statement
		return p.parseExpressionStatement()
	}
}

// parseExpressionStatement parses an expression statement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// Parse the expression
	stmt.Expression = p.parseExpression(
		LOWEST,
	)

	// Check if the next token is a semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// noPrefixParseFnError adds an error to the parser
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(
		p.errors,
		msg,
	)
}

// parseExpression parses an expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))
	// Check if there is a prefix parsing function for the current token
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// Parse the prefix expression
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// Check if there is an infix parsing function for the next token
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// Read the next token
		p.nextToken()

		// Parse the infix expression
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseLetStatement parses a let statement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))
	stmt := &ast.LetStatement{Token: p.curToken}

	// Check if the next token is an identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Set the identifier
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check if the next token is an equal sign
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	// Parse the expression
	stmt.Value = p.parseExpression(LOWEST)

	// Check if the next token is a semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer untrace(trace("parseReturnStatement"))
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// Parse the return value
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// Check if the next token is a semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// curTokenIs checks if the current token is of a certain type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the next token is of a certain type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// peekError checks if the next token is of a certain type
// and adds an error if it is not
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t,
		p.peekToken.Type,
	)
	p.errors = append(
		p.errors,
		msg,
	)
}

// expectPeek checks if the next token is of a certain type
// and advances the tokens if it is
func (p *Parser) expectPeek(t token.TokenType) bool {
	// Check if the next token is of the expected type
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}
