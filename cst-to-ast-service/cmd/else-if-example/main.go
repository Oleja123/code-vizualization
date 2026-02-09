package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
	// Пример с else if конструкциями
	sourceCode := []byte(`
int grade(int score) {
	if (score >= 90) {
		return 5;
	} else if (score >= 80) {
		return 4;
	} else if (score >= 70) {
		return 3;
	} else if (score >= 60) {
		return 2;
	} else {
		return 1;
	}
}

int main() {
	int result = grade(85);
	return result;
}
`)

	fmt.Println("=== Исходный код C (с else if) ===")
	fmt.Println(string(sourceCode))
	fmt.Println()

	// Создаем конвертер
	conv := converter.NewCConverter()

	// Парсим исходный код
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		log.Fatalf("Ошибка парсинга: %v", err)
	}

	// Конвертируем в AST
	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		log.Fatalf("Ошибка конвертации: %v", err)
	}

	// Выводим AST в JSON
	fmt.Println("=== AST (JSON) ===")
	jsonData, err := json.MarshalIndent(ast, "", "  ")
	if err != nil {
		log.Fatalf("Ошибка сериализации в JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
