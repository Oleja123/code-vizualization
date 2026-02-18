# Code Vizualiztion

Приложение для обучения основам программирования с визуализацией выполнения кода. В качестве языка программирования для визуализации используется "подмножество" языка C. Его описание будет дано далее.

## Описание подмножества языка C
- Переменные, массивы (не более двух измерений)
- Типы: `int` и  `void` (только для функций)
- Объявление и инициализация переменных и массивов 
- Присваивание (операторы: `=`, `+=`, `-=`, `*=`, `/=`, `%=`)
- Бинарные операторы (`+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`)
- Поддержка скобок и индексации в массивах
- Управляющие конструкции (`if`, `else`, `for`, `while`, `do while`, `continue`, `break`, `goto`)
- Функции (с возвратом значения и без, тип возвращаемого значения только `int` или `void`, возврат значения через `return`, **передачи массивов нет**)
- **Указателей нет**
- **Препроцессора нет**

## Основные фишки
- Подсветка изменяющихся элементов после шага (благодаря `StepChanged` у `Variable` и `ArrayElement`)
- Поддержка видимых переменных
- Визуализация scope'ов
- Предупреждения об `undefined behavior` при работе с неинициализированными данными

## Общие принципы архитекутуры приложения
- Приложение подключает `semantic-analyzer-service` как библиотеку и получает из него семантически верное ast дерево
- Архитектурная база приложения - `event driven`. Текущее состояние выполнения (`call stack`, `stack frame`, `scope`, `переменные`) задается последовательностью событий, восстановив которые можно получить текущее состояние.
- Интерпретатор генерирует события. После этого они последовательно применяются диспетчером событий.

## Подробное описание архитектуры

### domain

#### events - события

- Event - пустой интерфейс события
- EnterScope - событие входа в scope
- ExitScope - событие выхода из scope
- DeclareVar - событие объявления переменной
    
    **Атрибуты**
    - Name: `string`
    - Value: `*int`
- DeclareArray - событие объявления массива
    
    **Атрибуты**
    - Name: `string`
    - Value: `[]int`
    - Size: `int`
- DeclareArray2D - событие объявления массива
    
    **Атрибуты**
    - Name: `string`
    - Value: `[][]int`
    - Size1: `int`
    - Size2: `int`
- VarChanged - событие изменения значения переменной
    
    **Атрибуты**
    - Name: `string`
    - Value: `int`
- ArrayElementChanged - событие изменения значения элемента массива
    
    **Атрибуты**
    - Name: `string`
    - Ind: `int`
    - Value: `int`
- Array2DElementChanged - событие изменения значения элемента двухмерного массива
    
    **Атрибуты**
    - Name: `string`
    - Ind1: `int`
    - Ind2: `int`
    - Value: `int`
- FunctionCall - событие вызова функции
    
    **Атрибуты**
    - Name: `string`
- FunctionReturn - событие возврата значения из функции
    
    **Атрибуты**
    - Name: `string`
    - Value: `int`
- LineChanged - событие смены строки
    
    **Атрибуты**
    - Line: `int`

#### runtime - сущности времени исполнения (по сути то, что и отрисовывается)

- Declared - пустой интерфейс объявления (для массивов и переменных)
- Variable - переменная
    
    **Атрибуты**
    - Name: `string`
    - Value: `*int`
    - StepChanged: `int`
    
    **Методы**
    - NewVariable(name: `string`, value: `*int`, step: `int`): `Variable`
    - ChangeValue(value: `int`, step: `int`)
    - GetValue(): `int`, `error`
- ArrayElement - элемент массива
    
    **Атрибуты**
    - Value: `*int`
    - StepChanged: `int`
    
    **Методы**
    - NewArrayElement(value: `*int`, step: `int`): `ArrayElement`
    - ChangeValue(value: `int`, step: `int`)
    - GetValue(): `int`, `error`
- Array - массив
    
    **Атрибуты**
    - Name: `string`
    - Size: `int`
    - Values: `[]ArrayElement`
    
    **Методы**
    - NewArray(name: `string`, size: `int`, values: `[]ArrayElement`): `Array`
    - ChangeElement(index: `int`, value: `int`, step: `int`): `error`
    - GetElement(index: `int`): `*ArrayElement`, `error`
- Array2D - двумерный массив
    
    **Атрибуты**
    - Name: `string`
    - Size1: `int`
    - Size2: `int`
    - Values: `[][]ArrayElement`
    
    **Методы**
    - NewArray2D(name: `string`, size1: `int`, size2: `int`, values: `[][]ArrayElement`): `Array2D`
    - ChangeElement(index1: `int`, index2: `int`, value: `int`, step: `int`): `error`
    - GetElement(index1: `int`, index2: `int`): `*ArrayElement`, `error`
- DeclarationStack - последовательность объявлений
    
    **Атрибуты**
    - Declarations: `[]Declared`
    
    **Методы**
    - NewDeclarationStack(): `DeclarationStack`
    - Declare(decl: `Declared`)
    - GetVariable(name: `string`): `*Variable`, `error`
    - GetArray(name: `string`): `*Array`, `error`
    - GetArray2D(name: `string`): `*Array2D`, `error`
- Scope
    
    **Атрибуты**
    - Parent: `*Scope`
    - Declarations: `DeclarationStack`
    
    **Методы**
    - NewScope(Parent: `*Scope`): `Scope`
    - Declare(decl: `Declared`)
    - GetVariable(name: `string`): `*Variable`, `error`
    - GetArray(name: `string`): `*Array`, `error`
    - GetArray2D(name: `string`): `*Array2D`, `error`
- StackFrame
    
    **Атрибуты**
    - FuncName: `string`
    - Scopes: `[]Scope`
    - ReturnValue: `*int*`
    
    **Методы**
    - NewStackFrame(funcName: `string`): `StackFrame`
    - EnterScope()
    - ExitScope(): `error`
    - GetCurrentScope(): `*Scope`
    - GetVariable(name: `string`): `*Variable`, `error`
    - GetArray(name: `string`): `*Array`, `error`
    - GetArray2D(name: `string`): `*Array2D`, `error`
- CallStack
    
    **Атрибуты**
    - Frames: `[]StackFrame`
    
    **Методы**
    - NewCallStack(): `CallStack`
    - PushFrame()
    - PopFrame(): `*int`, `error`
    - GetCurrentFrame(): `*StackFrame`, `error`
- Snapshot - "снимок" текущего состояния исполнения
    
    **Атрибуты**
    - CallStack: `CallStack`
    - Line: `int`
    - Step: `int`
    
    **Методы**
    - Apply(event: `event`): `error`
    - Reset()
    - applyEnterScope(): `error`
    - applyExitScope(): `error`
    - applyDeclareVariable(e: `Event`)`: `error`
    - applyVariableChanged(e: `Event`): `error`
    - applyDeclareArray(e: `Event`): `error`
    - applyArrayElementChanged(e: `Event`): `error`
    - applyDeclareArray2D(e: `Event`): `error`
    - applyArray2DElementChanged(e: `Event`): `error`
    - applyFunctionCall(e: `Event`): `error`
    - applyFunctionReturn(e: `Event`): `error`
    - applyLineChanged(e: `Event`): `error`
#### event_dispatcher - диспетчер событий
- Step - шаг выполнения
    
    **Атрибуты**
    - Events: `[]Event`
- EventDispatcher - диспетчер событий
    
    **Атрибуты**
    - Snapshot: `Snapshot`
    - Steps: `[]Step`
    - currentStep: `Step`
    
    **Методы**
    - BeginStep()
    - Emit(event: `Event`) : `error`
    - EndStep(): `Step`
    - ApplyStep(step: `Step`): `error`
    - Replay(): `error`
    - PopStep()
#### scope_collection - все области видимости программы
- ScopeInfo - информация о Scope

    **Атрибуты**
    - ParentId: `int`
    - LoopCtx: `*LoopContext`
    - BlockAST: `BlockStmt`

- ScopeCollection - коллекция scope'ов

    **Атрибуты**
    - ScopeIds: map[`BlockStmt`]`int`
    - nextId: `int`
    - ScopeInfo: map[`int`]`ScopeInfo`

    **Методы**
    - NewScopeCollection(): `ScopeCollection`
    - EnsureScope(node: `Stmt`, parentId `int`): `int`
    - GetScopeId(node: `BlockStmt`): `int`, `bool`
    - GetScopeInfo(id: `int`): `ScopeInfo`
#### loop - информация о цикле
- LoopContext - информация о цикле

    **Атрибуты**
    - BreakTarget: `Stmt`
    - ContinueTarget: `Stmt`
#### label_collector - сборщик меток для goto
- LabelInfo - информация о метке

    **Атрибуты**
    - Node: `LabelStmt`
    - ScopeChain: `int[]`

- LabelCollector - сборщик меток для goto

    **Атрибуты**
    - ScopeCollection: `ScopeCollection`
    - Labels: `map[string]LabelInfo`

    **Методы**
    - NewLabelCollector(): `LabelCollector`
    - Collect(node: `Node`, chain `[]int`)

### application
#### interpreter - интерпретатор
- Interpreter - интерпретатор

    **Атрибуты**
    - ScopeCollection: `ScopeCollection`
    - LabelCollector: `LabelCollector`
    - CallStack: `CallStack`
    - EventDispatcher: `EventDispatcher`
    - GlobalScope: `Scope`

    **Методы**
    - NewInterpreter(): `Interpreter`
    - PrecomputeLabels(prog: `Program`)
    - ExecuteProgram(prog: `Program`)
    - ExecuteStatement(stmt: `Stmt`) : `error`
    - JumpToNode(stmt: `Stmt`) : `error`
    - executeVariableDecl(v: `VariableDecl`) : `error`
    - executeFunctionDecl(v: `FunctionDecl`) : `error`
    - executeBlockStmt(b: `BlockStmt`) : `error`
    - executeIfStmt(ifStmt: `IfStmt`) : `error`
    - executeLoop(loopStmt: `LoopStmt`) : `error`
    - executeBreak() : `error`
    - executeContinue() : `error`
    - executeGoto(g: `GotoStmt`) : `error`
    - executeLabelStmt(l: `LabelStmt`) : `error`
    - executeReturnStmt(r: `ReturnStmt`) : `*int`, `error`
    - executeExprStmt(e: `ExprStmt`) : `*int`, `error`
    - executeBinaryExprStmt(e: `BinaryExprStmt`) : `*int`, `error`
    - executeUnaryExprStmt(e: `UnaryExprStmt`) : `*int`, `error`
    - executeAssignmentExprStmt(e: `AssignmentExpr`) : `int`, `error`
    - executeCallExpr(c: `CallExpr`) : `int`, `error`
    - executeArrayAccessExpr(a: `ArrayAccessExpr`) : `int`, `error`
    - recomputeLoops(stmt: `Stmt`) : `LoopContext[]`

## Ключевые моменты логики работы интерпретатора

### Работа с goto, continue и break
- Заранее собираются все метки (метка записывается как `имя_метки.имя_функции`), информация о метках будет лежать в `LabelCollector`
- Во время этого прохода, каждому новому scope присваивается свой id
- Также для каждого scope (если это тело цикла) считается свой LoopContext (то есть узлы ast куда нужно прыгать при break и continue)
- Также для каждой метки хранится список id scope'ов, в которые она входит
- При прыжке по метке, текущий stack_frame удаляет все лишние scope и добавляет новые при необходимости. Далее происходит обновление информации о текущих циклах (пробежавшись по scopes) и обновление текущей информации о циклах в StackFrame.