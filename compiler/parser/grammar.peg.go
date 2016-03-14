package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 77, col: 1, offset: 2164},
			expr: &actionExpr{
				pos: position{line: 77, col: 11, offset: 2176},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 77, col: 11, offset: 2176},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 77, col: 11, offset: 2176},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 77, col: 14, offset: 2179},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 25, offset: 2190},
								expr: &seqExpr{
									pos: position{line: 77, col: 27, offset: 2192},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 77, col: 27, offset: 2192},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 37, offset: 2202},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 77, col: 44, offset: 2209},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 77, col: 44, offset: 2209},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 77, col: 50, offset: 2215},
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
			pos:  position{line: 148, col: 1, offset: 4823},
			expr: &actionExpr{
				pos: position{line: 148, col: 15, offset: 4839},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 148, col: 15, offset: 4839,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 152, col: 1, offset: 4897},
			expr: &actionExpr{
				pos: position{line: 152, col: 13, offset: 4911},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 152, col: 13, offset: 4911},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 152, col: 13, offset: 4911},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 152, col: 20, offset: 4918},
								expr: &seqExpr{
									pos: position{line: 152, col: 21, offset: 4919},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 152, col: 21, offset: 4919},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 152, col: 31, offset: 4929},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 152, col: 36, offset: 4934},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 152, col: 47, offset: 4945},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 152, col: 47, offset: 4945},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 152, col: 65, offset: 4963},
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
			pos:  position{line: 165, col: 1, offset: 5434},
			expr: &choiceExpr{
				pos: position{line: 165, col: 19, offset: 5454},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 165, col: 19, offset: 5454},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 29, offset: 5464},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 41, offset: 5476},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 49, offset: 5484},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 56, offset: 5491},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 66, offset: 5501},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 75, offset: 5510},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 87, offset: 5522},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 95, offset: 5530},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 167, col: 1, offset: 5539},
			expr: &actionExpr{
				pos: position{line: 167, col: 11, offset: 5551},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 167, col: 11, offset: 5551},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 167, col: 11, offset: 5551},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 167, col: 21, offset: 5561},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 167, col: 23, offset: 5563},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 167, col: 28, offset: 5568},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 167, col: 36, offset: 5576},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5753},
			expr: &actionExpr{
				pos: position{line: 175, col: 13, offset: 5767},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 13, offset: 5767},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 13, offset: 5767},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 25, offset: 5779},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 27, offset: 5781},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 33, offset: 5787},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 33, offset: 5787},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 43, offset: 5797},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 45, offset: 5799},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 48, offset: 5802},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 59, offset: 5813},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5938},
			expr: &actionExpr{
				pos: position{line: 182, col: 9, offset: 5948},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 9, offset: 5948},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 9, offset: 5948},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 17, offset: 5956},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 19, offset: 5958},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 23, offset: 5962},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 33, offset: 5972},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 35, offset: 5974},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 40, offset: 5979},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 51, offset: 5990},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 53, offset: 5992},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 57, offset: 5996},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 59, offset: 5998},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 65, offset: 6004},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 76, offset: 6015},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6147},
			expr: &actionExpr{
				pos: position{line: 190, col: 8, offset: 6156},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 8, offset: 6156},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 8, offset: 6156},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 15, offset: 6163},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 17, offset: 6165},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 22, offset: 6170},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 33, offset: 6181},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 36, offset: 6184},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 40, offset: 6188},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 43, offset: 6191},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 50, offset: 6198},
								expr: &seqExpr{
									pos: position{line: 190, col: 51, offset: 6199},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 51, offset: 6199},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 61, offset: 6209},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 66, offset: 6214},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 70, offset: 6218},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6830},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6844},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6844},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6844},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6851},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6852},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6852},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6862},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6867},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6872},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6883},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6885},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6891},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6892},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6892},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6896},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6898},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6912},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6912},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7248},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7260},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7260},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7260},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7270},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7272},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7276},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7286},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7288},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7293},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7304},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7413},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7424},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7424},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7424},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7433},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7435},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7438},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7478},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7492},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7492},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7492},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7504},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7506},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7509},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7560},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7570},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7570},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7570},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7578},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7580},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7583},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7630},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7645},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7645},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7645},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7650},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7661},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7664},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7668},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7671},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7678},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7688},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7692},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7853},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7867},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7867},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7874},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7875},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7875},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7881},
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
			pos:  position{line: 257, col: 1, offset: 8062},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 8072},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 8072},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 8072},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 8079},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 8080},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 8080},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 8090},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 8095},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 8098},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8110},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8112},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8116},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8118},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8122},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8122},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 74, offset: 8137},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 76, offset: 8139},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 80, offset: 8143},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 90, offset: 8153},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 92, offset: 8155},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 97, offset: 8160},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 108, offset: 8171},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 111, offset: 8174},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 115, offset: 8178},
								expr: &seqExpr{
									pos: position{line: 257, col: 116, offset: 8179},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 116, offset: 8179},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 120, offset: 8183},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 122, offset: 8185},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 135, offset: 8198},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 135, offset: 8198},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 279, col: 1, offset: 8660},
			expr: &actionExpr{
				pos: position{line: 279, col: 17, offset: 8678},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 279, col: 18, offset: 8679},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 18, offset: 8679},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 279, col: 31, offset: 8692},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 287, col: 1, offset: 8835},
			expr: &actionExpr{
				pos: position{line: 287, col: 11, offset: 8847},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 287, col: 11, offset: 8847},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 287, col: 11, offset: 8847},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 21, offset: 8857},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 23, offset: 8859},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 287, col: 28, offset: 8864},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 39, offset: 8875},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 41, offset: 8877},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 287, col: 49, offset: 8885},
								expr: &seqExpr{
									pos: position{line: 287, col: 50, offset: 8886},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 287, col: 50, offset: 8886},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 60, offset: 8896},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 63, offset: 8899},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 74, offset: 8910},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 79, offset: 8915},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 287, col: 82, offset: 8918},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 86, offset: 8922},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 89, offset: 8925},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 287, col: 97, offset: 8933},
								expr: &seqExpr{
									pos: position{line: 287, col: 98, offset: 8934},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 287, col: 98, offset: 8934},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 107, offset: 8943},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 287, col: 113, offset: 8949},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 287, col: 113, offset: 8949},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 287, col: 119, offset: 8955},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 138, offset: 8974},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 303, col: 1, offset: 9358},
			expr: &actionExpr{
				pos: position{line: 303, col: 21, offset: 9380},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 303, col: 21, offset: 9380,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 307, col: 1, offset: 9449},
			expr: &actionExpr{
				pos: position{line: 307, col: 12, offset: 9462},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 307, col: 12, offset: 9462},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 307, col: 12, offset: 9462},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 19, offset: 9469},
								expr: &seqExpr{
									pos: position{line: 307, col: 20, offset: 9470},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 20, offset: 9470},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 30, offset: 9480},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 35, offset: 9485},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 42, offset: 9492},
								expr: &seqExpr{
									pos: position{line: 307, col: 43, offset: 9493},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 43, offset: 9493},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 52, offset: 9502},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 57, offset: 9507},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 61, offset: 9511},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 74, offset: 9524},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 77, offset: 9527},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 82, offset: 9532},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 93, offset: 9543},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 307, col: 95, offset: 9545},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 99, offset: 9549},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 102, offset: 9552},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 112, offset: 9562},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 307, col: 122, offset: 9572},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 126, offset: 9576},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 129, offset: 9579},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 140, offset: 9590},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 140, offset: 9590},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 307, col: 148, offset: 9598},
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 148, offset: 9598},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 334, col: 1, offset: 10193},
			expr: &actionExpr{
				pos: position{line: 334, col: 16, offset: 10210},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 334, col: 16, offset: 10210},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 334, col: 21, offset: 10215},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 334, col: 21, offset: 10215},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 30, offset: 10224},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 341, col: 1, offset: 10346},
			expr: &actionExpr{
				pos: position{line: 341, col: 10, offset: 10357},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 341, col: 10, offset: 10357},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 10, offset: 10357},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 19, offset: 10366},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 341, col: 22, offset: 10369},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 26, offset: 10373},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 341, col: 29, offset: 10376},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 40, offset: 10387},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 341, col: 50, offset: 10397},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 345, col: 1, offset: 10433},
			expr: &actionExpr{
				pos: position{line: 345, col: 13, offset: 10447},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 13, offset: 10447},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 345, col: 18, offset: 10452},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 345, col: 18, offset: 10452},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 29, offset: 10463},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 45, offset: 10479},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 352, col: 1, offset: 10604},
			expr: &actionExpr{
				pos: position{line: 352, col: 12, offset: 10617},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 352, col: 13, offset: 10618},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 352, col: 13, offset: 10618},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 22, offset: 10627},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 31, offset: 10636},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 39, offset: 10644},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 47, offset: 10652},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 55, offset: 10660},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 66, offset: 10671},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 77, offset: 10682},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 356, col: 1, offset: 10742},
			expr: &actionExpr{
				pos: position{line: 356, col: 17, offset: 10760},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 17, offset: 10760},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 22, offset: 10765},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 356, col: 22, offset: 10765},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 32, offset: 10775},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 42, offset: 10785},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 360, col: 1, offset: 10820},
			expr: &actionExpr{
				pos: position{line: 360, col: 11, offset: 10832},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 360, col: 11, offset: 10832},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 360, col: 11, offset: 10832},
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 11, offset: 10832},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 360, col: 20, offset: 10841},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 27, offset: 10848},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 30, offset: 10851},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 34, offset: 10855},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 44, offset: 10865},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 47, offset: 10868},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 51, offset: 10872},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 54, offset: 10875},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 60, offset: 10881},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 70, offset: 10891},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 73, offset: 10894},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 368, col: 1, offset: 11017},
			expr: &actionExpr{
				pos: position{line: 368, col: 11, offset: 11029},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 11, offset: 11029},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 368, col: 11, offset: 11029},
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 11, offset: 11029},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 368, col: 20, offset: 11038},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 27, offset: 11045},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 30, offset: 11048},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 34, offset: 11052},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 44, offset: 11062},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 47, offset: 11065},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 375, col: 1, offset: 11156},
			expr: &actionExpr{
				pos: position{line: 375, col: 12, offset: 11169},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 12, offset: 11169},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 12, offset: 11169},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 20, offset: 11177},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 375, col: 23, offset: 11180},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 27, offset: 11184},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 37, offset: 11194},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 375, col: 40, offset: 11197},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 382, col: 1, offset: 11289},
			expr: &actionExpr{
				pos: position{line: 382, col: 11, offset: 11301},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 11, offset: 11301},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 382, col: 11, offset: 11301},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 382, col: 22, offset: 11312},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 30, offset: 11320},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 386, col: 1, offset: 11357},
			expr: &choiceExpr{
				pos: position{line: 386, col: 14, offset: 11372},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 386, col: 14, offset: 11372},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 24, offset: 11382},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 41, offset: 11399},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 55, offset: 11413},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 66, offset: 11424},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 78, offset: 11436},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 388, col: 1, offset: 11448},
			expr: &actionExpr{
				pos: position{line: 388, col: 15, offset: 11464},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 388, col: 15, offset: 11464},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 388, col: 15, offset: 11464},
							expr: &charClassMatcher{
								pos:        position{line: 388, col: 15, offset: 11464},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 388, col: 21, offset: 11470},
							expr: &ruleRefExpr{
								pos:  position{line: 388, col: 21, offset: 11470},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 392, col: 1, offset: 11534},
			expr: &actionExpr{
				pos: position{line: 392, col: 18, offset: 11553},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 392, col: 18, offset: 11553},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 18, offset: 11553},
							expr: &charClassMatcher{
								pos:        position{line: 392, col: 18, offset: 11553},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 24, offset: 11559},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 24, offset: 11559},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 31, offset: 11566},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 35, offset: 11570},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 35, offset: 11570},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 392, col: 42, offset: 11577},
							expr: &seqExpr{
								pos: position{line: 392, col: 44, offset: 11579},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 392, col: 44, offset: 11579},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 392, col: 51, offset: 11586},
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
			pos:  position{line: 396, col: 1, offset: 11656},
			expr: &actionExpr{
				pos: position{line: 396, col: 13, offset: 11670},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 396, col: 13, offset: 11670},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 396, col: 13, offset: 11670},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 17, offset: 11674},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 396, col: 20, offset: 11677},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 396, col: 27, offset: 11684},
								expr: &seqExpr{
									pos: position{line: 396, col: 28, offset: 11685},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 396, col: 28, offset: 11685},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 39, offset: 11696},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 396, col: 42, offset: 11699},
											expr: &ruleRefExpr{
												pos:  position{line: 396, col: 42, offset: 11699},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 57, offset: 11714},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 62, offset: 11719},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 396, col: 65, offset: 11722},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 405, col: 1, offset: 11916},
			expr: &actionExpr{
				pos: position{line: 405, col: 12, offset: 11929},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 405, col: 12, offset: 11929},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 405, col: 12, offset: 11929},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 405, col: 16, offset: 11933},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 405, col: 19, offset: 11936},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 405, col: 26, offset: 11943},
								expr: &seqExpr{
									pos: position{line: 405, col: 27, offset: 11944},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 405, col: 27, offset: 11944},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 38, offset: 11955},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 405, col: 41, offset: 11958},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 45, offset: 11962},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 48, offset: 11965},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 59, offset: 11976},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 405, col: 63, offset: 11980},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 405, col: 63, offset: 11980},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 405, col: 69, offset: 11986},
													expr: &litMatcher{
														pos:        position{line: 405, col: 70, offset: 11987},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 75, offset: 11992},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 405, col: 80, offset: 11997},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 425, col: 1, offset: 12547},
			expr: &ruleRefExpr{
				pos:  position{line: 425, col: 19, offset: 12567},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 427, col: 1, offset: 12574},
			expr: &actionExpr{
				pos: position{line: 427, col: 9, offset: 12584},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 427, col: 9, offset: 12584},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 427, col: 9, offset: 12584},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 16, offset: 12591},
								expr: &seqExpr{
									pos: position{line: 427, col: 17, offset: 12592},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 17, offset: 12592},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 27, offset: 12602},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 427, col: 32, offset: 12607},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 40, offset: 12615},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 43, offset: 12618},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 427, col: 48, offset: 12623},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 59, offset: 12634},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 62, offset: 12637},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 69, offset: 12644},
								expr: &ruleRefExpr{
									pos:  position{line: 427, col: 69, offset: 12644},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 77, offset: 12652},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 427, col: 80, offset: 12655},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 84, offset: 12659},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 87, offset: 12662},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 427, col: 98, offset: 12673},
								expr: &seqExpr{
									pos: position{line: 427, col: 99, offset: 12674},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 99, offset: 12674},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 109, offset: 12684},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 427, col: 115, offset: 12690},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 427, col: 115, offset: 12690},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 427, col: 121, offset: 12696},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 138, offset: 12713},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 448, col: 1, offset: 13258},
			expr: &actionExpr{
				pos: position{line: 448, col: 19, offset: 13278},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 448, col: 19, offset: 13278,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 452, col: 1, offset: 13345},
			expr: &actionExpr{
				pos: position{line: 452, col: 10, offset: 13356},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 452, col: 10, offset: 13356},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 452, col: 10, offset: 13356},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 19, offset: 13365},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 22, offset: 13368},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 452, col: 34, offset: 13380},
							expr: &seqExpr{
								pos: position{line: 452, col: 35, offset: 13381},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 452, col: 35, offset: 13381},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 452, col: 39, offset: 13385},
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
			pos:  position{line: 457, col: 1, offset: 13516},
			expr: &choiceExpr{
				pos: position{line: 457, col: 15, offset: 13532},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 457, col: 16, offset: 13533},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 457, col: 16, offset: 13533},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 457, col: 20, offset: 13537},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 457, col: 31, offset: 13548},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 457, col: 38, offset: 13555},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 459, col: 1, offset: 13567},
			expr: &oneOrMoreExpr{
				pos: position{line: 459, col: 14, offset: 13582},
				expr: &charClassMatcher{
					pos:        position{line: 459, col: 14, offset: 13582},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 461, col: 1, offset: 13600},
			expr: &actionExpr{
				pos: position{line: 461, col: 13, offset: 13614},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 461, col: 13, offset: 13614},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 461, col: 13, offset: 13614},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 461, col: 20, offset: 13621},
								expr: &seqExpr{
									pos: position{line: 461, col: 21, offset: 13622},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 461, col: 21, offset: 13622},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 31, offset: 13632},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 461, col: 36, offset: 13637},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 41, offset: 13642},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 52, offset: 13653},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 461, col: 54, offset: 13655},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 58, offset: 13659},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 61, offset: 13662},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 65, offset: 13666},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 76, offset: 13677},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 477, col: 1, offset: 14188},
			expr: &actionExpr{
				pos: position{line: 477, col: 11, offset: 14200},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 477, col: 12, offset: 14201},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 477, col: 13, offset: 14202},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 477, col: 13, offset: 14202},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 477, col: 17, offset: 14206},
									expr: &choiceExpr{
										pos: position{line: 477, col: 18, offset: 14207},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 477, col: 18, offset: 14207},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 477, col: 25, offset: 14214},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 477, col: 32, offset: 14221},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 477, col: 40, offset: 14229},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 477, col: 40, offset: 14229},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 477, col: 45, offset: 14234},
									expr: &choiceExpr{
										pos: position{line: 477, col: 46, offset: 14235},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 477, col: 46, offset: 14235},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 477, col: 53, offset: 14242},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 477, col: 60, offset: 14249},
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
			pos:  position{line: 484, col: 1, offset: 14465},
			expr: &actionExpr{
				pos: position{line: 484, col: 14, offset: 14480},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 484, col: 14, offset: 14480},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 484, col: 14, offset: 14480},
							expr: &choiceExpr{
								pos: position{line: 484, col: 15, offset: 14481},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 484, col: 15, offset: 14481},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 484, col: 24, offset: 14490},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 484, col: 30, offset: 14496},
							expr: &choiceExpr{
								pos: position{line: 484, col: 31, offset: 14497},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 484, col: 31, offset: 14497},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 484, col: 40, offset: 14506},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 484, col: 48, offset: 14514},
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
			pos:  position{line: 488, col: 1, offset: 14569},
			expr: &charClassMatcher{
				pos:        position{line: 488, col: 17, offset: 14587},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 489, col: 1, offset: 14592},
			expr: &charClassMatcher{
				pos:        position{line: 489, col: 10, offset: 14603},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 490, col: 1, offset: 14612},
			expr: &charClassMatcher{
				pos:        position{line: 490, col: 9, offset: 14622},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 492, col: 1, offset: 14629},
			expr: &anyMatcher{
				line: 492, col: 14, offset: 14644,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 493, col: 1, offset: 14646},
			expr: &actionExpr{
				pos: position{line: 493, col: 13, offset: 14660},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 493, col: 13, offset: 14660},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 493, col: 13, offset: 14660},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 493, col: 20, offset: 14667},
							expr: &seqExpr{
								pos: position{line: 493, col: 22, offset: 14669},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 493, col: 22, offset: 14669},
										expr: &litMatcher{
											pos:        position{line: 493, col: 23, offset: 14670},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 493, col: 28, offset: 14675},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 493, col: 42, offset: 14689},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 499, col: 1, offset: 14869},
			expr: &choiceExpr{
				pos: position{line: 499, col: 11, offset: 14881},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 499, col: 11, offset: 14881},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 499, col: 30, offset: 14900},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 500, col: 1, offset: 14918},
			expr: &seqExpr{
				pos: position{line: 500, col: 20, offset: 14939},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 500, col: 20, offset: 14939},
						expr: &ruleRefExpr{
							pos:  position{line: 500, col: 21, offset: 14940},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 500, col: 31, offset: 14950},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 500, col: 36, offset: 14955},
						expr: &seqExpr{
							pos: position{line: 500, col: 38, offset: 14957},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 500, col: 38, offset: 14957},
									expr: &litMatcher{
										pos:        position{line: 500, col: 39, offset: 14958},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 500, col: 44, offset: 14963},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 500, col: 58, offset: 14977},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 501, col: 1, offset: 14982},
			expr: &seqExpr{
				pos: position{line: 501, col: 36, offset: 15019},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 501, col: 36, offset: 15019},
						expr: &ruleRefExpr{
							pos:  position{line: 501, col: 37, offset: 15020},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 501, col: 47, offset: 15030},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 501, col: 52, offset: 15035},
						expr: &seqExpr{
							pos: position{line: 501, col: 54, offset: 15037},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 501, col: 54, offset: 15037},
									expr: &choiceExpr{
										pos: position{line: 501, col: 57, offset: 15040},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 501, col: 57, offset: 15040},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 501, col: 64, offset: 15047},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 501, col: 70, offset: 15053},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 501, col: 84, offset: 15067},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 502, col: 1, offset: 15072},
			expr: &choiceExpr{
				pos: position{line: 502, col: 21, offset: 15094},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 502, col: 22, offset: 15095},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 502, col: 22, offset: 15095},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 502, col: 27, offset: 15100},
								expr: &seqExpr{
									pos: position{line: 502, col: 29, offset: 15102},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 502, col: 29, offset: 15102},
											expr: &ruleRefExpr{
												pos:  position{line: 502, col: 30, offset: 15103},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 502, col: 34, offset: 15107},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 502, col: 52, offset: 15125},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 502, col: 52, offset: 15125},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 502, col: 56, offset: 15129},
								expr: &seqExpr{
									pos: position{line: 502, col: 58, offset: 15131},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 502, col: 58, offset: 15131},
											expr: &ruleRefExpr{
												pos:  position{line: 502, col: 59, offset: 15132},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 502, col: 63, offset: 15136},
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
			pos:  position{line: 504, col: 1, offset: 15152},
			expr: &zeroOrMoreExpr{
				pos: position{line: 504, col: 6, offset: 15159},
				expr: &choiceExpr{
					pos: position{line: 504, col: 8, offset: 15161},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 504, col: 8, offset: 15161},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 21, offset: 15174},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 27, offset: 15180},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 505, col: 1, offset: 15191},
			expr: &zeroOrMoreExpr{
				pos: position{line: 505, col: 5, offset: 15197},
				expr: &choiceExpr{
					pos: position{line: 505, col: 7, offset: 15199},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 505, col: 7, offset: 15199},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 505, col: 20, offset: 15212},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 506, col: 1, offset: 15248},
			expr: &zeroOrMoreExpr{
				pos: position{line: 506, col: 6, offset: 15255},
				expr: &ruleRefExpr{
					pos:  position{line: 506, col: 6, offset: 15255},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 508, col: 1, offset: 15268},
			expr: &charClassMatcher{
				pos:        position{line: 508, col: 14, offset: 15283},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 509, col: 1, offset: 15291},
			expr: &litMatcher{
				pos:        position{line: 509, col: 7, offset: 15299},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 510, col: 1, offset: 15304},
			expr: &choiceExpr{
				pos: position{line: 510, col: 7, offset: 15312},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 510, col: 7, offset: 15312},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 510, col: 7, offset: 15312},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 510, col: 10, offset: 15315},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 510, col: 16, offset: 15321},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 510, col: 16, offset: 15321},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 510, col: 18, offset: 15323},
								expr: &ruleRefExpr{
									pos:  position{line: 510, col: 18, offset: 15323},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 510, col: 37, offset: 15342},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 510, col: 43, offset: 15348},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 510, col: 43, offset: 15348},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 510, col: 46, offset: 15351},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 512, col: 1, offset: 15356},
			expr: &notExpr{
				pos: position{line: 512, col: 7, offset: 15364},
				expr: &anyMatcher{
					line: 512, col: 8, offset: 15365,
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

func (c *current) onInclude1(file interface{}) (interface{}, error) {
	name := file.(string)
	if ix := strings.LastIndex(name, "."); ix > 0 {
		name = name[:ix]
	}
	return &Include{Name: name, Value: file.(string)}, nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"])
}

func (c *current) onNamespace1(scope, ns interface{}) (interface{}, error) {
	return &Namespace{
		Scope: ifaceSliceToString(scope),
		Value: string(ns.(Identifier)),
	}, nil
}

func (p *parser) callonNamespace1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNamespace1(stack["scope"], stack["ns"])
}

func (c *current) onConst1(typ, name, value interface{}) (interface{}, error) {
	return &Constant{
		Name:  string(name.(Identifier)),
		Type:  typ.(*Type),
		Value: value,
	}, nil
}

func (p *parser) callonConst1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConst1(stack["typ"], stack["name"], stack["value"])
}

func (c *current) onEnum1(name, values interface{}) (interface{}, error) {
	vs := toIfaceSlice(values)
	en := &Enum{
		Name:   string(name.(Identifier)),
		Values: make(map[string]*EnumValue, len(vs)),
	}
	// Assigns numbers in order. This will behave badly if some values are
	// defined and other are not, but I think that's ok since that's a silly
	// thing to do.
	next := 0
	for _, v := range vs {
		ev := v.([]interface{})[0].(*EnumValue)
		if ev.Value < 0 {
			ev.Value = next
		}
		if ev.Value >= next {
			next = ev.Value + 1
		}
		en.Values[ev.Name] = ev
	}
	return en, nil
}

func (p *parser) callonEnum1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnum1(stack["name"], stack["values"])
}

func (c *current) onEnumValue1(docstr, name, value interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:  string(name.(Identifier)),
		Value: -1,
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
	return p.cur.onEnumValue1(stack["docstr"], stack["name"], stack["value"])
}

func (c *current) onTypeDef1(typ, name interface{}) (interface{}, error) {
	return &TypeDef{
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
	}, nil
}

func (p *parser) callonTypeDef1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeDef1(stack["typ"], stack["name"])
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

func (c *current) onStructLike1(name, fields interface{}) (interface{}, error) {
	st := &Struct{
		Name: string(name.(Identifier)),
	}
	if fields != nil {
		st.Fields = fields.([]*Field)
	}
	return st, nil
}

func (p *parser) callonStructLike1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStructLike1(stack["name"], stack["fields"])
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

func (c *current) onField1(docstr, id, mod, typ, name, def interface{}) (interface{}, error) {
	f := &Field{
		ID:   int(id.(int64)),
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
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
	return p.cur.onField1(stack["docstr"], stack["id"], stack["mod"], stack["typ"], stack["name"], stack["def"])
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

func (c *current) onService1(name, extends, methods interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	svc := &Service{
		Name:    string(name.(Identifier)),
		Methods: make([]*Method, len(ms)),
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
	return p.cur.onService1(stack["name"], stack["extends"], stack["methods"])
}

func (c *current) onEndOfServiceError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of service")
}

func (p *parser) callonEndOfServiceError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfServiceError1()
}

func (c *current) onFunction1(docstr, oneway, typ, name, arguments, exceptions interface{}) (interface{}, error) {
	m := &Method{
		Name: string(name.(Identifier)),
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
	return p.cur.onFunction1(stack["docstr"], stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"])
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

func (c *current) onBaseType1() (interface{}, error) {
	return &Type{Name: string(c.text)}, nil
}

func (p *parser) callonBaseType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseType1()
}

func (c *current) onContainerType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonContainerType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onContainerType1(stack["typ"])
}

func (c *current) onMapType1(key, value interface{}) (interface{}, error) {
	return &Type{
		Name:      "map",
		KeyType:   key.(*Type),
		ValueType: value.(*Type),
	}, nil
}

func (p *parser) callonMapType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMapType1(stack["key"], stack["value"])
}

func (c *current) onSetType1(typ interface{}) (interface{}, error) {
	return &Type{
		Name:      "set",
		ValueType: typ.(*Type),
	}, nil
}

func (p *parser) callonSetType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSetType1(stack["typ"])
}

func (c *current) onListType1(typ interface{}) (interface{}, error) {
	return &Type{
		Name:      "list",
		ValueType: typ.(*Type),
	}, nil
}

func (p *parser) callonListType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListType1(stack["typ"])
}

func (c *current) onCppType1(cppType interface{}) (interface{}, error) {
	return cppType, nil
}

func (p *parser) callonCppType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCppType1(stack["cppType"])
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

func (c *current) onScope1(docstr, name, prefix, operations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:       string(name.(Identifier)),
		Operations: make([]*Operation, len(ops)),
		Prefix:     defaultPrefix,
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
	return p.cur.onScope1(stack["docstr"], stack["name"], stack["prefix"], stack["operations"])
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

func (c *current) onOperation1(docstr, name, typ interface{}) (interface{}, error) {
	o := &Operation{
		Name: string(name.(Identifier)),
		Type: &Type{Name: string(typ.(Identifier))},
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
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["typ"])
}

func (c *current) onLiteral1() (interface{}, error) {
	if len(c.text) != 0 && c.text[0] == '\'' {
		return strconv.Unquote(`"` + strings.Replace(string(c.text[1:len(c.text)-1]), `\'`, `'`, -1) + `"`)
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
	pe := &parserError{Inner: err, prefix: buf.String()}
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
