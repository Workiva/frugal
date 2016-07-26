package test

import (
	"github.com/Workiva/frugal/compiler/parser"
	"testing"
)

const (
	validAuditThrift = "idl/valid_audit_thrift.frugal"
	validAuditScope  = "idl/valid_audit_scope.frugal"
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
	if err.Error() != "Exception InvalidOperation.why changed to DEFAULT from REQUIRED" {
		t.Fatal(err.Error())
	}

	err = parser.Compare(validAuditThrift, validFile)
	if err.Error() != "Exception InvalidOperation.why changed to REQUIRED from DEFAULT" {
		t.Fatal(err.Error())
	}
}

func TestValidAuditScope(t *testing.T) {
	err := parser.Compare(validFile, validAuditScope)
	if err.Error() != "Scope Foo Prefix foo.bar.{0}.qux not compatible with foo.bar.{0}.{1}.qux" {
		t.Fatal(err.Error())
	}

	err = parser.Compare(validAuditScope, validFile)
	if err.Error() != "Scope Foo Prefix foo.bar.{0}.{1}.qux not compatible with foo.bar.{0}.qux" {
		t.Fatal(err.Error())
	}
}
