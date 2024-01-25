package ast

import (
	"bytes"
	"strings"

	"github.com/rielj/go-interpreter/token"
)

// Node is a type that represents a node in the AST
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is a type that implements the Node interface
type Statement interface {
	Node
	statementNode()
}

// Expression is a type that implements the Node interface
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST our parser produces
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	// Check if there are any statements
	if len(p.Statements) > 0 {
		// Return the token literal of the first statement
		return p.Statements[0].TokenLiteral()
	} else {
		// Return an empty string
		return ""
	}
}

// String returns the string representation of the program
func (p *Program) String() string {
	// Create a string builder
	var out string

	// Loop through all the statements
	for _, s := range p.Statements {
		// Append the string representation of the statement
		out += s.String()
	}

	// Return the string
	return out
}

// ExpressionStatement is a type that implements the Statement interface
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression  // the expression itself
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the literal value of the token
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

// String returns the string representation of the expression statement
func (es *ExpressionStatement) String() string {
	// Check if the expression is not nil
	if es.Expression != nil {
		// Return the string representation of the expression
		return es.Expression.String()
	}

	// Return an empty string
	return ""
}

// ReturnStatement is a type that implements the Statement interface
type ReturnStatement struct {
	Token       token.Token // the token.RETURN token
	ReturnValue Expression  // the value the return statement returns
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the literal value of the token
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

// String returns the string representation of the return statement
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	// Write the return token
	out.WriteString(rs.TokenLiteral() + " ")

	// Check if the return value is not nil
	if rs.ReturnValue != nil {
		// Write the return value
		out.WriteString(rs.ReturnValue.String())
	}

	// Write the semicolon
	out.WriteString(";")

	// Return the string
	return out.String()
}

// LetStatement is a type that implements the Statement interface
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier // the name of the variable
	Value Expression  // the value the variable is bound to
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the literal value of the token
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String returns the string representation of the let statement
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	// Write the let token
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	// Check if the value is not nil
	if ls.Value != nil {
		// Write the value
		out.WriteString(ls.Value.String())
	}

	// Write the semicolon
	out.WriteString(";")

	// Return the string
	return out.String()
}

// Identifier is a type that implements the Expression interface
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string      // the value of the identifier
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// String returns the string representation of the identifier
func (i *Identifier) String() string {
	return i.Value
}

// IntegerLiteral is a type that implements the Expression interface
type IntegerLiteral struct {
	Token token.Token // the token.INT token
	Value int64       // the value of the integer literal
}

func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

// String returns the string representation of the integer literal
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// PrefixExpression is a type that implements the Expression interface
type PrefixExpression struct {
	Token    token.Token // the prefix token, e.g. !
	Operator string      // the operator, e.g. !
	Right    Expression  // the right expression
}

func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

// String returns the string representation of the prefix expression
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	// Write the operator
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	// Return the string
	return out.String()
}

// InfixExpression is a type that implements the Expression interface
type InfixExpression struct {
	Token    token.Token // the infix token, e.g. +
	Left     Expression  // the left expression
	Operator string      // the operator, e.g. +
	Right    Expression  // the right expression
}

func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String returns the string representation of the infix expression
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	// Write the left expression
	out.WriteString("(")
	out.WriteString(ie.Left.String())

	// Write the operator
	out.WriteString(" " + ie.Operator + " ")

	// Write the right expression
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	// Return the string
	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

// String returns the string representation of the boolean
func (b *Boolean) String() string {
	return b.Token.Literal
}

type IfExpression struct {
	Token       token.Token // the token.IF token
	Condition   Expression  // the condition
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String returns the string representation of the if expression
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	// Write the if token
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	// Check if there is an alternative
	if ie.Alternative != nil {
		// Write the else token
		out.WriteString("else")
		out.WriteString(ie.Alternative.String())
	}

	// Return the string
	return out.String()
}

type BlockStatement struct {
	Token      token.Token // the token.LBRACE token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the literal value of the token
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

// String returns the string representation of the block statement
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	// Loop through all the statements
	for _, s := range bs.Statements {
		// Write the statement
		out.WriteString(s.String())
	}

	// Return the string
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // the token.FUNCTION token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

// String returns the string representation of the function literal
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	// Write the function token
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	// Write the parameters
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	// Return the string
	return out.String()
}

type CallExpression struct {
	Token     token.Token // the token.LPAREN token
	Function  Expression  // the function to call
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

// String returns the string representation of the call expression
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	// Write the function
	out.WriteString(ce.Function.String())
	out.WriteString("(")

	// Write the arguments
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	// Return the string
	return out.String()
}
