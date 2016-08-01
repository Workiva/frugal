package test

import (
	"fmt"
	"testing"
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

const (
	testFileThrift = "idl/breaking_changes/test.thrift"
	testWarning    = "idl/breaking_changes/warning.thrift"
	scopeFile      = "idl/breaking_changes/scope.frugal"
)

type MockValidationLogger struct {
	errors []string
	warnings []string
}

func (m *MockValidationLogger) LogWarning(pieces ...string) {
	m.warnings = append(m.warnings, strings.Join(pieces, " "))
}

func (m *MockValidationLogger) LogError(pieces ...string) {
	m.errors = append(m.errors, strings.Join(pieces, " "))
}

func (m *MockValidationLogger) ErrorsLogged() bool {
	return len(m.errors) > 0
}

func TestPassingAudit(t *testing.T) {
	auditor := parser.NewAuditorWithLogger(&MockValidationLogger{})
	if err := auditor.Compare(validFile, validFile); err != nil {
		t.Fatal("unexpected error", err)
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
		"Enum test_enum1.enum1_value0: value removed 0",
		"Enum test_enum2.enum2_value3: value removed 3",
		"Enum test_enum1.enum1_value2: value removed 2",
		"struct test_struct4.struct4_member3, ID=3: additional field is required",
		"Service derived1: extention not equal , base",
		"Service derived2: extention not equal derived1, base",
		"Service base, Method base_function1, Argument function1_arg3, Type: not equal double, i64",
		"Service base, Method base_function2, Argument function2_arg8, Type, ValueType: not equal test_enum3, test_enum1",
		"Service derived1, Method derived1_function5, Argument function5_arg1, Type: not equal list, map",
		"Service base, Method base_function2, Argument function2_arg5, Type: not equal string, list",
		"Service derived1, Method, derived1_function6, ReturnType: not equal map, test_struct1",
		"Service base, Method base_function2: Cannot remove exceptions",
		"exception test_exception1.code, Type: not equal i64, i32",
		"Service derived1, Method derived1_function1, Exception e, Type: not equal test_exception1, test_exception2",
	}
	for i := 0; i < 33; i++ {

		badFile := fmt.Sprintf("idl/breaking_changes/break%d.thrift", i+1)
		logger := &MockValidationLogger{}
		auditor := parser.NewAuditorWithLogger(logger)
		err := auditor.Compare(badFile, testFileThrift)
		if err != nil {
			if logger.errors[0] != expected[i] {
				t.Fatalf("checking %s\nExpected: %s\nBut got : %s\n", badFile, expected[i], logger.errors[0])
			}
		} else {
			t.Fatalf("No errors found for %s\n", badFile)
		}
	}
}

func TestWarningChanges(t *testing.T) {
	auditor := parser.NewAuditorWithLogger(&MockValidationLogger{})
	err := auditor.Compare(testWarning, testFileThrift)
	if err != nil {
		t.Fatalf("\nExpected no errors, but got: %s", err.Error())
	}
}

func TestScopeBreakingChanges(t *testing.T) {
	expected := []string{
		"Scope Foo, Prefix: normalized foo.bar.{0}.{1}.{2}.qux not compatible with foo.bar.{0}.{1}.qux",
		"Scope Foo, Prefix: normalized foo.bar.{0}.qux not compatible with foo.bar.{0}.{1}.qux",
		"Scope blah: removed",
		"Scope Foo, Prefix: normalized foo.bar.{0}.{1}.qux.que not compatible with foo.bar.{0}.{1}.qux",
		"Scope Foo, Prefix: normalized foo.bar.{0}.{1} not compatible with foo.bar.{0}.{1}.qux",
		"Scope Foo, Operation Bar: removed",
		"Scope Foo, Operation Foo, Type: not equal int, Thing",
	}
	for i := 0; i < 7; i++ {
		badFile := fmt.Sprintf("idl/breaking_changes/scope%d.frugal", i+1)
		logger := &MockValidationLogger{}
		auditor := parser.NewAuditorWithLogger(logger)
		err := auditor.Compare(badFile, scopeFile)
		if err != nil {
			if logger.errors[0] != expected[i] {
				t.Fatalf("checking %s\nExpected: %s\nBut got : %s\n", badFile, expected[i], logger.errors[0])
			}
		} else {
			t.Fatalf("No errors found for %s\n", badFile)
		}
	}
}
