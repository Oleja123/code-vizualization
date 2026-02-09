package main

import (
	"context"
	"fmt"
	"log"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
)

func printTree(node *sitter.Node, indent string) {
	fmt.Printf("%sNode: %s\n", indent, node.Type())
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		printTree(child, indent+"  ")
	}
}

func main() {
	sourceCode := []byte(`
if (score >= 90) {
	return 5;
} else if (score >= 80) {
	return 4;
} else {
	return 1;
}
`)

	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())
	
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Выводим дерево tree-sitter со структурой
	fmt.Println("=== Полная структура tree-sitter ===")
	
	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() == "if_statement" {
			fmt.Println("\nif_statement детально:")
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				fmt.Printf("[%d] %s: %q\n", j, subChild.Type(), subChild.Content(sourceCode))
			}
			
			// Разворачиваем alternative
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "alternative" {
					fmt.Printf("\nalternative детально:\n")
					for k := 0; k < int(subChild.ChildCount()); k++ {
						subSubChild := subChild.Child(k)
						fmt.Printf("  [%d] %s\n", k, subSubChild.Type())
						
						if subSubChild.Type() == "else_clause" {
							fmt.Printf("    else_clause детально:\n")
							for l := 0; l < int(subSubChild.ChildCount()); l++ {
								sssChild := subSubChild.Child(l)
								fmt.Printf("      [%d] %s: %q\n", l, sssChild.Type(), sssChild.Content(sourceCode)[:min(30, len(sssChild.Content(sourceCode)))])
								
								if sssChild.Type() == "if_statement" {
									fmt.Printf("        if_statement детально:\n")
									for m := 0; m < int(sssChild.ChildCount()); m++ {
										ssssChild := sssChild.Child(m)
										fmt.Printf("          [%d] %s\n", m, ssssChild.Type())
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
