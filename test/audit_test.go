package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Workiva/frugal/compiler/parser"
	"github.com/stretchr/testify/assert"
)

const (
	testFileThrift = "idl/breaking_changes/test.thrift"
	testWarning    = "idl/breaking_changes/warning.thrift"
	scopeFile      = "idl/breaking_changes/scope.frugal"
)

type MockValidationLogger struct {
	errors   []string
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
	if err := auditor.Audit(validFile, validFile); err != nil {
		t.Fatal("unexpected error", err)
	}
}

func TestBreakingChanges(t *testing.T) {
	expected := []string{
		"service base: missing method: base_function3",
		"struct test_struct1: field struct1_member1: types not equal: 'i16' -> 'i32'",
		"struct test_struct1: field struct1_member9: types not equal: 'test_enum1' -> 'test_enum2'",
		"struct test_struct1: field struct1_member6: types not equal: 'bool' -> 'string'",
		"struct test_struct1: field struct1_member6: types not equal: 'bool' -> 'list'",
		"struct test_struct2: field struct2_member4: value type: types not equal: 'double' -> 'i16'",
		"struct test_struct6: field struct6_member2: field presence modifier changed: 'REQUIRED' -> 'DEFAULT'",
		"struct test_struct5: field struct5_member2: field presence modifier changed: 'DEFAULT' -> 'REQUIRED'",
		"struct test_struct1: field struct1_member7: field removed with ID=7",
		"struct test_struct2: field struct2_member1: field removed with ID=1",
		"struct test_struct3: field struct3_member7: field removed with ID=7",
		"service derived1: method derived1_function1: return type: types not equal: 'test_enum1' -> 'test_enum2'",
		"service derived1: method derived1_function6: return type: types not equal: 'test_struct1' -> 'test_struct2'",
		"service derived1: method derived1_function4: return type: types not equal: 'string' -> 'double'",
		"service derived2: method derived2_function1: return type: value type: types not equal: 'i32' -> 'i16'",
		"service derived2: method derived2_function5: return type: key type: types not equal: 'test_enum1' -> 'test_enum3'",
		"service derived2: method derived2_function6: return type: value type: types not equal: 'test_struct2' -> 'test_struct3'",
		"service base: method base_oneway: one way modifier changed",
		"service base: method base_function1: one way modifier changed",
		"enum test_enum1: variant enum1_value0: removed with ID=0",
		"enum test_enum2: variant enum2_value3: removed with ID=3",
		"enum test_enum1: variant enum1_value2: removed with ID=2",
		"struct test_struct4: field struct4_member3: added field is required",
		"service derived1: extends changed: 'base' -> ''",
		"service derived2: extends changed: 'base' -> 'derived1'",
		"service base: method base_function1: field function1_arg3: types not equal: 'i64' -> 'double'",
		"service base: method base_function2: field function2_arg8: value type: types not equal: 'test_enum1' -> 'test_enum3'",
		"service derived1: method derived1_function5: field function5_arg1: types not equal: 'map' -> 'list'",
		"service base: method base_function2: field function2_arg5: types not equal: 'list' -> 'string'",
		"service derived1: method derived1_function6: return type: types not equal: 'test_struct1' -> 'map'",
		"service base: method base_function2: can't remove exceptions with nil return type",
		"struct test_exception1: field code: types not equal: 'i32' -> 'i64'",
		"service derived1: method derived1_function1: field e: types not equal: 'test_exception2' -> 'test_exception1'",
	}
	for i := 0; i < 33; i++ {

		badFile := fmt.Sprintf("idl/breaking_changes/break%d.thrift", i+1)
		logger := &MockValidationLogger{}
		auditor := parser.NewAuditorWithLogger(logger)
		err := auditor.Audit(testFileThrift, badFile)
		if err != nil {
			if logger.errors[0] != expected[i] {
				t.Fatalf("checking %s\nExpected: %s\nBut got : %s\n", badFile, expected[i], logger.errors[0])
			}
			assert.Len(t, logger.errors, 1)
			assert.Len(t, logger.warnings, 0)
		} else {
			t.Fatalf("No errors found for %s\n", badFile)
		}
	}
}

func TestWarningChanges(t *testing.T) {
	auditor := parser.NewAuditorWithLogger(&MockValidationLogger{})
	err := auditor.Audit(testFileThrift, testWarning)
	if err != nil {
		t.Fatalf("\nExpected no errors, but got: %s", err.Error())
	}
}

func TestScopeBreakingChanges(t *testing.T) {
	expected := []string{
		"scope Foo: prefix changed: 'foo.bar.{}.{}.qux' -> 'foo.bar.{}.{}.{}.qux'",
		"scope Foo: prefix changed: 'foo.bar.{}.{}.qux' -> 'foo.bar.{}.qux'",
		"missing scope: blah",
		"scope Foo: prefix changed: 'foo.bar.{}.{}.qux' -> 'foo.bar.{}.{}.qux.que'",
		"scope Foo: prefix changed: 'foo.bar.{}.{}.qux' -> 'foo.bar.{}.{}'",
		"scope Foo: operation removed: Bar",
		"scope Foo: operation Foo: types not equal: 'Thing' -> 'int'",
	}
	for i := 0; i < 7; i++ {
		badFile := fmt.Sprintf("idl/breaking_changes/scope%d.frugal", i+1)
		logger := &MockValidationLogger{}
		auditor := parser.NewAuditorWithLogger(logger)
		err := auditor.Audit(scopeFile, badFile)
		if err != nil {
			if logger.errors[0] != expected[i] {
				t.Fatalf("checking %s\nExpected: %s\nBut got : %s\n", badFile, expected[i], logger.errors[0])
			}
		} else {
			t.Fatalf("No errors found for %s\n", badFile)
		}
	}
}
