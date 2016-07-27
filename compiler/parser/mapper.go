package parser

func makeFieldMap(f []*Field) (out map[int]*Field, e Error) {
	defer e.Prefix(" ")
	// map the fields according to their ID
	out = make(map[int]*Field)
	for i := range f {
		if _, ok := out[f[i].ID]; !ok {
			out[f[i].ID] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate IDs present %d: %s, %s", f[i].ID, f[i].Name, out[f[i].ID].Name))
		}
	}
	return out, e
}

func makeScopeMap(f []*Scope) (out map[string]*Scope, e Error) {
	defer e.Prefix(" ")
	// map the scopes according to their Name
	out = make(map[string]*Scope)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeOperationMap(f []*Operation) (out map[string]*Operation, e Error) {
	defer e.Prefix(" ")
	// map the operations according to their Name
	out = make(map[string]*Operation)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeStructureMap(f []*Struct) (out map[string]*Struct, e Error) {
	defer e.Prefix(" ")
	// map the structures according to their Name
	out = make(map[string]*Struct)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeServiceMap(f []*Service) (out map[string]*Service, e Error) {
	defer e.Prefix(" ")
	// map the services according to their Name
	out = make(map[string]*Service)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeMethodMap(f []*Method) (out map[string]*Method, e Error) {
	defer e.Prefix(" ")
	// map the methods according to their Name
	out = make(map[string]*Method)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeTypeDefMap(f []*TypeDef) (out map[string]*TypeDef, e Error) {
	defer e.Prefix(" ")
	// map the typedefs according to their Name
	out = make(map[string]*TypeDef)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeConstantMap(f []*Constant) (out map[string]*Constant, e Error) {
	defer e.Prefix(" ")
	// map the constants according to their Name
	out = make(map[string]*Constant)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeEnumMap(f []*Enum) (out map[string]*Enum, e Error) {
	defer e.Prefix(" ")
	// map the enums according to their Name
	out = make(map[string]*Enum)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeEnumValueMap(f []*EnumValue) (out map[int]*EnumValue, e Error) {
	defer e.Prefix(" ")
	// map the enumvalues according to their int value (because name can change values cannot)
	out = make(map[int]*EnumValue)
	for i := range f {
		if _, ok := out[f[i].Value]; !ok {
			out[f[i].Value] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate values present: %s", f[i].Name))
		}
	}
	return out, e
}

func makeNamespaceMap(f []*Namespace) (out map[string]*Namespace, e Error) {
	defer e.Prefix(" ")
	// map the namespaces according to their Scope
	out = make(map[string]*Namespace)
	for i := range f {
		if _, ok := out[f[i].Scope]; !ok {
			out[f[i].Scope] = f[i]
		} else {
			e.Append(NewErrorf("Duplicate scope present: %s", f[i].Scope))
		}
	}
	return out, e
}
