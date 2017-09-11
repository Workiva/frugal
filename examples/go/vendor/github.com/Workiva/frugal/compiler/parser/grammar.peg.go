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
			pos:  position{line: 150, col: 1, offset: 4792},
			expr: &actionExpr{
				pos: position{line: 150, col: 16, offset: 4807},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 150, col: 16, offset: 4807,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 154, col: 1, offset: 4865},
			expr: &actionExpr{
				pos: position{line: 154, col: 14, offset: 4878},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 154, col: 14, offset: 4878},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 154, col: 14, offset: 4878},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 154, col: 21, offset: 4885},
								expr: &seqExpr{
									pos: position{line: 154, col: 22, offset: 4886},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 154, col: 22, offset: 4886},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 154, col: 32, offset: 4896},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 154, col: 37, offset: 4901},
							label: "statement",
							expr: &ruleRefExpr{
								pos:  position{line: 154, col: 47, offset: 4911},
								name: "FrugalStatement",
							},
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 167, col: 1, offset: 5381},
			expr: &choiceExpr{
				pos: position{line: 167, col: 20, offset: 5400},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 167, col: 20, offset: 5400},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 30, offset: 5410},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 42, offset: 5422},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 50, offset: 5430},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 57, offset: 5437},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 67, offset: 5447},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 76, offset: 5456},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 88, offset: 5468},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 96, offset: 5476},
						name: "Service",
					},
					&ruleRefExpr{
						pos:  position{line: 167, col: 106, offset: 5486},
						name: "Scope",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 169, col: 1, offset: 5493},
			expr: &actionExpr{
				pos: position{line: 169, col: 12, offset: 5504},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 169, col: 12, offset: 5504},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 169, col: 12, offset: 5504},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 22, offset: 5514},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 169, col: 24, offset: 5516},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 169, col: 29, offset: 5521},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 37, offset: 5529},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 169, col: 39, offset: 5531},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 169, col: 51, offset: 5543},
								expr: &ruleRefExpr{
									pos:  position{line: 169, col: 51, offset: 5543},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 169, col: 68, offset: 5560},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 181, col: 1, offset: 5837},
			expr: &actionExpr{
				pos: position{line: 181, col: 14, offset: 5850},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 181, col: 14, offset: 5850},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 181, col: 14, offset: 5850},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 26, offset: 5862},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 28, offset: 5864},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 181, col: 34, offset: 5870},
								expr: &charClassMatcher{
									pos:        position{line: 181, col: 34, offset: 5870},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 44, offset: 5880},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 46, offset: 5882},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 181, col: 49, offset: 5885},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 60, offset: 5896},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 62, offset: 5898},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 181, col: 74, offset: 5910},
								expr: &ruleRefExpr{
									pos:  position{line: 181, col: 74, offset: 5910},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 91, offset: 5927},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 189, col: 1, offset: 6113},
			expr: &actionExpr{
				pos: position{line: 189, col: 10, offset: 6122},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 189, col: 10, offset: 6122},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 189, col: 10, offset: 6122},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 18, offset: 6130},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 20, offset: 6132},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 24, offset: 6136},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 34, offset: 6146},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 36, offset: 6148},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 41, offset: 6153},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 52, offset: 6164},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 189, col: 54, offset: 6166},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 58, offset: 6170},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 60, offset: 6172},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 66, offset: 6178},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 77, offset: 6189},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 79, offset: 6191},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 189, col: 91, offset: 6203},
								expr: &ruleRefExpr{
									pos:  position{line: 189, col: 91, offset: 6203},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 108, offset: 6220},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 198, col: 1, offset: 6414},
			expr: &actionExpr{
				pos: position{line: 198, col: 9, offset: 6422},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 198, col: 9, offset: 6422},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 198, col: 9, offset: 6422},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 16, offset: 6429},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 18, offset: 6431},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 198, col: 23, offset: 6436},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 34, offset: 6447},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 198, col: 37, offset: 6450},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 41, offset: 6454},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 44, offset: 6457},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 198, col: 51, offset: 6464},
								expr: &seqExpr{
									pos: position{line: 198, col: 52, offset: 6465},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 198, col: 52, offset: 6465},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 198, col: 62, offset: 6475},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 198, col: 67, offset: 6480},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 71, offset: 6484},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 73, offset: 6486},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 198, col: 85, offset: 6498},
								expr: &ruleRefExpr{
									pos:  position{line: 198, col: 85, offset: 6498},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 102, offset: 6515},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 222, col: 1, offset: 7177},
			expr: &actionExpr{
				pos: position{line: 222, col: 14, offset: 7190},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 222, col: 14, offset: 7190},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 222, col: 14, offset: 7190},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 21, offset: 7197},
								expr: &seqExpr{
									pos: position{line: 222, col: 22, offset: 7198},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 222, col: 22, offset: 7198},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 32, offset: 7208},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 222, col: 37, offset: 7213},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 42, offset: 7218},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 53, offset: 7229},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 55, offset: 7231},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 61, offset: 7237},
								expr: &seqExpr{
									pos: position{line: 222, col: 62, offset: 7238},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 222, col: 62, offset: 7238},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 66, offset: 7242},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 68, offset: 7244},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 82, offset: 7258},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 84, offset: 7260},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 96, offset: 7272},
								expr: &ruleRefExpr{
									pos:  position{line: 222, col: 96, offset: 7272},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 222, col: 113, offset: 7289},
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 113, offset: 7289},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 238, col: 1, offset: 7687},
			expr: &actionExpr{
				pos: position{line: 238, col: 12, offset: 7698},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 238, col: 12, offset: 7698},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 238, col: 12, offset: 7698},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 22, offset: 7708},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 24, offset: 7710},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 28, offset: 7714},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 38, offset: 7724},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7726},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 45, offset: 7731},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 56, offset: 7742},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 58, offset: 7744},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 238, col: 70, offset: 7756},
								expr: &ruleRefExpr{
									pos:  position{line: 238, col: 70, offset: 7756},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 87, offset: 7773},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 246, col: 1, offset: 7945},
			expr: &actionExpr{
				pos: position{line: 246, col: 11, offset: 7955},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 246, col: 11, offset: 7955},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 246, col: 11, offset: 7955},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 246, col: 20, offset: 7964},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 246, col: 22, offset: 7966},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 246, col: 25, offset: 7969},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 247, col: 1, offset: 8009},
			expr: &actionExpr{
				pos: position{line: 247, col: 14, offset: 8022},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 247, col: 14, offset: 8022},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 247, col: 14, offset: 8022},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 247, col: 26, offset: 8034},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 247, col: 28, offset: 8036},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 247, col: 31, offset: 8039},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 248, col: 1, offset: 8090},
			expr: &actionExpr{
				pos: position{line: 248, col: 10, offset: 8099},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 248, col: 10, offset: 8099},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 248, col: 10, offset: 8099},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 18, offset: 8107},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 20, offset: 8109},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 23, offset: 8112},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 249, col: 1, offset: 8159},
			expr: &actionExpr{
				pos: position{line: 249, col: 15, offset: 8173},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 249, col: 15, offset: 8173},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 249, col: 15, offset: 8173},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 20, offset: 8178},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 31, offset: 8189},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 249, col: 34, offset: 8192},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 38, offset: 8196},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 41, offset: 8199},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 48, offset: 8206},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 249, col: 58, offset: 8216},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 62, offset: 8220},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 64, offset: 8222},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 249, col: 76, offset: 8234},
								expr: &ruleRefExpr{
									pos:  position{line: 249, col: 76, offset: 8234},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 93, offset: 8251},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 260, col: 1, offset: 8468},
			expr: &actionExpr{
				pos: position{line: 260, col: 14, offset: 8481},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 260, col: 14, offset: 8481},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 260, col: 21, offset: 8488},
						expr: &seqExpr{
							pos: position{line: 260, col: 22, offset: 8489},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 260, col: 22, offset: 8489},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 260, col: 28, offset: 8495},
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
			pos:  position{line: 269, col: 1, offset: 8676},
			expr: &actionExpr{
				pos: position{line: 269, col: 10, offset: 8685},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 269, col: 10, offset: 8685},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 269, col: 10, offset: 8685},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 17, offset: 8692},
								expr: &seqExpr{
									pos: position{line: 269, col: 18, offset: 8693},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 269, col: 18, offset: 8693},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 28, offset: 8703},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 269, col: 33, offset: 8708},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 36, offset: 8711},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 48, offset: 8723},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 269, col: 50, offset: 8725},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 54, offset: 8729},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 56, offset: 8731},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 60, offset: 8735},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 60, offset: 8735},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 75, offset: 8750},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 77, offset: 8752},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 81, offset: 8756},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 91, offset: 8766},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 93, offset: 8768},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 98, offset: 8773},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 109, offset: 8784},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 112, offset: 8787},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 116, offset: 8791},
								expr: &seqExpr{
									pos: position{line: 269, col: 117, offset: 8792},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 269, col: 117, offset: 8792},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 121, offset: 8796},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 123, offset: 8798},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 136, offset: 8811},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 138, offset: 8813},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 150, offset: 8825},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 150, offset: 8825},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 269, col: 167, offset: 8842},
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 167, offset: 8842},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 292, col: 1, offset: 9374},
			expr: &actionExpr{
				pos: position{line: 292, col: 18, offset: 9391},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 292, col: 19, offset: 9392},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 292, col: 19, offset: 9392},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 292, col: 32, offset: 9405},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 300, col: 1, offset: 9548},
			expr: &actionExpr{
				pos: position{line: 300, col: 12, offset: 9559},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 300, col: 12, offset: 9559},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 300, col: 12, offset: 9559},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 22, offset: 9569},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 24, offset: 9571},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 29, offset: 9576},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 40, offset: 9587},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 42, offset: 9589},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 50, offset: 9597},
								expr: &seqExpr{
									pos: position{line: 300, col: 51, offset: 9598},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 300, col: 51, offset: 9598},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 61, offset: 9608},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 64, offset: 9611},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 75, offset: 9622},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 80, offset: 9627},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 300, col: 83, offset: 9630},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 87, offset: 9634},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 90, offset: 9637},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 300, col: 98, offset: 9645},
								expr: &seqExpr{
									pos: position{line: 300, col: 99, offset: 9646},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 300, col: 99, offset: 9646},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 108, offset: 9655},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 300, col: 114, offset: 9661},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 300, col: 114, offset: 9661},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 300, col: 120, offset: 9667},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 139, offset: 9686},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 141, offset: 9688},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 153, offset: 9700},
								expr: &ruleRefExpr{
									pos:  position{line: 300, col: 153, offset: 9700},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 170, offset: 9717},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 317, col: 1, offset: 10158},
			expr: &actionExpr{
				pos: position{line: 317, col: 22, offset: 10179},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 317, col: 22, offset: 10179,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 321, col: 1, offset: 10248},
			expr: &actionExpr{
				pos: position{line: 321, col: 13, offset: 10260},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 321, col: 13, offset: 10260},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 321, col: 13, offset: 10260},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 20, offset: 10267},
								expr: &seqExpr{
									pos: position{line: 321, col: 21, offset: 10268},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 321, col: 21, offset: 10268},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 31, offset: 10278},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 36, offset: 10283},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 43, offset: 10290},
								expr: &seqExpr{
									pos: position{line: 321, col: 44, offset: 10291},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 321, col: 44, offset: 10291},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 53, offset: 10300},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 58, offset: 10305},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 62, offset: 10309},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 75, offset: 10322},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 78, offset: 10325},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 83, offset: 10330},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 94, offset: 10341},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 321, col: 96, offset: 10343},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 100, offset: 10347},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 103, offset: 10350},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 113, offset: 10360},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 321, col: 123, offset: 10370},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 127, offset: 10374},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 130, offset: 10377},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 141, offset: 10388},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 141, offset: 10388},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 149, offset: 10396},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 151, offset: 10398},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 163, offset: 10410},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 163, offset: 10410},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 321, col: 180, offset: 10427},
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 180, offset: 10427},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 349, col: 1, offset: 11078},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 11094},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 11094},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 11099},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 349, col: 22, offset: 11099},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 31, offset: 11108},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 356, col: 1, offset: 11230},
			expr: &actionExpr{
				pos: position{line: 356, col: 11, offset: 11240},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 356, col: 11, offset: 11240},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 356, col: 11, offset: 11240},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 20, offset: 11249},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 356, col: 23, offset: 11252},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 27, offset: 11256},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 30, offset: 11259},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 41, offset: 11270},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 356, col: 51, offset: 11280},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 360, col: 1, offset: 11316},
			expr: &actionExpr{
				pos: position{line: 360, col: 14, offset: 11329},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 360, col: 14, offset: 11329},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 360, col: 19, offset: 11334},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 360, col: 19, offset: 11334},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 30, offset: 11345},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 46, offset: 11361},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 367, col: 1, offset: 11486},
			expr: &actionExpr{
				pos: position{line: 367, col: 13, offset: 11498},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 367, col: 13, offset: 11498},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 367, col: 13, offset: 11498},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 367, col: 18, offset: 11503},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 367, col: 31, offset: 11516},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 367, col: 33, offset: 11518},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 367, col: 45, offset: 11530},
								expr: &ruleRefExpr{
									pos:  position{line: 367, col: 45, offset: 11530},
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
			pos:  position{line: 374, col: 1, offset: 11666},
			expr: &actionExpr{
				pos: position{line: 374, col: 17, offset: 11682},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 374, col: 18, offset: 11683},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 374, col: 18, offset: 11683},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 27, offset: 11692},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 36, offset: 11701},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 44, offset: 11709},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 52, offset: 11717},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 60, offset: 11725},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 71, offset: 11736},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 82, offset: 11747},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 378, col: 1, offset: 11794},
			expr: &actionExpr{
				pos: position{line: 378, col: 18, offset: 11811},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 378, col: 18, offset: 11811},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 378, col: 23, offset: 11816},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 378, col: 23, offset: 11816},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 33, offset: 11826},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 43, offset: 11836},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 382, col: 1, offset: 11871},
			expr: &actionExpr{
				pos: position{line: 382, col: 12, offset: 11882},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 12, offset: 11882},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 382, col: 12, offset: 11882},
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 12, offset: 11882},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 382, col: 21, offset: 11891},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 28, offset: 11898},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 31, offset: 11901},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 35, offset: 11905},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 45, offset: 11915},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 48, offset: 11918},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 52, offset: 11922},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 55, offset: 11925},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 61, offset: 11931},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 71, offset: 11941},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 74, offset: 11944},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 78, offset: 11948},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 80, offset: 11950},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 382, col: 92, offset: 11962},
								expr: &ruleRefExpr{
									pos:  position{line: 382, col: 92, offset: 11962},
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
			pos:  position{line: 391, col: 1, offset: 12160},
			expr: &actionExpr{
				pos: position{line: 391, col: 12, offset: 12171},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 391, col: 12, offset: 12171},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 391, col: 12, offset: 12171},
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 12, offset: 12171},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 391, col: 21, offset: 12180},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 28, offset: 12187},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 31, offset: 12190},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 35, offset: 12194},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 45, offset: 12204},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 391, col: 48, offset: 12207},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 52, offset: 12211},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 54, offset: 12213},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 391, col: 66, offset: 12225},
								expr: &ruleRefExpr{
									pos:  position{line: 391, col: 66, offset: 12225},
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
			pos:  position{line: 399, col: 1, offset: 12387},
			expr: &actionExpr{
				pos: position{line: 399, col: 13, offset: 12399},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 399, col: 13, offset: 12399},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 399, col: 13, offset: 12399},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 21, offset: 12407},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 24, offset: 12410},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 28, offset: 12414},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 38, offset: 12424},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 399, col: 41, offset: 12427},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 45, offset: 12431},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 47, offset: 12433},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 399, col: 59, offset: 12445},
								expr: &ruleRefExpr{
									pos:  position{line: 399, col: 59, offset: 12445},
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
			pos:  position{line: 407, col: 1, offset: 12608},
			expr: &actionExpr{
				pos: position{line: 407, col: 12, offset: 12619},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 407, col: 12, offset: 12619},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 407, col: 12, offset: 12619},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 407, col: 23, offset: 12630},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 407, col: 31, offset: 12638},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 411, col: 1, offset: 12675},
			expr: &choiceExpr{
				pos: position{line: 411, col: 15, offset: 12689},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 411, col: 15, offset: 12689},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 25, offset: 12699},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 40, offset: 12714},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 57, offset: 12731},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 71, offset: 12745},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 82, offset: 12756},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 94, offset: 12768},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 413, col: 1, offset: 12780},
			expr: &actionExpr{
				pos: position{line: 413, col: 20, offset: 12799},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 413, col: 20, offset: 12799},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 413, col: 20, offset: 12799},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 413, col: 24, offset: 12803},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 413, col: 27, offset: 12806},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 413, col: 39, offset: 12818},
								expr: &ruleRefExpr{
									pos:  position{line: 413, col: 39, offset: 12818},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 413, col: 55, offset: 12834},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 421, col: 1, offset: 12998},
			expr: &actionExpr{
				pos: position{line: 421, col: 19, offset: 13016},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 421, col: 19, offset: 13016},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 421, col: 19, offset: 13016},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 24, offset: 13021},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 35, offset: 13032},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 421, col: 37, offset: 13034},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 421, col: 43, offset: 13040},
								expr: &actionExpr{
									pos: position{line: 421, col: 44, offset: 13041},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 421, col: 44, offset: 13041},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 421, col: 44, offset: 13041},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 421, col: 48, offset: 13045},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 421, col: 51, offset: 13048},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 421, col: 57, offset: 13054},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 421, col: 89, offset: 13086},
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 89, offset: 13086},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 104, offset: 13101},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 432, col: 1, offset: 13297},
			expr: &actionExpr{
				pos: position{line: 432, col: 17, offset: 13313},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 432, col: 18, offset: 13314},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 432, col: 18, offset: 13314},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 432, col: 27, offset: 13323},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 436, col: 1, offset: 13378},
			expr: &actionExpr{
				pos: position{line: 436, col: 16, offset: 13393},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 436, col: 16, offset: 13393},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 436, col: 16, offset: 13393},
							expr: &charClassMatcher{
								pos:        position{line: 436, col: 16, offset: 13393},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 436, col: 22, offset: 13399},
							expr: &ruleRefExpr{
								pos:  position{line: 436, col: 22, offset: 13399},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 440, col: 1, offset: 13463},
			expr: &actionExpr{
				pos: position{line: 440, col: 19, offset: 13481},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 440, col: 19, offset: 13481},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 440, col: 19, offset: 13481},
							expr: &charClassMatcher{
								pos:        position{line: 440, col: 19, offset: 13481},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 25, offset: 13487},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 25, offset: 13487},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 440, col: 32, offset: 13494},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 36, offset: 13498},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 36, offset: 13498},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 440, col: 43, offset: 13505},
							expr: &seqExpr{
								pos: position{line: 440, col: 45, offset: 13507},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 440, col: 45, offset: 13507},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 440, col: 52, offset: 13514},
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
			pos:  position{line: 444, col: 1, offset: 13584},
			expr: &actionExpr{
				pos: position{line: 444, col: 14, offset: 13597},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 444, col: 14, offset: 13597},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 444, col: 14, offset: 13597},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 18, offset: 13601},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 444, col: 21, offset: 13604},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 444, col: 28, offset: 13611},
								expr: &seqExpr{
									pos: position{line: 444, col: 29, offset: 13612},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 444, col: 29, offset: 13612},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 40, offset: 13623},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 444, col: 43, offset: 13626},
											expr: &ruleRefExpr{
												pos:  position{line: 444, col: 43, offset: 13626},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 58, offset: 13641},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 63, offset: 13646},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 444, col: 66, offset: 13649},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 453, col: 1, offset: 13843},
			expr: &actionExpr{
				pos: position{line: 453, col: 13, offset: 13855},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 453, col: 13, offset: 13855},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 453, col: 13, offset: 13855},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 453, col: 17, offset: 13859},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 453, col: 20, offset: 13862},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 453, col: 27, offset: 13869},
								expr: &seqExpr{
									pos: position{line: 453, col: 28, offset: 13870},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 453, col: 28, offset: 13870},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 39, offset: 13881},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 453, col: 42, offset: 13884},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 46, offset: 13888},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 49, offset: 13891},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 60, offset: 13902},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 453, col: 64, offset: 13906},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 453, col: 64, offset: 13906},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 453, col: 70, offset: 13912},
													expr: &litMatcher{
														pos:        position{line: 453, col: 71, offset: 13913},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 76, offset: 13918},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 453, col: 81, offset: 13923},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 473, col: 1, offset: 14473},
			expr: &actionExpr{
				pos: position{line: 473, col: 10, offset: 14482},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 473, col: 10, offset: 14482},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 473, col: 10, offset: 14482},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 17, offset: 14489},
								expr: &seqExpr{
									pos: position{line: 473, col: 18, offset: 14490},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 473, col: 18, offset: 14490},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 28, offset: 14500},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 473, col: 33, offset: 14505},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 41, offset: 14513},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 44, offset: 14516},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 473, col: 49, offset: 14521},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 60, offset: 14532},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 63, offset: 14535},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 70, offset: 14542},
								expr: &ruleRefExpr{
									pos:  position{line: 473, col: 70, offset: 14542},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 78, offset: 14550},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 473, col: 81, offset: 14553},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 85, offset: 14557},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 88, offset: 14560},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 473, col: 99, offset: 14571},
								expr: &seqExpr{
									pos: position{line: 473, col: 100, offset: 14572},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 473, col: 100, offset: 14572},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 110, offset: 14582},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 473, col: 116, offset: 14588},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 473, col: 116, offset: 14588},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 473, col: 122, offset: 14594},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 139, offset: 14611},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 141, offset: 14613},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 153, offset: 14625},
								expr: &ruleRefExpr{
									pos:  position{line: 473, col: 153, offset: 14625},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 170, offset: 14642},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 495, col: 1, offset: 15239},
			expr: &actionExpr{
				pos: position{line: 495, col: 20, offset: 15258},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 495, col: 20, offset: 15258,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 499, col: 1, offset: 15325},
			expr: &actionExpr{
				pos: position{line: 499, col: 11, offset: 15335},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 499, col: 11, offset: 15335},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 499, col: 11, offset: 15335},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 20, offset: 15344},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 23, offset: 15347},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 499, col: 35, offset: 15359},
							expr: &seqExpr{
								pos: position{line: 499, col: 36, offset: 15360},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 499, col: 36, offset: 15360},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 499, col: 40, offset: 15364},
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
			pos:  position{line: 504, col: 1, offset: 15495},
			expr: &choiceExpr{
				pos: position{line: 504, col: 16, offset: 15510},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 504, col: 17, offset: 15511},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 504, col: 17, offset: 15511},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 504, col: 21, offset: 15515},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 504, col: 32, offset: 15526},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 504, col: 39, offset: 15533},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 506, col: 1, offset: 15545},
			expr: &oneOrMoreExpr{
				pos: position{line: 506, col: 15, offset: 15559},
				expr: &charClassMatcher{
					pos:        position{line: 506, col: 15, offset: 15559},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 508, col: 1, offset: 15577},
			expr: &actionExpr{
				pos: position{line: 508, col: 14, offset: 15590},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 508, col: 14, offset: 15590},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 508, col: 14, offset: 15590},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 508, col: 21, offset: 15597},
								expr: &seqExpr{
									pos: position{line: 508, col: 22, offset: 15598},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 508, col: 22, offset: 15598},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 508, col: 32, offset: 15608},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 508, col: 37, offset: 15613},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 42, offset: 15618},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 53, offset: 15629},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 508, col: 55, offset: 15631},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 59, offset: 15635},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 508, col: 62, offset: 15638},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 66, offset: 15642},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 76, offset: 15652},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 508, col: 78, offset: 15654},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 508, col: 90, offset: 15666},
								expr: &ruleRefExpr{
									pos:  position{line: 508, col: 90, offset: 15666},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 508, col: 107, offset: 15683},
							expr: &ruleRefExpr{
								pos:  position{line: 508, col: 107, offset: 15683},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 525, col: 1, offset: 16243},
			expr: &actionExpr{
				pos: position{line: 525, col: 12, offset: 16254},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 525, col: 13, offset: 16255},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 525, col: 14, offset: 16256},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 525, col: 14, offset: 16256},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 525, col: 18, offset: 16260},
									expr: &choiceExpr{
										pos: position{line: 525, col: 19, offset: 16261},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 525, col: 19, offset: 16261},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 525, col: 26, offset: 16268},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 525, col: 33, offset: 16275},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 525, col: 41, offset: 16283},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 525, col: 41, offset: 16283},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 525, col: 46, offset: 16288},
									expr: &choiceExpr{
										pos: position{line: 525, col: 47, offset: 16289},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 525, col: 47, offset: 16289},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 525, col: 54, offset: 16296},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 525, col: 61, offset: 16303},
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
			pos:  position{line: 534, col: 1, offset: 16589},
			expr: &actionExpr{
				pos: position{line: 534, col: 15, offset: 16603},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 534, col: 15, offset: 16603},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 534, col: 15, offset: 16603},
							expr: &choiceExpr{
								pos: position{line: 534, col: 16, offset: 16604},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 534, col: 16, offset: 16604},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 534, col: 25, offset: 16613},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 534, col: 31, offset: 16619},
							expr: &choiceExpr{
								pos: position{line: 534, col: 32, offset: 16620},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 534, col: 32, offset: 16620},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 534, col: 41, offset: 16629},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 534, col: 49, offset: 16637},
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
			pos:  position{line: 538, col: 1, offset: 16692},
			expr: &charClassMatcher{
				pos:        position{line: 538, col: 18, offset: 16709},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 539, col: 1, offset: 16714},
			expr: &charClassMatcher{
				pos:        position{line: 539, col: 11, offset: 16724},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 540, col: 1, offset: 16733},
			expr: &charClassMatcher{
				pos:        position{line: 540, col: 10, offset: 16742},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 542, col: 1, offset: 16749},
			expr: &anyMatcher{
				line: 542, col: 15, offset: 16763,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 543, col: 1, offset: 16765},
			expr: &actionExpr{
				pos: position{line: 543, col: 14, offset: 16778},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 543, col: 14, offset: 16778},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 543, col: 14, offset: 16778},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 543, col: 21, offset: 16785},
							expr: &seqExpr{
								pos: position{line: 543, col: 23, offset: 16787},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 543, col: 23, offset: 16787},
										expr: &litMatcher{
											pos:        position{line: 543, col: 24, offset: 16788},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 543, col: 29, offset: 16793},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 543, col: 43, offset: 16807},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 549, col: 1, offset: 16987},
			expr: &choiceExpr{
				pos: position{line: 549, col: 12, offset: 16998},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 549, col: 12, offset: 16998},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 549, col: 31, offset: 17017},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 550, col: 1, offset: 17035},
			expr: &seqExpr{
				pos: position{line: 550, col: 21, offset: 17055},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 550, col: 21, offset: 17055},
						expr: &ruleRefExpr{
							pos:  position{line: 550, col: 22, offset: 17056},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 550, col: 32, offset: 17066},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 550, col: 37, offset: 17071},
						expr: &seqExpr{
							pos: position{line: 550, col: 39, offset: 17073},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 550, col: 39, offset: 17073},
									expr: &litMatcher{
										pos:        position{line: 550, col: 40, offset: 17074},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 550, col: 45, offset: 17079},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 550, col: 59, offset: 17093},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 551, col: 1, offset: 17098},
			expr: &seqExpr{
				pos: position{line: 551, col: 37, offset: 17134},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 551, col: 37, offset: 17134},
						expr: &ruleRefExpr{
							pos:  position{line: 551, col: 38, offset: 17135},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 551, col: 48, offset: 17145},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 551, col: 53, offset: 17150},
						expr: &seqExpr{
							pos: position{line: 551, col: 55, offset: 17152},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 551, col: 55, offset: 17152},
									expr: &choiceExpr{
										pos: position{line: 551, col: 58, offset: 17155},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 551, col: 58, offset: 17155},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 551, col: 65, offset: 17162},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 551, col: 71, offset: 17168},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 551, col: 85, offset: 17182},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 552, col: 1, offset: 17187},
			expr: &choiceExpr{
				pos: position{line: 552, col: 22, offset: 17208},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 552, col: 23, offset: 17209},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 552, col: 23, offset: 17209},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 552, col: 28, offset: 17214},
								expr: &seqExpr{
									pos: position{line: 552, col: 30, offset: 17216},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 552, col: 30, offset: 17216},
											expr: &ruleRefExpr{
												pos:  position{line: 552, col: 31, offset: 17217},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 552, col: 35, offset: 17221},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 552, col: 53, offset: 17239},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 552, col: 53, offset: 17239},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 552, col: 57, offset: 17243},
								expr: &seqExpr{
									pos: position{line: 552, col: 59, offset: 17245},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 552, col: 59, offset: 17245},
											expr: &ruleRefExpr{
												pos:  position{line: 552, col: 60, offset: 17246},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 552, col: 64, offset: 17250},
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
			pos:  position{line: 554, col: 1, offset: 17266},
			expr: &zeroOrMoreExpr{
				pos: position{line: 554, col: 7, offset: 17272},
				expr: &choiceExpr{
					pos: position{line: 554, col: 9, offset: 17274},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 554, col: 9, offset: 17274},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 554, col: 22, offset: 17287},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 554, col: 28, offset: 17293},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 555, col: 1, offset: 17304},
			expr: &zeroOrMoreExpr{
				pos: position{line: 555, col: 6, offset: 17309},
				expr: &choiceExpr{
					pos: position{line: 555, col: 8, offset: 17311},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 555, col: 8, offset: 17311},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 555, col: 21, offset: 17324},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 556, col: 1, offset: 17360},
			expr: &zeroOrMoreExpr{
				pos: position{line: 556, col: 7, offset: 17366},
				expr: &ruleRefExpr{
					pos:  position{line: 556, col: 7, offset: 17366},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 558, col: 1, offset: 17379},
			expr: &charClassMatcher{
				pos:        position{line: 558, col: 15, offset: 17393},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 559, col: 1, offset: 17401},
			expr: &litMatcher{
				pos:        position{line: 559, col: 8, offset: 17408},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 560, col: 1, offset: 17413},
			expr: &choiceExpr{
				pos: position{line: 560, col: 8, offset: 17420},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 560, col: 8, offset: 17420},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 8, offset: 17420},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 560, col: 11, offset: 17423},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 560, col: 17, offset: 17429},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 17, offset: 17429},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 560, col: 19, offset: 17431},
								expr: &ruleRefExpr{
									pos:  position{line: 560, col: 19, offset: 17431},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 560, col: 38, offset: 17450},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 560, col: 44, offset: 17456},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 560, col: 44, offset: 17456},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 560, col: 47, offset: 17459},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 562, col: 1, offset: 17464},
			expr: &notExpr{
				pos: position{line: 562, col: 8, offset: 17471},
				expr: &anyMatcher{
					line: 562, col: 9, offset: 17472,
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
		Type:        typ.(*Type),
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
