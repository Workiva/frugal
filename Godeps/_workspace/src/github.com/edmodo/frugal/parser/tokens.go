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
	"strings"
)

type TokenKind int

type Position struct {
	Line int
	Col  int
}

func (this Position) String() string {
	return fmt.Sprintf("line %d, col %d", this.Line, this.Col)
}

type Location struct {
	Start Position
	End   Position
}

type Token struct {
	Kind TokenKind
	Data interface{}
	Loc  Location
}

const (
	// Compound sets.
	TOK_ERROR TokenKind = iota
	TOK_EOF
	TOK_IDENTIFIER     // [_A-Za-z][_A-Za-z0-9]*
	TOK_LITERAL_INT    // [0-9]*
	TOK_LITERAL_STRING // "[^"]*"

	// Keywords.
	TOK_BOOL
	TOK_CONST
	TOK_DOUBLE
	TOK_ENUM
	TOK_EXCEPTION
	TOK_EXTENDS
	TOK_FALSE
	TOK_I16
	TOK_I32
	TOK_I64
	TOK_INCLUDE
	TOK_LIST
	TOK_MAP
	TOK_NAMESPACE
	TOK_ONEWAY
	TOK_OPTIONAL
	TOK_REQUIRED
	TOK_SERVICE
	TOK_STRING
	TOK_STRUCT
	TOK_THROWS
	TOK_TRUE
	TOK_TYPEDEF
	TOK_VOID

	// Chars.
	TOK_LBRACE
	TOK_RBRACE
	TOK_LBRACKET
	TOK_RBRACKET
	TOK_LPAREN
	TOK_RPAREN
	TOK_LT
	TOK_GT
	TOK_ASSIGN
	TOK_COLON
	TOK_DOT
	TOK_COMMA
	TOK_SEMICOLON
)

var KeywordMap = map[string]TokenKind{
	"bool":      TOK_BOOL,
	"const":     TOK_CONST,
	"double":    TOK_DOUBLE,
	"enum":      TOK_ENUM,
	"exception": TOK_EXCEPTION,
	"extends":   TOK_EXTENDS,
	"false":     TOK_FALSE,
	"i16":       TOK_I16,
	"i32":       TOK_I32,
	"i64":       TOK_I64,
	"include":   TOK_INCLUDE,
	"list":      TOK_LIST,
	"map":       TOK_MAP,
	"namespace": TOK_NAMESPACE,
	"oneway":    TOK_ONEWAY,
	"optional":  TOK_OPTIONAL,
	"required":  TOK_REQUIRED,
	"service":   TOK_SERVICE,
	"string":    TOK_STRING,
	"struct":    TOK_STRUCT,
	"throws":    TOK_THROWS,
	"true":      TOK_TRUE,
	"typedef":   TOK_TYPEDEF,
	"void":      TOK_VOID,
}

var PrettyPrintMap = map[TokenKind]string{
	TOK_IDENTIFIER:     "<identifier>",
	TOK_LITERAL_INT:    "<integer>",
	TOK_LITERAL_STRING: "<string>",
	TOK_BOOL:           "bool",
	TOK_CONST:          "const",
	TOK_DOUBLE:         "double",
	TOK_ENUM:           "enum",
	TOK_EXCEPTION:      "exception",
	TOK_EXTENDS:        "extends",
	TOK_FALSE:          "false",
	TOK_I16:            "i16",
	TOK_I32:            "i32",
	TOK_I64:            "i64",
	TOK_INCLUDE:        "include",
	TOK_LIST:           "list",
	TOK_MAP:            "map",
	TOK_NAMESPACE:      "namespace",
	TOK_ONEWAY:         "oneway",
	TOK_OPTIONAL:       "optional",
	TOK_REQUIRED:       "required",
	TOK_SERVICE:        "service",
	TOK_STRING:         "string",
	TOK_STRUCT:         "struct",
	TOK_THROWS:         "throws",
	TOK_TRUE:           "true",
	TOK_TYPEDEF:        "typedef",
	TOK_VOID:           "void",
	TOK_LBRACE:         "{",
	TOK_RBRACE:         "}",
	TOK_LBRACKET:       "[",
	TOK_RBRACKET:       "[",
	TOK_LPAREN:         "(",
	TOK_RPAREN:         ")",
	TOK_LT:             "<",
	TOK_GT:             ">",
	TOK_ASSIGN:         "=",
	TOK_COLON:          ":",
	TOK_DOT:            ".",
	TOK_COMMA:          ",",
	TOK_SEMICOLON:      ";",
}

// Pretty-prints the token to a string.
func (this *Token) String() string {
	if this.Data == nil {
		return this.Name()
	}
	return fmt.Sprintf("%s %v", this.Name(), this.Data)
}

func (this *Token) Name() string {
	if name, ok := PrettyPrintMap[this.Kind]; ok {
		return name
	}
	return "<unknown>"
}

func (this *Token) Identifier() string {
	if this.Kind != TOK_IDENTIFIER {
		panic(fmt.Errorf("only valid for identifier tokens (got %s)", this))
	}
	return this.Data.(string)
}

func (this *Token) StringLiteral() string {
	if this.Kind != TOK_LITERAL_STRING {
		panic("only valid for string tokens")
	}
	return this.Data.(string)
}

func (this *Token) IntLiteral() int64 {
	if this.Kind != TOK_LITERAL_INT {
		panic("only valid for integer tokens")
	}
	return this.Data.(int64)
}

func JoinIdentifiers(tokens []*Token) string {
	strs := []string{}
	for _, tok := range tokens {
		strs = append(strs, tok.Identifier())
	}
	return strings.Join(strs, ".")
}
