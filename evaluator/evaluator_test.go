package evaluator

import (
	"testing"

	"github.com/rielj/go-interpreter/lexer"
	"github.com/rielj/go-interpreter/object"
	"github.com/rielj/go-interpreter/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}
	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	// Boolean literals
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		// Boolean object
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

// Boolean object test helper
func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	// Type assertion
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	// Value assertion
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

// Test bang operator
func TestBangOperator(t *testing.T) {
	// Boolean literals
	tests := []struct {
		input    string
		expected bool
	}{
		// True
		{"!true", false},
		// False
		{"!false", true},
		// Bang operator on integer
		{"!5", false},
		// Bang operator on integer
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

// Test if-else expressions
func TestIfElseExpressions(t *testing.T) {
	// Boolean literals
	tests := []struct {
		input    string
		expected interface{}
	}{
		// If statement
		{"if (true) { 10 }", 10},
		// If-else statement
		{"if (false) { 10 } else { 20 }", 20},
		// If-else statement
		{"if (1) { 10 } else { 20 }", 10},
		// If-else statement
		{"if (1 < 2) { 10 } else { 20 }", 10},
		// If-else statement
		{"if (1 > 2) { 10 } else { 20 }", 20},
		// If-else statement
		{"if (1 > 2) { 10 }", nil},
		// If-else statement
		{"if (1 < 2) { 10 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		// Integer object
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			// Null object
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	// Type assertion
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

// Test return statements
func TestReturnStatements(t *testing.T) {
	// Integer literals
	tests := []struct {
		input    string
		expected int64
	}{
		// Return statement
		{"return 10;", 10},
		// Return statement
		{"return 10; 9;", 10},
		// Return statement
		{"return 2 * 5; 9;", 10},
		// Return statement
		{"9; return 2 * 5; 9;", 10},
		// Return statement
		{`
		if (10 > 1) {
			if (10 > 1) {
				return 10;
			}
			return 1;
		}
		`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		// Integer object
		testIntegerObject(t, evaluated, tt.expected)
	}
}

// Test error handling
func TestErrorHandling(t *testing.T) {
	// Error handling
	tests := []struct {
		input          string
		expectedErrMsg string
	}{
		// Error handling
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		// Error handling
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		// Error handling
		{"-true", "unknown operator: -BOOLEAN"},
		// Error handling
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		// Error handling
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		// Error handling
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		// Error handling
		{`
		if (10 > 1) {
			if (10 > 1) {
				return true + false;
			}
			return 1;
		}
		`, "unknown operator: BOOLEAN + BOOLEAN"},
		// Error handling
		{"foobar", "identifier not found: foobar"},
		// Error handling
		// {`"Hello" - "World"`, "unknown operator: STRING - STRING"},
		// Error handling
		// {`{"name": "Monkey"}[fn(x) { x }];`, "unusable as hash key: FUNCTION"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		// Error object
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}
		// Error message
		if errObj.Message != tt.expectedErrMsg {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedErrMsg, errObj.Message)
		}
	}
}

// Test let statements
func TestLetStatements(t *testing.T) {
	// Integer literals
	tests := []struct {
		input    string
		expected int64
	}{
		// Let statement
		{"let a = 5; a;", 5},
		// Let statement
		{"let a = 5 * 5; a;", 25},
		// Let statement
		{"let a = 5; let b = a; b;", 5},
		// Let statement
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

// Test function object
func TestFunctionObject(t *testing.T) {
	// Function literal
	input := "fn(x) { x + 2; };"
	// Evaluate
	evaluated := testEval(input)
	// Function object
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}
	// Parameters
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}
	// Parameter name
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}
	// Function body
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

// Test function application
func TestFunctionApplication(t *testing.T) {
	// Function literal
	tests := []struct {
		input    string
		expected int64
	}{
		// Function application
		{"let identity = fn(x) { x; }; identity(5);", 5},
		// Function application
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		// Function application
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		// Function application
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		// Function application
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		// Function application
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}
