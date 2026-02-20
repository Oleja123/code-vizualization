package main

import (
	"fmt"
	"os"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/application/interpreter"
)

func main() {
	sourceCode := `
int main() {
    int a = 1;
	int b = 3;

    return a + b;
}
`

	conv := converter.New()
	program, convErr := conv.ParseToAST(sourceCode)
	if convErr != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", convErr)
		os.Exit(1)
	}

	runner := interpreter.NewInterpreter()
	result, err := runner.ExecuteProgram(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		os.Exit(1)
	}

	if result == nil {
		fmt.Println("program finished with no return value")
		return
	}

	fmt.Printf("program returned: %d\n", *result)
}
