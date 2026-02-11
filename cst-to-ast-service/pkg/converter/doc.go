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
//	    program, err := conv.Parse(`
//	        int factorial(int n) {
//	            if (n <= 1) return 1;
//	            return n * factorial(n - 1);
//	        }
//	    `)
//	    if err != nil {
//	        // err имеет тип *ConverterError с полной информацией
//	        println("Error at", err.Loc.Line, ":", err.Message)
//	        println("Type:", string(err.Code))
//	        return
//	    }
//	
//	    // program имеет тип *Program
//	    for _, decl := range program.Declarations {
//	        if fn, ok := decl.(*FunctionDecl); ok {
//	            println("Function:", fn.Name)
//	        }
//	    }
//	}
//
// API:
//
//	conv := converter.New()                      // Создает новый конвертер
//	program, err := conv.Parse(code)             // Парсит код в AST
//	tree, err := conv.ParseCST(code)             // Парсит в CST (для отладки)
//	program, err := conv.ConvertCST(tree, code)  // Конвертирует CST в AST
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
//	        fmt.Printf("Error at %d:%d\n", convErr.Loc.Line, convErr.Loc.Column)
//	        fmt.Printf("Code: %s\n", convErr.Code)
//	        fmt.Printf("Message: %s\n", convErr.Message)
//	    }
//	}
//
// Поддерживаемое подмножество C:
//   - Типы: int, int*, int**, int[N]
//   - Переменные с инициализацией
//   - Функции с параметрами
//   - Операторы: if/else if/else, while, for, return, break, continue
//   - Выражения: арифметика, логика, битовые операции, вызовы функций
//   - Присваивание: =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=
//
// Подробная документация находится в ARCHITECTURE.md
package converter
