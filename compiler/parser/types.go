package parser

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

var frugalBaseTypes = map[string]bool{
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

var frugalContainerTypes = map[string]bool{
	"list": true,
	"set":  true,
	"map":  true,
}

// FieldModifier represents a Frugal IDL field modifier (required, optional
// default).
type FieldModifier int

const (
	// Required indicates always written for the writer and must read or error
	// for the reader.
	Required FieldModifier = iota

	// Optional indicates written if set for the writer and read if present for
	// the reader.
	Optional

	// Default indicates always written for the writer and read if present for
	// the reader.
	Default
)

func (f *FieldModifier) String() string {
	switch *f {
	case Required:
		return "REQUIRED"
	case Optional:
		return "OPTIONAL"
	case Default:
		return "DEFAULT"
	default:
		panic(fmt.Sprintf("unsupported modifier: %v", *f))
	}
}

// FieldFromType returns a new Field from the given Type and name.
func FieldFromType(t *Type, name string) *Field {
	return &Field{
		Comment:  nil,
		ID:       0,
		Name:     name,
		Modifier: Required,
		Type:     t,
		Default:  nil,
	}
}

// TypeFromStruct returns a new Type from the given Struct.
func TypeFromStruct(s *Struct) *Type {
	return &Type{
		Name:      s.Name,
		KeyType:   nil,
		ValueType: nil,
	}
}

// Include represents an IDL file include.
type Include struct {
	Name        string
	Value       string
	Annotations Annotations
}

// Namespace represents an IDL namespace.
type Namespace struct {
	Scope       string
	Value       string
	Annotations Annotations
}

// Wildcard indicates if this Namespace is a wildcard (*).
func (n *Namespace) Wildcard() bool {
	return n.Scope == "*"
}

// Type represents an IDL data type.
type Type struct {
	Name        string
	KeyType     *Type // If map
	ValueType   *Type // If map, list, or set
	Annotations Annotations
}

// IsPrimitive indicates if the type is a Frugal primitive type.
func (t *Type) IsPrimitive() bool {
	_, ok := frugalBaseTypes[t.Name]
	return ok
}

// IsCustom indicates if the type is not a container or primitive type.
func (t *Type) IsCustom() bool {
	return !t.IsPrimitive() && !t.IsContainer()
}

// IsContainer indicates if the type is a Frugal container type (list, set, or
// map).
func (t *Type) IsContainer() bool {
	_, ok := frugalContainerTypes[t.Name]
	return ok
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

// String returns a human-readable version of the Type.
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

// TypeDef represents an IDL typedef.
type TypeDef struct {
	Comment     []string
	Name        string
	Type        *Type
	Annotations Annotations
}

// EnumValue represents an IDL enum value.
type EnumValue struct {
	Comment     []string
	Name        string
	Value       int
	Annotations Annotations
}

// Enum represents an IDL enum.
type Enum struct {
	Comment     []string
	Name        string
	Values      []*EnumValue
	Annotations Annotations
}

// Constant represents an IDL constant.
type Constant struct {
	Comment     []string
	Name        string
	Type        *Type
	Value       interface{}
	Annotations Annotations
}

// Field represents an IDL field on a struct or method.
type Field struct {
	Comment     []string
	ID          int
	Name        string
	Modifier    FieldModifier
	Type        *Type
	Default     interface{}
	Annotations Annotations
}

// StructType represents what "type" a struct is (struct, exception, or union).
type StructType int

// String returns a human-readable version of the StructType.
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

// Valid StructTypes.
const (
	StructTypeStruct = iota
	StructTypeException
	StructTypeUnion
)

// Struct represents an IDL struct (or exception or union).
type Struct struct {
	Comment     []string
	Name        string
	Fields      []*Field
	Type        StructType
	Annotations Annotations
}

// Method represents an IDL service method.
type Method struct {
	Comment     []string
	Name        string
	Oneway      bool
	ReturnType  *Type
	Arguments   []*Field
	Exceptions  []*Field
	Annotations Annotations
}

// Service represents an IDL service.
type Service struct {
	Comment     []string
	Name        string
	Extends     string
	Methods     []*Method
	Annotations Annotations
	Frugal      *Frugal // Pointer back to containing Frugal
}

// ExtendsInclude returns the name of the include this service extends from, if
// applicable, or an empty string if not.
func (s *Service) ExtendsInclude() string {
	includeAndService := strings.Split(s.Extends, ".")
	if len(includeAndService) == 2 {
		return includeAndService[0]
	}
	return ""
}

// ExtendsService returns the name of the service this service extends, if
// applicable, or an empty string if not.
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
func (s *Service) ReferencedIncludes() ([]*Include, error) {
	var err error
	includes := []*Include{}
	includesSet := make(map[string]*Include)

	// Check extended service.
	if s.Extends != "" && strings.Contains(s.Extends, ".") {
		includeName := s.Extends[0:strings.Index(s.Extends, ".")]
		include := s.Frugal.Include(includeName)
		if include == nil {
			return nil, fmt.Errorf("Service %s extends references invalid include %s",
				s.Name, s.Extends)
		}
		if _, ok := includesSet[includeName]; !ok {
			includesSet[includeName] = include
			includes = append(includes, include)
		}
	}

	// Check methods.
	for _, method := range s.Methods {
		// Check arguments.
		for _, arg := range method.Arguments {
			includesSet, includes, err = addInclude(includesSet, includes, arg.Type, s.Frugal)
			if err != nil {
				return nil, err
			}
		}
		// Check return type.
		if method.ReturnType != nil {
			includesSet, includes, err = addInclude(includesSet, includes, method.ReturnType, s.Frugal)
		}
		if err != nil {
			return nil, err
		}
		// Check exceptions.
		for _, exception := range method.Exceptions {
			includesSet, includes, err = addInclude(includesSet, includes, exception.Type, s.Frugal)
		}
		if err != nil {
			return nil, err
		}
	}

	return includes, nil
}

// addInclude checks the given Type and adds any includes for it to the given
// map and slice, returning the new map and slice.
func addInclude(includesSet map[string]*Include, includes []*Include, t *Type, frugal *Frugal) (map[string]*Include, []*Include, error) {
	var err error
	if strings.Contains(t.Name, ".") {
		includeName := t.Name[0:strings.Index(t.Name, ".")]
		include := frugal.Include(includeName)
		if include == nil {
			return nil, nil, fmt.Errorf("Type %s references invalid include %s", t.Name, include.Name)
		}
		if _, ok := includesSet[includeName]; !ok {
			includesSet[includeName] = include
			includes = append(includes, include)
		}
	}
	// Check container types.
	if t.KeyType != nil {
		includesSet, includes, err = addInclude(includesSet, includes, t.KeyType, frugal)
	}
	if err != nil {
		return nil, nil, err
	}
	if t.ValueType != nil {
		includesSet, includes, err = addInclude(includesSet, includes, t.ValueType, frugal)
	}
	return includesSet, includes, err
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

// validate ensures Service oneways don't return anything and field ids aren't
// duplicated.
func (s *Service) validate() error {
	for _, method := range s.Methods {
		// Ensure oneways don't return anything.
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

		// Ensure field ids aren't duplicated.
		ids := make(map[int]struct{})
		for _, arg := range method.Arguments {
			if _, ok := ids[arg.ID]; ok {
				return fmt.Errorf("Duplicate field id %d in method %s.%s",
					arg.ID, s.Name, method.Name)
			}
			ids[arg.ID] = struct{}{}
		}
	}
	return nil
}

// Identifier represents an IDL identifier.
type Identifier string

// KeyValue is a key-value pair.
type KeyValue struct {
	Key, Value interface{}
}

// Annotation is key-value metadata attached to an IDL definition.
type Annotation struct {
	Name  string
	Value string
}

// Annotations is the collection of Annotations present on an IDL definition.
type Annotations []*Annotation

// Get returns true if the given annotation name is present and its associated
// value, if any.
func (a Annotations) Get(name string) (string, bool) {
	for _, annotation := range a {
		if annotation.Name == name {
			return annotation.Value, true
		}
	}

	return "", false
}

// Vendor returns true if the "vendor" annotation is present and its associated
// value, if any.
func (a Annotations) Vendor() (string, bool) {
	return a.Get(VendorAnnotation)
}

// Deprecated returns true if the "deprecated" annotation is present and its
// associated value, if any.
func (a Annotations) Deprecated() (string, bool) {
	return a.Get(DeprecatedAnnotation)
}

func getImports(t *Type) []string {
	list := []string{}
	switch t.Name {
	case "bool":
	case "byte":
	case "i8":
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

// LowercaseFirstLetter of the string.
func LowercaseFirstLetter(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Operation is a pub/sub scope operation. Corresponding publish and
// subscribe methods are generated from this for publishers and subscribers,
// respectively.
type Operation struct {
	Comment     []string
	Name        string
	Type        *Type
	Annotations Annotations
	Scope       *Scope // Pointer back to containing Scope
}

// ScopePrefix is the string prefix prepended to a pub/sub topic. The string
// can contain variables of the form {foo}, e.g. "foo.{bar}.baz" where "bar"
// is supplied at publish/subscribe time.
type ScopePrefix struct {
	String    string
	Variables []string
}

// Template returns the prefix where variables are replaced with the given
// string.
func (n *ScopePrefix) Template(s string) string {
	return prefixVariable.ReplaceAllString(n.String, s)
}

// Scope is a pub/sub namespace.
type Scope struct {
	Comment     []string
	Name        string
	Prefix      *ScopePrefix
	Operations  []*Operation
	Annotations Annotations
	Frugal      *Frugal // Pointer back to containing Frugal
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code for this Scope.
func (s *Scope) ReferencedIncludes() ([]*Include, error) {
	var err error
	includes := []*Include{}
	includesSet := make(map[string]*Include)
	for _, op := range s.Operations {
		includesSet, includes, err = addInclude(includesSet, includes, op.Type, s.Frugal)
		if err != nil {
			return nil, err
		}
	}
	return includes, nil
}

func (s *Scope) assignScope() {
	for _, op := range s.Operations {
		op.Scope = s
	}
}

// IdentifierType indicates if a Identifier is a local/include constant/enum
type IdentifierType uint8

const (
	NonIdentifier IdentifierType = iota
	LocalConstant
	LocalEnum
	IncludeConstant
	IncludeEnum
)

// IdentifierContext gives information about the identifier for a given
// contextual reference
type IdentifierContext struct {
	Type      IdentifierType
	Constant  *Constant
	Enum      *Enum
	EnumValue *EnumValue
	Include   *Frugal
}

// Frugal contains the complete IDL parse tree.
type Frugal struct {
	Name           string
	File           string
	Dir            string
	Path           string
	ParsedIncludes map[string]*Frugal

	Includes   []*Include
	Typedefs   []*TypeDef
	Namespaces []*Namespace
	Constants  []*Constant
	Enums      []*Enum
	Structs    []*Struct
	Exceptions []*Struct
	Unions     []*Struct
	Services   []*Service
	Scopes     []*Scope

	typedefIndex   map[string]*TypeDef
	namespaceIndex map[string]*Namespace
}

// Namespace returns namespace value for the given scope.
func (f *Frugal) Namespace(scope string) *Namespace {
	namespace := f.namespaceIndex[scope]
	if namespace != nil {
		return namespace
	}
	return f.namespaceIndex["*"]
}

// Include returns the Include with the given name.
func (f *Frugal) Include(name string) *Include {
	name = filepath.Base(name)
	for _, include := range f.Includes {
		if filepath.Base(include.Name) == name {
			return include
		}
	}
	return nil
}

// NamespaceForInclude returns the Namespace for the given inclue name and
// language.
func (f *Frugal) NamespaceForInclude(include, lang string) *Namespace {
	parsed, ok := f.ParsedIncludes[include]
	if !ok {
		return nil
	}
	return parsed.Namespace(lang)
}

// ContainsFrugalDefinitions indicates if the parse tree contains any
// scope or service definitions.
func (f *Frugal) ContainsFrugalDefinitions() bool {
	return len(f.Scopes)+len(f.Services) > 0
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code.
func (f *Frugal) ReferencedIncludes() ([]*Include, error) {
	includes := []*Include{}
	includesSet := make(map[string]*Include)
	for _, serv := range f.Services {
		servIncludes, err := serv.ReferencedIncludes()
		if err != nil {
			return nil, err
		}
		for _, include := range servIncludes {
			if _, ok := includesSet[include.Name]; !ok {
				includesSet[include.Name] = include
				includes = append(includes, include)
			}
		}
	}
	return includes, nil
}

// ReferencedInternals returns a slice containing the referenced internals
// which will need to be imported in generated code.
func (f *Frugal) ReferencedInternals() []string {
	internals := []string{}
	internalsSet := make(map[string]bool)
	for _, serv := range f.Services {
		for _, include := range serv.ReferencedInternals() {
			if _, ok := internalsSet[include]; !ok {
				internalsSet[include] = true
				internals = append(internals, include)
			}
		}
	}
	return internals
}

// ReferencedScopeIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for scopes.
func (f *Frugal) ReferencedScopeIncludes() ([]*Include, error) {
	includeNames := []string{}
	includesSet := make(map[string]*Include)
	for _, scope := range f.Scopes {
		scopeIncludes, err := scope.ReferencedIncludes()
		if err != nil {
			return nil, err
		}
		for _, include := range scopeIncludes {
			if _, ok := includesSet[include.Name]; !ok {
				includesSet[include.Name] = include
				includeNames = append(includeNames, include.Name)
			}
		}
	}
	sort.Strings(includeNames)
	includes := make([]*Include, len(includeNames))
	for i, include := range includeNames {
		includes[i] = includesSet[include]
	}
	return includes, nil
}

// ReferencedServiceIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for services.
func (f *Frugal) ReferencedServiceIncludes() ([]*Include, error) {
	includeNames := []string{}
	includesSet := make(map[string]*Include)
	for _, service := range f.Services {
		servIncludes, err := service.ReferencedIncludes()
		if err != nil {
			return nil, err
		}
		for _, include := range servIncludes {
			if _, ok := includesSet[include.Name]; !ok {
				includesSet[include.Name] = include
				includeNames = append(includeNames, include.Name)
			}
		}
	}
	sort.Strings(includeNames)
	includes := make([]*Include, len(includeNames))
	for i, include := range includeNames {
		includes[i] = includesSet[include]
	}
	return includes, nil
}

// DataStructures returns a slice containing all structs, exceptions, and
// unions.
func (f *Frugal) DataStructures() []*Struct {
	structs := []*Struct{}
	for _, s := range f.Structs {
		structs = append(structs, s)
	}
	for _, s := range f.Exceptions {
		structs = append(structs, s)
	}
	for _, s := range f.Unions {
		structs = append(structs, s)
	}
	return structs
}

// UnderlyingType follows any typedefs to get the base IDL type.
func (f *Frugal) UnderlyingType(t *Type) *Type {
	if t == nil {
		panic("Attempted to get underlying type of nil type")
	}
	typedefIndex := f.typedefIndex
	include := t.IncludeName()
	if include != "" {
		parsed, ok := f.ParsedIncludes[include]
		if !ok {
			return t
		}
		typedefIndex = parsed.typedefIndex
	}
	if typedef, ok := typedefIndex[t.ParamName()]; ok {
		// Recursively call underlying type to handle typedef nesting.
		return f.UnderlyingType(typedef.Type)
	}
	return t
}

// ConstantFromField returns a new Constant from the given Field and value.
func (f *Frugal) ConstantFromField(field *Field, value interface{}) *Constant {
	return &Constant{
		Name:  field.Name,
		Type:  field.Type,
		Value: value,
	}
}

// ContextFromIdentifier returns a IdentifierContext for the given Identifier.
func (f *Frugal) ContextFromIdentifier(identifier Identifier) *IdentifierContext {
	name := string(identifier)

	// Split based on '.', if present, it should be from an include.
	pieces := strings.Split(name, ".")
	if len(pieces) == 1 {
		// From this file.
		for _, constant := range f.Constants {
			if name == constant.Name {
				return &IdentifierContext{
					Type:     LocalConstant,
					Constant: constant,
				}
			}
		}
	} else if len(pieces) == 2 {
		// Either from an include or part of an enum.
		for _, enum := range f.Enums {
			if pieces[0] == enum.Name {
				for _, value := range enum.Values {
					if pieces[1] == value.Name {
						return &IdentifierContext{
							Type:      LocalEnum,
							Enum:      enum,
							EnumValue: value,
						}
					}
				}
			}
		}

		// If not part of an enum, it's from an include.
		include, ok := f.ParsedIncludes[pieces[0]]
		if !ok {
			panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
		}
		for _, constant := range include.Constants {
			if pieces[1] == constant.Name {
				return &IdentifierContext{
					Type:     IncludeConstant,
					Constant: constant,
					Include:  include,
				}
			}
		}
	} else if len(pieces) == 3 {
		// Enum from an include.
		include, ok := f.ParsedIncludes[pieces[0]]
		if !ok {
			panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
		}
		for _, enum := range include.Enums {
			if pieces[1] == enum.Name {
				for _, value := range enum.Values {
					if pieces[2] == value.Name {
						return &IdentifierContext{
							Type:      IncludeEnum,
							Enum:      enum,
							EnumValue: value,
							Include:   include,
						}
					}
				}
				panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
			}
		}
	}

	panic("referenced constant doesn't exist: " + name)
}

// IsStruct indicates if the underlying Type is a struct.
func (f *Frugal) IsStruct(t *Type) bool {
	t = f.UnderlyingType(t)
	if _, ok := frugalBaseTypes[t.Name]; ok {
		return false
	}
	return t.KeyType == nil && t.ValueType == nil && !f.IsEnum(t)
}

// IsUnion indicates if the underlying types is a union.
func (f *Frugal) IsUnion(t *Type) bool {
	t = f.UnderlyingType(t)

	frugal := f
	if t.IncludeName() != "" {
		// The type is from an include
		frugal = f.ParsedIncludes[t.IncludeName()]
	}

	for _, union := range frugal.Unions {
		if t.ParamName() == union.Name {
			return true
		}
	}
	return false
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
	for _, enum := range containingFrugal.Enums {
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

// validate parsed Frugal IDL by ensuring there are no duplicate service/scope
// names and the Frugal IDL is valid.
func (f *Frugal) validate() error {
	// Ensure there are no duplicate names between services and scopes.
	names := make(map[string]string)
	for _, service := range f.Services {
		// Since not every language supports (exported) upper/lowercase
		// class first letters, index by lowercasing the first letter.
		lowercaseService := LowercaseFirstLetter(service.Name)
		if providedService, ok := names[lowercaseService]; ok {
			if service.Name == providedService {
				return fmt.Errorf("Duplicate service name %s", service.Name)
			}
			return getConflictError("Services", service.Name, providedService)
		}
		names[lowercaseService] = service.Name

		methodNames := make(map[string]string)
		for _, method := range service.Methods {
			// Since not every language supports (exported) upper/lowercase
			// method first letters, index by lowercasing the first letter.
			lowercaseMethod := LowercaseFirstLetter(method.Name)
			if providedMethod, ok := methodNames[lowercaseMethod]; ok {
				if method.Name == providedMethod {
					return fmt.Errorf("Duplicate method name %s", method.Name)
				}
				return getConflictError("Methods", method.Name, providedMethod)
			}
			methodNames[lowercaseMethod] = method.Name
		}
	}

	names = make(map[string]string)
	for _, scope := range f.Scopes {
		// Since not every language supports (exported) upper/lowercase
		// class first letters, index by lowercasing the first letter.
		lowercaseScope := LowercaseFirstLetter(scope.Name)
		if providedScope, ok := names[lowercaseScope]; ok {
			if scope.Name == providedScope {
				return fmt.Errorf("Duplicate scope name %s", scope.Name)
			}
			return getConflictError("Scopes", scope.Name, providedScope)
		}
		names[lowercaseScope] = scope.Name

		opNames := make(map[string]string)
		for _, op := range scope.Operations {
			// Since not every language supports (exported) upper/lowercase
			// method first letters, index by lowercasing the first letter.
			lowercaseOp := LowercaseFirstLetter(op.Name)
			if providedOp, ok := opNames[lowercaseOp]; ok {
				if op.Name == providedOp {
					return fmt.Errorf("Duplicate operation name %s", op.Name)
				}
				return getConflictError("Operations", op.Name, providedOp)
			}
			opNames[lowercaseOp] = op.Name
		}
	}

	if err := f.validateNamespaces(); err != nil {
		return err
	}
	if err := f.validateIncludes(); err != nil {
		return err
	}
	if err := f.validateConstants(); err != nil {
		return err
	}
	if err := f.validateTypedefs(); err != nil {
		return err
	}
	if err := f.validateStructs(); err != nil {
		return err
	}
	if err := f.validateUnions(); err != nil {
		return err
	}
	if err := f.validateExceptions(); err != nil {
		return err
	}
	if err := f.validateServices(f.ParsedIncludes); err != nil {
		return err
	}
	return nil
}

func (f *Frugal) validateNamespaces() error {
	for _, namespace := range f.Namespaces {
		_, vendor := namespace.Annotations.Vendor()
		if namespace.Wildcard() && vendor {
			return fmt.Errorf("\"%s\" annotation not compatible with * namespace", VendorAnnotation)
		}
	}
	return nil
}

func (f *Frugal) validateIncludes() error {
	includes := map[string]struct{}{}
	for _, include := range f.Includes {
		if _, ok := includes[include.Name]; ok {
			return fmt.Errorf("Duplicate include: %s", include.Name)
		}
		includes[include.Name] = struct{}{}
	}
	return nil
}

func (f *Frugal) validateConstants() error {
	for _, constant := range f.Constants {
		if err := f.validateConstant(constant); err != nil {
			return err
		}
	}

	return nil
}

func (f *Frugal) validateConstant(constant *Constant) error {
	// validate the type exists
	if ok := f.isValidType(constant.Type); !ok {
		return fmt.Errorf("Invalid type %s", constant.Type.Name)
	}

	identifier, ok := constant.Value.(Identifier)
	if !ok {
		// Just a value, which is fine
		return nil
	}

	// The value of a constant is the name of another constant,
	// make sure it exists
	name := string(identifier)

	// split based on '.', if present, it should be from an include
	pieces := strings.Split(name, ".")
	if len(pieces) == 1 {
		// From this file
		for _, c := range f.Constants {
			if name == c.Name {
				return nil
			}
		}
		return fmt.Errorf("Referenced constant %s not found", name)
	} else if len(pieces) == 2 {
		// From an include
		frugal := f
		includeName := pieces[0]
		paramName := pieces[1]
		if includeName != "" {
			frugalInclude, ok := f.ParsedIncludes[includeName]
			if !ok {
				return fmt.Errorf("Include %s not found", includeName)
			}
			frugal = frugalInclude
		}
		for _, c := range frugal.Constants {
			if paramName == c.Name {
				return nil
			}
		}
		return fmt.Errorf("Referenced constant %s from include %s not found",
			paramName, includeName)
	}

	return fmt.Errorf("Invalid constant name %s", name)
}

func (f *Frugal) validateTypedefs() error {
	for _, typedef := range f.Typedefs {
		if !f.isValidType(typedef.Type) {
			return fmt.Errorf("Invalid alias %s, type %s doesn't exist",
				typedef.Name, typedef.Type.Name)
		}
	}
	return nil
}

func (f *Frugal) validateStructs() error {
	for _, s := range f.Structs {
		if err := f.validateStructLike(s); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frugal) validateUnions() error {
	for _, union := range f.Unions {
		if err := f.validateStructLike(union); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frugal) validateExceptions() error {
	for _, exception := range f.Exceptions {
		if err := f.validateStructLike(exception); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frugal) validateStructLike(s *Struct) error {
	ids := make(map[int]struct{})
	for _, field := range s.Fields {
		if !f.isValidType(field.Type) {
			return fmt.Errorf("Invalid type %s on struct %s", field.Type.String(), s.Name)
		}
		if _, ok := ids[field.ID]; ok {
			return fmt.Errorf("Duplicate field id %d in struct %s", field.ID, s.Name)
		}
		ids[field.ID] = struct{}{}
	}
	return nil
}

func (f *Frugal) isValidType(typ *Type) bool {
	// Check base types
	if typ.IsPrimitive() {
		return true
	} else if typ.IsContainer() {
		switch typ.Name {
		case "list", "set":
			return f.isValidType(typ.ValueType)
		case "map":
			return f.isValidType(typ.KeyType) && f.isValidType(typ.ValueType)
		}
	}

	frugal := f
	includeName := typ.IncludeName()
	paramName := typ.ParamName()
	if includeName != "" {
		frugalInclude, ok := f.ParsedIncludes[includeName]
		if !ok {
			return false
		}
		frugal = frugalInclude
	}

	// Check structs
	for _, s := range frugal.Structs {
		if paramName == s.Name {
			return true
		}
	}

	// Check unions
	for _, union := range frugal.Unions {
		if paramName == union.Name {
			return true
		}
	}

	// Check exceptions
	for _, exception := range frugal.Exceptions {
		if paramName == exception.Name {
			return true
		}
	}

	// Check enums
	for _, enum := range frugal.Enums {
		if paramName == enum.Name {
			return true
		}
	}

	// Check typedefs
	for _, typedef := range frugal.Typedefs {
		if paramName == typedef.Name {
			return true
		}
	}

	return false
}

func (f *Frugal) validateServices(includes map[string]*Frugal) error {
	for _, service := range f.Services {
		if err := f.validateServiceTypes(service, includes); err != nil {
			return err
		}
		if err := service.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frugal) validateServiceTypes(service *Service, includes map[string]*Frugal) error {
	for _, method := range service.Methods {
		if method.ReturnType != nil {
			if !f.isValidType(method.ReturnType) {
				return fmt.Errorf("Invalid return type %s for %s.%s",
					method.ReturnType.Name, service.Name, method.Name)
			}
		}
		for _, field := range method.Arguments {
			if !f.isValidType(field.Type) {
				return fmt.Errorf("Invalid argument type %s for %s.%s",
					field.Type.Name, service.Name, method.Name)
			}
		}
		for _, field := range method.Exceptions {
			if !f.isValidType(field.Type) {
				return fmt.Errorf("Invalid exception type %s for %s.%s",
					field.Type.Name, service.Name, method.Name)
			}
		}
	}
	return nil
}

func getConflictError(type_, name1, name2 string) error {
	return fmt.Errorf("%s %s and %s conflict. Some languages do not support"+
		" exported lowercase classes/methods. Only one of %s or %s may be used.",
		type_, name1, name2, name1, name2)
}

func (f *Frugal) sort() {
	sort.Sort(scopesByName(f.Scopes))
}

type scopesByName []*Scope

func (b scopesByName) Len() int {
	return len(b)
}

func (b scopesByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b scopesByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
