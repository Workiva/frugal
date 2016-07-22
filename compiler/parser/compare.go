package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	// "github.com/Workiva/frugal/compiler/parser"
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
)

func Compare(compare_sha, slug, token, file string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovering from Error!\n", r)
		}
	}()

	// get frugal to compare with
	tmpfile, _ := ioutil.TempFile("", file)
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name()) // clean up
	tmpfile.Write(getFile(slug, file, token))

	// parse the current frugal
	f1, err := ParseFrugal(file)
	if err != nil || f1 == nil {
		panic(fmt.Sprintf("Aawww snap! Could not parse frugal file: %s\n", file))
	}

	// parse the frugal to compare with
	f2, err := ParseFrugal(tmpfile.Name())
	if err != nil || f2 == nil {
		panic(fmt.Sprintf("Aawww snap! Could not parse frugal file: %s\n", file))
	}

	// If the two definitions are not equal, do the checks
	if !reflect.DeepEqual(f1, f2) {
		checkScopes(f1.Scopes, f2.Scopes)
		checkThrift(f1.Thrift, f2.Thrift)
		// Can ignore if and when thrift is dead
		// checkOrderedDefinitions(f.OrderedDefinitions, f2.OrderedDefinitions)
	}
	fmt.Printf("Passed: %s\n", file)
	return err
}

func checkScopes(scopes1, scopes2 []*Scope) {
	if len(scopes1) == len(scopes2) {
		for i := range scopes1 {
			checkString(scopes1[i].Name, scopes2[i].Name, SCOPE)
			// check scope prefix
			norm1 := normScopePrefix(scopes1[i].Prefix)
			norm2 := normScopePrefix(scopes2[i].Prefix)
			if norm1 != norm2 {
				panic(fmt.Sprintf("Scope prefix not compatible! %s, %s\n", norm1, norm2))
			}
			// check scope operations
			if len(scopes1[i].Operations) == len(scopes2[i].Operations) {
				for j := range scopes1[i].Operations {
					checkString(scopes1[i].Operations[j].Name, scopes2[i].Operations[j].Name, OPERATION)
					checkType(scopes1[i].Operations[j].Type, scopes2[i].Operations[j].Type)
				}
			} else {
				checkLen(len(scopes1[i].Operations), len(scopes1[i].Operations), OPERATION)
			}
		}
	} else {
		checkLen(len(scopes1), len(scopes2), SCOPE)
	}
}

func normScopePrefix(pre *ScopePrefix) string {
	for i, v := range pre.Variables {
		pre.String = strings.Replace(pre.String, v, strconv.Itoa(i), -1)
	}
	return pre.String
}

func checkType(t1, t2 *Type) {
	if t1 == nil || t2 == nil {
		if t1 != t2 {
			panic(fmt.Sprintf("types not compatible! %s, %s\n", t1, t2))
		}
		return
	}
	checkString(t1.Name, t2.Name, SCOPE+" "+OPERATION)
	checkType(t1.KeyType, t2.KeyType)
	checkType(t1.ValueType, t2.ValueType)
}

func checkThrift(thrift1, thrift2 *Thrift) {
	if !reflect.DeepEqual(thrift1, thrift2) {
		checkThriftStructs(thrift1.Structs, thrift2.Structs)
		checkThriftStructs(thrift1.Exceptions, thrift2.Exceptions)
		checkThriftStructs(thrift1.Unions, thrift2.Unions)
		checkThriftServices(thrift1.Services, thrift2.Services)
		checkThriftTypeDefs(thrift1.Typedefs, thrift2.Typedefs)
		checkThriftConstants(thrift1.Constants, thrift2.Constants)
		checkThriftEnums(thrift1.Enums, thrift2.Enums)
		checkThriftNamespaces(thrift1.Namespaces, thrift2.Namespaces)
	}
}

func checkThriftStructs(structs1, structs2 []*Struct) {
	if len(structs1) == len(structs2) {
		for i := range structs1 {
			checkString(structs1[i].Name, structs2[i].Name, STRUCT)
			// check StructType (struct, exception, union)
			checkString(structs1[i].Type.String(), structs2[i].Type.String(), STRUCT+"TYPE")
			// check Fields
			checkFields(structs1[i].Fields, structs2[i].Fields)
		}
	} else {
		checkLen(len(structs1), len(structs2), STRUCT)
	}
}

func checkFields(f1s, f2s []*Field) {
	f1_map := makeFieldMap(f1s)
	f2_map := makeFieldMap(f2s)
	if len(f2s) > len(f1s) {
		checkAddedFields(f2_map, f1_map)
	} else if len(f1s) > len(f2s) {
		checkAddedFields(f1_map, f2_map)
	} else {
		for key, _ := range f1_map {
			checkField(f1_map[key], f2_map[key])
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

func checkField(f1, f2 *Field) {
	// We can let this one slide
	if f1.Name != f2.Name {
		fmt.Printf("Field names not equal, but thats ok... %s, %s\n", f1.Name, f2.Name)
	}
	// check ID
	if f1.ID != f2.ID {
		panic(fmt.Sprintf("Field IDs not equal! %#v, %#v\n", f1.ID, f2.ID))
	}
	// check type
	checkType(f1.Type, f2.Type)
	// check modifier (Required, Optional, Default)
	checkFieldModifier(f1, f2)
}

func checkFieldModifier(f1, f2 *Field) {
	if f1.Modifier == Required && f2.Modifier != Required {
		panic("Field changed to required from optional or default")
	}
	if f1.Modifier != Required && f2.Modifier == Required {
		panic("Field changed to optional or default from required")
	}
	if f1.Modifier == Default && f2.Modifier == Default {
		if !reflect.DeepEqual(f1.Default, f2.Default) {
			panic("Default values have changed")
		}
	} else if f1.Modifier != Default && f2.Modifier == Default || f1.Modifier == Default && f2.Modifier != Default {
		panic("Field default <-> non-default")
	}
}

func checkAddedFields(f1s, f2s map[int]*Field) {
	fmt.Printf("there are added/removed fields, but it could be ok...\n")
	// first go through the first N old fields and make sure they have not changed
	for key, _ := range f1s {
		_, ok := f2s[key]
		if ok {
			checkField(f1s[key], f2s[key])
		} else {
			if f1s[key].Modifier == Required {
				panic("Added/Removed field is required!")
			}
		}
	}
}

func checkThriftServices(services1, services2 []*Service) {
	if len(services1) == len(services2) {
		for i := range services1 {
			checkString(services1[i].Name, services2[i].Name, SERVICE)
			checkString(services1[i].Extends, services2[i].Extends, SERVICE+"Extends")
			checkThriftServiceMethods(services1[i].Methods, services2[i].Methods)
		}
	} else {
		checkLen(len(services1), len(services2), SERVICE)
	}
}

func checkThriftServiceMethods(meths1, meths2 []*Method) {
	if len(meths1) == len(meths2) {
		for i := range meths1 {
			checkString(meths1[i].Name, meths2[i].Name, METHOD)
			// check direction of method
			if meths1[i].Oneway != meths2[i].Oneway {
				panic(fmt.Sprintf("Method oneway not equal! %#v, %#v\n", meths1[i].Oneway, meths2[i].Oneway))
			}
			checkType(meths1[i].ReturnType, meths2[i].ReturnType)
			checkFields(meths1[i].Arguments, meths2[i].Arguments)
			checkFields(meths1[i].Exceptions, meths2[i].Exceptions)
		}
	} else {
		checkLen(len(meths1), len(meths2), METHOD)
	}
}

func checkThriftTypeDefs(typedefs1, typedefs2 []*TypeDef) {
	if len(typedefs1) == len(typedefs2) {
		for i := range typedefs1 {
			checkString(typedefs1[i].Name, typedefs2[i].Name, TYPEDEF)
			checkType(typedefs1[i].Type, typedefs2[i].Type)
		}
	} else {
		checkLen(len(typedefs1), len(typedefs2), TYPEDEF)
	}
}

func checkThriftConstants(constants1, constants2 []*Constant) {
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

func checkThriftEnums(enums1, enums2 []*Enum) {
	if len(enums1) == len(enums2) {
		for i := range enums1 {
			checkString(enums1[i].Name, enums2[i].Name, ENUM)
			checkEnumValues(enums1[i].Values, enums2[i].Values)
		}
	} else {
		checkLen(len(enums1), len(enums2), ENUM)
	}
}

func checkEnumValues(vals1, vals2 []*EnumValue) {
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

func checkThriftNamespaces(namespaces1, namespaces2 []*Namespace) {
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

func checkString(s1, s2, name string) {
	if s1 != s2 {
		panic(fmt.Sprintf("%s name/string not equal! %s, %s\n", name, s1, s2))
	}
}

func checkLen(l1, l2 int, t string) {
	if l1 > l2 {
		panic(fmt.Sprintf("There are new %ss\n", t))
	} else {
		panic(fmt.Sprintf("There are removed %ss\n", t))
	}
}

func getFile(slug, file, token string) []byte {

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+slug+"/contents/"+file, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "token "+token)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	obj := struct {
		Content []byte `json:"content"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(&obj); err != nil {
		panic(err)
	}
	return obj.Content
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
