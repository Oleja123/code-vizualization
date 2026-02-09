package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
	// Пример простой C программы
	sourceCode := []byte(`
int factorial(int n) {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}

int main() {
	int x = 5;
	int result = factorial(x);
	return result;
}
`)

	fmt.Println("=== Исходный код C ===")
	fmt.Println(string(sourceCode))
	fmt.Println()

	// Создаем конвертер
	conv := converter.NewCConverter()

	// Парсим исходный код в CST (tree-sitter)
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		log.Fatalf("Ошибка парсинга: %v", err)
	}

	fmt.Println("=== Tree-sitter CST ===")
	fmt.Println(tree.RootNode().String())
	fmt.Println()

	// Конвертируем CST в AST
	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		log.Fatalf("Ошибка конвертации: %v", err)
	}

	// Выводим AST в JSON формате
	fmt.Println("=== AST (JSON) ===")
	jsonData, err := json.MarshalIndent(ast, "", "  ")
	if err != nil {
		log.Fatalf("Ошибка сериализации в JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
