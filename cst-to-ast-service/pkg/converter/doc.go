// Package converter предоставляет публичный API для парсинга C кода в AST
//
// Основное использование:
//
//	import (
//	    "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
//	)
//
//	func main() {
//	    conv := converter.New()
//	    program, err := conv.ParseToAST(`
//	        int factorial(int n) {
//	            if (n <= 1) return 1;
//	            return n * factorial(n - 1);
//	        }
//	    `)
//	    if err != nil {
//	        // err имеет тип *ConverterError с полной информацией
//	        loc := err.GetLocation()
//	        println("Error at", loc.Line, ":", err.GetMessage())
//	        println("Type:", string(err.GetCode()))
//	        return
//	    }
//
//	    // program имеет тип *Program
//	    for _, decl := range program.Declarations {
//	        if fn, ok := decl.(*converter.FunctionDecl); ok {
//	            println("Function:", fn.Name)
//	        }
//	    }
//	}
//
// API:
//
//	conv := converter.New()                                  // Создает новый конвертер
//	program, err := conv.ParseToAST(code)                    // Парсит код в AST
//	tree, err := conv.Parse([]byte(code))                    // Парсит в CST (для отладки)
//	node, err := conv.ConvertToProgram(tree, []byte(code))   // Конвертирует CST в AST
//
// Ошибки:
//
// Все ошибки имеют тип *ConverterError с полной информацией для интерпретатора:
//   - Code: ErrorCode - тип ошибки (ParseFailed, StmtUnsupported, etc.)
//   - Message: string - понятное описание
//   - Loc: Location - позиция в коде (line, column)
//   - NodeType: string - тип узла tree-sitter (для отладки)
//
// Интерпретатор может обработать ошибку:
//
//	if err != nil {
//	    if convErr, ok := err.(*converter.ConverterError); ok {
//	        loc := convErr.GetLocation()
//	        fmt.Printf("Error at %d:%d\n", loc.Line, loc.Column)
//	        fmt.Printf("Code: %s\n", convErr.GetCode())
//	        fmt.Printf("Message: %s\n", convErr.GetMessage())
//	    }
//	}
//
// Поддерживаемое подмножество C:
//   - Типы: int, int*, int**, int[N]
//   - Переменные с инициализацией
//   - Функции с параметрами
//   - Операторы: if/else if/else, while, do-while, for, return, break, continue, goto, label
//   - Выражения: арифметика, логика, битовые операции, вызовы функций
//   - Присваивание: =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=
//
// Подробная документация находится в ARCHITECTURE.md
package converter
