package parser

import (
	"fmt"
	"reflect"
	"strings"
)

type ValidationLogger interface {
	LogWarning(...string)
	LogError(...string)
	ErrorsLogged() bool
}

type StdOutLogger struct {
	errorsLogged bool
}

// TODO use tty colors to emphasize things?
func (s *StdOutLogger) LogWarning(warning ...string) {
	fmt.Println("WARNING:", warning)
}

func (s *StdOutLogger) LogError(errorMessage ...string) {
	s.errorsLogged = true
	fmt.Println("ERROR:", errorMessage)
}

func (s *StdOutLogger) ErrorsLogged() bool {
	return s.errorsLogged
}

type Auditor struct {
	logger    ValidationLogger
	oldFrugal *Frugal
	newFrugal *Frugal
}

func NewAuditor() *Auditor {
	return &Auditor{
		logger: &StdOutLogger{},
	}
}

func NewAuditorWithLogger(logger ValidationLogger) *Auditor {
	return &Auditor{
		logger: logger,
	}
}

func (a *Auditor) Compare(oldFile, newFile string) error {
	newFrugal, err := ParseFrugal(newFile)
	if err != nil {
		return err
	}

	oldFrugal, err := ParseFrugal(oldFile)
	if err != nil {
		return err
	}

	a.oldFrugal = oldFrugal
	a.newFrugal = newFrugal

	a.checkScopes(oldFrugal.Scopes, newFrugal.Scopes)

	a.checkNamespaces(oldFrugal.Thrift.Namespaces, newFrugal.Thrift.Namespaces)
	a.checkConstants(oldFrugal.Thrift.Constants, newFrugal.Thrift.Constants)
	a.checkEnums(oldFrugal.Thrift.Enums, newFrugal.Thrift.Enums)
	a.checkStructLike(oldFrugal.Thrift.Structs, newFrugal.Thrift.Structs)
	a.checkStructLike(oldFrugal.Thrift.Exceptions, newFrugal.Thrift.Exceptions)
	a.checkStructLike(oldFrugal.Thrift.Unions, newFrugal.Thrift.Unions)
	a.checkServices(oldFrugal.Thrift.Services, newFrugal.Thrift.Services)

	if a.logger.ErrorsLogged() {
		return fmt.Errorf("FAILED: audit of %s against %s", newFile, oldFile)
	}
	return nil
}

func (a *Auditor) checkScopes(old, new []*Scope) {
	newMap := make(map[string]*Scope)
	for _, scope := range new {
		newMap[scope.Name] = scope
	}

	for _, oldScope := range old {
		if newScope, ok := newMap[oldScope.Name]; ok {
			context := fmt.Sprintf("scope %s:", oldScope.Name)
			a.checkScopePrefix(oldScope.Prefix, newScope.Prefix, context)
			a.checkOperations(oldScope.Operations, newScope.Operations, context)
		} else {
			a.logger.LogError("missing scope:", oldScope.Name)
		}
	}
}

func (a *Auditor) checkScopePrefix(old, new *ScopePrefix, context string) {
	oldNorm := normalizeScopePrefix(old.String)
	newNorm := normalizeScopePrefix(new.String)
	if oldNorm != newNorm {
		a.logger.LogError(context, fmt.Sprintf("prefix changed: '%s' -> '%s'", oldNorm, newNorm))
	}
}

func normalizeScopePrefix(s string) string {
	separated := strings.Split(s, ".")
	for idx, piece := range separated {
		if strings.HasPrefix(piece, "{") && strings.HasSuffix(piece, "}") {
			separated[idx] = "{}"
		}
	}
	return strings.Join(separated, ".")
}

func (a *Auditor) checkOperations(old, new []*Operation, context string) {
	newMap := make(map[string]*Operation)
	for _, op := range new {
		newMap[op.Name] = op
	}

	for _, oldOp := range old {
		if newOp, ok := newMap[oldOp.Name]; ok {
			opContext := fmt.Sprintf("%s operation %s:", context, oldOp.Name)
			a.checkType(oldOp.Type, newOp.Type, false, opContext)
		} else {
			a.logger.LogError(context, "operation removed:", oldOp.Name)
		}
	}
}

func (a *Auditor) checkNamespaces(old, new []*Namespace) {
	newMap := make(map[string]*Namespace)
	for _, namespace := range new {
		newMap[namespace.Scope] = namespace
	}

	for _, oldNamespace := range old {
		if newNamespace, ok := newMap[oldNamespace.Scope]; ok {
			if oldNamespace.Value != newNamespace.Value {
				a.logger.LogWarning("namespace changed:", oldNamespace.Scope)
			}
		} else {
			a.logger.LogWarning("namespace removed:", oldNamespace.Scope)
		}
	}
}

func (a *Auditor) checkConstants(old, new []*Constant) {
	newMap := make(map[string]*Constant)
	for _, constant := range new {
		newMap[constant.Name] = constant
	}

	for _, oldConstant := range old {
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

func (a *Auditor) checkEnums(old, new []*Enum) {
	newMap := make(map[string]*Enum)
	for _, enum := range new {
		newMap[enum.Name] = enum
	}

	for _, oldEnum := range old {
		if newEnum, ok := newMap[oldEnum.Name]; ok {
			context := fmt.Sprintf("enum %s:", oldEnum.Name)
			a.checkEnumValues(oldEnum.Values, newEnum.Values, context)
		} else {
			a.logger.LogWarning("enum removed:", oldEnum.Name)
		}
	}
}

func (a *Auditor) checkEnumValues(old, new []*EnumValue, context string) {
	newMap := make(map[int]*EnumValue)
	for _, value := range new {
		newMap[value.Value] = value
	}

	for _, oldValue := range old {
		if newValue, ok := newMap[oldValue.Value]; ok {
			if oldValue.Name != newValue.Name {
				a.logger.LogWarning("enum variant name changed:", oldValue.Name)
			}
		} else {
			a.logger.LogError(fmt.Sprintf("%s variant %s: removed with ID=%d",
				context, oldValue.Name, oldValue.Value))
		}
	}
}

func (a *Auditor) checkStructLike(old, new []*Struct) {
	newMap := make(map[string]*Struct)
	for _, s := range new {
		newMap[s.Name] = s
	}

	for _, oldStruct := range old {
		if newStruct, ok := newMap[oldStruct.Name]; ok {
			context := fmt.Sprintf("struct %s:", oldStruct.Name)
			a.checkFields(oldStruct.Fields, newStruct.Fields, context)
		} else {
			a.logger.LogError("missing struct:", oldStruct.Name)
		}
	}
}

func (a *Auditor) checkServices(old, new []*Service) {
	newMap := make(map[string]*Service)
	for _, service := range new {
		newMap[service.Name] = service
	}

	for _, oldService := range old {
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

func (a *Auditor) checkServiceMethods(old, new []*Method, context string) {
	newMap := make(map[string]*Method)
	for _, method := range new {
		newMap[method.Name] = method
	}

	for _, oldMethod := range old {
		if newMethod, ok := newMap[oldMethod.Name]; ok {
			// one way must be equal
			methodContext := fmt.Sprintf("%s method %s:", context, oldMethod.Name)
			if oldMethod.Oneway != newMethod.Oneway {
				a.logger.LogError(methodContext, "one way modifier changed")
			}

			// return types must be equal
			a.checkType(oldMethod.ReturnType, newMethod.ReturnType, false, methodContext+" return type:")

			a.checkFields(oldMethod.Arguments, newMethod.Arguments, methodContext)
			a.checkFields(oldMethod.Exceptions, newMethod.Exceptions, methodContext)

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

func (a *Auditor) checkFields(old, new []*Field, context string) {
	oldMap := makeFieldsMap(old)
	newMap := makeFieldsMap(new)

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

func (a *Auditor) checkType(old, new *Type, warn bool, context string) {
	logMismatch := a.logger.LogWarning
	if !warn {
		logMismatch = a.logger.LogError
	}

	// guarding here makes recursive calls easier
	if old == nil || new == nil {
		if old != new {
			logMismatch(context, fmt.Sprintf("types not equal: '%v' -> '%v'", old, new))
		}
		return
	}

	underlyingOldType := a.oldFrugal.UnderlyingType(old)
	underlyingNewType := a.newFrugal.UnderlyingType(new)
	// TODO should this exclude the include name?
	if underlyingOldType.Name != underlyingNewType.Name {
		logMismatch(context, fmt.Sprintf("types not equal: '%s' -> '%s'",
			underlyingOldType.Name, underlyingNewType.Name))
		return
	}

	a.checkType(underlyingOldType.KeyType, underlyingNewType.KeyType, warn, context+" key type:")
	a.checkType(underlyingOldType.ValueType, underlyingNewType.ValueType, warn, context+" value type:")
}
