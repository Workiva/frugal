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

type include string

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
			pos:  position{line: 79, col: 1, offset: 2189},
			expr: &actionExpr{
				pos: position{line: 79, col: 11, offset: 2201},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 79, col: 11, offset: 2201},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 79, col: 11, offset: 2201},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 79, col: 14, offset: 2204},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 79, col: 25, offset: 2215},
								expr: &seqExpr{
									pos: position{line: 79, col: 27, offset: 2217},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 79, col: 27, offset: 2217},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 79, col: 37, offset: 2227},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 79, col: 44, offset: 2234},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 79, col: 44, offset: 2234},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 79, col: 50, offset: 2240},
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
			pos:  position{line: 152, col: 1, offset: 4849},
			expr: &actionExpr{
				pos: position{line: 152, col: 15, offset: 4865},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 152, col: 15, offset: 4865,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 156, col: 1, offset: 4923},
			expr: &actionExpr{
				pos: position{line: 156, col: 13, offset: 4937},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 156, col: 13, offset: 4937},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 156, col: 13, offset: 4937},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 156, col: 20, offset: 4944},
								expr: &seqExpr{
									pos: position{line: 156, col: 21, offset: 4945},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 156, col: 21, offset: 4945},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 156, col: 31, offset: 4955},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 156, col: 36, offset: 4960},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 156, col: 47, offset: 4971},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 156, col: 47, offset: 4971},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 156, col: 65, offset: 4989},
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
			pos:  position{line: 169, col: 1, offset: 5460},
			expr: &choiceExpr{
				pos: position{line: 169, col: 19, offset: 5480},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 169, col: 19, offset: 5480},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 29, offset: 5490},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 41, offset: 5502},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 49, offset: 5510},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 56, offset: 5517},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 66, offset: 5527},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 75, offset: 5536},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 87, offset: 5548},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 95, offset: 5556},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 171, col: 1, offset: 5565},
			expr: &actionExpr{
				pos: position{line: 171, col: 11, offset: 5577},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 171, col: 11, offset: 5577},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 171, col: 11, offset: 5577},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 21, offset: 5587},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 23, offset: 5589},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 28, offset: 5594},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 36, offset: 5602},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5650},
			expr: &actionExpr{
				pos: position{line: 175, col: 13, offset: 5664},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 13, offset: 5664},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 13, offset: 5664},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 25, offset: 5676},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 27, offset: 5678},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 33, offset: 5684},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 33, offset: 5684},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 43, offset: 5694},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 45, offset: 5696},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 48, offset: 5699},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 59, offset: 5710},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5835},
			expr: &actionExpr{
				pos: position{line: 182, col: 9, offset: 5845},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 9, offset: 5845},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 9, offset: 5845},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 17, offset: 5853},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 19, offset: 5855},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 23, offset: 5859},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 33, offset: 5869},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 35, offset: 5871},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 40, offset: 5876},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 51, offset: 5887},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 53, offset: 5889},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 57, offset: 5893},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 59, offset: 5895},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 65, offset: 5901},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 76, offset: 5912},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6044},
			expr: &actionExpr{
				pos: position{line: 190, col: 8, offset: 6053},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 8, offset: 6053},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 8, offset: 6053},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 15, offset: 6060},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 17, offset: 6062},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 22, offset: 6067},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 33, offset: 6078},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 36, offset: 6081},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 40, offset: 6085},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 43, offset: 6088},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 50, offset: 6095},
								expr: &seqExpr{
									pos: position{line: 190, col: 51, offset: 6096},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 51, offset: 6096},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 61, offset: 6106},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 66, offset: 6111},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 70, offset: 6115},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6727},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6741},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6741},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6741},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6748},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6749},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6749},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6759},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6764},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6769},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6780},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6782},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6788},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6789},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6789},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6793},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6795},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6809},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6809},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7145},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7157},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7157},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7157},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7167},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7169},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7173},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7183},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7185},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7190},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7201},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7310},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7321},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7321},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7321},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7330},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7332},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7335},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7375},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7389},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7389},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7389},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7401},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7403},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7406},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7457},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7467},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7467},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7467},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7475},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7477},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7480},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7527},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7542},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7542},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7542},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7547},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7558},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7561},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7565},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7568},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7575},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7585},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7589},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7750},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7764},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7764},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7771},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7772},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7772},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7778},
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
			pos:  position{line: 257, col: 1, offset: 7959},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 7969},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 7969},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 7969},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 7976},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 7977},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 7977},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 7987},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 7992},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 7995},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8007},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8009},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8013},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8015},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8019},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8019},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 74, offset: 8034},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 76, offset: 8036},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 80, offset: 8040},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 90, offset: 8050},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 92, offset: 8052},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 97, offset: 8057},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 108, offset: 8068},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 111, offset: 8071},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 115, offset: 8075},
								expr: &seqExpr{
									pos: position{line: 257, col: 116, offset: 8076},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 116, offset: 8076},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 120, offset: 8080},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 122, offset: 8082},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 135, offset: 8095},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 135, offset: 8095},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 279, col: 1, offset: 8557},
			expr: &actionExpr{
				pos: position{line: 279, col: 17, offset: 8575},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 279, col: 18, offset: 8576},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 18, offset: 8576},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 279, col: 31, offset: 8589},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 287, col: 1, offset: 8732},
			expr: &actionExpr{
				pos: position{line: 287, col: 11, offset: 8744},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 287, col: 11, offset: 8744},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 287, col: 11, offset: 8744},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 21, offset: 8754},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 23, offset: 8756},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 287, col: 28, offset: 8761},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 39, offset: 8772},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 41, offset: 8774},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 287, col: 49, offset: 8782},
								expr: &seqExpr{
									pos: position{line: 287, col: 50, offset: 8783},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 287, col: 50, offset: 8783},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 60, offset: 8793},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 63, offset: 8796},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 74, offset: 8807},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 79, offset: 8812},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 287, col: 82, offset: 8815},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 86, offset: 8819},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 89, offset: 8822},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 287, col: 97, offset: 8830},
								expr: &seqExpr{
									pos: position{line: 287, col: 98, offset: 8831},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 287, col: 98, offset: 8831},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 107, offset: 8840},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 287, col: 113, offset: 8846},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 287, col: 113, offset: 8846},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 287, col: 119, offset: 8852},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 138, offset: 8871},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 303, col: 1, offset: 9255},
			expr: &actionExpr{
				pos: position{line: 303, col: 21, offset: 9277},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 303, col: 21, offset: 9277,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 307, col: 1, offset: 9346},
			expr: &actionExpr{
				pos: position{line: 307, col: 12, offset: 9359},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 307, col: 12, offset: 9359},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 307, col: 12, offset: 9359},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 19, offset: 9366},
								expr: &seqExpr{
									pos: position{line: 307, col: 20, offset: 9367},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 20, offset: 9367},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 30, offset: 9377},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 35, offset: 9382},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 42, offset: 9389},
								expr: &seqExpr{
									pos: position{line: 307, col: 43, offset: 9390},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 43, offset: 9390},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 52, offset: 9399},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 57, offset: 9404},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 61, offset: 9408},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 74, offset: 9421},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 77, offset: 9424},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 82, offset: 9429},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 93, offset: 9440},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 307, col: 95, offset: 9442},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 99, offset: 9446},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 102, offset: 9449},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 112, offset: 9459},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 307, col: 122, offset: 9469},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 126, offset: 9473},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 129, offset: 9476},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 140, offset: 9487},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 140, offset: 9487},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 307, col: 148, offset: 9495},
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 148, offset: 9495},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 334, col: 1, offset: 10090},
			expr: &actionExpr{
				pos: position{line: 334, col: 16, offset: 10107},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 334, col: 16, offset: 10107},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 334, col: 21, offset: 10112},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 334, col: 21, offset: 10112},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 30, offset: 10121},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 341, col: 1, offset: 10243},
			expr: &actionExpr{
				pos: position{line: 341, col: 10, offset: 10254},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 341, col: 10, offset: 10254},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 10, offset: 10254},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 19, offset: 10263},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 341, col: 22, offset: 10266},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 26, offset: 10270},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 341, col: 29, offset: 10273},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 40, offset: 10284},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 341, col: 50, offset: 10294},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 345, col: 1, offset: 10330},
			expr: &actionExpr{
				pos: position{line: 345, col: 13, offset: 10344},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 13, offset: 10344},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 345, col: 18, offset: 10349},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 345, col: 18, offset: 10349},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 29, offset: 10360},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 45, offset: 10376},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 352, col: 1, offset: 10501},
			expr: &actionExpr{
				pos: position{line: 352, col: 12, offset: 10514},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 352, col: 13, offset: 10515},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 352, col: 13, offset: 10515},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 22, offset: 10524},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 31, offset: 10533},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 39, offset: 10541},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 47, offset: 10549},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 55, offset: 10557},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 66, offset: 10568},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 77, offset: 10579},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 356, col: 1, offset: 10639},
			expr: &actionExpr{
				pos: position{line: 356, col: 17, offset: 10657},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 17, offset: 10657},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 22, offset: 10662},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 356, col: 22, offset: 10662},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 32, offset: 10672},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 42, offset: 10682},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 360, col: 1, offset: 10717},
			expr: &actionExpr{
				pos: position{line: 360, col: 11, offset: 10729},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 360, col: 11, offset: 10729},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 360, col: 11, offset: 10729},
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 11, offset: 10729},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 360, col: 20, offset: 10738},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 27, offset: 10745},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 30, offset: 10748},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 34, offset: 10752},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 44, offset: 10762},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 47, offset: 10765},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 51, offset: 10769},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 54, offset: 10772},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 60, offset: 10778},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 70, offset: 10788},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 73, offset: 10791},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 368, col: 1, offset: 10914},
			expr: &actionExpr{
				pos: position{line: 368, col: 11, offset: 10926},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 11, offset: 10926},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 368, col: 11, offset: 10926},
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 11, offset: 10926},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 368, col: 20, offset: 10935},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 27, offset: 10942},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 30, offset: 10945},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 34, offset: 10949},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 44, offset: 10959},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 47, offset: 10962},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 375, col: 1, offset: 11053},
			expr: &actionExpr{
				pos: position{line: 375, col: 12, offset: 11066},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 12, offset: 11066},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 12, offset: 11066},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 20, offset: 11074},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 375, col: 23, offset: 11077},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 27, offset: 11081},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 37, offset: 11091},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 375, col: 40, offset: 11094},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 382, col: 1, offset: 11186},
			expr: &actionExpr{
				pos: position{line: 382, col: 11, offset: 11198},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 11, offset: 11198},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 382, col: 11, offset: 11198},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 382, col: 22, offset: 11209},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 30, offset: 11217},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 386, col: 1, offset: 11254},
			expr: &choiceExpr{
				pos: position{line: 386, col: 14, offset: 11269},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 386, col: 14, offset: 11269},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 24, offset: 11279},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 41, offset: 11296},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 55, offset: 11310},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 66, offset: 11321},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 78, offset: 11333},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 388, col: 1, offset: 11345},
			expr: &actionExpr{
				pos: position{line: 388, col: 15, offset: 11361},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 388, col: 15, offset: 11361},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 388, col: 15, offset: 11361},
							expr: &charClassMatcher{
								pos:        position{line: 388, col: 15, offset: 11361},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 388, col: 21, offset: 11367},
							expr: &ruleRefExpr{
								pos:  position{line: 388, col: 21, offset: 11367},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 392, col: 1, offset: 11431},
			expr: &actionExpr{
				pos: position{line: 392, col: 18, offset: 11450},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 392, col: 18, offset: 11450},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 18, offset: 11450},
							expr: &charClassMatcher{
								pos:        position{line: 392, col: 18, offset: 11450},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 24, offset: 11456},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 24, offset: 11456},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 31, offset: 11463},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 35, offset: 11467},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 35, offset: 11467},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 392, col: 42, offset: 11474},
							expr: &seqExpr{
								pos: position{line: 392, col: 44, offset: 11476},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 392, col: 44, offset: 11476},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 392, col: 51, offset: 11483},
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
			pos:  position{line: 396, col: 1, offset: 11553},
			expr: &actionExpr{
				pos: position{line: 396, col: 13, offset: 11567},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 396, col: 13, offset: 11567},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 396, col: 13, offset: 11567},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 17, offset: 11571},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 396, col: 20, offset: 11574},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 396, col: 27, offset: 11581},
								expr: &seqExpr{
									pos: position{line: 396, col: 28, offset: 11582},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 396, col: 28, offset: 11582},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 39, offset: 11593},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 396, col: 42, offset: 11596},
											expr: &ruleRefExpr{
												pos:  position{line: 396, col: 42, offset: 11596},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 57, offset: 11611},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 62, offset: 11616},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 396, col: 65, offset: 11619},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 405, col: 1, offset: 11813},
			expr: &actionExpr{
				pos: position{line: 405, col: 12, offset: 11826},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 405, col: 12, offset: 11826},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 405, col: 12, offset: 11826},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 405, col: 16, offset: 11830},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 405, col: 19, offset: 11833},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 405, col: 26, offset: 11840},
								expr: &seqExpr{
									pos: position{line: 405, col: 27, offset: 11841},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 405, col: 27, offset: 11841},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 38, offset: 11852},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 405, col: 41, offset: 11855},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 45, offset: 11859},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 48, offset: 11862},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 59, offset: 11873},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 405, col: 63, offset: 11877},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 405, col: 63, offset: 11877},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 405, col: 69, offset: 11883},
													expr: &litMatcher{
														pos:        position{line: 405, col: 70, offset: 11884},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 75, offset: 11889},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 405, col: 80, offset: 11894},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 425, col: 1, offset: 12444},
			expr: &choiceExpr{
				pos: position{line: 425, col: 19, offset: 12464},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 425, col: 19, offset: 12464},
						name: "Scope",
					},
					&ruleRefExpr{
						pos:  position{line: 425, col: 27, offset: 12472},
						name: "Async",
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 427, col: 1, offset: 12479},
			expr: &actionExpr{
				pos: position{line: 427, col: 9, offset: 12489},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 427, col: 9, offset: 12489},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 427, col: 9, offset: 12489},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 16, offset: 12496},
								expr: &seqExpr{
									pos: position{line: 427, col: 17, offset: 12497},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 17, offset: 12497},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 27, offset: 12507},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 427, col: 32, offset: 12512},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 40, offset: 12520},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 43, offset: 12523},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 427, col: 48, offset: 12528},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 59, offset: 12539},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 62, offset: 12542},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 69, offset: 12549},
								expr: &ruleRefExpr{
									pos:  position{line: 427, col: 69, offset: 12549},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 77, offset: 12557},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 427, col: 80, offset: 12560},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 84, offset: 12564},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 87, offset: 12567},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 427, col: 98, offset: 12578},
								expr: &seqExpr{
									pos: position{line: 427, col: 99, offset: 12579},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 99, offset: 12579},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 109, offset: 12589},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 427, col: 115, offset: 12595},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 427, col: 115, offset: 12595},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 427, col: 121, offset: 12601},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 138, offset: 12618},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 448, col: 1, offset: 13163},
			expr: &actionExpr{
				pos: position{line: 448, col: 19, offset: 13183},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 448, col: 19, offset: 13183,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 452, col: 1, offset: 13250},
			expr: &actionExpr{
				pos: position{line: 452, col: 10, offset: 13261},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 452, col: 10, offset: 13261},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 452, col: 10, offset: 13261},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 19, offset: 13270},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 22, offset: 13273},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 452, col: 34, offset: 13285},
							expr: &seqExpr{
								pos: position{line: 452, col: 35, offset: 13286},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 452, col: 35, offset: 13286},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 452, col: 39, offset: 13290},
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
			pos:  position{line: 457, col: 1, offset: 13421},
			expr: &choiceExpr{
				pos: position{line: 457, col: 15, offset: 13437},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 457, col: 16, offset: 13438},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 457, col: 16, offset: 13438},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 457, col: 20, offset: 13442},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 457, col: 31, offset: 13453},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 457, col: 38, offset: 13460},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 459, col: 1, offset: 13472},
			expr: &oneOrMoreExpr{
				pos: position{line: 459, col: 14, offset: 13487},
				expr: &charClassMatcher{
					pos:        position{line: 459, col: 14, offset: 13487},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 461, col: 1, offset: 13505},
			expr: &actionExpr{
				pos: position{line: 461, col: 13, offset: 13519},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 461, col: 13, offset: 13519},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 461, col: 13, offset: 13519},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 461, col: 20, offset: 13526},
								expr: &seqExpr{
									pos: position{line: 461, col: 21, offset: 13527},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 461, col: 21, offset: 13527},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 31, offset: 13537},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 461, col: 36, offset: 13542},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 41, offset: 13547},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 52, offset: 13558},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 461, col: 54, offset: 13560},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 58, offset: 13564},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 61, offset: 13567},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 65, offset: 13571},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 76, offset: 13582},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Async",
			pos:  position{line: 473, col: 1, offset: 13852},
			expr: &actionExpr{
				pos: position{line: 473, col: 9, offset: 13862},
				run: (*parser).callonAsync1,
				expr: &seqExpr{
					pos: position{line: 473, col: 9, offset: 13862},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 473, col: 9, offset: 13862},
							val:        "async",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 17, offset: 13870},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 19, offset: 13872},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 473, col: 24, offset: 13877},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 35, offset: 13888},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 37, offset: 13890},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 473, col: 45, offset: 13898},
								expr: &seqExpr{
									pos: position{line: 473, col: 46, offset: 13899},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 473, col: 46, offset: 13899},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 56, offset: 13909},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 59, offset: 13912},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 70, offset: 13923},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 75, offset: 13928},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 473, col: 78, offset: 13931},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 82, offset: 13935},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 473, col: 85, offset: 13938},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 473, col: 93, offset: 13946},
								expr: &seqExpr{
									pos: position{line: 473, col: 94, offset: 13947},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 473, col: 94, offset: 13947},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 473, col: 103, offset: 13956},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 473, col: 109, offset: 13962},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 473, col: 109, offset: 13962},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 473, col: 115, offset: 13968},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 473, col: 134, offset: 13987},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 493, col: 1, offset: 14618},
			expr: &actionExpr{
				pos: position{line: 493, col: 11, offset: 14630},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 493, col: 12, offset: 14631},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 493, col: 13, offset: 14632},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 493, col: 13, offset: 14632},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 493, col: 17, offset: 14636},
									expr: &choiceExpr{
										pos: position{line: 493, col: 18, offset: 14637},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 493, col: 18, offset: 14637},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 493, col: 25, offset: 14644},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 493, col: 32, offset: 14651},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 493, col: 40, offset: 14659},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 493, col: 40, offset: 14659},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 493, col: 45, offset: 14664},
									expr: &choiceExpr{
										pos: position{line: 493, col: 46, offset: 14665},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 493, col: 46, offset: 14665},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 493, col: 53, offset: 14672},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 493, col: 60, offset: 14679},
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
			pos:  position{line: 500, col: 1, offset: 14895},
			expr: &actionExpr{
				pos: position{line: 500, col: 14, offset: 14910},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 500, col: 14, offset: 14910},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 500, col: 14, offset: 14910},
							expr: &choiceExpr{
								pos: position{line: 500, col: 15, offset: 14911},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 500, col: 15, offset: 14911},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 500, col: 24, offset: 14920},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 500, col: 30, offset: 14926},
							expr: &choiceExpr{
								pos: position{line: 500, col: 31, offset: 14927},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 500, col: 31, offset: 14927},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 500, col: 40, offset: 14936},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 500, col: 48, offset: 14944},
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
			pos:  position{line: 504, col: 1, offset: 14999},
			expr: &charClassMatcher{
				pos:        position{line: 504, col: 17, offset: 15017},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 505, col: 1, offset: 15022},
			expr: &charClassMatcher{
				pos:        position{line: 505, col: 10, offset: 15033},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 506, col: 1, offset: 15042},
			expr: &charClassMatcher{
				pos:        position{line: 506, col: 9, offset: 15052},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 508, col: 1, offset: 15059},
			expr: &anyMatcher{
				line: 508, col: 14, offset: 15074,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 509, col: 1, offset: 15076},
			expr: &actionExpr{
				pos: position{line: 509, col: 13, offset: 15090},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 509, col: 13, offset: 15090},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 509, col: 13, offset: 15090},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 509, col: 20, offset: 15097},
							expr: &seqExpr{
								pos: position{line: 509, col: 22, offset: 15099},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 509, col: 22, offset: 15099},
										expr: &litMatcher{
											pos:        position{line: 509, col: 23, offset: 15100},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 509, col: 28, offset: 15105},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 509, col: 42, offset: 15119},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 515, col: 1, offset: 15299},
			expr: &choiceExpr{
				pos: position{line: 515, col: 11, offset: 15311},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 515, col: 11, offset: 15311},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 515, col: 30, offset: 15330},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 516, col: 1, offset: 15348},
			expr: &seqExpr{
				pos: position{line: 516, col: 20, offset: 15369},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 516, col: 20, offset: 15369},
						expr: &ruleRefExpr{
							pos:  position{line: 516, col: 21, offset: 15370},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 516, col: 31, offset: 15380},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 516, col: 36, offset: 15385},
						expr: &seqExpr{
							pos: position{line: 516, col: 38, offset: 15387},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 516, col: 38, offset: 15387},
									expr: &litMatcher{
										pos:        position{line: 516, col: 39, offset: 15388},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 516, col: 44, offset: 15393},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 516, col: 58, offset: 15407},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 517, col: 1, offset: 15412},
			expr: &seqExpr{
				pos: position{line: 517, col: 36, offset: 15449},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 517, col: 36, offset: 15449},
						expr: &ruleRefExpr{
							pos:  position{line: 517, col: 37, offset: 15450},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 517, col: 47, offset: 15460},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 517, col: 52, offset: 15465},
						expr: &seqExpr{
							pos: position{line: 517, col: 54, offset: 15467},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 517, col: 54, offset: 15467},
									expr: &choiceExpr{
										pos: position{line: 517, col: 57, offset: 15470},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 517, col: 57, offset: 15470},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 517, col: 64, offset: 15477},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 517, col: 70, offset: 15483},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 517, col: 84, offset: 15497},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 518, col: 1, offset: 15502},
			expr: &choiceExpr{
				pos: position{line: 518, col: 21, offset: 15524},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 518, col: 22, offset: 15525},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 518, col: 22, offset: 15525},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 518, col: 27, offset: 15530},
								expr: &seqExpr{
									pos: position{line: 518, col: 29, offset: 15532},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 518, col: 29, offset: 15532},
											expr: &ruleRefExpr{
												pos:  position{line: 518, col: 30, offset: 15533},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 518, col: 34, offset: 15537},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 518, col: 52, offset: 15555},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 518, col: 52, offset: 15555},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 518, col: 56, offset: 15559},
								expr: &seqExpr{
									pos: position{line: 518, col: 58, offset: 15561},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 518, col: 58, offset: 15561},
											expr: &ruleRefExpr{
												pos:  position{line: 518, col: 59, offset: 15562},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 518, col: 63, offset: 15566},
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
			pos:  position{line: 520, col: 1, offset: 15582},
			expr: &zeroOrMoreExpr{
				pos: position{line: 520, col: 6, offset: 15589},
				expr: &choiceExpr{
					pos: position{line: 520, col: 8, offset: 15591},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 520, col: 8, offset: 15591},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 520, col: 21, offset: 15604},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 520, col: 27, offset: 15610},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 521, col: 1, offset: 15621},
			expr: &zeroOrMoreExpr{
				pos: position{line: 521, col: 5, offset: 15627},
				expr: &choiceExpr{
					pos: position{line: 521, col: 7, offset: 15629},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 521, col: 7, offset: 15629},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 521, col: 20, offset: 15642},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 522, col: 1, offset: 15678},
			expr: &zeroOrMoreExpr{
				pos: position{line: 522, col: 6, offset: 15685},
				expr: &ruleRefExpr{
					pos:  position{line: 522, col: 6, offset: 15685},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 524, col: 1, offset: 15698},
			expr: &charClassMatcher{
				pos:        position{line: 524, col: 14, offset: 15713},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 525, col: 1, offset: 15721},
			expr: &litMatcher{
				pos:        position{line: 525, col: 7, offset: 15729},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 526, col: 1, offset: 15734},
			expr: &choiceExpr{
				pos: position{line: 526, col: 7, offset: 15742},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 526, col: 7, offset: 15742},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 526, col: 7, offset: 15742},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 526, col: 10, offset: 15745},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 526, col: 16, offset: 15751},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 526, col: 16, offset: 15751},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 526, col: 18, offset: 15753},
								expr: &ruleRefExpr{
									pos:  position{line: 526, col: 18, offset: 15753},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 526, col: 37, offset: 15772},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 526, col: 43, offset: 15778},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 526, col: 43, offset: 15778},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 526, col: 46, offset: 15781},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 528, col: 1, offset: 15786},
			expr: &notExpr{
				pos: position{line: 528, col: 7, offset: 15794},
				expr: &anyMatcher{
					line: 528, col: 8, offset: 15795,
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
	frugal := &Frugal{
		Thrift:         thrift,
		Scopes:         []*Scope{},
		Asyncs:         []*Async{},
		ParsedIncludes: make(map[string]*Frugal),
	}

	stmts := toIfaceSlice(statements)
	for _, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
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
			v.Comment = wrapper.comment
			thrift.Structs = append(thrift.Structs, v)
		case exception:
			strct := (*Struct)(v)
			strct.Comment = wrapper.comment
			thrift.Exceptions = append(thrift.Exceptions, strct)
		case union:
			strct := unionToStruct(v)
			strct.Comment = wrapper.comment
			thrift.Unions = append(thrift.Unions, strct)
		case *Service:
			v.Comment = wrapper.comment
			thrift.Services = append(thrift.Services, v)
		case include:
			name := string(v)
			if ix := strings.LastIndex(name, "."); ix > 0 {
				name = name[:ix]
			}
			thrift.Includes = append(thrift.Includes, &Include{Name: name, Value: string(v)})
		case *Scope:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Scopes = append(frugal.Scopes, v)
		case *Async:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Asyncs = append(frugal.Asyncs, v)
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
	return include(file.(string)), nil
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

func (c *current) onAsync1(name, extends, methods interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	async := &Async{
		Name:    string(name.(Identifier)),
		Methods: make([]*Method, len(ms)),
	}
	if extends != nil {
		async.Extends = string(extends.([]interface{})[2].(Identifier))
	}
	for i, m := range ms {
		mt := m.([]interface{})[0].(*Method)
		async.Methods[i] = mt
	}
	return async, nil
}

func (p *parser) callonAsync1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAsync1(stack["name"], stack["extends"], stack["methods"])
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
