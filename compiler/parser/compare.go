package parser

import (
	"log"
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
	NAME               = "Name"
	RTYPE              = "ReturnType"
	ARGUMENT           = "Argument"
	KEYTYPE            = "KeyType"
	VALUETYPE          = "ValueType"
)

//need global variable for typedefs to use in type checks
var oldTypeDefs, newTypeDefs map[string]*TypeDef

//Comparison of file with the audit file
func Compare(file, audit string) error {
	var err Error

	// parse the current frugal
	newFrugal, e := ParseFrugal(file)
	if e != nil {
		return e
	}

	// parse the frugal to compare with
	oldFrugal, e := ParseFrugal(audit)
	if e != nil {
		return e
	}

	// check typedefs first so we can use them later
	err.Append(checkThriftTypeDefs(newFrugal.Thrift.Typedefs, oldFrugal.Thrift.Typedefs))
	// check scopes
	err.Append(checkScopes(newFrugal.Scopes, oldFrugal.Scopes))
	// check thrift models
	err.Append(checkThriftStructs(newFrugal.Thrift.Structs, oldFrugal.Thrift.Structs))
	err.Append(checkThriftStructs(newFrugal.Thrift.Exceptions, oldFrugal.Thrift.Exceptions))
	err.Append(checkThriftStructs(newFrugal.Thrift.Unions, oldFrugal.Thrift.Unions))
	err.Append(checkThriftServices(newFrugal.Thrift.Services, oldFrugal.Thrift.Services))
	err.Append(checkThriftEnums(newFrugal.Thrift.Enums, oldFrugal.Thrift.Enums))
	// checking namespaces and constants only generate warnings
	err.Append(checkThriftNamespaces(newFrugal.Thrift.Namespaces, oldFrugal.Thrift.Namespaces))
	err.Append(checkThriftConstants(newFrugal.Thrift.Constants, oldFrugal.Thrift.Constants))

	// return nil if no errors
	if err.Error() == "" {
		return nil
	}
	return &err
}

func checkScopes(scopes1, scopes2 []*Scope) (err Error) {
	defer err.Prefix(SCOPE)

	sc1_map, e := makeScopeMap(scopes1)
	err.Append(e)
	sc2_map, e := makeScopeMap(scopes2)
	err.Append(e)
	for key, _ := range sc1_map {
		if _, ok := sc2_map[key]; ok {
			// check scope prefix
			err.Append(checkPrefix(sc1_map[key].Prefix, sc2_map[key].Prefix, " "+key+", "+PREFIX))
			// check scope operations
			op1_map, e := makeOperationMap(sc1_map[key].Operations)
			err.Append(e)
			op2_map, e := makeOperationMap(sc2_map[key].Operations)
			err.Append(e)
			for op, _ := range op1_map {
				if _, ok := op2_map[op]; ok {
					err.Append(checkType(op1_map[op].Type, op2_map[op].Type, TYPE))
				}
			}
			// can add scope operations but not remove them
			for op, _ := range op2_map {
				if _, ok := op1_map[op]; !ok {
					err.Append(NewErrorf("%s, Operation %s: removed", key, op2_map[op].Name))
				}
			}
		}
	}
	// can add scopes but not remove them
	for key, _ := range sc2_map {
		if _, ok := sc1_map[key]; !ok {
			err.Append(NewErrorf("%s: removed", sc2_map[key].Name))
		}
	}
	return err
}

func checkPrefix(pre1, pre2 *ScopePrefix, trace string) (err Error) {
	defer err.Prefix(trace)
	norm1 := normScopePrefix(pre1)
	norm2 := normScopePrefix(pre2)
	if norm1 != norm2 {
		err.Append(NewErrorf(": normalized %s not compatible with %s", pre1.String, pre2.String))
	}
	return err
}

func normScopePrefix(pre *ScopePrefix) string {
	for i, v := range pre.Variables {
		pre.String = strings.Replace(pre.String, v, strconv.Itoa(i), -1)
	}
	return pre.String
}

func checkType(t1, t2 *Type, trace string) (err Error) {
	defer err.Prefix(", " + trace)
	if t1 == nil || t2 == nil {
		if t1 != t2 {
			err.Append(NewErrorf(": not compatible %s, %s", t1, t2))
		}
		return err
	}
	// if the types are different we dont need to recurse further
	if t1.Name != t2.Name {
		t1, t1IsTypeDef := checkForTypeDef(t1, newTypeDefs)
		t2, t2IsTypeDef := checkForTypeDef(t2, oldTypeDefs)
		if t1IsTypeDef || t2IsTypeDef {
			err.Append(checkType(t1, t2, TYPEDEF))
		} else {
			err.Append(NewErrorf(": not equal %s, %s", t1.Name, t2.Name))
		}
	} else {
		err.Append(checkType(t1.KeyType, t2.KeyType, KEYTYPE))
		err.Append(checkType(t1.ValueType, t2.ValueType, VALUETYPE))
	}
	return err
}

func checkForTypeDef(t *Type, typeDefs map[string]*TypeDef) (*Type, bool) {
	// check if t is in thriftBaseTypes
	for key, _ := range thriftBaseTypes {
		if key == t.Name {
			return t, false
		}
	}
	// check if t is in thriftContainerTypes
	for key, _ := range thriftContainerTypes {
		if key == t.Name {
			return t, false
		}
	}
	// check if t is a typedef
	for key, _ := range typeDefs {
		if key == t.Name {
			return typeDefs[key].Type, true
		}
	}
	return t, false
}

func checkThriftStructs(structs1, structs2 []*Struct) (err Error) {
	if len(structs1) != 0 {
		defer err.Prefix(structs1[0].Type.String())
	}
	str1_map, e := makeStructureMap(structs1)
	err.Append(e)
	str2_map, e := makeStructureMap(structs2)
	err.Append(e)
	for key, _ := range str1_map {
		if _, ok := str2_map[key]; ok {
			// check Fields
			err.Append(checkFields(str1_map[key].Fields, str2_map[key].Fields, " "+str1_map[key].Name+"."))
		}
	}
	// can add structs but not remove them
	for key, _ := range str2_map {
		if _, ok := str1_map[key]; !ok {
			err.Append(NewErrorf("%s: removed", str2_map[key].Name))
		}
	}
	return err
}

func checkFields(f1s, f2s []*Field, trace string) (err Error) {
	defer err.Prefix(trace)
	f1_map, e := makeFieldMap(f1s)
	err.Append(e)
	f2_map, e := makeFieldMap(f2s)
	err.Append(e)
	// look for removed fields
	var max_id int
	var min_id int
	i := 0
	for key, _ := range f2_map {
		// get max and min key to check that additions fall outside [min,max]
		i += 1
		if i == 1 {
			max_id = key
			min_id = key
		}
		if key > max_id {
			max_id = key
		}
		if key < min_id {
			min_id = key
		}
		if _, ok := f1_map[key]; !ok && f2_map[key].Modifier != Optional {
			err.Append(NewErrorf("%s, ID=%d: removed", f2_map[key].Name, key))
		}
	}
	// look at new fields
	for key, _ := range f1_map {
		if _, ok := f2_map[key]; ok {
			err.Append(checkField(f1_map[key], f2_map[key], f1_map[key].Name))
		} else {
			if f1_map[key].Modifier == Required {
				err.Append(NewErrorf("%s, ID=%d: additional field is required", f1_map[key].Name, key))
			}
			if key < max_id && key > min_id {
				log.Printf("WARNING! %s, ID=%d: additional field does not have ID outside original range", f1_map[key].Name, key)
			}
		}
	}

	return err
}

func checkField(f1, f2 *Field, trace string) (err Error) {
	defer err.Prefix(trace)

	// check type
	err.Append(checkType(f1.Type, f2.Type, TYPE))
	// check modifier (Required, Optional, Default)
	err.Append(checkFieldModifier(f1, f2))
	return err
}

func checkFieldModifier(f1, f2 *Field) (err Error) {
	defer err.Prefix(": ")
	if (f1.Modifier == Required && f2.Modifier != Required) ||
		(f1.Modifier != Required && f2.Modifier == Required) ||
		(f1.Modifier != Default && f2.Modifier == Default) ||
		(f1.Modifier == Default && f2.Modifier != Default) {
		err.Append(NewErrorf("changed to %s from %s", f1.Modifier.String(), f2.Modifier.String()))
	}
	// changing default values only generates a warning
	if !reflect.DeepEqual(f1.Default, f2.Default) {
		log.Printf("WARNING! Default values changed %s: %#v, %#v\n", f1.Name, f1.Default, f2.Default)
	}
	return err
}

func checkThriftServices(services1, services2 []*Service) (err Error) {
	defer err.Prefix(SERVICE)

	serv1_map, e := makeServiceMap(services1)
	err.Append(e)
	serv2_map, e := makeServiceMap(services2)
	err.Append(e)
	for key, _ := range serv1_map {
		if _, ok := serv2_map[key]; ok {
			if serv1_map[key].Extends != serv2_map[key].Extends {
				err.Append(NewErrorf(" %s: extention not equal %s, %s", key, serv1_map[key].Extends, serv2_map[key].Extends))
			}
			err.Append(checkThriftServiceMethods(serv1_map[key].Methods, serv2_map[key].Methods, key))
		}
	}
	// can add services but not remove them
	for key, _ := range serv2_map {
		if _, ok := serv1_map[key]; !ok {
			err.Append(NewErrorf(" %s: removed", serv2_map[key].Name))
		}
	}
	return err
}

func checkThriftServiceMethods(meths1, meths2 []*Method, trace string) (err Error) {
	defer err.Prefix(" " + trace + ", " + METHOD)

	meth1_map, e := makeMethodMap(meths1)
	err.Append(e)
	meth2_map, e := makeMethodMap(meths2)
	err.Append(e)
	for key, _ := range meth1_map {
		if _, ok := meth2_map[key]; ok {
			// check direction of method
			if meth1_map[key].Oneway != meth2_map[key].Oneway {
				err.Append(NewErrorf(" %s: oneway not equal %t, %t", key, meth1_map[key].Oneway, meth2_map[key].Oneway))
			}
			err.Append(checkType(meth1_map[key].ReturnType, meth2_map[key].ReturnType, key+", "+RTYPE))
			err.Append(checkFields(meth1_map[key].Arguments, meth2_map[key].Arguments, " "+key+", "+ARGUMENT+" "))
			// check exceptions to method
			err.Append(checkFields(meth1_map[key].Exceptions, meth2_map[key].Exceptions, " "+key+", "+EXCEPTION+" "))
			// cant add exception with non-void return
			if meth1_map[key].ReturnType != nil {
				if len(meth1_map[key].Exceptions) > len(meth2_map[key].Exceptions) {
					err.Append(NewErrorf(": Cannot add a new exception to %s", key))
				}
			}
		}
	}
	// can add methods but not remove or rename them
	for key, _ := range meth2_map {
		if _, ok := meth1_map[key]; !ok {
			err.Append(NewErrorf(" %s: removed", meth2_map[key].Name))
		}
	}
	return err
}

func checkThriftTypeDefs(typedefs1, typedefs2 []*TypeDef) (err Error) {
	defer err.Prefix(TYPEDEF)
	var e Error
	newTypeDefs, e = makeTypeDefMap(typedefs1)
	err.Append(e)
	oldTypeDefs, e = makeTypeDefMap(typedefs2)
	err.Append(e)
	for key, _ := range newTypeDefs {
		if _, ok := oldTypeDefs[key]; ok {
			err.Append(checkType(newTypeDefs[key].Type, oldTypeDefs[key].Type, key+" "+TYPE))
		}
	}
	return err
}

func checkThriftEnums(enums1, enums2 []*Enum) (err Error) {
	defer err.Prefix(ENUM)

	enum1_map, e := makeEnumMap(enums1)
	err.Append(e)
	enum2_map, e := makeEnumMap(enums2)
	err.Append(e)
	for key, _ := range enum1_map {
		if _, ok := enum2_map[key]; ok {
			err.Append(checkEnumValues(enum1_map[key].Values, enum2_map[key].Values, " "+enum1_map[key].Name+"."))
		}
	}
	// can add enum but not remove them
	for key, _ := range enum2_map {
		if _, ok := enum1_map[key]; !ok {
			err.Append(NewErrorf("%s: removed", enum2_map[key].Name))
		}
	}
	return err
}

func checkEnumValues(vals1, vals2 []*EnumValue, trace string) (err Error) {
	defer err.Prefix(trace)
	// EnumValue map keys are the int values
	eval1_map, e := makeEnumValueMap(vals1)
	err.Append(e)
	eval2_map, e := makeEnumValueMap(vals2)
	err.Append(e)
	for key, _ := range eval1_map {
		if _, ok := eval2_map[key]; !ok {
			err.Append(NewErrorf("%s: value added %d", eval1_map[key].Name, key))
		}
	}
	// can add enum value but not remove them
	for key, _ := range eval2_map {
		if _, ok := eval1_map[key]; !ok {
			err.Append(NewErrorf("%s: value removed %d", eval2_map[key].Name, key))
		}
	}
	return err
}

// THE FOLLOWING FUNCTIONS ONLY GENERATE WARNINGS

func checkThriftConstants(constants1, constants2 []*Constant) (err Error) {
	// constant changes only generates warnings
	cons1_map, e := makeConstantMap(constants1)
	err.Append(e)
	cons2_map, e := makeConstantMap(constants2)
	err.Append(e)
	for key, _ := range cons1_map {
		if _, ok := cons2_map[key]; ok {
			// do deep equal on constant values
			if !reflect.DeepEqual(cons1_map[key].Value, cons2_map[key].Value) {
				log.Printf("WARNING! Constant value changed %s: %#v, %#v\n", key, cons1_map[key].Value, cons2_map[key].Value)
			}
		}
	}
	// warn if constant value is removed
	for key, _ := range cons2_map {
		if _, ok := cons1_map[key]; !ok {
			log.Printf("WARNING! Constant removed %s, %#v\n", key, cons2_map[key].Value)
		}
	}
	return err
}

func checkThriftNamespaces(namespaces1, namespaces2 []*Namespace) (err Error) {
	// Namespace changes only generate warnings
	ns1_map, e := makeNamespaceMap(namespaces1)
	err.Append(e)
	ns2_map, e := makeNamespaceMap(namespaces2)
	err.Append(e)
	for key, _ := range ns1_map {
		if _, ok := ns2_map[key]; ok {
			// do deep equal on namespaces
			if !reflect.DeepEqual(ns1_map[key], ns2_map[key]) {
				log.Printf("WARNING! Namespaces not equal %#v, %#v\n", ns1_map[key], ns2_map[key])
			}
		}
	}
	// warn if namespace is removed
	for key, _ := range ns2_map {
		if _, ok := ns1_map[key]; !ok {
			log.Printf("WARNING! Namespace removed %s\n", key)
		}
	}
	return err
}
