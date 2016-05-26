package parser

import (
	"fmt"
	"sort"
)

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

type Operation struct {
	Comment []string
	Name    string
	Type    *Type
	Scope   *Scope // Pointer back to containing Scope
}

type ScopePrefix struct {
	String    string
	Variables []string
}

func (n *ScopePrefix) Template(s string) string {
	return prefixVariable.ReplaceAllString(n.String, s)
}

type Scope struct {
	Comment    []string
	Name       string
	Prefix     *ScopePrefix
	Operations []*Operation
	Frugal     *Frugal // Pointer back to containing Frugal
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code for this Scope.
func (s *Scope) ReferencedIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)
	for _, op := range s.Operations {
		includesSet, includes = addInclude(includesSet, includes, op.Type)
	}
	return includes
}

func (s *Scope) assignScope() {
	for _, op := range s.Operations {
		op.Scope = s
	}
}

type Frugal struct {
	Name           string
	File           string
	Dir            string
	Path           string
	Scopes         []*Scope
	Thrift         *Thrift
	ParsedIncludes map[string]*Frugal

	// This retains a list of all definitions in the order they are defined in
	// the IDL. The Thrift compiler has several bugs which make the generated
	// code sensitive to the ordering of IDL definitions, e.g.
	// https://issues.apache.org/jira/browse/THRIFT-3465
	// This is used to retain ordering in the generated intermediate Thrift.
	// TODO: Remove this once the dependency on Thrift is removed or the bugs
	// in Thrift are fixed.
	OrderedDefinitions []interface{}
}

func (f *Frugal) NamespaceForInclude(include, lang string) (string, bool) {
	parsed, ok := f.ParsedIncludes[include]
	if !ok {
		return "", ok
	}
	namespace, ok := parsed.Thrift.Namespace(lang)
	return namespace, ok
}

func (f *Frugal) ContainsFrugalDefinitions() bool {
	return len(f.Scopes)+len(f.Thrift.Services) > 0
}

// ReferencedScopeIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for scopes.
func (f *Frugal) ReferencedScopeIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)
	for _, scope := range f.Scopes {
		for _, include := range scope.ReferencedIncludes() {
			if _, ok := includesSet[include]; !ok {
				includesSet[include] = true
				includes = append(includes, include)
			}
		}
	}
	sort.Strings(includes)
	return includes
}

// ReferencedServiceIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for services.
func (f *Frugal) ReferencedServiceIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)
	for _, service := range f.Thrift.Services {
		for _, include := range service.ReferencedIncludes() {
			if _, ok := includesSet[include]; !ok {
				includesSet[include] = true
				includes = append(includes, include)
			}
		}
	}
	sort.Strings(includes)
	return includes
}

// UnderlyingType follows any typedefs to get the base IDL type.
func (f *Frugal) UnderlyingType(t *Type) *Type {
	if t == nil {
		panic("Attempted to get underlying type of nil type")
	}
	typedefIndex := f.Thrift.typedefIndex
	include := t.IncludeName()
	if include != "" {
		parsed, ok := f.ParsedIncludes[include]
		if !ok {
			return t
		}
		typedefIndex = parsed.Thrift.typedefIndex
	}
	if typedef, ok := typedefIndex[t.ParamName()]; ok {
		return typedef.Type
	}
	return t
}

func (f *Frugal) ConstantFromField(field *Field, value interface{}) *Constant {
	return &Constant{
		Name:  field.Name,
		Type:  field.Type,
		Value: value,
	}
}

// IsStruct indicates if the underlying Type is a struct.
func (f *Frugal) IsStruct(t *Type) bool {
	t = f.UnderlyingType(t)
	if _, ok := thriftBaseTypes[t.Name]; ok {
		return false
	}
	return t.KeyType == nil && t.ValueType == nil && !f.IsEnum(t)
}

// IsEnum indicates if the underlying Type is an enum.
func (f *Frugal) IsEnum(t *Type) bool {
	include := t.IncludeName()
	containingFrugal := f
	if include != "" {
		if containing, ok := f.ParsedIncludes[include]; ok {
			containingFrugal = containing
		}
	}
	for _, enum := range containingFrugal.Thrift.Enums {
		if enum.Name == t.ParamName() {
			return true
		}
	}
	return false
}

func (f *Frugal) assignFrugal() {
	for _, scope := range f.Scopes {
		scope.assignScope()
	}
}

func (f *Frugal) validate() error {
	// Ensure there are no duplicate names between services and scopes.
	names := make(map[string]struct{})
	for _, service := range f.Thrift.Services {
		if _, ok := names[service.Name]; ok {
			return fmt.Errorf("Duplicate service name %s", service.Name)
		}
		names[service.Name] = struct{}{}
	}
	for _, scope := range f.Scopes {
		if _, ok := names[scope.Name]; ok {
			return fmt.Errorf("Duplicate scope name %s", scope.Name)
		}
		names[scope.Name] = struct{}{}
	}
	return f.Thrift.validate(f.ParsedIncludes)
}

func (f *Frugal) sort() {
	sort.Sort(ScopesByName(f.Scopes))
}

type ScopesByName []*Scope

func (b ScopesByName) Len() int {
	return len(b)
}

func (b ScopesByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ScopesByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
