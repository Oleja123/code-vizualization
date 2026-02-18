package loop

import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

type LoopContext struct {
	ContinueTarget converter.Stmt
	BreakTarget    converter.Stmt
}
