package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

var thriftBaseTypes = map[string]bool{
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

var thriftContainerTypes = map[string]bool{
	"list": true,
	"set":  true,
	"map":  true,
}

// FieldModifier represents a Thrift IDL field modifier (required, optional
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

// IsPrimitive indicates if the type is a Thrift primitive type.
func (t *Type) IsPrimitive() bool {
	_, ok := thriftBaseTypes[t.Name]
	return ok
}

// IsCustom indicates if the type is not a container or primitive type.
func (t *Type) IsCustom() bool {
	return !t.IsPrimitive() && !t.IsContainer()
}

// IsContainer indicates if the type is a Thrift container type (list, set, or
// map).
func (t *Type) IsContainer() bool {
	_, ok := thriftContainerTypes[t.Name]
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
	Thrift      *Thrift // Pointer back to containing Thrift
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
		include := s.Thrift.Include(includeName)
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
			includesSet, includes, err = addInclude(includesSet, includes, arg.Type, s.Thrift)
			if err != nil {
				return nil, err
			}
		}
		// Check return type.
		if method.ReturnType != nil {
			includesSet, includes, err = addInclude(includesSet, includes, method.ReturnType, s.Thrift)
		}
		if err != nil {
			return nil, err
		}
		// Check exceptions.
		for _, exception := range method.Exceptions {
			includesSet, includes, err = addInclude(includesSet, includes, exception.Type, s.Thrift)
		}
		if err != nil {
			return nil, err
		}
	}

	return includes, nil
}

// addInclude checks the given Type and adds any includes for it to the given
// map and slice, returning the new map and slice.
func addInclude(includesSet map[string]*Include, includes []*Include, t *Type, thrift *Thrift) (map[string]*Include, []*Include, error) {
	var err error
	if strings.Contains(t.Name, ".") {
		includeName := t.Name[0:strings.Index(t.Name, ".")]
		include := thrift.Include(includeName)
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
		includesSet, includes, err = addInclude(includesSet, includes, t.KeyType, thrift)
	}
	if err != nil {
		return nil, nil, err
	}
	if t.ValueType != nil {
		includesSet, includes, err = addInclude(includesSet, includes, t.ValueType, thrift)
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

// Thrift contains the Thrift-specific IDL parse tree.
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

// Namespace returns namespace value for the given scope.
func (t *Thrift) Namespace(scope string) *Namespace {
	namespace := t.namespaceIndex[scope]
	if namespace != nil {
		return namespace
	}
	return t.namespaceIndex["*"]
}

// Include returns the Include with the given name.
func (t *Thrift) Include(name string) *Include {
	name = filepath.Base(name)
	for _, include := range t.Includes {
		if filepath.Base(include.Name) == name {
			return include
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

// Vendor returns true if the "vendor" annotation is present and its associated
// value, if any.
func (a Annotations) Vendor() (string, bool) {
	for _, annotation := range a {
		if annotation.Name == VendorAnnotation {
			return annotation.Value, true
		}
	}
	return "", false
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code.
func (t *Thrift) ReferencedIncludes() ([]*Include, error) {
	includes := []*Include{}
	includesSet := make(map[string]*Include)
	for _, serv := range t.Services {
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

// DataStructures returns a slice containing all structs, exceptions, and
// unions.
func (t *Thrift) DataStructures() []*Struct {
	structs := []*Struct{}
	for _, s := range t.Structs {
		structs = append(structs, s)
	}
	for _, s := range t.Exceptions {
		structs = append(structs, s)
	}
	for _, s := range t.Unions {
		structs = append(structs, s)
	}
	return structs
}

// validate parsed Thrift IDL, e.g. no duplicate identifiers, valid types,
// no wildcard namespace with a "vendor" annotation, etc.
func (t *Thrift) validate(includes map[string]*Frugal) error {
	if err := t.validateNamespaces(); err != nil {
		return err
	}
	if err := t.validateIncludes(); err != nil {
		return err
	}
	if err := t.validateConstants(includes); err != nil {
		return err
	}
	if err := t.validateTypedefs(includes); err != nil {
		return err
	}
	if err := t.validateStructs(includes); err != nil {
		return err
	}
	if err := t.validateUnions(includes); err != nil {
		return err
	}
	if err := t.validateExceptions(includes); err != nil {
		return err
	}
	if err := t.validateServices(includes); err != nil {
		return err
	}
	return nil
}

func (t *Thrift) validateNamespaces() error {
	for _, namespace := range t.Namespaces {
		_, vendor := namespace.Annotations.Vendor()
		if namespace.Wildcard() && vendor {
			return fmt.Errorf("\"%s\" annotation not compatible with * namespace", VendorAnnotation)
		}
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

func (t *Thrift) validateConstants(includes map[string]*Frugal) error {
	for _, constant := range t.Constants {
		if err := t.validateConstant(constant, includes); err != nil {
			return err
		}
	}

	return nil
}

func (t *Thrift) validateConstant(constant *Constant, includes map[string]*Frugal) error {
	// validate the type exists
	if ok := t.isValidType(constant.Type, includes); !ok {
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
		for _, c := range t.Constants {
			if name == c.Name {
				return nil
			}
		}
		return fmt.Errorf("Referenced constant %s not found", name)
	} else if len(pieces) == 2 {
		// From an include
		thrift := t
		includeName := pieces[0]
		paramName := pieces[1]
		if includeName != "" {
			frugalInclude, ok := includes[includeName]
			if !ok {
				return fmt.Errorf("Include %s not found", includeName)
			}
			thrift = frugalInclude.Thrift
		}
		for _, c := range thrift.Constants {
			if paramName == c.Name {
				return nil
			}
		}
		return fmt.Errorf("Referenced constant %s from include %s not found",
			paramName, includeName)
	}

	return fmt.Errorf("Invalid constant name %s", name)
}

func (t *Thrift) validateTypedefs(includes map[string]*Frugal) error {
	for _, typedef := range t.Typedefs {
		if !t.isValidType(typedef.Type, includes) {
			return fmt.Errorf("Invalid alias %s, type %s doesn't exist",
				typedef.Name, typedef.Type.Name)
		}
	}
	return nil
}

func (t *Thrift) validateStructs(includes map[string]*Frugal) error {
	for _, s := range t.Structs {
		if err := t.validateStructLike(s, includes); err != nil {
			return err
		}
	}
	return nil
}

func (t *Thrift) validateUnions(includes map[string]*Frugal) error {
	for _, union := range t.Unions {
		if err := t.validateStructLike(union, includes); err != nil {
			return err
		}
	}
	return nil
}

func (t *Thrift) validateExceptions(includes map[string]*Frugal) error {
	for _, exception := range t.Exceptions {
		if err := t.validateStructLike(exception, includes); err != nil {
			return err
		}
	}
	return nil
}

func (t *Thrift) validateStructLike(s *Struct, includes map[string]*Frugal) error {
	ids := make(map[int]struct{})
	for _, field := range s.Fields {
		if !t.isValidType(field.Type, includes) {
			return fmt.Errorf("Invalid type %s on struct %s", field.Type.String(), s.Name)
		}
		if _, ok := ids[field.ID]; ok {
			return fmt.Errorf("Duplicate field id %d in struct %s", field.ID, s.Name)
		}
		ids[field.ID] = struct{}{}
	}
	return nil
}

func (t *Thrift) isValidType(typ *Type, includes map[string]*Frugal) bool {
	// Check base types
	if typ.IsPrimitive() {
		return true
	} else if typ.IsContainer() {
		switch typ.Name {
		case "list", "set":
			return t.isValidType(typ.ValueType, includes)
		case "map":
			return t.isValidType(typ.KeyType, includes) && t.isValidType(typ.ValueType, includes)
		}
	}

	thrift := t
	includeName := typ.IncludeName()
	paramName := typ.ParamName()
	if includeName != "" {
		frugalInclude, ok := includes[includeName]
		if !ok {
			return false
		}
		thrift = frugalInclude.Thrift
	}

	// Check structs
	for _, s := range thrift.Structs {
		if paramName == s.Name {
			return true
		}
	}

	// Check unions
	for _, union := range thrift.Unions {
		if paramName == union.Name {
			return true
		}
	}

	// Check exceptions
	for _, exception := range thrift.Exceptions {
		if paramName == exception.Name {
			return true
		}
	}

	// Check enums
	for _, enum := range thrift.Enums {
		if paramName == enum.Name {
			return true
		}
	}

	// Check typedefs
	for _, typedef := range thrift.Typedefs {
		if paramName == typedef.Name {
			return true
		}
	}

	return false
}

func (t *Thrift) validateServices(includes map[string]*Frugal) error {
	for _, service := range t.Services {
		if err := t.validateServiceTypes(service, includes); err != nil {
			return err
		}
		if err := service.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Thrift) validateServiceTypes(service *Service, includes map[string]*Frugal) error {
	for _, method := range service.Methods {
		if method.ReturnType != nil {
			if !t.isValidType(method.ReturnType, includes) {
				return fmt.Errorf("Invalid return type %s for %s.%s",
					method.ReturnType.Name, service.Name, method.Name)
			}
		}
		for _, field := range method.Arguments {
			if !t.isValidType(field.Type, includes) {
				return fmt.Errorf("Invalid argument type %s for %s.%s",
					field.Type.Name, service.Name, method.Name)
			}
		}
		for _, field := range method.Exceptions {
			if !t.isValidType(field.Type, includes) {
				return fmt.Errorf("Invalid exception type %s for %s.%s",
					field.Type.Name, service.Name, method.Name)
			}
		}
	}
	return nil
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
