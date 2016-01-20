package parser

import (
	"sort"
	"strings"
)

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

type Operation struct {
	Comment []string
	Name    string
	Param   string
	Scope   *Scope // Pointer back to containing Scope
}

// IncludeName returns the base include name of the parameter, if any.
func (o *Operation) IncludeName() string {
	if strings.Contains(o.Param, ".") {
		return o.Param[0:strings.Index(o.Param, ".")]
	}
	return ""
}

// ParamName returns the base parameter name with any include prefix removed.
func (o *Operation) ParamName() string {
	name := o.Param
	if strings.Contains(name, ".") {
		name = name[strings.Index(name, ".")+1:]
	}
	return name
}

type ScopePrefix struct {
	String    string
	Variables []string
}

func (n *ScopePrefix) Template() string {
	return prefixVariable.ReplaceAllString(n.String, "%s")
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
		if strings.Contains(op.Param, ".") {
			reducedStr := op.Param[0:strings.Index(op.Param, ".")]
			if _, ok := includesSet[reducedStr]; !ok {
				includesSet[reducedStr] = true
				includes = append(includes, reducedStr)
			}
		}
	}
	return includes
}

func (s *Scope) assignScope() {
	for _, op := range s.Operations {
		op.Scope = s
	}
}

type Async struct {
	Comment []string
	Name    string
	Extends string
	Methods []*Method
	Frugal  *Frugal // Pointer back to containing Frugal
}

type Frugal struct {
	Name           string
	Dir            string
	Path           string
	Scopes         []*Scope
	Asyncs         []*Async
	Thrift         *Thrift
	ParsedIncludes map[string]*Frugal
}

func (f *Frugal) NamespaceForInclude(include, lang string) (string, bool) {
	namespace, ok := f.ParsedIncludes[include].Thrift.Namespace(lang)
	return namespace, ok
}

func (f *Frugal) ContainsFrugalDefinitions() bool {
	return len(f.Scopes)+len(f.Thrift.Services) > 0
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code.
func (f *Frugal) ReferencedIncludes() []string {
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
	return includes
}

func (f *Frugal) UnderlyingType(t *Type) string {
	if t == nil {
		panic("frugal: Attempted to get underlying type of nil type")
	}
	typedefIndex := f.Thrift.typedefIndex
	include := t.IncludeName()
	if include != "" {
		typedefIndex = f.ParsedIncludes[include].Thrift.typedefIndex
	}
	if typedef, ok := typedefIndex[t.ParamName()]; ok {
		return typedef.Type.Name
	}
	return t.Name
}

func (f *Frugal) assignFrugal() {
	for _, scope := range f.Scopes {
		scope.assignScope()
	}
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
