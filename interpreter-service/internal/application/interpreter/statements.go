package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime/errors"
)

func (i *Interpreter) executeStatement(stmt converter.Stmt) (ExecResult, error) {
	switch t := stmt.(type) {
	case *VariableDecl:
		return i.executeNonFunctionDecl(t)
	case *converter.VariableDecl:
		return i.executeNonFunctionDecl(&VariableDecl{VariableDecl: *t, IsGlobal: false})
	case *converter.FunctionDecl:
		return i.executeFunctionDecl(t)
	case *converter.IfStmt:
		return i.executeIfStmt(t)
	case *converter.WhileStmt:
		return i.executeWhileStmt(t)
	case *converter.DoWhileStmt:
		return i.executeDoWhileStmt(t)
	case *converter.ForStmt:
		return i.executeForStmt(t)
	case *converter.ReturnStmt:
		return i.executeReturnStmt(t)
	case *converter.BlockStmt:
		return i.executeBlockStmt(t)
	case *converter.ExprStmt:
		return i.executeExprStmt(t)
	case *converter.BreakStmt:
		return i.executeBreakStmt(t)
	case *converter.ContinueStmt:
		return i.executeContinueStmt(t)

	default:
		return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unknown statement type %T", stmt))
	}
}

func (i *Interpreter) executeNonFunctionDecl(v *VariableDecl) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(v.Loc.Line)})
	i.addStep()

	switch len(v.VarType.ArraySizes) {
	case 0:
		return i.executeVariableDecl(*v)
	case 1:
		return i.executeArrayDecl(*v)
	case 2:
		return i.executeArray2DDecl(*v)
	default:
		return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("unknown decaration type")
	}
}

func (i *Interpreter) executeVariableDecl(v VariableDecl) (ExecResult, error) {
	var value *int
	if v.InitExpr != nil {
		val, err := i.executeExpression(v.InitExpr)
		if err != nil {
			return NormalResult(), err
		}
		v, ok := val.(int)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
		}
		value = &v
	}

	variable := runtime.NewVariable(v.Name, value, i.currentStepNumber, v.IsGlobal)

	frame := i.CallStack.GetCurrentFrame()
	currentScope := frame.GetCurrentScope()
	currentScope.Declare(variable)

	i.addEvents(events.DeclareVar{Name: v.Name, Value: cloneValue(value), IsGlobal: v.IsGlobal})

	return NormalResult(), nil
}

func (i *Interpreter) executeArrayDecl(v VariableDecl) (ExecResult, error) {
	var value []runtime.ArrayElement
	if v.InitExpr != nil {
		val, err := i.executeExpression(v.InitExpr)
		if err != nil {
			return NormalResult(), err
		}
		v, ok := val.([]runtime.ArrayElement)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("types mismatch: expected []ArrayElement, got %T", val))
		}
		value = v
	}

	variable := runtime.NewArray(v.Name, v.VarType.ArraySizes[0], value, i.currentStepNumber, v.IsGlobal)

	frame := i.CallStack.GetCurrentFrame()
	currentScope := frame.GetCurrentScope()
	currentScope.Declare(variable)

	i.addEvents(events.DeclareArray{Name: v.Name, Value: cloneArrayElements(value), Size: v.VarType.ArraySizes[0], IsGlobal: v.IsGlobal})

	return NormalResult(), nil
}

func (i *Interpreter) executeArray2DDecl(v VariableDecl) (ExecResult, error) {
	var value []runtime.Array
	if v.InitExpr != nil {
		val, err := i.executeExpression(v.InitExpr)
		if err != nil {
			return NormalResult(), err
		}
		v, ok := val.([]runtime.Array)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
		}
		value = v
	}

	variable := runtime.NewArray2D(v.Name, v.VarType.ArraySizes[0], v.VarType.ArraySizes[1], value, i.currentStepNumber, v.IsGlobal)

	frame := i.CallStack.GetCurrentFrame()
	currentScope := frame.GetCurrentScope()
	currentScope.Declare(variable)

	i.addEvents(events.DeclareArray2D{Name: v.Name, Value: cloneArrayRows(value), Size1: v.VarType.ArraySizes[0], Size2: v.VarType.ArraySizes[1], IsGlobal: v.IsGlobal})

	return NormalResult(), nil
}

func (i *Interpreter) executeBlockStmt(b *converter.BlockStmt) (ExecResult, error) {
	frame := i.CallStack.GetCurrentFrame()
	frame.EnterScope()
	i.addEvents(events.EnterScope{})
	defer func() {
		frame.ExitScope()
		i.addEvents(events.ExitScope{})
	}()

	for _, stmt := range b.Statements {
		res, err := i.executeStatement(stmt)
		if err != nil {
			return res, err
		}
		if res.Signal != SignalNormal {
			return res, nil
		}
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeIfStmt(ifStmt *converter.IfStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(ifStmt.Loc.Line)})
	i.addStep()

	cond, err := i.executeExpression(ifStmt.Condition)
	if err != nil {
		return NormalResult(), err
	}

	v, ok := cond.(int)
	if !ok {
		return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
	}

	if v != 0 {
		return i.executeStatement(ifStmt.ThenBlock)
	} else if ifStmt.ElseBlock != nil {
		return i.executeStatement(ifStmt.ElseBlock)
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeReturnStmt(r *converter.ReturnStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(r.Loc.Line)})
	i.addStep()

	var val *int
	if r.Value != nil {
		v, err := i.executeExpression(r.Value)
		if err != nil {
			return NormalResult(), err
		}
		t, ok := v.(int)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
		}
		val = &t
	}
	return ReturnResult(val), nil
}

func (i *Interpreter) executeExprStmt(e *converter.ExprStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(e.Loc.Line)})
	i.addStep()

	if e.Expression == nil {
		return NormalResult(), nil
	}

	_, err := i.executeExpression(e.Expression)
	if err != nil {
		return NormalResult(), err
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeWhileStmt(loop *converter.WhileStmt) (ExecResult, error) {
	for {
		i.addEvents(events.LineChanged{Line: int(loop.Loc.Line)})
		i.addStep()

		condVal, err := i.executeExpression(loop.Condition)
		if err != nil {
			return NormalResult(), err
		}

		v, ok := condVal.(int)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
		}

		if v == 0 {
			break
		}

		res, err := i.executeStatement(loop.Body)
		if err != nil {
			return res, err
		}

		switch res.Signal {
		case SignalBreak:
			return NormalResult(), nil
		case SignalContinue:
			continue
		case SignalReturn:
			return res, nil
		}
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeDoWhileStmt(loop *converter.DoWhileStmt) (ExecResult, error) {
	for {
		res, err := i.executeStatement(loop.Body)
		if err != nil {
			return res, err
		}

		switch res.Signal {
		case SignalBreak:
			return NormalResult(), nil
		case SignalContinue:
			// просто продолжаем на проверку условия
		case SignalReturn:
			return res, nil
		}

		i.addEvents(events.LineChanged{Line: int(loop.Condition.GetLocation().Line)})
		i.addStep()

		condVal, err := i.executeExpression(loop.Condition)
		if err != nil {
			return NormalResult(), err
		}

		v, ok := condVal.(int)
		if !ok {
			return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
		}

		if v == 0 {
			break
		}
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeForStmt(loop *converter.ForStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(loop.Loc.Line)})
	i.addStep()

	frame := i.CallStack.GetCurrentFrame()

	frame.EnterScope()
	i.addEvents(events.EnterScope{})
	defer func() {
		frame.ExitScope()
		i.addEvents(events.ExitScope{})
	}()

	if loop.Init != nil {
		_, err := i.executeStatement(loop.Init)
		if err != nil {
			return NormalResult(), err
		}
	}

	for {
		if loop.Condition != nil {
			condVal, err := i.executeExpression(loop.Condition)
			if err != nil {
				return NormalResult(), err
			}

			v, ok := condVal.(int)
			if !ok {
				return ExecResult{}, runtimeerrors.NewErrUnexpectedInternalError("types mismatch")
			}

			if v == 0 {
				break
			}
		}

		res, err := i.executeStatement(loop.Body)
		if err != nil {
			return res, err
		}

		switch res.Signal {
		case SignalBreak:
			return NormalResult(), nil
		case SignalContinue:
		case SignalReturn:
			return res, nil
		}

		if loop.Post != nil {
			_, err := i.executeStatement(loop.Post)
			if err != nil {
				return NormalResult(), err
			}
		}
	}

	return NormalResult(), nil
}

func (i *Interpreter) executeBreakStmt(b *converter.BreakStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(b.Loc.Line)})
	i.addStep()
	return BreakResult(), nil
}

func (i *Interpreter) executeContinueStmt(c *converter.ContinueStmt) (ExecResult, error) {
	i.addEvents(events.LineChanged{Line: int(c.Loc.Line)})
	i.addStep()
	return ContinueResult(), nil
}

func (i *Interpreter) executeFunctionDecl(f *converter.FunctionDecl) (ExecResult, error) {
	if _, exists := i.Functions[f.Name]; exists {
		return NormalResult(), runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("functions with the same name: %s", f.Name))
	}

	i.Functions[f.Name] = f
	return NormalResult(), nil
}

// cloneArrayElements создает deep-copy массива элементов для сохранения снимка значений
func cloneArrayElements(src []runtime.ArrayElement) []runtime.ArrayElement {
	if src == nil {
		return nil
	}
	dst := make([]runtime.ArrayElement, len(src))
	for i := range src {
		dst[i].StepChanged = src[i].StepChanged
		if src[i].Value != nil {
			val := *src[i].Value
			dst[i].Value = &val
		}
	}
	return dst
}

// cloneArrayRows создает deep-copy 2D массива для сохранения снимка значений
func cloneArrayRows(src []runtime.Array) []runtime.Array {
	if src == nil {
		return nil
	}
	dst := make([]runtime.Array, len(src))
	for i := range src {
		dst[i] = runtime.Array{
			Name:   src[i].Name,
			Size:   src[i].Size,
			Values: cloneArrayElements(src[i].Values),
		}
	}
	return dst
}

// cloneValue создает копию значения указателя для сохранения снимка
func cloneValue(src *int) *int {
	if src == nil {
		return nil
	}
	val := *src
	return &val
}
