package main

import (
	"fmt"
	"log"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/interpreter"
)

const factorialProgram = `int bebra() {
  int a;
  return a + 1;
}

int main() {
  return bebra();
}
`

func main() {
	conv := converter.New()
	program, parseErr := conv.ParseToAST(factorialProgram)
	if parseErr != nil {
		log.Fatalf("parse error: %v", parseErr)
	}

	runner := interpreter.NewInterpreter()
	result, steps, _, runtimeErr := runner.ExecuteProgram(program)
	_ = steps
	if runtimeErr != nil {
		log.Fatalf("runtime error: %v", runtimeErr)
	}

	if result == nil {
		fmt.Println("Program finished. main returned: <nil>")
		return
	}

	fmt.Printf("Program finished. main returned: %d\n", *result)
}
