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
