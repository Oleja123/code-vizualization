package snapshot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/interpreter-service/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime"
)

func TestNewSnapshot(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	assert.NotNil(t, sn.CallStack)
	assert.NotNil(t, sn.GlobalScope)
	assert.Equal(t, globalScope, sn.GlobalScope)
	assert.Equal(t, 0, sn.Line)
	assert.NotNil(t, sn.CallStack.GetCurrentFrame())
}

func TestSnapshotApplyDeclareVar(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	event := events.DeclareVar{
		Name:  "x",
		Value: nil,
	}

	err := sn.Apply(event, 0)
	require.NoError(t, err)

	variable, ok := sn.GetVariable("x")
	assert.True(t, ok)
	assert.NotNil(t, variable)
	assert.Equal(t, "x", variable.Name)
}

func TestSnapshotApplyDeclareVarWithInitialValue(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	val := 42
	event := events.DeclareVar{
		Name:  "x",
		Value: &val,
	}

	err := sn.Apply(event, 0)
	require.NoError(t, err)

	variable, ok := sn.GetVariable("x")
	assert.True(t, ok)
	value, err := variable.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestSnapshotApplyDeclareArray(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	event := events.DeclareArray{
		Name:  "arr",
		Value: []int{1, 2, 3},
		Size:  3,
	}

	err := sn.Apply(event, 0)
	require.NoError(t, err)

	arr, ok := sn.GetArray("arr")
	assert.True(t, ok)
	assert.NotNil(t, arr)
	assert.Equal(t, "arr", arr.Name)
	assert.Equal(t, 3, arr.Size)
}

func TestSnapshotApplyDeclareArray2D(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	event := events.DeclareArray2D{
		Name:  "matrix",
		Value: [][]int{{1, 2}, {3, 4}},
		Size1: 2,
		Size2: 2,
	}

	err := sn.Apply(event, 0)
	require.NoError(t, err)

	arr2d, ok := sn.GetArray2D("matrix")
	assert.True(t, ok)
	assert.NotNil(t, arr2d)
	assert.Equal(t, "matrix", arr2d.Name)
	assert.Equal(t, 2, arr2d.Size1)
	assert.Equal(t, 2, arr2d.Size2)
}

func TestSnapshotApplyEnterExitScope(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Declare variable in global scope
	declareEvent := events.DeclareVar{
		Name:  "global_var",
		Value: nil,
	}
	err := sn.Apply(declareEvent, 0)
	require.NoError(t, err)

	// Enter new scope
	enterEvent := events.EnterScope{}
	err = sn.Apply(enterEvent, 0)
	require.NoError(t, err)

	frame := sn.GetCurrentFrame()
	assert.NotNil(t, frame)
	assert.Equal(t, 2, len(frame.Scopes))

	// Exit scope
	exitEvent := events.ExitScope{}
	err = sn.Apply(exitEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 1, len(frame.Scopes))
}

func TestSnapshotApplyVarChanged(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Declare variable
	declareEvent := events.DeclareVar{
		Name:  "x",
		Value: nil,
	}
	err := sn.Apply(declareEvent, 0)
	require.NoError(t, err)

	// Change variable
	changeEvent := events.VarChanged{
		Name:  "x",
		Value: 100,
	}
	err = sn.Apply(changeEvent, 0)
	require.NoError(t, err)

	variable, ok := sn.GetVariable("x")
	assert.True(t, ok)
	value, err := variable.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 100, value)
}

func TestSnapshotApplyVarChangedNotFound(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	changeEvent := events.VarChanged{
		Name:  "nonexistent",
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
}

func TestSnapshotApplyArrayElementChanged(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Declare array
	declareEvent := events.DeclareArray{
		Name:  "arr",
		Value: []int{0, 0, 0},
		Size:  3,
	}
	err := sn.Apply(declareEvent, 0)
	require.NoError(t, err)

	// Change array element
	changeEvent := events.ArrayElementChanged{
		Name:  "arr",
		Ind:   1,
		Value: 42,
	}
	err = sn.Apply(changeEvent, 0)
	require.NoError(t, err)

	arr, ok := sn.GetArray("arr")
	assert.True(t, ok)
	value, err := arr.GetElement(1)
	assert.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestSnapshotApplyArray2DElementChanged(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Declare 2D array
	declareEvent := events.DeclareArray2D{
		Name:  "matrix",
		Value: [][]int{{0, 0}, {0, 0}},
		Size1: 2,
		Size2: 2,
	}
	err := sn.Apply(declareEvent, 0)
	require.NoError(t, err)

	// Change element
	changeEvent := events.Array2DElementChanged{
		Name:  "matrix",
		Ind1:  1,
		Ind2:  1,
		Value: 99,
	}
	err = sn.Apply(changeEvent, 0)
	require.NoError(t, err)

	arr2d, ok := sn.GetArray2D("matrix")
	assert.True(t, ok)
	value, err := arr2d.GetElement(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)
}

func TestSnapshotApplyFunctionCallAndReturn(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	initialFramesCount := sn.GetFramesCount()
	assert.Equal(t, 1, initialFramesCount)

	// Function call
	callEvent := events.FunctionCall{
		Name: "myFunc",
	}
	err := sn.Apply(callEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 2, sn.GetFramesCount())
	frame := sn.GetCurrentFrame()
	assert.NotNil(t, frame)
	assert.Equal(t, "myFunc", frame.FuncName)

	// Function return
	retVal := 55
	returnEvent := events.FunctionReturn{
		Name:        "myFunc",
		ReturnValue: &retVal,
	}
	err = sn.Apply(returnEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 1, sn.GetFramesCount())
}

func TestSnapshotApplyFunctionReturnVoid(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	callEvent := events.FunctionCall{
		Name: "voidFunc",
	}
	err := sn.Apply(callEvent, 0)
	require.NoError(t, err)

	// Return without value
	returnEvent := events.FunctionReturn{
		Name:        "voidFunc",
		ReturnValue: nil,
	}
	err = sn.Apply(returnEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 1, sn.GetFramesCount())
}

func TestSnapshotApplyLineChanged(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	assert.Equal(t, 0, sn.GetCurrentLine())

	lineEvent := events.LineChanged{
		Line: 42,
	}
	err := sn.Apply(lineEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 42, sn.GetCurrentLine())
}

func TestSnapshotReset(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Add some state
	declareEvent := events.DeclareVar{
		Name:  "x",
		Value: nil,
	}
	err := sn.Apply(declareEvent, 0)
	require.NoError(t, err)

	lineEvent := events.LineChanged{
		Line: 100,
	}
	err = sn.Apply(lineEvent, 0)
	require.NoError(t, err)

	// Verify state is set
	assert.Equal(t, 100, sn.GetCurrentLine())

	// Reset
	sn.Reset()

	// After reset, CallStack should be fresh, Line and Step should be 0
	assert.Equal(t, 0, sn.GetCurrentLine())
	assert.Equal(t, 1, sn.GetFramesCount())
	// Note: Global scope retains declarations (this is correct behavior)
	// NewSnapshot should be created with a fresh scope for a clean slate
}

func TestSnapshotComplexScenario(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Declare global variables
	events1 := events.DeclareVar{Name: "x", Value: nil}
	assert.NoError(t, sn.Apply(events1, 0))

	declareArrayEvent := events.DeclareArray{
		Name:  "arr",
		Value: []int{10, 20, 30},
		Size:  3,
	}
	assert.NoError(t, sn.Apply(declareArrayEvent, 0))

	// Change variable
	varChangeEvent := events.VarChanged{Name: "x", Value: 5}
	assert.NoError(t, sn.Apply(varChangeEvent, 0))

	// Function call
	callEvent := events.FunctionCall{Name: "process"}
	assert.NoError(t, sn.Apply(callEvent, 0))

	// Enter scope
	enterScopeEvent := events.EnterScope{}
	assert.NoError(t, sn.Apply(enterScopeEvent, 0))

	// Declare local variable
	localVarEvent := events.DeclareVar{Name: "y", Value: nil}
	assert.NoError(t, sn.Apply(localVarEvent, 0))

	// Change array element
	arrChangeEvent := events.ArrayElementChanged{Name: "arr", Ind: 0, Value: 100}
	assert.NoError(t, sn.Apply(arrChangeEvent, 0))

	// Exit scope
	exitScopeEvent := events.ExitScope{}
	assert.NoError(t, sn.Apply(exitScopeEvent, 0))

	// Return from function
	returnVal := 42
	returnEvent := events.FunctionReturn{Name: "process", ReturnValue: &returnVal}
	assert.NoError(t, sn.Apply(returnEvent, 0))

	// Add final step for line change

	// Verify final state
	assert.Equal(t, 1, sn.GetFramesCount())

	x, ok := sn.GetVariable("x")
	assert.True(t, ok)
	xVal, err := x.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 5, xVal)

	arr, ok := sn.GetArray("arr")
	assert.True(t, ok)
	elemVal, err := arr.GetElement(0)
	assert.NoError(t, err)
	assert.Equal(t, 100, elemVal)
}

func TestSnapshotGetArrayNotFound(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	arr, ok := sn.GetArray("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, arr)
}

func TestSnapshotGetArray2DNotFound(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	arr2d, ok := sn.GetArray2D("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, arr2d)
}

func TestSnapshotGetVariableNotFound(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	variable, ok := sn.GetVariable("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, variable)
}

func TestSnapshotMultipleFunctionCalls(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// First function call
	call1 := events.FunctionCall{Name: "func1"}
	assert.NoError(t, sn.Apply(call1, 0))
	assert.Equal(t, 2, sn.GetFramesCount())

	// Nested call
	call2 := events.FunctionCall{Name: "func2"}
	assert.NoError(t, sn.Apply(call2, 0))
	assert.Equal(t, 3, sn.GetFramesCount())

	frame := sn.GetCurrentFrame()
	assert.NotNil(t, frame)
	assert.Equal(t, "func2", frame.FuncName)

	// Return from nested call
	return2 := events.FunctionReturn{Name: "func2", ReturnValue: nil}
	assert.NoError(t, sn.Apply(return2, 0))
	assert.Equal(t, 2, sn.GetFramesCount())

	frame = sn.GetCurrentFrame()
	assert.NotNil(t, frame)
	assert.Equal(t, "func1", frame.FuncName)

	// Return from first call
	return1 := events.FunctionReturn{Name: "func1", ReturnValue: nil}
	assert.NoError(t, sn.Apply(return1, 0))
	assert.Equal(t, 1, sn.GetFramesCount())
}

// =================== Реальные программы ===================

// Тест: Факториал (рекурсия)
func TestScenarioFactorial(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Объявление глобальных переменных
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.LineChanged{Line: 1}, 0)

	// n = 5
	sn.Apply(events.VarChanged{Name: "n", Value: 5}, 0)
	sn.Apply(events.LineChanged{Line: 2}, 0)

	// Объявление результата
	sn.Apply(events.DeclareVar{Name: "result", Value: nil}, 0)
	sn.Apply(events.LineChanged{Line: 3}, 0)

	// result = factorial(5)
	sn.Apply(events.FunctionCall{Name: "factorial"}, 0)
	sn.Apply(events.LineChanged{Line: 4}, 0)

	// =================== factorial(5) ===================
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.LineChanged{Line: 10}, 0)

	sn.Apply(events.VarChanged{Name: "n", Value: 5}, 0)

	// if (n == 1) return 1
	sn.Apply(events.LineChanged{Line: 11}, 0)

	// Рекурсивный вызов factorial(4)
	sn.Apply(events.FunctionCall{Name: "factorial"}, 0)
	sn.Apply(events.LineChanged{Line: 13}, 0)

	// =================== factorial(4) ===================
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "n", Value: 4}, 0)

	// Ещё один рекурсивный вызов factorial(3)
	sn.Apply(events.FunctionCall{Name: "factorial"}, 0)

	// =================== factorial(3) ===================
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "n", Value: 3}, 0)

	// Рекурсивный вызов factorial(2)
	sn.Apply(events.FunctionCall{Name: "factorial"}, 0)

	// =================== factorial(2) ===================
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "n", Value: 2}, 0)

	// Рекурсивный вызов factorial(1)
	sn.Apply(events.FunctionCall{Name: "factorial"}, 0)

	// =================== factorial(1) - базовый случай ===================
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "n", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "n", Value: 1}, 0)

	// return 1
	retVal1 := 1
	sn.Apply(events.FunctionReturn{Name: "factorial", ReturnValue: &retVal1}, 0)

	// =================== Возврат из factorial(2) ===================
	sn.Apply(events.ExitScope{}, 0)
	retVal2 := 2 // 2 * 1
	sn.Apply(events.FunctionReturn{Name: "factorial", ReturnValue: &retVal2}, 0)

	// =================== Возврат из factorial(3) ===================
	sn.Apply(events.ExitScope{}, 0)
	retVal6 := 6 // 3 * 2
	sn.Apply(events.FunctionReturn{Name: "factorial", ReturnValue: &retVal6}, 0)

	// =================== Возврат из factorial(4) ===================
	sn.Apply(events.ExitScope{}, 0)
	retVal24 := 24 // 4 * 6
	sn.Apply(events.FunctionReturn{Name: "factorial", ReturnValue: &retVal24}, 0)

	// =================== Возврат из factorial(5) ===================
	sn.Apply(events.ExitScope{}, 0)
	retVal120 := 120 // 5 * 24
	sn.Apply(events.FunctionReturn{Name: "factorial", ReturnValue: &retVal120}, 0)

	// Присвоение результата
	sn.Apply(events.VarChanged{Name: "result", Value: 120}, 0)
	sn.Apply(events.LineChanged{Line: 5}, 0)

	// Проверяем финальное состояние
	result, ok := sn.GetVariable("result")
	assert.True(t, ok)
	resultVal, err := result.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 120, resultVal)
	assert.Equal(t, 1, sn.GetFramesCount()) // Вернулись в main frame
}

// Тест: Сортировка массива (bubble sort)
func TestScenarioBubbleSort(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Объявление и инициализация массива
	sn.Apply(events.DeclareArray{
		Name:  "arr",
		Value: []int{5, 2, 8, 1, 9},
		Size:  5,
	}, 0)
	sn.Apply(events.LineChanged{Line: 1}, 0)

	// Переменные для сортировки
	sn.Apply(events.DeclareVar{Name: "i", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "j", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "temp", Value: nil}, 0)
	sn.Apply(events.LineChanged{Line: 2}, 0)

	// Bubble sort: две вложенные ячейки
	// i = 0
	sn.Apply(events.VarChanged{Name: "i", Value: 0}, 0)
	sn.Apply(events.LineChanged{Line: 4}, 0)

	// j = 0, arr[0]=5 > arr[1]=2 -> swap
	sn.Apply(events.VarChanged{Name: "j", Value: 0}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 0, Value: 2}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 1, Value: 5}, 0)

	// j = 1, arr[1]=5 < arr[2]=8 -> no swap
	sn.Apply(events.VarChanged{Name: "j", Value: 1}, 0)

	// j = 2, arr[2]=8 > arr[3]=1 -> swap
	sn.Apply(events.VarChanged{Name: "j", Value: 2}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 2, Value: 1}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 3, Value: 8}, 0)

	// j = 3, arr[3]=8 < arr[4]=9 -> no swap
	sn.Apply(events.VarChanged{Name: "j", Value: 3}, 0)

	// Второй проход (i = 1)
	sn.Apply(events.VarChanged{Name: "i", Value: 1}, 0)
	sn.Apply(events.LineChanged{Line: 4}, 0)

	// j = 0, arr[0]=2 < arr[1]=5 -> no swap
	sn.Apply(events.VarChanged{Name: "j", Value: 0}, 0)

	// j = 1, arr[1]=5 > arr[2]=1 -> swap
	sn.Apply(events.VarChanged{Name: "j", Value: 1}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 1, Value: 1}, 0)
	sn.Apply(events.ArrayElementChanged{Name: "arr", Ind: 2, Value: 5}, 0)

	// j = 2, arr[2]=5 < arr[3]=8 -> no swap
	sn.Apply(events.VarChanged{Name: "j", Value: 2}, 0)

	// j = 3, arr[3]=8 < arr[4]=9 -> no swap
	sn.Apply(events.VarChanged{Name: "j", Value: 3}, 0)

	// После нескольких итераций проверяем состояние
	arr, ok := sn.GetArray("arr")
	assert.True(t, ok)

	// Проверяем, что массив содержит правильные элементы
	val0, _ := arr.GetElement(0)
	val1, _ := arr.GetElement(1)
	val2, _ := arr.GetElement(2)
	assert.Equal(t, 2, val0)
	assert.Equal(t, 1, val1)
	assert.Equal(t, 5, val2)
}

// Тест: Поиск в матрице (2D массив)
func TestScenarioMatrix2DSearch(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Объявление матрицы 3x3
	sn.Apply(events.DeclareArray2D{
		Name:  "matrix",
		Value: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
		Size1: 3,
		Size2: 3,
	}, 0)
	sn.Apply(events.LineChanged{Line: 1}, 0)

	// Переменные для поиска
	sn.Apply(events.DeclareVar{Name: "target", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "found", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "i", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "j", Value: nil}, 0)
	sn.Apply(events.LineChanged{Line: 2}, 0)

	// target = 5
	sn.Apply(events.VarChanged{Name: "target", Value: 5}, 0)
	sn.Apply(events.VarChanged{Name: "found", Value: 0}, 0)

	// Поиск по матрице
	// i = 0
	sn.Apply(events.VarChanged{Name: "i", Value: 0}, 0)
	sn.Apply(events.LineChanged{Line: 5}, 0)

	// j = 0, matrix[0][0] = 1 != 5
	sn.Apply(events.VarChanged{Name: "j", Value: 0}, 0)

	// j = 1, matrix[0][1] = 2 != 5
	sn.Apply(events.VarChanged{Name: "j", Value: 1}, 0)

	// j = 2, matrix[0][2] = 3 != 5
	sn.Apply(events.VarChanged{Name: "j", Value: 2}, 0)

	// i = 1
	sn.Apply(events.VarChanged{Name: "i", Value: 1}, 0)

	// j = 0, matrix[1][0] = 4 != 5
	sn.Apply(events.VarChanged{Name: "j", Value: 0}, 0)

	// j = 1, matrix[1][1] = 5 == 5 -> НАЙДЕНО!
	sn.Apply(events.VarChanged{Name: "j", Value: 1}, 0)
	sn.Apply(events.VarChanged{Name: "found", Value: 1}, 0)
	sn.Apply(events.LineChanged{Line: 8}, 0)

	// Проверяем финальное состояние
	found, ok := sn.GetVariable("found")
	assert.True(t, ok)
	foundVal, err := found.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 1, foundVal) // Найдено

	matrix, ok := sn.GetArray2D("matrix")
	assert.True(t, ok)
	matVal, err := matrix.GetElement(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 5, matVal) // Правильный элемент матрицы
}

// Тест: Вычисление суммы и среднего (с тремя вложенными функциями)
func TestScenarioNotationCalculations(t *testing.T) {
	globalScope := runtime.NewScope(nil)
	sn := NewSnapshot(globalScope)

	// Массив чисел
	sn.Apply(events.DeclareArray{
		Name:  "numbers",
		Value: []int{10, 20, 30, 40},
		Size:  4,
	}, 0)
	sn.Apply(events.LineChanged{Line: 1}, 0)

	// sum, avg, count
	sn.Apply(events.DeclareVar{Name: "sum", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "avg", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "count", Value: nil}, 0)

	// sum = calculateSum(numbers, 4)
	sn.Apply(events.FunctionCall{Name: "calculateSum"}, 0)

	// входим в calculateSum
	sn.Apply(events.EnterScope{}, 0)
	sn.Apply(events.DeclareVar{Name: "total", Value: nil}, 0)
	sn.Apply(events.DeclareVar{Name: "i", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "total", Value: 0}, 0)

	// Цикл суммирования
	sn.Apply(events.VarChanged{Name: "i", Value: 0}, 0)
	sn.Apply(events.VarChanged{Name: "total", Value: 10}, 0)

	sn.Apply(events.VarChanged{Name: "i", Value: 1}, 0)
	sn.Apply(events.VarChanged{Name: "total", Value: 30}, 0)

	sn.Apply(events.VarChanged{Name: "i", Value: 2}, 0)
	sn.Apply(events.VarChanged{Name: "total", Value: 60}, 0)

	sn.Apply(events.VarChanged{Name: "i", Value: 3}, 0)
	sn.Apply(events.VarChanged{Name: "total", Value: 100}, 0)

	// Возврат из calculateSum
	retSum := 100
	sn.Apply(events.FunctionReturn{Name: "calculateSum", ReturnValue: &retSum}, 0)
	sn.Apply(events.VarChanged{Name: "sum", Value: 100}, 0)

	// count = 4
	sn.Apply(events.VarChanged{Name: "count", Value: 4}, 0)

	// avg = calculateAverage(sum, count)
	sn.Apply(events.FunctionCall{Name: "calculateAverage"}, 0)

	// входим в calculateAverage
	sn.Apply(events.EnterScope{}, 0)

	// Возврат из calculateAverage
	retAvg := 25
	sn.Apply(events.FunctionReturn{Name: "calculateAverage", ReturnValue: &retAvg}, 0)
	sn.Apply(events.VarChanged{Name: "avg", Value: 25}, 0)

	// Проверяем финальное состояние
	sum, ok := sn.GetVariable("sum")
	assert.True(t, ok)
	sumVal, _ := sum.GetValue()
	assert.Equal(t, 100, sumVal)

	avg, ok := sn.GetVariable("avg")
	assert.True(t, ok)
	avgVal, _ := avg.GetValue()
	assert.Equal(t, 25, avgVal)

	count, ok := sn.GetVariable("count")
	assert.True(t, ok)
	countVal, _ := count.GetValue()
	assert.Equal(t, 4, countVal)

	assert.Equal(t, 1, sn.GetFramesCount()) // Вернулись в main
}
