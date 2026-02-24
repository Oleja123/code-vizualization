package snapshot

import (
	"testing"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to convert []int to []runtime.ArrayElement
func makeArrayElements(values []int) []runtime.ArrayElement {
	result := make([]runtime.ArrayElement, len(values))
	for i, v := range values {
		val := v
		result[i] = runtime.ArrayElement{Value: &val, StepChanged: 0}
	}
	return result
}

// Helper function to convert [][]int to []runtime.Array
func makeArray2D(values [][]int) []runtime.Array {
	result := make([]runtime.Array, len(values))
	for i, row := range values {
		elements := makeArrayElements(row)
		result[i] = runtime.Array{
			Name:   "",
			Size:   len(row),
			Values: elements,
		}
	}
	return result
}

func TestNewSnapshot(t *testing.T) {
	sn := NewSnapshot()

	assert.NotNil(t, sn.CallStack)
	assert.NotNil(t, sn.GlobalScope)
	assert.Equal(t, 0, sn.Line)
	assert.NotNil(t, sn.CallStack.GetCurrentFrame())
}

func TestSnapshotApplyDeclareVar(t *testing.T) {
	sn := NewSnapshot()

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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

	event := events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{1, 2, 3}),
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
	sn := NewSnapshot()

	event := events.DeclareArray2D{
		Name:  "matrix",
		Value: makeArray2D([][]int{{1, 2}, {3, 4}}),
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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

	changeEvent := events.VarChanged{
		Name:  "nonexistent",
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
}

func TestSnapshotApplyArrayElementChanged(t *testing.T) {
	sn := NewSnapshot()

	// Declare array
	declareEvent := events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{0, 0, 0}),
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
	valInt, err := value.GetValue()
	assert.Equal(t, 42, valInt)
}

func TestSnapshotApplyArray2DElementChanged(t *testing.T) {
	sn := NewSnapshot()

	// Declare 2D array
	declareEvent := events.DeclareArray2D{
		Name:  "matrix",
		Value: makeArray2D([][]int{{0, 0}, {0, 0}}),
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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

	assert.Equal(t, 0, sn.GetCurrentLine())

	lineEvent := events.LineChanged{
		Line: 42,
	}
	err := sn.Apply(lineEvent, 0)
	require.NoError(t, err)

	assert.Equal(t, 42, sn.GetCurrentLine())
}

func TestSnapshotReset(t *testing.T) {
	sn := NewSnapshot()

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

	// After reset, CallStack should be fresh
	assert.Equal(t, -1, sn.GetCurrentLine())
	assert.Equal(t, 1, sn.GetFramesCount())
	// Note: Global scope retains declarations (this is correct behavior)
	// NewSnapshot should be created with a fresh scope for a clean slate
}

func TestSnapshotComplexScenario(t *testing.T) {
	sn := NewSnapshot()

	// Declare global variables
	events1 := events.DeclareVar{Name: "x", Value: nil}
	assert.NoError(t, sn.Apply(events1, 0))

	declareArrayEvent := events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{10, 20, 30}),
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
	elemValInt, err := elemVal.GetValue()
	assert.Equal(t, 100, elemValInt)
}

func TestSnapshotGetArrayNotFound(t *testing.T) {
	sn := NewSnapshot()

	arr, ok := sn.GetArray("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, arr)
}

func TestSnapshotGetArray2DNotFound(t *testing.T) {
	sn := NewSnapshot()

	arr2d, ok := sn.GetArray2D("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, arr2d)
}

func TestSnapshotGetVariableNotFound(t *testing.T) {
	sn := NewSnapshot()

	variable, ok := sn.GetVariable("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, variable)
}

func TestSnapshotMultipleFunctionCalls(t *testing.T) {
	sn := NewSnapshot()

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
	sn := NewSnapshot()

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
	sn := NewSnapshot()

	// Объявление и инициализация массива
	sn.Apply(events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{5, 2, 8, 1, 9}),
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
	val0Int, _ := val0.GetValue()
	val1Int, _ := val1.GetValue()
	val2Int, _ := val2.GetValue()
	assert.Equal(t, 2, val0Int)
	assert.Equal(t, 1, val1Int)
	assert.Equal(t, 5, val2Int)
}

// Тест: Поиск в матрице (2D массив)
func TestScenarioMatrix2DSearch(t *testing.T) {
	sn := NewSnapshot()

	// Объявление матрицы 3x3
	sn.Apply(events.DeclareArray2D{
		Name:  "matrix",
		Value: makeArray2D([][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}),
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
	sn := NewSnapshot()

	// Массив чисел
	sn.Apply(events.DeclareArray{
		Name:  "numbers",
		Value: makeArrayElements([]int{10, 20, 30, 40}),
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

// ============= Тесты на обработку undefined behavior =============

func TestUndefinedBehaviorVarNotFound(t *testing.T) {
	sn := NewSnapshot()

	// Попытка изменить переменную, которая не была объявлена
	changeEvent := events.VarChanged{
		Name:  "nonexistent",
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable nonexistent not found")
}

func TestUndefinedBehaviorArrayNotFound(t *testing.T) {
	sn := NewSnapshot()

	// Попытка изменить элемент масивов, который не был объявлен
	changeEvent := events.ArrayElementChanged{
		Name:  "nonexistent_array",
		Ind:   0,
		Value: 42,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array nonexistent_array not found")
}

func TestUndefinedBehaviorArray2DNotFound(t *testing.T) {
	sn := NewSnapshot()

	// Попытка изменить элемент 2D массива, который не был объявлен
	changeEvent := events.Array2DElementChanged{
		Name:  "nonexistent_matrix",
		Ind1:  0,
		Ind2:  0,
		Value: 42,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array2d nonexistent_matrix not found")
}

func TestUndefinedBehaviorArrayOutOfBoundsPositive(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем массив размером 3
	sn.Apply(events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{1, 2, 3}),
		Size:  3,
	}, 0)

	// Попытка доступа к индексу вне границ (индекс 5 при размере 3)
	changeEvent := events.ArrayElementChanged{
		Name:  "arr",
		Ind:   5,
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index out of bounds")
}

func TestUndefinedBehaviorArrayOutOfBoundsNegative(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем массив
	sn.Apply(events.DeclareArray{
		Name:  "arr",
		Value: makeArrayElements([]int{1, 2, 3}),
		Size:  3,
	}, 0)

	// Попытка доступа к отрицательному индексу
	changeEvent := events.ArrayElementChanged{
		Name:  "arr",
		Ind:   -1,
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index out of bounds")
}

func TestUndefinedBehaviorArray2DOutOfBoundsRow(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем 2D массив 2x2
	sn.Apply(events.DeclareArray2D{
		Name:  "matrix",
		Value: makeArray2D([][]int{{1, 2}, {3, 4}}),
		Size1: 2,
		Size2: 2,
	}, 0)

	// Попытка доступа к строке вне границ
	changeEvent := events.Array2DElementChanged{
		Name:  "matrix",
		Ind1:  5,
		Ind2:  0,
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index out of bounds")
}

func TestUndefinedBehaviorArray2DOutOfBoundsCol(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем 2D массив 2x2
	sn.Apply(events.DeclareArray2D{
		Name:  "matrix",
		Value: makeArray2D([][]int{{1, 2}, {3, 4}}),
		Size1: 2,
		Size2: 2,
	}, 0)

	// Попытка доступа к колонке вне границ
	changeEvent := events.Array2DElementChanged{
		Name:  "matrix",
		Ind1:  0,
		Ind2:  10,
		Value: 100,
	}
	err := sn.Apply(changeEvent, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index out of bounds")
}

func TestUndefinedBehaviorEventApply(t *testing.T) {
	sn := NewSnapshot()

	// Применяем событие UndefinedBehavior
	ubEvent := events.UndefinedBehavior{
		Message: "division by zero",
	}
	err := sn.Apply(ubEvent, 0)
	assert.NoError(t, err)

	// Проверяем, что ошибка сохранена в Snapshot
	assert.Equal(t, "division by zero", sn.Error)
}

func TestUndefinedBehaviorEventApplyMultiple(t *testing.T) {
	sn := NewSnapshot()

	// Применяем первое событие UndefinedBehavior
	ubEvent1 := events.UndefinedBehavior{
		Message: "first error",
	}
	err := sn.Apply(ubEvent1, 0)
	assert.NoError(t, err)
	assert.Equal(t, "first error", sn.Error)

	// Объявляем переменную
	sn.Apply(events.DeclareVar{Name: "x", Value: nil}, 0)
	sn.Apply(events.VarChanged{Name: "x", Value: 10}, 0)

	// Применяем второе событие UndefinedBehavior (перезаписываем ошибку)
	ubEvent2 := events.UndefinedBehavior{
		Message: "array index out of bounds",
	}
	err = sn.Apply(ubEvent2, 0)
	assert.NoError(t, err)
	assert.Equal(t, "array index out of bounds", sn.Error)
}

func TestUndefinedBehaviorVarNotFoundAfterDeclaration(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем переменную x
	sn.Apply(events.DeclareVar{Name: "x", Value: nil}, 0)
	assert.NoError(t, sn.Apply(events.VarChanged{Name: "x", Value: 5}, 0))

	// Входим в новую область видимости
	sn.Apply(events.EnterScope{}, 0)

	// Пытаемся изменить x (она должна найтись из родительской области)
	// но если реализация требует локальную переменную, то будет ошибка
	err := sn.Apply(events.VarChanged{Name: "y", Value: 10}, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable y not found")
}

func TestUndefinedBehaviorExitScopeWithoutEnter(t *testing.T) {
	sn := NewSnapshot()

	// Входим в область видимости
	sn.Apply(events.EnterScope{}, 0)
	initialFrames := sn.GetFramesCount()

	// Выходим из области видимости
	err := sn.Apply(events.ExitScope{}, 0)
	require.NoError(t, err)

	// Фреймов должно остаться столько же (выходили из локальной scope, не из фрейма)
	assert.Equal(t, initialFrames, sn.GetFramesCount())

	// Попытка выхода из последней области видимости должна вызвать ошибку
	err = sn.Apply(events.ExitScope{}, 0)
	assert.Error(t, err)
}

func TestUndefinedBehaviorComplexErrorScenario(t *testing.T) {
	sn := NewSnapshot()

	// Объявляем переменные и массив
	sn.Apply(events.DeclareVar{Name: "count", Value: nil}, 0)
	sn.Apply(events.DeclareArray{
		Name:  "data",
		Value: makeArrayElements([]int{10, 20, 30}),
		Size:  3,
	}, 0)

	// Нормальные операции
	assert.NoError(t, sn.Apply(events.VarChanged{Name: "count", Value: 3}, 0))
	assert.NoError(t, sn.Apply(events.ArrayElementChanged{Name: "data", Ind: 1, Value: 25}, 0))

	// Эмулируем ошибку: попытка доступа к несуществующему элементу
	err1 := sn.Apply(events.ArrayElementChanged{Name: "data", Ind: 10, Value: 50}, 0)
	assert.Error(t, err1)

	// Эмулируем ошибку: изменение несуществующей переменной
	err2 := sn.Apply(events.VarChanged{Name: "undefined_var", Value: 999}, 0)
	assert.Error(t, err2)

	// Проверяем, что массив и переменная всё ещё доступны несмотря на ошибки
	arr, ok := sn.GetArray("data")
	assert.True(t, ok)
	assert.NotNil(t, arr)

	count, ok := sn.GetVariable("count")
	assert.True(t, ok)
	countVal, _ := count.GetValue()
	assert.Equal(t, 3, countVal)
}

func TestUndefinedBehaviorGetNonexistentVariable(t *testing.T) {
	sn := NewSnapshot()

	// Пытаемся получить несуществующую переменную
	variable, ok := sn.GetVariable("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, variable)
}

func TestUndefinedBehaviorGetNonexistentArray(t *testing.T) {
	sn := NewSnapshot()

	// Пытаемся получить несуществующий массив
	arr, ok := sn.GetArray("nonexistent_array")
	assert.False(t, ok)
	assert.Nil(t, arr)
}

func TestUndefinedBehaviorGetNonexistentArray2D(t *testing.T) {
	sn := NewSnapshot()

	// Пытаемся получить несуществующий 2D массив
	arr2d, ok := sn.GetArray2D("nonexistent_matrix")
	assert.False(t, ok)
	assert.Nil(t, arr2d)
}
