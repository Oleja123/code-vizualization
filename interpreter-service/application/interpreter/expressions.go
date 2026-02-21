package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
	runtimeinterfaces "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/interfaces"
)

func (i *Interpreter) executeExpression(expr converter.Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *converter.IntLiteral:
		return e.Value, nil
	case *converter.VariableExpr:
		return i.executeExpression(&VariableExpr{VariableExpr: *e, IsLvalue: false})
	case *converter.ArrayAccessExpr:
		return i.executeExpression(&ArrayAccessExpr{ArrayAccessExpr: *e, IsLvalue: false})

	case *VariableExpr:
		v, err := i.resolveVariable(e.Name)
		if err != nil {
			return nil, err
		}
		if e.IsLvalue {
			return v, nil
		}
		val, ok := v.(*runtime.Variable)
		if !ok {
			return nil, runtimeerrors.NewErrUnexpectedInternalError("no variable as rvalue")
		}
		value, err := val.GetValue()
		if err != nil {
			return nil, err
		}
		return value, nil
	case *ArrayAccessExpr:
		return i.executeArrayAccessExpr(e)
	case *converter.BinaryExpr:
		return i.executeBinaryExpr(e)
	case *converter.AssignmentExpr:
		return i.executeAssignmentExpr(e)
	case *converter.ArrayInitExpr:
		return i.executeArrayInitExpr(e)
	case *converter.UnaryExpr:
		return i.executeUnaryExpr(e)
	case *converter.CallExpr:
		return i.executeCallExpr(e)
	default:
		return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown expression type %T", expr))
	}
}

func (i *Interpreter) executeArrayAccessExpr(a *ArrayAccessExpr) (interface{}, error) {
	arrayLvalue, err := i.convertToLvalue(a.Array)
	if err != nil {
		return nil, err
	}

	array, err := i.executeExpression(arrayLvalue)

	if err != nil {
		return nil, err
	}

	index, err := i.executeExpression(a.Index)

	if err != nil {
		return nil, err
	}

	ind, ok := index.(int)

	if !ok {
		return nil, runtimeerrors.NewErrUnexpectedInternalError("index is not an intenger")
	}

	switch arr := array.(type) {
	case *runtime.Array:
		val, err := arr.GetElement(ind)
		if err != nil {
			return nil, err
		}

		if a.IsLvalue {
			return val, nil
		} else {
			intVal, err := val.GetValue()
			if err != nil {
				return nil, err
			}
			return intVal, nil
		}
	case *runtime.Array2D:
		val, err := arr.GetArray(ind)
		if err != nil {
			return nil, err
		}

		if a.IsLvalue {
			return val, nil
		} else {
			return nil, runtimeerrors.NewErrUnexpectedInternalError("array as rvalue")
		}
	default:
		return nil, runtimeerrors.NewErrUnexpectedInternalError("array type mismatch")
	}
}

func (i *Interpreter) executeBinaryExpr(expr *converter.BinaryExpr) (int, error) {
	leftRvalue, err := i.convertToRvalue(expr.Left)

	if err != nil {
		return 0, err
	}

	leftValRaw, err := i.executeExpression(leftRvalue)
	if err != nil {
		return 0, err
	}

	leftVal, ok := leftValRaw.(int)
	if !ok {
		return 0, runtimeerrors.NewErrUnexpectedInternalError("left expression is not int")
	}

	var result int
	var was bool

	switch expr.Operator {
	case "&&":
		if leftVal == 0 {
			result = 0
			was = true
		}
	case "||":
		if leftVal == 1 {
			result = 1
			was = true
		}
	}

	if was {
		return result, nil
	}

	rightRvalue, err := i.convertToRvalue(expr.Right)

	if err != nil {
		return 0, err
	}

	rightValRaw, err := i.executeExpression(rightRvalue)
	if err != nil {
		return 0, err
	}

	rightVal, ok := rightValRaw.(int)
	if !ok {
		return 0, runtimeerrors.NewErrUnexpectedInternalError("right expression is not int")
	}

	switch expr.Operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			return 0, runtimeerrors.NewErrRuntime("division by zero")
		}
		result = leftVal / rightVal
	case "%":
		if rightVal == 0 {
			return 0, runtimeerrors.NewErrRuntime("modulo by zero")
		}
		result = leftVal % rightVal
	case "==":
		if leftVal == rightVal {
			result = 1
		} else {
			result = 0
		}
	case "!=":
		if leftVal != rightVal {
			result = 1
		} else {
			result = 0
		}
	case "<":
		if leftVal < rightVal {
			result = 1
		} else {
			result = 0
		}
	case "<=":
		if leftVal <= rightVal {
			result = 1
		} else {
			result = 0
		}
	case ">":
		if leftVal > rightVal {
			result = 1
		} else {
			result = 0
		}
	case ">=":
		if leftVal >= rightVal {
			result = 1
		} else {
			result = 0
		}
	case "&&":
		if leftVal != 0 && rightVal != 0 {
			result = 1
		} else {
			result = 0
		}
	case "||":
		if leftVal != 0 || rightVal != 0 {
			result = 1
		} else {
			result = 0
		}
	default:
		return 0, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown binary operator: %s", expr.Operator))
	}

	return result, nil
}

func (i *Interpreter) executeAssignmentExpr(expr *converter.AssignmentExpr) (interface{}, error) {
	leftLvalue, err := i.convertToLvalue(expr.Left)

	if err != nil {
		return 0, err
	}

	leftValRaw, err := i.executeExpression(leftLvalue)
	if err != nil {
		return nil, err
	}

	leftVal, ok := leftValRaw.(runtimeinterfaces.Changeable)
	if !ok {
		return nil, runtimeerrors.NewErrUnexpectedInternalError("left expression is not lvalue")
	}

	rightValRaw, err := i.executeExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	rightVal, ok := rightValRaw.(int)
	if !ok {
		return nil, runtimeerrors.NewErrUnexpectedInternalError("right expression is not int")
	}

	switch expr.Operator {
	case "=":
		leftVal.ChangeValue(rightVal, 0)
	case "+=":
		curVal, err := leftVal.GetValue()
		if err != nil {
			return nil, err
		}
		leftVal.ChangeValue(curVal+rightVal, 0)
	case "-=":
		curVal, err := leftVal.GetValue()
		if err != nil {
			return nil, err
		}
		leftVal.ChangeValue(curVal-rightVal, 0)
	case "*=":
		curVal, err := leftVal.GetValue()
		if err != nil {
			return nil, err
		}
		leftVal.ChangeValue(curVal*rightVal, 0)
	case "/=":
		curVal, err := leftVal.GetValue()
		if err != nil {
			return nil, err
		}
		if rightVal == 0 {
			return nil, runtimeerrors.NewErrRuntime("division by zero")
		}
		leftVal.ChangeValue(curVal/rightVal, 0)
	case "%=":
		curVal, err := leftVal.GetValue()
		if err != nil {
			return nil, err
		}
		if rightVal == 0 {
			return nil, runtimeerrors.NewErrRuntime("division by zero")
		}
		leftVal.ChangeValue(curVal%rightVal, 0)
	default:
		return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown assignment operator: %s", expr.Operator))
	}

	return nil, nil
}

func (i *Interpreter) executeArrayInitExpr(expr *converter.ArrayInitExpr) (interface{}, error) {
	interfaceSlice := make([]interface{}, len(expr.Elements))
	for ind, element := range expr.Elements {
		val, err := i.executeExpression(element)
		if err != nil {
			return nil, err
		}
		interfaceSlice[ind] = val
	}

	switch interfaceSlice[0].(type) {
	case int:
		arrayElementSlice := make([]runtime.ArrayElement, len(expr.Elements))
		for ind, element := range interfaceSlice {
			val, ok := element.(int)
			if !ok {
				valPtr, ok := element.(*int)
				if !ok {
					return nil, runtimeerrors.NewErrUnexpectedInternalError("array element is not int")
				}
				arrayElementSlice[ind] = *runtime.NewArrayElement(valPtr, 0, false)
			} else {
				arrayElementSlice[ind] = *runtime.NewArrayElement(&val, 0, false)
			}
		}
		return arrayElementSlice, nil
	case []runtime.ArrayElement:
		arraySlice := make([]runtime.Array, len(expr.Elements))
		for ind, element := range interfaceSlice {
			val, ok := element.([]runtime.ArrayElement)
			if !ok {
				return nil, runtimeerrors.NewErrUnexpectedInternalError("array2d element is not slice of array elements")
			} else {
				arraySlice[ind] = *runtime.NewArray("", len(val), val, 0, false)
			}
		}
		return arraySlice, nil
	default:
		return nil, runtimeerrors.NewErrUnexpectedInternalError("unexpected array element type")
	}
}

func (i *Interpreter) executeUnaryExpr(expr *converter.UnaryExpr) (int, error) {
	if expr.Operator == "!" || expr.Operator == "+" || expr.Operator == "-" {
		operandRvalue, err := i.convertToRvalue(expr.Operand)
		if err != nil {
			return 0, err
		}

		operand, err := i.executeExpression(operandRvalue)
		if err != nil {
			return 0, err
		}

		v, ok := operand.(int)

		if !ok {
			return 0, runtimeerrors.NewErrUnexpectedInternalError("unary operator ! get non int value")
		}

		switch expr.Operator {
		case "!":
			if v == 0 {
				return 1, nil
			}
			return 0, nil
		case "-":
			return -v, nil
		case "+":
			return v, nil
		default:
			return 0, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown unary operator: %s", expr.Operator))
		}
	} else {
		operandLvalue, err := i.convertToLvalue(expr.Operand)
		if err != nil {
			return 0, err
		}

		operand, err := i.executeExpression(operandLvalue)
		if err != nil {
			return 0, err
		}

		operandVal, ok := operand.(runtimeinterfaces.Changeable)

		if !ok {
			return 0, runtimeerrors.NewErrUnexpectedInternalError("unary operator ++ or -- get non lvalue")
		}

		old, err := operandVal.GetValue()
		if err != nil {
			return 0, err
		}

		switch expr.Operator {
		case "++":
			operandVal.ChangeValue(old+1, 0)
			if expr.IsPostfix {
				return old, nil
			} else {
				return old + 1, nil
			}
		case "--":
			operandVal.ChangeValue(old-1, 0)
			if expr.IsPostfix {
				return old, nil
			} else {
				return old - 1, nil
			}
		default:
			return 0, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown unary operator: %s", expr.Operator))
		}
	}
}

func (i *Interpreter) executeCallExpr(expr *converter.CallExpr) (interface{}, error) {
	declNode, ok := i.Functions[expr.FunctionName]

	if !ok {
		return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown function named: %s", expr.FunctionName))
	}

	if len(expr.Arguments) != len(declNode.Parameters) {
		return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("function %s expects %d arguments, got %d", expr.FunctionName, len(declNode.Parameters), len(expr.Arguments)))
	}

	// Evaluate arguments in the CALLER's context BEFORE pushing new frame
	argumentValues := make([]int, len(expr.Arguments))
	for ind, val := range expr.Arguments {
		val, err := i.executeExpression(val)
		if err != nil {
			return nil, err
		}

		value, ok := val.(int)
		if !ok {
			return nil, runtimeerrors.NewErrUnexpectedInternalError("function argument is not an integer")
		}
		argumentValues[ind] = value
	}

	// NOW push the new frame and initialize parameters with evaluated values
	defer i.CallStack.PopFrame()
	i.CallStack.PushFrame(runtime.NewStackFrame(expr.FunctionName, i.GlobalScope))
	i.CallStack.GetCurrentFrame().EnterScope()

	parameters := make([]*runtime.Variable, len(declNode.Parameters))

	for ind, val := range declNode.Parameters {
		variable := runtime.NewVariable(val.Name, nil, 0, false)
		parameters[ind] = variable

		frame := i.CallStack.GetCurrentFrame()
		frame.GetCurrentScope().Declare(variable)

		// Initialize with the pre-evaluated argument value
		parameters[ind].ChangeValue(argumentValues[ind], 0)
	}

	res, err := i.executeStatement(declNode.Body)
	if err != nil {
		return nil, err
	}

	if res.Signal == SignalReturn {
		if res.Value == nil {
			return nil, nil
		}
		return *res.Value, nil
	}

	return nil, nil
}

func (i *Interpreter) convertToLvalue(expr converter.Expr) (converter.Expr, error) {
	switch e := expr.(type) {
	case *converter.VariableExpr:
		return &VariableExpr{*e, true}, nil
	case *converter.ArrayAccessExpr:
		return &ArrayAccessExpr{*e, true}, nil
	default:
		return nil, runtimeerrors.NewErrUnexpectedInternalError("unexpected lvalue type")
	}
}

func (i *Interpreter) convertToRvalue(expr converter.Expr) (converter.Expr, error) {
	switch e := expr.(type) {
	case *converter.VariableExpr:
		return &VariableExpr{*e, false}, nil
	case *converter.ArrayAccessExpr:
		return &ArrayAccessExpr{*e, false}, nil
	default:
		return e, nil
	}
}
