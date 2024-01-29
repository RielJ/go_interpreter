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
		{`"Hello" - "World"`, "unknown operator: STRING - STRING"},
		// Error handling
		{`{"name": "Monkey"}[fn(x) { x }];`, "unusable as hash key: FUNCTION"},
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

// Test String literal
func TestStringLiteral(t *testing.T) {
	// String literal
	input := `"Hello World!"`
	// Evaluate
	evaluated := testEval(input)
	// String object
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	// Value
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

// Test string concatenation
func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	// Evaluate
	evaluated := testEval(input)
	// String object
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	// Value
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

// Test builtin functions
func TestBuiltinFunctions(t *testing.T) {
	// Builtin functions
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Len
		{`len("")`, 0},
		// Len
		{`len("four")`, 4},
		// Len
		{`len("hello world")`, 11},
		// Len
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		// Len
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		// Evaluate
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		// Integer
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		// String
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			// Error message
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

// Test array literal
func TestArrayLiterals(t *testing.T) {
	// Array literal
	input := "[1, 2 * 2, 3 + 3]"
	// Evaluate
	evaluated := testEval(input)
	// Array object
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}
	// Length
	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}
	// Element 1
	testIntegerObject(t, result.Elements[0], 1)
	// Element 2
	testIntegerObject(t, result.Elements[1], 4)
	// Element 3
	testIntegerObject(t, result.Elements[2], 6)
}

// Test array index expression
func TestArrayIndexExpressions(t *testing.T) {
	// Array index expression
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Array index expression
		{"[1, 2, 3][0]", 1},
		// Array index expression
		{"[1, 2, 3][1]", 2},
		// Array index expression
		{"[1, 2, 3][2]", 3},
		// Array index expression
		{"let i = 0; [1][i];", 1},
		// Array index expression
		{"[1, 2, 3][1 + 1];", 3},
		// Array index expression
		{"let myArray = [1, 2, 3]; myArray[2];", 3},
		// Array index expression
		{"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", 6},
		// Array index expression
		{"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]", 2},
		// Array index expression
		{"[1, 2, 3][3]", nil},
		// Array index expression
		{"[1, 2, 3][-1]", nil},
	}

	for _, tt := range tests {
		// Evaluate
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		// Integer
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		// Nil
		case nil:
			testNullObject(t, evaluated)
		}
	}
}

// Test hash literal
func TestHashLiterals(t *testing.T) {
	// Hash literal
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`
	// Evaluate
	evaluated := testEval(input)
	// Hash object
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}
	// Pairs
	expected := map[object.HashKey]int64{
		// String
		(&object.String{Value: "one"}).HashKey(): 1,
		// String
		(&object.String{Value: "two"}).HashKey(): 2,
		// String
		(&object.String{Value: "three"}).HashKey(): 3,
		// Integer
		(&object.Integer{Value: 4}).HashKey(): 4,
		// Boolean
		TRUE.HashKey(): 5,
		// Boolean
		FALSE.HashKey(): 6,
	}
	// Length
	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}
	// Pairs
	for expectedKey, expectedValue := range expected {
		// Pair
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}
		// Integer
		testIntegerObject(t, pair.Value, expectedValue)
	}
}

// Test hash index expression
func TestHashIndexExpressions(t *testing.T) {
	// Hash index expression
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Hash index expression
		{`{"foo": 5}["foo"]`, 5},
		// Hash index expression
		{`{"foo": 5}["bar"]`, nil},
		// Hash index expression
		{`let key = "foo"; {"foo": 5}[key]`, 5},
		// Hash index expression
		{`{}["foo"]`, nil},
		// Hash index expression
		{`{5: 5}[5]`, 5},
		// Hash index expression
		{`{true: 5}[true]`, 5},
		// Hash index expression
		{`{false: 5}[false]`, 5},
	}

	for _, tt := range tests {
		// Evaluate
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		// Integer
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		// Nil
		case nil:
			testNullObject(t, evaluated)
		}
	}
}
