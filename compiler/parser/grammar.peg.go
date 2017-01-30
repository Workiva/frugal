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
			pos:  position{line: 150, col: 1, offset: 4800},
			expr: &actionExpr{
				pos: position{line: 150, col: 16, offset: 4815},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 150, col: 16, offset: 4815,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 154, col: 1, offset: 4873},
			expr: &actionExpr{
				pos: position{line: 154, col: 14, offset: 4886},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 154, col: 14, offset: 4886},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 154, col: 14, offset: 4886},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 154, col: 21, offset: 4893},
								expr: &seqExpr{
									pos: position{line: 154, col: 22, offset: 4894},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 154, col: 22, offset: 4894},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 154, col: 32, offset: 4904},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 154, col: 37, offset: 4909},
							label: "statement",
							expr: &ruleRefExpr{
								pos:  position{line: 154, col: 47, offset: 4919},
								name: "FrugalStatement",
							},
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 167, col: 1, offset: 5389},
			expr: &choiceExpr{
				pos: position{line: 167, col: 20, offset: 5408},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 167, col: 20, offset: 5408},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 30, offset: 5418},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 42, offset: 5430},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 50, offset: 5438},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 57, offset: 5445},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 67, offset: 5455},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 76, offset: 5464},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 88, offset: 5476},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 96, offset: 5484},
						name: "Service",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 106, offset: 5494},
						name: "Scope",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 169, col: 1, offset: 5501},
			expr: &actionExpr{
				pos: position{line: 169, col: 12, offset: 5512},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 169, col: 12, offset: 5512},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 169, col: 12, offset: 5512},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 22, offset: 5522},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 169, col: 24, offset: 5524},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 169, col: 29, offset: 5529},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 37, offset: 5537},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 169, col: 39, offset: 5539},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 169, col: 51, offset: 5551},
								expr: &ruleRefExpr{
									pos:  position{line: 169, col: 51, offset: 5551},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 68, offset: 5568},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 181, col: 1, offset: 5845},
			expr: &actionExpr{
				pos: position{line: 181, col: 14, offset: 5858},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 181, col: 14, offset: 5858},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 181, col: 14, offset: 5858},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 26, offset: 5870},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 28, offset: 5872},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 181, col: 34, offset: 5878},
								expr: &charClassMatcher{
									pos:        position{line: 181, col: 34, offset: 5878},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 44, offset: 5888},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 46, offset: 5890},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 181, col: 49, offset: 5893},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 60, offset: 5904},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 62, offset: 5906},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 181, col: 74, offset: 5918},
								expr: &ruleRefExpr{
									pos:  position{line: 181, col: 74, offset: 5918},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 91, offset: 5935},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 189, col: 1, offset: 6121},
			expr: &actionExpr{
				pos: position{line: 189, col: 10, offset: 6130},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 189, col: 10, offset: 6130},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 189, col: 10, offset: 6130},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 18, offset: 6138},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 20, offset: 6140},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 24, offset: 6144},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 34, offset: 6154},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 36, offset: 6156},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 41, offset: 6161},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 52, offset: 6172},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 189, col: 54, offset: 6174},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 58, offset: 6178},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 60, offset: 6180},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 66, offset: 6186},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 77, offset: 6197},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 79, offset: 6199},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 189, col: 91, offset: 6211},
								expr: &ruleRefExpr{
									pos:  position{line: 189, col: 91, offset: 6211},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 108, offset: 6228},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 198, col: 1, offset: 6422},
			expr: &actionExpr{
				pos: position{line: 198, col: 9, offset: 6430},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 198, col: 9, offset: 6430},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 198, col: 9, offset: 6430},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 16, offset: 6437},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 18, offset: 6439},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 198, col: 23, offset: 6444},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 34, offset: 6455},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 198, col: 37, offset: 6458},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 41, offset: 6462},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 44, offset: 6465},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 198, col: 51, offset: 6472},
								expr: &seqExpr{
									pos: position{line: 198, col: 52, offset: 6473},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 198, col: 52, offset: 6473},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 198, col: 62, offset: 6483},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 198, col: 67, offset: 6488},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 71, offset: 6492},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 73, offset: 6494},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 198, col: 85, offset: 6506},
								expr: &ruleRefExpr{
									pos:  position{line: 198, col: 85, offset: 6506},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 102, offset: 6523},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 222, col: 1, offset: 7185},
			expr: &actionExpr{
				pos: position{line: 222, col: 14, offset: 7198},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 222, col: 14, offset: 7198},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 222, col: 14, offset: 7198},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 21, offset: 7205},
								expr: &seqExpr{
									pos: position{line: 222, col: 22, offset: 7206},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 222, col: 22, offset: 7206},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 32, offset: 7216},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 222, col: 37, offset: 7221},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 42, offset: 7226},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 53, offset: 7237},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 55, offset: 7239},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 61, offset: 7245},
								expr: &seqExpr{
									pos: position{line: 222, col: 62, offset: 7246},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 222, col: 62, offset: 7246},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 66, offset: 7250},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 68, offset: 7252},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 82, offset: 7266},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 84, offset: 7268},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 96, offset: 7280},
								expr: &ruleRefExpr{
									pos:  position{line: 222, col: 96, offset: 7280},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 222, col: 113, offset: 7297},
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 113, offset: 7297},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 238, col: 1, offset: 7695},
			expr: &actionExpr{
				pos: position{line: 238, col: 12, offset: 7706},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 238, col: 12, offset: 7706},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 238, col: 12, offset: 7706},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 22, offset: 7716},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 24, offset: 7718},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 28, offset: 7722},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 38, offset: 7732},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7734},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 45, offset: 7739},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 56, offset: 7750},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 58, offset: 7752},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 238, col: 70, offset: 7764},
								expr: &ruleRefExpr{
									pos:  position{line: 238, col: 70, offset: 7764},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 87, offset: 7781},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 246, col: 1, offset: 7953},
			expr: &actionExpr{
				pos: position{line: 246, col: 11, offset: 7963},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 246, col: 11, offset: 7963},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 246, col: 11, offset: 7963},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 246, col: 20, offset: 7972},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 246, col: 22, offset: 7974},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 246, col: 25, offset: 7977},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 247, col: 1, offset: 8017},
			expr: &actionExpr{
				pos: position{line: 247, col: 14, offset: 8030},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 247, col: 14, offset: 8030},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 247, col: 14, offset: 8030},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 247, col: 26, offset: 8042},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 247, col: 28, offset: 8044},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 247, col: 31, offset: 8047},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 248, col: 1, offset: 8098},
			expr: &actionExpr{
				pos: position{line: 248, col: 10, offset: 8107},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 248, col: 10, offset: 8107},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 248, col: 10, offset: 8107},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 18, offset: 8115},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 20, offset: 8117},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 23, offset: 8120},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 249, col: 1, offset: 8167},
			expr: &actionExpr{
				pos: position{line: 249, col: 15, offset: 8181},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 249, col: 15, offset: 8181},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 249, col: 15, offset: 8181},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 20, offset: 8186},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 31, offset: 8197},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 249, col: 34, offset: 8200},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 38, offset: 8204},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 41, offset: 8207},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 48, offset: 8214},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 249, col: 58, offset: 8224},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 62, offset: 8228},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 64, offset: 8230},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 249, col: 76, offset: 8242},
								expr: &ruleRefExpr{
									pos:  position{line: 249, col: 76, offset: 8242},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 93, offset: 8259},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 260, col: 1, offset: 8476},
			expr: &actionExpr{
				pos: position{line: 260, col: 14, offset: 8489},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 260, col: 14, offset: 8489},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 260, col: 21, offset: 8496},
						expr: &seqExpr{
							pos: position{line: 260, col: 22, offset: 8497},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 260, col: 22, offset: 8497},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 260, col: 28, offset: 8503},
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
			pos:  position{line: 269, col: 1, offset: 8684},
			expr: &actionExpr{
				pos: position{line: 269, col: 10, offset: 8693},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 269, col: 10, offset: 8693},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 269, col: 10, offset: 8693},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 17, offset: 8700},
								expr: &seqExpr{
									pos: position{line: 269, col: 18, offset: 8701},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 269, col: 18, offset: 8701},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 28, offset: 8711},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 269, col: 33, offset: 8716},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 36, offset: 8719},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 48, offset: 8731},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 269, col: 50, offset: 8733},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 54, offset: 8737},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 56, offset: 8739},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 60, offset: 8743},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 60, offset: 8743},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 75, offset: 8758},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 77, offset: 8760},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 81, offset: 8764},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 91, offset: 8774},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 93, offset: 8776},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 98, offset: 8781},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 109, offset: 8792},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 112, offset: 8795},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 116, offset: 8799},
								expr: &seqExpr{
									pos: position{line: 269, col: 117, offset: 8800},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 269, col: 117, offset: 8800},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 121, offset: 8804},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 123, offset: 8806},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 136, offset: 8819},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 138, offset: 8821},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 150, offset: 8833},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 150, offset: 8833},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 269, col: 167, offset: 8850},
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 167, offset: 8850},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 292, col: 1, offset: 9382},
			expr: &actionExpr{
				pos: position{line: 292, col: 18, offset: 9399},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 292, col: 19, offset: 9400},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 292, col: 19, offset: 9400},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 292, col: 32, offset: 9413},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 300, col: 1, offset: 9556},
			expr: &actionExpr{
				pos: position{line: 300, col: 12, offset: 9567},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 300, col: 12, offset: 9567},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 300, col: 12, offset: 9567},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 22, offset: 9577},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 24, offset: 9579},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 29, offset: 9584},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 40, offset: 9595},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 42, offset: 9597},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 50, offset: 9605},
								expr: &seqExpr{
									pos: position{line: 300, col: 51, offset: 9606},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 300, col: 51, offset: 9606},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 61, offset: 9616},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 64, offset: 9619},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 75, offset: 9630},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 80, offset: 9635},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 300, col: 83, offset: 9638},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 87, offset: 9642},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 90, offset: 9645},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 300, col: 98, offset: 9653},
								expr: &seqExpr{
									pos: position{line: 300, col: 99, offset: 9654},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 300, col: 99, offset: 9654},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 108, offset: 9663},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 300, col: 114, offset: 9669},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 300, col: 114, offset: 9669},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 300, col: 120, offset: 9675},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 139, offset: 9694},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 141, offset: 9696},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 153, offset: 9708},
								expr: &ruleRefExpr{
									pos:  position{line: 300, col: 153, offset: 9708},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 170, offset: 9725},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 317, col: 1, offset: 10166},
			expr: &actionExpr{
				pos: position{line: 317, col: 22, offset: 10187},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 317, col: 22, offset: 10187,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 321, col: 1, offset: 10256},
			expr: &actionExpr{
				pos: position{line: 321, col: 13, offset: 10268},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 321, col: 13, offset: 10268},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 321, col: 13, offset: 10268},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 20, offset: 10275},
								expr: &seqExpr{
									pos: position{line: 321, col: 21, offset: 10276},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 321, col: 21, offset: 10276},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 31, offset: 10286},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 36, offset: 10291},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 43, offset: 10298},
								expr: &seqExpr{
									pos: position{line: 321, col: 44, offset: 10299},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 321, col: 44, offset: 10299},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 53, offset: 10308},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 58, offset: 10313},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 62, offset: 10317},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 75, offset: 10330},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 78, offset: 10333},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 83, offset: 10338},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 94, offset: 10349},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 321, col: 96, offset: 10351},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 100, offset: 10355},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 103, offset: 10358},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 113, offset: 10368},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 321, col: 123, offset: 10378},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 127, offset: 10382},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 130, offset: 10385},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 141, offset: 10396},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 141, offset: 10396},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 149, offset: 10404},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 151, offset: 10406},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 163, offset: 10418},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 163, offset: 10418},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 321, col: 180, offset: 10435},
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 180, offset: 10435},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 349, col: 1, offset: 11086},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 11102},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 11102},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 11107},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 349, col: 22, offset: 11107},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 31, offset: 11116},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 356, col: 1, offset: 11238},
			expr: &actionExpr{
				pos: position{line: 356, col: 11, offset: 11248},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 356, col: 11, offset: 11248},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 356, col: 11, offset: 11248},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 20, offset: 11257},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 356, col: 23, offset: 11260},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 27, offset: 11264},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 30, offset: 11267},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 41, offset: 11278},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 356, col: 51, offset: 11288},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 360, col: 1, offset: 11324},
			expr: &actionExpr{
				pos: position{line: 360, col: 14, offset: 11337},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 360, col: 14, offset: 11337},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 360, col: 19, offset: 11342},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 360, col: 19, offset: 11342},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 30, offset: 11353},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 46, offset: 11369},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 367, col: 1, offset: 11494},
			expr: &actionExpr{
				pos: position{line: 367, col: 13, offset: 11506},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 367, col: 13, offset: 11506},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 367, col: 13, offset: 11506},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 367, col: 18, offset: 11511},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 367, col: 31, offset: 11524},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 367, col: 33, offset: 11526},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 367, col: 45, offset: 11538},
								expr: &ruleRefExpr{
									pos:  position{line: 367, col: 45, offset: 11538},
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
			pos:  position{line: 374, col: 1, offset: 11674},
			expr: &actionExpr{
				pos: position{line: 374, col: 17, offset: 11690},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 374, col: 18, offset: 11691},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 374, col: 18, offset: 11691},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 27, offset: 11700},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 36, offset: 11709},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 44, offset: 11717},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 52, offset: 11725},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 60, offset: 11733},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 71, offset: 11744},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 82, offset: 11755},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 378, col: 1, offset: 11802},
			expr: &actionExpr{
				pos: position{line: 378, col: 18, offset: 11819},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 378, col: 18, offset: 11819},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 378, col: 23, offset: 11824},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 378, col: 23, offset: 11824},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 33, offset: 11834},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 43, offset: 11844},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 382, col: 1, offset: 11879},
			expr: &actionExpr{
				pos: position{line: 382, col: 12, offset: 11890},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 12, offset: 11890},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 382, col: 12, offset: 11890},
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 12, offset: 11890},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 382, col: 21, offset: 11899},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 28, offset: 11906},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 31, offset: 11909},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 35, offset: 11913},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 45, offset: 11923},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 48, offset: 11926},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 52, offset: 11930},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 55, offset: 11933},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 61, offset: 11939},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 71, offset: 11949},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 74, offset: 11952},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 78, offset: 11956},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 80, offset: 11958},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 382, col: 92, offset: 11970},
								expr: &ruleRefExpr{
									pos:  position{line: 382, col: 92, offset: 11970},
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
			pos:  position{line: 391, col: 1, offset: 12168},
			expr: &actionExpr{
				pos: position{line: 391, col: 12, offset: 12179},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 391, col: 12, offset: 12179},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 391, col: 12, offset: 12179},
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 12, offset: 12179},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 391, col: 21, offset: 12188},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 28, offset: 12195},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 31, offset: 12198},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 35, offset: 12202},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 45, offset: 12212},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 391, col: 48, offset: 12215},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 52, offset: 12219},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 54, offset: 12221},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 391, col: 66, offset: 12233},
								expr: &ruleRefExpr{
									pos:  position{line: 391, col: 66, offset: 12233},
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
			pos:  position{line: 399, col: 1, offset: 12395},
			expr: &actionExpr{
				pos: position{line: 399, col: 13, offset: 12407},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 399, col: 13, offset: 12407},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 399, col: 13, offset: 12407},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 21, offset: 12415},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 24, offset: 12418},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 28, offset: 12422},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 38, offset: 12432},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 399, col: 41, offset: 12435},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 45, offset: 12439},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 47, offset: 12441},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 399, col: 59, offset: 12453},
								expr: &ruleRefExpr{
									pos:  position{line: 399, col: 59, offset: 12453},
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
			pos:  position{line: 407, col: 1, offset: 12616},
			expr: &actionExpr{
				pos: position{line: 407, col: 12, offset: 12627},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 407, col: 12, offset: 12627},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 407, col: 12, offset: 12627},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 407, col: 23, offset: 12638},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 407, col: 31, offset: 12646},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 411, col: 1, offset: 12683},
			expr: &choiceExpr{
				pos: position{line: 411, col: 15, offset: 12697},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 411, col: 15, offset: 12697},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 25, offset: 12707},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 40, offset: 12722},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 57, offset: 12739},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 71, offset: 12753},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 82, offset: 12764},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 94, offset: 12776},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 413, col: 1, offset: 12788},
			expr: &actionExpr{
				pos: position{line: 413, col: 20, offset: 12807},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 413, col: 20, offset: 12807},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 413, col: 20, offset: 12807},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 413, col: 24, offset: 12811},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 413, col: 27, offset: 12814},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 413, col: 39, offset: 12826},
								expr: &ruleRefExpr{
									pos:  position{line: 413, col: 39, offset: 12826},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 413, col: 55, offset: 12842},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 421, col: 1, offset: 13006},
			expr: &actionExpr{
				pos: position{line: 421, col: 19, offset: 13024},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 421, col: 19, offset: 13024},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 421, col: 19, offset: 13024},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 24, offset: 13029},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 35, offset: 13040},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 421, col: 37, offset: 13042},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 421, col: 43, offset: 13048},
								expr: &actionExpr{
									pos: position{line: 421, col: 44, offset: 13049},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 421, col: 44, offset: 13049},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 421, col: 44, offset: 13049},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 421, col: 48, offset: 13053},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 421, col: 51, offset: 13056},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 421, col: 57, offset: 13062},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 421, col: 89, offset: 13094},
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 89, offset: 13094},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 104, offset: 13109},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 432, col: 1, offset: 13305},
			expr: &actionExpr{
				pos: position{line: 432, col: 17, offset: 13321},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 432, col: 18, offset: 13322},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 432, col: 18, offset: 13322},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 432, col: 27, offset: 13331},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 436, col: 1, offset: 13386},
			expr: &actionExpr{
				pos: position{line: 436, col: 16, offset: 13401},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 436, col: 16, offset: 13401},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 436, col: 16, offset: 13401},
							expr: &charClassMatcher{
								pos:        position{line: 436, col: 16, offset: 13401},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 436, col: 22, offset: 13407},
							expr: &ruleRefExpr{
								pos:  position{line: 436, col: 22, offset: 13407},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 440, col: 1, offset: 13471},
			expr: &actionExpr{
				pos: position{line: 440, col: 19, offset: 13489},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 440, col: 19, offset: 13489},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 440, col: 19, offset: 13489},
							expr: &charClassMatcher{
								pos:        position{line: 440, col: 19, offset: 13489},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 25, offset: 13495},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 25, offset: 13495},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 440, col: 32, offset: 13502},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 36, offset: 13506},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 36, offset: 13506},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 440, col: 43, offset: 13513},
							expr: &seqExpr{
								pos: position{line: 440, col: 45, offset: 13515},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 440, col: 45, offset: 13515},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 440, col: 52, offset: 13522},
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
			pos:  position{line: 444, col: 1, offset: 13592},
			expr: &actionExpr{
				pos: position{line: 444, col: 14, offset: 13605},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 444, col: 14, offset: 13605},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 444, col: 14, offset: 13605},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 18, offset: 13609},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 444, col: 21, offset: 13612},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 444, col: 28, offset: 13619},
								expr: &seqExpr{
									pos: position{line: 444, col: 29, offset: 13620},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 444, col: 29, offset: 13620},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 40, offset: 13631},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 444, col: 43, offset: 13634},
											expr: &ruleRefExpr{
												pos:  position{line: 444, col: 43, offset: 13634},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 58, offset: 13649},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 63, offset: 13654},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 444, col: 66, offset: 13657},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 453, col: 1, offset: 13851},
			expr: &actionExpr{
				pos: position{line: 453, col: 13, offset: 13863},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 453, col: 13, offset: 13863},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 453, col: 13, offset: 13863},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 453, col: 17, offset: 13867},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 453, col: 20, offset: 13870},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 453, col: 27, offset: 13877},
								expr: &seqExpr{
									pos: position{line: 453, col: 28, offset: 13878},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 453, col: 28, offset: 13878},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 39, offset: 13889},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 453, col: 42, offset: 13892},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 46, offset: 13896},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 49, offset: 13899},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 60, offset: 13910},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 453, col: 64, offset: 13914},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 453, col: 64, offset: 13914},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 453, col: 70, offset: 13920},
													expr: &litMatcher{
														pos:        position{line: 453, col: 71, offset: 13921},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 76, offset: 13926},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 453, col: 81, offset: 13931},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 473, col: 1, offset: 14481},
			expr: &actionExpr{
				pos: position{line: 473, col: 10, offset: 14490},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 473, col: 10, offset: 14490},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 473, col: 10, offset: 14490},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 17, offset: 14497},
								expr: &seqExpr{
									pos: position{line: 473, col: 18, offset: 14498},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 473, col: 18, offset: 14498},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 28, offset: 14508},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 473, col: 33, offset: 14513},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 41, offset: 14521},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 44, offset: 14524},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 473, col: 49, offset: 14529},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 60, offset: 14540},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 63, offset: 14543},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 70, offset: 14550},
								expr: &ruleRefExpr{
									pos:  position{line: 473, col: 70, offset: 14550},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 78, offset: 14558},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 473, col: 81, offset: 14561},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 85, offset: 14565},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 88, offset: 14568},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 473, col: 99, offset: 14579},
								expr: &seqExpr{
									pos: position{line: 473, col: 100, offset: 14580},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 473, col: 100, offset: 14580},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 110, offset: 14590},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 473, col: 116, offset: 14596},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 473, col: 116, offset: 14596},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 473, col: 122, offset: 14602},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 139, offset: 14619},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 141, offset: 14621},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 153, offset: 14633},
								expr: &ruleRefExpr{
									pos:  position{line: 473, col: 153, offset: 14633},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 170, offset: 14650},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 495, col: 1, offset: 15247},
			expr: &actionExpr{
				pos: position{line: 495, col: 20, offset: 15266},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 495, col: 20, offset: 15266,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 499, col: 1, offset: 15333},
			expr: &actionExpr{
				pos: position{line: 499, col: 11, offset: 15343},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 499, col: 11, offset: 15343},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 499, col: 11, offset: 15343},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 20, offset: 15352},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 23, offset: 15355},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 499, col: 35, offset: 15367},
							expr: &seqExpr{
								pos: position{line: 499, col: 36, offset: 15368},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 499, col: 36, offset: 15368},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 499, col: 40, offset: 15372},
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
			pos:  position{line: 504, col: 1, offset: 15503},
			expr: &choiceExpr{
				pos: position{line: 504, col: 16, offset: 15518},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 504, col: 17, offset: 15519},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 504, col: 17, offset: 15519},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 504, col: 21, offset: 15523},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 504, col: 32, offset: 15534},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 504, col: 39, offset: 15541},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 506, col: 1, offset: 15553},
			expr: &oneOrMoreExpr{
				pos: position{line: 506, col: 15, offset: 15567},
				expr: &charClassMatcher{
					pos:        position{line: 506, col: 15, offset: 15567},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 508, col: 1, offset: 15585},
			expr: &actionExpr{
				pos: position{line: 508, col: 14, offset: 15598},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 508, col: 14, offset: 15598},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 508, col: 14, offset: 15598},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 508, col: 21, offset: 15605},
								expr: &seqExpr{
									pos: position{line: 508, col: 22, offset: 15606},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 508, col: 22, offset: 15606},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 508, col: 32, offset: 15616},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 508, col: 37, offset: 15621},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 42, offset: 15626},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 53, offset: 15637},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 508, col: 55, offset: 15639},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 59, offset: 15643},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 508, col: 62, offset: 15646},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 66, offset: 15650},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 77, offset: 15661},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 508, col: 79, offset: 15663},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 508, col: 91, offset: 15675},
								expr: &ruleRefExpr{
									pos:  position{line: 508, col: 91, offset: 15675},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 508, col: 108, offset: 15692},
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 108, offset: 15692},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 525, col: 1, offset: 16278},
			expr: &actionExpr{
				pos: position{line: 525, col: 12, offset: 16289},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 525, col: 13, offset: 16290},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 525, col: 14, offset: 16291},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 525, col: 14, offset: 16291},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 525, col: 18, offset: 16295},
									expr: &choiceExpr{
										pos: position{line: 525, col: 19, offset: 16296},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 525, col: 19, offset: 16296},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 525, col: 26, offset: 16303},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 525, col: 33, offset: 16310},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 525, col: 41, offset: 16318},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 525, col: 41, offset: 16318},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 525, col: 46, offset: 16323},
									expr: &choiceExpr{
										pos: position{line: 525, col: 47, offset: 16324},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 525, col: 47, offset: 16324},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 525, col: 54, offset: 16331},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 525, col: 61, offset: 16338},
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
			pos:  position{line: 534, col: 1, offset: 16624},
			expr: &actionExpr{
				pos: position{line: 534, col: 15, offset: 16638},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 534, col: 15, offset: 16638},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 534, col: 15, offset: 16638},
							expr: &choiceExpr{
								pos: position{line: 534, col: 16, offset: 16639},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 534, col: 16, offset: 16639},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 534, col: 25, offset: 16648},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 534, col: 31, offset: 16654},
							expr: &choiceExpr{
								pos: position{line: 534, col: 32, offset: 16655},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 534, col: 32, offset: 16655},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 534, col: 41, offset: 16664},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 534, col: 49, offset: 16672},
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
			pos:  position{line: 538, col: 1, offset: 16727},
			expr: &charClassMatcher{
				pos:        position{line: 538, col: 18, offset: 16744},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 539, col: 1, offset: 16749},
			expr: &charClassMatcher{
				pos:        position{line: 539, col: 11, offset: 16759},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 540, col: 1, offset: 16768},
			expr: &charClassMatcher{
				pos:        position{line: 540, col: 10, offset: 16777},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 542, col: 1, offset: 16784},
			expr: &anyMatcher{
				line: 542, col: 15, offset: 16798,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 543, col: 1, offset: 16800},
			expr: &actionExpr{
				pos: position{line: 543, col: 14, offset: 16813},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 543, col: 14, offset: 16813},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 543, col: 14, offset: 16813},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 543, col: 21, offset: 16820},
							expr: &seqExpr{
								pos: position{line: 543, col: 23, offset: 16822},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 543, col: 23, offset: 16822},
										expr: &litMatcher{
											pos:        position{line: 543, col: 24, offset: 16823},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 543, col: 29, offset: 16828},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 543, col: 43, offset: 16842},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 549, col: 1, offset: 17022},
			expr: &choiceExpr{
				pos: position{line: 549, col: 12, offset: 17033},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 549, col: 12, offset: 17033},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 549, col: 31, offset: 17052},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 550, col: 1, offset: 17070},
			expr: &seqExpr{
				pos: position{line: 550, col: 21, offset: 17090},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 550, col: 21, offset: 17090},
						expr: &ruleRefExpr{
							pos:  position{line: 550, col: 22, offset: 17091},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 550, col: 32, offset: 17101},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 550, col: 37, offset: 17106},
						expr: &seqExpr{
							pos: position{line: 550, col: 39, offset: 17108},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 550, col: 39, offset: 17108},
									expr: &litMatcher{
										pos:        position{line: 550, col: 40, offset: 17109},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 550, col: 45, offset: 17114},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 550, col: 59, offset: 17128},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 551, col: 1, offset: 17133},
			expr: &seqExpr{
				pos: position{line: 551, col: 37, offset: 17169},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 551, col: 37, offset: 17169},
						expr: &ruleRefExpr{
							pos:  position{line: 551, col: 38, offset: 17170},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 551, col: 48, offset: 17180},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 551, col: 53, offset: 17185},
						expr: &seqExpr{
							pos: position{line: 551, col: 55, offset: 17187},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 551, col: 55, offset: 17187},
									expr: &choiceExpr{
										pos: position{line: 551, col: 58, offset: 17190},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 551, col: 58, offset: 17190},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 551, col: 65, offset: 17197},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 551, col: 71, offset: 17203},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 551, col: 85, offset: 17217},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 552, col: 1, offset: 17222},
			expr: &choiceExpr{
				pos: position{line: 552, col: 22, offset: 17243},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 552, col: 23, offset: 17244},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 552, col: 23, offset: 17244},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 552, col: 28, offset: 17249},
								expr: &seqExpr{
									pos: position{line: 552, col: 30, offset: 17251},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 552, col: 30, offset: 17251},
											expr: &ruleRefExpr{
												pos:  position{line: 552, col: 31, offset: 17252},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 552, col: 35, offset: 17256},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 552, col: 53, offset: 17274},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 552, col: 53, offset: 17274},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 552, col: 57, offset: 17278},
								expr: &seqExpr{
									pos: position{line: 552, col: 59, offset: 17280},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 552, col: 59, offset: 17280},
											expr: &ruleRefExpr{
												pos:  position{line: 552, col: 60, offset: 17281},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 552, col: 64, offset: 17285},
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
			pos:  position{line: 554, col: 1, offset: 17301},
			expr: &zeroOrMoreExpr{
				pos: position{line: 554, col: 7, offset: 17307},
				expr: &choiceExpr{
					pos: position{line: 554, col: 9, offset: 17309},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 554, col: 9, offset: 17309},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 554, col: 22, offset: 17322},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 554, col: 28, offset: 17328},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 555, col: 1, offset: 17339},
			expr: &zeroOrMoreExpr{
				pos: position{line: 555, col: 6, offset: 17344},
				expr: &choiceExpr{
					pos: position{line: 555, col: 8, offset: 17346},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 555, col: 8, offset: 17346},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 555, col: 21, offset: 17359},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 556, col: 1, offset: 17395},
			expr: &zeroOrMoreExpr{
				pos: position{line: 556, col: 7, offset: 17401},
				expr: &ruleRefExpr{
					pos:  position{line: 556, col: 7, offset: 17401},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 558, col: 1, offset: 17414},
			expr: &charClassMatcher{
				pos:        position{line: 558, col: 15, offset: 17428},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 559, col: 1, offset: 17436},
			expr: &litMatcher{
				pos:        position{line: 559, col: 8, offset: 17443},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 560, col: 1, offset: 17448},
			expr: &choiceExpr{
				pos: position{line: 560, col: 8, offset: 17455},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 560, col: 8, offset: 17455},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 8, offset: 17455},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 560, col: 11, offset: 17458},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 560, col: 17, offset: 17464},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 17, offset: 17464},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 560, col: 19, offset: 17466},
								expr: &ruleRefExpr{
									pos:  position{line: 560, col: 19, offset: 17466},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 560, col: 38, offset: 17485},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 560, col: 44, offset: 17491},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 44, offset: 17491},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 560, col: 47, offset: 17494},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 562, col: 1, offset: 17499},
			expr: &notExpr{
				pos: position{line: 562, col: 8, offset: 17506},
				expr: &anyMatcher{
					line: 562, col: 9, offset: 17507,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	stmts := toIfaceSlice(statements)
	frugal := &Frugal{
		Scopes:         []*Scope{},
		ParsedIncludes: make(map[string]*Frugal),
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

	for _, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
		switch v := wrapper.statement.(type) {
		case *Namespace:
			frugal.Namespaces = append(frugal.Namespaces, v)
			frugal.namespaceIndex[v.Scope] = v
		case *Constant:
			v.Comment = wrapper.comment
			frugal.Constants = append(frugal.Constants, v)
		case *Enum:
			v.Comment = wrapper.comment
			frugal.Enums = append(frugal.Enums, v)
		case *TypeDef:
			v.Comment = wrapper.comment
			frugal.Typedefs = append(frugal.Typedefs, v)
			frugal.typedefIndex[v.Name] = v
		case *Struct:
			v.Type = StructTypeStruct
			v.Comment = wrapper.comment
			frugal.Structs = append(frugal.Structs, v)
		case exception:
			strct := (*Struct)(v)
			strct.Type = StructTypeException
			strct.Comment = wrapper.comment
			frugal.Exceptions = append(frugal.Exceptions, strct)
		case union:
			strct := unionToStruct(v)
			strct.Type = StructTypeUnion
			strct.Comment = wrapper.comment
			frugal.Unions = append(frugal.Unions, strct)
		case *Service:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Services = append(frugal.Services, v)
		case *Include:
			frugal.Includes = append(frugal.Includes, v)
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
