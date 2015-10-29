package parser

import (
	"path/filepath"
	"strings"
)

// A simple recursive-descent parser for Thrift IDL.
type Parser struct {
	Context *CompileContext
	scanner *Scanner
	tree    *ParseTree
}

func NewParser(context *CompileContext) (*Parser, error) {
	scanner, err := NewScanner(context)
	if err != nil {
		return nil, err
	}

	return &Parser{
		Context: context,
		scanner: scanner,
		tree:    NewParseTree(context.CurFile),
	}, nil
}

// If the next token matches |kind|, return the token. Otherwise, return nil.
func (this *Parser) match(kind TokenKind) *Token {
	tok := this.scanner.next()
	if tok.Kind != kind {
		this.scanner.undo()
		return nil
	}
	return tok
}

// If the next token matches |kind|, return the token. Otherwise, report an
// error and return nil.
func (this *Parser) need(kind TokenKind) *Token {
	tok := this.scanner.next()
	if kind != tok.Kind {
		name := PrettyPrintMap[kind]
		this.Context.ReportError(tok.Loc.Start, "expected %s, but got %s", name, tok.String())
		return nil
	}
	return tok
}

func (this *Parser) requireTerminator() {
	// Currently, thrift has no concept of terminators. It allows, optionally,
	// ',' or ';'. We should consider deviating from the official grammar and
	// requiring at the very least, a newline.
	if this.match(TOK_COMMA) != nil {
		return
	}
	this.match(TOK_SEMICOLON)
}

// Parse the following:
//   name-path ::= identifier ("." identifier)*
//   namespace ::= "namespace" name-path
//
// The first identifier is a language extension, and the rest of the production
// is the namespace for that language. These do not appear in the formal AST,
// rather, they are collected into a map during parsing.
func (this *Parser) parseNamespace() bool {
	tok := this.need(TOK_IDENTIFIER)
	if tok == nil {
		return false
	}

	language := tok.Identifier()

	// At least one part of a namespace is required.
	if tok = this.need(TOK_IDENTIFIER); tok == nil {
		return false
	}
	parts := []string{tok.Identifier()}

	// Read any additional components.
	for this.match(TOK_DOT) != nil {
		if tok = this.need(TOK_IDENTIFIER); tok == nil {
			return false
		}
		parts = append(parts, tok.Identifier())
	}

	namespace := strings.Join(parts, ".")
	this.tree.Namespaces[language] = namespace

	return true
}

// Parse the following:
//   include ::= "include" literal-string
//
// Like namespaces, these are not in the AST proper.
func (this *Parser) parseInclude() bool {
	tok := this.need(TOK_LITERAL_STRING)
	if tok == nil {
		return false
	}

	folder, name := filepath.Split(tok.StringLiteral())
	if folder != "" {
		this.Context.ReportError(tok.Loc.Start, "External paths are not yet supported")
		return false
	}

	// Take ".thrift" out of the path.
	index := strings.Index(name, ".thrift")
	if index != -1 {
		name = name[:index]
	}

	this.tree.Includes[name] = &Include{
		Tok:     tok,
		Package: name,
		Tree:    nil,
	}
	return true
}

// Parse:
//   expr ::= string-literal
//          | integer-literal
//          | name-path
//          | "[" (expr ","?)* "]"
//          | "{" (expr ":" expr ","?)* "}"
func (this *Parser) parseExpr() Node {
	tok := this.scanner.next()
	switch tok.Kind {
	case TOK_LITERAL_STRING, TOK_LITERAL_INT, TOK_TRUE, TOK_FALSE:
		return &LiteralNode{tok}

	case TOK_IDENTIFIER:
		path := this.parseNames(tok)
		if path == nil {
			return nil
		}
		return NewNameProxyNode(path)

	// Parse a list of expressions.
	case TOK_LBRACKET:
		exprs := []Node{}
		for this.match(TOK_RBRACKET) == nil {
			expr := this.parseExpr()
			if expr == nil {
				return nil
			}
			exprs = append(exprs, expr)

			this.requireTerminator()
		}
		return &ListNode{exprs, nil}

	// Parse a list of key-value pairs.
	case TOK_LBRACE:
		start := tok.Loc.Start
		entries := []MapNodeEntry{}
		for this.match(TOK_RBRACE) == nil {
			left := this.parseExpr()
			if left == nil {
				return nil
			}

			if this.need(TOK_COLON) == nil {
				return nil
			}

			right := this.parseExpr()
			if right == nil {
				return nil
			}

			entries = append(entries, MapNodeEntry{
				Key:   left,
				Value: right,
			})

			this.requireTerminator()
		}
		return &MapNode{
			Location{
				Start: start,
				End:   this.scanner.Position(),
			},
			entries,
		}
	}

	this.Context.ReportError(tok.Loc.Start, "expected a constant expression, got %s", tok.String())
	return nil
}

// Parse the following:
//   enum ::= "enum" identifier "{" (identifier ","?)* "}"
func (this *Parser) parseEnum(start *Token) *EnumNode {
	name := this.need(TOK_IDENTIFIER)
	if name == nil {
		return nil
	}
	if this.need(TOK_LBRACE) == nil {
		return nil
	}

	entries := []*EnumEntry{}
	for this.match(TOK_RBRACE) == nil {
		name := this.need(TOK_IDENTIFIER)
		if name == nil {
			return nil
		}

		var value *Token
		if this.match(TOK_ASSIGN) != nil {
			if value = this.need(TOK_LITERAL_INT); value == nil {
				return nil
			}
		}

		entries = append(entries, &EnumEntry{
			Name:  name,
			Value: value,
		})

		this.requireTerminator()
	}

	return NewEnumNode(
		Location{
			Start: start.Loc.Start,
			End:   this.scanner.Position(),
		},
		name,
		entries,
	)
}

// Parse the rest of a fully-qualified name.
func (this *Parser) parseNames(first *Token) []*Token {
	path := []*Token{first}
	for this.match(TOK_DOT) != nil {
		ident := this.need(TOK_IDENTIFIER)
		if ident == nil {
			return nil
		}
		path = append(path, ident)
	}
	return path
}

// Parse a fully-qualified name.
func (this *Parser) parseFullName() *NameProxyNode {
	tok := this.need(TOK_IDENTIFIER)
	if tok == nil {
		return nil
	}
	path := this.parseNames(tok)
	if path == nil {
		return nil
	}
	return NewNameProxyNode(path)
}

// Parse the following:
//   type ::= i32
//          | i64
//          | string
//          | double
//          | name-path
//          | "list" "<" type ">"
//          | "map" "<" type "," type ">"
func (this *Parser) parseType() Type {
	tok := this.scanner.next()
	switch tok.Kind {
	case TOK_I16,
		TOK_I32,
		TOK_I64,
		TOK_BOOL,
		TOK_VOID,
		TOK_STRING,
		TOK_DOUBLE:
		return &BuiltinType{tok}

	case TOK_IDENTIFIER:
		path := this.parseNames(tok)
		if path == nil {
			return nil
		}
		return NewNameProxyNode(path)

	// list<type>
	case TOK_LIST:
		if this.need(TOK_LT) == nil {
			return nil
		}
		ttype := this.parseType()
		if ttype == nil {
			return nil
		}
		if this.need(TOK_GT) == nil {
			return nil
		}
		return &ListType{ttype}

	// map<type, type>
	case TOK_MAP:
		if this.need(TOK_LT) == nil {
			return nil
		}
		left := this.parseType()
		if left == nil || this.need(TOK_COMMA) == nil {
			return nil
		}
		right := this.parseType()
		if right == nil {
			return nil
		}
		if this.need(TOK_GT) == nil {
			return nil
		}
		return &MapType{left, right}
	}

	this.Context.ReportError(tok.Loc.Start, "expected type name, got: %s", tok.String())
	return nil
}

// Parse the following:
//   struct              ::= "struct" identifier "{" struct-body "}"
//   struct-body         ::= (struct-member ","?)*
//   struct-member       ::= struct-member-order? struct-member-spec? type identifier ("=" expression)?
//   struct-member-order ::= integer-literal ":"
//   struct-member-spec  ::= "required" | "optional"
//
func (this *Parser) parseStruct(start *Token) *StructNode {
	name := this.need(TOK_IDENTIFIER)
	if name == nil {
		return nil
	}
	if this.need(TOK_LBRACE) == nil {
		return nil
	}

	fields := []*StructField{}
	for this.match(TOK_RBRACE) == nil {
		order := this.match(TOK_LITERAL_INT)
		if order != nil {
			if this.need(TOK_COLON) == nil {
				return nil
			}
		}

		spec := this.match(TOK_REQUIRED)
		if spec == nil {
			spec = this.match(TOK_OPTIONAL)
		}

		ttype := this.parseType()
		if ttype == nil {
			return nil
		}

		name := this.need(TOK_IDENTIFIER)
		if name == nil {
			return nil
		}

		var expr Node
		if this.match(TOK_ASSIGN) != nil {
			if expr = this.parseExpr(); expr == nil {
				return nil
			}
		}

		fields = append(fields, &StructField{
			Order:   order,
			Spec:    spec,
			Type:    ttype,
			Name:    name,
			Default: expr,
		})

		this.requireTerminator()
	}

	return NewStructNode(
		Location{
			Start: start.Loc.Start,
			End:   this.scanner.Position(),
		},
		start,
		name,
		fields,
	)
}

// Parse:
//   service-method-arg ::= (integer-literal ":")? type identifier ","?
func (this *Parser) parseArgs() []*ServiceMethodArg {
	if this.need(TOK_LPAREN) == nil {
		return nil
	}

	args := []*ServiceMethodArg{}
	for this.match(TOK_RPAREN) == nil {
		order := this.match(TOK_LITERAL_INT)
		if order != nil {
			if this.need(TOK_COLON) == nil {
				return nil
			}
		}

		ttype := this.parseType()
		if ttype == nil {
			return nil
		}

		name := this.need(TOK_IDENTIFIER)
		if name == nil {
			return nil
		}

		args = append(args, &ServiceMethodArg{
			Order: order,
			Type:  ttype,
			Name:  name,
		})

		this.requireTerminator()
	}

	return args
}

// Parse:
//   service ::= "service" identifier ("extends" name-path) "{" service-body "}"
//   service-body ::= service-method*
//   service-method ::= type identifier "(" service-method-arg* ")" service-method-throws?
//   service-method-throws ::= "throws" "(" service-method-arg* ")"
func (this *Parser) parseService(start *Token) *ServiceNode {
	name := this.need(TOK_IDENTIFIER)
	if name == nil {
		return nil
	}

	var extends *NameProxyNode
	if this.match(TOK_EXTENDS) != nil {
		if extends = this.parseFullName(); extends == nil {
			return nil
		}
	}

	if this.need(TOK_LBRACE) == nil {
		return nil
	}

	methods := []*ServiceMethod{}
	for this.match(TOK_RBRACE) == nil {
		oneway := this.match(TOK_ONEWAY)

		ttype := this.parseType()
		if ttype == nil {
			return nil
		}

		name := this.need(TOK_IDENTIFIER)
		if name == nil {
			return nil
		}

		args := this.parseArgs()
		if args == nil {
			return nil
		}

		var throws []*ServiceMethodArg
		if this.match(TOK_THROWS) != nil {
			if throws = this.parseArgs(); throws == nil {
				return nil
			}
		}

		method := &ServiceMethod{
			OneWay:     oneway,
			ReturnType: ttype,
			Name:       name,
			Args:       args,
			Throws:     throws,
		}
		methods = append(methods, method)
	}

	return &ServiceNode{
		Range: Location{
			Start: start.Loc.Start,
			End:   this.scanner.Position(),
		},
		Name:    name,
		Extends: extends,
		Methods: methods,
	}
}

func (this *Parser) parseTypedef(start *Token) *TypedefNode {
	ttype := this.parseType()
	if ttype == nil {
		return nil
	}

	name := this.need(TOK_IDENTIFIER)
	if name == nil {
		return nil
	}

	return &TypedefNode{
		Range: Location{
			Start: start.Loc.Start,
			End:   this.scanner.Position(),
		},
		Type: ttype,
		Name: name,
	}
}

// Parse:
//   const ::= "const" type identifier "=" expr
func (this *Parser) parseConst(start *Token) *ConstNode {
	ttype := this.parseType()
	if ttype == nil {
		return nil
	}

	name := this.need(TOK_IDENTIFIER)
	if name == nil {
		return nil
	}

	if this.need(TOK_ASSIGN) == nil {
		return nil
	}

	init := this.parseExpr()
	if init == nil {
		return nil
	}

	return &ConstNode{
		Range: Location{
			Start: start.Loc.Start,
			End:   this.scanner.Position(),
		},
		Type: ttype,
		Name: name,
		Init: init,
	}
}

// Parse:
//   body ::= statement*
//   statement ::= namespace
//               | include
//               | enum
//               | struct
//               | service
//               | const
func (this *Parser) parse() bool {
	for {
		tok := this.scanner.next()
		switch tok.Kind {
		case TOK_NAMESPACE:
			if !this.parseNamespace() {
				return false
			}

		case TOK_INCLUDE:
			if !this.parseInclude() {
				return false
			}

		case TOK_ENUM:
			node := this.parseEnum(tok)
			if node == nil {
				return false
			}
			this.tree.Nodes = append(this.tree.Nodes, node)

		case TOK_STRUCT, TOK_EXCEPTION:
			node := this.parseStruct(tok)
			if node == nil {
				return false
			}
			this.tree.Nodes = append(this.tree.Nodes, node)

		case TOK_SERVICE:
			node := this.parseService(tok)
			if node == nil {
				return false
			}
			this.tree.Nodes = append(this.tree.Nodes, node)

		case TOK_CONST:
			node := this.parseConst(tok)
			if node == nil {
				return false
			}
			this.tree.Nodes = append(this.tree.Nodes, node)

		case TOK_TYPEDEF:
			node := this.parseTypedef(tok)
			if node == nil {
				return false
			}
			this.tree.Nodes = append(this.tree.Nodes, node)

		case TOK_ERROR:
			return false

		case TOK_EOF:
			return true

		default:
			this.Context.ReportError(tok.Loc.Start, "expected definition, got: %s", tok.String())
		}
	}
}

func (this *Parser) Parse() *ParseTree {
	if !this.parse() {
		return nil
	}
	if this.Context.HasErrors() {
		return nil
	}
	return this.tree
}
