package parser

import (
	"fmt"
	"testing"

	"github.com/rielj/go-interpreter/ast"
	"github.com/rielj/go-interpreter/lexer"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedName  string
		expectedValue interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	// Loop through the tests
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		// Check the parser errors
		checkParserErrors(t, p)

		// Check the length of the program
		if len(program.Statements) != 1 {
			t.Fatalf(
				"program has not enough statements. got=%d",
				len(program.Statements),
			)
		}

		// Check the type of the statement
		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedName) {
			return
		}

		// Type assertion
		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

// testLetStatement tests the let statement
func testLetStatement(
	t *testing.T,
	stmt ast.Statement,
	name string,
) bool {
	if stmt.TokenLiteral() != "let" {
		t.Errorf(
			"stmt.TokenLiteral not 'let'. got=%q",
			stmt.TokenLiteral(),
		)
		return false
	}

	// Type assertion
	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		t.Errorf(
			"stmt not *ast.LetStatement. got=%T",
			stmt,
		)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf(
			"letStmt.Name.Value not '%s'. got=%s",
			name,
			letStmt.Name.Value,
		)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf(
			"letStmt.Name.TokenLiteral() not '%s'. got=%s",
			name,
			letStmt.Name.TokenLiteral(),
		)
		return false
	}

	return true
}

// checkParserErrors checks the parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	// Print the errors
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	// Check the parser errors
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf(
			"program.Statements does not contain 3 statements. got=%d",
			len(program.Statements),
		)
	}

	// Loop through the statements and test the return statement
	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf(
				"stmt not *ast.ReturnStatement. got=%T",
				stmt,
			)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf(
				"returnStmt.TokenLiteral not 'return', got %q",
				returnStmt.TokenLiteral(),
			)
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf(
			"exp not *ast.Identifier. got=%T",
			stmt.Expression,
		)
	}

	// Check the value of the identifier
	if ident.Value != "foobar" {
		t.Errorf(
			"ident.Value not %s. got=%s",
			"foobar",
			ident.Value,
		)
	}

	// Check the literal value of the identifier
	if ident.TokenLiteral() != "foobar" {
		t.Errorf(
			"ident.TokenLiteral not %s. got=%s",
			"foobar",
			ident.TokenLiteral(),
		)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf(
			"exp not *ast.IntegerLiteral. got=%T",
			stmt.Expression,
		)
	}

	// Check the value of the literal
	if literal.Value != 5 {
		t.Errorf(
			"literal.Value not %d. got=%d",
			5,
			literal.Value,
		)
	}

	// Check the literal value of the identifier
	if literal.TokenLiteral() != "5" {
		t.Errorf(
			"literal.TokenLiteral not %s. got=%s",
			"5",
			literal.TokenLiteral(),
		)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	// Loop through the tests
	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Check the length of the program
		if len(program.Statements) != 1 {
			t.Fatalf(
				"program has not enough statements. got=%d",
				len(program.Statements),
			)
		}

		// Check the type of the statement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf(
				"program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0],
			)
		}

		// Check the type of the expression
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf(
				"stmt is not ast.PrefixExpression. got=%T",
				stmt.Expression,
			)
		}

		// Check the operator
		if exp.Operator != tt.operator {
			t.Fatalf(
				"exp.Operator is not '%s'. got=%s",
				tt.operator,
				exp.Operator,
			)
		}

		// Check the value of the literal
		if !testLiteralExpression(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

// testIntegerLiteral tests the integer literal
func testIntegerLiteral(
	t *testing.T,
	il ast.Expression,
	value int64,
) bool {
	// Type assertion
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf(
			"il not *ast.IntegerLiteral. got=%T",
			il,
		)
		return false
	}

	// Check the value of the literal
	if integ.Value != value {
		t.Errorf(
			"integ.Value not %d. got=%d",
			value,
			integ.Value,
		)
		return false
	}

	// Check the literal value of the identifier
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf(
			"integ.TokenLiteral not %d. got=%s",
			value,
			integ.TokenLiteral(),
		)
		return false
	}

	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	// Loop through the tests
	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Check the length of the program
		if len(program.Statements) != 1 {
			t.Fatalf(
				"program has not enough statements. got=%d",
				len(program.Statements),
			)
		}

		// Check the type of the statement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf(
				"program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0],
			)
		}

		// Check the type of the expression
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf(
				"stmt is not ast.InfixExpression. got=%T",
				stmt.Expression,
			)
		}

		// Check the operator
		if exp.Operator != tt.operator {
			t.Fatalf(
				"exp.Operator is not '%s'. got=%s",
				tt.operator,
				exp.Operator,
			)
		}

		// Check the value of the literal
		if !testLiteralExpression(t, exp.Left, tt.leftValue) {
			return
		}

	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"1 + (2 + 3) + 4;",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2;",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5);",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5);",
			"(-(5 + 5))",
		},
		{
			"!(true == true);",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d;",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8));",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g);",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
	}

	// Loop through the tests
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Check the string representation of the program
		actual := program.String()
		if actual != tt.expected {
			t.Errorf(
				"expected=%q, got=%q",
				tt.expected,
				actual,
			)
		}
		t.Logf("TestOperatorPrecedenceParsing passed, parsed %d statements", len(tests))
		t.Logf("Program: %s", program.String())
		t.Logf("Expected: %s", tt.expected)
		t.Logf("Actual: %s", actual)
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	// Type assertion
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf(
			"exp not *ast.Identifier. got=%T",
			exp,
		)
		return false
	}

	// Check the value of the literal
	if ident.Value != value {
		t.Errorf(
			"ident.Value not %s. got=%s",
			value,
			ident.Value,
		)
		return false
	}

	// Check the literal value of the identifier
	if ident.TokenLiteral() != value {
		t.Errorf(
			"ident.TokenLiteral not %s. got=%s",
			value,
			ident.TokenLiteral(),
		)
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf(
		"type of exp not handled. got=%T",
		exp,
	)
	return false
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	left interface{},
	operator string,
	right interface{},
) bool {
	// Type assertion
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf(
			"exp not *ast.InfixExpression. got=%T(%s)",
			exp,
			exp,
		)
		return false
	}

	// Check the operator
	if opExp.Operator != operator {
		t.Errorf(
			"exp.Operator is not '%s'. got=%s",
			operator,
			opExp.Operator,
		)
		return false
	}

	// Check the value of the literal
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	// Check the value of the literal
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testBooleanLiteral(
	t *testing.T,
	exp ast.Expression,
	value bool,
) bool {
	// Type assertion
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf(
			"exp not *ast.Boolean. got=%T",
			exp,
		)
		return false
	}

	// Check the value of the literal
	if bo.Value != value {
		t.Errorf(
			"bo.Value not %t. got=%t",
			value,
			bo.Value,
		)
		return false
	}

	// Check the literal value of the identifier
	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf(
			"bo.TokenLiteral not %t. got=%s",
			value,
			bo.TokenLiteral(),
		)
		return false
	}

	return true
}

func TestIfExpression(t *testing.T) {
	input := `
		if (x < y) { x }
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// Check the parser errors
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf(
			"stmt is not ast.IfExpression. got=%T",
			stmt.Expression,
		)
	}

	// Check the condition
	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	// Check the consequence
	if len(exp.Consequence.Statements) != 1 {
		t.Errorf(
			"consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements),
		)
	}

	// Check the consequence
	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0],
		)
	}

	// Check the consequence
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	// Check the alternative
	if exp.Alternative != nil {
		t.Errorf(
			"exp.Alternative.Statements was not nil. got=%+v",
			exp.Alternative,
		)
	}

	t.Logf("TestIfExpression passed, parsed %d statements", len(program.Statements))
	t.Logf("Program: %s", program.String())
}

func TestIfElseExpression(t *testing.T) {
	input := `
		if (x < y) { x } else { y }
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// Check the parser errors
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf(
			"stmt is not ast.IfExpression. got=%T",
			stmt.Expression,
		)
	}

	// Check the condition
	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	// Check the consequence
	if len(exp.Consequence.Statements) != 1 {
		t.Errorf(
			"consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements),
		)
	}

	// Check the consequence
	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0],
		)
	}

	// Check the consequence
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	// Check the alternative
	if len(exp.Alternative.Statements) != 1 {
		t.Errorf(
			"alternative is not 1 statements. got=%d\n",
			len(exp.Alternative.Statements),
		)
	}

	// Check the alternative
	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0],
		)
	}

	// Check the alternative
	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}

	t.Logf("TestIfElseExpression passed, parsed %d statements", len(program.Statements))
	t.Logf("Program: %s", program.String())
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// Check the parser errors
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf(
			"stmt is not ast.FunctionLiteral. got=%T",
			stmt.Expression,
		)
	}

	// Check the parameters
	if len(function.Parameters) != 2 {
		t.Fatalf(
			"function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters),
		)
	}

	// Check the first parameter
	testLiteralExpression(t, function.Parameters[0], "x")

	// Check the second parameter
	testLiteralExpression(t, function.Parameters[1], "y")

	// Check the body
	if len(function.Body.Statements) != 1 {
		t.Fatalf(
			"function.Body.Statements has not 1 statements. got=%d\n",
			len(function.Body.Statements),
		)
	}

	// Check the body
	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0],
		)
	}

	// Check the body
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")

	t.Logf("TestFunctionLiteralParsing passed, parsed %d statements", len(program.Statements))
	t.Logf("Program: %s", program.String())
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{"fn() {};", []string{}},
		{"fn(x) {};", []string{"x"}},
		{"fn(x, y, z) {};", []string{"x", "y", "z"}},
	}

	// Loop through the tests
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		// Check the parser errors
		checkParserErrors(t, p)

		// Check the length of the program
		if len(program.Statements) != 1 {
			t.Fatalf(
				"program has not enough statements. got=%d",
				len(program.Statements),
			)
		}

		// Check the type of the statement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf(
				"program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0],
			)
		}

		// Check the type of the expression
		function, ok := stmt.Expression.(*ast.FunctionLiteral)
		if !ok {
			t.Fatalf(
				"stmt is not ast.FunctionLiteral. got=%T",
				stmt.Expression,
			)
		}

		// Check the parameters
		if len(function.Parameters) != len(tt.expectedParams) {
			t.Fatalf(
				"length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams),
				len(function.Parameters),
			)
		}

		// Loop through the parameters
		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}

		t.Logf("TestFunctionLiteralParsing passed, parsed %d statements", len(program.Statements))
		t.Logf("Program: %s", program.String())
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := `add(1, 2 * 3, 4 + 5);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// Check the parser errors
	checkParserErrors(t, p)

	// Check the length of the program
	if len(program.Statements) != 1 {
		t.Fatalf(
			"program has not enough statements. got=%d",
			len(program.Statements),
		)
	}

	// Check the type of the statement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf(
			"program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0],
		)
	}

	// Check the type of the expression
	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf(
			"stmt is not ast.CallExpression. got=%T",
			stmt.Expression,
		)
	}

	// Check the function name
	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	// Check the arguments
	if len(exp.Arguments) != 3 {
		t.Fatalf(
			"wrong length of arguments. got=%d",
			len(exp.Arguments),
		)
	}

	// Check the arguments
	testLiteralExpression(t, exp.Arguments[0], 1)

	// Check the arguments
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)

	// Check the arguments
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)

	t.Logf("TestCallExpressionParsing passed, parsed %d statements", len(program.Statements))
	t.Logf("Program: %s", program.String())
}

func TestCallExpressionParameterParsing(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedArguments  []string
	}{
		{
			"add();",
			"add",
			[]string{},
		},
		{
			"add(1);",
			"add",
			[]string{"1"},
		},
		{
			"add(1, 2 * 3, 4 + 5);",
			"add",
			[]string{"1", "(2 * 3)", "(4 + 5)"},
		},
	}

	// Loop through the tests
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		// Check the parser errors
		checkParserErrors(t, p)

		// Check the length of the program
		if len(program.Statements) != 1 {
			t.Fatalf(
				"program has not enough statements. got=%d",
				len(program.Statements),
			)
		}

		// Check the type of the statement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf(
				"program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0],
			)
		}

		// Check the type of the expression
		exp, ok := stmt.Expression.(*ast.CallExpression)
		if !ok {
			t.Fatalf(
				"stmt is not ast.CallExpression. got=%T",
				stmt.Expression,
			)
		}

		// Check the function name
		if !testIdentifier(t, exp.Function, tt.expectedIdentifier) {
			return
		}

		// Check the arguments
		if len(exp.Arguments) != len(tt.expectedArguments) {
			t.Fatalf(
				"wrong length of arguments. got=%d",
				len(exp.Arguments),
			)
		}

		// Loop through the arguments
		for i, arg := range tt.expectedArguments {
			if exp.Arguments[i].String() != arg {
				t.Errorf(
					"argument %d wrong. expected=%q, got=%q",
					i,
					arg,
					exp.Arguments[i].String(),
				)
			}
		}
		t.Logf(
			"TestCallExpressionParameterParsing passed, parsed %d statements",
			len(program.Statements),
		)

		t.Logf("Program: %s", program.String())

	}
}
