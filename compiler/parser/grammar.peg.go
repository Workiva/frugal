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
			pos:  position{line: 213, col: 1, offset: 6819},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6833},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6833},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6833},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6840},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6841},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6841},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6851},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6856},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6861},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6872},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6874},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6880},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6881},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6881},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6885},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6887},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6901},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6901},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7237},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7249},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7249},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7249},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7259},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7261},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7265},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7275},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7277},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7282},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7293},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7402},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7413},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7413},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7413},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7422},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7424},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7427},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7467},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7481},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7481},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7481},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7493},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7495},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7498},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7549},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7559},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7559},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7559},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7567},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7569},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7572},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7619},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7634},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7634},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7634},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7639},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7650},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7653},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7657},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7660},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7667},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7677},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7681},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7842},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7856},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7856},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7863},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7864},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7864},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7870},
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
			pos:  position{line: 257, col: 1, offset: 8051},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 8061},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 8061},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 8061},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 8068},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 8069},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 8069},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 8079},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 8084},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 8087},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8099},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8101},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8105},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8107},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8111},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8111},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 74, offset: 8126},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 76, offset: 8128},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 80, offset: 8132},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 90, offset: 8142},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 92, offset: 8144},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 97, offset: 8149},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 108, offset: 8160},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 111, offset: 8163},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 115, offset: 8167},
								expr: &seqExpr{
									pos: position{line: 257, col: 116, offset: 8168},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 116, offset: 8168},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 120, offset: 8172},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 122, offset: 8174},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 135, offset: 8187},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 135, offset: 8187},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 279, col: 1, offset: 8649},
			expr: &actionExpr{
				pos: position{line: 279, col: 17, offset: 8667},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 279, col: 18, offset: 8668},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 18, offset: 8668},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 279, col: 31, offset: 8681},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 287, col: 1, offset: 8824},
			expr: &actionExpr{
				pos: position{line: 287, col: 11, offset: 8836},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 287, col: 11, offset: 8836},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 287, col: 11, offset: 8836},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 21, offset: 8846},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 23, offset: 8848},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 287, col: 28, offset: 8853},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 39, offset: 8864},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 41, offset: 8866},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 287, col: 49, offset: 8874},
								expr: &seqExpr{
									pos: position{line: 287, col: 50, offset: 8875},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 287, col: 50, offset: 8875},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 60, offset: 8885},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 63, offset: 8888},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 74, offset: 8899},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 79, offset: 8904},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 287, col: 82, offset: 8907},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 86, offset: 8911},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 89, offset: 8914},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 287, col: 97, offset: 8922},
								expr: &seqExpr{
									pos: position{line: 287, col: 98, offset: 8923},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 287, col: 98, offset: 8923},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 107, offset: 8932},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 287, col: 113, offset: 8938},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 287, col: 113, offset: 8938},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 287, col: 119, offset: 8944},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 138, offset: 8963},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 303, col: 1, offset: 9347},
			expr: &actionExpr{
				pos: position{line: 303, col: 21, offset: 9369},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 303, col: 21, offset: 9369,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 307, col: 1, offset: 9438},
			expr: &actionExpr{
				pos: position{line: 307, col: 12, offset: 9451},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 307, col: 12, offset: 9451},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 307, col: 12, offset: 9451},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 19, offset: 9458},
								expr: &seqExpr{
									pos: position{line: 307, col: 20, offset: 9459},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 20, offset: 9459},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 30, offset: 9469},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 35, offset: 9474},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 42, offset: 9481},
								expr: &seqExpr{
									pos: position{line: 307, col: 43, offset: 9482},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 43, offset: 9482},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 52, offset: 9491},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 57, offset: 9496},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 61, offset: 9500},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 74, offset: 9513},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 77, offset: 9516},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 82, offset: 9521},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 93, offset: 9532},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 307, col: 95, offset: 9534},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 99, offset: 9538},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 102, offset: 9541},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 112, offset: 9551},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 307, col: 122, offset: 9561},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 126, offset: 9565},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 129, offset: 9568},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 140, offset: 9579},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 140, offset: 9579},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 307, col: 148, offset: 9587},
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 148, offset: 9587},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 334, col: 1, offset: 10182},
			expr: &actionExpr{
				pos: position{line: 334, col: 16, offset: 10199},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 334, col: 16, offset: 10199},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 334, col: 21, offset: 10204},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 334, col: 21, offset: 10204},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 30, offset: 10213},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 341, col: 1, offset: 10335},
			expr: &actionExpr{
				pos: position{line: 341, col: 10, offset: 10346},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 341, col: 10, offset: 10346},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 10, offset: 10346},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 19, offset: 10355},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 341, col: 22, offset: 10358},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 26, offset: 10362},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 341, col: 29, offset: 10365},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 40, offset: 10376},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 341, col: 50, offset: 10386},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 345, col: 1, offset: 10422},
			expr: &actionExpr{
				pos: position{line: 345, col: 13, offset: 10436},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 13, offset: 10436},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 345, col: 18, offset: 10441},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 345, col: 18, offset: 10441},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 29, offset: 10452},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 45, offset: 10468},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 352, col: 1, offset: 10593},
			expr: &actionExpr{
				pos: position{line: 352, col: 12, offset: 10606},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 352, col: 13, offset: 10607},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 352, col: 13, offset: 10607},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 22, offset: 10616},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 31, offset: 10625},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 39, offset: 10633},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 47, offset: 10641},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 55, offset: 10649},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 66, offset: 10660},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 77, offset: 10671},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 356, col: 1, offset: 10731},
			expr: &actionExpr{
				pos: position{line: 356, col: 17, offset: 10749},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 17, offset: 10749},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 22, offset: 10754},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 356, col: 22, offset: 10754},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 32, offset: 10764},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 42, offset: 10774},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 360, col: 1, offset: 10809},
			expr: &actionExpr{
				pos: position{line: 360, col: 11, offset: 10821},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 360, col: 11, offset: 10821},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 360, col: 11, offset: 10821},
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 11, offset: 10821},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 360, col: 20, offset: 10830},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 27, offset: 10837},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 30, offset: 10840},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 34, offset: 10844},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 44, offset: 10854},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 47, offset: 10857},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 51, offset: 10861},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 54, offset: 10864},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 60, offset: 10870},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 70, offset: 10880},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 73, offset: 10883},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 368, col: 1, offset: 11006},
			expr: &actionExpr{
				pos: position{line: 368, col: 11, offset: 11018},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 11, offset: 11018},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 368, col: 11, offset: 11018},
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 11, offset: 11018},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 368, col: 20, offset: 11027},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 27, offset: 11034},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 30, offset: 11037},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 34, offset: 11041},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 44, offset: 11051},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 47, offset: 11054},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 375, col: 1, offset: 11145},
			expr: &actionExpr{
				pos: position{line: 375, col: 12, offset: 11158},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 12, offset: 11158},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 12, offset: 11158},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 20, offset: 11166},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 375, col: 23, offset: 11169},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 27, offset: 11173},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 37, offset: 11183},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 375, col: 40, offset: 11186},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 382, col: 1, offset: 11278},
			expr: &actionExpr{
				pos: position{line: 382, col: 11, offset: 11290},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 11, offset: 11290},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 382, col: 11, offset: 11290},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 382, col: 22, offset: 11301},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 30, offset: 11309},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 386, col: 1, offset: 11346},
			expr: &choiceExpr{
				pos: position{line: 386, col: 14, offset: 11361},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 386, col: 14, offset: 11361},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 24, offset: 11371},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 39, offset: 11386},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 56, offset: 11403},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 70, offset: 11417},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 81, offset: 11428},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 93, offset: 11440},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 388, col: 1, offset: 11452},
			expr: &actionExpr{
				pos: position{line: 388, col: 16, offset: 11469},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 388, col: 17, offset: 11470},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 388, col: 17, offset: 11470},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 388, col: 26, offset: 11479},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 392, col: 1, offset: 11534},
			expr: &actionExpr{
				pos: position{line: 392, col: 15, offset: 11550},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 392, col: 15, offset: 11550},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 15, offset: 11550},
							expr: &charClassMatcher{
								pos:        position{line: 392, col: 15, offset: 11550},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 392, col: 21, offset: 11556},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 21, offset: 11556},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 396, col: 1, offset: 11620},
			expr: &actionExpr{
				pos: position{line: 396, col: 18, offset: 11639},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 396, col: 18, offset: 11639},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 396, col: 18, offset: 11639},
							expr: &charClassMatcher{
								pos:        position{line: 396, col: 18, offset: 11639},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 396, col: 24, offset: 11645},
							expr: &ruleRefExpr{
								pos:  position{line: 396, col: 24, offset: 11645},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 396, col: 31, offset: 11652},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 396, col: 35, offset: 11656},
							expr: &ruleRefExpr{
								pos:  position{line: 396, col: 35, offset: 11656},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 396, col: 42, offset: 11663},
							expr: &seqExpr{
								pos: position{line: 396, col: 44, offset: 11665},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 396, col: 44, offset: 11665},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 396, col: 51, offset: 11672},
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
			pos:  position{line: 400, col: 1, offset: 11742},
			expr: &actionExpr{
				pos: position{line: 400, col: 13, offset: 11756},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 400, col: 13, offset: 11756},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 400, col: 13, offset: 11756},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 17, offset: 11760},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 20, offset: 11763},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 400, col: 27, offset: 11770},
								expr: &seqExpr{
									pos: position{line: 400, col: 28, offset: 11771},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 400, col: 28, offset: 11771},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 39, offset: 11782},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 400, col: 42, offset: 11785},
											expr: &ruleRefExpr{
												pos:  position{line: 400, col: 42, offset: 11785},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 57, offset: 11800},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 62, offset: 11805},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 400, col: 65, offset: 11808},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 409, col: 1, offset: 12002},
			expr: &actionExpr{
				pos: position{line: 409, col: 12, offset: 12015},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 409, col: 12, offset: 12015},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 409, col: 12, offset: 12015},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 409, col: 16, offset: 12019},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 409, col: 19, offset: 12022},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 409, col: 26, offset: 12029},
								expr: &seqExpr{
									pos: position{line: 409, col: 27, offset: 12030},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 409, col: 27, offset: 12030},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 38, offset: 12041},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 409, col: 41, offset: 12044},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 45, offset: 12048},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 48, offset: 12051},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 59, offset: 12062},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 409, col: 63, offset: 12066},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 409, col: 63, offset: 12066},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 409, col: 69, offset: 12072},
													expr: &litMatcher{
														pos:        position{line: 409, col: 70, offset: 12073},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 75, offset: 12078},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 409, col: 80, offset: 12083},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 429, col: 1, offset: 12633},
			expr: &ruleRefExpr{
				pos:  position{line: 429, col: 19, offset: 12653},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 431, col: 1, offset: 12660},
			expr: &actionExpr{
				pos: position{line: 431, col: 9, offset: 12670},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 431, col: 9, offset: 12670},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 431, col: 9, offset: 12670},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 431, col: 16, offset: 12677},
								expr: &seqExpr{
									pos: position{line: 431, col: 17, offset: 12678},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 431, col: 17, offset: 12678},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 431, col: 27, offset: 12688},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 431, col: 32, offset: 12693},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 40, offset: 12701},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 43, offset: 12704},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 431, col: 48, offset: 12709},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 59, offset: 12720},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 62, offset: 12723},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 431, col: 69, offset: 12730},
								expr: &ruleRefExpr{
									pos:  position{line: 431, col: 69, offset: 12730},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 77, offset: 12738},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 431, col: 80, offset: 12741},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 84, offset: 12745},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 87, offset: 12748},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 431, col: 98, offset: 12759},
								expr: &seqExpr{
									pos: position{line: 431, col: 99, offset: 12760},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 431, col: 99, offset: 12760},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 431, col: 109, offset: 12770},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 431, col: 115, offset: 12776},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 431, col: 115, offset: 12776},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 431, col: 121, offset: 12782},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 138, offset: 12799},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 452, col: 1, offset: 13344},
			expr: &actionExpr{
				pos: position{line: 452, col: 19, offset: 13364},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 452, col: 19, offset: 13364,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 456, col: 1, offset: 13431},
			expr: &actionExpr{
				pos: position{line: 456, col: 10, offset: 13442},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 456, col: 10, offset: 13442},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 456, col: 10, offset: 13442},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 19, offset: 13451},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 22, offset: 13454},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 456, col: 34, offset: 13466},
							expr: &seqExpr{
								pos: position{line: 456, col: 35, offset: 13467},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 456, col: 35, offset: 13467},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 456, col: 39, offset: 13471},
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
			pos:  position{line: 461, col: 1, offset: 13602},
			expr: &choiceExpr{
				pos: position{line: 461, col: 15, offset: 13618},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 461, col: 16, offset: 13619},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 461, col: 16, offset: 13619},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 461, col: 20, offset: 13623},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 461, col: 31, offset: 13634},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 461, col: 38, offset: 13641},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 463, col: 1, offset: 13653},
			expr: &oneOrMoreExpr{
				pos: position{line: 463, col: 14, offset: 13668},
				expr: &charClassMatcher{
					pos:        position{line: 463, col: 14, offset: 13668},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 465, col: 1, offset: 13686},
			expr: &actionExpr{
				pos: position{line: 465, col: 13, offset: 13700},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 465, col: 13, offset: 13700},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 465, col: 13, offset: 13700},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 465, col: 20, offset: 13707},
								expr: &seqExpr{
									pos: position{line: 465, col: 21, offset: 13708},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 465, col: 21, offset: 13708},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 465, col: 31, offset: 13718},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 465, col: 36, offset: 13723},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 465, col: 41, offset: 13728},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 52, offset: 13739},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 465, col: 54, offset: 13741},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 58, offset: 13745},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 465, col: 61, offset: 13748},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 465, col: 65, offset: 13752},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 76, offset: 13763},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 481, col: 1, offset: 14274},
			expr: &actionExpr{
				pos: position{line: 481, col: 11, offset: 14286},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 481, col: 12, offset: 14287},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 481, col: 13, offset: 14288},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 13, offset: 14288},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 17, offset: 14292},
									expr: &choiceExpr{
										pos: position{line: 481, col: 18, offset: 14293},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 18, offset: 14293},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 25, offset: 14300},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 32, offset: 14307},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 481, col: 40, offset: 14315},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 40, offset: 14315},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 45, offset: 14320},
									expr: &choiceExpr{
										pos: position{line: 481, col: 46, offset: 14321},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 46, offset: 14321},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 53, offset: 14328},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 60, offset: 14335},
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
			pos:  position{line: 488, col: 1, offset: 14551},
			expr: &actionExpr{
				pos: position{line: 488, col: 14, offset: 14566},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 488, col: 14, offset: 14566},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 488, col: 14, offset: 14566},
							expr: &choiceExpr{
								pos: position{line: 488, col: 15, offset: 14567},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 15, offset: 14567},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 488, col: 24, offset: 14576},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 488, col: 30, offset: 14582},
							expr: &choiceExpr{
								pos: position{line: 488, col: 31, offset: 14583},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 31, offset: 14583},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 488, col: 40, offset: 14592},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 488, col: 48, offset: 14600},
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
			pos:  position{line: 492, col: 1, offset: 14655},
			expr: &charClassMatcher{
				pos:        position{line: 492, col: 17, offset: 14673},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 493, col: 1, offset: 14678},
			expr: &charClassMatcher{
				pos:        position{line: 493, col: 10, offset: 14689},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 494, col: 1, offset: 14698},
			expr: &charClassMatcher{
				pos:        position{line: 494, col: 9, offset: 14708},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 496, col: 1, offset: 14715},
			expr: &anyMatcher{
				line: 496, col: 14, offset: 14730,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 497, col: 1, offset: 14732},
			expr: &actionExpr{
				pos: position{line: 497, col: 13, offset: 14746},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 497, col: 13, offset: 14746},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 497, col: 13, offset: 14746},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 497, col: 20, offset: 14753},
							expr: &seqExpr{
								pos: position{line: 497, col: 22, offset: 14755},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 497, col: 22, offset: 14755},
										expr: &litMatcher{
											pos:        position{line: 497, col: 23, offset: 14756},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 497, col: 28, offset: 14761},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 497, col: 42, offset: 14775},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 503, col: 1, offset: 14955},
			expr: &choiceExpr{
				pos: position{line: 503, col: 11, offset: 14967},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 503, col: 11, offset: 14967},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 503, col: 30, offset: 14986},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 504, col: 1, offset: 15004},
			expr: &seqExpr{
				pos: position{line: 504, col: 20, offset: 15025},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 504, col: 20, offset: 15025},
						expr: &ruleRefExpr{
							pos:  position{line: 504, col: 21, offset: 15026},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 31, offset: 15036},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 504, col: 36, offset: 15041},
						expr: &seqExpr{
							pos: position{line: 504, col: 38, offset: 15043},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 504, col: 38, offset: 15043},
									expr: &litMatcher{
										pos:        position{line: 504, col: 39, offset: 15044},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 504, col: 44, offset: 15049},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 58, offset: 15063},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 505, col: 1, offset: 15068},
			expr: &seqExpr{
				pos: position{line: 505, col: 36, offset: 15105},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 505, col: 36, offset: 15105},
						expr: &ruleRefExpr{
							pos:  position{line: 505, col: 37, offset: 15106},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 47, offset: 15116},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 505, col: 52, offset: 15121},
						expr: &seqExpr{
							pos: position{line: 505, col: 54, offset: 15123},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 505, col: 54, offset: 15123},
									expr: &choiceExpr{
										pos: position{line: 505, col: 57, offset: 15126},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 505, col: 57, offset: 15126},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 505, col: 64, offset: 15133},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 505, col: 70, offset: 15139},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 84, offset: 15153},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 506, col: 1, offset: 15158},
			expr: &choiceExpr{
				pos: position{line: 506, col: 21, offset: 15180},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 506, col: 22, offset: 15181},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 22, offset: 15181},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 27, offset: 15186},
								expr: &seqExpr{
									pos: position{line: 506, col: 29, offset: 15188},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 29, offset: 15188},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 30, offset: 15189},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 34, offset: 15193},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 506, col: 52, offset: 15211},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 52, offset: 15211},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 56, offset: 15215},
								expr: &seqExpr{
									pos: position{line: 506, col: 58, offset: 15217},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 58, offset: 15217},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 59, offset: 15218},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 63, offset: 15222},
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
			pos:  position{line: 508, col: 1, offset: 15238},
			expr: &zeroOrMoreExpr{
				pos: position{line: 508, col: 6, offset: 15245},
				expr: &choiceExpr{
					pos: position{line: 508, col: 8, offset: 15247},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 508, col: 8, offset: 15247},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 21, offset: 15260},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 27, offset: 15266},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 509, col: 1, offset: 15277},
			expr: &zeroOrMoreExpr{
				pos: position{line: 509, col: 5, offset: 15283},
				expr: &choiceExpr{
					pos: position{line: 509, col: 7, offset: 15285},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 509, col: 7, offset: 15285},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 509, col: 20, offset: 15298},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 510, col: 1, offset: 15334},
			expr: &zeroOrMoreExpr{
				pos: position{line: 510, col: 6, offset: 15341},
				expr: &ruleRefExpr{
					pos:  position{line: 510, col: 6, offset: 15341},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 512, col: 1, offset: 15354},
			expr: &charClassMatcher{
				pos:        position{line: 512, col: 14, offset: 15369},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 513, col: 1, offset: 15377},
			expr: &litMatcher{
				pos:        position{line: 513, col: 7, offset: 15385},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 514, col: 1, offset: 15390},
			expr: &choiceExpr{
				pos: position{line: 514, col: 7, offset: 15398},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 514, col: 7, offset: 15398},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 7, offset: 15398},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 514, col: 10, offset: 15401},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 16, offset: 15407},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 16, offset: 15407},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 514, col: 18, offset: 15409},
								expr: &ruleRefExpr{
									pos:  position{line: 514, col: 18, offset: 15409},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 37, offset: 15428},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 43, offset: 15434},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 43, offset: 15434},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 46, offset: 15437},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 516, col: 1, offset: 15442},
			expr: &notExpr{
				pos: position{line: 516, col: 7, offset: 15450},
				expr: &anyMatcher{
					line: 516, col: 8, offset: 15451,
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
		Values: make([]*EnumValue, len(vs)),
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
