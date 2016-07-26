package parser

import (
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
	NAME               = "Name"
)

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

	// check scopes
	err.Append(checkScopes(newFrugal.Scopes, oldFrugal.Scopes, SCOPE))

	// check thrift models
	if !reflect.DeepEqual(newFrugal.Thrift, oldFrugal.Thrift) {
		err.Append(checkThriftStructs(newFrugal.Thrift.Structs, oldFrugal.Thrift.Structs, STRUCT))
		err.Append(checkThriftStructs(newFrugal.Thrift.Exceptions, oldFrugal.Thrift.Exceptions, EXCEPTION))
		err.Append(checkThriftStructs(newFrugal.Thrift.Unions, oldFrugal.Thrift.Unions, UNION))
		err.Append(checkThriftServices(newFrugal.Thrift.Services, oldFrugal.Thrift.Services, SERVICE))
		err.Append(checkThriftTypeDefs(newFrugal.Thrift.Typedefs, oldFrugal.Thrift.Typedefs, TYPEDEF))
		err.Append(checkThriftConstants(newFrugal.Thrift.Constants, oldFrugal.Thrift.Constants, CONSTANT))
		err.Append(checkThriftEnums(newFrugal.Thrift.Enums, oldFrugal.Thrift.Enums, ENUM))
		err.Append(checkThriftNamespaces(newFrugal.Thrift.Namespaces, oldFrugal.Thrift.Namespaces, NAMESPACE))
	}

	// return nil if no errors
	if err.Error() == "" {
		return nil
	}
	return &err
}

func checkScopes(scopes1, scopes2 []*Scope, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(scopes1), len(scopes2)))
	sc1_map, e := makeScopeMap(scopes1)
	err.Append(e)
	sc2_map, e := makeScopeMap(scopes2)
	err.Append(e)
	for key, _ := range sc1_map {
		if _, ok := sc2_map[key]; ok {
			// check scope prefix
			err.Append(checkPrefix(sc1_map[key].Prefix, sc2_map[key].Prefix, " "+key+" "+PREFIX))
			// check scope operations
			err.Append(checkLen(len(sc1_map[key].Operations), len(sc2_map[key].Operations)))
			op1_map, e := makeOperationMap(sc1_map[key].Operations)
			err.Append(e)
			op2_map, e := makeOperationMap(sc2_map[key].Operations)
			err.Append(e)
			for op, _ := range op1_map {
				if _, ok := op2_map[op]; ok {
					err.Append(checkType(op1_map[op].Type, op2_map[op].Type, TYPE))
				}
			}
		}
	}
	return err
}

func checkPrefix(pre1, pre2 *ScopePrefix, trace string) (err Error) {
	defer err.Prefix(trace)
	norm1 := normScopePrefix(pre1)
	norm2 := normScopePrefix(pre2)
	if norm1 != norm2 {
		err.Append(NewErrorf(" %s not compatible with %s", pre1.String, pre2.String))
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
	defer err.Prefix(" " + trace)
	if t1 == nil || t2 == nil {
		if t1 != t2 {
			err.Append(NewErrorf("types not compatible: %s, %s", t1, t2))
		}
		return err
	}
	err.Append(checkString(t1.Name, t2.Name))
	err.Append(checkType(t1.KeyType, t2.KeyType, TYPE))
	err.Append(checkType(t1.ValueType, t2.ValueType, TYPE))
	return err
}

func checkThriftStructs(structs1, structs2 []*Struct, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(structs1), len(structs2)))
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
	return err
}

func checkFields(f1s, f2s []*Field, trace string) (err Error) {
	defer err.Prefix(trace)
	f1_map, e := makeFieldMap(f1s)
	err.Append(e)
	f2_map, e := makeFieldMap(f2s)
	err.Append(e)
	for key, _ := range f1_map {
		if _, ok := f2_map[key]; ok {
			err.Append(checkField(f1_map[key], f2_map[key], f1_map[key].Name))
		} else {
			if f1_map[key].Modifier == Required {
				err.Append(NewErrorf("%s additional field is required", f1_map[key].Name))
			}
		}
	}
	// look for removed fields
	if len(f2_map) > len(f1_map) {
		for key, _ := range f2_map {
			if _, ok := f1_map[key]; !ok {
				if f2_map[key].Modifier == Required {
					err.Append(NewErrorf("%s removed field is required", f2_map[key].Name))
				}
			}
		}
	}
	return err
}

func checkField(f1, f2 *Field, trace string) (err Error) {
	defer err.Prefix(trace)
	// check ID
	if f1.ID != f2.ID {
		err.Append(NewErrorf("Field IDs not equal: %#v, %#v", f1.ID, f2.ID))
	}
	// check type
	err.Append(checkType(f1.Type, f2.Type, TYPE))
	// check modifier (Required, Optional, Default)
	err.Append(checkFieldModifier(f1, f2))
	return err
}

func checkFieldModifier(f1, f2 *Field) (err Error) {
	defer err.Prefix(" ")
	if (f1.Modifier == Required && f2.Modifier != Required) ||
		(f1.Modifier != Required && f2.Modifier == Required) ||
		(f1.Modifier != Default && f2.Modifier == Default) ||
		(f1.Modifier == Default && f2.Modifier != Default) {
		err.Append(NewErrorf("changed to %s from %s", f1.Modifier.String(), f2.Modifier.String()))
	}
	// TODO is this check necessary?
	if f1.Modifier == Default && f2.Modifier == Default {
		if !reflect.DeepEqual(f1.Default, f2.Default) {
			err.Append(NewErrorf("default values have changed"))
		}
	}
	return err
}

func checkThriftServices(services1, services2 []*Service, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(services1), len(services2)))
	serv1_map, e := makeServiceMap(services1)
	err.Append(e)
	serv2_map, e := makeServiceMap(services2)
	err.Append(e)
	for key, _ := range serv1_map {
		if _, ok := serv2_map[key]; ok {
			err.Append(checkString(serv1_map[key].Extends, serv2_map[key].Extends))
			err.Append(checkThriftServiceMethods(serv1_map[key].Methods, serv2_map[key].Methods, METHOD))
		}
	}
	return err
}

func checkThriftServiceMethods(meths1, meths2 []*Method, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(meths1), len(meths2)))
	meth1_map, e := makeMethodMap(meths1)
	err.Append(e)
	meth2_map, e := makeMethodMap(meths2)
	err.Append(e)
	for key, _ := range meth1_map {
		if _, ok := meth2_map[key]; ok {
			// check direction of method
			if meth1_map[key].Oneway != meth2_map[key].Oneway {
				err.Append(NewErrorf("Method oneway not equal: %#v, %#v", meth1_map[key].Oneway, meth2_map[key].Oneway))
			}
			err.Append(checkType(meth1_map[key].ReturnType, meth2_map[key].ReturnType, TYPE))
			err.Append(checkFields(meth1_map[key].Arguments, meth2_map[key].Arguments, FIELDS))
			// TODO this should be vetted more thoroughly
			err.Append(checkFields(meth1_map[key].Exceptions, meth2_map[key].Exceptions, FIELDS))
		}
	}
	return err
}

func checkThriftTypeDefs(typedefs1, typedefs2 []*TypeDef, trace string) (err Error) {
	defer err.Prefix(trace)

	tdef1_map, e := makeTypeDefMap(typedefs1)
	err.Append(e)
	tdef2_map, e := makeTypeDefMap(typedefs2)
	err.Append(e)
	for key, _ := range tdef1_map {
		if _, ok := tdef2_map[key]; ok {
			err.Append(checkType(tdef1_map[key].Type, tdef2_map[key].Type, TYPE))
		}
	}
	// can add typedefs but not remove them
	if len(typedefs1) < len(typedefs2) {
		for key, _ := range tdef2_map {
			if _, ok := tdef1_map[key]; !ok {
				err.Append(NewErrorf(" removed: %s", tdef2_map[key].Name))
			}
		}
	}
	return err
}

func checkThriftConstants(constants1, constants2 []*Constant, trace string) (err Error) {
	defer err.Prefix(trace)

	cons1_map, e := makeConstantMap(constants1)
	err.Append(e)
	cons2_map, e := makeConstantMap(constants2)
	err.Append(e)
	for key, _ := range cons1_map {
		if _, ok := cons2_map[key]; ok {
			err.Append(checkType(cons1_map[key].Type, cons2_map[key].Type, TYPE))
			// do deep equal on value
			// TODO is this check necessary?
			if !reflect.DeepEqual(cons1_map[key].Value, cons2_map[key].Value) {
				err.Append(NewErrorf(" %s values differ: new(%#v), old(%#v)", key, cons1_map[key].Value, cons2_map[key].Value))
			}
		}
	}
	// can add constants but not remove them
	if len(constants1) < len(constants2) {
		for key, _ := range cons2_map {
			if _, ok := cons1_map[key]; !ok {
				err.Append(NewErrorf(" removed: %s", cons2_map[key].Name))
			}
		}
	}
	return err

}

func checkThriftEnums(enums1, enums2 []*Enum, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(enums1), len(enums2)))
	enum1_map, e := makeEnumMap(enums1)
	err.Append(e)
	enum2_map, e := makeEnumMap(enums2)
	err.Append(e)
	for key, _ := range enum1_map {
		if _, ok := enum2_map[key]; ok {
			err.Append(checkEnumValues(enum1_map[key].Values, enum2_map[key].Values, " "+enum1_map[key].Name+"."))
		}
	}
	return err
}

func checkEnumValues(vals1, vals2 []*EnumValue, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(vals1), len(vals2)))
	eval1_map, e := makeEnumValueMap(vals1)
	err.Append(e)
	eval2_map, e := makeEnumValueMap(vals2)
	err.Append(e)
	for key, _ := range eval1_map {
		if _, ok := eval2_map[key]; ok {
			// check value
			if eval1_map[key].Value != eval2_map[key].Value {
				err.Append(NewErrorf("%s values differ: new(%#v), old(%#v)", eval1_map[key].Name, eval1_map[key].Value, eval2_map[key].Value))
			}
		}
	}
	return err
}

func checkThriftNamespaces(namespaces1, namespaces2 []*Namespace, trace string) (err Error) {
	defer err.Prefix(trace)
	err.Append(checkLen(len(namespaces1), len(namespaces2)))
	ns1_map, e := makeNamespaceMap(namespaces1)
	err.Append(e)
	ns2_map, e := makeNamespaceMap(namespaces2)
	err.Append(e)
	for key, _ := range ns1_map {
		if _, ok := ns2_map[key]; ok {
			// do deep equal on namespaces
			if !reflect.DeepEqual(ns1_map[key], ns2_map[key]) {
				err.Append(NewErrorf(" not equal: %#v, %#v", ns1_map[key], ns2_map[key]))
			}
		}
	}
	return err
}

func checkString(s1, s2 string) (err Error) {
	defer err.Prefix(" ")
	if s1 != s2 {
		err.Append(NewErrorf("not equal: %s, %s", s1, s2))
	}
	return err
}

func checkLen(l1, l2 int) (err Error) {
	defer err.Prefix(" ")
	if l1 != l2 {
		err.Append(NewErrorf(fmt.Sprintf("lengths differ new(%s), old(%s)", strconv.Itoa(l1), strconv.Itoa(l2))))
	}
	return err
}
