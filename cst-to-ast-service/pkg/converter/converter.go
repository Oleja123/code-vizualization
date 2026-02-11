package converter

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/interfaces"
	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/structs"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
)

// Переэкспорт для публичного API
type (
	// structs переэкспортируются через пакет
	Program      = structs.Program
	VariableDecl = structs.VariableDecl
	FunctionDecl = structs.FunctionDecl
	IfStmt       = structs.IfStmt
	ElseIfClause = structs.ElseIfClause
	WhileStmt    = structs.WhileStmt
	ForStmt      = structs.ForStmt
	ReturnStmt   = structs.ReturnStmt
	BlockStmt    = structs.BlockStmt
	ExprStmt     = structs.ExprStmt
	BreakStmt    = structs.BreakStmt
	ContinueStmt = structs.ContinueStmt

	// Выражения
	VariableExpr    = structs.VariableExpr
	IntLiteral      = structs.IntLiteral
	BinaryExpr      = structs.BinaryExpr
	UnaryExpr       = structs.UnaryExpr
	AssignmentExpr  = structs.AssignmentExpr
	CallExpr        = structs.CallExpr
	ArrayAccessExpr = structs.ArrayAccessExpr
	ArrayInitExpr   = structs.ArrayInitExpr

	// Базовые типы
	Location = structs.Location
	Type     = structs.Type
)

// CConverter реализует конвертер tree-sitter CST в AST для языка C
type CConverter struct {
	parser *sitter.Parser
}

// NewCConverter создает новый конвертер для C
func NewCConverter() *CConverter {
	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())

	return &CConverter{
		parser: parser,
	}
}

// New создает новый конвертер для C (публичный API)
func New() *CConverter {
	return NewCConverter()
}

// ParseToAST парсит C код и возвращает AST или детальную ошибку
// Это основной публичный метод для интерпретаторов
func (c *CConverter) ParseToAST(sourceCode string) (*Program, *ConverterError) {
	tree, err := c.parser.ParseCtx(context.Background(), nil, []byte(sourceCode))
	if err != nil {
		return nil, &ConverterError{
			Code:    ErrParseFailed,
			Message: "tree-sitter parsing failed: " + err.Error(),
			Cause:   err,
		}
	}

	// Используем внутренний метод ConvertToProgram для полной конвертации
	node, err := c.ConvertToProgram(tree, []byte(sourceCode))
	if err != nil {
		// Преобразуем внутреннюю ошибку в публичный формат
		if convErr, ok := err.(*ConverterError); ok {
			return nil, convErr
		}
		// Если это не наша ошибка, оборачиваем в ConverterError
		return nil, &ConverterError{
			Code:    ErrStmtConversion,
			Message: "failed to convert AST: " + err.Error(),
			Cause:   err,
		}
	}

	program, ok := node.(*Program)
	if !ok {
		return nil, &ConverterError{
			Code:    ErrStmtConversion,
			Message: "expected Program node, got something else",
		}
	}

	return program, nil
}

// Parse парсит исходный код C и возвращает дерево tree-sitter
func (c *CConverter) Parse(sourceCode []byte) (*sitter.Tree, error) {
	tree, err := c.parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, newConverterError(ErrParseFailed, "failed to parse source code", nil, err)
	}
	return tree, nil
}

// ConvertToProgram преобразует корневой узел tree-sitter в Program
func (c *CConverter) ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (interfaces.Node, error) {
	rootNode := tree.RootNode()

	program := &structs.Program{
		Type:         "Program",
		Declarations: make([]interfaces.Stmt, 0),
		Loc:          c.getLocation(rootNode),
	}

	// Проходим по всем дочерним узлам корня
	for i := 0; i < int(rootNode.ChildCount()); i++ {
		child := rootNode.Child(i)

		// Пропускаем комментарии и пробелы
		if child.Type() == "comment" {
			continue
		}

		stmt, err := c.ConvertStmt(child, sourceCode)
		if err != nil {
			return nil, newConverterError(ErrStmtConversion, "failed to convert statement", child, err)
		}

		if stmt != nil {
			program.Declarations = append(program.Declarations, stmt)
		}
	}

	return program, nil
}

// ConvertStmt преобразует узел tree-sitter в Statement
func (c *CConverter) ConvertStmt(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	if node == nil {
		return nil, nil
	}

	// Проверяем, есть ли ошибки парсинга tree-sitter для этого узла
	if c.hasTreeSitterError(node) {
		errMsg := c.getTreeSitterErrorMessage(node, sourceCode)
		return nil, newConverterError(ErrTreeSitterError, errMsg, node, nil)
	}

	nodeType := node.Type()

	switch nodeType {
	case "declaration":
		return c.convertDeclaration(node, sourceCode)
	case "function_definition":
		return c.convertFunctionDefinition(node, sourceCode)
	case "if_statement":
		return c.convertIfStatement(node, sourceCode)
	case "while_statement":
		return c.convertWhileStatement(node, sourceCode)
	case "for_statement":
		return c.convertForStatement(node, sourceCode)
	case "return_statement":
		return c.convertReturnStatement(node, sourceCode)
	case "compound_statement":
		return c.convertBlockStatement(node, sourceCode)
	case "expression_statement":
		return c.convertExpressionStatement(node, sourceCode)
	case "break_statement":
		return &structs.BreakStmt{Type: "BreakStmt", Loc: c.getLocation(node)}, nil
	case "continue_statement":
		return &structs.ContinueStmt{Type: "ContinueStmt", Loc: c.getLocation(node)}, nil
	case "comment":
		// Пропускаем комментарии - они не являются AST узлами
		return nil, nil
	default:
		// Неподдерживаемый тип оператора
		return nil, newConverterError(ErrStmtUnsupported, fmt.Sprintf("unsupported statement type: %s", nodeType), node, nil)
	}
}

// ConvertExpr преобразует узел tree-sitter в Expression
func (c *CConverter) ConvertExpr(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	if node == nil {
		return nil, nil
	}

	// Проверяем, есть ли ошибки парсинга tree-sitter для этого узла
	if c.hasTreeSitterError(node) {
		errMsg := c.getTreeSitterErrorMessage(node, sourceCode)
		return nil, newConverterError(ErrTreeSitterError, errMsg, node, nil)
	}

	nodeType := node.Type()

	switch nodeType {
	case "identifier":
		return c.convertIdentifier(node, sourceCode)
	case "number_literal":
		return c.convertIntLiteral(node, sourceCode)
	case "binary_expression":
		return c.convertBinaryExpression(node, sourceCode)
	case "unary_expression":
		return c.convertUnaryExpression(node, sourceCode)
	case "update_expression":
		return c.convertUnaryExpression(node, sourceCode)
	case "assignment_expression":
		return c.convertAssignmentExpression(node, sourceCode)
	case "call_expression":
		return c.convertCallExpression(node, sourceCode)
	case "subscript_expression":
		return c.convertArrayAccessExpression(node, sourceCode)
	case "initializer_list":
		return c.convertArrayInitExpression(node, sourceCode)
	case "pointer_expression":
		return c.convertUnaryExpression(node, sourceCode)
	case "parenthesized_expression":
		// Разворачиваем скобки, пропускаем комментарии
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() != "(" && child.Type() != ")" && child.Type() != "comment" {
				return c.ConvertExpr(child, sourceCode)
			}
		}
		return nil, nil
	case "comment":
		// Пропускаем комментарии в выражениях
		return nil, nil
	default:
		// Неподдерживаемый тип выражения
		return nil, newConverterError(ErrExprUnsupported, fmt.Sprintf("unsupported expression type: %s", nodeType), node, nil)
	}
}

// ============= Helper functions =============

// hasTreeSitterError проверяет, содержит ли узел ошибки парсинга tree-sitter
// Проверяет только сам узел - потомков мы проверим при рекурсивном вызове ConvertStmt/ConvertExpr
func (c *CConverter) hasTreeSitterError(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	return node.HasError() || node.IsMissing()
}

// getTreeSitterErrorMessage извлекает информативное сообщение об ошибке парсинга tree-sitter
// Анализирует дерево ошибки и выводит контекст с информацией о синтаксической проблеме
func (c *CConverter) getTreeSitterErrorMessage(node *sitter.Node, sourceCode []byte) string {
	if node == nil {
		return "tree-sitter parsing error"
	}

	// Если сам узел - ERROR узел, выводим информацию о нем
	if node.Type() == "ERROR" {
		startPoint := node.StartPoint()
		errChar := c.getCharAtPosition(sourceCode, startPoint)
		return fmt.Sprintf("syntax error at line %d, column %d near '%s'",
			startPoint.Row+1, startPoint.Column, errChar)
	}

	// Рекурсивно ищем ERROR узлы в потомках
	errorNode := c.findErrorNode(node)
	if errorNode != nil {
		startPoint := errorNode.StartPoint()
		errChar := c.getCharAtPosition(sourceCode, startPoint)
		return fmt.Sprintf("syntax error at line %d, column %d near '%s'",
			startPoint.Row+1, startPoint.Column, errChar)
	}

	// Ищем IsMissing узлы
	missingNode := c.findMissingNode(node)
	if missingNode != nil {
		startPoint := missingNode.StartPoint()
		return fmt.Sprintf("syntax error at line %d, column %d: missing '%s'",
			startPoint.Row+1, startPoint.Column, missingNode.Type())
	}

	// Общее сообщение по типу узла
	return fmt.Sprintf("tree-sitter parsing error in %s", node.Type())
}

// getCharAtPosition возвращает символ в заданной позиции исходного кода
func (c *CConverter) getCharAtPosition(sourceCode []byte, point sitter.Point) string {
	lines := bytes.Split(sourceCode, []byte("\n"))
	if int(point.Row) >= len(lines) {
		return "EOF"
	}

	line := lines[point.Row]
	if int(point.Column) >= len(line) {
		return "\\n"
	}

	char := line[point.Column]
	if char == '\t' {
		return "\\t"
	}
	if char == ' ' {
		return "SPACE"
	}
	return string(char)
}

// findErrorNode рекурсивно ищет первый ERROR узел в дереве
func (c *CConverter) findErrorNode(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	if node.Type() == "ERROR" {
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if found := c.findErrorNode(node.Child(i)); found != nil {
			return found
		}
	}

	return nil
}

// findMissingNode рекурсивно ищет первый missing узел в дереве
func (c *CConverter) findMissingNode(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	if node.IsMissing() {
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if found := c.findMissingNode(node.Child(i)); found != nil {
			return found
		}
	}

	return nil
}

func (c *CConverter) getLocation(node *sitter.Node) structs.Location {
	startPoint := node.StartPoint()
	endPoint := node.EndPoint()

	return structs.Location{
		Line:      startPoint.Row + 1,
		Column:    startPoint.Column,
		EndLine:   endPoint.Row + 1,
		EndColumn: endPoint.Column,
	}
}

func (c *CConverter) getNodeText(node *sitter.Node, sourceCode []byte) string {
	return node.Content(sourceCode)
}

// ============= Statement converters =============

func (c *CConverter) convertDeclaration(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	// declaration: type declarator [= initializer] ;

	var typeNode *sitter.Node
	var declaratorNode *sitter.Node
	var initNode *sitter.Node

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "primitive_type", "type_identifier":
			typeNode = child
		case "init_declarator":
			// init_declarator содержит: declarator = initializer
			// Структура: [declarator, "=", initializer]
			foundEquals := false
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				childType := subChild.Type()

				if childType == "=" {
					foundEquals = true
					continue
				}

				// До '=' это declarator
				if !foundEquals {
					if childType == "identifier" || childType == "pointer_declarator" ||
						childType == "array_declarator" {
						declaratorNode = subChild
					}
				} else {
					// После '=' это initializer
					if childType != "comment" {
						initNode = subChild
					}
				}
			}
		case "identifier", "pointer_declarator", "array_declarator":
			declaratorNode = child
		}
	}

	if typeNode == nil || declaratorNode == nil {
		return nil, newConverterError(ErrStmtConversion, "invalid declaration: missing type or declarator", node, nil)
	}

	varType, varName := c.parseDeclarator(declaratorNode, sourceCode)
	varType.BaseType = c.getNodeText(typeNode, sourceCode)

	// Проверка на пустое имя переменной
	if varName == "" {
		return nil, newConverterError(ErrStmtConversion, "declaration without variable name", node, nil)
	}

	var initExpr interfaces.Expr
	var err error
	if initNode != nil {
		initExpr, err = c.ConvertExpr(initNode, sourceCode)
		if err != nil {
			return nil, newConverterError(ErrStmtConversion, "failed to convert initializer", initNode, err)
		}
	}

	return &structs.VariableDecl{
		Type:     "VariableDecl",
		VarType:  varType,
		Name:     varName,
		InitExpr: initExpr,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) parseDeclarator(node *sitter.Node, sourceCode []byte) (structs.Type, string) {
	varType := structs.Type{
		BaseType:     "int",
		PointerLevel: 0,
		ArraySizes:   make([]int, 0),
	}

	var name string

	switch node.Type() {
	case "identifier":
		name = c.getNodeText(node, sourceCode)

	case "pointer_declarator":
		// Подсчитываем уровень указателей
		current := node
		for current.Type() == "pointer_declarator" {
			varType.PointerLevel++
			// Ищем дочерний узел, который не является '*'
			for i := 0; i < int(current.ChildCount()); i++ {
				child := current.Child(i)
				if child.Type() != "*" {
					current = child
					break
				}
			}
		}
		// После цикла current может быть identifier или array_declarator
		if current.Type() == "identifier" {
			name = c.getNodeText(current, sourceCode)
		} else if current.Type() == "array_declarator" {
			// Рекурсивно обработать array_declarator (например, int *arr[10])
			baseType, baseName := c.parseDeclarator(current, sourceCode)
			name = baseName
			varType.ArraySizes = baseType.ArraySizes
		}

	case "array_declarator":
		// Рекурсивно обрабатываем вложенные array_declarator для многомерных массивов
		// Сначала ищем вложенный declarator
		var nestedDeclarator *sitter.Node
		var sizes []int

		current := node
		// Собираем все размеры массива в обратном порядке (от внешнего к внутреннему)
		for current.Type() == "array_declarator" {
			// Ищем размер в этом array_declarator
			for i := 0; i < int(current.ChildCount()); i++ {
				child := current.Child(i)
				if child.Type() == "number_literal" {
					sizeStr := c.getNodeText(child, sourceCode)
					if size, err := strconv.Atoi(sizeStr); err == nil {
						sizes = append(sizes, size)
					}
				} else if child.Type() != "[" && child.Type() != "]" && child.Type() != "number_literal" {
					// Это вложенный declarator
					nestedDeclarator = child
				}
			}
			// Если вложенный элемент тоже array_declarator, продолжаем
			if nestedDeclarator != nil && nestedDeclarator.Type() == "array_declarator" {
				current = nestedDeclarator
				nestedDeclarator = nil
			} else {
				break
			}
		}

		// Переворачиваем размеры чтобы они шли от внутреннего к внешнему
		for i, j := 0, len(sizes)-1; i < j; i, j = i+1, j-1 {
			sizes[i], sizes[j] = sizes[j], sizes[i]
		}
		varType.ArraySizes = sizes

		// Теперь обрабатываем вложенный declarator если есть
		if nestedDeclarator != nil {
			name = c.getNodeText(nestedDeclarator, sourceCode)
		} else if current.Type() == "identifier" {
			name = c.getNodeText(current, sourceCode)
		}
	}

	return varType, name
}

func (c *CConverter) convertFunctionDefinition(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var funcName string
	var returnType structs.Type
	var params []structs.Parameter
	var body *structs.BlockStmt
	var funcPointerLevel int

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		switch child.Type() {
		case "primitive_type", "type_identifier":
			typeName := c.getNodeText(child, sourceCode)
			// Принимаем любой тип - семантика обработает
			returnType = structs.Type{
				BaseType:     typeName,
				PointerLevel: 0,
				ArraySizes:   make([]int, 0),
			}

		case "function_declarator":
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					funcName = c.getNodeText(subChild, sourceCode)
				} else if subChild.Type() == "parameter_list" {
					params = c.parseParameterList(subChild, sourceCode)
				}
			}

		case "pointer_declarator":
			name, parsedParams, pointerLevel := c.parseFunctionDeclarator(child, sourceCode)
			if name != "" {
				funcName = name
			}
			if parsedParams != nil {
				params = parsedParams
			}
			funcPointerLevel = pointerLevel

		case "compound_statement":
			blockStmt, err := c.convertBlockStatement(child, sourceCode)
			if err != nil {
				return nil, err
			}
			body = blockStmt.(*structs.BlockStmt)
		}
	}

	if funcPointerLevel > 0 {
		returnType.PointerLevel = funcPointerLevel
	}

	return &structs.FunctionDecl{
		Type:       "FunctionDecl",
		Name:       funcName,
		ReturnType: returnType,
		Parameters: params,
		Body:       body,
		Loc:        c.getLocation(node),
	}, nil
}

func (c *CConverter) parseFunctionDeclarator(node *sitter.Node, sourceCode []byte) (string, []structs.Parameter, int) {
	var name string
	var params []structs.Parameter
	var pointerLevel int

	current := node
	for current.Type() == "pointer_declarator" {
		pointerLevel++
		for i := 0; i < int(current.ChildCount()); i++ {
			child := current.Child(i)
			if child.Type() != "*" {
				current = child
				break
			}
		}
	}

	if current.Type() == "function_declarator" {
		for i := 0; i < int(current.ChildCount()); i++ {
			child := current.Child(i)
			if child.Type() == "identifier" {
				name = c.getNodeText(child, sourceCode)
			} else if child.Type() == "parameter_list" {
				params = c.parseParameterList(child, sourceCode)
			}
		}
	} else if current.Type() == "identifier" {
		name = c.getNodeText(current, sourceCode)
	}

	return name, params, pointerLevel
}

func (c *CConverter) parseParameterList(node *sitter.Node, sourceCode []byte) []structs.Parameter {
	params := make([]structs.Parameter, 0)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() == "parameter_declaration" {
			var paramType structs.Type
			var paramName string

			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)

				if subChild.Type() == "primitive_type" || subChild.Type() == "type_identifier" {
					paramType.BaseType = c.getNodeText(subChild, sourceCode)
				} else if subChild.Type() == "identifier" || subChild.Type() == "pointer_declarator" ||
					subChild.Type() == "array_declarator" {
					paramType, paramName = c.parseDeclarator(subChild, sourceCode)
				}
			}

			params = append(params, structs.Parameter{
				Type: paramType,
				Name: paramName,
				Loc:  c.getLocation(child),
			})
		}
	}

	return params
}

func (c *CConverter) convertIfStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var condition interfaces.Expr
	var thenBlock interfaces.Stmt
	var elseIfList []structs.ElseIfClause
	var elseBlock interfaces.Stmt

	// Структура tree-sitter для if/else if/else:
	// if_statement:
	//   "if"
	//   parenthesized_expression
	//   compound_statement
	//   else_clause (optional)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		switch childType {
		case "parenthesized_expression":
			// Это условие if
			if condition == nil && child.ChildCount() > 1 {
				condExpr, err := c.ConvertExpr(child.Child(1), sourceCode)
				if err != nil {
					return nil, err
				}
				condition = condExpr
			}

		case "compound_statement", "expression_statement", "return_statement",
			"while_statement", "for_statement":
			// Это тело if (если мы ещё не установили тело)
			if thenBlock == nil && condition != nil {
				stmt, err := c.ConvertStmt(child, sourceCode)
				if err != nil {
					return nil, err
				}
				thenBlock = stmt
			}

		case "else_clause":
			// Это else или else if блок
			c.processElseClause(child, sourceCode, &elseIfList, &elseBlock)

		case "if":
			// Ключевое слово "if", пропускаем
		}
	}

	return &structs.IfStmt{
		Type:       "IfStmt",
		Condition:  condition,
		ThenBlock:  thenBlock,
		ElseIfList: elseIfList,
		ElseBlock:  elseBlock,
		Loc:        c.getLocation(node),
	}, nil
}

// processElseClause обрабатывает else_clause
// Структура: else_clause может содержать:
// - if_statement (для else if)
// - compound_statement (для else { ... })
func (c *CConverter) processElseClause(node *sitter.Node, sourceCode []byte, elseIfList *[]structs.ElseIfClause, elseBlock *interfaces.Stmt) error {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		switch child.Type() {
		case "if_statement":
			// Это else if конструкция
			nestedIfStmt, err := c.convertIfStatement(child, sourceCode)
			if err != nil {
				return err
			}

			nestedIf := nestedIfStmt.(*structs.IfStmt)

			// Добавляем условие вложенного if в elseIfList
			*elseIfList = append(*elseIfList, structs.ElseIfClause{
				Condition: nestedIf.Condition,
				Block:     nestedIf.ThenBlock,
				Loc:       nestedIf.Loc,
			})

			// Добавляем все else if из вложенного if
			*elseIfList = append(*elseIfList, nestedIf.ElseIfList...)

			// Последний else блок становится нашим else
			*elseBlock = nestedIf.ElseBlock

		case "compound_statement", "expression_statement", "return_statement":
			// Это обычный else блок
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return err
			}
			*elseBlock = stmt

		case "else":
			// Ключевое слово "else", пропускаем
		}
	}
	return nil
}

func (c *CConverter) convertWhileStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var condition interfaces.Expr
	var body interfaces.Stmt

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		switch child.Type() {
		case "parenthesized_expression":
			if child.ChildCount() > 1 {
				condExpr, err := c.ConvertExpr(child.Child(1), sourceCode)
				if err != nil {
					return nil, err
				}
				condition = condExpr
			}

		case "compound_statement", "expression_statement":
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return nil, err
			}
			body = stmt
		}
	}

	return &structs.WhileStmt{
		Type:      "WhileStmt",
		Condition: condition,
		Body:      body,
		Loc:       c.getLocation(node),
	}, nil
}

func (c *CConverter) convertForStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var init interfaces.Stmt
	var condition interfaces.Expr
	var post interfaces.Stmt
	var body interfaces.Stmt
	var usedFields bool

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		field := node.FieldNameForChild(i)
		if field != "" {
			usedFields = true
		}

		switch field {
		case "initializer":
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return nil, err
			}
			init = stmt
		case "condition":
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			condition = expr
		case "update":
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			post = &structs.ExprStmt{
				Type:       "ExprStmt",
				Expression: expr,
				Loc:        c.getLocation(child),
			}
		case "body":
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return nil, err
			}
			body = stmt
		}
	}

	if usedFields {
		return &structs.ForStmt{
			Type:      "ForStmt",
			Init:      init,
			Condition: condition,
			Post:      post,
			Body:      body,
			Loc:       c.getLocation(node),
		}, nil
	}

	partIndex := 0
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() == "(" {
			continue
		}
		if child.Type() == ")" {
			break
		}
		if child.Type() == ";" {
			partIndex++
			continue
		}
		// Пропускаем комментарии
		if child.Type() == "comment" {
			continue
		}

		switch child.Type() {
		case "declaration":
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return nil, err
			}
			init = stmt

		case "expression_statement":
			if partIndex == 0 {
				stmt, err := c.ConvertStmt(child, sourceCode)
				if err != nil {
					return nil, err
				}
				init = stmt
			}

		case "compound_statement":
			stmt, err := c.ConvertStmt(child, sourceCode)
			if err != nil {
				return nil, err
			}
			body = stmt

		default:
			switch partIndex {
			case 1:
				// Condition
				expr, err := c.ConvertExpr(child, sourceCode)
				if err != nil {
					return nil, err
				}
				condition = expr
			case 2:
				// Post - обрабатываем как выражение и оборачиваем в ExprStmt
				expr, err := c.ConvertExpr(child, sourceCode)
				if err != nil {
					return nil, err
				}
				post = &structs.ExprStmt{
					Type:       "ExprStmt",
					Expression: expr,
					Loc:        c.getLocation(child),
				}
			}
		}
	}

	return &structs.ForStmt{
		Type:      "ForStmt",
		Init:      init,
		Condition: condition,
		Post:      post,
		Body:      body,
		Loc:       c.getLocation(node),
	}, nil
}

func (c *CConverter) convertReturnStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var value interfaces.Expr
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() != "return" && child.Type() != ";" {
			var err error
			value, err = c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
		}
	}

	return &structs.ReturnStmt{
		Type:  "ReturnStmt",
		Value: value,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertBlockStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	statements := make([]interfaces.Stmt, 0)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		// Пропускаем фигурные скобки и комментарии
		if child.Type() == "{" || child.Type() == "}" || child.Type() == "comment" {
			continue
		}

		stmt, err := c.ConvertStmt(child, sourceCode)
		if err != nil {
			return nil, err
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return &structs.BlockStmt{
		Type:       "BlockStmt",
		Statements: statements,
		Loc:        c.getLocation(node),
	}, nil
}

func (c *CConverter) convertExpressionStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var expr interfaces.Expr

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() != ";" {
			var err error
			expr, err = c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
		}
	}

	return &structs.ExprStmt{
		Type:       "ExprStmt",
		Expression: expr,
		Loc:        c.getLocation(node),
	}, nil
}

// ============= Expression converters =============

func (c *CConverter) convertIdentifier(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	name := c.getNodeText(node, sourceCode)

	return &structs.VariableExpr{
		Type: "VariableExpr",
		Name: name,
		Loc:  c.getLocation(node),
	}, nil
}

func (c *CConverter) convertIntLiteral(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	text := c.getNodeText(node, sourceCode)
	value, err := strconv.Atoi(text)
	if err != nil {
		return nil, newConverterError(ErrIntLiteralParse, "failed to parse integer literal", node, err)
	}

	// Если числовой литерал отрицательный, представляем его как UnaryExpr
	// (унарный минус с положительным литералом), а не как отрицательный IntLiteral
	if value < 0 {
		posValue := -value
		posLiteral := &structs.IntLiteral{
			Type:  "IntLiteral",
			Value: posValue,
			Loc:   c.getLocation(node),
		}
		return &structs.UnaryExpr{
			Type:      "UnaryExpr",
			Operator:  "-",
			Operand:   posLiteral,
			IsPostfix: false,
			Loc:       c.getLocation(node),
		}, nil
	}

	return &structs.IntLiteral{
		Type:  "IntLiteral",
		Value: value,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertBinaryExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var left interfaces.Expr
	var operator string
	var right interfaces.Expr
	var partIndex int

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		// Пропускаем комментарии
		if child.Type() == "comment" {
			continue
		}

		switch partIndex {
		case 0:
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			left = expr
		case 1:
			operator = c.getNodeText(child, sourceCode)
		case 2:
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			right = expr
		}
		partIndex++
	}

	return &structs.BinaryExpr{
		Type:     "BinaryExpr",
		Left:     left,
		Operator: operator,
		Right:    right,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) convertUnaryExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var operator string
	var operand interfaces.Expr
	var operatorIndex int
	var operandIndex int

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		// Пропускаем комментарии
		if child.Type() == "comment" {
			continue
		}

		if child.Type() == "!" || child.Type() == "-" || child.Type() == "*" ||
			child.Type() == "&" || child.Type() == "~" || child.Type() == "++" || child.Type() == "--" {
			operator = c.getNodeText(child, sourceCode)
			operatorIndex = i
		} else {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			operand = expr
			operandIndex = i
		}
	}

	// Определяем, постфиксный ли оператор (оператор идёт после операнда)
	isPostfix := operatorIndex > operandIndex

	return &structs.UnaryExpr{
		Type:      "UnaryExpr",
		Operator:  operator,
		Operand:   operand,
		IsPostfix: isPostfix,
		Loc:       c.getLocation(node),
	}, nil
}

func (c *CConverter) convertAssignmentExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var left interfaces.Expr
	var right interfaces.Expr
	var operator string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() == "comment" {
			continue
		}

		if c.isAssignmentOperator(child.Type()) {
			operator = c.getNodeText(child, sourceCode)
			continue
		}

		if left == nil {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			left = expr
		} else {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			right = expr
		}
	}

	return &structs.AssignmentExpr{
		Type:     "AssignmentExpr",
		Left:     left,
		Operator: operator,
		Right:    right,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) convertCallExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var funcName string
	arguments := make([]interfaces.Expr, 0)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() == "identifier" {
			funcName = c.getNodeText(child, sourceCode)
		} else if child.Type() == "argument_list" {
			// Парсим список аргументов
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)

				if arg.Type() != "(" && arg.Type() != ")" && arg.Type() != "," {
					expr, err := c.ConvertExpr(arg, sourceCode)
					if err != nil {
						return nil, err
					}
					arguments = append(arguments, expr)
				}
			}
		}
	}

	return &structs.CallExpr{
		Type:         "CallExpr",
		FunctionName: funcName,
		Arguments:    arguments,
		Loc:          c.getLocation(node),
	}, nil
}

func (c *CConverter) convertArrayAccessExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var array interfaces.Expr
	var index interfaces.Expr

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if i == 0 {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			array = expr
		} else if child.Type() != "[" && child.Type() != "]" {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			index = expr
		}
	}

	return &structs.ArrayAccessExpr{
		Type:  "ArrayAccessExpr",
		Array: array,
		Index: index,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertArrayInitExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	elements := make([]interfaces.Expr, 0)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)

		if child.Type() != "{" && child.Type() != "}" && child.Type() != "," && child.Type() != "comment" {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)
		}
	}

	return &structs.ArrayInitExpr{
		Type:     "ArrayInitExpr",
		Elements: elements,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) isAssignmentOperator(tokenType string) bool {
	return tokenType == "=" || tokenType == "+=" || tokenType == "-=" || tokenType == "/=" ||
		tokenType == "%=" || tokenType == "&=" || tokenType == "|=" || tokenType == "^=" ||
		tokenType == "<<=" || tokenType == ">>="
}
