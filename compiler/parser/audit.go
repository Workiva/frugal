/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parser

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidationLogger provides an interface to output validation results.
type ValidationLogger interface {
	// LogWarning should log a warning message
	LogWarning(...string)
	// LogError should log an error message
	LogError(...string)
	// ErrorsLogged should return true if any errors have been logged,
	// and false otherwise
	ErrorsLogged() bool
}

// StdOutLogger is a validation logger that prints message to standard out
type stdOutLogger struct {
	errorsLogged bool
}

// TODO use tty colors to emphasize things?
func (s *stdOutLogger) LogWarning(warning ...string) {
	fmt.Println("WARNING:", warning)
}

func (s *stdOutLogger) LogError(errorMessage ...string) {
	s.errorsLogged = true
	fmt.Println("ERROR:", errorMessage)
}

func (s *stdOutLogger) ErrorsLogged() bool {
	return s.errorsLogged
}

// Auditor provides an interface for auditing one frugal file against another
// for breaking API changes.
type Auditor struct {
	logger    ValidationLogger
	oldFrugal *Frugal
	newFrugal *Frugal
}

// NewAuditor constructs an auditor that logs warnings and errors to
// standard output
func NewAuditor() *Auditor {
	return &Auditor{
		logger: &stdOutLogger{},
	}
}

// NewAuditorWithLogger constructs an auditor which uses the given logger
// to log warnings and errors
func NewAuditorWithLogger(logger ValidationLogger) *Auditor {
	return &Auditor{
		logger: logger,
	}
}

// Compare checks the contents of newFile for breaking changes with respect to
// oldFile
func (a *Auditor) Audit(oldFile, newFile string, includeDirs []string) error {
	newFrugal, err := ParseFrugal(newFile, includeDirs)
	if err != nil {
		return err
	}

	oldFrugal, err := ParseFrugal(oldFile, includeDirs)
	if err != nil {
		return err
	}

	a.oldFrugal = oldFrugal
	a.newFrugal = newFrugal

	a.checkScopes(oldFrugal.Scopes, newFrugal.Scopes)

	a.checkNamespaces(oldFrugal.Namespaces, newFrugal.Namespaces)
	a.checkConstants(oldFrugal.Constants, newFrugal.Constants)
	a.checkEnums(oldFrugal.Enums, newFrugal.Enums)
	a.checkStructLike(oldFrugal.Structs, newFrugal.Structs)
	a.checkStructLike(oldFrugal.Exceptions, newFrugal.Exceptions)
	a.checkStructLike(oldFrugal.Unions, newFrugal.Unions)
	a.checkServices(oldFrugal.Services, newFrugal.Services)

	if a.logger.ErrorsLogged() {
		return fmt.Errorf("FAILED: audit of %s against %s", newFile, oldFile)
	}
	return nil
}

// checkScopes requirements:
// Error:
// - Scopes removed
// - Scope prefix changed in any way other than renaming variables
// - Operation removed
// - Operation type changed
func (a *Auditor) checkScopes(oldScopes, newScopes []*Scope) {
	newMap := make(map[string]*Scope)
	for _, scope := range newScopes {
		newMap[scope.Name] = scope
	}

	for _, oldScope := range oldScopes {
		if newScope, ok := newMap[oldScope.Name]; ok {
			context := fmt.Sprintf("scope %s:", oldScope.Name)
			a.checkScopePrefix(oldScope.Prefix, newScope.Prefix, context)
			a.checkOperations(oldScope.Operations, newScope.Operations, context)
		} else {
			a.logger.LogError("missing scope:", oldScope.Name)
		}
	}
}

func (a *Auditor) checkScopePrefix(oldPrefix, newPrefix *ScopePrefix, context string) {
	// variable names in scope prefixes should be able to change,
	// but nothing else should be able to. Changing all the variables
	// to '{}' allows this
	oldNorm := normalizeScopePrefix(oldPrefix.String)
	newNorm := normalizeScopePrefix(newPrefix.String)
	if oldNorm != newNorm {
		a.logger.LogError(context, fmt.Sprintf("prefix changed: '%s' -> '%s'", oldNorm, newNorm))
	}
}

// normalizeScopePrefix changes variables in scope prefixes to '{}',
// i.e. "foo.{bar}.baz" -> "foo.{}.baz"
func normalizeScopePrefix(s string) string {
	separated := strings.Split(s, ".")
	for idx, piece := range separated {
		if strings.HasPrefix(piece, "{") && strings.HasSuffix(piece, "}") {
			separated[idx] = "{}"
		}
	}
	return strings.Join(separated, ".")
}

func (a *Auditor) checkOperations(oldOps, newOps []*Operation, context string) {
	newMap := make(map[string]*Operation)
	for _, op := range newOps {
		newMap[op.Name] = op
	}

	for _, oldOp := range oldOps {
		if newOp, ok := newMap[oldOp.Name]; ok {
			opContext := fmt.Sprintf("%s operation %s:", context, oldOp.Name)
			a.checkType(oldOp.Type, newOp.Type, false, opContext)
		} else {
			a.logger.LogError(context, "operation removed:", oldOp.Name)
		}
	}
}

// checkNamespaces requirements:
// Warning:
// - Namespace changed
// - Namespace removed
func (a *Auditor) checkNamespaces(oldNamespace, newNamespace []*Namespace) {
	newMap := make(map[string]*Namespace)
	for _, namespace := range newNamespace {
		newMap[namespace.Scope] = namespace
	}

	// These are warnings as namespace information isn't sent over the
	// network
	for _, oldNamespace := range oldNamespace {
		if newNamespace, ok := newMap[oldNamespace.Scope]; ok {
			if oldNamespace.Value != newNamespace.Value {
				a.logger.LogWarning("namespace changed:", oldNamespace.Scope)
			}
		} else {
			a.logger.LogWarning("namespace removed:", oldNamespace.Scope)
		}
	}
}

// checkConstants requirements
// Warning:
// - Constant removed
// - Constant value changed
// - Constant type changed
func (a *Auditor) checkConstants(oldConstants, newConstants []*Constant) {
	newMap := make(map[string]*Constant)
	for _, constant := range newConstants {
		newMap[constant.Name] = constant
	}

	// These are warnings as only the actual value is sent over the network
	for _, oldConstant := range oldConstants {
		if newConstant, ok := newMap[oldConstant.Name]; ok {
			context := fmt.Sprintf("constant %s:", oldConstant.Name)
			a.checkType(oldConstant.Type, newConstant.Type, true, context)
			if !reflect.DeepEqual(oldConstant.Value, newConstant.Value) {
				a.logger.LogWarning("constant value changed:", oldConstant.Name)
			}
		} else {
			a.logger.LogWarning("constant value removed:", oldConstant.Name)
		}
	}
}

// checkEnums requirements
// Warning:
// - Enum struct removed
// - Enum variant name changed
// Error:
// - Enum variant removed
func (a *Auditor) checkEnums(oldEnums, newEnums []*Enum) {
	newMap := make(map[string]*Enum)
	for _, enum := range newEnums {
		newMap[enum.Name] = enum
	}

	for _, oldEnum := range oldEnums {
		if newEnum, ok := newMap[oldEnum.Name]; ok {
			context := fmt.Sprintf("enum %s:", oldEnum.Name)
			a.checkEnumValues(oldEnum.Values, newEnum.Values, context)
		} else {
			a.logger.LogWarning("enum removed:", oldEnum.Name)
		}
	}
}

func (a *Auditor) checkEnumValues(oldValues, newValues []*EnumValue, context string) {
	newMap := make(map[int]*EnumValue)
	for _, value := range newValues {
		newMap[value.Value] = value
	}

	for _, oldValue := range oldValues {
		if newValue, ok := newMap[oldValue.Value]; ok {
			if oldValue.Name != newValue.Name {
				// enum variant names are allowed to change as
				// only the numeric value is sent over the
				// network
				a.logger.LogWarning("enum variant name changed:", oldValue.Name)
			}
		} else {
			a.logger.LogError(fmt.Sprintf("%s variant %s: removed with ID=%d",
				context, oldValue.Name, oldValue.Value))
		}
	}
}

// checkStructLike requirements:
// Warning:
// - Field name changed
// - Default value of field changed
// - Adding a field "in the middle"
// Error:
// - Struct removed
// - Presence modifier changed from optional/default to required (or vice versa)
// - Field type changed
// - Non-optional field removed
// - Addition of required field
func (a *Auditor) checkStructLike(oldStructs, newStructs []*Struct) {
	newMap := make(map[string]*Struct)
	for _, s := range newStructs {
		newMap[s.Name] = s
	}

	for _, oldStruct := range oldStructs {
		if newStruct, ok := newMap[oldStruct.Name]; ok {
			context := fmt.Sprintf("struct %s:", oldStruct.Name)
			a.checkFields(oldStruct.Fields, newStruct.Fields, context)
		} else {
			a.logger.LogError("missing struct:", oldStruct.Name)
		}
	}
}

// checkService requirements:
// Warning:
// - Name of argument changed
// - Name of exception changed
// - Adding argument "in the middle"
// - Adding exception "in the middle"
// Error:
// - Service inheritance changed
// - Service removed/renamed
// - Method removed/renamed
// - Method one-way changed
// - Method return type change
// - Method argument type changed
// - Method exception type changed
// - Adding an exception with a nil return value and no current exceptions
// - Removing an exception with a nil return value and only one current exception
func (a *Auditor) checkServices(oldServices, newServices []*Service) {
	newMap := make(map[string]*Service)
	for _, service := range newServices {
		newMap[service.Name] = service
	}

	for _, oldService := range oldServices {
		if newService, ok := newMap[oldService.Name]; ok {
			// It's fine to add inheritance, but not change it if it already exists
			if oldService.Extends != "" && oldService.Extends != newService.Extends {
				a.logger.LogError(fmt.Sprintf("service %s: extends changed: '%s' -> '%s'",
					oldService.Name, oldService.Extends, newService.Extends))
			}
			context := fmt.Sprintf("service %s:", oldService.Name)
			a.checkServiceMethods(oldService.Methods, newService.Methods, context)
		} else {
			a.logger.LogError("missing service:", oldService.Name)
		}
	}
}

func (a *Auditor) checkServiceMethods(oldMethods, newMethods []*Method, context string) {
	newMap := make(map[string]*Method)
	for _, method := range newMethods {
		newMap[method.Name] = method
	}

	for _, oldMethod := range oldMethods {
		if newMethod, ok := newMap[oldMethod.Name]; ok {
			methodContext := fmt.Sprintf("%s method %s:", context, oldMethod.Name)
			if oldMethod.Oneway != newMethod.Oneway {
				a.logger.LogError(methodContext, "one way modifier changed")
			}

			a.checkType(oldMethod.ReturnType, newMethod.ReturnType, false, methodContext+" return type:")

			a.checkFields(oldMethod.Arguments, newMethod.Arguments, methodContext)
			a.checkFields(oldMethod.Exceptions, newMethod.Exceptions, methodContext)

			// If the return type is nil and not exceptions exist,
			// the generated code doesn't expect anything to be
			// returned, so transitioning between the states of
			// "nothing can be returned" and "something can be
			// returned" isn't allowed
			if oldMethod.ReturnType == nil && len(oldMethod.Exceptions) == 0 && len(newMethod.Exceptions) > 0 {
				a.logger.LogError(methodContext, "can't add exceptions with nil return type")
			}

			if newMethod.ReturnType == nil && len(newMethod.Exceptions) == 0 && len(oldMethod.Exceptions) > 0 {
				a.logger.LogError(methodContext, "can't remove exceptions with nil return type")
			}
		} else {
			a.logger.LogError(context, "missing method: "+oldMethod.Name)
		}
	}
}

func (a *Auditor) checkFields(oldFields, newFields []*Field, context string) {
	oldMap := makeFieldsMap(oldFields)
	newMap := makeFieldsMap(newFields)

	min := int(^uint(0) >> 1)
	max := 0
	for _, oldField := range oldMap {
		if oldField.ID < min {
			min = oldField.ID
		}
		if oldField.ID > max {
			max = oldField.ID
		}

		fieldContext := fmt.Sprintf("%s field %s:", context, oldField.Name)
		if newField, ok := newMap[oldField.ID]; ok {
			a.checkType(oldField.Type, newField.Type, false, fieldContext)

			oldFieldReq := oldField.Modifier == Required
			newFieldReq := newField.Modifier == Required
			if oldFieldReq != newFieldReq {
				a.logger.LogError(fieldContext, fmt.Sprintf("field presence modifier changed: '%s' -> '%s'",
					oldField.Modifier.String(), newField.Modifier.String()))
			}

			if !reflect.DeepEqual(oldField.Default, newField.Default) {
				a.logger.LogWarning(fieldContext, "default value changed")
			}
			if oldField.Name != newField.Name {
				a.logger.LogWarning(fieldContext, "name changed")
			}
		} else if oldField.Modifier != Optional {
			a.logger.LogError(fieldContext, fmt.Sprintf("field removed with ID=%d", oldField.ID))
		}
	}

	for _, newField := range newMap {
		if _, ok := oldMap[newField.ID]; !ok {
			fieldContext := fmt.Sprintf("%s field %s:", context, newField.Name)
			// Adding a field "in the middle" is generally a sign
			// of field ID reuse, which isn't allowed
			if min < newField.ID && newField.ID < max {
				a.logger.LogWarning(fieldContext, fmt.Sprintf("added field in the middle with ID=%d", newField.ID))
			}

			if newField.Modifier == Required {
				a.logger.LogError(fieldContext, "added field is required")
			}
		}
	}
}

func makeFieldsMap(fields []*Field) map[int]*Field {
	fieldsMap := make(map[int]*Field)
	for _, field := range fields {
		fieldsMap[field.ID] = field
	}
	return fieldsMap
}

func (a *Auditor) checkType(oldType, newType *Type, warn bool, context string) {
	logMismatch := a.logger.LogWarning
	if !warn {
		logMismatch = a.logger.LogError
	}

	// guarding here makes recursive calls easier
	if oldType == nil || newType == nil {
		if oldType != newType {
			logMismatch(context, fmt.Sprintf("types not equal: '%v' -> '%v'", oldType, newType))
		}
		return
	}

	underlyingOldType := a.oldFrugal.UnderlyingType(oldType)
	underlyingNewType := a.newFrugal.UnderlyingType(newType)
	// TODO should this exclude the include name?
	if underlyingOldType.Name != underlyingNewType.Name {
		logMismatch(context, fmt.Sprintf("types not equal: '%s' -> '%s'",
			underlyingOldType.Name, underlyingNewType.Name))
		return
	}

	a.checkType(underlyingOldType.KeyType, underlyingNewType.KeyType, warn, context+" key type:")
	a.checkType(underlyingOldType.ValueType, underlyingNewType.ValueType, warn, context+" value type:")
}
