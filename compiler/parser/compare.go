package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	NAMESPACE          = "Namespace"
	TYPEDEF            = "TypeDef"
	CONSTANT           = "Constant"
	ENUM               = "Enum"
	STRUCT             = "Struct"
	EXCEPTION          = "Exception"
	UNION              = "Union"
	SERVICE            = "Service"
	ORDEREDDEFINITIONS = "OrderedDefinitions"
	SCOPE              = "Scope"
	OPERATION          = "Operation"
	FIELD              = "Field"
	ENUMVALUE          = "EnumValue"
	METHOD             = "Method"
	PREFIX             = "Prefix"
	TYPE               = "Type"
	THRIFT             = "Thrift"
	MODIFIER           = "Modifier"
	FIELDS             = "Fields"
)

func cleanTrace(name string) {
	if r := recover(); r != nil {
		panic(fmt.Sprintf("%s -> %s", name, r))
	}
}

func Compare(file, audit string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
		}
	}()

	// parse the current frugal
	var f1 *Frugal
	f1, err = ParseFrugal(file)
	if err != nil || f1 == nil {
		return err
	}

	// parse the frugal to compare with
	var f2 *Frugal
	f2, err = ParseFrugal(audit)
	if err != nil || f2 == nil {
		return err
	}

	// check scopes
	checkScopes(f1.Scopes, f2.Scopes, SCOPE)

	// check thrift models
	checkThrift(f1.Thrift, f2.Thrift, THRIFT)

	return err
}

func checkScopes(scopes1, scopes2 []*Scope, trace string) {
	defer cleanTrace(trace)

	if len(scopes1) == len(scopes2) {
		for i := range scopes1 {
			checkString(scopes1[i].Name, scopes2[i].Name, SCOPE)
			// check scope prefix
			checkPrefix(scopes1[i].Prefix, scopes2[i].Prefix, PREFIX)

			// check scope operations
			if len(scopes1[i].Operations) == len(scopes2[i].Operations) {
				for j := range scopes1[i].Operations {
					checkString(scopes1[i].Operations[j].Name, scopes2[i].Operations[j].Name, OPERATION)
					checkType(scopes1[i].Operations[j].Type, scopes2[i].Operations[j].Type, TYPE)
				}
			} else {
				checkLen(len(scopes1[i].Operations), len(scopes1[i].Operations), OPERATION)
			}
		}
	} else {
		checkLen(len(scopes1), len(scopes2), SCOPE)
	}
}

func checkPrefix(pre1, pre2 *ScopePrefix, trace string) {
	defer cleanTrace(trace)
	norm1 := normScopePrefix(pre1)
	norm2 := normScopePrefix(pre2)
	if norm1 != norm2 {
		panic(fmt.Sprintf("%s not compatible with %s", pre1.String, pre2.String))
	}
}

func normScopePrefix(pre *ScopePrefix) string {
	for i, v := range pre.Variables {
		pre.String = strings.Replace(pre.String, v, strconv.Itoa(i), -1)
	}
	return pre.String
}

func checkType(t1, t2 *Type, trace string) {
	defer cleanTrace(trace)
	if t1 == nil || t2 == nil {
		if t1 != t2 {
			panic(fmt.Sprintf("types not compatible! %s, %s\n", t1, t2))
		}
		return
	}
	checkString(t1.Name, t2.Name, SCOPE+" "+OPERATION)
	checkType(t1.KeyType, t2.KeyType, TYPE)
	checkType(t1.ValueType, t2.ValueType, TYPE)
}

func checkThrift(thrift1, thrift2 *Thrift, trace string) {
	defer cleanTrace(trace)
	if !reflect.DeepEqual(thrift1, thrift2) {
		checkThriftStructs(thrift1.Structs, thrift2.Structs, STRUCT)
		checkThriftStructs(thrift1.Exceptions, thrift2.Exceptions, STRUCT)
		checkThriftStructs(thrift1.Unions, thrift2.Unions, STRUCT)
		checkThriftServices(thrift1.Services, thrift2.Services, SERVICE)
		checkThriftTypeDefs(thrift1.Typedefs, thrift2.Typedefs, TYPEDEF)
		checkThriftConstants(thrift1.Constants, thrift2.Constants, CONSTANT)
		checkThriftEnums(thrift1.Enums, thrift2.Enums, ENUM)
		checkThriftNamespaces(thrift1.Namespaces, thrift2.Namespaces, NAMESPACE)
	}
}

func checkThriftStructs(structs1, structs2 []*Struct, trace string) {
	defer cleanTrace(trace)
	if len(structs1) == len(structs2) {
		for i := range structs1 {
			checkString(structs1[i].Name, structs2[i].Name, STRUCT)
			// check StructType (struct, exception, union)
			checkString(structs1[i].Type.String(), structs2[i].Type.String(), STRUCT+"Type")
			// check Fields
			checkFields(structs1[i].Fields, structs2[i].Fields, FIELDS)
		}
	} else {
		checkLen(len(structs1), len(structs2), STRUCT)
	}
}

func checkFields(f1s, f2s []*Field, trace string) {
	defer cleanTrace(trace)
	f1_map := makeFieldMap(f1s)
	f2_map := makeFieldMap(f2s)
	if len(f2s) > len(f1s) {
		checkAddedFields(f2_map, f1_map, FIELD)
	} else if len(f1s) > len(f2s) {
		checkAddedFields(f1_map, f2_map, FIELD)
	} else {
		for key, _ := range f1_map {
			checkField(f1_map[key], f2_map[key], FIELD)
		}
	}
}

func makeFieldMap(f []*Field) map[int]*Field {
	// map the fields according to their ID
	out := make(map[int]*Field)
	for i := range f {
		for key, _ := range out {
			if f[i].ID == key {
				panic(fmt.Sprintf("Duplicate IDs present: %#v\n", key))
			}
		}
		out[f[i].ID] = f[i]
	}
	return out
}

func checkField(f1, f2 *Field, trace string) {
	defer cleanTrace(trace)
	// check ID
	if f1.ID != f2.ID {
		panic(fmt.Sprintf("Field IDs not equal! %#v, %#v\n", f1.ID, f2.ID))
	}
	// check type
	checkType(f1.Type, f2.Type, TYPE)
	// check modifier (Required, Optional, Default)
	checkFieldModifier(f1, f2, MODIFIER)
}

func checkFieldModifier(f1, f2 *Field, trace string) {
	defer cleanTrace(trace)
	if f1.Modifier == Required && f2.Modifier != Required {
		panic(fmt.Sprintf("(%s) changed to required from optional or default (%s)", f1.Name, f2.Name))
	}
	if f1.Modifier != Required && f2.Modifier == Required {
		panic(fmt.Sprintf("(%s) changed to optional or default from required (%s)", f1.Name, f2.Name))
	}
	if f1.Modifier == Default && f2.Modifier == Default {
		if !reflect.DeepEqual(f1.Default, f2.Default) {
			panic(fmt.Sprintf("Default values have changed (%s)", f1.Name))
		}
	} else if f1.Modifier != Default && f2.Modifier == Default || f1.Modifier == Default && f2.Modifier != Default {
		panic(fmt.Sprintf("(%s) is no longer defaulted", f1.Name))
	} else if f1.Modifier == Default && f2.Modifier != Default {
		panic(fmt.Sprintf("(%s) is defaulted", f1.Name))
	}
}

func checkAddedFields(f1s, f2s map[int]*Field, trace string) {
	defer cleanTrace(trace)
	// first go through the first N old fields and make sure they have not changed
	for key, _ := range f1s {
		_, ok := f2s[key]
		if ok {
			checkField(f1s[key], f2s[key], FIELD)
		} else {
			if f1s[key].Modifier == Required {
				panic(fmt.Sprintf("Added/Removed field is required! %s", f1s[key].Name))
			}
		}
	}
}

func checkThriftServices(services1, services2 []*Service, trace string) {
	defer cleanTrace(trace)
	if len(services1) == len(services2) {
		for i := range services1 {
			checkString(services1[i].Name, services2[i].Name, SERVICE)
			checkString(services1[i].Extends, services2[i].Extends, SERVICE+"Extends")
			checkThriftServiceMethods(services1[i].Methods, services2[i].Methods, METHOD)
		}
	} else {
		checkLen(len(services1), len(services2), SERVICE)
	}
}

func checkThriftServiceMethods(meths1, meths2 []*Method, trace string) {
	defer cleanTrace(trace)
	if len(meths1) == len(meths2) {
		for i := range meths1 {
			checkString(meths1[i].Name, meths2[i].Name, METHOD)
			// check direction of method
			if meths1[i].Oneway != meths2[i].Oneway {
				panic(fmt.Sprintf("Method oneway not equal! %#v, %#v\n", meths1[i].Oneway, meths2[i].Oneway))
			}
			checkType(meths1[i].ReturnType, meths2[i].ReturnType, TYPE)
			checkFields(meths1[i].Arguments, meths2[i].Arguments, FIELDS)
			checkFields(meths1[i].Exceptions, meths2[i].Exceptions, FIELDS)
		}
	} else {
		checkLen(len(meths1), len(meths2), METHOD)
	}
}

func checkThriftTypeDefs(typedefs1, typedefs2 []*TypeDef, trace string) {
	defer cleanTrace(trace)
	if len(typedefs1) == len(typedefs2) {
		for i := range typedefs1 {
			checkString(typedefs1[i].Name, typedefs2[i].Name, TYPEDEF)
			checkType(typedefs1[i].Type, typedefs2[i].Type, TYPE)
		}
	} else {
		checkLen(len(typedefs1), len(typedefs2), TYPEDEF)
	}
}

func checkThriftConstants(constants1, constants2 []*Constant, trace string) {
	defer cleanTrace(trace)
	if len(constants1) == len(constants2) {
		for i := range constants1 {
			checkString(constants1[i].Name, constants2[i].Name, CONSTANT)
			// do deep equal on value
			if !reflect.DeepEqual(constants1[i].Value, constants2[i].Value) {
				panic(fmt.Sprintf("Constant values not equal! %#v, %#v\n", constants1[i].Value, constants2[i].Value))
			}
		}
	} else {
		checkLen(len(constants1), len(constants2), CONSTANT)
	}
}

func checkThriftEnums(enums1, enums2 []*Enum, trace string) {
	defer cleanTrace(trace)
	if len(enums1) == len(enums2) {
		for i := range enums1 {
			checkString(enums1[i].Name, enums2[i].Name, ENUM)
			checkEnumValues(enums1[i].Values, enums2[i].Values, ENUMVALUE)
		}
	} else {
		checkLen(len(enums1), len(enums2), ENUM)
	}
}

func checkEnumValues(vals1, vals2 []*EnumValue, trace string) {
	defer cleanTrace(trace)
	if len(vals1) == len(vals2) {
		for i := range vals1 {
			checkString(vals1[i].Name, vals2[i].Name, ENUMVALUE)
			// check value
			if vals1[i].Value != vals2[i].Value {
				panic(fmt.Sprintf("EnumValue name not equal! %#v, %#v\n", vals1[i].Value, vals2[i].Value))
			}
		}
	} else {
		checkLen(len(vals1), len(vals2), ENUMVALUE)
	}
}

func checkThriftNamespaces(namespaces1, namespaces2 []*Namespace, trace string) {
	defer cleanTrace(trace)
	if len(namespaces1) == len(namespaces2) {
		for i := range namespaces1 {
			// do deep equal on namespaces
			if !reflect.DeepEqual(namespaces1[i], namespaces2[i]) {
				panic(fmt.Sprintf("Namespace not equal! %#v, %#v\n", namespaces1[i], namespaces2[i]))
			}
		}
	} else {
		checkLen(len(namespaces1), len(namespaces2), NAMESPACE)
	}
}

func checkString(s1, s2, trace string) {
	defer cleanTrace(trace)
	if s1 != s2 {
		panic(fmt.Sprintf("%s not equal! %s, %s\n", trace, s1, s2))
	}
}

func checkLen(l1, l2 int, trace string) {
	defer cleanTrace(trace)
	if l1 > l2 {
		panic(fmt.Sprintf("There are new %ss", trace))
	} else {
		panic(fmt.Sprintf("There are removed %ss", trace))
	}
}

// func checkOrderedDefinitions(defs1, defs2 []interface{}) {
// 	if len(defs1) == len(defs2) {
// 		for i := range defs1 {
// 			if !reflect.DeepEqual(defs1[i], defs2[i]) {
// 				panic("OrderedDefinitions not equal!")
// 			}
// 		}
// 	} else {
// 		checkLen(len(defs1), len(defs2), ORDEREDDEFINITIONS)
// 	}
// }
