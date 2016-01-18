package parser

import (
	"fmt"
	"strings"
)

type Include struct {
	Name  string
	Value string
}

type Namespace struct {
	Scope string
	Value string
}

type Type struct {
	Name      string
	KeyType   *Type // If map
	ValueType *Type // If map, list, or set
}

// IncludeName returns the base include name of the type, if any.
func (t *Type) IncludeName() string {
	if strings.Contains(t.Name, ".") {
		return t.Name[0:strings.Index(t.Name, ".")]
	}
	return ""
}

// ParamName returns the base type name with any include prefix removed.
func (t *Type) ParamName() string {
	name := t.Name
	if strings.Contains(name, ".") {
		name = name[strings.Index(name, ".")+1:]
	}
	return name
}

func (t *Type) String() string {
	switch t.Name {
	case "map":
		return fmt.Sprintf("map<%s,%s>", t.KeyType.String(), t.ValueType.String())
	case "list":
		return fmt.Sprintf("list<%s>", t.ValueType.String())
	case "set":
		return fmt.Sprintf("set<%s>", t.ValueType.String())
	}
	return t.Name
}

type TypeDef struct {
	Comment []string
	Name    string
	Type    *Type
}

type EnumValue struct {
	Comment []string
	Name    string
	Value   int
}

type Enum struct {
	Comment []string
	Name    string
	Values  map[string]*EnumValue
}

type Constant struct {
	Comment []string
	Name    string
	Type    *Type
	Value   interface{}
}

type Field struct {
	Comment  []string
	ID       int
	Name     string
	Optional bool
	Type     *Type
	Default  interface{}
}

type Struct struct {
	Comment []string
	Name    string
	Fields  []*Field
}

type Method struct {
	Comment    []string
	Name       string
	Oneway     bool
	ReturnType *Type
	Arguments  []*Field
	Exceptions []*Field
}

type Service struct {
	Comment []string
	Name    string
	Extends string
	Methods []*Method
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code for this Service.
func (s *Service) ReferencedIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)
	for _, method := range s.Methods {
		for _, arg := range method.Arguments {
			if strings.Contains(arg.Type.Name, ".") {
				reducedStr := arg.Name[0:strings.Index(arg.Type.Name, ".")]
				if _, ok := includesSet[reducedStr]; !ok {
					includesSet[reducedStr] = true
					includes = append(includes, reducedStr)
				}
			}
		}
	}
	return includes
}

// ReferencedInternals returns a slice containing the referenced internals
// which will need to be imported in generated code for this Service.
// TODO: Clean this mess up
func (s *Service) ReferencedInternals() []string {
	internals := []string{}
	internalsSet := make(map[string]bool)
	for _, method := range s.Methods {
		for _, arg := range method.Arguments {
			if !strings.Contains(arg.Type.Name, ".") {
				// Check to see if it's a struct
				for _, param := range getImports(arg.Type) {
					if _, ok := internalsSet[param]; !ok {
						internalsSet[param] = true
						internals = append(internals, param)
					}
				}
			}
		}
	}
	return internals
}

type Thrift struct {
	Includes   []*Include
	Typedefs   []*TypeDef
	Namespaces []*Namespace
	Constants  []*Constant
	Enums      []*Enum
	Structs    []*Struct
	Exceptions []*Struct
	Unions     []*Struct
	Services   []*Service

	typedefIndex   map[string]*TypeDef
	namespaceIndex map[string]*Namespace
}

func (t *Thrift) Namespace(scope string) (string, bool) {
	namespace, ok := t.namespaceIndex[scope]
	value := ""
	if ok {
		value = namespace.Value
	}
	return value, ok
}

type Identifier string

type KeyValue struct {
	Key, Value interface{}
}

func (t *Thrift) NamespaceForInclude(include, lang string) (string, bool) {
	namespace, ok := t.namespaceIndex[lang]
	if !ok {
		return "", ok
	}
	return namespace.Value, ok
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code.
func (t *Thrift) ReferencedIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)
	for _, serv := range t.Services {
		for _, include := range serv.ReferencedIncludes() {
			if _, ok := includesSet[include]; !ok {
				includesSet[include] = true
				includes = append(includes, include)
			}
		}
	}
	return includes
}

// ReferencedInternals returns a slice containing the referenced internals
// which will need to be imported in generated code.
func (t *Thrift) ReferencedInternals() []string {
	internals := []string{}
	internalsSet := make(map[string]bool)
	for _, serv := range t.Services {
		for _, include := range serv.ReferencedInternals() {
			if _, ok := internalsSet[include]; !ok {
				internalsSet[include] = true
				internals = append(internals, include)
			}
		}
	}
	return internals
}

func getImports(t *Type) []string {
	list := []string{}
	switch t.Name {
	case "bool":
	case "byte":
	case "i16":
	case "i32":
	case "i64":
	case "double":
	case "string":
	case "binary":
	case "list":
		for _, imp := range getImports(t.ValueType) {
			list = append(list, imp)
		}
	case "set":
		for _, imp := range getImports(t.ValueType) {
			list = append(list, imp)
		}
	case "map":
		for _, imp := range getImports(t.KeyType) {
			list = append(list, imp)
		}
		for _, imp := range getImports(t.ValueType) {
			list = append(list, imp)
		}
		return list
	default:
		list = append(list, t.Name)
	}
	return list
}
