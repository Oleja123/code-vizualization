package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ Variable Tests ============

func TestNewVariableWithInitialValue(t *testing.T) {
	val := 42
	variable := NewVariable("x", &val, 0, false)

	require.NotNil(t, variable)
	assert.Equal(t, "x", variable.Name)
	assert.NotNil(t, variable.Value)
	assert.Equal(t, 42, *variable.Value)
	assert.Equal(t, 0, variable.StepChanged)
}

func TestNewVariableWithoutInit(t *testing.T) {
	variable := NewVariable("y", nil, 1, false)

	require.NotNil(t, variable)
	assert.Equal(t, "y", variable.Name)
	assert.Nil(t, variable.Value)
	assert.Equal(t, 1, variable.StepChanged)
}

func TestNewVariableGlobal(t *testing.T) {
	variable := NewVariable("global_x", nil, 0, true)

	require.NotNil(t, variable)
	assert.Equal(t, "global_x", variable.Name)
	assert.NotNil(t, variable.Value)
	assert.Equal(t, 0, *variable.Value)
}

func TestVariableGetValue(t *testing.T) {
	val := 99
	variable := NewVariable("x", &val, 0, false)

	result, err := variable.GetValue()

	assert.NoError(t, err)
	assert.Equal(t, 99, result)
}

func TestVariableGetValueUninitialized(t *testing.T) {
	variable := NewVariable("x", nil, 0, false)

	_, err := variable.GetValue()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined behavior")
	assert.Contains(t, err.Error(), "uninitialized")
}

func TestVariableChangeValue(t *testing.T) {
	variable := NewVariable("x", nil, 0, true)

	variable.ChangeValue(50, 2)

	assert.NotNil(t, variable.Value)
	assert.Equal(t, 50, *variable.Value)
	assert.Equal(t, 2, variable.StepChanged)
}

func TestVariableChangeValueOnNilAllocates(t *testing.T) {
	variable := NewVariable("x", nil, 0, false)
	assert.Nil(t, variable.Value)

	variable.ChangeValue(75, 1)

	// Should allocate memory
	assert.NotNil(t, variable.Value)
	assert.Equal(t, 75, *variable.Value)
	assert.Equal(t, 1, variable.StepChanged)
}

func TestVariableChangeValueMultipleTimes(t *testing.T) {
	variable := NewVariable("x", nil, 0, true)

	variable.ChangeValue(10, 1)
	assert.Equal(t, 10, *variable.Value)

	variable.ChangeValue(20, 2)
	assert.Equal(t, 20, *variable.Value)
	assert.Equal(t, 2, variable.StepChanged)
}

func TestVariableChangeValueTracksStep(t *testing.T) {
	variable := NewVariable("x", nil, 0, true)

	variable.ChangeValue(100, 5)
	assert.Equal(t, 5, variable.StepChanged)

	variable.ChangeValue(200, 10)
	assert.Equal(t, 10, variable.StepChanged)
}

// ============ ArrayElement Tests ============

func TestNewArrayElementWithInitialValue(t *testing.T) {
	val := 123
	element := NewArrayElement(&val, 0, false)

	require.NotNil(t, element)
	assert.NotNil(t, element.Value)
	assert.Equal(t, 123, *element.Value)
	assert.Equal(t, 0, element.StepChanged)
}

func TestNewArrayElementWithoutInit(t *testing.T) {
	element := NewArrayElement(nil, 1, false)

	require.NotNil(t, element)
	assert.Nil(t, element.Value)
	assert.Equal(t, 1, element.StepChanged)
}

func TestNewArrayElementGlobal(t *testing.T) {
	element := NewArrayElement(nil, 0, true)

	require.NotNil(t, element)
	assert.NotNil(t, element.Value)
	assert.Equal(t, 0, *element.Value)
}

func TestArrayElementGetValue(t *testing.T) {
	val := 77
	element := NewArrayElement(&val, 0, false)

	result, err := element.GetValue()

	assert.NoError(t, err)
	assert.Equal(t, 77, result)
}

func TestArrayElementGetValueUninitialized(t *testing.T) {
	element := NewArrayElement(nil, 0, false)

	_, err := element.GetValue()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined behavior")
}

func TestArrayElementChangeValue(t *testing.T) {
	element := NewArrayElement(nil, 0, true)

	element.ChangeValue(88, 3)

	assert.NotNil(t, element.Value)
	assert.Equal(t, 88, *element.Value)
	assert.Equal(t, 3, element.StepChanged)
}

func TestArrayElementChangeValueOnNilAllocates(t *testing.T) {
	element := NewArrayElement(nil, 0, false)
	assert.Nil(t, element.Value)

	element.ChangeValue(50, 1)

	// Should allocate memory
	assert.NotNil(t, element.Value)
	assert.Equal(t, 50, *element.Value)
	assert.Equal(t, 1, element.StepChanged)
}

func TestArrayElementChangeValueMultipleTimes(t *testing.T) {
	element := NewArrayElement(nil, 0, true)

	element.ChangeValue(5, 1)
	assert.Equal(t, 5, *element.Value)

	element.ChangeValue(15, 2)
	assert.Equal(t, 15, *element.Value)
	assert.Equal(t, 2, element.StepChanged)
}

func TestArrayElementChangeValueTracksStep(t *testing.T) {
	element := NewArrayElement(nil, 0, true)

	element.ChangeValue(100, 7)
	assert.Equal(t, 7, element.StepChanged)

	element.ChangeValue(200, 14)
	assert.Equal(t, 14, element.StepChanged)
}

// ============ Array Tests ============

func TestNewArrayWithoutInit(t *testing.T) {
	arr := NewArray("nums", 5, nil, 0, true)

	require.NotNil(t, arr)
	assert.Equal(t, "nums", arr.Name)
	assert.Equal(t, 5, arr.Size)
	assert.Len(t, arr.Values, 5)

	// All elements should be initialized to 0 (global array)
	for i, elem := range arr.Values {
		assert.NotNil(t, elem.Value, "element at index %d should have value", i)
		assert.Equal(t, 0, *elem.Value, "element at index %d should be 0", i)
	}
}

func TestNewArrayLocal(t *testing.T) {
	arr := NewArray("local_nums", 3, nil, 0, false)

	require.NotNil(t, arr)
	assert.Equal(t, "local_nums", arr.Name)
	assert.Equal(t, 3, arr.Size)
	assert.Len(t, arr.Values, 3)

	// Local array elements should be uninitialized
	for _, elem := range arr.Values {
		assert.Nil(t, elem.Value)
	}
}

func TestNewArrayWithInitializedValues(t *testing.T) {
	// Create pre-initialized array elements
	elements := make([]ArrayElement, 3)
	for i := range elements {
		val := i * 10
		elements[i] = *NewArrayElement(&val, 0, false)
	}

	arr := NewArray("initialized_arr", 3, elements, 5, false)

	require.NotNil(t, arr)
	assert.Equal(t, "initialized_arr", arr.Name)
	assert.Equal(t, 3, arr.Size)
	assert.Len(t, arr.Values, 3)

	// Check that the provided values are used
	for i, elem := range arr.Values {
		assert.NotNil(t, elem.Value)
		assert.Equal(t, i*10, *elem.Value)
	}
}

func TestArrayChangeElement(t *testing.T) {
	arr := NewArray("arr", 5, nil, 0, true)

	err := arr.ChangeElement(2, 42, 1)

	assert.NoError(t, err)
	assert.Equal(t, 42, *arr.Values[2].Value)
	assert.Equal(t, 1, arr.Values[2].StepChanged)
}

func TestArrayChangeElementOutOfBounds(t *testing.T) {
	arr := NewArray("arr", 5, nil, 0, true)

	errNegative := arr.ChangeElement(-1, 10, 1)
	assert.Error(t, errNegative)
	assert.Contains(t, errNegative.Error(), "undefined behavior")

	errTooLarge := arr.ChangeElement(10, 20, 1)
	assert.Error(t, errTooLarge)
	assert.Contains(t, errTooLarge.Error(), "undefined behavior")
}

func TestArrayGetElement(t *testing.T) {
	arr := NewArray("arr", 5, nil, 0, true)
	arr.ChangeElement(1, 99, 1)

	val, err := arr.GetElement(1)

	assert.NoError(t, err)
	valInt, _ := val.GetValue()
	assert.Equal(t, 99, valInt)
}

func TestArrayGetElementOutOfBounds(t *testing.T) {
	arr := NewArray("arr", 5, nil, 0, true)

	_, errNegative := arr.GetElement(-1)
	assert.Error(t, errNegative)

	_, errTooLarge := arr.GetElement(10)
	assert.Error(t, errTooLarge)
}

func TestArrayGetElementUninitialized(t *testing.T) {
	arr := NewArray("arr", 5, nil, 0, false)

	val, err := arr.GetElement(0)
	_, err = val.GetValue()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined behavior")
}

// ============ DeclarationStack Tests ============

func TestDeclarationStackDeclareVariable(t *testing.T) {
	ds := &DeclarationStack{}
	variable := NewVariable("x", nil, 0, true)

	ds.Declare(variable)

	assert.Len(t, ds.Declarations, 1)
}

func TestDeclarationStackDeclareMultiple(t *testing.T) {
	ds := &DeclarationStack{}
	var1 := NewVariable("x", nil, 0, true)
	var2 := NewVariable("y", nil, 0, true)
	arr := NewArray("arr", 5, nil, 0, true)

	ds.Declare(var1)
	ds.Declare(var2)
	ds.Declare(arr)

	assert.Len(t, ds.Declarations, 3)
}

func TestDeclarationStackGetVariable(t *testing.T) {
	ds := &DeclarationStack{}
	variable := NewVariable("x", nil, 0, true)
	ds.Declare(variable)

	found, ok := ds.GetVariable("x")

	assert.True(t, ok)
	assert.NotNil(t, found)
	assert.Equal(t, "x", found.Name)
}

func TestDeclarationStackGetVariableNotFound(t *testing.T) {
	ds := &DeclarationStack{}
	variable := NewVariable("x", nil, 0, true)
	ds.Declare(variable)

	_, ok := ds.GetVariable("y")

	assert.False(t, ok)
}

func TestDeclarationStackGetArray(t *testing.T) {
	ds := &DeclarationStack{}
	arr := NewArray("numbers", 10, nil, 0, true)
	ds.Declare(arr)

	found, ok := ds.GetArray("numbers")

	assert.True(t, ok)
	assert.NotNil(t, found)
	assert.Equal(t, "numbers", found.Name)
	assert.Equal(t, 10, found.Size)
}

func TestDeclarationStackGetArrayNotFound(t *testing.T) {
	ds := &DeclarationStack{}
	arr := NewArray("numbers", 10, nil, 0, true)
	ds.Declare(arr)

	_, ok := ds.GetArray("missing")

	assert.False(t, ok)
}

func TestDeclarationStackMixedTypes(t *testing.T) {
	ds := &DeclarationStack{}

	v1 := NewVariable("x", nil, 0, true)
	v2 := NewVariable("y", nil, 0, true)
	arr := NewArray("arr", 5, nil, 0, true)

	ds.Declare(v1)
	ds.Declare(arr)
	ds.Declare(v2)

	// Check order
	assert.Len(t, ds.Declarations, 3)

	// Check retrieval
	foundV1, ok1 := ds.GetVariable("x")
	assert.True(t, ok1)
	assert.Equal(t, "x", foundV1.Name)

	foundArr, okArr := ds.GetArray("arr")
	assert.True(t, okArr)
	assert.Equal(t, "arr", foundArr.Name)

	foundV2, ok2 := ds.GetVariable("y")
	assert.True(t, ok2)
	assert.Equal(t, "y", foundV2.Name)

}

func TestDeclarationStackGetVariableWrongType(t *testing.T) {
	ds := &DeclarationStack{}
	arr := NewArray("arr", 5, nil, 0, true)
	ds.Declare(arr)

	// Trying to get array as variable should fail
	_, ok := ds.GetVariable("arr")

	assert.False(t, ok)
}

func TestDeclarationStackGetArrayWrongType(t *testing.T) {
	ds := &DeclarationStack{}
	variable := NewVariable("x", nil, 0, true)
	ds.Declare(variable)

	// Trying to get variable as array should fail
	_, ok := ds.GetArray("x")

	assert.False(t, ok)
}

func TestDeclarationStackOrderPreservation(t *testing.T) {
	ds := &DeclarationStack{}

	names := []string{"first", "second", "third"}
	for _, name := range names {
		v := NewVariable(name, nil, 0, true)
		ds.Declare(v)
	}

	// Check that order is preserved
	for i, name := range names {
		v, ok := ds.GetVariable(name)
		assert.True(t, ok)
		assert.Equal(t, name, v.Name)
		assert.Equal(t, i, findDeclIndex(ds, name))
	}
}

// Helper function to find index of declaration
func findDeclIndex(ds *DeclarationStack, name string) int {
	for i, decl := range ds.Declarations {
		if v, ok := decl.(*Variable); ok && v.Name == name {
			return i
		}
		if a, ok := decl.(*Array); ok && a.Name == name {
			return i
		}
		if a2, ok := decl.(*Array2D); ok && a2.Name == name {
			return i
		}
	}
	return -1
}

// ============ Scope Tests ============

func TestNewScopeGlobal(t *testing.T) {
	scope := NewScope(nil)

	assert.NotNil(t, scope)
	assert.Nil(t, scope.Parent)
	assert.Empty(t, scope.Declarations.Declarations)
}

func TestNewScopeWithParent(t *testing.T) {
	parentScope := NewScope(nil)
	childScope := NewScope(parentScope)

	assert.NotNil(t, childScope)
	assert.Equal(t, parentScope, childScope.Parent)
}

func TestScopeDeclareVariable(t *testing.T) {
	scope := NewScope(nil)
	variable := NewVariable("x", nil, 0, true)

	scope.Declare(variable)

	assert.Len(t, scope.Declarations.Declarations, 1)
}

func TestScopeDeclareMultiple(t *testing.T) {
	scope := NewScope(nil)
	v1 := NewVariable("x", nil, 0, true)
	v2 := NewVariable("y", nil, 0, true)
	arr := NewArray("arr", 5, nil, 0, true)

	scope.Declare(v1)
	scope.Declare(v2)
	scope.Declare(arr)

	assert.Len(t, scope.Declarations.Declarations, 3)
}

func TestScopeGetVariable(t *testing.T) {
	scope := NewScope(nil)
	variable := NewVariable("x", nil, 0, true)
	scope.Declare(variable)

	found, ok := scope.GetVariable("x")

	assert.True(t, ok)
	assert.NotNil(t, found)
	assert.Equal(t, "x", found.Name)
}

func TestScopeGetVariableNotFound(t *testing.T) {
	scope := NewScope(nil)
	variable := NewVariable("x", nil, 0, true)
	scope.Declare(variable)

	_, ok := scope.GetVariable("y")

	assert.False(t, ok)
}

func TestScopeGetVariableCurrentScopeOnly(t *testing.T) {
	// Scope should only search in its own declarations
	// Not in parent scope (that's StackFrame's responsibility)
	parentScope := NewScope(nil)
	parentVar := NewVariable("parent_var", nil, 0, true)
	parentScope.Declare(parentVar)

	childScope := NewScope(parentScope)
	childVar := NewVariable("child_var", nil, 0, true)
	childScope.Declare(childVar)

	// Should find in child scope
	found, ok := childScope.GetVariable("child_var")
	assert.True(t, ok)
	assert.Equal(t, "child_var", found.Name)

	// Should NOT find in parent scope (scope doesn't search parents)
	_, ok = childScope.GetVariable("parent_var")
	assert.False(t, ok)
}

func TestScopeGetArray(t *testing.T) {
	scope := NewScope(nil)
	arr := NewArray("numbers", 10, nil, 0, true)
	scope.Declare(arr)

	found, ok := scope.GetArray("numbers")

	assert.True(t, ok)
	assert.NotNil(t, found)
	assert.Equal(t, "numbers", found.Name)
}

func TestScopeGetArrayNotFound(t *testing.T) {
	scope := NewScope(nil)
	arr := NewArray("numbers", 10, nil, 0, true)
	scope.Declare(arr)

	_, ok := scope.GetArray("missing")

	assert.False(t, ok)
}

func TestScopeMixedDeclarations(t *testing.T) {
	scope := NewScope(nil)

	v1 := NewVariable("x", nil, 0, true)
	arr := NewArray("arr", 5, nil, 0, true)
	v2 := NewVariable("y", nil, 0, true)

	scope.Declare(v1)
	scope.Declare(arr)
	scope.Declare(v2)

	// Check we can find each
	foundV1, okV1 := scope.GetVariable("x")
	assert.True(t, okV1)
	assert.Equal(t, "x", foundV1.Name)

	foundArr, okArr := scope.GetArray("arr")
	assert.True(t, okArr)
	assert.Equal(t, "arr", foundArr.Name)

	foundV2, okV2 := scope.GetVariable("y")
	assert.True(t, okV2)
	assert.Equal(t, "y", foundV2.Name)
}

func TestScopeHierarchy(t *testing.T) {
	// Test that parent pointer is correctly set
	level1 := NewScope(nil)
	level2 := NewScope(level1)
	level3 := NewScope(level2)

	assert.Nil(t, level1.Parent)
	assert.Equal(t, level1, level2.Parent)
	assert.Equal(t, level2, level3.Parent)
}

func TestScopeDeclareAndRetrieveSameVariable(t *testing.T) {
	scope := NewScope(nil)
	v := NewVariable("test", nil, 0, true)
	scope.Declare(v)

	// Change the variable
	v.ChangeValue(42, 1)

	// Retrieve it and check the change is visible
	found, ok := scope.GetVariable("test")
	assert.True(t, ok)
	val, err := found.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)
}

func TestScopeDeclareAndRetrieveSameArray(t *testing.T) {
	scope := NewScope(nil)
	arr := NewArray("numbers", 5, nil, 0, true)
	scope.Declare(arr)

	// Change the array element
	err := arr.ChangeElement(2, 99, 1)
	assert.NoError(t, err)

	// Retrieve it and check the change is visible
	found, ok := scope.GetArray("numbers")
	assert.True(t, ok)
	val, err := found.GetElement(2)
	assert.NoError(t, err)
	valInt, err := val.GetValue()
	assert.Equal(t, 99, valInt)
}

// ============ StackFrame Tests ============

func TestNewStackFrame(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	require.NotNil(t, stackFrame)
	assert.Equal(t, "main", stackFrame.FuncName)
	assert.Len(t, stackFrame.Scopes, 1)
	assert.Equal(t, globalScope, stackFrame.Scopes[0])
}

func TestStackFrameGetCurrentScope(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	current := stackFrame.GetCurrentScope()

	assert.NotNil(t, current)
	assert.Equal(t, globalScope, current)
}

func TestStackFrameEnterScope(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	stackFrame.EnterScope()

	assert.Len(t, stackFrame.Scopes, 2)
	current := stackFrame.GetCurrentScope()
	assert.NotNil(t, current)
	assert.Equal(t, globalScope, current.Parent)
}

func TestStackFrameEnterMultipleScopes(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	stackFrame.EnterScope()
	stackFrame.EnterScope()
	stackFrame.EnterScope()

	assert.Len(t, stackFrame.Scopes, 4)
}

func TestStackFrameExitScope(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)
	stackFrame.EnterScope()

	err := stackFrame.ExitScope()

	assert.NoError(t, err)
	assert.Len(t, stackFrame.Scopes, 1)
}

func TestStackFrameExitScopeGlobal(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	// Should not be able to exit global scope
	err := stackFrame.ExitScope()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected internal error")
	assert.Len(t, stackFrame.Scopes, 1)
}

func TestStackFrameDeclare(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)
	variable := NewVariable("x", nil, 0, true)

	stackFrame.Declare(variable)

	found, ok := stackFrame.GetCurrentScope().GetVariable("x")
	assert.True(t, ok)
	assert.Equal(t, "x", found.Name)
}

func TestStackFrameSetReturnValue(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("func", globalScope)

	stackFrame.SetReturnValue(42)

	assert.NotNil(t, stackFrame.ReturnValue)
	assert.Equal(t, 42, *stackFrame.ReturnValue)
}

func TestStackFrameGetReturnValue(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("func", globalScope)
	stackFrame.SetReturnValue(99)

	val, err := stackFrame.GetReturnValue()

	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, 99, *val)
}

func TestStackFrameGetReturnValueNotSet(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("func", globalScope)

	_, err := stackFrame.GetReturnValue()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected internal error")
}

func TestStackFrameSetReturnValueMultipleTimes(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("func", globalScope)

	stackFrame.SetReturnValue(10)
	assert.Equal(t, 10, *stackFrame.ReturnValue)

	stackFrame.SetReturnValue(20)
	assert.Equal(t, 20, *stackFrame.ReturnValue)
}

func TestStackFrameGetVariableCurrentScope(t *testing.T) {
	globalScope := NewScope(nil)
	v := NewVariable("x", nil, 0, true)
	globalScope.Declare(v)

	stackFrame := NewStackFrame("main", globalScope)

	found, ok := stackFrame.GetVariable("x")
	assert.True(t, ok)
	assert.Equal(t, "x", found.Name)
}

func TestStackFrameGetVariableFromParentScope(t *testing.T) {
	globalScope := NewScope(nil)
	globalVar := NewVariable("global_x", nil, 0, true)
	globalScope.Declare(globalVar)

	stackFrame := NewStackFrame("main", globalScope)
	stackFrame.EnterScope()

	// Variable declared in global scope should be found from child scope
	found, ok := stackFrame.GetVariable("global_x")
	assert.True(t, ok)
	assert.Equal(t, "global_x", found.Name)
}

func TestStackFrameGetVariableShadowing(t *testing.T) {
	globalScope := NewScope(nil)
	globalVar := NewVariable("x", nil, 0, true)
	globalVar.ChangeValue(10, 0)
	globalScope.Declare(globalVar)

	stackFrame := NewStackFrame("main", globalScope)
	stackFrame.EnterScope()

	localVar := NewVariable("x", nil, 0, true)
	localVar.ChangeValue(20, 0)
	stackFrame.GetCurrentScope().Declare(localVar)

	// Should find the local variable, not global
	found, ok := stackFrame.GetVariable("x")
	assert.True(t, ok)
	val, err := found.GetValue()
	assert.NoError(t, err)
	assert.Equal(t, 20, val)
}

func TestStackFrameGetArrayFromParentScope(t *testing.T) {
	globalScope := NewScope(nil)
	arr := NewArray("numbers", 5, nil, 0, true)
	globalScope.Declare(arr)

	stackFrame := NewStackFrame("main", globalScope)
	stackFrame.EnterScope()

	found, ok := stackFrame.GetArray("numbers")
	assert.True(t, ok)
	assert.Equal(t, "numbers", found.Name)
}

func TestStackFrameGetVariableDeepHierarchy(t *testing.T) {
	globalScope := NewScope(nil)
	v := NewVariable("x", nil, 0, true)
	globalScope.Declare(v)

	stackFrame := NewStackFrame("main", globalScope)
	stackFrame.EnterScope()
	stackFrame.EnterScope()
	stackFrame.EnterScope()

	// Should find variable even deep in hierarchy
	found, ok := stackFrame.GetVariable("x")
	assert.True(t, ok)
	assert.Equal(t, "x", found.Name)
}

func TestStackFrameGetVariableNotFound(t *testing.T) {
	globalScope := NewScope(nil)
	stackFrame := NewStackFrame("main", globalScope)

	_, ok := stackFrame.GetVariable("nonexistent")

	assert.False(t, ok)
}

// ============ CallStack Tests ============

func TestNewCallStack(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	assert.NotNil(t, callStack)
	assert.Len(t, callStack.Frames, 1)
	assert.Equal(t, "global", callStack.Frames[0].FuncName)
}

func TestCallStackGetCurrentFrame(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	frame := callStack.GetCurrentFrame()

	assert.NotNil(t, frame)
	assert.Equal(t, "global", frame.FuncName)
}

func TestCallStackFramesCount(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	assert.Equal(t, 1, callStack.FramesCount())
}

func TestCallStackPushFrame(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	newFrame := NewStackFrame("func1", globalScope)
	callStack.PushFrame(newFrame)

	assert.Equal(t, 2, callStack.FramesCount())
	assert.Equal(t, "func1", callStack.GetCurrentFrame().FuncName)
}

func TestCallStackPushMultipleFrames(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	frame1 := NewStackFrame("func1", globalScope)
	frame2 := NewStackFrame("func2", globalScope)
	frame3 := NewStackFrame("func3", globalScope)

	callStack.PushFrame(frame1)
	callStack.PushFrame(frame2)
	callStack.PushFrame(frame3)

	assert.Equal(t, 4, callStack.FramesCount())
	assert.Equal(t, "func3", callStack.GetCurrentFrame().FuncName)
}

func TestCallStackPopFrame(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	frame := NewStackFrame("func1", globalScope)
	callStack.PushFrame(frame)

	err := callStack.PopFrame()

	assert.NoError(t, err)
	assert.Equal(t, 1, callStack.FramesCount())
	assert.Equal(t, "global", callStack.GetCurrentFrame().FuncName)
}

func TestCallStackPopMultipleFrames(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	frame1 := NewStackFrame("func1", globalScope)
	frame2 := NewStackFrame("func2", globalScope)

	callStack.PushFrame(frame1)
	callStack.PushFrame(frame2)

	callStack.PopFrame()
	assert.Equal(t, 2, callStack.FramesCount())
	assert.Equal(t, "func1", callStack.GetCurrentFrame().FuncName)

	callStack.PopFrame()
	assert.Equal(t, 1, callStack.FramesCount())
	assert.Equal(t, "global", callStack.GetCurrentFrame().FuncName)
}

func TestCallStackPopMainFrame(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	// Should not be able to pop main frame
	err := callStack.PopFrame()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected internal error")
	assert.Equal(t, 1, callStack.FramesCount())
}

func TestCallStackDeclareInCurrentFrame(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	variable := NewVariable("x", nil, 0, true)
	callStack.DeclareInCurrentFrame(variable)

	frame := callStack.GetCurrentFrame()
	found, ok := frame.GetVariable("x")
	assert.True(t, ok)
	assert.Equal(t, "x", found.Name)
}

func TestCallStackGetVariableInCurrentFrame(t *testing.T) {
	globalScope := NewScope(nil)
	v := NewVariable("x", nil, 0, true)
	globalScope.Declare(v)

	callStack := NewCallStack(globalScope)

	found, ok := callStack.GetVariableInCurrentFrame("x")
	assert.True(t, ok)
	assert.Equal(t, "x", found.Name)
}

func TestCallStackGetArrayInCurrentFrame(t *testing.T) {
	globalScope := NewScope(nil)
	arr := NewArray("numbers", 5, nil, 0, true)
	globalScope.Declare(arr)

	callStack := NewCallStack(globalScope)

	found, ok := callStack.GetArrayInCurrentFrame("numbers")
	assert.True(t, ok)
	assert.Equal(t, "numbers", found.Name)
}

func TestCallStackMultipleFramesWithDifferentVariables(t *testing.T) {
	globalScope := NewScope(nil)
	globalVar := NewVariable("global", nil, 0, true)
	globalScope.Declare(globalVar)

	callStack := NewCallStack(globalScope)

	// First frame
	func1Frame := NewStackFrame("func1", globalScope)
	func1Frame.EnterScope()
	var1 := NewVariable("local1", nil, 0, true)
	func1Frame.GetCurrentScope().Declare(var1)
	callStack.PushFrame(func1Frame)

	// Second frame
	func2Frame := NewStackFrame("func2", globalScope)
	func2Frame.EnterScope()
	var2 := NewVariable("local2", nil, 0, true)
	func2Frame.GetCurrentScope().Declare(var2)
	callStack.PushFrame(func2Frame)

	// In func2, should find local2 and global (which is in global scope)
	found1, ok1 := callStack.GetVariableInCurrentFrame("local2")
	assert.True(t, ok1)
	assert.Equal(t, "local2", found1.Name)

	foundGlobal, okGlobal := callStack.GetVariableInCurrentFrame("global")
	assert.True(t, okGlobal)
	assert.Equal(t, "global", foundGlobal.Name)

	// local1 is in func1, not in func2
	_, ok2 := callStack.GetVariableInCurrentFrame("local1")
	assert.False(t, ok2)
}

func TestCallStackDeclareInDifferentFrames(t *testing.T) {
	globalScope := NewScope(nil)
	callStack := NewCallStack(globalScope)

	// Declare in main frame's local scope
	callStack.GetCurrentFrame().EnterScope()
	mainVar := NewVariable("main_var", nil, 0, true)
	callStack.GetCurrentFrame().GetCurrentScope().Declare(mainVar)

	// Create and push func1 frame
	func1Frame := NewStackFrame("func1", globalScope)
	func1Frame.EnterScope()
	func1Var := NewVariable("func1_var", nil, 0, true)
	func1Frame.GetCurrentScope().Declare(func1Var)
	callStack.PushFrame(func1Frame)

	// In func1, should find func1_var (in local scope)
	found, ok := callStack.GetVariableInCurrentFrame("func1_var")
	assert.True(t, ok)
	assert.Equal(t, "func1_var", found.Name)

	// In func1, should NOT find main_var (it's in main frame, not accessible)
	_, ok2 := callStack.GetVariableInCurrentFrame("main_var")
	assert.False(t, ok2)

	// Pop back to main
	callStack.PopFrame()

	// In main, should find main_var (in local scope)
	found3, ok3 := callStack.GetVariableInCurrentFrame("main_var")
	assert.True(t, ok3)
	assert.Equal(t, "main_var", found3.Name)

	// In main, should NOT find func1_var (it's in func1 frame)
	_, ok4 := callStack.GetVariableInCurrentFrame("func1_var")
	assert.False(t, ok4)
}
