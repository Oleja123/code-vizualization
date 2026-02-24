package runtime

type Declared interface{}

type DeclarationStack struct {
	Declarations []Declared
}

func (ds *DeclarationStack) Declare(d Declared) {
	ds.Declarations = append(ds.Declarations, d)
}

func (ds *DeclarationStack) GetVariable(name string) (*Variable, bool) {
	for i := range ds.Declarations {
		if v, ok := ds.Declarations[i].(*Variable); ok {
			if v.Name == name {
				return v, true
			}
		}
	}
	return nil, false
}

func (ds *DeclarationStack) GetArray(name string) (*Array, bool) {
	for i := range ds.Declarations {
		if v, ok := ds.Declarations[i].(*Array); ok {
			if v.Name == name {
				return v, true
			}
		}
	}
	return nil, false
}

func (ds *DeclarationStack) GetArray2D(name string) (*Array2D, bool) {
	for i := range ds.Declarations {
		if v, ok := ds.Declarations[i].(*Array2D); ok {
			if v.Name == name {
				return v, true
			}
		}
	}
	return nil, false
}
