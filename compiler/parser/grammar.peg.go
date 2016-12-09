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
			pos:  position{line: 188, col: 1, offset: 6130},
			expr: &actionExpr{
				pos: position{line: 188, col: 14, offset: 6143},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 188, col: 14, offset: 6143},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 188, col: 14, offset: 6143},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 26, offset: 6155},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 28, offset: 6157},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 188, col: 34, offset: 6163},
								expr: &charClassMatcher{
									pos:        position{line: 188, col: 34, offset: 6163},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 44, offset: 6173},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 46, offset: 6175},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 188, col: 49, offset: 6178},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 60, offset: 6189},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 188, col: 62, offset: 6191},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 188, col: 74, offset: 6203},
								expr: &ruleRefExpr{
									pos:  position{line: 188, col: 74, offset: 6203},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 188, col: 91, offset: 6220},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 196, col: 1, offset: 6406},
			expr: &actionExpr{
				pos: position{line: 196, col: 10, offset: 6415},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 196, col: 10, offset: 6415},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 196, col: 10, offset: 6415},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 18, offset: 6423},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 20, offset: 6425},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 24, offset: 6429},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 34, offset: 6439},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 36, offset: 6441},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 41, offset: 6446},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 52, offset: 6457},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 196, col: 54, offset: 6459},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 58, offset: 6463},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 60, offset: 6465},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 66, offset: 6471},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 77, offset: 6482},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 79, offset: 6484},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 196, col: 91, offset: 6496},
								expr: &ruleRefExpr{
									pos:  position{line: 196, col: 91, offset: 6496},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 108, offset: 6513},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 205, col: 1, offset: 6707},
			expr: &actionExpr{
				pos: position{line: 205, col: 9, offset: 6715},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 205, col: 9, offset: 6715},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 205, col: 9, offset: 6715},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 16, offset: 6722},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 18, offset: 6724},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 205, col: 23, offset: 6729},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 34, offset: 6740},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 205, col: 37, offset: 6743},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 41, offset: 6747},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 44, offset: 6750},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 205, col: 51, offset: 6757},
								expr: &seqExpr{
									pos: position{line: 205, col: 52, offset: 6758},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 205, col: 52, offset: 6758},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 205, col: 62, offset: 6768},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 205, col: 67, offset: 6773},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 71, offset: 6777},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 205, col: 73, offset: 6779},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 205, col: 85, offset: 6791},
								expr: &ruleRefExpr{
									pos:  position{line: 205, col: 85, offset: 6791},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 102, offset: 6808},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 229, col: 1, offset: 7470},
			expr: &actionExpr{
				pos: position{line: 229, col: 14, offset: 7483},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 229, col: 14, offset: 7483},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 229, col: 14, offset: 7483},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 21, offset: 7490},
								expr: &seqExpr{
									pos: position{line: 229, col: 22, offset: 7491},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 229, col: 22, offset: 7491},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 32, offset: 7501},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 229, col: 37, offset: 7506},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 42, offset: 7511},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 53, offset: 7522},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 55, offset: 7524},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 61, offset: 7530},
								expr: &seqExpr{
									pos: position{line: 229, col: 62, offset: 7531},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 229, col: 62, offset: 7531},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 66, offset: 7535},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 229, col: 68, offset: 7537},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 82, offset: 7551},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 84, offset: 7553},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 229, col: 96, offset: 7565},
								expr: &ruleRefExpr{
									pos:  position{line: 229, col: 96, offset: 7565},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 229, col: 113, offset: 7582},
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 113, offset: 7582},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 245, col: 1, offset: 7980},
			expr: &actionExpr{
				pos: position{line: 245, col: 12, offset: 7991},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 245, col: 12, offset: 7991},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 245, col: 12, offset: 7991},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 22, offset: 8001},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 24, offset: 8003},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 245, col: 28, offset: 8007},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 38, offset: 8017},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 40, offset: 8019},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 245, col: 45, offset: 8024},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 56, offset: 8035},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 245, col: 58, offset: 8037},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 245, col: 70, offset: 8049},
								expr: &ruleRefExpr{
									pos:  position{line: 245, col: 70, offset: 8049},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 245, col: 87, offset: 8066},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 253, col: 1, offset: 8238},
			expr: &actionExpr{
				pos: position{line: 253, col: 11, offset: 8248},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 253, col: 11, offset: 8248},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 253, col: 11, offset: 8248},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 20, offset: 8257},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 22, offset: 8259},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 25, offset: 8262},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 254, col: 1, offset: 8302},
			expr: &actionExpr{
				pos: position{line: 254, col: 14, offset: 8315},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 254, col: 14, offset: 8315},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 254, col: 14, offset: 8315},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 26, offset: 8327},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 28, offset: 8329},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 31, offset: 8332},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 255, col: 1, offset: 8383},
			expr: &actionExpr{
				pos: position{line: 255, col: 10, offset: 8392},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 255, col: 10, offset: 8392},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 255, col: 10, offset: 8392},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 255, col: 18, offset: 8400},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 255, col: 20, offset: 8402},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 255, col: 23, offset: 8405},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 256, col: 1, offset: 8452},
			expr: &actionExpr{
				pos: position{line: 256, col: 15, offset: 8466},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 256, col: 15, offset: 8466},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 256, col: 15, offset: 8466},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 20, offset: 8471},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 31, offset: 8482},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 256, col: 34, offset: 8485},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 38, offset: 8489},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 41, offset: 8492},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 48, offset: 8499},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 256, col: 58, offset: 8509},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 62, offset: 8513},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 64, offset: 8515},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 76, offset: 8527},
								expr: &ruleRefExpr{
									pos:  position{line: 256, col: 76, offset: 8527},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 93, offset: 8544},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 267, col: 1, offset: 8761},
			expr: &actionExpr{
				pos: position{line: 267, col: 14, offset: 8774},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 267, col: 14, offset: 8774},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 267, col: 21, offset: 8781},
						expr: &seqExpr{
							pos: position{line: 267, col: 22, offset: 8782},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 267, col: 22, offset: 8782},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 267, col: 28, offset: 8788},
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
			pos:  position{line: 276, col: 1, offset: 8969},
			expr: &actionExpr{
				pos: position{line: 276, col: 10, offset: 8978},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 276, col: 10, offset: 8978},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 276, col: 10, offset: 8978},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 17, offset: 8985},
								expr: &seqExpr{
									pos: position{line: 276, col: 18, offset: 8986},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 276, col: 18, offset: 8986},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 28, offset: 8996},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 276, col: 33, offset: 9001},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 36, offset: 9004},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 48, offset: 9016},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 276, col: 50, offset: 9018},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 54, offset: 9022},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 56, offset: 9024},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 60, offset: 9028},
								expr: &ruleRefExpr{
									pos:  position{line: 276, col: 60, offset: 9028},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 75, offset: 9043},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 77, offset: 9045},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 81, offset: 9049},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 91, offset: 9059},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 93, offset: 9061},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 98, offset: 9066},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 109, offset: 9077},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 112, offset: 9080},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 116, offset: 9084},
								expr: &seqExpr{
									pos: position{line: 276, col: 117, offset: 9085},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 276, col: 117, offset: 9085},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 121, offset: 9089},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 276, col: 123, offset: 9091},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 276, col: 136, offset: 9104},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 276, col: 138, offset: 9106},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 276, col: 150, offset: 9118},
								expr: &ruleRefExpr{
									pos:  position{line: 276, col: 150, offset: 9118},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 276, col: 167, offset: 9135},
							expr: &ruleRefExpr{
								pos:  position{line: 276, col: 167, offset: 9135},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 299, col: 1, offset: 9667},
			expr: &actionExpr{
				pos: position{line: 299, col: 18, offset: 9684},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 299, col: 19, offset: 9685},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 299, col: 19, offset: 9685},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 299, col: 32, offset: 9698},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 307, col: 1, offset: 9841},
			expr: &actionExpr{
				pos: position{line: 307, col: 12, offset: 9852},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 307, col: 12, offset: 9852},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 307, col: 12, offset: 9852},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 22, offset: 9862},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 24, offset: 9864},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 29, offset: 9869},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 40, offset: 9880},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 42, offset: 9882},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 50, offset: 9890},
								expr: &seqExpr{
									pos: position{line: 307, col: 51, offset: 9891},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 51, offset: 9891},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 61, offset: 9901},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 64, offset: 9904},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 75, offset: 9915},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 80, offset: 9920},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 307, col: 83, offset: 9923},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 87, offset: 9927},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 90, offset: 9930},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 307, col: 98, offset: 9938},
								expr: &seqExpr{
									pos: position{line: 307, col: 99, offset: 9939},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 99, offset: 9939},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 108, offset: 9948},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 307, col: 114, offset: 9954},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 307, col: 114, offset: 9954},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 307, col: 120, offset: 9960},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 139, offset: 9979},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 141, offset: 9981},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 153, offset: 9993},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 153, offset: 9993},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 170, offset: 10010},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 324, col: 1, offset: 10451},
			expr: &actionExpr{
				pos: position{line: 324, col: 22, offset: 10472},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 324, col: 22, offset: 10472,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 328, col: 1, offset: 10541},
			expr: &actionExpr{
				pos: position{line: 328, col: 13, offset: 10553},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 328, col: 13, offset: 10553},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 328, col: 13, offset: 10553},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 20, offset: 10560},
								expr: &seqExpr{
									pos: position{line: 328, col: 21, offset: 10561},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 328, col: 21, offset: 10561},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 328, col: 31, offset: 10571},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 328, col: 36, offset: 10576},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 43, offset: 10583},
								expr: &seqExpr{
									pos: position{line: 328, col: 44, offset: 10584},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 328, col: 44, offset: 10584},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 328, col: 53, offset: 10593},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 328, col: 58, offset: 10598},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 62, offset: 10602},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 75, offset: 10615},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 78, offset: 10618},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 83, offset: 10623},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 94, offset: 10634},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 328, col: 96, offset: 10636},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 100, offset: 10640},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 103, offset: 10643},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 113, offset: 10653},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 328, col: 123, offset: 10663},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 127, offset: 10667},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 130, offset: 10670},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 141, offset: 10681},
								expr: &ruleRefExpr{
									pos:  position{line: 328, col: 141, offset: 10681},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 149, offset: 10689},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 151, offset: 10691},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 328, col: 163, offset: 10703},
								expr: &ruleRefExpr{
									pos:  position{line: 328, col: 163, offset: 10703},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 328, col: 180, offset: 10720},
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 180, offset: 10720},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 356, col: 1, offset: 11371},
			expr: &actionExpr{
				pos: position{line: 356, col: 17, offset: 11387},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 17, offset: 11387},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 22, offset: 11392},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 356, col: 22, offset: 11392},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 31, offset: 11401},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 363, col: 1, offset: 11523},
			expr: &actionExpr{
				pos: position{line: 363, col: 11, offset: 11533},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 363, col: 11, offset: 11533},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 363, col: 11, offset: 11533},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 20, offset: 11542},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 363, col: 23, offset: 11545},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 27, offset: 11549},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 363, col: 30, offset: 11552},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 41, offset: 11563},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 363, col: 51, offset: 11573},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 367, col: 1, offset: 11609},
			expr: &actionExpr{
				pos: position{line: 367, col: 14, offset: 11622},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 367, col: 14, offset: 11622},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 367, col: 19, offset: 11627},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 367, col: 19, offset: 11627},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 367, col: 30, offset: 11638},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 367, col: 46, offset: 11654},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 374, col: 1, offset: 11779},
			expr: &actionExpr{
				pos: position{line: 374, col: 13, offset: 11791},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 374, col: 13, offset: 11791},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 374, col: 13, offset: 11791},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 374, col: 18, offset: 11796},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 31, offset: 11809},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 374, col: 33, offset: 11811},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 374, col: 45, offset: 11823},
								expr: &ruleRefExpr{
									pos:  position{line: 374, col: 45, offset: 11823},
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
			pos:  position{line: 381, col: 1, offset: 11959},
			expr: &actionExpr{
				pos: position{line: 381, col: 17, offset: 11975},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 381, col: 18, offset: 11976},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 381, col: 18, offset: 11976},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 27, offset: 11985},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 36, offset: 11994},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 44, offset: 12002},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 52, offset: 12010},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 60, offset: 12018},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 71, offset: 12029},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 381, col: 82, offset: 12040},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 385, col: 1, offset: 12087},
			expr: &actionExpr{
				pos: position{line: 385, col: 18, offset: 12104},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 385, col: 18, offset: 12104},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 385, col: 23, offset: 12109},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 385, col: 23, offset: 12109},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 385, col: 33, offset: 12119},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 385, col: 43, offset: 12129},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 389, col: 1, offset: 12164},
			expr: &actionExpr{
				pos: position{line: 389, col: 12, offset: 12175},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 389, col: 12, offset: 12175},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 389, col: 12, offset: 12175},
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 12, offset: 12175},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 389, col: 21, offset: 12184},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 28, offset: 12191},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 31, offset: 12194},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 35, offset: 12198},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 45, offset: 12208},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 389, col: 48, offset: 12211},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 52, offset: 12215},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 55, offset: 12218},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 389, col: 61, offset: 12224},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 71, offset: 12234},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 389, col: 74, offset: 12237},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 78, offset: 12241},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 80, offset: 12243},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 389, col: 92, offset: 12255},
								expr: &ruleRefExpr{
									pos:  position{line: 389, col: 92, offset: 12255},
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
			pos:  position{line: 398, col: 1, offset: 12453},
			expr: &actionExpr{
				pos: position{line: 398, col: 12, offset: 12464},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 398, col: 12, offset: 12464},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 398, col: 12, offset: 12464},
							expr: &ruleRefExpr{
								pos:  position{line: 398, col: 12, offset: 12464},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 398, col: 21, offset: 12473},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 28, offset: 12480},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 31, offset: 12483},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 398, col: 35, offset: 12487},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 45, offset: 12497},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 398, col: 48, offset: 12500},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 52, offset: 12504},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 54, offset: 12506},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 398, col: 66, offset: 12518},
								expr: &ruleRefExpr{
									pos:  position{line: 398, col: 66, offset: 12518},
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
			pos:  position{line: 406, col: 1, offset: 12680},
			expr: &actionExpr{
				pos: position{line: 406, col: 13, offset: 12692},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 406, col: 13, offset: 12692},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 406, col: 13, offset: 12692},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 21, offset: 12700},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 406, col: 24, offset: 12703},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 406, col: 28, offset: 12707},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 38, offset: 12717},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 406, col: 41, offset: 12720},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 406, col: 45, offset: 12724},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 406, col: 47, offset: 12726},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 406, col: 59, offset: 12738},
								expr: &ruleRefExpr{
									pos:  position{line: 406, col: 59, offset: 12738},
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
			pos:  position{line: 414, col: 1, offset: 12901},
			expr: &actionExpr{
				pos: position{line: 414, col: 12, offset: 12912},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 414, col: 12, offset: 12912},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 414, col: 12, offset: 12912},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 414, col: 23, offset: 12923},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 414, col: 31, offset: 12931},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 418, col: 1, offset: 12968},
			expr: &choiceExpr{
				pos: position{line: 418, col: 15, offset: 12982},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 418, col: 15, offset: 12982},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 25, offset: 12992},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 40, offset: 13007},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 57, offset: 13024},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 71, offset: 13038},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 82, offset: 13049},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 94, offset: 13061},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 420, col: 1, offset: 13073},
			expr: &actionExpr{
				pos: position{line: 420, col: 20, offset: 13092},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 420, col: 20, offset: 13092},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 420, col: 20, offset: 13092},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 24, offset: 13096},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 27, offset: 13099},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 420, col: 39, offset: 13111},
								expr: &ruleRefExpr{
									pos:  position{line: 420, col: 39, offset: 13111},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 420, col: 55, offset: 13127},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 428, col: 1, offset: 13291},
			expr: &actionExpr{
				pos: position{line: 428, col: 19, offset: 13309},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 428, col: 19, offset: 13309},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 428, col: 19, offset: 13309},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 428, col: 24, offset: 13314},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 428, col: 35, offset: 13325},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 428, col: 37, offset: 13327},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 428, col: 43, offset: 13333},
								expr: &actionExpr{
									pos: position{line: 428, col: 44, offset: 13334},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 428, col: 44, offset: 13334},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 428, col: 44, offset: 13334},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 428, col: 48, offset: 13338},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 428, col: 51, offset: 13341},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 428, col: 57, offset: 13347},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 428, col: 89, offset: 13379},
							expr: &ruleRefExpr{
								pos:  position{line: 428, col: 89, offset: 13379},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 428, col: 104, offset: 13394},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 439, col: 1, offset: 13590},
			expr: &actionExpr{
				pos: position{line: 439, col: 17, offset: 13606},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 439, col: 18, offset: 13607},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 439, col: 18, offset: 13607},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 439, col: 27, offset: 13616},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 443, col: 1, offset: 13671},
			expr: &actionExpr{
				pos: position{line: 443, col: 16, offset: 13686},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 443, col: 16, offset: 13686},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 443, col: 16, offset: 13686},
							expr: &charClassMatcher{
								pos:        position{line: 443, col: 16, offset: 13686},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 443, col: 22, offset: 13692},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 22, offset: 13692},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 447, col: 1, offset: 13756},
			expr: &actionExpr{
				pos: position{line: 447, col: 19, offset: 13774},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 447, col: 19, offset: 13774},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 447, col: 19, offset: 13774},
							expr: &charClassMatcher{
								pos:        position{line: 447, col: 19, offset: 13774},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 447, col: 25, offset: 13780},
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 25, offset: 13780},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 447, col: 32, offset: 13787},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 447, col: 36, offset: 13791},
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 36, offset: 13791},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 447, col: 43, offset: 13798},
							expr: &seqExpr{
								pos: position{line: 447, col: 45, offset: 13800},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 447, col: 45, offset: 13800},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 447, col: 52, offset: 13807},
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
			pos:  position{line: 451, col: 1, offset: 13877},
			expr: &actionExpr{
				pos: position{line: 451, col: 14, offset: 13890},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 451, col: 14, offset: 13890},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 451, col: 14, offset: 13890},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 18, offset: 13894},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 451, col: 21, offset: 13897},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 451, col: 28, offset: 13904},
								expr: &seqExpr{
									pos: position{line: 451, col: 29, offset: 13905},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 451, col: 29, offset: 13905},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 451, col: 40, offset: 13916},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 451, col: 43, offset: 13919},
											expr: &ruleRefExpr{
												pos:  position{line: 451, col: 43, offset: 13919},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 451, col: 58, offset: 13934},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 63, offset: 13939},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 451, col: 66, offset: 13942},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 460, col: 1, offset: 14136},
			expr: &actionExpr{
				pos: position{line: 460, col: 13, offset: 14148},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 460, col: 13, offset: 14148},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 460, col: 13, offset: 14148},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 460, col: 17, offset: 14152},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 460, col: 20, offset: 14155},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 460, col: 27, offset: 14162},
								expr: &seqExpr{
									pos: position{line: 460, col: 28, offset: 14163},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 460, col: 28, offset: 14163},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 39, offset: 14174},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 460, col: 42, offset: 14177},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 46, offset: 14181},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 49, offset: 14184},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 60, offset: 14195},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 460, col: 64, offset: 14199},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 460, col: 64, offset: 14199},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 460, col: 70, offset: 14205},
													expr: &litMatcher{
														pos:        position{line: 460, col: 71, offset: 14206},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 460, col: 76, offset: 14211},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 460, col: 81, offset: 14216},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 480, col: 1, offset: 14766},
			expr: &ruleRefExpr{
				pos:  position{line: 480, col: 20, offset: 14785},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 482, col: 1, offset: 14792},
			expr: &actionExpr{
				pos: position{line: 482, col: 10, offset: 14801},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 482, col: 10, offset: 14801},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 482, col: 10, offset: 14801},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 17, offset: 14808},
								expr: &seqExpr{
									pos: position{line: 482, col: 18, offset: 14809},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 482, col: 18, offset: 14809},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 482, col: 28, offset: 14819},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 482, col: 33, offset: 14824},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 41, offset: 14832},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 44, offset: 14835},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 482, col: 49, offset: 14840},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 60, offset: 14851},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 63, offset: 14854},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 70, offset: 14861},
								expr: &ruleRefExpr{
									pos:  position{line: 482, col: 70, offset: 14861},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 78, offset: 14869},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 482, col: 81, offset: 14872},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 85, offset: 14876},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 88, offset: 14879},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 482, col: 99, offset: 14890},
								expr: &seqExpr{
									pos: position{line: 482, col: 100, offset: 14891},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 482, col: 100, offset: 14891},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 482, col: 110, offset: 14901},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 482, col: 116, offset: 14907},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 482, col: 116, offset: 14907},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 482, col: 122, offset: 14913},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 139, offset: 14930},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 482, col: 141, offset: 14932},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 482, col: 153, offset: 14944},
								expr: &ruleRefExpr{
									pos:  position{line: 482, col: 153, offset: 14944},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 482, col: 170, offset: 14961},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 504, col: 1, offset: 15558},
			expr: &actionExpr{
				pos: position{line: 504, col: 20, offset: 15577},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 504, col: 20, offset: 15577,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 508, col: 1, offset: 15644},
			expr: &actionExpr{
				pos: position{line: 508, col: 11, offset: 15654},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 508, col: 11, offset: 15654},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 508, col: 11, offset: 15654},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 20, offset: 15663},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 23, offset: 15666},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 508, col: 35, offset: 15678},
							expr: &seqExpr{
								pos: position{line: 508, col: 36, offset: 15679},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 508, col: 36, offset: 15679},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 508, col: 40, offset: 15683},
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
			pos:  position{line: 513, col: 1, offset: 15814},
			expr: &choiceExpr{
				pos: position{line: 513, col: 16, offset: 15829},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 513, col: 17, offset: 15830},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 513, col: 17, offset: 15830},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 513, col: 21, offset: 15834},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 513, col: 32, offset: 15845},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 513, col: 39, offset: 15852},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 515, col: 1, offset: 15864},
			expr: &oneOrMoreExpr{
				pos: position{line: 515, col: 15, offset: 15878},
				expr: &charClassMatcher{
					pos:        position{line: 515, col: 15, offset: 15878},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 517, col: 1, offset: 15896},
			expr: &actionExpr{
				pos: position{line: 517, col: 14, offset: 15909},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 517, col: 14, offset: 15909},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 517, col: 14, offset: 15909},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 517, col: 21, offset: 15916},
								expr: &seqExpr{
									pos: position{line: 517, col: 22, offset: 15917},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 517, col: 22, offset: 15917},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 517, col: 32, offset: 15927},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 517, col: 37, offset: 15932},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 42, offset: 15937},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 53, offset: 15948},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 517, col: 55, offset: 15950},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 59, offset: 15954},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 517, col: 62, offset: 15957},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 66, offset: 15961},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 517, col: 77, offset: 15972},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 517, col: 79, offset: 15974},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 517, col: 91, offset: 15986},
								expr: &ruleRefExpr{
									pos:  position{line: 517, col: 91, offset: 15986},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 517, col: 108, offset: 16003},
							expr: &ruleRefExpr{
								pos:  position{line: 517, col: 108, offset: 16003},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 534, col: 1, offset: 16589},
			expr: &actionExpr{
				pos: position{line: 534, col: 12, offset: 16600},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 534, col: 13, offset: 16601},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 534, col: 14, offset: 16602},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 534, col: 14, offset: 16602},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 534, col: 18, offset: 16606},
									expr: &choiceExpr{
										pos: position{line: 534, col: 19, offset: 16607},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 534, col: 19, offset: 16607},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 534, col: 26, offset: 16614},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 534, col: 33, offset: 16621},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 534, col: 41, offset: 16629},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 534, col: 41, offset: 16629},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 534, col: 46, offset: 16634},
									expr: &choiceExpr{
										pos: position{line: 534, col: 47, offset: 16635},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 534, col: 47, offset: 16635},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 534, col: 54, offset: 16642},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 534, col: 61, offset: 16649},
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
			pos:  position{line: 543, col: 1, offset: 16935},
			expr: &actionExpr{
				pos: position{line: 543, col: 15, offset: 16949},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 543, col: 15, offset: 16949},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 543, col: 15, offset: 16949},
							expr: &choiceExpr{
								pos: position{line: 543, col: 16, offset: 16950},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 543, col: 16, offset: 16950},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 543, col: 25, offset: 16959},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 543, col: 31, offset: 16965},
							expr: &choiceExpr{
								pos: position{line: 543, col: 32, offset: 16966},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 543, col: 32, offset: 16966},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 543, col: 41, offset: 16975},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 543, col: 49, offset: 16983},
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
			pos:  position{line: 547, col: 1, offset: 17038},
			expr: &charClassMatcher{
				pos:        position{line: 547, col: 18, offset: 17055},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 548, col: 1, offset: 17060},
			expr: &charClassMatcher{
				pos:        position{line: 548, col: 11, offset: 17070},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 549, col: 1, offset: 17079},
			expr: &charClassMatcher{
				pos:        position{line: 549, col: 10, offset: 17088},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 551, col: 1, offset: 17095},
			expr: &anyMatcher{
				line: 551, col: 15, offset: 17109,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 552, col: 1, offset: 17111},
			expr: &actionExpr{
				pos: position{line: 552, col: 14, offset: 17124},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 552, col: 14, offset: 17124},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 552, col: 14, offset: 17124},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 552, col: 21, offset: 17131},
							expr: &seqExpr{
								pos: position{line: 552, col: 23, offset: 17133},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 552, col: 23, offset: 17133},
										expr: &litMatcher{
											pos:        position{line: 552, col: 24, offset: 17134},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 552, col: 29, offset: 17139},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 552, col: 43, offset: 17153},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 558, col: 1, offset: 17333},
			expr: &choiceExpr{
				pos: position{line: 558, col: 12, offset: 17344},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 558, col: 12, offset: 17344},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 558, col: 31, offset: 17363},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 559, col: 1, offset: 17381},
			expr: &seqExpr{
				pos: position{line: 559, col: 21, offset: 17401},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 559, col: 21, offset: 17401},
						expr: &ruleRefExpr{
							pos:  position{line: 559, col: 22, offset: 17402},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 559, col: 32, offset: 17412},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 559, col: 37, offset: 17417},
						expr: &seqExpr{
							pos: position{line: 559, col: 39, offset: 17419},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 559, col: 39, offset: 17419},
									expr: &litMatcher{
										pos:        position{line: 559, col: 40, offset: 17420},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 559, col: 45, offset: 17425},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 559, col: 59, offset: 17439},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 560, col: 1, offset: 17444},
			expr: &seqExpr{
				pos: position{line: 560, col: 37, offset: 17480},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 560, col: 37, offset: 17480},
						expr: &ruleRefExpr{
							pos:  position{line: 560, col: 38, offset: 17481},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 560, col: 48, offset: 17491},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 560, col: 53, offset: 17496},
						expr: &seqExpr{
							pos: position{line: 560, col: 55, offset: 17498},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 560, col: 55, offset: 17498},
									expr: &choiceExpr{
										pos: position{line: 560, col: 58, offset: 17501},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 560, col: 58, offset: 17501},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 560, col: 65, offset: 17508},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 560, col: 71, offset: 17514},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 560, col: 85, offset: 17528},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 561, col: 1, offset: 17533},
			expr: &choiceExpr{
				pos: position{line: 561, col: 22, offset: 17554},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 561, col: 23, offset: 17555},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 561, col: 23, offset: 17555},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 561, col: 28, offset: 17560},
								expr: &seqExpr{
									pos: position{line: 561, col: 30, offset: 17562},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 561, col: 30, offset: 17562},
											expr: &ruleRefExpr{
												pos:  position{line: 561, col: 31, offset: 17563},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 561, col: 35, offset: 17567},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 561, col: 53, offset: 17585},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 561, col: 53, offset: 17585},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 561, col: 57, offset: 17589},
								expr: &seqExpr{
									pos: position{line: 561, col: 59, offset: 17591},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 561, col: 59, offset: 17591},
											expr: &ruleRefExpr{
												pos:  position{line: 561, col: 60, offset: 17592},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 561, col: 64, offset: 17596},
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
			pos:  position{line: 563, col: 1, offset: 17612},
			expr: &zeroOrMoreExpr{
				pos: position{line: 563, col: 7, offset: 17618},
				expr: &choiceExpr{
					pos: position{line: 563, col: 9, offset: 17620},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 563, col: 9, offset: 17620},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 563, col: 22, offset: 17633},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 563, col: 28, offset: 17639},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 564, col: 1, offset: 17650},
			expr: &zeroOrMoreExpr{
				pos: position{line: 564, col: 6, offset: 17655},
				expr: &choiceExpr{
					pos: position{line: 564, col: 8, offset: 17657},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 564, col: 8, offset: 17657},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 564, col: 21, offset: 17670},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 565, col: 1, offset: 17706},
			expr: &zeroOrMoreExpr{
				pos: position{line: 565, col: 7, offset: 17712},
				expr: &ruleRefExpr{
					pos:  position{line: 565, col: 7, offset: 17712},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 567, col: 1, offset: 17725},
			expr: &charClassMatcher{
				pos:        position{line: 567, col: 15, offset: 17739},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 568, col: 1, offset: 17747},
			expr: &litMatcher{
				pos:        position{line: 568, col: 8, offset: 17754},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 569, col: 1, offset: 17759},
			expr: &choiceExpr{
				pos: position{line: 569, col: 8, offset: 17766},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 569, col: 8, offset: 17766},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 8, offset: 17766},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 569, col: 11, offset: 17769},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 569, col: 17, offset: 17775},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 17, offset: 17775},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 569, col: 19, offset: 17777},
								expr: &ruleRefExpr{
									pos:  position{line: 569, col: 19, offset: 17777},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 569, col: 38, offset: 17796},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 569, col: 44, offset: 17802},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 569, col: 44, offset: 17802},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 569, col: 47, offset: 17805},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 571, col: 1, offset: 17810},
			expr: &notExpr{
				pos: position{line: 571, col: 8, offset: 17817},
				expr: &anyMatcher{
					line: 571, col: 9, offset: 17818,
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
	name := file.(string)
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
