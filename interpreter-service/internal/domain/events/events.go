package events

import "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime"

type Event interface{}

type EnterScope struct{}

type ExitScope struct{}

type DeclareVar struct {
	Name     string `json:"name"`
	Value    *int   `json:"value"`
	IsGlobal bool   `json:"isGlobal"`
}

type DeclareArray struct {
	Name     string                 `json:"name"`
	Value    []runtime.ArrayElement `json:"value"`
	Size     int                    `json:"size"`
	IsGlobal bool                   `json:"isGlobal"`
}

type DeclareArray2D struct {
	Name     string          `json:"name"`
	Value    []runtime.Array `json:"value"`
	Size1    int             `json:"size1"`
	Size2    int             `json:"size2"`
	IsGlobal bool            `json:"isGlobal"`
}

type VarChanged struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type ArrayElementChanged struct {
	Name  string `json:"name"`
	Ind   int    `json:"ind"`
	Value int    `json:"value"`
}

type Array2DElementChanged struct {
	Name  string `json:"name"`
	Ind1  int    `json:"ind1"`
	Ind2  int    `json:"ind2"`
	Value int    `json:"value"`
}

type FunctionCall struct {
	Name string `json:"name"`
	// каждый parameter в ast будет порождать DeclareVar
}

type FunctionReturn struct {
	Name        string `json:"name"`
	ReturnValue *int   `json:"returnValue"` // nil если void функция
}

type LineChanged struct {
	Line int `json:"line"`
}

type UndefinedBehavior struct {
	Message string `json:"message"`
}

type RuntimeError struct {
	Message string `json:"message"`
}
