package runtime

type Scope struct {
	Parent       *Scope
	Declarations DeclarationStack
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent: parent,
	}
}

func (sc *Scope) Declare(decl Declared) {
	sc.Declarations.Declare(decl)
}

func (sc *Scope) GetVariable(name string) (*Variable, bool) {
	return sc.Declarations.GetVariable(name)
}

func (sc *Scope) GetArray(name string) (*Array, bool) {
	return sc.Declarations.GetArray(name)
}

func (sc *Scope) GetArray2D(name string) (*Array2D, bool) {
	return sc.Declarations.GetArray2D(name)
}
