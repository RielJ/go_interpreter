package evaluator

import (
	"fmt"

	"github.com/rielj/go-interpreter/ast"
	"github.com/rielj/go-interpreter/object"
)

var (
	// Singleton objects
	// TRUE
	TRUE = &object.Boolean{Value: true}
	// FALSE
	FALSE = &object.Boolean{Value: false}
	// NULL
	NULL = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node, env)
	// Expressions
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	// Integer
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	// Boolean
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	// Prefix expressions
	case *ast.PrefixExpression:
		// Evaluate the right side of the expression
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		// Evaluate the prefix operator
		return evalPrefixExpression(node.Operator, right)
	// Infix expressions
	case *ast.InfixExpression:
		// Evaluate the left side of the expression
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		// Evaluate the right side of the expression
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		// Evaluate the infix operator
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	// If statements
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	// Return statements
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Let statements
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		// Add the evaluated value to the environment
		// This is how we implement variable bindings

	// Identifiers
	case *ast.Identifier:
		return evalIdentifier(node, env)

	// Function literals
	case *ast.FunctionLiteral:
		// Return a Function object
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}

	// Call expressions
	case *ast.CallExpression:
		// Evaluate the function
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		// Evaluate the arguments
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		// Call the function
		return applyFunction(function, args)
	}

	return nil
}

// Helper function to apply functions
func applyFunction(fn object.Object, args []object.Object) object.Object {
	// Type assertion
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	// Create a new environment for the function
	extendedEnv := extendFunctionEnv(function, args)

	// Evaluate the function body
	evaluated := Eval(function.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
}

// Helper function to extend the environment for a function
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	// Create a new environment
	env := object.NewEnclosedEnvironment(fn.Env)

	// Add the arguments to the environment
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// Helper function to unwrap return values
func unwrapReturnValue(obj object.Object) object.Object {
	// If the object is a ReturnValue object, return the value
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	// Otherwise, return the object
	return obj
}

// Helper function to evaluate expressions
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	// Evaluate each expression
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		// Add the evaluated expression to the result
		result = append(result, evaluated)
	}

	return result
}

// Helper function to evaluate identifiers
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// Get the value of the identifier from the environment
	val, ok := env.Get(node.Value)
	if !ok {
		// If the identifier is not found in the environment, return an error
		return newError("identifier not found: " + node.Value)
	}
	// Otherwise, return the value
	return val
}

// Helper function to evaluate programs
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	// Evaluate each statement in the program
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		// If the result is an Error object, return the error
		case *object.Error:
			return result
		case *object.ReturnValue:
			return result.Value
		}
		// // If the result is a ReturnValue object, return the value
		// if returnValue, ok := result.(*object.ReturnValue); ok {
		// 	return returnValue.Value
		// }
	}

	return result
}

// Helper function to evaluate block statements
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	// Evaluate each statement in the block
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		// If the result is a ReturnValue object, return the value
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// Helper function to convert Go bool to our Boolean object
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// Helper function to evaluate prefix expressions
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	// If the operator is "!", return the result of the bang operator
	case "!":
		return evalBangOperatorExpression(right)
	// If the operator is "-", return the result of the minus operator
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	// If the operator is anything else, return an error
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// Helper function to evaluate infix expressions
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	// If the left and right sides are both integers, evaluate the integer expression
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	// If the left and right sides are both booleans, evaluate the boolean expression
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	// If the left and right sides are not the same type, return NULL
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	// If the left and right sides are not both integers or booleans, return NULL
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// Helper function to evaluate integer infix expressions
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	// Get the values of the left and right sides
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	// Evaluate the integer expression based on the operator
	switch operator {
	// Addition
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	// Subtraction
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	// Multiplication
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	// Division
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	// Less than
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	// Greater than
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	// Equality
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	// Inequality
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	// If the operator is anything else, return NULL
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// Helper function to evaluate bang operator expressions
func evalBangOperatorExpression(right object.Object) object.Object {
	// If the right side is TRUE, return FALSE
	// If the right side is FALSE, return TRUE
	// If the right side is NULL, return TRUE
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	// If the right side is anything else, return FALSE
	default:
		return FALSE
	}
}

// Helper function to evaluate minus prefix operator expressions
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	// If the right side is not an integer, return NULL
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	// Otherwise, return the negative of the integer
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// Helper function to evaluate if expressions
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	// Evaluate the condition
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	// If the condition is TRUE, evaluate the consequence
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
		// If the condition is FALSE, evaluate the alternative
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
		// If there is no alternative, return NULL
	} else {
		return NULL
	}
}

// Helper function to determine if an object is truthy
func isTruthy(obj object.Object) bool {
	// TRUE and FALSE are truthy and falsy, respectively
	// NULL is falsy
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	// If the object is anything else, it is truthy
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	// If the object is not nil and its type is ERROR_OBJ, it is an error
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	// Otherwise, it is not an error
	return false
}
