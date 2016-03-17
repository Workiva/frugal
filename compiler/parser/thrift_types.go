package parser

import (
	"fmt"
	"strings"
)

var thriftTypes = map[string]bool{
	"bool":   true,
	"byte":   true,
	"i8":     true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"double": true,
	"string": true,
	"binary": true,
}

// Thrift Field Modifiers
type FieldModifier int

const (
	Required FieldModifier = iota
	Optional
	Default
)

func IsThriftPrimitive(typ *Type) bool {
	_, ok := thriftTypes[typ.Name]
	return ok
}

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
	Modifier FieldModifier
	Type     *Type
	Default  interface{}
}

type StructType int

func (s StructType) String() string {
	switch s {
	case StructTypeStruct:
		return "struct"
	case StructTypeException:
		return "exception"
	case StructTypeUnion:
		return "union"
	default:
		panic(fmt.Sprintf("unknown struct type %d", s))
	}
}

const (
	StructTypeStruct = iota
	StructTypeException
	StructTypeUnion
)

type Struct struct {
	Comment []string
	Name    string
	Fields  []*Field
	Type    StructType
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

func (s *Service) ExtendsInclude() string {
	includeAndService := strings.Split(s.Extends, ".")
	if len(includeAndService) == 2 {
		return includeAndService[0]
	}
	return ""
}

func (s *Service) ExtendsService() string {
	includeAndService := strings.Split(s.Extends, ".")
	if len(includeAndService) == 2 {
		return includeAndService[1]
	}
	return s.Extends
}

// TwowayMethods returns a slice of the non-oneway methods defined in this
// Service.
func (s *Service) TwowayMethods() []*Method {
	methods := make([]*Method, 0, len(s.Methods))
	for _, method := range s.Methods {
		if !method.Oneway {
			methods = append(methods, method)
		}
	}
	return methods
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code for this Service.
func (s *Service) ReferencedIncludes() []string {
	includes := []string{}
	includesSet := make(map[string]bool)

	// Check extended service.
	if s.Extends != "" && strings.Contains(s.Extends, ".") {
		reducedStr := s.Extends[0:strings.Index(s.Extends, ".")]
		if _, ok := includesSet[reducedStr]; !ok {
			includesSet[reducedStr] = true
			includes = append(includes, reducedStr)
		}
	}

	// Check methods.
	for _, method := range s.Methods {
		// Check arguments.
		for _, arg := range method.Arguments {
			includesSet, includes = addInclude(includesSet, includes, arg.Type)
		}
		// Check return type.
		if method.ReturnType != nil {
			includesSet, includes = addInclude(includesSet, includes, method.ReturnType)
		}
	}

	return includes
}

// addInclude checks the given Type and adds any includes for it to the given
// map and slice, returning the new map and slice.
func addInclude(includesSet map[string]bool, includes []string, t *Type) (map[string]bool, []string) {
	if strings.Contains(t.Name, ".") {
		reducedStr := t.Name[0:strings.Index(t.Name, ".")]
		if _, ok := includesSet[reducedStr]; !ok {
			includesSet[reducedStr] = true
			includes = append(includes, reducedStr)
		}
	}
	// Check container types.
	if t.KeyType != nil {
		includesSet, includes = addInclude(includesSet, includes, t.KeyType)
	}
	if t.ValueType != nil {
		includesSet, includes = addInclude(includesSet, includes, t.ValueType)
	}
	return includesSet, includes
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

func (s *Service) validate() error {
	for _, method := range s.Methods {
		if method.Oneway {
			if len(method.Exceptions) > 0 {
				return fmt.Errorf("Oneway method %s.%s cannot throw an exception",
					s.Name, method.Name)
			}
			if method.ReturnType != nil {
				return fmt.Errorf("Void method %s.%s cannot return %s",
					s.Name, method.Name, method.ReturnType)
			}
		}
	}
	return nil
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
	} else {
		namespace, ok = t.namespaceIndex["*"]
		if ok {
			value = namespace.Value
		}
	}
	return value, ok
}

type Identifier string

type KeyValue struct {
	Key, Value interface{}
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

func (t *Thrift) validate() error {
	if err := t.validateIncludes(); err != nil {
		return err
	}
	if err := t.validateServices(); err != nil {
		return err
	}
	return nil
}

func (t *Thrift) validateIncludes() error {
	includes := map[string]struct{}{}
	for _, include := range t.Includes {
		if _, ok := includes[include.Name]; ok {
			return fmt.Errorf("Duplicate include: %s", include.Name)
		}
		includes[include.Name] = struct{}{}
	}
	return nil
}

func (t *Thrift) validateServices() error {
	for _, service := range t.Services {
		if err := service.validate(); err != nil {
			return err
		}
	}
	return nil
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
