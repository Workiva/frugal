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

package java

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

const (
	lang                        = "java"
	defaultOutputDir            = "gen-java"
	generatedAnnotations        = "generated_annotations"
	useVendorOption             = "use_vendor"
	suppressDeprecatedLogging   = "suppress_deprecated_logging"
	tab                         = "\t"
	tabtab                      = tab + tab
	tabtabtab                   = tab + tab + tab
	tabtabtabtab                = tab + tab + tab + tab
	tabtabtabtabtab             = tab + tab + tab + tab + tab
	tabtabtabtabtabtab          = tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtab       = tab + tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtabtab    = tab + tab + tab + tab + tab + tab + tab + tab
	tabtabtabtabtabtabtabtabtab = tab + tab + tab + tab + tab + tab + tab + tab + tab
)

type Generator struct {
	*generator.BaseGenerator
	time      time.Time
	outputDir string
}

func NewGenerator(options map[string]string) generator.LanguageGenerator {
	return &Generator{
		&generator.BaseGenerator{Options: options},
		globals.Now,
		"",
	}
}

// ADTs would be really nice
type IsSetType int64

const (
	IsSetNone IsSetType = iota
	IsSetBitfield
	IsSetBitSet
)

// This is how java does isset checks, I'm open to changing this.
func (g *Generator) getIsSetType(s *parser.Struct) (IsSetType, string) {
	primitiveCount := 0
	for _, field := range s.Fields {
		if g.isJavaPrimitive(field.Type) {
			primitiveCount += 1
		}
	}

	switch {
	case primitiveCount == 0:
		return IsSetNone, ""
	case 0 < primitiveCount && primitiveCount <= 8:
		return IsSetBitfield, "byte"
	case 8 < primitiveCount && primitiveCount <= 16:
		return IsSetBitfield, "short"
	case 16 < primitiveCount && primitiveCount <= 32:
		return IsSetBitfield, "int"
	case 32 < primitiveCount && primitiveCount <= 64:
		return IsSetBitfield, "long"
	default:
		return IsSetBitSet, ""
	}
}

// Suppress deprecated API usage warning logging
func (g *Generator) suppressDeprecatedLogging() bool {
	_, ok := g.Options[suppressDeprecatedLogging]
	return ok
}

func (g *Generator) SetupGenerator(outputDir string) error {
	g.outputDir = outputDir
	return nil
}

func (g *Generator) TeardownGenerator() error {
	return nil
}

func (g *Generator) GenerateConstantsContents(constants []*parser.Constant) error {
	if len(constants) == 0 {
		return nil
	}

	contents := ""

	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation("")
	}
	contents += fmt.Sprintf("public class %sConstants {\n", g.Frugal.Name)

	for _, constant := range constants {
		val := g.generateConstantValueWrapper(constant.Name, constant.Type, constant.Value, true, true, tab)
		contents += fmt.Sprintf("%s\n", val)
	}

	contents += "}\n"

	file, err := g.GenerateFile(fmt.Sprintf("%sConstants", g.Frugal.Name), g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}
	_, err = file.WriteString(contents)

	return err
}

// generateConstantValueWrapper generates a constant value. Unlike other languages,
// constants can't be initialized in a single expression, so temp variables
// are needed. Due to this, the entire constant, not just the value, is
// generated.
func (g *Generator) generateConstantValueWrapper(fieldName string, t *parser.Type, value interface{}, declare, needsStatic bool, indent string) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	contents := indent

	if needsStatic {
		contents += "public static final "
	}
	if declare {
		contents += g.getJavaTypeFromThriftType(underlyingType) + " "
	}

	if value == nil {
		return fmt.Sprintf("%s%s = %s;\n", contents, fieldName, "null")
	}

	if underlyingType.IsPrimitive() || g.Frugal.IsEnum(underlyingType) {
		_, val := g.generateConstantValueRec(t, value, indent)
		return fmt.Sprintf("%s%s = %s;\n", contents, fieldName, val)
	} else if g.Frugal.IsStruct(underlyingType) {
		s := g.Frugal.FindStruct(underlyingType)
		if s == nil {
			panic("no struct for type " + underlyingType.Name)
		}

		contents += fmt.Sprintf("%s = new %s();\n", fieldName, g.getJavaTypeFromThriftType(underlyingType))

		ind := indent
		if needsStatic {
			contents += ind + "static {\n"
			ind += tab
		}

		for _, pair := range value.([]parser.KeyValue) {
			name := pair.KeyToString()
			for _, field := range s.Fields {
				if name == field.Name {
					preamble, val := g.generateConstantValueRec(field.Type, pair.Value, ind)
					contents += preamble
					contents += ind + fmt.Sprintf("%s.set%s(%s);\n", fieldName, strings.Title(name), val)
				}
			}
		}

		if needsStatic {
			contents += indent + "}\n"
		}

		return contents
	} else if underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf("%s = new ArrayList<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := indent
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v, ind)
				contents += preamble
				contents += ind + fmt.Sprintf("%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += indent + "}\n"
			}
			return contents
		case "set":
			contents += fmt.Sprintf("%s = new HashSet<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := indent
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v, ind)
				contents += preamble
				contents += ind + fmt.Sprintf("%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += indent + "}\n"
			}
			return contents
		case "map":
			contents += fmt.Sprintf("%s = new HashMap<%s,%s>();\n",
				fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)),
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := indent
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, pair := range value.([]parser.KeyValue) {
				preamble, key := g.generateConstantValueRec(underlyingType.KeyType, pair.Key, ind)
				contents += preamble
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, pair.Value, ind)
				contents += preamble
				contents += ind + fmt.Sprintf("%s.put(%s, %s);\n", fieldName, key, val)
			}

			if needsStatic {
				contents += indent + "}\n"
			}
			return contents
		}
	}

	panic("Unrecognized type: " + underlyingType.Name)
}

func (g *Generator) generateEnumConstValue(frugal *parser.Frugal, pieces []string, t *parser.Type) (string, bool) {
	for _, enum := range frugal.Enums {
		if pieces[0] == enum.Name {
			for _, value := range enum.Values {
				if pieces[1] == value.Name {
					return fmt.Sprintf("%s.%s", g.getJavaTypeFromThriftType(t), value.Name), true
				}
			}
			panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
		}
	}
	return "", false
}

func (g *Generator) generateEnumConstFromValue(t *parser.Type, value int) string {
	frugal := g.Frugal
	if t.IncludeName() != "" {
		// The type is from an include
		frugal = g.Frugal.ParsedIncludes[t.IncludeName()]
	}

	for _, enum := range frugal.Enums {
		if enum.Name == t.ParamName() {
			// found the enum
			for _, enumValue := range enum.Values {
				if enumValue.Value == value {
					// found the value
					return fmt.Sprintf("%s.%s", g.getJavaTypeFromThriftType(t), enumValue.Name)
				}
			}
		}
	}

	panic("value not found")
}

// quote creates a Java string literal for a string.
func (g *Generator) quote(s string) string {
	// For now, just use Go quoting rules.
	return strconv.Quote(s)
}

func (g *Generator) namespaceForInclude(includeName string) string {
	if includeName == "" {
		return ""
	}

	namespace := g.Frugal.NamespaceForInclude(includeName, lang)
	if namespace == nil {
		return ""
	}

	if vendorPath, _ := namespace.Annotations.Vendor(); vendorPath != "" && g.UseVendor() && g.isVendoredInclude(includeName) {
		return vendorPath
	}

	return namespace.Value
}

func (g *Generator) generateConstantValueRec(t *parser.Type, value interface{}, indent string) (string, string) {
	underlyingType := g.Frugal.UnderlyingType(t)

	// If the value being referenced is of type Identifier, it's referencing
	// another constant.
	identifier, ok := value.(parser.Identifier)
	if ok {
		idCtx := g.Frugal.ContextFromIdentifier(identifier)
		switch idCtx.Type {
		case parser.LocalConstant:
			return "", fmt.Sprintf("%sConstants.%s", g.Frugal.Name, idCtx.Constant.Name)
		case parser.LocalEnum:
			return "", fmt.Sprintf("%s.%s", idCtx.Enum.Name, idCtx.EnumValue.Name)
		case parser.IncludeConstant:
			include := idCtx.Include.Name
			if namespace := g.namespaceForInclude(include); namespace != "" {
				include = namespace
			}
			return "", fmt.Sprintf("%s.%sConstants.%s", include, idCtx.Include.Name, idCtx.Constant.Name)
		case parser.IncludeEnum:
			include := idCtx.Include.Name
			if namespace := g.namespaceForInclude(include); namespace != "" {
				include = namespace
			}
			return "", fmt.Sprintf("%s.%s.%s", include, idCtx.Enum.Name, idCtx.EnumValue.Name)
		default:
			panic(fmt.Sprintf("The Identifier %s has unexpected type %d", identifier, idCtx.Type))
		}
	}

	if underlyingType.IsPrimitive() {
		switch underlyingType.Name {
		case "bool":
			return "", fmt.Sprintf("%v", value)
		case "byte", "i8":
			return "", fmt.Sprintf("(byte)%v", value)
		case "i16":
			return "", fmt.Sprintf("(short)%v", value)
		case "i32":
			return "", fmt.Sprintf("%v", value)
		case "i64":
			return "", fmt.Sprintf("%vL", value)
		case "double":
			return "", fmt.Sprintf("%v", value)
		case "string":
			return "", g.quote(value.(string))
		case "binary":
			return "", fmt.Sprintf("java.nio.ByteBuffer.wrap(\"%v\".getBytes())", value)
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return "", g.generateEnumConstFromValue(underlyingType, int(value.(int64)))
	}
	elem := g.GetElem()
	preamble := g.generateConstantValueWrapper(elem, t, value, true, false, indent)
	return preamble, elem

}

func (g *Generator) GenerateTypeDef(*parser.TypeDef) error {
	// No typedefs in java
	return nil
}

func (g *Generator) GenerateEnum(enum *parser.Enum) error {
	contents := ""
	contents += fmt.Sprintf("public enum %s implements org.apache.thrift.TEnum {\n", enum.Name)
	for idx, value := range enum.Values {
		terminator := ","
		if idx == len(enum.Values)-1 {
			terminator = ";"
		}
		contents += g.generateCommentWithDeprecated(value.Comment, tab, value.Annotations)
		contents += tab + fmt.Sprintf("%s(%d)%s\n", value.Name, value.Value, terminator)
	}
	contents += "\n"

	contents += tab + "private final int value;\n\n"
	contents += tab + fmt.Sprintf("private %s(int value) {\n", enum.Name)
	contents += tabtab + "this.value = value;\n"
	contents += tab + "}\n\n"

	contents += tab + "public int getValue() {\n"
	contents += tabtab + "return value;\n"
	contents += tab + "}\n\n"

	contents += tab + fmt.Sprintf("public static %s findByValue(int value) {\n", enum.Name)
	contents += tabtab + "switch (value) {\n"
	for _, value := range enum.Values {
		contents += tabtabtab + fmt.Sprintf("case %d:\n", value.Value)
		contents += tabtabtabtab + fmt.Sprintf("return %s;\n", value.Name)
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "return null;\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n"

	contents += "}\n"

	file, err := g.GenerateFile(enum.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.GenerateDocStringComment(file); err != nil {
		return err
	}
	if _, err = file.WriteString("\n"); err != nil {
		return err
	}
	if err = g.generatePackage(file); err != nil {
		return err
	}
	if _, err = file.WriteString("\n\n"); err != nil {
		return err
	}
	if err = g.GenerateEnumImports(file); err != nil {
		return err
	}

	_, err = file.WriteString(contents)

	return err
}

func (g *Generator) initStructFile(file *os.File) error {
	if err := g.GenerateDocStringComment(file); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}
	if err := g.generatePackage(file); err != nil {
		return err
	}

	if _, err := file.WriteString("\n\n"); err != nil {
		return err
	}

	return g.GenerateStructImports(file)
}

func (g *Generator) GenerateStruct(s *parser.Struct) error {
	file, err := g.GenerateFile(s.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}

	_, err = file.WriteString(g.generateStruct(s, false, false, ""))
	return err
}

func (g *Generator) GenerateUnion(union *parser.Struct) error {
	// I have no idea why java uses this convention as the fields really
	// should be optional...
	for _, field := range union.Fields {
		field.Modifier = parser.Default
	}

	file, err := g.GenerateFile(union.Name, g.outputDir, generator.ObjectFile)
	defer file.Close()
	if err != nil {
		return err
	}

	if err = g.initStructFile(file); err != nil {
		return err
	}

	contents := g.generateUnion(union, false, false)
	_, err = file.WriteString(contents)
	return err
}

func (g *Generator) generateUnion(union *parser.Struct, isArg, isResult bool) string {
	contents := ""

	if g.includeGeneratedAnnotation() && !isArg && !isResult {
		contents += g.generatedAnnotation("")
	}

	static := ""
	if isArg || isResult {
		static = "static "
	}
	contents += fmt.Sprintf("public %sclass %s extends org.apache.thrift.TUnion<%s, %s._Fields> {\n",
		static, union.Name, union.Name, union.Name)

	contents += g.generateDescriptors(union, tab)
	contents += g.generateFieldsEnum(union, tab)
	contents += g.generateUnionConstructors(union, tab)
	contents += g.generateUnionFieldConstructors(union, tab)
	contents += g.generateUnionCheckType(union, tab)

	contents += g.generateUnionStandardRead(union, tab)
	contents += g.generateUnionStandardWrite(union, tab)
	contents += g.generateUnionTupleRead(union, tab)
	contents += g.generateUnionTupleWrite(union, tab)

	contents += g.generateUnionGetDescriptors(union, tab)
	contents += g.generateUnionFieldForId(tab)
	contents += g.generateUnionGetSetFields(union, tab)
	contents += g.generateUnionIsSetFields(union, tab)

	contents += g.generateUnionEquals(union, tab)
	contents += g.generateUnionCompareTo(union, tab)
	contents += g.generateUnionHashCode(union, tab)
	contents += g.generateWriteObject(union, tab)
	contents += g.generateReadObject(union, tab)

	contents += "}\n"
	return contents
}

func (g *Generator) generateUnionConstructors(union *parser.Struct, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("public %s() {\n", union.Name)
	contents += indent + tab + "super();\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("public %s(_Fields setField, Object value) {\n", union.Name)
	contents += indent + tab + "super(setField, value);\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("public %s(%s other) {\n", union.Name, union.Name)
	contents += indent + tab + "super(other);\n"
	contents += indent + "}\n"

	contents += indent + fmt.Sprintf("public %s deepCopy() {\n", union.Name)
	contents += indent + tab + fmt.Sprintf("return new %s(this);\n", union.Name)
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionFieldConstructors(union *parser.Struct, indent string) string {
	contents := ""

	for _, field := range union.Fields {
		contents += g.generateCommentWithDeprecated(field.Comment, indent, field.Annotations)
		contents += indent + fmt.Sprintf("public static %s %s(%s value) {\n",
			union.Name, field.Name, g.getJavaTypeFromThriftType(field.Type))
		contents += indent + tab + fmt.Sprintf("%s x = new %s();\n", union.Name, union.Name)
		contents += indent + tab + fmt.Sprintf("x.set%s(value);\n", strings.Title(field.Name))
		contents += indent + tab + "return x;\n"
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionCheckType(union *parser.Struct, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + "protected void checkType(_Fields setField, Object value) throws ClassCastException {\n"
	contents += indent + tab + "switch (setField) {\n"
	for _, field := range union.Fields {
		fieldType := containerType(g.getJavaTypeFromThriftType(field.Type))
		unparametrizedType := containerType(g.getUnparametrizedJavaType(field.Type))
		contents += indent + tabtab + fmt.Sprintf("case %s:\n", toConstantName(field.Name))
		contents += indent + tabtabtab + fmt.Sprintf("if (value instanceof %s) {\n", unparametrizedType)
		contents += indent + tabtabtabtab + "break;\n"
		contents += indent + tabtabtab + "}\n"
		contents += indent + tabtabtab + fmt.Sprintf("throw new ClassCastException(\"Was expecting value of type %s for field '%s', but got \" + value.getClass().getSimpleName());\n",
			fieldType, field.Name)
	}
	contents += indent + tabtab + "default:\n"
	contents += indent + tabtabtab + fmt.Sprintf("throw new IllegalArgumentException(\"Unknown field id \" + setField);\n")
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardRead(union *parser.Struct, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + "protected Object standardSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, org.apache.thrift.protocol.TField field) throws org.apache.thrift.TException {\n"
	contents += indent + tab + "_Fields setField = _Fields.findByThriftId(field.id);\n"
	contents += indent + tab + "if (setField != null) {\n"
	contents += indent + tabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += indent + tabtabtab + fmt.Sprintf("case %s:\n", constantName)
		contents += indent + tabtabtabtab + fmt.Sprintf("if (field.type == %s_FIELD_DESC.type) {\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, indent+tabtabtabtabtab)
		contents += indent + tabtabtabtabtab + fmt.Sprintf("return %s;\n", field.Name)
		contents += indent + tabtabtabtab + "} else {\n"
		contents += indent + tabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
		contents += indent + tabtabtabtabtab + "return null;\n"
		contents += indent + tabtabtabtab + "}\n"
	}
	contents += indent + tabtabtab + "default:\n"
	contents += indent + tabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tab + "} else {\n"
	contents += indent + tabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
	contents += indent + tabtab + "return null;\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardWrite(union *parser.Struct, indent string) string {
	return g.generateUnionWrite(union, "standard", indent)
}

func (g *Generator) generateUnionTupleRead(union *parser.Struct, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + "protected Object tupleSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, short fieldID) throws org.apache.thrift.TException {\n"
	contents += indent + tab + "_Fields setField = _Fields.findByThriftId(fieldID);\n"
	contents += indent + tab + "if (setField != null) {\n"
	contents += indent + tabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += indent + tabtabtab + fmt.Sprintf("case %s:\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, indent+tabtabtabtab)
		contents += indent + tabtabtabtab + fmt.Sprintf("return %s;\n", field.Name)
	}
	contents += indent + tabtabtab + "default:\n"
	contents += indent + tabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tab + "} else {\n"
	contents += indent + tabtab + "throw new TProtocolException(\"Couldn't find a field with field id \" + fieldID);\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionTupleWrite(union *parser.Struct, indent string) string {
	return g.generateUnionWrite(union, "tuple", indent)
}

func (g *Generator) generateUnionWrite(union *parser.Struct, scheme string, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + fmt.Sprintf("protected void %sSchemeWriteValue(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {\n", scheme)
	contents += indent + tab + "switch (setField_) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		javaContainerType := containerType(g.getJavaTypeFromThriftType(field.Type))
		contents += indent + tabtab + fmt.Sprintf("case %s:\n", constantName)
		contents += indent + tabtabtab + fmt.Sprintf("%s %s = (%s)value_;\n", javaContainerType, field.Name, javaContainerType)
		contents += g.generateWriteFieldRec(field, false, false, indent+tabtabtab)
		contents += indent + tabtabtab + "return;\n"
	}
	contents += indent + tabtab + "default:\n"
	contents += indent + tabtabtab + "throw new IllegalStateException(\"Cannot write union with unknown field \" + setField_);\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionGetDescriptors(union *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + "protected org.apache.thrift.protocol.TField getFieldDesc(_Fields setField) {\n"
	contents += indent + tab + "switch (setField) {\n"

	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += indent + tabtab + fmt.Sprintf("case %s:\n", constantName)
		contents += indent + tabtabtab + fmt.Sprintf("return %s_FIELD_DESC;\n", constantName)
	}

	contents += indent + tabtab + "default:\n"
	contents += indent + tabtabtab + "throw new IllegalArgumentException(\"Unknown field id \" + setField);\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	contents += indent + "@Override\n"
	contents += indent + "protected org.apache.thrift.protocol.TStruct getStructDesc() {\n"
	contents += indent + tab + "return STRUCT_DESC;\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateUnionFieldForId(indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + "protected _Fields enumForId(short id) {\n"
	contents += indent + tab + "return _Fields.findByThriftIdOrThrow(id);\n"
	contents += indent + "}\n\n"

	contents += indent + "public _Fields fieldForId(int fieldId) {\n"
	contents += indent + tab + "return _Fields.findByThriftId(fieldId);\n"
	contents += indent + "}\n\n\n"
	return contents
}

func (g *Generator) generateUnionGetSetFields(union *parser.Struct, indent string) string {
	contents := ""

	for _, field := range union.Fields {
		titleName := strings.Title(field.Name)
		constantName := toConstantName(field.Name)
		javaType := g.getJavaTypeFromThriftType(field.Type)

		// get
		if field.Annotations.IsDeprecated() {
			contents += indent + "@Deprecated\n"
		}
		contents += indent + fmt.Sprintf("public %s get%s() {\n", javaType, titleName)
		contents += indent + tab + fmt.Sprintf("if (getSetField() == _Fields.%s) {\n", constantName)
		contents += indent + tabtab + fmt.Sprintf("return (%s)getFieldValue();\n", containerType(javaType))
		contents += indent + tab + "} else {\n"
		contents += indent + tabtab + fmt.Sprintf("throw new RuntimeException(\"Cannot get field '%s' because union is currently set to \" + getFieldDesc(getSetField()).name);\n", field.Name)
		contents += indent + tab + "}\n"
		contents += indent + "}\n\n"

		// set
		if field.Annotations.IsDeprecated() {
			contents += indent + "@Deprecated\n"
		}
		contents += indent + fmt.Sprintf("public void set%s(%s value) {\n", titleName, javaType)
		if !g.isJavaPrimitive(field.Type) {
			contents += indent + tab + "if (value == null) throw new NullPointerException();\n"
		}
		contents += indent + tab + fmt.Sprintf("setField_ = _Fields.%s;\n", constantName)
		contents += indent + tab + "value_ = value;\n"
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionIsSetFields(union *parser.Struct, indent string) string {
	contents := ""

	for _, field := range union.Fields {
		if field.Annotations.IsDeprecated() {
			contents += indent + "@Deprecated\n"
		}
		contents += indent + fmt.Sprintf("public boolean isSet%s() {\n", strings.Title(field.Name))
		contents += indent + tab + fmt.Sprintf("return setField_ == _Fields.%s;\n", toConstantName(field.Name))
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionEquals(union *parser.Struct, indent string) string {
	contents := "\n"

	contents += indent + "public boolean equals(Object other) {\n"
	contents += indent + tab + fmt.Sprintf("if (other instanceof %s) {\n", union.Name)
	contents += indent + tabtab + fmt.Sprintf("return equals((%s)other);\n", union.Name)
	contents += indent + tab + "} else {\n"
	contents += indent + tabtab + "return false;\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("public boolean equals(%s other) {\n", union.Name)
	contents += indent + tab + "return other != null && getSetField() == other.getSetField() && getFieldValue().equals(other.getFieldValue());\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnionCompareTo(union *parser.Struct, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + fmt.Sprintf("public int compareTo(%s other) {\n", union.Name)
	contents += indent + tab + "int lastComparison = org.apache.thrift.TBaseHelper.compareTo(getSetField(), other.getSetField());\n"
	contents += indent + tab + "if (lastComparison == 0) {\n"
	contents += indent + tabtab + "return org.apache.thrift.TBaseHelper.compareTo(getFieldValue(), other.getFieldValue());\n"
	contents += indent + tab + "}\n"
	contents += indent + tab + "return lastComparison;\n"
	contents += indent + "}\n\n\n"

	return contents
}

func (g *Generator) generateUnionHashCode(union *parser.Struct, indent string) string {
	contents := ""

	contents += indent + "@Override\n"
	contents += indent + "public int hashCode() {\n"
	contents += indent + tab + "List<Object> list = new ArrayList<Object>();\n"
	contents += indent + tab + "list.add(this.getClass().getName());\n"
	contents += indent + tab + "org.apache.thrift.TFieldIdEnum setField = getSetField();\n"
	contents += indent + tab + "if (setField != null) {\n"
	contents += indent + tabtab + "list.add(setField.getThriftFieldId());\n"
	contents += indent + tabtab + "Object value = getFieldValue();\n"
	contents += indent + tabtab + "if (value instanceof org.apache.thrift.TEnum) {\n"
	contents += indent + tabtabtab + "list.add(((org.apache.thrift.TEnum)getFieldValue()).getValue());\n"
	contents += indent + tabtab + "} else {\n"
	contents += indent + tabtabtab + "list.add(value);\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tab + "}\n"
	contents += indent + tab + "return list.hashCode();\n"
	contents += indent + "}\n"

	return contents
}

func (g *Generator) GenerateException(exception *parser.Struct) error {
	return g.GenerateStruct(exception)
}

// generateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) generateServiceArgsResults(service *parser.Service, indent string) string {
	contents := ""
	for _, s := range g.GetServiceMethodTypes(service) {
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				field.Modifier = parser.Default
			}
		}
		contents += g.generateStruct(s, strings.HasSuffix(s.Name, "_args"), strings.HasSuffix(s.Name, "_result"), indent)
		contents += "\n"
	}
	return contents
}

func (g *Generator) generateStruct(s *parser.Struct, isArg, isResult bool, indent string) string {
	contents := ""

	if s.Comment != nil {
		contents += g.GenerateBlockComment(s.Comment, indent)
	}
	if g.includeGeneratedAnnotation() && !isArg && !isResult {
		contents += g.generatedAnnotation(indent)
	}
	static := ""
	if isArg || isResult {
		static = "static "
	}
	exception := ""
	if s.Type == parser.StructTypeException {
		exception = "extends TException "
	}
	contents += fmt.Sprintf("%spublic %sclass %s %simplements org.apache.thrift.TBase<%s, %s._Fields>, java.io.Serializable, Cloneable, Comparable<%s> {\n",
		indent, static, s.Name, exception, s.Name, s.Name, s.Name)

	nestedIndent := indent + tab

	contents += g.generateDescriptors(s, nestedIndent)

	contents += g.generateSchemeMap(s, nestedIndent)

	contents += g.generateInstanceVars(s, nestedIndent)

	contents += g.generateFieldsEnum(s, nestedIndent)

	contents += g.generateIsSetVars(s, nestedIndent)

	contents += g.generateDefaultConstructor(s, nestedIndent)
	contents += g.generateFullConstructor(s, nestedIndent)
	contents += g.generateCopyConstructor(s, nestedIndent)
	contents += g.generateDeepCopyMethod(s, nestedIndent)
	contents += g.generateClear(s, nestedIndent)

	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if underlyingType.IsContainer() {
			contents += g.generateContainerGetSize(field, nestedIndent)
			contents += g.generateContainerIterator(field, nestedIndent)
			contents += g.generateContainerAddTo(field, nestedIndent)
		}

		contents += g.generateGetField(field, nestedIndent)
		contents += g.generateSetField(s.Name, field, nestedIndent)
		contents += g.generateUnsetField(s, field, nestedIndent)
		contents += g.generateIsSetField(s, field, nestedIndent)
		contents += g.generateSetIsSetField(s, field, nestedIndent)
	}

	contents += g.generateSetValue(s, nestedIndent)
	contents += g.generateGetValue(s, nestedIndent)
	contents += g.generateIsSetValue(s, nestedIndent)

	contents += g.generateEquals(s, nestedIndent)
	contents += g.generateHashCode(s, nestedIndent)
	contents += g.generateCompareTo(s, nestedIndent)

	contents += g.generateFieldForId(s, nestedIndent)
	contents += g.generateReadWrite(s, nestedIndent)

	contents += g.generateToString(s, nestedIndent)
	contents += g.generateValidate(s, nestedIndent)

	contents += g.generateWriteObject(s, nestedIndent)
	contents += g.generateReadObject(s, nestedIndent)

	contents += g.generateStandardScheme(s, isResult, nestedIndent)
	contents += g.generateTupleScheme(s, nestedIndent)

	contents += indent + "}\n"
	return contents
}

func (g *Generator) generateDescriptors(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("private static final org.apache.thrift.protocol.TStruct STRUCT_DESC = new org.apache.thrift.protocol.TStruct(\"%s\");\n\n",
		s.Name)
	for _, field := range s.Fields {
		contents += indent + fmt.Sprintf("private static final org.apache.thrift.protocol.TField %s_FIELD_DESC = new org.apache.thrift.protocol.TField(\"%s\", %s, (short)%d);\n",
			toConstantName(field.Name), field.Name, g.getTType(field.Type), field.ID)
	}
	contents += "\n"
	return contents
}

func (g *Generator) generateSchemeMap(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "private static final Map<Class<? extends IScheme>, SchemeFactory> schemes = new HashMap<Class<? extends IScheme>, SchemeFactory>();\n"
	contents += indent + "static {\n"
	contents += indent + tab + fmt.Sprintf("schemes.put(StandardScheme.class, new %sStandardSchemeFactory());\n", s.Name)
	contents += indent + tab + fmt.Sprintf("schemes.put(TupleScheme.class, new %sTupleSchemeFactory());\n", s.Name)
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateInstanceVars(s *parser.Struct, indent string) string {
	contents := ""
	for _, field := range s.Fields {
		contents += g.generateCommentWithDeprecated(field.Comment, indent, field.Annotations)
		modifier := ""
		if field.Modifier == parser.Required {
			modifier = "required"
		} else if field.Modifier == parser.Optional {
			modifier = "optional"
		}
		modifierComment := ""
		if modifier != "" {
			modifierComment = " // " + modifier
		}
		contents += indent + fmt.Sprintf("public %s %s;%s\n",
			g.getJavaTypeFromThriftType(field.Type), field.Name, modifierComment)
	}
	return contents
}

func (g *Generator) generateFieldsEnum(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "/** The set of fields this struct contains, along with convenience methods for finding and manipulating them. */\n"
	contents += indent + "public enum _Fields implements org.apache.thrift.TFieldIdEnum {\n"

	for idx, field := range s.Fields {
		terminator := ""
		if idx != len(s.Fields)-1 {
			terminator = ","
		}

		if field.Comment != nil {
			contents += g.GenerateBlockComment(field.Comment, indent+tab)
		}
		contents += indent + tab + fmt.Sprintf("%s((short)%d, \"%s\")%s\n", toConstantName(field.Name), field.ID, field.Name, terminator)
	}
	// Do it this was as the semi colon is needed no matter what
	contents += indent + tab + ";\n"
	contents += "\n"

	contents += indent + tab + "private static final Map<String, _Fields> byName = new HashMap<String, _Fields>();\n\n"
	contents += indent + tab + "static {\n"
	contents += indent + tabtab + "for (_Fields field : EnumSet.allOf(_Fields.class)) {\n"
	contents += indent + tabtabtab + "byName.put(field.getFieldName(), field);\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches fieldId, or null if its not found."}, indent+tab)
	contents += indent + tab + "public static _Fields findByThriftId(int fieldId) {\n"
	contents += indent + tabtab + "switch(fieldId) {\n"
	for _, field := range s.Fields {
		contents += indent + tabtabtab + fmt.Sprintf("case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += indent + tabtabtabtab + fmt.Sprintf("return %s;\n", toConstantName(field.Name))
	}
	contents += indent + tabtabtab + "default:\n"
	contents += indent + tabtabtabtab + "return null;\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tab + "}\n\n"

	contents += g.GenerateBlockComment([]string{
		"Find the _Fields constant that matches fieldId, throwing an exception",
		"if it is not found.",
	}, indent+tab)
	contents += indent + tab + "public static _Fields findByThriftIdOrThrow(int fieldId) {\n"
	contents += indent + tabtab + "_Fields fields = findByThriftId(fieldId);\n"
	contents += indent + tabtab + "if (fields == null) throw new IllegalArgumentException(\"Field \" + fieldId + \" doesn't exist!\");\n"
	contents += indent + tabtab + "return fields;\n"
	contents += indent + tab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches name, or null if its not found."}, indent+tab)
	contents += indent + tab + "public static _Fields findByName(String name) {\n"
	contents += indent + tabtab + "return byName.get(name);\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "private final short _thriftId;\n"
	contents += indent + tab + "private final String _fieldName;\n\n"

	contents += indent + tab + "_Fields(short thriftId, String fieldName) {\n"
	contents += indent + tabtab + "_thriftId = thriftId;\n"
	contents += indent + tabtab + "_fieldName = fieldName;\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "public short getThriftFieldId() {\n"
	contents += indent + tabtab + "return _thriftId;\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "public String getFieldName() {\n"
	contents += indent + tabtab + "return _fieldName;\n"
	contents += indent + tab + "}\n"

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetVars(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "// isset id assignments\n"
	primitiveCount := 0
	for _, field := range s.Fields {
		if g.isJavaPrimitive(field.Type) {
			contents += indent + fmt.Sprintf("private static final int %s = %d;\n",
				g.getIsSetID(field.Name), primitiveCount)
			primitiveCount++
		}
	}
	isSetType, bitFieldType := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
	// Do nothing
	case IsSetBitfield:
		contents += indent + fmt.Sprintf("private %s __isset_bitfield = 0;\n", bitFieldType)
	case IsSetBitSet:
		contents += indent + fmt.Sprintf("private BitSet __isset_bit_vector = new BitSet(%d);\n", primitiveCount)
	}
	return contents
}

func (g *Generator) generateDefaultConstructor(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("public %s() {\n", s.Name)
	for _, field := range s.Fields {
		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false, indent+tab)
			contents += fmt.Sprintf("%s\n", val)
		}
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateFullConstructor(s *parser.Struct, indent string) string {
	contents := ""
	argsList := ""
	sep := "\n" + indent + tab
	numNonOptional := 0
	for _, field := range s.Fields {
		if field.Modifier == parser.Optional {
			continue
		}
		argsList += fmt.Sprintf("%s%s %s", sep, g.getJavaTypeFromThriftType(field.Type), field.Name)
		sep = ",\n" + indent + tab
		numNonOptional++
	}

	if numNonOptional > 0 {
		contents += indent + fmt.Sprintf("public %s(%s) {\n", s.Name, argsList)
		contents += indent + tab + "this();\n"
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				continue
			}

			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += indent + tab + fmt.Sprintf("this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name, field.Name)
			} else {
				contents += indent + tab + fmt.Sprintf("this.%s = %s;\n", field.Name, field.Name)
			}

			if g.isJavaPrimitive(field.Type) {
				contents += indent + tab + fmt.Sprintf("set%sIsSet(true);\n", strings.Title(field.Name))
			}
		}
		contents += indent + "}\n\n"
	}
	return contents
}

func (g *Generator) generateCopyConstructor(s *parser.Struct, indent string) string {
	contents := ""
	contents += g.GenerateBlockComment([]string{"Performs a deep copy on <i>other</i>."}, indent)
	contents += indent + fmt.Sprintf("public %s(%s other) {\n", s.Name, s.Name)

	isSetType, _ := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
		// do nothing
	case IsSetBitfield:
		contents += indent + tab + "__isset_bitfield = other.__isset_bitfield;\n"
	case IsSetBitSet:
		contents += indent + tab + "__isset_bit_vector.clear();\n"
		contents += indent + tab + "__isset_bit_vector.or(other.__isset_bit_vector);\n"
	}

	for _, field := range s.Fields {
		isPrimitive := g.isJavaPrimitive(g.Frugal.UnderlyingType(field.Type))
		ind := indent + tab
		if !isPrimitive {
			contents += indent + tab + fmt.Sprintf("if (other.isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
		}
		contents += g.generateCopyConstructorField(field, "other."+field.Name, true, ind)
		if !isPrimitive {
			contents += indent + tab + "}\n"
		}
	}

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateDeepCopyMethod(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("public %s deepCopy() {\n", s.Name)
	contents += indent + tab + fmt.Sprintf("return new %s(this);\n", s.Name)
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateClear(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + fmt.Sprintf("public void clear() {\n")
	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)

		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false, indent+tab)
			contents += val + "\n"
		} else if g.isJavaPrimitive(field.Type) {
			contents += indent + tab + fmt.Sprintf("set%sIsSet(false);\n", strings.Title(field.Name))
			val := g.getPrimitiveDefaultValue(underlyingType)
			contents += indent + tab + fmt.Sprintf("this.%s = %s;\n\n", field.Name, val)
		} else {
			contents += indent + tab + fmt.Sprintf("this.%s = null;\n\n", field.Name)
		}
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateContainerGetSize(field *parser.Field, indent string) string {
	contents := ""
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public int get%sSize() {\n", strings.Title(field.Name))
	contents += indent + tab + fmt.Sprintf("return (this.%s == null) ? 0 : this.%s.size();\n", field.Name, field.Name)
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateContainerIterator(field *parser.Field, indent string) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	// maps don't get iterators
	if underlyingType.Name == "map" {
		return ""
	}

	contents := ""
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public java.util.Iterator<%s> get%sIterator() {\n",
		containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)), strings.Title(field.Name))
	contents += indent + tab + fmt.Sprintf("return (this.%s == null) ? null : this.%s.iterator();\n", field.Name, field.Name)
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateContainerAddTo(field *parser.Field, indent string) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	valType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
	fieldTitle := strings.Title(field.Name)

	contents := ""
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}

	if underlyingType.Name == "list" || underlyingType.Name == "set" {
		contents += indent + fmt.Sprintf("public void addTo%s(%s elem) {\n", fieldTitle, valType)
		newContainer := ""
		if underlyingType.Name == "list" {
			newContainer = fmt.Sprintf("new ArrayList<%s>()", containerType(valType))
		} else {
			newContainer = fmt.Sprintf("new HashSet<%s>()", containerType(valType))
		}
		contents += indent + tab + fmt.Sprintf("if (this.%s == null) {\n", field.Name)
		contents += indent + tabtab + fmt.Sprintf("this.%s = %s;\n", field.Name, newContainer)
		contents += indent + tab + "}\n"
		contents += indent + tab + fmt.Sprintf("this.%s.add(elem);\n", field.Name)
		contents += indent + "}\n\n"
	} else {
		contents += indent + fmt.Sprintf("public void putTo%s(%s key, %s val) {\n",
			fieldTitle, g.getJavaTypeFromThriftType(underlyingType.KeyType), valType)
		contents += indent + tab + fmt.Sprintf("if (this.%s == null) {\n", field.Name)
		contents += indent + tabtab + fmt.Sprintf("this.%s = new HashMap<%s,%s>();\n",
			field.Name, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)), containerType(valType))
		contents += indent + tab + "}\n"
		contents += indent + tab + fmt.Sprintf("this.%s.put(key, val);\n", field.Name)
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) getAccessorPrefix(t *parser.Type) string {
	if g.Frugal.UnderlyingType(t).Name == "bool" {
		return "is"
	}
	return "get"
}

func (g *Generator) generateGetField(field *parser.Field, indent string) string {
	contents := ""
	fieldTitle := strings.Title(field.Name)
	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, indent)
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	returnType := g.getJavaTypeFromThriftType(underlyingType)
	// There's some weird overlap between the ByteBuffer and byte[] methods
	if underlyingType.Name == "binary" {
		returnType = "byte[]"
	}

	accessPrefix := g.getAccessorPrefix(field.Type)
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public %s %s%s() {\n", returnType, accessPrefix, fieldTitle)
	if underlyingType.Name == "binary" {
		contents += indent + tab + fmt.Sprintf("set%s(org.apache.thrift.TBaseHelper.rightSize(%s));\n",
			strings.Title(field.Name), field.Name)
		contents += indent + tab + fmt.Sprintf("return %s == null ? null : %s.array();\n", field.Name, field.Name)
	} else {
		contents += indent + tab + fmt.Sprintf("return this.%s;\n", field.Name)
	}
	contents += indent + "}\n\n"

	if underlyingType.Name == "binary" {
		contents += indent + fmt.Sprintf("public java.nio.ByteBuffer bufferFor%s() {\n", fieldTitle)
		contents += indent + tab + fmt.Sprintf("return org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name)
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) generateSetField(structName string, field *parser.Field, indent string) string {
	fieldTitle := strings.Title(field.Name)
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	contents := ""

	if underlyingType.Name == "binary" {
		// Special additional binary set
		if field.Annotations.IsDeprecated() {
			contents += indent + "@Deprecated\n"
		}
		contents += indent + fmt.Sprintf("public %s set%s(byte[] %s) {\n", structName, fieldTitle, field.Name)
		contents += indent + tab + fmt.Sprintf("this.%s = %s == null ? (java.nio.ByteBuffer)null : java.nio.ByteBuffer.wrap(Arrays.copyOf(%s, %s.length));\n",
			field.Name, field.Name, field.Name, field.Name)
		contents += indent + tab + "return this;\n"
		contents += indent + "}\n\n"
	}

	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, indent)
	}
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public %s set%s(%s %s) {\n",
		structName, fieldTitle, g.getJavaTypeFromThriftType(field.Type), field.Name)

	if underlyingType.Name == "binary" {
		contents += indent + tab + fmt.Sprintf("this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name, field.Name)
	} else {
		contents += indent + tab + fmt.Sprintf("this.%s = %s;\n", field.Name, field.Name)
	}

	if g.isJavaPrimitive(field.Type) {
		contents += indent + tab + fmt.Sprintf("set%sIsSet(true);\n", fieldTitle)
	}

	contents += indent + tab + "return this;\n"
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateUnsetField(s *parser.Struct, field *parser.Field, indent string) string {
	contents := ""

	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public void unset%s() {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += indent + tab + fmt.Sprintf("__isset_bitfield = EncodingUtils.clearBit(__isset_bitfield, %s);\n", isSetID)
		case IsSetBitSet:
			contents += indent + tab + fmt.Sprintf("__isset_bit_vector.clear(%s);\n", isSetID)
		}
	} else {
		contents += indent + tab + fmt.Sprintf("this.%s = null;\n", field.Name)
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) getIsSetID(fieldName string) string {
	return fmt.Sprintf("__%s_ISSET_ID", strings.ToUpper(fieldName))
}

func (g *Generator) generateIsSetField(s *parser.Struct, field *parser.Field, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("/** Returns true if field %s is set (has been assigned a value) and false otherwise */\n", field.Name)
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public boolean isSet%s() {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += indent + tab + fmt.Sprintf("return EncodingUtils.testBit(__isset_bitfield, %s);\n", isSetID)
		case IsSetBitSet:
			contents += indent + tab + fmt.Sprintf("return __isset_bit_vector.get(%s);\n", isSetID)
		}
	} else {
		contents += indent + tab + fmt.Sprintf("return this.%s != null;\n", field.Name)
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateSetIsSetField(s *parser.Struct, field *parser.Field, indent string) string {
	contents := ""
	if field.Annotations.IsDeprecated() {
		contents += indent + "@Deprecated\n"
	}
	contents += indent + fmt.Sprintf("public void set%sIsSet(boolean value) {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += indent + tab + fmt.Sprintf("__isset_bitfield = EncodingUtils.setBit(__isset_bitfield, %s, value);\n", isSetID)
		case IsSetBitSet:
			contents += indent + tab + fmt.Sprintf("__isset_bit_vector.set(%s, value);\n", isSetID)
		}
	} else {
		contents += indent + tab + "if (!value) {\n"
		contents += indent + tabtab + fmt.Sprintf("this.%s = null;\n", field.Name)
		contents += indent + tab + "}\n"
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateSetValue(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "public void setFieldValue(_Fields field, Object value) {\n"
	contents += indent + tab + "switch (field) {\n"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += indent + tab + fmt.Sprintf("case %s:\n", toConstantName(field.Name))
		contents += indent + tabtab + "if (value == null) {\n"
		contents += indent + tabtabtab + fmt.Sprintf("unset%s();\n", fieldTitle)
		contents += indent + tabtab + "} else {\n"
		contents += indent + tabtabtab + fmt.Sprintf("set%s((%s)value);\n", fieldTitle, containerType(g.getJavaTypeFromThriftType(field.Type)))
		contents += indent + tabtab + "}\n"
		contents += indent + tabtab + "break;\n\n"
	}
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateGetValue(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "public Object getFieldValue(_Fields field) {\n"
	contents += indent + tab + "switch (field) {\n"
	for _, field := range s.Fields {
		contents += indent + tab + fmt.Sprintf("case %s:\n", toConstantName(field.Name))
		contents += indent + tabtab + fmt.Sprintf("return %s%s();\n\n",
			g.getAccessorPrefix(field.Type), strings.Title(field.Name))
	}
	contents += indent + tab + "}\n"
	contents += indent + tab + "throw new IllegalStateException();\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetValue(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "/** Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise */\n"
	contents += indent + "public boolean isSet(_Fields field) {\n"
	contents += indent + tab + "if (field == null) {\n"
	contents += indent + tabtab + "throw new IllegalArgumentException();\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "switch (field) {\n"
	for _, field := range s.Fields {
		contents += indent + tab + fmt.Sprintf("case %s:\n", toConstantName(field.Name))
		contents += indent + tabtab + fmt.Sprintf("return isSet%s();\n", strings.Title(field.Name))
	}
	contents += indent + tab + "}\n"
	contents += indent + tab + "throw new IllegalStateException();\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateEquals(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + "public boolean equals(Object that) {\n"
	contents += indent + tab + "if (that == null)\n"
	contents += indent + tabtab + "return false;\n"
	contents += indent + tab + fmt.Sprintf("if (that instanceof %s)\n", s.Name)
	contents += indent + tabtab + fmt.Sprintf("return this.equals((%s)that);\n", s.Name)
	contents += indent + tab + "return false;\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("public boolean equals(%s that) {\n", s.Name)
	contents += indent + tab + "if (that == null)\n"
	contents += indent + tabtab + "return false;\n\n"

	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		primitive := g.isJavaPrimitive(field.Type)

		// TODO 2.0 this looks so ugly
		thisPresentArg := "true"
		thatPresentArg := "true"
		if optional || !primitive {
			thisPresentArg += fmt.Sprintf(" && this.isSet%s()", strings.Title(field.Name))
			thatPresentArg += fmt.Sprintf(" && that.isSet%s()", strings.Title(field.Name))
		}

		contents += indent + tab + fmt.Sprintf("boolean this_present_%s = %s;\n", field.Name, thisPresentArg)
		contents += indent + tab + fmt.Sprintf("boolean that_present_%s = %s;\n", field.Name, thatPresentArg)
		contents += indent + tab + fmt.Sprintf("if (this_present_%s || that_present_%s) {\n", field.Name, field.Name)
		contents += indent + tabtab + fmt.Sprintf("if (!(this_present_%s && that_present_%s))\n", field.Name, field.Name)
		contents += indent + tabtabtab + "return false;\n"

		unequalTest := ""
		if primitive {
			unequalTest = fmt.Sprintf("this.%s != that.%s", field.Name, field.Name)
		} else {
			unequalTest = fmt.Sprintf("!this.%s.equals(that.%s)", field.Name, field.Name)
		}
		contents += indent + tabtab + fmt.Sprintf("if (%s)\n", unequalTest)
		contents += indent + tabtabtab + "return false;\n"
		contents += indent + tab + "}\n\n"
	}

	contents += indent + tab + "return true;\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateHashCode(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + "public int hashCode() {\n"
	contents += indent + tab + "List<Object> list = new ArrayList<Object>();\n\n"
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		primitive := g.isJavaPrimitive(field.Type)

		presentArg := "true"
		if optional || !primitive {
			presentArg += fmt.Sprintf(" && (isSet%s())", strings.Title(field.Name))
		}

		contents += indent + tab + fmt.Sprintf("boolean present_%s = %s;\n", field.Name, presentArg)
		contents += indent + tab + fmt.Sprintf("list.add(present_%s);\n", field.Name)
		contents += indent + tab + fmt.Sprintf("if (present_%s)\n", field.Name)
		if g.Frugal.IsEnum(field.Type) {
			contents += indent + tabtab + fmt.Sprintf("list.add(%s.getValue());\n\n", field.Name)
		} else {
			contents += indent + tabtab + fmt.Sprintf("list.add(%s);\n\n", field.Name)
		}
	}
	contents += indent + tab + "return list.hashCode();\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateCompareTo(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + fmt.Sprintf("public int compareTo(%s other) {\n", s.Name)
	contents += indent + tab + "if (!getClass().equals(other.getClass())) {\n"
	contents += indent + tabtab + "return getClass().getName().compareTo(other.getClass().getName());\n"
	contents += indent + tab + "}\n\n"
	contents += indent + tab + "int lastComparison = 0;\n\n"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += indent + tab + fmt.Sprintf("lastComparison = Boolean.valueOf(isSet%s()).compareTo(other.isSet%s());\n", fieldTitle, fieldTitle)
		contents += indent + tab + "if (lastComparison != 0) {\n"
		contents += indent + tabtab + "return lastComparison;\n"
		contents += indent + tab + "}\n"
		contents += indent + tab + fmt.Sprintf("if (isSet%s()) {\n", fieldTitle)
		contents += indent + tabtab + fmt.Sprintf("lastComparison = org.apache.thrift.TBaseHelper.compareTo(this.%s, other.%s);\n", field.Name, field.Name)
		contents += indent + tabtab + "if (lastComparison != 0) {\n"
		contents += indent + tabtabtab + "return lastComparison;\n"
		contents += indent + tabtab + "}\n"
		contents += indent + tab + "}\n"
	}
	contents += indent + tab + "return 0;\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateFieldForId(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "public _Fields fieldForId(int fieldId) {\n"
	contents += indent + tab + "return _Fields.findByThriftId(fieldId);\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateReadWrite(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "public void read(org.apache.thrift.protocol.TProtocol iprot) throws org.apache.thrift.TException {\n"
	contents += indent + tab + "schemes.get(iprot.getScheme()).getScheme().read(iprot, this);\n"
	contents += indent + "}\n\n"

	contents += indent + "public void write(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {\n"
	contents += indent + tab + "schemes.get(oprot.getScheme()).getScheme().write(oprot, this);\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateToString(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "@Override\n"
	contents += indent + "public String toString() {\n"
	contents += indent + tab + fmt.Sprintf("StringBuilder sb = new StringBuilder(\"%s(\");\n", s.Name)
	contents += indent + tab + "boolean first = true;\n\n"
	first := true
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		ind := ""
		if optional {
			contents += indent + tab + fmt.Sprintf("if (isSet%s()) {\n", strings.Title(field.Name))
			ind = tab
		}

		if !first {
			contents += indent + tab + ind + "if (!first) sb.append(\", \");\n"
		}
		contents += indent + tab + ind + fmt.Sprintf("sb.append(\"%s:\");\n", field.Name)
		if !g.isJavaPrimitive(field.Type) {
			contents += indent + tab + ind + fmt.Sprintf("if (this.%s == null) {\n", field.Name)
			contents += indent + tabtab + ind + "sb.append(\"null\");\n"
			contents += indent + tab + ind + "} else {\n"
			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += indent + tabtab + ind + fmt.Sprintf("org.apache.thrift.TBaseHelper.toString(this.%s, sb);\n", field.Name)
			} else {
				contents += indent + tabtab + ind + fmt.Sprintf("sb.append(this.%s);\n", field.Name)
			}
			contents += indent + tab + ind + "}\n"
		} else {
			contents += indent + tab + ind + fmt.Sprintf("sb.append(this.%s);\n", field.Name)
		}
		contents += indent + tab + ind + "first = false;\n"

		if optional {
			contents += indent + tab + "}\n"
		}
		first = false
	}

	contents += indent + tab + "sb.append(\")\");\n"
	contents += indent + tab + "return sb.toString();\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateValidate(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "public void validate() throws org.apache.thrift.TException {\n"
	contents += indent + tab + "// check for required fields\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && !g.isJavaPrimitive(field.Type) {
			contents += indent + tab + fmt.Sprintf("if (%s == null) {\n", field.Name)
			contents += indent + tabtab + fmt.Sprintf("throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' is not present in struct '%s'\");\n",
				field.Name, s.Name)
			contents += indent + tab + "}\n"
		}
	}

	contents += indent + tab + "// check for sub-struct validity\n"
	for _, field := range s.Fields {
		if g.Frugal.IsStruct(field.Type) && !g.Frugal.IsUnion(field.Type) {
			contents += indent + tab + fmt.Sprintf("if (%s != null) {\n", field.Name)
			contents += indent + tabtab + fmt.Sprintf("%s.validate();\n", field.Name)
			contents += indent + tab + "}\n"
		}
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateWriteObject(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "private void writeObject(java.io.ObjectOutputStream out) throws java.io.IOException {\n"
	contents += indent + tab + "try {\n"
	contents += indent + tabtab + "write(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(out)));\n"
	contents += indent + tab + "} catch (org.apache.thrift.TException te) {\n"
	contents += indent + tabtab + "throw new java.io.IOException(te);\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateReadObject(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + "private void readObject(java.io.ObjectInputStream in) throws java.io.IOException, ClassNotFoundException {\n"
	// isset stuff, don't do for unions
	contents += indent + tab + "try {\n"
	if s.Type != parser.StructTypeUnion {
		contents += indent + tabtab + "// it doesn't seem like you should have to do this, but java serialization is wacky, and doesn't call the default constructor.\n"
		isSetType, _ := g.getIsSetType(s)
		switch isSetType {
		case IsSetNone:
		// Do nothing
		case IsSetBitfield:
			contents += indent + tabtab + "__isset_bitfield = 0;\n"
		case IsSetBitSet:
			contents += indent + tabtab + "__isset_bit_vector = new BitSet(1);\n"
		}
	}

	contents += indent + tabtab + "read(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(in)));\n"
	contents += indent + tab + "} catch (org.apache.thrift.TException te) {\n"
	contents += indent + tabtab + "throw new java.io.IOException(te);\n"
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateStandardScheme(s *parser.Struct, isResult bool, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("private static class %sStandardSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += indent + tab + fmt.Sprintf("public %sStandardScheme getScheme() {\n", s.Name)
	contents += indent + tabtab + fmt.Sprintf("return new %sStandardScheme();\n", s.Name)
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("private static class %sStandardScheme extends StandardScheme<%s> {\n\n", s.Name, s.Name)

	// read
	contents += indent + tab + fmt.Sprintf("public void read(org.apache.thrift.protocol.TProtocol iprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += indent + tabtab + "org.apache.thrift.protocol.TField schemeField;\n"
	contents += indent + tabtab + "iprot.readStructBegin();\n"
	contents += indent + tabtab + "while (true) {\n"
	contents += indent + tabtabtab + "schemeField = iprot.readFieldBegin();\n"
	contents += indent + tabtabtab + "if (schemeField.type == org.apache.thrift.protocol.TType.STOP) {\n"
	contents += indent + tabtabtabtab + "break;\n"
	contents += indent + tabtabtab + "}\n"
	contents += indent + tabtabtab + "switch (schemeField.id) {\n"
	for _, field := range s.Fields {
		contents += indent + tabtabtabtab + fmt.Sprintf("case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += indent + tabtabtabtabtab + fmt.Sprintf("if (schemeField.type == %s) {\n", g.getTType(field.Type))
		contents += g.generateReadFieldRec(field, true, false, false, indent+tabtabtabtabtabtab)
		contents += indent + tabtabtabtabtabtab + fmt.Sprintf("struct.set%sIsSet(true);\n", strings.Title(field.Name))
		contents += indent + tabtabtabtabtab + "} else {\n"
		contents += indent + tabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
		contents += indent + tabtabtabtabtab + "}\n"
		contents += indent + tabtabtabtabtab + "break;\n"
	}
	contents += indent + tabtabtabtab + "default:\n"
	contents += indent + tabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
	contents += indent + tabtabtab + "}\n"
	contents += indent + tabtabtab + "iprot.readFieldEnd();\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tabtab + "iprot.readStructEnd();\n\n"

	contents += indent + tabtab + "// check for required fields of primitive type, which can't be checked in the validate method\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && g.isJavaPrimitive(field.Type) {
			contents += indent + tabtab + fmt.Sprintf("if (!struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += indent + tabtabtab + fmt.Sprintf("throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' was not found in serialized data for struct type '%s'\");\n", field.Name, s.Name)
			contents += indent + tabtab + "}\n"
		}
	}

	contents += indent + tabtab + "struct.validate();\n"
	contents += indent + tab + "}\n\n"

	// write
	contents += indent + tab + fmt.Sprintf("public void write(org.apache.thrift.protocol.TProtocol oprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += indent + tabtab + "struct.validate();\n\n"
	contents += indent + tabtab + "oprot.writeStructBegin(STRUCT_DESC);\n"
	for _, field := range s.Fields {
		isKindOfPrimitive := g.canBeJavaPrimitive(field.Type)
		ind := tabtab
		optInd := tabtab
		if !isKindOfPrimitive {
			contents += indent + ind + fmt.Sprintf("if (struct.%s != null) {\n", field.Name)
			ind += tab
			optInd += tab
		}
		opt := field.Modifier == parser.Optional || (isResult && isKindOfPrimitive)
		if opt {
			contents += indent + ind + fmt.Sprintf("if (struct.isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
		}

		contents += indent + ind + fmt.Sprintf("oprot.writeFieldBegin(%s_FIELD_DESC);\n", toConstantName(field.Name))
		contents += g.generateWriteFieldRec(field, true, false, indent+ind)
		contents += indent + ind + "oprot.writeFieldEnd();\n"

		if opt {
			contents += indent + optInd + "}\n"
		}
		if !isKindOfPrimitive {
			contents += indent + tabtab + "}\n"
		}
	}
	contents += indent + tabtab + "oprot.writeFieldStop();\n"
	contents += indent + tabtab + "oprot.writeStructEnd();\n"

	contents += indent + tab + "}\n\n"

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateTupleScheme(s *parser.Struct, indent string) string {
	contents := ""
	contents += indent + fmt.Sprintf("private static class %sTupleSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += indent + tab + fmt.Sprintf("public %sTupleScheme getScheme() {\n", s.Name)
	contents += indent + tabtab + fmt.Sprintf("return new %sTupleScheme();\n", s.Name)
	contents += indent + tab + "}\n"
	contents += indent + "}\n\n"

	contents += indent + fmt.Sprintf("private static class %sTupleScheme extends TupleScheme<%s> {\n\n", s.Name, s.Name)
	contents += indent + tab + "@Override\n"
	contents += indent + tab + fmt.Sprintf("public void write(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += indent + tabtab + "TTupleProtocol oprot = (TTupleProtocol) prot;\n"
	// write required fields
	numNonReqs := 0
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			numNonReqs++
			continue
		}

		contents += g.generateWriteFieldRec(field, true, true, indent+tabtab)
	}

	if numNonReqs > 0 {
		// write optional/default fields
		nonReqFieldCount := 0
		contents += indent + tabtab + "BitSet optionals = new BitSet();\n"
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += indent + tabtab + fmt.Sprintf("if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += indent + tabtabtab + fmt.Sprintf("optionals.set(%d);\n", nonReqFieldCount)
			contents += indent + tabtab + "}\n"
			nonReqFieldCount++
		}

		contents += indent + tabtab + fmt.Sprintf("oprot.writeBitSet(optionals, %d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += indent + tabtab + fmt.Sprintf("if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += g.generateWriteFieldRec(field, true, true, indent+tabtabtab)
			contents += indent + tabtab + "}\n"
		}
	}

	contents += indent + tab + "}\n\n"

	contents += indent + tab + "@Override\n"
	contents += indent + tab + fmt.Sprintf("public void read(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += indent + tabtab + "TTupleProtocol iprot = (TTupleProtocol) prot;\n"
	// read required fields
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			continue
		}

		contents += g.generateReadFieldRec(field, true, true, false, indent+tabtab)
		contents += indent + tabtab + fmt.Sprintf("struct.set%sIsSet(true);\n", strings.Title(field.Name))
	}

	if numNonReqs > 0 {
		// read default/optional fields
		nonReqFieldCount := 0
		contents += indent + tabtab + fmt.Sprintf("BitSet incoming = iprot.readBitSet(%d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += indent + tabtab + fmt.Sprintf("if (incoming.get(%d)) {\n", nonReqFieldCount)
			contents += g.generateReadFieldRec(field, true, true, false, indent+tabtabtab)
			contents += indent + tabtabtab + fmt.Sprintf("struct.set%sIsSet(true);\n", strings.Title(field.Name))
			contents += indent + tabtab + "}\n"
			nonReqFieldCount++
		}
	}
	contents += indent + tab + "}\n\n"

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateCopyConstructorField(field *parser.Field, otherFieldName string, first bool, indent string) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isPrimitive := g.canBeJavaPrimitive(underlyingType)
	accessPrefix := "this."
	declPrefix := "this."
	if !first {
		accessPrefix = ""
		declPrefix = g.getJavaTypeFromThriftType(underlyingType) + " "
	}

	if isPrimitive || underlyingType.Name == "string" {
		return indent + fmt.Sprintf("%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.Name == "binary" {
		return indent + fmt.Sprintf("%s%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", declPrefix, field.Name, otherFieldName)
	} else if g.Frugal.IsStruct(underlyingType) {
		return indent + fmt.Sprintf("%s%s = new %s(%s);\n", declPrefix, field.Name, g.getJavaTypeFromThriftType(underlyingType), otherFieldName)
	} else if g.Frugal.IsEnum(underlyingType) {
		return indent + fmt.Sprintf("%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.IsContainer() {
		contents := ""
		valueType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		containerValType := containerType(valueType)
		otherValElem := g.GetElem()
		thisValElem := g.GetElem()
		thisValField := parser.FieldFromType(underlyingType.ValueType, thisValElem)

		switch underlyingType.Name {
		case "list":
			contents += indent + fmt.Sprintf("%s%s = new ArrayList<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += indent + fmt.Sprintf("for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, indent+tab)
			contents += indent + tab + fmt.Sprintf("%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += indent + "}\n"
		case "set":
			contents += indent + fmt.Sprintf("%s%s = new HashSet<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += indent + fmt.Sprintf("for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, indent+tab)
			contents += indent + tab + fmt.Sprintf("%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += indent + "}\n"
		case "map":
			keyType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyUnderlying := g.Frugal.UnderlyingType(underlyingType.KeyType)
			valUnderlying := g.Frugal.UnderlyingType(underlyingType.ValueType)
			containerKeyType := containerType(keyType)

			// If it's all primitives, optimization. Otherwise need to iterate
			if (g.isJavaPrimitive(keyUnderlying) || keyUnderlying.Name == "string") &&
				(g.isJavaPrimitive(valUnderlying) || valUnderlying.Name == "string") {
				contents += indent + fmt.Sprintf("%s%s = new HashMap<%s,%s>(%s);\n",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
			} else {
				thisKeyElem := g.GetElem()
				thisKeyField := parser.FieldFromType(underlyingType.KeyType, thisKeyElem)

				contents += indent + fmt.Sprintf("%s%s = new HashMap<%s,%s>(%s.size());\n",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
				contents += indent + fmt.Sprintf("for (Map.Entry<%s, %s> %s : %s.entrySet()) {\n",
					containerKeyType, containerValType, otherValElem, otherFieldName)
				contents += g.generateCopyConstructorField(thisKeyField, otherValElem+".getKey()", false, indent+tab)
				contents += g.generateCopyConstructorField(thisValField, otherValElem+".getValue()", false, indent+tab)
				contents += indent + tab + fmt.Sprintf("%s%s.put(%s, %s);\n", accessPrefix, field.Name, thisKeyElem, thisValElem)
				contents += indent + "}\n"
			}
		}

		return contents
	}
	panic("unrecognized type: " + underlyingType.Name)
}

// succinct means only read collection length instead of the whole header,
// and don't read collection end.
// containerTypes causes variables to be declared as objects instead of
// potentially primitives
func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, succinct bool, containerTypes bool, indent string) string {
	contents := ""
	declPrefix := "struct."
	accessPrefix := "struct."
	javaType := g.getJavaTypeFromThriftType(field.Type)
	if !first {
		if containerTypes {
			declPrefix = containerType(javaType) + " "
		} else {
			declPrefix = javaType + " "
		}
		accessPrefix = ""
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	if underlyingType.IsPrimitive() {
		thriftType := ""
		switch underlyingType.Name {
		case "bool":
			thriftType = "Bool"
		case "byte", "i8":
			thriftType = "Byte"
		case "i16":
			thriftType = "I16"
		case "i32":
			thriftType = "I32"
		case "i64":
			thriftType = "I64"
		case "double":
			thriftType = "Double"
		case "string":
			thriftType = "String"
		case "binary":
			thriftType = "Binary"
		default:
			panic("unkown thrift type: " + underlyingType.Name)
		}

		contents += indent + fmt.Sprintf("%s%s = iprot.read%s();\n", declPrefix, field.Name, thriftType)
	} else if g.Frugal.IsEnum(underlyingType) {
		contents += indent + fmt.Sprintf("%s%s = %s.findByValue(iprot.readI32());\n", declPrefix, field.Name, javaType)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += indent + fmt.Sprintf("%s%s = new %s();\n", declPrefix, field.Name, javaType)
		contents += indent + fmt.Sprintf("%s%s.read(iprot);\n", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		containerElem := g.GetElem()
		counterElem := g.GetElem()

		valType := containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType))
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valContents := g.generateReadFieldRec(valField, false, succinct, containerTypes, indent+tab)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TList %s = new org.apache.thrift.protocol.TList(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TList %s = iprot.readListBegin();\n", containerElem)
			}
			contents += indent + fmt.Sprintf("%s%s = new ArrayList<%s>(%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += indent + fmt.Sprintf("for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += indent + tab + fmt.Sprintf("%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "iprot.readListEnd();\n"
			}
		case "set":
			if succinct {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TSet %s = new org.apache.thrift.protocol.TSet(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TSet %s = iprot.readSetBegin();\n", containerElem)
			}
			contents += indent + fmt.Sprintf("%s%s = new HashSet<%s>(2*%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += indent + fmt.Sprintf("for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += indent + tab + fmt.Sprintf("%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "iprot.readSetEnd();\n"
			}
		case "map":
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TMap %s = new org.apache.thrift.protocol.TMap(%s, %s, iprot.readI32());\n",
					containerElem, keyTType, valTType)
			} else {
				contents += indent + fmt.Sprintf("org.apache.thrift.protocol.TMap %s = iprot.readMapBegin();\n", containerElem)
			}

			keyType := containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType))
			contents += indent + fmt.Sprintf("%s%s = new HashMap<%s,%s>(2*%s.size);\n", declPrefix, field.Name, keyType, valType, containerElem)
			contents += indent + fmt.Sprintf("for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, succinct, containerTypes, indent+tab)
			contents += valContents
			contents += indent + tab + fmt.Sprintf("%s%s.put(%s, %s);\n", accessPrefix, field.Name, keyElem, valElem)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "iprot.readMapEnd();\n"
			}
		}
	}

	return contents
}

// succinct means only write collection length instead of the whole header,
// and don't write collection end.
func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, succinct bool, indent string) string {
	contents := ""
	accessPrefix := "struct."
	if !first {
		accessPrefix = ""
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isEnum := g.Frugal.IsEnum(underlyingType)
	if underlyingType.IsPrimitive() || isEnum {
		elem := g.GetElem()

		// Store the value in an intermittent value
		// This allows writing a default value if using boxed primitives
		// and the value is "null"
		contents += indent + fmt.Sprintf("%s %s = %s%s;\n",
			g.getJavaTypeFromThriftType(underlyingType), elem, accessPrefix, field.Name)
		if g.canBeJavaPrimitive(underlyingType) && g.generateBoxedPrimitives() {
			contents += indent + fmt.Sprintf("if (%s == null) {\n", elem)
			val := g.getPrimitiveDefaultValue(underlyingType)
			contents += indent + tab + fmt.Sprintf("%s = %s;\n", elem, val)
			contents += indent + "}\n"
		}

		write := indent + "oprot.write"
		switch underlyingType.Name {
		case "bool":
			write += "Bool(%s);\n"
		case "byte", "i8":
			write += "Byte(%s);\n"
		case "i16":
			write += "I16(%s);\n"
		case "i32":
			write += "I32(%s);\n"
		case "i64":
			write += "I64(%s);\n"
		case "double":
			write += "Double(%s);\n"
		case "string":
			write += "String(%s);\n"
		case "binary":
			write += "Binary(%s);\n"
		default:
			if isEnum {
				write += "I32(%s.getValue());\n"
			} else {
				panic("unknown thrift type: " + underlyingType.Name)
			}
		}

		contents += fmt.Sprintf(write, elem)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += indent + fmt.Sprintf("%s%s.write(oprot);\n", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		iterElem := g.GetElem()
		valJavaType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += indent + fmt.Sprintf("oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += indent + fmt.Sprintf("oprot.writeListBegin(new org.apache.thrift.protocol.TList(%s, %s%s.size()));\n",
					valTType, accessPrefix, field.Name)
			}
			contents += indent + fmt.Sprintf("for (%s %s : %s%s) {\n", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, indent+tab)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "oprot.writeListEnd();\n"
			}
		case "set":
			if succinct {
				contents += indent + fmt.Sprintf("oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += indent + fmt.Sprintf("oprot.writeSetBegin(new org.apache.thrift.protocol.TSet(%s, %s%s.size()));\n",
					valTType, accessPrefix, field.Name)
			}
			contents += indent + fmt.Sprintf("for (%s %s : %s%s) {\n", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, indent+tab)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "oprot.writeSetEnd();\n"
			}
		case "map":
			keyJavaType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += indent + fmt.Sprintf("oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += indent + fmt.Sprintf("oprot.writeMapBegin(new org.apache.thrift.protocol.TMap(%s, %s, %s%s.size()));\n",
					keyTType, valTType, accessPrefix, field.Name)
			}
			contents += indent + fmt.Sprintf("for (Map.Entry<%s, %s> %s : %s%s.entrySet()) {\n",
				containerType(keyJavaType), containerType(valJavaType), iterElem, accessPrefix, field.Name)
			keyField := parser.FieldFromType(underlyingType.KeyType, iterElem+".getKey()")
			valField := parser.FieldFromType(underlyingType.ValueType, iterElem+".getValue()")
			contents += g.generateWriteFieldRec(keyField, false, succinct, indent+tab)
			contents += g.generateWriteFieldRec(valField, false, succinct, indent+tab)
			contents += indent + "}\n"
			if !succinct {
				contents += indent + "oprot.writeMapEnd();\n"
			}
		default:
			panic("unknown type: " + underlyingType.Name)
		}
	}

	return contents
}

func (g *Generator) GetOutputDir(dir string) string {
	if namespace := g.Frugal.Namespace(lang); namespace != nil {
		path := generator.GetPackageComponents(namespace.Value)
		dir = filepath.Join(append([]string{dir}, path...)...)
	}
	return dir
}

func (g *Generator) DefaultOutputDir() string {
	return defaultOutputDir
}

func (g *Generator) PostProcess(f *os.File) error { return nil }

func (g *Generator) GenerateDependencies(dir string) error {
	return nil
}

func (g *Generator) GenerateFile(name, outputDir string, fileType generator.FileType) (*os.File, error) {
	switch fileType {
	case generator.PublishFile:
		return g.CreateFile(strings.Title(name)+"Publisher", outputDir, lang, false)
	case generator.SubscribeFile:
		return g.CreateFile(strings.Title(name)+"Subscriber", outputDir, lang, false)
	case generator.CombinedServiceFile:
		return g.CreateFile("F"+name, outputDir, lang, false)
	case generator.ObjectFile:
		return g.CreateFile(name, outputDir, lang, false)
	default:
		return nil, fmt.Errorf("Bad file type for Java generator: %s", fileType)
	}
}

func (g *Generator) GenerateDocStringComment(file *os.File) error {
	comment := fmt.Sprintf(
		"/**\n"+
			" * Autogenerated by Frugal Compiler (%s)\n"+
			" * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING\n"+
			" *\n"+
			" * @generated\n"+
			" */",
		globals.Version)

	_, err := file.WriteString(comment)
	return err
}

func (g *Generator) GenerateServicePackage(file *os.File, s *parser.Service) error {
	return g.generatePackage(file)
}

func (g *Generator) GenerateScopePackage(file *os.File, s *parser.Scope) error {
	return g.generatePackage(file)
}

func (g *Generator) generatePackage(file *os.File) error {
	namespace := g.Frugal.Namespace(lang)
	if namespace == nil {
		return nil
	}
	_, err := file.WriteString(fmt.Sprintf("package %s;", namespace.Value))
	return err
}

func (g *Generator) GenerateEnumImports(file *os.File) error {
	imports := ""
	imports += "import java.util.Map;\n"
	imports += "import java.util.HashMap;\n"
	imports += "import org.apache.thrift.TEnum;\n"
	imports += "\n"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateStructImports(file *os.File) error {
	_, err := file.WriteString(g.generateStructImports())
	return err
}

func (g *Generator) generateStructImports() string {
	imports := ""
	imports += "import org.apache.thrift.scheme.IScheme;\n"
	imports += "import org.apache.thrift.scheme.SchemeFactory;\n"
	imports += "import org.apache.thrift.scheme.StandardScheme;\n\n"
	imports += "import org.apache.thrift.scheme.TupleScheme;\n"
	imports += "import org.apache.thrift.protocol.TTupleProtocol;\n"
	imports += "import org.apache.thrift.protocol.TProtocolException;\n"
	imports += "import org.apache.thrift.EncodingUtils;\n"
	imports += "import org.apache.thrift.TException;\n"
	imports += "import org.apache.thrift.async.AsyncMethodCallback;\n"
	imports += "import org.apache.thrift.server.AbstractNonblockingServer.*;\n"
	imports += "import java.util.List;\n"
	imports += "import java.util.ArrayList;\n"
	imports += "import java.util.Map;\n"
	imports += "import java.util.HashMap;\n"
	imports += "import java.util.EnumMap;\n"
	imports += "import java.util.Set;\n"
	imports += "import java.util.HashSet;\n"
	imports += "import java.util.EnumSet;\n"
	imports += "import java.util.Collections;\n"
	imports += "import java.util.BitSet;\n"
	imports += "import java.nio.ByteBuffer;\n"
	imports += "import java.util.Arrays;\n"
	if g.includeGeneratedAnnotation() {
		imports += "import javax.annotation.Generated;\n"
	}
	imports += "import org.slf4j.Logger;\n"
	imports += "import org.slf4j.LoggerFactory;\n"

	imports += "\n"

	return imports
}

func (g *Generator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := ""

	imports += g.generateStructImports()

	imports += "import com.workiva.frugal.FContext;\n"
	imports += "import com.workiva.frugal.exception.TApplicationExceptionType;\n"
	imports += "import com.workiva.frugal.exception.TTransportExceptionType;\n"
	imports += "import com.workiva.frugal.middleware.InvocationHandler;\n"
	imports += "import com.workiva.frugal.middleware.ServiceMiddleware;\n"
	imports += "import com.workiva.frugal.processor.FBaseProcessor;\n"
	imports += "import com.workiva.frugal.processor.FProcessor;\n"
	imports += "import com.workiva.frugal.processor.FProcessorFunction;\n"
	imports += "import com.workiva.frugal.protocol.*;\n"
	imports += "import com.workiva.frugal.provider.FServiceProvider;\n"
	imports += "import com.workiva.frugal.transport.FTransport;\n"
	imports += "import com.workiva.frugal.transport.TMemoryOutputBuffer;\n"
	imports += "import org.apache.thrift.TApplicationException;\n"
	imports += "import org.apache.thrift.TException;\n"
	imports += "import org.apache.thrift.protocol.TMessage;\n"
	imports += "import org.apache.thrift.protocol.TMessageType;\n"
	imports += "import org.apache.thrift.transport.TTransport;\n"
	imports += "import org.apache.thrift.transport.TTransportException;\n"
	if g.includeGeneratedAnnotation() {
		imports += "import javax.annotation.Generated;\n"
	}
	imports += "import java.util.Arrays;\n"
	imports += "import java.util.concurrent.*;\n"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import com.workiva.frugal.FContext;\n"
	imports += "import com.workiva.frugal.exception.TApplicationExceptionType;\n"
	imports += "import com.workiva.frugal.middleware.InvocationHandler;\n"
	imports += "import com.workiva.frugal.middleware.ServiceMiddleware;\n"
	imports += "import com.workiva.frugal.protocol.*;\n"
	imports += "import com.workiva.frugal.provider.FScopeProvider;\n"
	imports += "import com.workiva.frugal.transport.FPublisherTransport;\n"
	imports += "import com.workiva.frugal.transport.FSubscriberTransport;\n"
	imports += "import com.workiva.frugal.transport.FSubscription;\n"
	imports += "import com.workiva.frugal.transport.TMemoryOutputBuffer;\n"
	imports += "import org.apache.thrift.TException;\n"
	imports += "import org.apache.thrift.TApplicationException;\n"
	imports += "import org.apache.thrift.transport.TTransport;\n"
	imports += "import org.apache.thrift.transport.TTransportException;\n"
	imports += "import org.apache.thrift.protocol.*;\n\n"

	imports += "import java.util.List;\n"
	imports += "import java.util.ArrayList;\n"
	imports += "import java.util.Map;\n"
	imports += "import java.util.HashMap;\n"
	imports += "import java.util.EnumMap;\n"
	imports += "import java.util.Set;\n"
	imports += "import java.util.HashSet;\n"
	imports += "import java.util.EnumSet;\n"
	imports += "import java.util.Collections;\n"
	imports += "import java.util.BitSet;\n"
	imports += "import java.nio.ByteBuffer;\n"
	imports += "import java.util.Arrays;\n"
	imports += "import org.slf4j.Logger;\n"
	imports += "import org.slf4j.LoggerFactory;\n"
	if g.includeGeneratedAnnotation() {
		imports += "import javax.annotation.Generated;\n"
	}

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateConstants(file *os.File, name string) error {
	return nil
}

func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	scopeTitle := strings.Title(scope.Name)
	contents := ""

	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation("")
	}

	contents += fmt.Sprintf("public class %sPublisher {\n\n", scopeTitle)

	contents += g.generatePublisherIface(scope, tab)
	contents += g.generatePublisherClient(scope, tab)

	contents += "}"

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generatePublisherIface(scope *parser.Scope, indent string) string {
	contents := ""

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, indent)
	}
	contents += indent + "public interface Iface {\n"

	contents += indent + tab + "public void open() throws TException;\n\n"
	contents += indent + tab + "public void close() throws TException;\n\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, indent+tab)
		}
		contents += indent + tab + fmt.Sprintf("public void publish%s(FContext ctx, %s%s req) throws TException;\n\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))
	}

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generatePublisherClient(scope *parser.Scope, indent string) string {
	contents := ""

	scopeTitle := strings.Title(scope.Name)

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, indent)
	}
	contents += indent + "public static class Client implements Iface {\n"
	contents += indent + tab + fmt.Sprintf("private static final String DELIMITER = \"%s\";\n\n", globals.TopicDelimiter)
	contents += indent + tab + "private final Iface target;\n"
	contents += indent + tab + "private final Iface proxy;\n\n"

	contents += indent + tab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {\n"
	contents += indent + tabtab + fmt.Sprintf("target = new Internal%sPublisher(provider);\n", scopeTitle)
	contents += indent + tabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	contents += indent + tabtab + "combined.addAll(provider.getMiddleware());\n"
	contents += indent + tabtab + "middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	contents += indent + tabtab + "proxy = InvocationHandler.composeMiddleware(target, Iface.class, middleware);\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "public void open() throws TException {\n"
	contents += indent + tabtab + "target.open();\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "public void close() throws TException {\n"
	contents += indent + tabtab + "target.close();\n"
	contents += indent + tab + "}\n\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, indent+tab)
		}
		contents += indent + tab + fmt.Sprintf("public void publish%s(FContext ctx, %s%s req) throws TException {\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))
		contents += indent + tabtab + fmt.Sprintf("proxy.publish%s(%s);\n", op.Name, g.generateScopeArgs(scope))
		contents += indent + tab + "}\n\n"
	}

	contents += indent + tab + fmt.Sprintf("protected static class Internal%sPublisher implements Iface {\n\n", scopeTitle)

	contents += indent + tabtab + "private FScopeProvider provider;\n"
	contents += indent + tabtab + "private FPublisherTransport transport;\n"

	contents += indent + tabtab + "private FProtocolFactory protocolFactory;\n\n"

	contents += indent + tabtab + fmt.Sprintf("protected Internal%sPublisher() {\n", scopeTitle)
	contents += indent + tabtab + "}\n\n"

	contents += indent + tabtab + fmt.Sprintf("public Internal%sPublisher(FScopeProvider provider) {\n", scopeTitle)
	contents += indent + tabtabtab + "this.provider = provider;\n"
	contents += indent + tabtab + "}\n\n"

	contents += indent + tabtab + "public void open() throws TException {\n"
	contents += indent + tabtabtab + "FScopeProvider.Publisher publisher = provider.buildPublisher();\n"
	contents += indent + tabtabtab + "transport = publisher.getTransport();\n"
	contents += indent + tabtabtab + "protocolFactory = publisher.getProtocolFactory();\n"
	contents += indent + tabtabtab + "transport.open();\n"
	contents += indent + tabtab + "}\n\n"

	contents += indent + tabtab + "public void close() throws TException {\n"
	contents += indent + tabtabtab + "transport.close();\n"
	contents += indent + tabtab + "}\n\n"

	prefix := ""
	for _, op := range scope.Operations {
		contents += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, indent+tabtab)
		}

		contents += indent + tabtab + fmt.Sprintf("public void publish%s(FContext ctx, %s%s req) throws TException {\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))

		// Inject the prefix variables into the FContext to send
		for _, prefixVar := range scope.Prefix.Variables {
			contents += indent + tabtabtab + fmt.Sprintf("ctx.addRequestHeader(\"_topic_%s\", %s);\n", prefixVar, prefixVar)
		}

		contents += indent + tabtabtab + fmt.Sprintf("String op = \"%s\";\n", op.Name)
		contents += indent + tabtabtab + fmt.Sprintf("String prefix = %s;\n", generatePrefixStringTemplate(scope))
		contents += indent + tabtabtab + "String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);\n"
		contents += indent + tabtabtab + "TMemoryOutputBuffer memoryBuffer = new TMemoryOutputBuffer(transport.getPublishSizeLimit());\n"
		contents += indent + tabtabtab + "FProtocol oprot = protocolFactory.getProtocol(memoryBuffer);\n"
		contents += indent + tabtabtab + "oprot.writeRequestHeader(ctx);\n"
		contents += indent + tabtabtab + "oprot.writeMessageBegin(new TMessage(op, TMessageType.CALL, 0));\n"
		contents += g.generateWriteFieldRec(parser.FieldFromType(op.Type, "req"), false, false, indent+tabtabtab)
		contents += indent + tabtabtab + "oprot.writeMessageEnd();\n"
		contents += indent + tabtabtab + "transport.publish(topic, memoryBuffer.getWriteBytes());\n"
		contents += indent + tabtab + "}\n"
	}

	contents += indent + tab + "}\n"
	contents += indent + "}\n"

	return contents
}

func generatePrefixStringTemplate(scope *parser.Scope) string {
	if len(scope.Prefix.Variables) == 0 {
		if scope.Prefix.String == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s%s"`, scope.Prefix.String, globals.TopicDelimiter)
	}
	template := "String.format(\""
	template += scope.Prefix.Template("%s")
	template += globals.TopicDelimiter + "\", "
	prefix := ""
	for _, variable := range scope.Prefix.Variables {
		template += prefix + variable
		prefix = ", "
	}
	template += ")"
	return template
}

func (g *Generator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	contents := ""
	scopeName := strings.Title(scope.Name)
	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation("")
	}

	contents += fmt.Sprintf("public class %sSubscriber {\n\n", scopeName)

	contents += g.generateSubscriberIface(scope, tab)
	contents += g.generateHandlerIfaces(scope, tab)
	contents += g.generateSubscriberClient(scope, tab)

	contents += "\n}"

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateSubscriberIface(scope *parser.Scope, indent string) string {
	contents := ""

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, indent)
	}

	// generate a non-throwable interface
	contents += indent + "public interface Iface {\n"
	args := g.generateScopePrefixArgs(scope)
	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, indent+tab)
		}
		contents += indent + tab + fmt.Sprintf("public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException;\n\n",
			op.Name, args, op.Name)
	}

	// generate a throwable interface
	contents += indent + "}\n\n"
	contents += indent + "public interface IfaceThrowable {\n"
	throwableArgs := g.generateScopePrefixArgs(scope)
	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, indent+tab)
		}
		contents += indent + tab + fmt.Sprintf("public FSubscription subscribe%sThrowable(%sfinal %sThrowableHandler handler) throws TException;\n\n",
			op.Name, throwableArgs, op.Name)
	}

	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) generateHandlerIfaces(scope *parser.Scope, indent string) string {
	contents := ""

	// generate non-throwable handler interfaces
	for _, op := range scope.Operations {
		contents += indent + fmt.Sprintf("public interface %sHandler {\n", op.Name)
		contents += indent + tab + fmt.Sprintf("void on%s(FContext ctx, %s req) throws TException;\n", op.Name, g.getJavaTypeFromThriftType(op.Type))
		contents += indent + "}\n\n"
	}

	// generate throwable handler interfaces
	for _, op := range scope.Operations {
		contents += indent + fmt.Sprintf("public interface %sThrowableHandler {\n", op.Name)
		contents += indent + tab + fmt.Sprintf("void on%s(FContext ctx, %s req) throws TException;\n", op.Name, g.getJavaTypeFromThriftType(op.Type))
		contents += indent + "}\n\n"
	}

	return contents
}

func (g *Generator) generateSubscriberClient(scope *parser.Scope, indent string) string {
	contents := ""

	prefix := ""
	args := g.generateScopePrefixArgs(scope)

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, indent)
	}
	contents += indent + "public static class Client implements Iface, IfaceThrowable {\n"

	contents += indent + tab + fmt.Sprintf("private static final String DELIMITER = \"%s\";\n", globals.TopicDelimiter)
	contents += indent + tab + "private static final Logger LOGGER = LoggerFactory.getLogger(Client.class);\n\n"

	contents += indent + tab + "private final FScopeProvider provider;\n"
	contents += indent + tab + "private final ServiceMiddleware[] middleware;\n\n"

	contents += indent + tab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {\n"
	contents += indent + tabtab + "this.provider = provider;\n"
	contents += indent + tabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	contents += indent + tabtab + "combined.addAll(provider.getMiddleware());\n"
	contents += indent + tabtab + "this.middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	contents += indent + tab + "}\n\n"

	throwable := false
	for i := 0; i < 2; i++ {
		for _, op := range scope.Operations {
			contents += prefix
			prefix = "\n\n"
			if op.Comment != nil {
				contents += g.GenerateBlockComment(op.Comment, indent+tab)
			}

			if throwable {
				contents += indent + tab + fmt.Sprintf("public FSubscription subscribe%sThrowable(%sfinal %sThrowableHandler handler) throws TException {\n", op.Name, args, op.Name)
			} else {
				contents += indent + tab + fmt.Sprintf("public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException {\n", op.Name, args, op.Name)
			}
			contents += indent + tabtab + fmt.Sprintf("final String op = \"%s\";\n", op.Name)
			contents += indent + tabtab + fmt.Sprintf("String prefix = %s;\n", generatePrefixStringTemplate(scope))
			contents += indent + tabtab + "final String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);\n"
			contents += indent + tabtab + "final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();\n"

			contents += indent + tabtab + "final FSubscriberTransport transport = subscriber.getTransport();\n"

			if throwable {
				contents += indent + tabtab + fmt.Sprintf(
					"final %sThrowableHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, %sThrowableHandler.class, middleware);\n",
					op.Name, op.Name)
			} else {
				contents += indent + tabtab + fmt.Sprintf(
					"final %sHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, %sHandler.class, middleware);\n",
					op.Name, op.Name)
			}

			contents += indent + tabtab + fmt.Sprintf("transport.subscribe(topic, recv%s(op, subscriber.getProtocolFactory(), proxiedHandler));\n", op.Name)
			contents += indent + tabtab + "return FSubscription.of(topic, transport);\n"
			contents += indent + tab + "}\n\n"

			callback := "FAsyncCallback"
			if throwable {
				contents += indent + tab + fmt.Sprintf("private %s recv%s(String op, FProtocolFactory pf, %sThrowableHandler handler) {\n", callback, op.Name, op.Name)
			} else {
				contents += indent + tab + fmt.Sprintf("private %s recv%s(String op, FProtocolFactory pf, %sHandler handler) {\n", callback, op.Name, op.Name)
			}

			contents += indent + tabtab + fmt.Sprintf("return new %s() {\n", callback)

			contents += indent + tabtabtab + "public void onMessage(TTransport tr) throws TException {\n"
			contents += indent + tabtabtabtab + "FProtocol iprot = pf.getProtocol(tr);\n"
			contents += indent + tabtabtabtab + "FContext ctx = iprot.readRequestHeader();\n"
			contents += indent + tabtabtabtab + "TMessage msg = iprot.readMessageBegin();\n"
			contents += indent + tabtabtabtab + "if (!msg.name.equals(op)) {\n"
			contents += indent + tabtabtabtabtab + "TProtocolUtil.skip(iprot, TType.STRUCT);\n"
			contents += indent + tabtabtabtabtab + "iprot.readMessageEnd();\n"
			contents += indent + tabtabtabtabtab + "throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);\n"
			contents += indent + tabtabtabtab + "}\n"
			contents += g.generateReadFieldRec(parser.FieldFromType(op.Type, "received"), false, false, false, indent+tabtabtabtab)
			contents += indent + tabtabtabtab + "iprot.readMessageEnd();\n"

			contents += indent + tabtabtabtab + fmt.Sprintf("handler.on%s(ctx, received);\n", op.Name)
			contents += indent + tabtabtab + "}\n"
			contents += indent + tabtab + "};\n"
			contents += indent + tab + "}"
		}
		throwable = true
	}
	contents += "\n" + indent + "}\n"

	return contents
}

func (g *Generator) generateScopePrefixArgs(scope *parser.Scope) string {
	args := ""
	if len(scope.Prefix.Variables) > 0 {
		for _, variable := range scope.Prefix.Variables {
			args = fmt.Sprintf("%sString %s, ", args, variable)
		}
	}
	return args
}

func (g *Generator) GenerateService(file *os.File, s *parser.Service) error {
	contents := ""
	if g.includeGeneratedAnnotation() {
		contents += g.generatedAnnotation("")
	}
	contents += fmt.Sprintf("public class F%s {\n\n", s.Name)
	contents += tab + fmt.Sprintf("private static final Logger logger = LoggerFactory.getLogger(F%s.class);\n\n", s.Name)
	contents += g.generateServiceInterface(s, tab)
	contents += g.generateClient(s, tab)
	contents += g.generateServer(s, tab)
	contents += g.generateServiceArgsResults(s, tab)
	contents += "}"

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateCommentWithDeprecated(comment []string, indent string, anns parser.Annotations) string {
	fullComment := []string{}
	if comment != nil {
		fullComment = append(fullComment, comment...)
	}

	deprecationValue, deprecated := anns.Deprecated()
	if deprecated && deprecationValue != "" {
		fullComment = append(fullComment, fmt.Sprintf("@deprecated %s", deprecationValue))
	}

	contents := ""
	if len(fullComment) != 0 {
		contents += g.GenerateBlockComment(fullComment, indent)
	}
	if deprecated {
		contents += indent + "@Deprecated\n"
	}
	return contents
}

func (g *Generator) generateServiceInterface(service *parser.Service, indent string) string {
	contents := ""
	if service.Comment != nil {
		contents += g.GenerateBlockComment(service.Comment, indent)
	}
	if service.Extends != "" {
		contents += indent + fmt.Sprintf("public interface Iface extends %s.Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += indent + "public interface Iface {\n\n"
	}
	for _, method := range service.Methods {
		contents += g.generateCommentWithDeprecated(method.Comment, indent+tab, method.Annotations)
		contents += indent + tab + fmt.Sprintf("public %s %s(FContext ctx%s) %s;\n\n",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	}
	contents += indent + "}\n\n"
	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "F" + service.ExtendsService()
	include := service.ExtendsInclude()
	if namespace := g.namespaceForInclude(include); namespace != "" {
		serviceName = namespace + "." + serviceName
	}
	return serviceName
}

func (g *Generator) generateReturnValue(method *parser.Method) string {
	return g.generateContextualReturnValue(method, false)
}

func (g *Generator) generateBoxedReturnValue(method *parser.Method) string {
	return g.generateContextualReturnValue(method, true)
}

func (g *Generator) generateContextualReturnValue(method *parser.Method, boxed bool) string {
	if method.ReturnType == nil {
		ret := "void"
		if boxed {
			ret = "Void"
		}
		return ret
	}
	ret := g.getJavaTypeFromThriftType(method.ReturnType)
	if boxed {
		ret = containerType(ret)
	}
	return ret
}

func (g *Generator) generateArgs(args []*parser.Field, final bool) string {
	argStr := ""
	modifier := ""
	if final {
		modifier = "final "
	}
	for _, arg := range args {
		argStr += ", " + modifier + g.getJavaTypeFromThriftType(arg.Type) + " " + arg.Name
	}
	return argStr
}

func (g *Generator) generateClient(service *parser.Service, indent string) string {
	contents := ""
	if service.Extends != "" {
		contents += indent + fmt.Sprintf("public static class Client extends %s.Client implements Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += indent + "public static class Client implements Iface {\n\n"
	}
	if service.Extends == "" {
		if g.generateAsync() {
			contents += indent + tab + "protected ExecutorService asyncExecutor = Executors.newFixedThreadPool(2);\n"
		}
	}
	contents += indent + tab + "private Iface proxy;\n\n"

	contents += indent + tab + "public Client(FServiceProvider provider, ServiceMiddleware... middleware) {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "super(provider, middleware);\n"
	}
	contents += indent + tabtab + "Iface client = new InternalClient(provider);\n"
	contents += indent + tabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	contents += indent + tabtab + "combined.addAll(provider.getMiddleware());\n"
	contents += indent + tabtab + "middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	contents += indent + tabtab + "proxy = InvocationHandler.composeMiddleware(client, Iface.class, middleware);\n"
	contents += indent + tab + "}\n\n"

	for _, method := range service.Methods {
		if method.Comment != nil {
			contents += g.GenerateBlockComment(method.Comment, indent+tab)
		}

		_, deprecated := method.Annotations.Deprecated()
		if deprecated {
			contents += indent + tab + "@Deprecated\n"
		}

		contents += indent + tab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {\n",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))

		if deprecated && !g.suppressDeprecatedLogging() {
			contents += indent + tabtab + fmt.Sprintf("logger.warn(\"Call to deprecated function '%s.%s'\");\n", service.Name, method.Name)
		}

		if method.ReturnType != nil {
			contents += indent + tabtab + fmt.Sprintf("return proxy.%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		} else {
			contents += indent + tabtab + fmt.Sprintf("proxy.%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		}
		contents += indent + tab + "}\n\n"

		if g.generateAsync() {
			contents += g.generateAsyncClientMethod(service, method, indent)
		}
	}
	contents += indent + "}\n\n"
	contents += g.generateInternalClient(service, indent)
	return contents
}

func (g *Generator) generateAsyncClientMethod(service *parser.Service, method *parser.Method, indent string) string {
	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, indent+tab)
	}
	contents += indent + tab + fmt.Sprintf("public Future<%s> %sAsync(final FContext ctx%s) {\n",
		g.generateBoxedReturnValue(method), method.Name, g.generateArgs(method.Arguments, true))
	contents += indent + tabtab + fmt.Sprintf("return asyncExecutor.submit(new Callable<%s>() {\n", g.generateBoxedReturnValue(method))
	contents += indent + tabtabtab + fmt.Sprintf("public %s call() throws Exception {\n", g.generateBoxedReturnValue(method))
	if method.ReturnType != nil {
		contents += indent + tabtabtabtab + fmt.Sprintf("return %s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
	} else {
		contents += indent + tabtabtabtab + fmt.Sprintf("%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		contents += indent + tabtabtabtab + "return null;\n"
	}
	contents += indent + tabtabtab + "}\n"
	contents += indent + tabtab + "});\n"
	contents += indent + tab + "}\n\n"
	return contents
}

func (g *Generator) generateInternalClient(service *parser.Service, indent string) string {
	contents := ""
	if service.Extends != "" {
		contents += indent + fmt.Sprintf("private static class InternalClient extends %s.Client implements Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += indent + "private static class InternalClient implements Iface {\n\n"
	}

	contents += indent + tab + "private FTransport transport;\n"
	contents += indent + tab + "private FProtocolFactory protocolFactory;\n"

	contents += indent + tab + "public InternalClient(FServiceProvider provider) {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "super(provider);\n"
	}
	contents += indent + tabtab + "this.transport = provider.getTransport();\n"
	contents += indent + tabtab + "this.protocolFactory = provider.getProtocolFactory();\n"
	contents += indent + tab + "}\n\n"

	for _, method := range service.Methods {
		contents += g.generateClientMethod(service, method, indent)
	}
	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateClientMethod(service *parser.Service, method *parser.Method, indent string) string {
	methodLower := parser.LowercaseFirstLetter(method.Name)

	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, indent+tab)
	}
	contents += indent + tab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {\n",
		g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	contents += indent + tabtab + "TMemoryOutputBuffer memoryBuffer = new TMemoryOutputBuffer(this.transport.getRequestSizeLimit());\n"
	contents += indent + tabtab + "FProtocol oprot = this.protocolFactory.getProtocol(memoryBuffer);\n"
	contents += indent + tabtab + "oprot.writeRequestHeader(ctx);\n"
	msgType := "CALL"
	if method.Oneway {
		msgType = "ONEWAY"
	}
	contents += indent + tabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.%s, 0));\n", methodLower, msgType)
	contents += indent + tabtab + fmt.Sprintf("%s_args args = new %s_args();\n", method.Name, method.Name)
	for _, arg := range method.Arguments {
		contents += indent + tabtab + fmt.Sprintf("args.set%s(%s);\n", strings.Title(arg.Name), arg.Name)
	}
	contents += indent + tabtab + "args.write(oprot);\n"
	contents += indent + tabtab + "oprot.writeMessageEnd();\n"
	if method.Oneway {
		contents += indent + tabtab + "this.transport.oneway(ctx, memoryBuffer.getWriteBytes());\n"
	} else {
		contents += indent + tabtab + "TTransport response = this.transport.request(ctx, memoryBuffer.getWriteBytes());\n"
	}
	if method.Oneway {
		contents += indent + tab + "}\n"
		return contents
	}

	contents += "\n"
	contents += indent + tabtab + "FProtocol iprot = this.protocolFactory.getProtocol(response);\n"
	contents += indent + tabtab + "iprot.readResponseHeader(ctx);\n"
	contents += indent + tabtab + "TMessage message = iprot.readMessageBegin();\n"
	contents += indent + tabtab + fmt.Sprintf("if (!message.name.equals(\"%s\")) {\n", methodLower)
	contents += indent + tabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationExceptionType.WRONG_METHOD_NAME, \"%s failed: wrong method name\");\n",
		method.Name)
	contents += indent + tabtab + "}\n"
	contents += indent + tabtab + "if (message.type == TMessageType.EXCEPTION) {\n"
	contents += indent + tabtabtab + "TApplicationException e = TApplicationException.read(iprot);\n"
	contents += indent + tabtabtab + "iprot.readMessageEnd();\n"
	contents += indent + tabtabtab + "TException returnedException = e;\n"
	contents += indent + tabtabtab + "if (e.getType() == TApplicationExceptionType.RESPONSE_TOO_LARGE) {\n"
	contents += indent + tabtabtabtab + "returnedException = new TTransportException(TTransportExceptionType.RESPONSE_TOO_LARGE, e.getMessage());\n"
	contents += indent + tabtabtab + "}\n"
	contents += indent + tabtabtab + "throw returnedException;\n"
	contents += indent + tabtab + "}\n"
	contents += indent + tabtab + "if (message.type != TMessageType.REPLY) {\n"
	contents += indent + tabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationExceptionType.INVALID_MESSAGE_TYPE, \"%s failed: invalid message type\");\n",
		method.Name)
	contents += indent + tabtab + "}\n"
	contents += indent + tabtab + fmt.Sprintf("%s_result res = new %s_result();\n", method.Name, method.Name)
	contents += indent + tabtab + "res.read(iprot);\n"
	contents += indent + tabtab + "iprot.readMessageEnd();\n"
	if method.ReturnType != nil {
		contents += indent + tabtab + "if (res.isSetSuccess()) {\n"
		contents += indent + tabtabtab + "return res.success;\n"
		contents += indent + tabtab + "}\n"
	}
	for _, exception := range method.Exceptions {
		contents += indent + tabtab + fmt.Sprintf("if (res.%s != null) {\n", exception.Name)
		contents += indent + tabtabtab + fmt.Sprintf("throw res.%s;\n", exception.Name)
		contents += indent + tabtab + "}\n"
	}
	if method.ReturnType != nil {
		contents += indent + tabtab + fmt.Sprintf(
			"throw new TApplicationException(TApplicationExceptionType.MISSING_RESULT, \"%s failed: unknown result\");\n",
			method.Name)
	}
	contents += indent + tab + "}\n"

	return contents
}

func (g *Generator) generateExceptions(exceptions []*parser.Field) string {
	contents := "throws TException"
	for _, exception := range exceptions {
		contents += ", " + g.getJavaTypeFromThriftType(exception.Type)
	}
	return contents
}

func (g *Generator) generateServer(service *parser.Service, indent string) string {
	contents := ""
	extends := "FBaseProcessor"
	if service.Extends != "" {
		extends = g.getServiceExtendsName(service) + ".Processor"
	}
	contents += indent + fmt.Sprintf("public static class Processor extends %s implements FProcessor {\n\n", extends)

	contents += indent + tab + "private Iface handler;\n\n"

	contents += indent + tab + "public Processor(Iface iface, ServiceMiddleware... middleware) {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "super(iface, middleware);\n"
	}
	contents += indent + tabtab + "handler = InvocationHandler.composeMiddleware(iface, Iface.class, middleware);\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "protected java.util.Map<String, FProcessorFunction> getProcessMap() {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "java.util.Map<String, FProcessorFunction> processMap = super.getProcessMap();\n"
	} else {
		contents += indent + tabtab + "java.util.Map<String, FProcessorFunction> processMap = new java.util.HashMap<>();\n"
	}
	for _, method := range service.Methods {
		contents += indent + tabtab + fmt.Sprintf("processMap.put(\"%s\", new %s());\n", parser.LowercaseFirstLetter(method.Name), strings.Title(method.Name))
	}
	contents += indent + tabtab + "return processMap;\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "protected java.util.Map<String, java.util.Map<String, String>> getAnnotationsMap() {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "java.util.Map<String, java.util.Map<String, String>> annotationsMap = super.getAnnotationsMap();\n"
	} else {
		contents += indent + tabtab + "java.util.Map<String, java.util.Map<String, String>> annotationsMap = new java.util.HashMap<>();\n"
	}
	for _, method := range service.Methods {
		if len(method.Annotations) > 0 {
			contents += indent + tabtab + fmt.Sprintf("java.util.Map<String, String> %sMap = new java.util.HashMap<>();\n", method.Name)
			for _, annotation := range method.Annotations {
				contents += indent + tabtab + fmt.Sprintf("%sMap.put(\"%s\", %s);\n", method.Name, annotation.Name, g.quote(annotation.Value))
			}
			contents += indent + tabtab + fmt.Sprintf("annotationsMap.put(\"%s\", %sMap);\n", parser.LowercaseFirstLetter(method.Name), method.Name)
		}
	}
	contents += indent + tabtab + "return annotationsMap;\n"
	contents += indent + tab + "}\n\n"

	contents += indent + tab + "@Override\n"
	contents += indent + tab + "public void addMiddleware(ServiceMiddleware middleware) {\n"
	if service.Extends != "" {
		contents += indent + tabtab + "super.addMiddleware(middleware);\n"
	}
	contents += indent + tabtab + "handler = InvocationHandler.composeMiddleware(handler, Iface.class, new ServiceMiddleware[]{middleware});\n"
	contents += indent + tab + "}\n\n"

	for _, method := range service.Methods {
		methodLower := parser.LowercaseFirstLetter(method.Name)
		contents += indent + tab + fmt.Sprintf("private class %s implements FProcessorFunction {\n\n", strings.Title(method.Name))

		contents += indent + tabtab + "public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {\n"

		if _, ok := method.Annotations.Deprecated(); ok && !g.suppressDeprecatedLogging() {
			contents += indent + tabtabtab + fmt.Sprintf("logger.warn(\"Deprecated function '%s.%s' was called by a client\");\n", service.Name, method.Name)
		}

		contents += indent + tabtabtab + fmt.Sprintf("%s_args args = new %s_args();\n", method.Name, method.Name)
		contents += indent + tabtabtab + "try {\n"
		contents += indent + tabtabtabtab + "args.read(iprot);\n"
		contents += indent + tabtabtab + "} catch (TException e) {\n"
		contents += indent + tabtabtabtab + "iprot.readMessageEnd();\n"
		if !method.Oneway {
			contents += indent + tabtabtabtab + "synchronized (WRITE_LOCK) {\n"
			contents += indent + tabtabtabtabtab + fmt.Sprintf("e = writeApplicationException(ctx, oprot, TApplicationExceptionType.PROTOCOL_ERROR, \"%s\", e.getMessage());\n", method.Name)
			contents += indent + tabtabtabtab + "}\n"
		}
		contents += indent + tabtabtabtab + "throw e;\n"
		contents += indent + tabtabtab + "}\n\n"

		contents += indent + tabtabtab + "iprot.readMessageEnd();\n"

		if method.Oneway {
			contents += indent + tabtabtab + fmt.Sprintf("handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += indent + tabtab + "}\n"
			contents += indent + tab + "}\n\n"
			continue
		}

		contents += indent + tabtabtab + fmt.Sprintf("%s_result result = new %s_result();\n", method.Name, method.Name)
		contents += indent + tabtabtab + "try {\n"
		if method.ReturnType == nil {
			contents += indent + tabtabtabtab + fmt.Sprintf("handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
		} else {
			contents += indent + tabtabtabtab + fmt.Sprintf("result.success = handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += indent + tabtabtabtab + "result.setSuccessIsSet(true);\n"
		}
		for _, exception := range method.Exceptions {
			contents += indent + tabtabtab + fmt.Sprintf("} catch (%s %s) {\n", g.getJavaTypeFromThriftType(exception.Type), exception.Name)
			contents += indent + tabtabtabtab + fmt.Sprintf("result.%s = %s;\n", exception.Name, exception.Name)
		}
		contents += indent + tabtabtab + "} catch (TApplicationException e) {\n"
		contents += indent + tabtabtabtab + "oprot.writeResponseHeader(ctx);\n"
		contents += indent + tabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.EXCEPTION, 0));\n", methodLower)
		contents += indent + tabtabtabtab + "e.write(oprot);\n"
		contents += indent + tabtabtabtab + "oprot.writeMessageEnd();\n"
		contents += indent + tabtabtabtab + "oprot.getTransport().flush();\n"
		contents += indent + tabtabtabtab + "return;\n"
		contents += indent + tabtabtab + "} catch (TException e) {\n"
		contents += indent + tabtabtabtab + "synchronized (WRITE_LOCK) {\n"
		contents += indent + tabtabtabtabtab + fmt.Sprintf(
			"e = (TApplicationException) writeApplicationException(ctx, oprot, TApplicationExceptionType.INTERNAL_ERROR, \"%s\", \"Internal error processing %s: \" + e.getMessage()).initCause(e);\n",
			methodLower, method.Name)
		contents += indent + tabtabtabtab + "}\n"
		contents += indent + tabtabtabtab + "throw e;\n"
		contents += indent + tabtabtab + "}\n"
		contents += indent + tabtabtab + "synchronized (WRITE_LOCK) {\n"
		contents += indent + tabtabtabtab + "try {\n"
		contents += indent + tabtabtabtabtab + "oprot.writeResponseHeader(ctx);\n"
		contents += indent + tabtabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.REPLY, 0));\n", methodLower)
		contents += indent + tabtabtabtabtab + "result.write(oprot);\n"
		contents += indent + tabtabtabtabtab + "oprot.writeMessageEnd();\n"
		contents += indent + tabtabtabtabtab + "oprot.getTransport().flush();\n"
		contents += indent + tabtabtabtab + "} catch (TTransportException e) {\n"
		contents += indent + tabtabtabtabtab + "if (e.getType() == TTransportExceptionType.REQUEST_TOO_LARGE) {\n"
		contents += indent + tabtabtabtabtabtab + fmt.Sprintf(
			"writeApplicationException(ctx, oprot, TApplicationExceptionType.RESPONSE_TOO_LARGE, \"%s\", \"response too large: \" + e.getMessage());\n",
			methodLower)
		contents += indent + tabtabtabtabtab + "} else {\n"
		contents += indent + tabtabtabtabtabtab + "throw e;\n"
		contents += indent + tabtabtabtabtab + "}\n"
		contents += indent + tabtabtabtab + "}\n"
		contents += indent + tabtabtab + "}\n"
		contents += indent + tabtab + "}\n"
		contents += indent + tab + "}\n\n"
	}

	contents += indent + "}\n\n"

	return contents
}

func (g *Generator) generateScopeArgs(scope *parser.Scope) string {
	args := "ctx"
	for _, v := range scope.Prefix.Variables {
		args += ", " + v
	}
	args += ", req"
	return args
}

func (g *Generator) generateClientCallArgs(args []*parser.Field) string {
	return g.generateCallArgs(args, "")
}

func (g *Generator) generateServerCallArgs(args []*parser.Field) string {
	return g.generateCallArgs(args, "args.")
}

func (g *Generator) generateCallArgs(args []*parser.Field, prefix string) string {
	contents := "ctx"
	prefix = ", " + prefix
	for _, arg := range args {
		contents += prefix + arg.Name
	}
	return contents
}

func (g *Generator) getJavaTypeFromThriftType(t *parser.Type) string {
	javaType := g._getJavaType(t, true)
	if g.generateBoxedPrimitives() {
		return containerType(javaType)
	}
	return javaType
}

func (g *Generator) getUnparametrizedJavaType(t *parser.Type) string {
	return g._getJavaType(t, false)
}

func (g *Generator) _getJavaType(t *parser.Type, parametrized bool) string {
	if t == nil {
		return "void"
	}
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool":
		return "boolean"
	case "byte", "i8":
		return "byte"
	case "i16":
		return "short"
	case "i32":
		return "int"
	case "i64":
		return "long"
	case "double":
		return "double"
	case "string":
		return "String"
	case "binary":
		return "java.nio.ByteBuffer"
	case "list":
		if parametrized {
			return fmt.Sprintf("java.util.List<%s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.List"
	case "set":
		if parametrized {
			return fmt.Sprintf("java.util.Set<%s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.Set"
	case "map":
		if parametrized {
			return fmt.Sprintf("java.util.Map<%s, %s>",
				containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)),
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
		}
		return "java.util.Map"
	default:
		// This is a custom type, return a pointer to it
		return g.qualifiedTypeName(t)
	}
}

func (g *Generator) getTType(t *parser.Type) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	ttype := ""
	switch underlyingType.Name {
	case "bool", "byte", "double", "i16", "i32", "i64", "list", "set", "map", "string":
		ttype = strings.ToUpper(underlyingType.Name)
	case "binary":
		ttype = "STRING"
	default:
		if g.Frugal.IsStruct(t) {
			ttype = "STRUCT"
		} else if g.Frugal.IsEnum(t) {
			ttype = "I32"
		} else {
			panic("shouldn't happen: " + underlyingType.Name)
		}
	}

	return fmt.Sprintf("org.apache.thrift.protocol.TType.%s", ttype)
}

func (g *Generator) canBeJavaPrimitive(t *parser.Type) bool {
	underlyingType := g.Frugal.UnderlyingType(t)
	switch underlyingType.Name {
	case "bool", "byte", "i8", "i16", "i32", "i64", "double":
		return true
	default:
		return false
	}
}

func (g *Generator) isJavaPrimitive(t *parser.Type) bool {
	if g.generateBoxedPrimitives() {
		// If boxing primitives, nothing is a primitive
		return false
	}

	return g.canBeJavaPrimitive(t)
}

func (g *Generator) getPrimitiveDefaultValue(t *parser.Type) string {
	switch name := g.Frugal.UnderlyingType(t).Name; name {
	case "bool":
		return "false"
	case "i8", "byte":
		return "(byte)0"
	case "i16":
		return "(short)0"
	case "i32":
		return "0"
	case "i64":
		return "0L"
	case "double":
		return "0.0"
	default:
		panic(fmt.Sprintf("%s is not a primitive", name))
	}
}

func (g *Generator) generateBoxedPrimitives() bool {
	_, ok := g.Options["boxed_primitives"]
	return ok
}

func containerType(typeName string) string {
	switch typeName {
	case "int":
		return "Integer"
	case "boolean", "byte", "short", "long", "double", "void":
		return strings.Title(typeName)
	default:
		return typeName
	}
}

func (g *Generator) isVendoredInclude(includeName string) bool {
	include := g.Frugal.Include(includeName)
	if include == nil {
		return false
	}
	_, vendored := include.Annotations.Vendor()
	return vendored
}

func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()
	if namespace := g.namespaceForInclude(include); namespace != "" {
		return fmt.Sprintf("%s.%s", namespace, param)
	}
	return param
}

func toConstantName(name string) string {
	// TODO fix for identifiers like "ID2"
	ret := ""
	tmp := []rune(name)
	is_prev_lc := true
	is_current_lc := tmp[0] == unicode.ToLower(tmp[0])
	is_next_lc := false

	for i, _ := range tmp {
		lc := unicode.ToLower(tmp[i])

		if i == len(name)-1 {
			is_next_lc = false
		} else {
			is_next_lc = (tmp[i+1] == unicode.ToLower(tmp[i+1]))
		}

		if i != 0 && !is_current_lc && (is_prev_lc || is_next_lc) {
			ret += "_"
		}
		ret += string(lc)

		is_prev_lc = is_current_lc
		is_current_lc = is_next_lc
	}
	return strings.ToUpper(ret)
}

func (g *Generator) includeGeneratedAnnotation() bool {
	return g.Options[generatedAnnotations] != "suppress"
}

func (g *Generator) generatedAnnotation(indent string) string {
	anno := indent + fmt.Sprintf("@Generated(value = \"Autogenerated by Frugal Compiler (%s)\"", globals.Version)
	if g.Options[generatedAnnotations] != "undated" {
		anno += fmt.Sprintf(", "+"date = \"%s\"", g.time.Format("2006-1-2"))
	}
	anno += ")\n"
	return anno
}

func (g *Generator) generateAsync() bool {
	_, ok := g.Options["async"]
	return ok
}

func (g *Generator) UseVendor() bool {
	_, ok := g.Options[useVendorOption]
	return ok
}
