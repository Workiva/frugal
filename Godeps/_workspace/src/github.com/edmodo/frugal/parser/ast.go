// vim: set ts=4 sw=4 tw=99 noet:
//
// Copyright 2014, Edmodo, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this work except in compliance with the License.
// You may obtain a copy of the License in the LICENSE file, or at:
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS"
// BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package parser

import (
	"fmt"
)

// Base interface for all AST nodes.
type Node interface {
	Loc() Location
	NodeType() string
}

// Base interface for all constructs that are parsed as a type expression.
type Type interface {
	Loc() Location
	String() string

	// Reach past all typedefs and return the actual type, and if relevant,
	// the node as well.
	Resolve() (Type, Node)
}

// A builtin type is just a single token (such as i32).
type BuiltinType struct {
	Tok *Token
}

func (this *BuiltinType) Loc() Location {
	return this.Tok.Loc
}

func (this *BuiltinType) String() string {
	return PrettyPrintMap[this.Tok.Kind]
}

func (this *BuiltinType) Resolve() (Type, Node) {
	return this, nil
}

// A list type is list<type>.
type ListType struct {
	Inner Type
}

func (this *ListType) Loc() Location {
	// Not really correct, we should store a Location.
	return this.Inner.Loc()
}

func (this *ListType) String() string {
	return fmt.Sprintf("list<%s>", this.Inner.String())
}

func (this *ListType) Resolve() (Type, Node) {
	return this, nil
}

// A map type is map<key, value>.
type MapType struct {
	Key   Type
	Value Type
}

func (this *MapType) Loc() Location {
	// Not really correct, we should store a Location.
	return this.Key.Loc()
}

func (this *MapType) String() string {
	return fmt.Sprintf("map<%s,%s>", this.Key.String(), this.Value.String())
}

func (this *MapType) Resolve() (Type, Node) {
	return this, nil
}

type EnumEntry struct {
	// Name token (always an identifier).
	Name *Token

	// Initializer (nil, or a TOK_LITERAL_INT).
	Value *Token

	// Constant value, filled in by semantic analysis.
	ConstVal int32
}

// Encapsulates an enum definition.
type EnumNode struct {
	Range   Location
	Name    *Token
	Entries []*EnumEntry

	// Map from name -> Entry. Filled in by semantic analysis.
	Names map[string]*EnumEntry
}

func NewEnumNode(loc Location, name *Token, fields []*EnumEntry) *EnumNode {
	return &EnumNode{
		Range:   loc,
		Name:    name,
		Entries: fields,
		Names:   map[string]*EnumEntry{},
	}
}

func (this *EnumNode) Loc() Location {
	return this.Range
}

func (this *EnumNode) NodeType() string {
	return "enum"
}

type StructField struct {
	// The token which contains the order number, or nil if not present.
	Order *Token

	// A token containing TOK_OPTIONAL or TOK_REQUIRED, or nil (required).
	Spec *Token

	// The type of the field.
	Type Type

	// The name of the field.
	Name *Token

	// The default value, or nil if not present. After semantic analysis, this is
	// converted to a ValueNode.
	Default Node
}

// Encapsulates struct definition.
type StructNode struct {
	Range Location

	// Either TOK_EXCEPTION or TOK_STRUCT.
	Tok *Token

	// Struct/exception name and fields.
	Name   *Token
	Fields []*StructField

	// Map from name -> StructField. Filled in by semantic analysis.
	Names map[string]*StructField
}

func NewStructNode(loc Location, kind *Token, name *Token, fields []*StructField) *StructNode {
	return &StructNode{
		Range:  loc,
		Tok:    kind,
		Name:   name,
		Fields: fields,
		Names:  map[string]*StructField{},
	}
}

func (this *StructNode) NodeType() string {
	return PrettyPrintMap[this.Tok.Kind]
}

func (this *StructNode) Loc() Location {
	return this.Range
}

// Encapsulates a literal.
type LiteralNode struct {
	// The token is either a TOK_LITERAL_INT or TOK_LITERAL_STRING.
	Lit *Token
}

func (this *LiteralNode) Loc() Location {
	return this.Lit.Loc
}

func (this *LiteralNode) NodeType() string {
	return "literal"
}

func (this *LiteralNode) TypeString() string {
	switch this.Lit.Kind {
	case TOK_TRUE, TOK_FALSE:
		return "bool"
	case TOK_LITERAL_STRING:
		return "string"
	case TOK_LITERAL_INT:
		return "integer"
	case TOK_VOID:
		return "void"
	}
	return "<unknown>"
}

// A sequence of expressions.
type ListNode struct {
	Exprs []Node

	// After semantic analysis, this contains the resolved values for each
	// expression.
	Values []*ValueNode
}

func (this *ListNode) Loc() Location {
	return Location{
		Start: this.Exprs[0].Loc().Start,
		End:   this.Exprs[len(this.Exprs)-1].Loc().End,
	}
}

func (this *ListNode) NodeType() string {
	return "list"
}

// A sequence of expression pairs, in a key-value mapping.
type MapNodeEntry struct {
	Key   Node
	Value Node

	// If this is resolved as a map, these are filled in by semantic analysis as
	// part of type resolution. Otherwise - such as for structs - they are left
	// nil.
	KeyVal   *ValueNode
	ValueVal *ValueNode
}

type MapNode struct {
	Range   Location
	Entries []MapNodeEntry
}

func (this *MapNode) Loc() Location {
	return this.Range
}

func (this *MapNode) NodeType() string {
	return "map"
}

// Encapsulates a name or path of names.
type NameProxyNode struct {
	// Path components.
	Path []*Token

	// If not a locally bound name, Import specifies the parse tree it came from.
	// Otherwise it is nil.
	// Set by semantic analysis.
	Import *ParseTree

	// The node this name is bound to. Set by semantic analysis.
	//
	// Note that thrift doesn't allow reaching through imports, i.e. with file x.thrift:
	//   include "y.thrift"
	//
	// Another file could not access "x.y". Therefore, a code generator can always
	// use Import to determine the path to compute.
	Binding Node

	// The remaining components of Path after resolving it to a Node. For example
	// if "x.y.z" resolves to a service "y" in package "x", Import will be "x",
	// "y" will be the *ServiceNode, and Tail will be ["z"].
	Tail []*Token
}

func NewNameProxyNode(path []*Token) *NameProxyNode {
	return &NameProxyNode{
		Path: path,
	}
}

func (this *NameProxyNode) NodeType() string {
	return "name"
}

func (this *NameProxyNode) Loc() Location {
	first := this.Path[0]
	last := this.Path[len(this.Path)-1]
	return Location{first.Loc.Start, last.Loc.End}
}

func (this *NameProxyNode) String() string {
	return JoinIdentifiers(this.Path)
}

func (this *NameProxyNode) Resolve() (Type, Node) {
	// Semantic analysis should have been run, so binding should be valid.
	if _, ok := this.Binding.(*TypedefNode); ok {
		typedef := this.Binding.(*TypedefNode)
		return typedef.Type.Resolve()
	}
	return this, this.Binding
}

type ServiceMethodArg struct {
	// The order of the argument, if present, as a TOK_LITERAL_INT
	Order *Token

	// The type expression of the argument.
	Type Type

	// The token containing the argument name.
	Name *Token
}

type ServiceMethod struct {
	// If non-nil, specifies that the method is one-way.
	OneWay *Token

	// The return type expression of the method.
	ReturnType Type

	// The name of the method.
	Name *Token

	// The argument list of the method.
	Args []*ServiceMethodArg

	// The list of throwable errors of the method.
	Throws []*ServiceMethodArg
}

// Returns whether or not a method has no return value. Should only be called
// after semantic analysis.
func (this *ServiceMethod) ReturnsVoid() bool {
	// Peek past typedefs.
	ttype, _ := this.ReturnType.Resolve()
	builtin, ok := ttype.(*BuiltinType)
	if !ok {
		return false
	}
	return builtin.Tok.Kind == TOK_VOID
}

// Encapsulates a service definition.
type ServiceNode struct {
	Range   Location
	Name    *Token
	Extends *NameProxyNode
	Methods []*ServiceMethod
}

func (this *ServiceNode) Loc() Location {
	return this.Range
}

func (this *ServiceNode) NodeType() string {
	return "service"
}

// Return the base service, if any. Only valid after semantic analysis.
func (this *ServiceNode) BaseService() *ServiceNode {
	if this.Extends == nil {
		return nil
	}

	return this.Extends.Binding.(*ServiceNode)
}

// Returns the entire inheritance chain. Only valid after semantic analysis.
// The derived-most service is first in the list.
func (this *ServiceNode) InheritanceChain() []*ServiceNode {
	chain := []*ServiceNode{}
	for current := this; current != nil; current = current.BaseService() {
		chain = append(chain, current)
	}
	return chain
}

// Encapsulates a constant variable definition.
type ConstNode struct {
	Range Location

	// The type of the constant variable.
	Type Type

	// The name of the constant variable.
	Name *Token

	// The initialization value of the constant variable.
	// This is always one of:
	//   LiteralNode
	//   ListNode
	//   MapNode
	//
	// After semantic analysis, this is converted to a ValueNode.
	Init Node
}

func (this *ConstNode) Loc() Location {
	return this.Range
}

func (this *ConstNode) NodeType() string {
	return "constant"
}

// Encapsulates a typedef definition.
type TypedefNode struct {
	Range Location
	Type  Type
	Name  *Token
}

func (this *TypedefNode) Loc() Location {
	return this.Range
}

func (this *TypedefNode) NodeType() string {
	return "typedef"
}

type StructInitializer map[*StructField]*ValueNode

// A Value node is not constructed by the parser. It is produced by Semantic
// Analysis during type checking, to replace NameProxy nodes that resolve to
// enum fields.
type ValueNode struct {
	Original Node

	// Describes the type below.
	Type TokenKind

	// One of:
	//   A bool true/false. (Type = BOOL)
	//   An i16. (Type = I16)
	//   An i32. (Type = I32)
	//   An i64. (Type = I64)
	//   A string. (Type = STRING)
	//   A *ListNode. (Type = LIST)
	//   A *MapNode. (Type = MAP)
	//   An *EnumEntry. (Type = ENUM)
	//   A StructInitializer. (Type = STRUCT)
	Result interface{}
}

func (this *ValueNode) Loc() Location {
	return this.Original.Loc()
}

func (this *ValueNode) NodeType() string {
	return "value"
}

// Include directive information.
type Include struct {
	// The token containing the include string.
	Tok *Token

	// The package derived from the include string.
	Package string

	// The parse tree, filled in by ParseRecursive().
	Tree *ParseTree
}

type ParseTree struct {
	// Mapping of language -> namespace.
	Namespaces map[string]string

	// Map of package names to includes.
	Includes map[string]*Include

	// Root nodes in the syntax tree.
	Nodes []Node

	// The original file path.
	Path string

	// The package name this file would be imported as, in thrift. For example,
	// "egg.thrift" will become package "egg".
	Package string

	// Name to node mapping, filled in by semantic analysis.
	Names map[string]Node

	// Set of which includes are used. Filled in by semantic analysis.
	UsedIncludes map[string]*ParseTree
}

func NewParseTree(file string) *ParseTree {
	return &ParseTree{
		Namespaces:   map[string]string{},
		Includes:     map[string]*Include{},
		Path:         file,
		Names:        map[string]Node{},
		UsedIncludes: map[string]*ParseTree{},
	}
}
