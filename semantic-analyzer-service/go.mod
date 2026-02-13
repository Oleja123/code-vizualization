module github.com/Oleja123/code-vizualization/semantic-analyzer-service

go 1.25.6

require (
	github.com/Oleja123/code-vizualization/cst-to-ast-service v0.0.0
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Oleja123/code-vizualization/cst-to-ast-service => ../cst-to-ast-service
