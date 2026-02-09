package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
	// Пример с массивами, указателями, циклами
	sourceCode := []byte(`
int sum_array(int *arr, int size) {
	int total = 0;
	for (int i = 0; i < size; i++) {
		total = total + arr[i];
	}
	return total;
}

int main() {
	int numbers[5];
	int i = 0;
	
	while (i < 5) {
		numbers[i] = i * 2;
		i = i + 1;
	}
	
	int *ptr = &numbers[0];
	int result = sum_array(ptr, 5);
	
	return result;
}
`)

	fmt.Println("=== Исходный код C (массивы, указатели, циклы) ===")
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
