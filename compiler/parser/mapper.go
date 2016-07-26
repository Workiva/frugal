package parser

import (
	"strconv"
)

func makeFieldMap(f []*Field) (out map[int]*Field, err Error) {
	// map the fields according to their ID
	out = make(map[int]*Field)
	for i := range f {
		if _, ok := out[f[i].ID]; !ok {
			out[f[i].ID] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate IDs present %s: %s, %s", strconv.Itoa(f[i].ID), f[i].Name, out[f[i].ID].Name))
		}
	}
	return out, err
}

func makeScopeMap(f []*Scope) (out map[string]*Scope, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Scope)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeOperationMap(f []*Operation) (out map[string]*Operation, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Operation)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeStructureMap(f []*Struct) (out map[string]*Struct, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Struct)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeServiceMap(f []*Service) (out map[string]*Service, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Service)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeMethodMap(f []*Method) (out map[string]*Method, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Method)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeTypeDefMap(f []*TypeDef) (out map[string]*TypeDef, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*TypeDef)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeConstantMap(f []*Constant) (out map[string]*Constant, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Constant)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeEnumMap(f []*Enum) (out map[string]*Enum, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Enum)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeEnumValueMap(f []*EnumValue) (out map[string]*EnumValue, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*EnumValue)
	for i := range f {
		if _, ok := out[f[i].Name]; !ok {
			out[f[i].Name] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Name))
		}
	}
	return out, err
}

func makeNamespaceMap(f []*Namespace) (out map[string]*Namespace, err Error) {
	// map the scopes according to their Name
	out = make(map[string]*Namespace)
	for i := range f {
		if _, ok := out[f[i].Scope]; !ok {
			out[f[i].Scope] = f[i]
		} else {
			err.Append(NewErrorf("Duplicate names present: %s", f[i].Scope))
		}
	}
	return out, err
}
