package events

type Event interface{}

type EnterScope struct{}

type ExitScope struct{}

type DeclareVar struct {
	Name     string
	Value    *int
	IsGlobal bool
}

type DeclareArray struct {
	Name     string
	Value    []int
	Size     int
	IsGlobal bool
}

type DeclareArray2D struct {
	Name     string
	Value    [][]int
	Size1    int
	Size2    int
	IsGlobal bool
}

type VarChanged struct {
	Name  string
	Value int
}

type ArrayElementChanged struct {
	Name  string
	Ind   int
	Value int
}

type Array2DElementChanged struct {
	Name  string
	Ind1  int
	Ind2  int
	Value int
}

type FunctionCall struct {
	Name string
	// каждый parameter в ast будет порождать DeclareVar
}

type FunctionReturn struct {
	Name        string
	ReturnValue *int // nil если void функция
}

type LineChanged struct {
	Line int
}
