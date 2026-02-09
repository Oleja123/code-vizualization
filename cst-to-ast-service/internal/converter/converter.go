package converter

import (
	"context"
	"fmt"
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/interfaces"
	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/structs"
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

// Parse парсит исходный код C и возвращает дерево tree-sitter
func (c *CConverter) Parse(sourceCode []byte) (*sitter.Tree, error) {
	tree, err := c.parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source code: %w", err)
	}
	return tree, nil
}

// ConvertToProgram преобразует корневой узел tree-sitter в Program
func (c *CConverter) ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (interfaces.Node, error) {
	rootNode := tree.RootNode()
	
	program := &structs.Program{
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
			return nil, fmt.Errorf("failed to convert statement at line %d: %w", 
				child.StartPoint().Row, err)
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
		return &structs.BreakStmt{Loc: c.getLocation(node)}, nil
	case "continue_statement":
		return &structs.ContinueStmt{Loc: c.getLocation(node)}, nil
	default:
		return nil, fmt.Errorf("unsupported statement type: %s", nodeType)
	}
}

// ConvertExpr преобразует узел tree-sitter в Expression
func (c *CConverter) ConvertExpr(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	if node == nil {
		return nil, nil
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
		// Разворачиваем скобки
		if node.ChildCount() > 0 {
			return c.ConvertExpr(node.Child(1), sourceCode)
		}
		return nil, fmt.Errorf("empty parenthesized expression")
	default:
		return nil, fmt.Errorf("unsupported expression type: %s", nodeType)
	}
}

// ============= Helper functions =============

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
			// init_declarator содержит declarator и инициализатор
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" || subChild.Type() == "pointer_declarator" || 
					subChild.Type() == "array_declarator" {
					declaratorNode = subChild
				} else if subChild.Type() != "=" {
					initNode = subChild
				}
			}
		case "identifier", "pointer_declarator", "array_declarator":
			declaratorNode = child
		}
	}
	
	if typeNode == nil || declaratorNode == nil {
		return nil, fmt.Errorf("invalid declaration: missing type or declarator")
	}
	
	varType, varName := c.parseDeclarator(declaratorNode, sourceCode)
	varType.BaseType = c.getNodeText(typeNode, sourceCode)
	
	var initExpr interfaces.Expr
	var err error
	if initNode != nil {
		initExpr, err = c.ConvertExpr(initNode, sourceCode)
		if err != nil {
			return nil, fmt.Errorf("failed to convert initializer: %w", err)
		}
	}
	
	return &structs.VariableDecl{
		Type:     varType,
		Name:     varName,
		InitExpr: initExpr,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) parseDeclarator(node *sitter.Node, sourceCode []byte) (structs.Type, string) {
	varType := structs.Type{
		BaseType:     "int",
		PointerLevel: 0,
		ArraySize:    0,
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
		if current.Type() == "identifier" {
			name = c.getNodeText(current, sourceCode)
		}
		
	case "array_declarator":
		// int arr[10]
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				name = c.getNodeText(child, sourceCode)
			} else if child.Type() == "number_literal" {
				sizeStr := c.getNodeText(child, sourceCode)
				if size, err := strconv.Atoi(sizeStr); err == nil {
					varType.ArraySize = size
				}
			}
		}
	}
	
	return varType, name
}

func (c *CConverter) convertFunctionDefinition(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var funcName string
	var params []structs.Parameter
	var body *structs.BlockStmt
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		switch child.Type() {
		case "function_declarator":
			// Извлекаем имя функции и параметры
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					funcName = c.getNodeText(subChild, sourceCode)
				} else if subChild.Type() == "parameter_list" {
					params = c.parseParameterList(subChild, sourceCode)
				}
			}
			
		case "compound_statement":
			blockStmt, err := c.convertBlockStatement(child, sourceCode)
			if err != nil {
				return nil, err
			}
			body = blockStmt.(*structs.BlockStmt)
		}
	}
	
	return &structs.FunctionDecl{
		Name:       funcName,
		Parameters: params,
		Body:       body,
		Loc:        c.getLocation(node),
	}, nil
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
		Condition: condition,
		Body:      body,
		Loc:       c.getLocation(node),
	}, nil
}

func (c *CConverter) convertForStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	var init interfaces.Stmt
	var condition interfaces.Expr
	var post interfaces.Expr
	var body interfaces.Stmt
	
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
			if partIndex == 1 {
				// Condition
				expr, err := c.ConvertExpr(child, sourceCode)
				if err != nil {
					return nil, err
				}
				condition = expr
			} else if partIndex == 2 {
				// Post
				expr, err := c.ConvertExpr(child, sourceCode)
				if err != nil {
					return nil, err
				}
				post = expr
			}
		}
	}
	
	return &structs.ForStmt{
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
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			value = expr
			break
		}
	}
	
	return &structs.ReturnStmt{
		Value: value,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertBlockStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	statements := make([]interfaces.Stmt, 0)
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		// Пропускаем фигурные скобки
		if child.Type() == "{" || child.Type() == "}" {
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
		Statements: statements,
		Loc:        c.getLocation(node),
	}, nil
}

func (c *CConverter) convertExpressionStatement(node *sitter.Node, sourceCode []byte) (interfaces.Stmt, error) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if child.Type() != ";" {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			
			return &structs.ExprStmt{
				Expression: expr,
				Loc:        c.getLocation(node),
			}, nil
		}
	}
	
	return nil, fmt.Errorf("empty expression statement")
}

// ============= Expression converters =============

func (c *CConverter) convertIdentifier(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	return &structs.Identifier{
		Name: c.getNodeText(node, sourceCode),
		Loc:  c.getLocation(node),
	}, nil
}

func (c *CConverter) convertIntLiteral(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	text := c.getNodeText(node, sourceCode)
	value, err := strconv.Atoi(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integer literal: %w", err)
	}
	
	return &structs.IntLiteral{
		Value: value,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertBinaryExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var left interfaces.Expr
	var operator string
	var right interfaces.Expr
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if i == 0 {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			left = expr
		} else if i == 1 {
			operator = c.getNodeText(child, sourceCode)
		} else if i == 2 {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			right = expr
		}
	}
	
	return &structs.BinaryExpr{
		Left:     left,
		Operator: operator,
		Right:    right,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) convertUnaryExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var operator string
	var operand interfaces.Expr
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if child.Type() == "!" || child.Type() == "-" || child.Type() == "*" || 
			child.Type() == "&" || child.Type() == "~" || child.Type() == "++" || child.Type() == "--" {
			operator = c.getNodeText(child, sourceCode)
		} else {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			operand = expr
		}
	}
	
	return &structs.UnaryExpr{
		Operator: operator,
		Operand:  operand,
		Loc:      c.getLocation(node),
	}, nil
}

func (c *CConverter) convertAssignmentExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	var left interfaces.Expr
	var right interfaces.Expr
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if i == 0 {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			left = expr
		} else if child.Type() != "=" && i > 0 {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			right = expr
		}
	}
	
	return &structs.AssignmentExpr{
		Left:  left,
		Right: right,
		Loc:   c.getLocation(node),
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
		Array: array,
		Index: index,
		Loc:   c.getLocation(node),
	}, nil
}

func (c *CConverter) convertArrayInitExpression(node *sitter.Node, sourceCode []byte) (interfaces.Expr, error) {
	elements := make([]interfaces.Expr, 0)
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if child.Type() != "{" && child.Type() != "}" && child.Type() != "," {
			expr, err := c.ConvertExpr(child, sourceCode)
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)
		}
	}
	
	return &structs.ArrayInitExpr{
		Elements: elements,
		Loc:      c.getLocation(node),
	}, nil
}
