package test

import (
	"fmt"
	"github.com/Workiva/frugal/compiler/parser"
	"testing"
)

const (
	validAuditThrift = "idl/valid_audit_thrift.frugal"
	validAuditScope  = "idl/valid_audit_scope.frugal"
	testFileThrift   = "idl/breaking_changes/test.thrift"
	testWarning      = "idl/breaking_changes/warning.thrift"
)

func TestPassingAudit(t *testing.T) {

	if err := parser.Compare(validFile, validFile); err != nil {
		t.Fatal("Unexpected error", err)
	}
}

func TestInvalidAudit(t *testing.T) {

	if err := parser.Compare(invalidFile, validFile); err == nil {
		t.Fatal("Compare should fail for invalid frugal file")
	}

	if err := parser.Compare(validFile, invalidFile); err == nil {
		t.Fatal("Compare should fail for invalid frugal file")
	}

	if err := parser.Compare(invalidFile, invalidFile); err == nil {
		t.Fatal("Compare should fail for invalid frugal file")
	}
}

func TestValidAuditThrift(t *testing.T) {
	err := parser.Compare(validFile, validAuditThrift)
	expected := "exception InvalidOperation.why: changed to DEFAULT from REQUIRED"
	if err.Error() != expected {
		t.Fatal(fmt.Sprintf("\nExpected: %s\nBut got : %s\n", expected, err.Error()))
	}

	err = parser.Compare(validAuditThrift, validFile)
	expected = "exception InvalidOperation.why: changed to REQUIRED from DEFAULT"
	if err.Error() != expected {
		t.Fatal(fmt.Sprintf("\nExpected: %s\nBut got : %s\n", expected, err.Error()))
	}
}

func TestValidAuditScope(t *testing.T) {
	err := parser.Compare(validFile, validAuditScope)
	expected := "Scope Foo, Prefix: normalized foo.bar.{0}.qux not compatible with foo.bar.{0}.{1}.qux"
	if err.Error() != expected {
		t.Fatal(fmt.Sprintf("\nExpected: %s\nBut got : %s\n", expected, err.Error()))
	}

	err = parser.Compare(validAuditScope, validFile)
	expected = "Scope Foo, Prefix: normalized foo.bar.{0}.{1}.qux not compatible with foo.bar.{0}.qux"
	if err.Error() != expected {
		t.Fatal(fmt.Sprintf("\nExpected: %s\nBut got : %s\n", expected, err.Error()))
	}
}

func TestBreakingChanges(t *testing.T) {
	expected := []string{
		"Service base, Method base_function3: removed",
		"struct test_struct1.struct1_member1, Type: not equal i32, i16",
		"struct test_struct1.struct1_member9, Type: not equal test_enum2, test_enum1",
		"struct test_struct1.struct1_member6, Type: not equal string, bool",
		"struct test_struct1.struct1_member6, Type: not equal list, bool",
		"struct test_struct2.struct2_member4, Type, ValueType: not equal i16, double",
		"struct test_struct6.struct6_member2: changed to DEFAULT from REQUIRED",
		"struct test_struct5.struct5_member2: changed to REQUIRED from DEFAULT",
		"struct test_struct1.struct1_member7, ID=7: removed",
		"struct test_struct2.struct2_member1, ID=1: removed",
		"struct test_struct3.struct3_member7, ID=7: removed",
		"Service derived1, Method, derived1_function1, ReturnType: not equal test_enum2, test_enum1",
		"Service derived1, Method, derived1_function6, ReturnType: not equal test_struct2, test_struct1",
		"Service derived1, Method, derived1_function4, ReturnType: not equal double, string",
		"Service derived2, Method, derived2_function1, ReturnType, ValueType: not equal i16, i32",
		"Service derived2, Method, derived2_function5, ReturnType, KeyType: not equal test_enum3, test_enum1",
		"Service derived2, Method, derived2_function6, ReturnType, ValueType: not equal test_struct3, test_struct2",
		"Service base, Method base_oneway: oneway not equal false, true",
		"Service base, Method base_function1: oneway not equal true, false",
		"Enum test_enum1.enum1_value0: removed",
		"Enum test_enum2.enum2_value3: removed",
		"Enum test_enum1.enum1_value2: removed",
		"struct test_struct4.struct4_member3, ID=3: additional field is required",
		"Service derived1: extention not equal , base",
		"Service derived2: extention not equal derived1, base",
		"Service base, Method base_function1, Argument function1_arg3, Type: not equal double, i64",
		"Service base, Method base_function2, Argument function2_arg8, Type, ValueType: not equal test_enum3, test_enum1",
		"Service derived1, Method derived1_function5, Argument function5_arg1, Type: not equal list, map",
		"Service base, Method base_function2, Argument function2_arg5, Type: not equal string, list",
		"Service derived1, Method, derived1_function6, ReturnType: not equal map, test_struct1",
		"Service base, Method base_function2, Exception e, ID=1: removed",
		"exception test_exception1.code, Type: not equal i64, i32",
		"Service derived1, Method derived1_function1, Exception e, Type: not equal test_exception1, test_exception2",
		"struct test_struct3.struct3_member6, ID=6: additional field does not have ID outside original range",
	}
	for i := 0; i < 34; i++ {

		badFile := fmt.Sprintf("idl/breaking_changes/break%d.thrift", i+1)
		err := parser.Compare(badFile, testFileThrift)
		if err.Error() != expected[i] {
			t.Fatalf("checking %s\nExpected: %s\nBut got : %s\n", badFile, expected[i], err.Error())
		}
	}
}

func TestWarningChanges(t *testing.T) {
	err := parser.Compare(testWarning, testFileThrift)
	if err != nil {
		t.Fatalf("\nExpected no errors, but got: %s", err.Error())
	}
}
