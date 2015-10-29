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
	"io"
)

type AstPrinter struct {
	fp     io.Writer
	tree   *ParseTree
	prefix string
}

func (this *AstPrinter) fprintf(msg string, args ...interface{}) {
	text := this.prefix + fmt.Sprintf(msg, args...)
	_, err := this.fp.Write([]byte(text))
	if err != nil {
		panic(err)
	}
}

func (this *AstPrinter) indent() {
	this.prefix += "  "
}

func (this *AstPrinter) dedent() {
	this.prefix = this.prefix[:len(this.prefix)-2]
}

func (this *AstPrinter) printArg(arg *ServiceMethodArg) {
	this.indent()
	msg := ""
	if arg.Order != nil {
		msg += fmt.Sprintf("%d: ", arg.Order.IntLiteral())
	}
	msg += fmt.Sprintf("%s ", arg.Type.String())
	msg += fmt.Sprintf("%s", arg.Name.Identifier())
	this.fprintf("%s\n", msg)
	this.dedent()
}

func (this *AstPrinter) dumpLiteral(node Node) {
	switch node.(type) {
	case *LiteralNode:
		node := node.(*LiteralNode)
		switch node.Lit.Kind {
		case TOK_LITERAL_INT:
			this.fprintf("%d\n", node.Lit.IntLiteral())
		case TOK_LITERAL_STRING:
			this.fprintf("\"%s\"\n", node.Lit.StringLiteral())
		}

	case *MapNode:
		node := node.(*MapNode)
		this.fprintf("{\n")
		this.indent()
		for _, entry := range node.Entries {
			this.fprintf("key = ")
			this.dumpLiteral(entry.Key)
			this.fprintf("value = ")
			this.dumpLiteral(entry.Value)
		}
		this.dedent()
		this.fprintf("}\n")

	case *NameProxyNode:
		node := node.(*NameProxyNode)
		this.fprintf("%s\n", node.String())

	default:
		this.fprintf("unrecognized node: %T %v\n", node, node)
	}
}

func (this *AstPrinter) printMethod(method *ServiceMethod) {
	extra := ""
	if method.OneWay != nil {
		extra = "oneway"
	}

	this.fprintf("[ method %s %s\n", method.Name.Identifier(), extra)
	this.indent()

	this.fprintf("args = \n")
	for _, arg := range method.Args {
		this.printArg(arg)
	}

	this.fprintf("throws = \n")
	for _, arg := range method.Throws {
		this.printArg(arg)
	}

	this.dedent()
}

func (this *AstPrinter) print() {
	for key, value := range this.tree.Namespaces {
		this.fprintf("namespace %s %s\n", key, value)
	}
	if len(this.tree.Namespaces) > 0 {
		this.fprintf("\n")
	}

	for include, _ := range this.tree.Includes {
		this.fprintf("include \"%s.thrift\"\n", include)
	}
	if len(this.tree.Includes) > 0 {
		this.fprintf("\n")
	}

	for _, node := range this.tree.Nodes {
		this.printNode(node)
	}
}

func (this *AstPrinter) printNode(node Node) {
	switch node.(type) {
	case *EnumNode:
		node := node.(*EnumNode)
		this.fprintf("[ enum %s\n", node.Name.Identifier())
		this.indent()
		for _, entry := range node.Entries {
			this.fprintf("%s\n", entry.Name.Identifier())
		}
		this.dedent()

	case *StructNode:
		node := node.(*StructNode)
		this.fprintf("[ %s %s\n", PrettyPrintMap[node.Tok.Kind], node.Name.Identifier())
		this.indent()
		for _, field := range node.Fields {
			msg := ""
			if field.Order != nil {
				msg += fmt.Sprintf("%d: ", field.Order.IntLiteral())
			}
			msg += fmt.Sprintf("%s ", PrettyPrintMap[field.Spec.Kind])
			msg += fmt.Sprintf("%s", field.Name.Identifier())
			this.fprintf("%s\n", msg)
		}
		this.dedent()

	case *ServiceNode:
		node := node.(*ServiceNode)
		header := fmt.Sprintf("[ service %s", node.Name.Identifier())
		if node.Extends != nil {
			header += fmt.Sprintf(" extends %s", node.Extends.String())
		}
		this.fprintf("%s\n", header)
		this.indent()
		for _, method := range node.Methods {
			this.printMethod(method)
		}
		this.dedent()

	case *ConstNode:
		node := node.(*ConstNode)
		this.fprintf("[ const %s %s = \n", node.Type.String(), node.Name.Identifier())
		this.indent()
		this.dumpLiteral(node.Init)
		this.dedent()

	case *TypedefNode:
		node := node.(*TypedefNode)
		this.fprintf("[ typedef %s as %s\n", node.Type.String(), node.Name.Identifier())

	default:
		this.fprintf("Unrecognized node! %T %v\n", node, node)
	}
}

func (this *ParseTree) Print(fp io.Writer) {
	printer := AstPrinter{
		fp:   fp,
		tree: this,
	}
	printer.print()
}
