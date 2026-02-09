package main

import (
	"fmt"
	"log"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
	sourceCode := []byte(`
int grade(int score) {
	if (score >= 90) {
		return 5;
	} else if (score >= 80) {
		return 4;
	} else {
		return 1;
	}
}
`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Выводим дерево tree-sitter
	fmt.Println("=== Tree-sitter CST ===")
	fmt.Println(tree.RootNode().String())
}
