package scopecollection

import (
	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/loop"
)

type ScopeInfo struct {
	ParentId int
	LoopCtx  *loop.LoopContext
	BlockAST *converter.BlockStmt
}

type ScopeCollection struct {
	ScopeIds  map[*converter.BlockStmt]int
	nextId    int
	ScopeInfo map[int]ScopeInfo
}

func NewScopeCollection() *ScopeCollection {
	return &ScopeCollection{
		ScopeIds:  make(map[*converter.BlockStmt]int),
		ScopeInfo: make(map[int]ScopeInfo),
		nextId:    0,
	}
}

func (sc *ScopeCollection) incrementId() {
	sc.nextId++
}

func (sc *ScopeCollection) EnsureScope(node *converter.BlockStmt, parentId int, loopCtx *loop.LoopContext) int {
	defer sc.incrementId()
	sc.ScopeIds[node] = sc.nextId
	sc.ScopeInfo[sc.nextId] = ScopeInfo{ParentId: parentId, LoopCtx: loopCtx, BlockAST: node}
	return sc.nextId
}

func (sc *ScopeCollection) GetScopeId(node *converter.BlockStmt) (int, bool) {
	id, ok := sc.ScopeIds[node]
	return id, ok
}

func (sc *ScopeCollection) GetScopeInfo(id int) (ScopeInfo, bool) {
	info, ok := sc.ScopeInfo[id]
	return info, ok
}
