# Getting Started - CST-to-AST Converter

## Quick Overview

This project converts C code parsed by tree-sitter into an Abstract Syntax Tree (AST), suitable for educational interpreters and code visualization tools.

## Installation

```bash
cd cst-to-ast-service
go mod tidy
```

## Run Examples

```bash
# Example 1: Recursive factorial function
go run cmd/example/main.go

# Example 2: Arrays, pointers, and loops
go run cmd/advanced-example/main.go

# Example 3: Else if chains (NEW!)
go run cmd/else-if-example/main.go
```

## What's Included

### Core Implementation
- **converter.go**: Main conversion engine (780+ lines)
- **structs/ast.go**: 18 AST node types (10 statements + 8 expressions)
- **interfaces/interfaces.go**: Type-safe interface definitions

### Examples & Debugging
- **cmd/example/**: Simple recursion example
- **cmd/advanced-example/**: Arrays, pointers, and loop constructs
- **cmd/else-if-example/**: Else if chain demonstration (NEW!)
- **cmd/debug-cst/**: Analyze tree-sitter CST structure
- **cmd/debug-else-if/**: Debug else if parsing

### Documentation
- **README.md**: Project overview
- **ARCHITECTURE.md**: Detailed architecture (500+ lines)
- **USAGE.md**: Complete usage guide (400+ lines)
- **PROJECT_COMPLETION_REPORT.md**: Implementation summary

## Supported C Constructs

### Data Types
- `int` with pointers (`int*`, `int**`)
- Arrays (`int[N]`)

### Statements
- Variable declarations: `int x = 10;`
- Function definitions: `int add(int a, int b) { ... }`
- Conditionals: `if { } else if { } else { }`
- Loops: `while { }`, `for { }`
- Control flow: `return`, `break`, `continue`

### Expressions
- Binary operations: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`
- Unary operations: `-`, `!`, `*`, `&`
- Function calls: `factorial(5)`
- Array access: `arr[i]`
- Assignment: `x = 10`

## Program Usage

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
    sourceCode := []byte(`
        int add(int a, int b) {
            return a + b;
        }
    `)
    
    conv := converter.NewCConverter()
    
    // Parse to CST
    tree, err := conv.Parse(sourceCode)
    if err != nil {
        log.Fatal(err)
    }
    
    // Convert to AST
    ast, err := conv.ConvertToProgram(tree, sourceCode)
    if err != nil {
        log.Fatal(err)
    }
    
    // Serialize to JSON
    jsonData, _ := json.MarshalIndent(ast, "", "  ")
    fmt.Println(string(jsonData))
}
```

## Key Features

### ✅ Complete AST Representation
- 18 different node types
- Each node includes Location (line, column, end position)
- Full support for nested structures

### ✅ Explicit Else If Support (NEW!)
```go
type ElseIfClause struct {
    Condition interfaces.Expr
    Block     interfaces.Stmt
    Location  Location
}

type IfStmt struct {
    Condition  interfaces.Expr
    ThenBlock  interfaces.Stmt
    ElseIfList []ElseIfClause  // Explicit!
    ElseBlock  interfaces.Stmt
    Location   Location
}
```

### ✅ Type-Safe Design
- Marker methods for interface enforcement
- Clear separation of concerns
- Domain-driven architecture

## Documentation

For detailed information:
- **Architecture details**: See [ARCHITECTURE.md](ARCHITECTURE.md)
- **Usage guide**: See [USAGE.md](USAGE.md)
- **Implementation report**: See [PROJECT_COMPLETION_REPORT.md](PROJECT_COMPLETION_REPORT.md)

## Example AST Output

### Input Code
```c
int grade(int score) {
    if (score >= 90) {
        return 5;
    } else if (score >= 80) {
        return 4;
    } else {
        return 3;
    }
}
```

### Output Structure
```json
{
    "type": "IfStmt",
    "condition": {
        "type": "BinaryExpr",
        "operator": ">=",
        "left": {"type": "Identifier", "name": "score"},
        "right": {"type": "IntLiteral", "value": 90}
    },
    "thenBlock": {...},
    "elseIf": [
        {
            "condition": {
                "type": "BinaryExpr",
                "operator": ">=",
                "left": {"type": "Identifier", "name": "score"},
                "right": {"type": "IntLiteral", "value": 80}
            },
            "block": {...}
        }
    ],
    "elseBlock": {...}
}
```

## Project Structure

```
cst-to-ast-service/
├── cmd/
│   ├── example/              # Factorial recursion
│   ├── advanced-example/     # Arrays, pointers, loops
│   ├── else-if-example/      # Else if chains
│   ├── debug-cst/           # CST analyzer
│   └── debug-else-if/       # Else if debugger
├── internal/
│   ├── converter/
│   │   └── converter.go      # Main conversion engine
│   └── domain/
│       ├── interfaces/
│       │   └── interfaces.go # Type interfaces
│       └── structs/
│           └── ast.go        # AST node definitions
├── go.mod
└── go.sum
```

## Testing

All examples compile and run successfully:

```bash
go run cmd/example/main.go > /dev/null && echo "✅ Works"
go run cmd/advanced-example/main.go > /dev/null && echo "✅ Works"
go run cmd/else-if-example/main.go > /dev/null && echo "✅ Works"
```

## Dependencies

- Go 1.21+
- `github.com/smacker/go-tree-sitter` for C parsing

## Limitations

- **Type system**: Only `int` type supported (by design)
- **Operators**: No `switch`, `do-while`, `goto`
- **Preprocessor**: No `#include`, `#define` support
- **Comments**: Not included in AST

## Next Steps

1. Review [ARCHITECTURE.md](ARCHITECTURE.md) for implementation details
2. Run the examples to see AST output
3. Integrate into your educational interpreter
4. Extend with additional C constructs as needed

## Support

For detailed questions:
- Check [USAGE.md](USAGE.md) FAQ section
- Review example code in `cmd/` directory
- Analyze AST structure with debug tools

---

**Project Status**: ✅ Complete and Production-Ready
**Latest Version**: 1.0.0
**Last Updated**: Today
