package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	identifier     = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]")
	prefixVariable = regexp.MustCompile("{\\w*}")
	defaultPrefix  = &ScopePrefix{String: "", Variables: make([]string, 0)}
)

type statementWrapper struct {
	comment   []string
	statement interface{}
}

type exception *Struct

type union *Struct

func newScopePrefix(prefix string) (*ScopePrefix, error) {
	variables := []string{}
	for _, variable := range prefixVariable.FindAllString(prefix, -1) {
		variable = variable[1 : len(variable)-1]
		if len(variable) == 0 || !identifier.MatchString(variable) {
			return nil, fmt.Errorf("parser: invalid prefix variable '%s'", variable)
		}
		variables = append(variables, variable)
	}
	return &ScopePrefix{String: prefix, Variables: variables}, nil
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func ifaceSliceToString(v interface{}) string {
	ifs := toIfaceSlice(v)
	b := make([]byte, len(ifs))
	for i, v := range ifs {
		b[i] = v.([]uint8)[0]
	}
	return string(b)
}

func rawCommentToDocStr(raw string) []string {
	rawLines := strings.Split(raw, "\n")
	comment := make([]string, len(rawLines))
	for i, line := range rawLines {
		comment[i] = strings.TrimLeft(line, "* ")
	}
	return comment
}

// toStruct converts a union to a struct with all fields optional.
func unionToStruct(u union) *Struct {
	st := (*Struct)(u)
	for _, f := range st.Fields {
		f.Modifier = Optional
	}
	return st
}

// toAnnotations converts an interface{} to an Annotation slice.
func toAnnotations(v interface{}) Annotations {
	if v == nil {
		return nil
	}
	return Annotations(v.([]*Annotation))
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 85, col: 1, offset: 2393},
			expr: &actionExpr{
				pos: position{line: 85, col: 12, offset: 2404},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 12, offset: 2404},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 12, offset: 2404},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 15, offset: 2407},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 26, offset: 2418},
								expr: &seqExpr{
									pos: position{line: 85, col: 28, offset: 2420},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 28, offset: 2420},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 38, offset: 2430},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 45, offset: 2437},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 45, offset: 2437},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 51, offset: 2443},
									name: "SyntaxError",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SyntaxError",
			pos:  position{line: 157, col: 1, offset: 5088},
			expr: &actionExpr{
				pos: position{line: 157, col: 16, offset: 5103},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 157, col: 16, offset: 5103,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 161, col: 1, offset: 5161},
			expr: &actionExpr{
				pos: position{line: 161, col: 14, offset: 5174},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 161, col: 14, offset: 5174},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 161, col: 14, offset: 5174},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 161, col: 21, offset: 5181},
								expr: &seqExpr{
									pos: position{line: 161, col: 22, offset: 5182},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 161, col: 22, offset: 5182},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 161, col: 32, offset: 5192},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 161, col: 37, offset: 5197},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 161, col: 48, offset: 5208},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 161, col: 48, offset: 5208},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 161, col: 66, offset: 5226},
										name: "FrugalStatement",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ThriftStatement",
			pos:  position{line: 174, col: 1, offset: 5697},
			expr: &choiceExpr{
				pos: position{line: 174, col: 20, offset: 5716},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 174, col: 20, offset: 5716},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 30, offset: 5726},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 42, offset: 5738},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 50, offset: 5746},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 57, offset: 5753},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 67, offset: 5763},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 76, offset: 5772},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 88, offset: 5784},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 174, col: 96, offset: 5792},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 176, col: 1, offset: 5801},
			expr: &actionExpr{
				pos: position{line: 176, col: 12, offset: 5812},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 176, col: 12, offset: 5812},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 176, col: 12, offset: 5812},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 176, col: 22, offset: 5822},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 176, col: 24, offset: 5824},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 176, col: 29, offset: 5829},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 176, col: 37, offset: 5837},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 176, col: 39, offset: 5839},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 176, col: 51, offset: 5851},
								expr: &ruleRefExpr{
									pos:  position{line: 176, col: 51, offset: 5851},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 176, col: 68, offset: 5868},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 188, col: 1, offset: 6145},
			expr: &actionExpr{
				pos: position{line: 188, col: 14, offset: 6158},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 188, col: 14, offset: 6158},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 188, col: 14, offset: 6158},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 26, offset: 6170},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 28, offset: 6172},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 188, col: 34, offset: 6178},
								expr: &charClassMatcher{
									pos:        position{line: 188, col: 34, offset: 6178},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 44, offset: 6188},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 46, offset: 6190},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 188, col: 49, offset: 6193},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 60, offset: 6204},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 62, offset: 6206},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 188, col: 74, offset: 6218},
								expr: &ruleRefExpr{
									pos:  position{line: 188, col: 74, offset: 6218},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 91, offset: 6235},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 196, col: 1, offset: 6421},
			expr: &actionExpr{
				pos: position{line: 196, col: 10, offset: 6430},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 196, col: 10, offset: 6430},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 196, col: 10, offset: 6430},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 18, offset: 6438},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 20, offset: 6440},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 24, offset: 6444},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 34, offset: 6454},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 36, offset: 6456},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 41, offset: 6461},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 52, offset: 6472},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 196, col: 54, offset: 6474},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 58, offset: 6478},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 60, offset: 6480},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 66, offset: 6486},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 77, offset: 6497},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 79, offset: 6499},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 196, col: 91, offset: 6511},
								expr: &ruleRefExpr{
									pos:  position{line: 196, col: 91, offset: 6511},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 108, offset: 6528},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 205, col: 1, offset: 6722},
			expr: &actionExpr{
				pos: position{line: 205, col: 9, offset: 6730},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 205, col: 9, offset: 6730},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 205, col: 9, offset: 6730},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 16, offset: 6737},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 18, offset: 6739},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 205, col: 23, offset: 6744},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 34, offset: 6755},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 205, col: 37, offset: 6758},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 41, offset: 6762},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 44, offset: 6765},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 205, col: 51, offset: 6772},
								expr: &seqExpr{
									pos: position{line: 205, col: 52, offset: 6773},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 205, col: 52, offset: 6773},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 205, col: 62, offset: 6783},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 205, col: 67, offset: 6788},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 71, offset: 6792},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 73, offset: 6794},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 205, col: 85, offset: 6806},
								expr: &ruleRefExpr{
									pos:  position{line: 205, col: 85, offset: 6806},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 102, offset: 6823},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 229, col: 1, offset: 7485},
			expr: &actionExpr{
				pos: position{line: 229, col: 14, offset: 7498},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 229, col: 14, offset: 7498},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 229, col: 14, offset: 7498},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 21, offset: 7505},
								expr: &seqExpr{
									pos: position{line: 229, col: 22, offset: 7506},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 229, col: 22, offset: 7506},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 32, offset: 7516},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 229, col: 37, offset: 7521},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 42, offset: 7526},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 53, offset: 7537},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 55, offset: 7539},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 61, offset: 7545},
								expr: &seqExpr{
									pos: position{line: 229, col: 62, offset: 7546},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 229, col: 62, offset: 7546},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 66, offset: 7550},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 68, offset: 7552},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 82, offset: 7566},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 84, offset: 7568},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 96, offset: 7580},
								expr: &ruleRefExpr{
									pos:  position{line: 229, col: 96, offset: 7580},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 229, col: 113, offset: 7597},
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 113, offset: 7597},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 245, col: 1, offset: 7995},
			expr: &actionExpr{
				pos: position{line: 245, col: 12, offset: 8006},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 245, col: 12, offset: 8006},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 245, col: 12, offset: 8006},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 22, offset: 8016},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 24, offset: 8018},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 245, col: 28, offset: 8022},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 38, offset: 8032},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 40, offset: 8034},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 245, col: 45, offset: 8039},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 56, offset: 8050},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 58, offset: 8052},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 245, col: 70, offset: 8064},
								expr: &ruleRefExpr{
									pos:  position{line: 245, col: 70, offset: 8064},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 87, offset: 8081},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 253, col: 1, offset: 8253},
			expr: &actionExpr{
				pos: position{line: 253, col: 11, offset: 8263},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 253, col: 11, offset: 8263},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 253, col: 11, offset: 8263},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 20, offset: 8272},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 22, offset: 8274},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 25, offset: 8277},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 254, col: 1, offset: 8317},
			expr: &actionExpr{
				pos: position{line: 254, col: 14, offset: 8330},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 254, col: 14, offset: 8330},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 254, col: 14, offset: 8330},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 26, offset: 8342},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 28, offset: 8344},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 31, offset: 8347},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 255, col: 1, offset: 8398},
			expr: &actionExpr{
				pos: position{line: 255, col: 10, offset: 8407},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 255, col: 10, offset: 8407},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 255, col: 10, offset: 8407},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 255, col: 18, offset: 8415},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 255, col: 20, offset: 8417},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 255, col: 23, offset: 8420},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 256, col: 1, offset: 8467},
			expr: &actionExpr{
				pos: position{line: 256, col: 15, offset: 8481},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 256, col: 15, offset: 8481},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 256, col: 15, offset: 8481},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 20, offset: 8486},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 31, offset: 8497},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 256, col: 34, offset: 8500},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 38, offset: 8504},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 41, offset: 8507},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 48, offset: 8514},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 256, col: 58, offset: 8524},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 62, offset: 8528},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 64, offset: 8530},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 76, offset: 8542},
								expr: &ruleRefExpr{
									pos:  position{line: 256, col: 76, offset: 8542},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 93, offset: 8559},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 267, col: 1, offset: 8776},
			expr: &actionExpr{
				pos: position{line: 267, col: 14, offset: 8789},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 267, col: 14, offset: 8789},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 267, col: 21, offset: 8796},
						expr: &seqExpr{
							pos: position{line: 267, col: 22, offset: 8797},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 267, col: 22, offset: 8797},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 267, col: 28, offset: 8803},
									name: "__",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Field",
			pos:  position{line: 276, col: 1, offset: 8984},
			expr: &actionExpr{
				pos: position{line: 276, col: 10, offset: 8993},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 276, col: 10, offset: 8993},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 276, col: 10, offset: 8993},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 17, offset: 9000},
								expr: &seqExpr{
									pos: position{line: 276, col: 18, offset: 9001},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 276, col: 18, offset: 9001},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 28, offset: 9011},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 276, col: 33, offset: 9016},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 36, offset: 9019},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 48, offset: 9031},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 276, col: 50, offset: 9033},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 54, offset: 9037},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 56, offset: 9039},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 60, offset: 9043},
								expr: &ruleRefExpr{
									pos:  position{line: 276, col: 60, offset: 9043},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 75, offset: 9058},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 77, offset: 9060},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 81, offset: 9064},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 91, offset: 9074},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 93, offset: 9076},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 98, offset: 9081},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 109, offset: 9092},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 112, offset: 9095},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 116, offset: 9099},
								expr: &seqExpr{
									pos: position{line: 276, col: 117, offset: 9100},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 276, col: 117, offset: 9100},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 121, offset: 9104},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 123, offset: 9106},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 136, offset: 9119},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 138, offset: 9121},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 150, offset: 9133},
								expr: &ruleRefExpr{
									pos:  position{line: 276, col: 150, offset: 9133},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 276, col: 167, offset: 9150},
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 167, offset: 9150},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 299, col: 1, offset: 9682},
			expr: &actionExpr{
				pos: position{line: 299, col: 18, offset: 9699},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 299, col: 19, offset: 9700},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 299, col: 19, offset: 9700},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 299, col: 32, offset: 9713},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 307, col: 1, offset: 9856},
			expr: &actionExpr{
				pos: position{line: 307, col: 12, offset: 9867},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 307, col: 12, offset: 9867},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 307, col: 12, offset: 9867},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 22, offset: 9877},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 24, offset: 9879},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 29, offset: 9884},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 40, offset: 9895},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 42, offset: 9897},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 50, offset: 9905},
								expr: &seqExpr{
									pos: position{line: 307, col: 51, offset: 9906},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 51, offset: 9906},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 61, offset: 9916},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 64, offset: 9919},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 75, offset: 9930},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 80, offset: 9935},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 307, col: 83, offset: 9938},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 87, offset: 9942},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 90, offset: 9945},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 307, col: 98, offset: 9953},
								expr: &seqExpr{
									pos: position{line: 307, col: 99, offset: 9954},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 99, offset: 9954},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 108, offset: 9963},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 307, col: 114, offset: 9969},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 307, col: 114, offset: 9969},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 307, col: 120, offset: 9975},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 139, offset: 9994},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 141, offset: 9996},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 153, offset: 10008},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 153, offset: 10008},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 170, offset: 10025},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 324, col: 1, offset: 10466},
			expr: &actionExpr{
				pos: position{line: 324, col: 22, offset: 10487},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 324, col: 22, offset: 10487,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 328, col: 1, offset: 10556},
			expr: &actionExpr{
				pos: position{line: 328, col: 13, offset: 10568},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 328, col: 13, offset: 10568},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 328, col: 13, offset: 10568},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 20, offset: 10575},
								expr: &seqExpr{
									pos: position{line: 328, col: 21, offset: 10576},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 328, col: 21, offset: 10576},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 328, col: 31, offset: 10586},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 328, col: 36, offset: 10591},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 43, offset: 10598},
								expr: &seqExpr{
									pos: position{line: 328, col: 44, offset: 10599},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 328, col: 44, offset: 10599},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 328, col: 53, offset: 10608},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 328, col: 58, offset: 10613},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 62, offset: 10617},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 75, offset: 10630},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 78, offset: 10633},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 83, offset: 10638},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 94, offset: 10649},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 328, col: 96, offset: 10651},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 100, offset: 10655},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 103, offset: 10658},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 113, offset: 10668},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 328, col: 123, offset: 10678},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 127, offset: 10682},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 130, offset: 10685},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 141, offset: 10696},
								expr: &ruleRefExpr{
									pos:  position{line: 328, col: 141, offset: 10696},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 149, offset: 10704},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 151, offset: 10706},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 163, offset: 10718},
								expr: &ruleRefExpr{
									pos:  position{line: 328, col: 163, offset: 10718},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 328, col: 180, offset: 10735},
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 180, offset: 10735},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 356, col: 1, offset: 11386},
			expr: &actionExpr{
				pos: position{line: 356, col: 17, offset: 11402},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 17, offset: 11402},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 22, offset: 11407},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 356, col: 22, offset: 11407},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 31, offset: 11416},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 363, col: 1, offset: 11538},
			expr: &actionExpr{
				pos: position{line: 363, col: 11, offset: 11548},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 363, col: 11, offset: 11548},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 363, col: 11, offset: 11548},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 20, offset: 11557},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 363, col: 23, offset: 11560},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 27, offset: 11564},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 363, col: 30, offset: 11567},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 41, offset: 11578},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 363, col: 51, offset: 11588},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 367, col: 1, offset: 11624},
			expr: &actionExpr{
				pos: position{line: 367, col: 14, offset: 11637},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 367, col: 14, offset: 11637},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 367, col: 19, offset: 11642},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 367, col: 19, offset: 11642},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 367, col: 30, offset: 11653},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 367, col: 46, offset: 11669},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 374, col: 1, offset: 11794},
			expr: &actionExpr{
				pos: position{line: 374, col: 13, offset: 11806},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 374, col: 13, offset: 11806},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 374, col: 13, offset: 11806},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 374, col: 18, offset: 11811},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 31, offset: 11824},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 374, col: 33, offset: 11826},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 374, col: 45, offset: 11838},
								expr: &ruleRefExpr{
									pos:  position{line: 374, col: 45, offset: 11838},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BaseTypeName",
			pos:  position{line: 381, col: 1, offset: 11974},
			expr: &actionExpr{
				pos: position{line: 381, col: 17, offset: 11990},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 381, col: 18, offset: 11991},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 381, col: 18, offset: 11991},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 27, offset: 12000},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 36, offset: 12009},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 44, offset: 12017},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 52, offset: 12025},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 60, offset: 12033},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 71, offset: 12044},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 82, offset: 12055},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 385, col: 1, offset: 12102},
			expr: &actionExpr{
				pos: position{line: 385, col: 18, offset: 12119},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 385, col: 18, offset: 12119},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 385, col: 23, offset: 12124},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 385, col: 23, offset: 12124},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 385, col: 33, offset: 12134},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 385, col: 43, offset: 12144},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 389, col: 1, offset: 12179},
			expr: &actionExpr{
				pos: position{line: 389, col: 12, offset: 12190},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 389, col: 12, offset: 12190},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 389, col: 12, offset: 12190},
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 12, offset: 12190},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 389, col: 21, offset: 12199},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 28, offset: 12206},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 31, offset: 12209},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 35, offset: 12213},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 45, offset: 12223},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 389, col: 48, offset: 12226},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 52, offset: 12230},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 55, offset: 12233},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 61, offset: 12239},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 71, offset: 12249},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 389, col: 74, offset: 12252},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 78, offset: 12256},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 80, offset: 12258},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 389, col: 92, offset: 12270},
								expr: &ruleRefExpr{
									pos:  position{line: 389, col: 92, offset: 12270},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 398, col: 1, offset: 12468},
			expr: &actionExpr{
				pos: position{line: 398, col: 12, offset: 12479},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 398, col: 12, offset: 12479},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 398, col: 12, offset: 12479},
							expr: &ruleRefExpr{
								pos:  position{line: 398, col: 12, offset: 12479},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 398, col: 21, offset: 12488},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 28, offset: 12495},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 31, offset: 12498},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 398, col: 35, offset: 12502},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 45, offset: 12512},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 398, col: 48, offset: 12515},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 52, offset: 12519},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 54, offset: 12521},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 398, col: 66, offset: 12533},
								expr: &ruleRefExpr{
									pos:  position{line: 398, col: 66, offset: 12533},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 406, col: 1, offset: 12695},
			expr: &actionExpr{
				pos: position{line: 406, col: 13, offset: 12707},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 406, col: 13, offset: 12707},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 406, col: 13, offset: 12707},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 21, offset: 12715},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 406, col: 24, offset: 12718},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 406, col: 28, offset: 12722},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 38, offset: 12732},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 406, col: 41, offset: 12735},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 45, offset: 12739},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 406, col: 47, offset: 12741},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 406, col: 59, offset: 12753},
								expr: &ruleRefExpr{
									pos:  position{line: 406, col: 59, offset: 12753},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 414, col: 1, offset: 12916},
			expr: &actionExpr{
				pos: position{line: 414, col: 12, offset: 12927},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 414, col: 12, offset: 12927},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 414, col: 12, offset: 12927},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 414, col: 23, offset: 12938},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 414, col: 31, offset: 12946},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 418, col: 1, offset: 12983},
			expr: &choiceExpr{
				pos: position{line: 418, col: 15, offset: 12997},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 418, col: 15, offset: 12997},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 25, offset: 13007},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 40, offset: 13022},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 57, offset: 13039},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 71, offset: 13053},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 82, offset: 13064},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 94, offset: 13076},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 420, col: 1, offset: 13088},
			expr: &actionExpr{
				pos: position{line: 420, col: 20, offset: 13107},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 420, col: 20, offset: 13107},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 420, col: 20, offset: 13107},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 24, offset: 13111},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 27, offset: 13114},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 420, col: 39, offset: 13126},
								expr: &ruleRefExpr{
									pos:  position{line: 420, col: 39, offset: 13126},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 420, col: 55, offset: 13142},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 428, col: 1, offset: 13306},
			expr: &actionExpr{
				pos: position{line: 428, col: 19, offset: 13324},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 428, col: 19, offset: 13324},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 428, col: 19, offset: 13324},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 428, col: 24, offset: 13329},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 428, col: 35, offset: 13340},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 428, col: 37, offset: 13342},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 428, col: 43, offset: 13348},
								expr: &actionExpr{
									pos: position{line: 428, col: 44, offset: 13349},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 428, col: 44, offset: 13349},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 428, col: 44, offset: 13349},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 428, col: 48, offset: 13353},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 428, col: 51, offset: 13356},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 428, col: 57, offset: 13362},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 428, col: 89, offset: 13394},
							expr: &ruleRefExpr{
								pos:  position{line: 428, col: 89, offset: 13394},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 428, col: 104, offset: 13409},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 439, col: 1, offset: 13605},
			expr: &actionExpr{
				pos: position{line: 439, col: 17, offset: 13621},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 439, col: 18, offset: 13622},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 439, col: 18, offset: 13622},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 439, col: 27, offset: 13631},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 443, col: 1, offset: 13686},
			expr: &actionExpr{
				pos: position{line: 443, col: 16, offset: 13701},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 443, col: 16, offset: 13701},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 443, col: 16, offset: 13701},
							expr: &charClassMatcher{
								pos:        position{line: 443, col: 16, offset: 13701},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 443, col: 22, offset: 13707},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 22, offset: 13707},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 447, col: 1, offset: 13771},
			expr: &actionExpr{
				pos: position{line: 447, col: 19, offset: 13789},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 447, col: 19, offset: 13789},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 447, col: 19, offset: 13789},
							expr: &charClassMatcher{
								pos:        position{line: 447, col: 19, offset: 13789},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 447, col: 25, offset: 13795},
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 25, offset: 13795},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 447, col: 32, offset: 13802},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 447, col: 36, offset: 13806},
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 36, offset: 13806},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 447, col: 43, offset: 13813},
							expr: &seqExpr{
								pos: position{line: 447, col: 45, offset: 13815},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 447, col: 45, offset: 13815},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 447, col: 52, offset: 13822},
										name: "IntConstant",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ConstList",
			pos:  position{line: 451, col: 1, offset: 13892},
			expr: &actionExpr{
				pos: position{line: 451, col: 14, offset: 13905},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 451, col: 14, offset: 13905},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 451, col: 14, offset: 13905},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 18, offset: 13909},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 451, col: 21, offset: 13912},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 451, col: 28, offset: 13919},
								expr: &seqExpr{
									pos: position{line: 451, col: 29, offset: 13920},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 451, col: 29, offset: 13920},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 451, col: 40, offset: 13931},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 451, col: 43, offset: 13934},
											expr: &ruleRefExpr{
												pos:  position{line: 451, col: 43, offset: 13934},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 451, col: 58, offset: 13949},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 63, offset: 13954},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 451, col: 66, offset: 13957},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 460, col: 1, offset: 14151},
			expr: &actionExpr{
				pos: position{line: 460, col: 13, offset: 14163},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 460, col: 13, offset: 14163},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 460, col: 13, offset: 14163},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 460, col: 17, offset: 14167},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 460, col: 20, offset: 14170},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 460, col: 27, offset: 14177},
								expr: &seqExpr{
									pos: position{line: 460, col: 28, offset: 14178},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 460, col: 28, offset: 14178},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 39, offset: 14189},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 460, col: 42, offset: 14192},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 46, offset: 14196},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 49, offset: 14199},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 60, offset: 14210},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 460, col: 64, offset: 14214},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 460, col: 64, offset: 14214},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 460, col: 70, offset: 14220},
													expr: &litMatcher{
														pos:        position{line: 460, col: 71, offset: 14221},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 76, offset: 14226},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 460, col: 81, offset: 14231},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 480, col: 1, offset: 14781},
			expr: &ruleRefExpr{
				pos:  position{line: 480, col: 20, offset: 14800},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 482, col: 1, offset: 14807},
			expr: &actionExpr{
				pos: position{line: 482, col: 10, offset: 14816},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 482, col: 10, offset: 14816},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 482, col: 10, offset: 14816},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 17, offset: 14823},
								expr: &seqExpr{
									pos: position{line: 482, col: 18, offset: 14824},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 482, col: 18, offset: 14824},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 482, col: 28, offset: 14834},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 482, col: 33, offset: 14839},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 41, offset: 14847},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 44, offset: 14850},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 482, col: 49, offset: 14855},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 60, offset: 14866},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 63, offset: 14869},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 70, offset: 14876},
								expr: &ruleRefExpr{
									pos:  position{line: 482, col: 70, offset: 14876},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 78, offset: 14884},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 482, col: 81, offset: 14887},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 85, offset: 14891},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 88, offset: 14894},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 482, col: 99, offset: 14905},
								expr: &seqExpr{
									pos: position{line: 482, col: 100, offset: 14906},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 482, col: 100, offset: 14906},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 482, col: 110, offset: 14916},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 482, col: 116, offset: 14922},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 482, col: 116, offset: 14922},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 482, col: 122, offset: 14928},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 139, offset: 14945},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 141, offset: 14947},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 153, offset: 14959},
								expr: &ruleRefExpr{
									pos:  position{line: 482, col: 153, offset: 14959},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 170, offset: 14976},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 504, col: 1, offset: 15573},
			expr: &actionExpr{
				pos: position{line: 504, col: 20, offset: 15592},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 504, col: 20, offset: 15592,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 508, col: 1, offset: 15659},
			expr: &actionExpr{
				pos: position{line: 508, col: 11, offset: 15669},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 508, col: 11, offset: 15669},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 508, col: 11, offset: 15669},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 20, offset: 15678},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 23, offset: 15681},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 508, col: 35, offset: 15693},
							expr: &seqExpr{
								pos: position{line: 508, col: 36, offset: 15694},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 508, col: 36, offset: 15694},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 508, col: 40, offset: 15698},
										name: "PrefixToken",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "PrefixToken",
			pos:  position{line: 513, col: 1, offset: 15829},
			expr: &choiceExpr{
				pos: position{line: 513, col: 16, offset: 15844},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 513, col: 17, offset: 15845},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 513, col: 17, offset: 15845},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 513, col: 21, offset: 15849},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 513, col: 32, offset: 15860},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 513, col: 39, offset: 15867},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 515, col: 1, offset: 15879},
			expr: &oneOrMoreExpr{
				pos: position{line: 515, col: 15, offset: 15893},
				expr: &charClassMatcher{
					pos:        position{line: 515, col: 15, offset: 15893},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 517, col: 1, offset: 15911},
			expr: &actionExpr{
				pos: position{line: 517, col: 14, offset: 15924},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 517, col: 14, offset: 15924},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 517, col: 14, offset: 15924},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 517, col: 21, offset: 15931},
								expr: &seqExpr{
									pos: position{line: 517, col: 22, offset: 15932},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 517, col: 22, offset: 15932},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 517, col: 32, offset: 15942},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 517, col: 37, offset: 15947},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 42, offset: 15952},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 53, offset: 15963},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 517, col: 55, offset: 15965},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 59, offset: 15969},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 517, col: 62, offset: 15972},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 66, offset: 15976},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 77, offset: 15987},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 517, col: 79, offset: 15989},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 517, col: 91, offset: 16001},
								expr: &ruleRefExpr{
									pos:  position{line: 517, col: 91, offset: 16001},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 517, col: 108, offset: 16018},
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 108, offset: 16018},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 534, col: 1, offset: 16604},
			expr: &actionExpr{
				pos: position{line: 534, col: 12, offset: 16615},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 534, col: 13, offset: 16616},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 534, col: 14, offset: 16617},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 534, col: 14, offset: 16617},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 534, col: 18, offset: 16621},
									expr: &choiceExpr{
										pos: position{line: 534, col: 19, offset: 16622},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 534, col: 19, offset: 16622},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 534, col: 26, offset: 16629},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 534, col: 33, offset: 16636},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 534, col: 41, offset: 16644},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 534, col: 41, offset: 16644},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 534, col: 46, offset: 16649},
									expr: &choiceExpr{
										pos: position{line: 534, col: 47, offset: 16650},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 534, col: 47, offset: 16650},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 534, col: 54, offset: 16657},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 534, col: 61, offset: 16664},
									val:        "'",
									ignoreCase: false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Identifier",
			pos:  position{line: 543, col: 1, offset: 16950},
			expr: &actionExpr{
				pos: position{line: 543, col: 15, offset: 16964},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 543, col: 15, offset: 16964},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 543, col: 15, offset: 16964},
							expr: &choiceExpr{
								pos: position{line: 543, col: 16, offset: 16965},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 543, col: 16, offset: 16965},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 543, col: 25, offset: 16974},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 543, col: 31, offset: 16980},
							expr: &choiceExpr{
								pos: position{line: 543, col: 32, offset: 16981},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 543, col: 32, offset: 16981},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 543, col: 41, offset: 16990},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 543, col: 49, offset: 16998},
										val:        "[._]",
										chars:      []rune{'.', '_'},
										ignoreCase: false,
										inverted:   false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ListSeparator",
			pos:  position{line: 547, col: 1, offset: 17053},
			expr: &charClassMatcher{
				pos:        position{line: 547, col: 18, offset: 17070},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 548, col: 1, offset: 17075},
			expr: &charClassMatcher{
				pos:        position{line: 548, col: 11, offset: 17085},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 549, col: 1, offset: 17094},
			expr: &charClassMatcher{
				pos:        position{line: 549, col: 10, offset: 17103},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 551, col: 1, offset: 17110},
			expr: &anyMatcher{
				line: 551, col: 15, offset: 17124,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 552, col: 1, offset: 17126},
			expr: &actionExpr{
				pos: position{line: 552, col: 14, offset: 17139},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 552, col: 14, offset: 17139},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 552, col: 14, offset: 17139},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 552, col: 21, offset: 17146},
							expr: &seqExpr{
								pos: position{line: 552, col: 23, offset: 17148},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 552, col: 23, offset: 17148},
										expr: &litMatcher{
											pos:        position{line: 552, col: 24, offset: 17149},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 552, col: 29, offset: 17154},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 552, col: 43, offset: 17168},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 558, col: 1, offset: 17348},
			expr: &choiceExpr{
				pos: position{line: 558, col: 12, offset: 17359},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 558, col: 12, offset: 17359},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 558, col: 31, offset: 17378},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 559, col: 1, offset: 17396},
			expr: &seqExpr{
				pos: position{line: 559, col: 21, offset: 17416},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 559, col: 21, offset: 17416},
						expr: &ruleRefExpr{
							pos:  position{line: 559, col: 22, offset: 17417},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 559, col: 32, offset: 17427},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 559, col: 37, offset: 17432},
						expr: &seqExpr{
							pos: position{line: 559, col: 39, offset: 17434},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 559, col: 39, offset: 17434},
									expr: &litMatcher{
										pos:        position{line: 559, col: 40, offset: 17435},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 559, col: 45, offset: 17440},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 559, col: 59, offset: 17454},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 560, col: 1, offset: 17459},
			expr: &seqExpr{
				pos: position{line: 560, col: 37, offset: 17495},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 560, col: 37, offset: 17495},
						expr: &ruleRefExpr{
							pos:  position{line: 560, col: 38, offset: 17496},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 560, col: 48, offset: 17506},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 560, col: 53, offset: 17511},
						expr: &seqExpr{
							pos: position{line: 560, col: 55, offset: 17513},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 560, col: 55, offset: 17513},
									expr: &choiceExpr{
										pos: position{line: 560, col: 58, offset: 17516},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 560, col: 58, offset: 17516},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 560, col: 65, offset: 17523},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 560, col: 71, offset: 17529},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 560, col: 85, offset: 17543},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 561, col: 1, offset: 17548},
			expr: &choiceExpr{
				pos: position{line: 561, col: 22, offset: 17569},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 561, col: 23, offset: 17570},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 561, col: 23, offset: 17570},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 561, col: 28, offset: 17575},
								expr: &seqExpr{
									pos: position{line: 561, col: 30, offset: 17577},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 561, col: 30, offset: 17577},
											expr: &ruleRefExpr{
												pos:  position{line: 561, col: 31, offset: 17578},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 561, col: 35, offset: 17582},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 561, col: 53, offset: 17600},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 561, col: 53, offset: 17600},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 561, col: 57, offset: 17604},
								expr: &seqExpr{
									pos: position{line: 561, col: 59, offset: 17606},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 561, col: 59, offset: 17606},
											expr: &ruleRefExpr{
												pos:  position{line: 561, col: 60, offset: 17607},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 561, col: 64, offset: 17611},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "__",
			pos:  position{line: 563, col: 1, offset: 17627},
			expr: &zeroOrMoreExpr{
				pos: position{line: 563, col: 7, offset: 17633},
				expr: &choiceExpr{
					pos: position{line: 563, col: 9, offset: 17635},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 563, col: 9, offset: 17635},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 563, col: 22, offset: 17648},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 563, col: 28, offset: 17654},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 564, col: 1, offset: 17665},
			expr: &zeroOrMoreExpr{
				pos: position{line: 564, col: 6, offset: 17670},
				expr: &choiceExpr{
					pos: position{line: 564, col: 8, offset: 17672},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 564, col: 8, offset: 17672},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 564, col: 21, offset: 17685},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 565, col: 1, offset: 17721},
			expr: &zeroOrMoreExpr{
				pos: position{line: 565, col: 7, offset: 17727},
				expr: &ruleRefExpr{
					pos:  position{line: 565, col: 7, offset: 17727},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 567, col: 1, offset: 17740},
			expr: &charClassMatcher{
				pos:        position{line: 567, col: 15, offset: 17754},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 568, col: 1, offset: 17762},
			expr: &litMatcher{
				pos:        position{line: 568, col: 8, offset: 17769},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 569, col: 1, offset: 17774},
			expr: &choiceExpr{
				pos: position{line: 569, col: 8, offset: 17781},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 569, col: 8, offset: 17781},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 8, offset: 17781},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 569, col: 11, offset: 17784},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 569, col: 17, offset: 17790},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 17, offset: 17790},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 569, col: 19, offset: 17792},
								expr: &ruleRefExpr{
									pos:  position{line: 569, col: 19, offset: 17792},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 569, col: 38, offset: 17811},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 569, col: 44, offset: 17817},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 44, offset: 17817},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 569, col: 47, offset: 17820},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 571, col: 1, offset: 17825},
			expr: &notExpr{
				pos: position{line: 571, col: 8, offset: 17832},
				expr: &anyMatcher{
					line: 571, col: 9, offset: 17833,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:       []*Include{},
		Namespaces:     []*Namespace{},
		Typedefs:       []*TypeDef{},
		Constants:      []*Constant{},
		Enums:          []*Enum{},
		Structs:        []*Struct{},
		Exceptions:     []*Struct{},
		Unions:         []*Struct{},
		Services:       []*Service{},
		typedefIndex:   make(map[string]*TypeDef),
		namespaceIndex: make(map[string]*Namespace),
	}
	stmts := toIfaceSlice(statements)
	frugal := &Frugal{
		Thrift:             thrift,
		Scopes:             []*Scope{},
		ParsedIncludes:     make(map[string]*Frugal),
		OrderedDefinitions: make([]interface{}, len(stmts)),
	}

	for i, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
		frugal.OrderedDefinitions[i] = wrapper.statement
		switch v := wrapper.statement.(type) {
		case *Namespace:
			thrift.Namespaces = append(thrift.Namespaces, v)
			thrift.namespaceIndex[v.Scope] = v
		case *Constant:
			v.Comment = wrapper.comment
			thrift.Constants = append(thrift.Constants, v)
		case *Enum:
			v.Comment = wrapper.comment
			thrift.Enums = append(thrift.Enums, v)
		case *TypeDef:
			v.Comment = wrapper.comment
			thrift.Typedefs = append(thrift.Typedefs, v)
			thrift.typedefIndex[v.Name] = v
		case *Struct:
			v.Type = StructTypeStruct
			v.Comment = wrapper.comment
			thrift.Structs = append(thrift.Structs, v)
		case exception:
			strct := (*Struct)(v)
			strct.Type = StructTypeException
			strct.Comment = wrapper.comment
			thrift.Exceptions = append(thrift.Exceptions, strct)
			frugal.OrderedDefinitions[i] = strct
		case union:
			strct := unionToStruct(v)
			strct.Type = StructTypeUnion
			strct.Comment = wrapper.comment
			thrift.Unions = append(thrift.Unions, strct)
			frugal.OrderedDefinitions[i] = strct
		case *Service:
			v.Comment = wrapper.comment
			v.Thrift = frugal.Thrift
			thrift.Services = append(thrift.Services, v)
		case *Include:
			thrift.Includes = append(thrift.Includes, v)
		case *Scope:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Scopes = append(frugal.Scopes, v)
		default:
			return nil, fmt.Errorf("parser: unknown value %#v", v)
		}
	}
	return frugal, nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["statements"])
}

func (c *current) onSyntaxError1() (interface{}, error) {
	return nil, errors.New("parser: syntax error")
}

func (p *parser) callonSyntaxError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSyntaxError1()
}

func (c *current) onStatement1(docstr, statement interface{}) (interface{}, error) {
	wrapper := &statementWrapper{statement: statement}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		wrapper.comment = rawCommentToDocStr(raw)
	}
	return wrapper, nil
}

func (p *parser) callonStatement1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStatement1(stack["docstr"], stack["statement"])
}

func (c *current) onInclude1(file, annotations interface{}) (interface{}, error) {
	name := filepath.Base(file.(string))
	if ix := strings.LastIndex(name, "."); ix > 0 {
		name = name[:ix]
	}
	return &Include{
		Name:        name,
		Value:       file.(string),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"], stack["annotations"])
}

func (c *current) onNamespace1(scope, ns, annotations interface{}) (interface{}, error) {
	return &Namespace{
		Scope:       ifaceSliceToString(scope),
		Value:       string(ns.(Identifier)),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonNamespace1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNamespace1(stack["scope"], stack["ns"], stack["annotations"])
}

func (c *current) onConst1(typ, name, value, annotations interface{}) (interface{}, error) {
	return &Constant{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Value:       value,
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonConst1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConst1(stack["typ"], stack["name"], stack["value"], stack["annotations"])
}

func (c *current) onEnum1(name, values, annotations interface{}) (interface{}, error) {
	vs := toIfaceSlice(values)
	en := &Enum{
		Name:        string(name.(Identifier)),
		Values:      make([]*EnumValue, len(vs)),
		Annotations: toAnnotations(annotations),
	}
	// Assigns numbers in order. This will behave badly if some values are
	// defined and other are not, but I think that's ok since that's a silly
	// thing to do.
	next := 0
	for idx, v := range vs {
		ev := v.([]interface{})[0].(*EnumValue)
		if ev.Value < 0 {
			ev.Value = next
		}
		if ev.Value >= next {
			next = ev.Value + 1
		}
		en.Values[idx] = ev
	}
	return en, nil
}

func (p *parser) callonEnum1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnum1(stack["name"], stack["values"], stack["annotations"])
}

func (c *current) onEnumValue1(docstr, name, value, annotations interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:        string(name.(Identifier)),
		Value:       -1,
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		ev.Comment = rawCommentToDocStr(raw)
	}
	if value != nil {
		ev.Value = int(value.([]interface{})[2].(int64))
	}
	return ev, nil
}

func (p *parser) callonEnumValue1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnumValue1(stack["docstr"], stack["name"], stack["value"], stack["annotations"])
}

func (c *current) onTypeDef1(typ, name, annotations interface{}) (interface{}, error) {
	return &TypeDef{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonTypeDef1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeDef1(stack["typ"], stack["name"], stack["annotations"])
}

func (c *current) onStruct1(st interface{}) (interface{}, error) {
	return st.(*Struct), nil
}

func (p *parser) callonStruct1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStruct1(stack["st"])
}

func (c *current) onException1(st interface{}) (interface{}, error) {
	return exception(st.(*Struct)), nil
}

func (p *parser) callonException1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onException1(stack["st"])
}

func (c *current) onUnion1(st interface{}) (interface{}, error) {
	return union(st.(*Struct)), nil
}

func (p *parser) callonUnion1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnion1(stack["st"])
}

func (c *current) onStructLike1(name, fields, annotations interface{}) (interface{}, error) {
	st := &Struct{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
	}
	if fields != nil {
		st.Fields = fields.([]*Field)
	}
	return st, nil
}

func (p *parser) callonStructLike1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStructLike1(stack["name"], stack["fields"], stack["annotations"])
}

func (c *current) onFieldList1(fields interface{}) (interface{}, error) {
	fs := fields.([]interface{})
	flds := make([]*Field, len(fs))
	for i, f := range fs {
		flds[i] = f.([]interface{})[0].(*Field)
	}
	return flds, nil
}

func (p *parser) callonFieldList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldList1(stack["fields"])
}

func (c *current) onField1(docstr, id, mod, typ, name, def, annotations interface{}) (interface{}, error) {
	f := &Field{
		ID:          int(id.(int64)),
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		f.Comment = rawCommentToDocStr(raw)
	}
	if mod != nil {
		f.Modifier = mod.(FieldModifier)
	} else {
		f.Modifier = Default
	}

	if def != nil {
		f.Default = def.([]interface{})[2]
	}
	return f, nil
}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["docstr"], stack["id"], stack["mod"], stack["typ"], stack["name"], stack["def"], stack["annotations"])
}

func (c *current) onFieldModifier1() (interface{}, error) {
	if bytes.Equal(c.text, []byte("required")) {
		return Required, nil
	} else {
		return Optional, nil
	}
}

func (p *parser) callonFieldModifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldModifier1()
}

func (c *current) onService1(name, extends, methods, annotations interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	svc := &Service{
		Name:        string(name.(Identifier)),
		Methods:     make([]*Method, len(ms)),
		Annotations: toAnnotations(annotations),
	}
	if extends != nil {
		svc.Extends = string(extends.([]interface{})[2].(Identifier))
	}
	for i, m := range ms {
		mt := m.([]interface{})[0].(*Method)
		svc.Methods[i] = mt
	}
	return svc, nil
}

func (p *parser) callonService1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onService1(stack["name"], stack["extends"], stack["methods"], stack["annotations"])
}

func (c *current) onEndOfServiceError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of service")
}

func (p *parser) callonEndOfServiceError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfServiceError1()
}

func (c *current) onFunction1(docstr, oneway, typ, name, arguments, exceptions, annotations interface{}) (interface{}, error) {
	m := &Method{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		m.Comment = rawCommentToDocStr(raw)
	}
	t := typ.(*Type)
	if t.Name != "void" {
		m.ReturnType = t
	}
	if oneway != nil {
		m.Oneway = true
	}
	if arguments != nil {
		m.Arguments = arguments.([]*Field)
	}
	if exceptions != nil {
		m.Exceptions = exceptions.([]*Field)
		for _, e := range m.Exceptions {
			e.Modifier = Optional
		}
	}
	return m, nil
}

func (p *parser) callonFunction1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunction1(stack["docstr"], stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"], stack["annotations"])
}

func (c *current) onFunctionType1(typ interface{}) (interface{}, error) {
	if t, ok := typ.(*Type); ok {
		return t, nil
	}
	return &Type{Name: string(c.text)}, nil
}

func (p *parser) callonFunctionType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionType1(stack["typ"])
}

func (c *current) onThrows1(exceptions interface{}) (interface{}, error) {
	return exceptions, nil
}

func (p *parser) callonThrows1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onThrows1(stack["exceptions"])
}

func (c *current) onFieldType1(typ interface{}) (interface{}, error) {
	if t, ok := typ.(Identifier); ok {
		return &Type{Name: string(t)}, nil
	}
	return typ, nil
}

func (p *parser) callonFieldType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldType1(stack["typ"])
}

func (c *current) onBaseType1(name, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        name.(string),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonBaseType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseType1(stack["name"], stack["annotations"])
}

func (c *current) onBaseTypeName1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonBaseTypeName1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseTypeName1()
}

func (c *current) onContainerType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonContainerType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onContainerType1(stack["typ"])
}

func (c *current) onMapType1(key, value, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "map",
		KeyType:     key.(*Type),
		ValueType:   value.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonMapType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMapType1(stack["key"], stack["value"], stack["annotations"])
}

func (c *current) onSetType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "set",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonSetType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSetType1(stack["typ"], stack["annotations"])
}

func (c *current) onListType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "list",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonListType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListType1(stack["typ"], stack["annotations"])
}

func (c *current) onCppType1(cppType interface{}) (interface{}, error) {
	return cppType, nil
}

func (p *parser) callonCppType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCppType1(stack["cppType"])
}

func (c *current) onTypeAnnotations1(annotations interface{}) (interface{}, error) {
	var anns []*Annotation
	for _, ann := range annotations.([]interface{}) {
		anns = append(anns, ann.(*Annotation))
	}
	return anns, nil
}

func (p *parser) callonTypeAnnotations1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotations1(stack["annotations"])
}

func (c *current) onTypeAnnotation8(value interface{}) (interface{}, error) {
	return value, nil
}

func (p *parser) callonTypeAnnotation8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation8(stack["value"])
}

func (c *current) onTypeAnnotation1(name, value interface{}) (interface{}, error) {
	var optValue string
	if value != nil {
		optValue = value.(string)
	}
	return &Annotation{
		Name:  string(name.(Identifier)),
		Value: optValue,
	}, nil
}

func (p *parser) callonTypeAnnotation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation1(stack["name"], stack["value"])
}

func (c *current) onBoolConstant1() (interface{}, error) {
	return string(c.text) == "true", nil
}

func (p *parser) callonBoolConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBoolConstant1()
}

func (c *current) onIntConstant1() (interface{}, error) {
	return strconv.ParseInt(string(c.text), 10, 64)
}

func (p *parser) callonIntConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIntConstant1()
}

func (c *current) onDoubleConstant1() (interface{}, error) {
	return strconv.ParseFloat(string(c.text), 64)
}

func (p *parser) callonDoubleConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDoubleConstant1()
}

func (c *current) onConstList1(values interface{}) (interface{}, error) {
	valueSlice := values.([]interface{})
	vs := make([]interface{}, len(valueSlice))
	for i, v := range valueSlice {
		vs[i] = v.([]interface{})[0]
	}
	return vs, nil
}

func (p *parser) callonConstList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConstList1(stack["values"])
}

func (c *current) onConstMap1(values interface{}) (interface{}, error) {
	if values == nil {
		return nil, nil
	}
	vals := values.([]interface{})
	kvs := make([]KeyValue, len(vals))
	for i, kv := range vals {
		v := kv.([]interface{})
		kvs[i] = KeyValue{
			Key:   v[0],
			Value: v[4],
		}
	}
	return kvs, nil
}

func (p *parser) callonConstMap1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConstMap1(stack["values"])
}

func (c *current) onScope1(docstr, name, prefix, operations, annotations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:        string(name.(Identifier)),
		Operations:  make([]*Operation, len(ops)),
		Prefix:      defaultPrefix,
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		scope.Comment = rawCommentToDocStr(raw)
	}
	if prefix != nil {
		scope.Prefix = prefix.(*ScopePrefix)
	}
	for i, o := range ops {
		op := o.([]interface{})[0].(*Operation)
		scope.Operations[i] = op
	}
	return scope, nil
}

func (p *parser) callonScope1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onScope1(stack["docstr"], stack["name"], stack["prefix"], stack["operations"], stack["annotations"])
}

func (c *current) onEndOfScopeError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of scope")
}

func (p *parser) callonEndOfScopeError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfScopeError1()
}

func (c *current) onPrefix1() (interface{}, error) {
	prefix := strings.TrimSpace(strings.TrimPrefix(string(c.text), "prefix"))
	return newScopePrefix(prefix)
}

func (p *parser) callonPrefix1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrefix1()
}

func (c *current) onOperation1(docstr, name, typ, annotations interface{}) (interface{}, error) {
	o := &Operation{
		Name:        string(name.(Identifier)),
		Type:        &Type{Name: string(typ.(Identifier))},
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		o.Comment = rawCommentToDocStr(raw)
	}
	return o, nil
}

func (p *parser) callonOperation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["typ"], stack["annotations"])
}

func (c *current) onLiteral1() (interface{}, error) {
	if len(c.text) != 0 && c.text[0] == '\'' {
		intermediate := strings.Replace(string(c.text[1:len(c.text)-1]), `\'`, `'`, -1)
		return strconv.Unquote(`"` + strings.Replace(intermediate, `"`, `\"`, -1) + `"`)
	}

	return strconv.Unquote(string(c.text))
}

func (p *parser) callonLiteral1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLiteral1()
}

func (c *current) onIdentifier1() (interface{}, error) {
	return Identifier(string(c.text)), nil
}

func (p *parser) callonIdentifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifier1()
}

func (c *current) onDocString1() (interface{}, error) {
	comment := string(c.text)
	comment = strings.TrimPrefix(comment, "/**@")
	comment = strings.TrimSuffix(comment, "*/")
	return strings.TrimSpace(comment), nil
}

func (p *parser) callonDocString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDocString1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug   bool
	depth   int

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
	p.errs.add(pe)
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n > 0 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
