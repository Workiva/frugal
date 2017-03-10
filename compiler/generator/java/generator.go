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
	tab                         = "\t"
	generatedAnnotations        = "generated_annotations"
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
		contents += g.generatedAnnotation()
	}
	contents += fmt.Sprintf("public class %sConstants {\n", g.Frugal.Name)

	for _, constant := range constants {
		val := g.generateConstantValueWrapper(constant.Name, constant.Type, constant.Value, true, true)
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
func (g *Generator) generateConstantValueWrapper(fieldName string, t *parser.Type, value interface{}, declare, needsStatic bool) string {
	underlyingType := g.Frugal.UnderlyingType(t)
	contents := tabtab

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
		_, val := g.generateConstantValueRec(t, value)
		return fmt.Sprintf("%s%s = %s;\n", contents, fieldName, val)
	} else if g.Frugal.IsStruct(underlyingType) {
		var s *parser.Struct
		for _, potential := range g.Frugal.Structs {
			if underlyingType.Name == potential.Name {
				s = potential
				break
			}
		}

		contents += fmt.Sprintf("%s = new %s();\n", fieldName, g.getJavaTypeFromThriftType(underlyingType))

		ind := tabtab
		if needsStatic {
			contents += ind + "static {\n"
			ind += tab
		}

		for _, pair := range value.([]parser.KeyValue) {
			name := pair.Key.(string)
			for _, field := range s.Fields {
				if name == field.Name {
					preamble, val := g.generateConstantValueRec(field.Type, pair.Value)
					contents += preamble
					contents += fmt.Sprintf(ind+"%s.set%s(%s);\n", fieldName, strings.Title(name), val)
				}
			}
		}

		if needsStatic {
			contents += tabtab + "}\n"
		}

		return contents
	} else if underlyingType.IsContainer() {
		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf("%s = new ArrayList<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
			}
			return contents
		case "set":
			contents += fmt.Sprintf("%s = new HashSet<%s>();\n", fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += ind + "static {\n"
				ind += tab
			}

			for _, v := range value.([]interface{}) {
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, v)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.add(%s);\n", fieldName, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
			}
			return contents
		case "map":
			contents += fmt.Sprintf("%s = new HashMap<%s,%s>();\n",
				fieldName, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)),
				containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)))
			ind := tabtab
			if needsStatic {
				contents += tabtab + "static {\n"
				ind += tab
			}

			for _, pair := range value.([]parser.KeyValue) {
				preamble, key := g.generateConstantValueRec(underlyingType.KeyType, pair.Key)
				contents += preamble
				preamble, val := g.generateConstantValueRec(underlyingType.ValueType, pair.Value)
				contents += preamble
				contents += fmt.Sprintf(ind+"%s.put(%s, %s);\n", fieldName, key, val)
			}

			if needsStatic {
				contents += tabtab + "}\n"
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

func (g *Generator) generateConstantValueRec(t *parser.Type, value interface{}) (string, string) {
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
			if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
				include = namespace.Value
			}
			return "", fmt.Sprintf("%s.%sConstants.%s", include, idCtx.Include.Name, idCtx.Constant.Name)
		case parser.IncludeEnum:
			include := idCtx.Include.Name
			if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
				include = namespace.Value
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
			return "", fmt.Sprintf("%v", strconv.Quote(value.(string)))
		case "binary":
			return "", fmt.Sprintf("java.nio.ByteBuffer.wrap(\"%v\".getBytes())", value)
		}
	} else if g.Frugal.IsEnum(underlyingType) {
		return "", g.generateEnumConstFromValue(underlyingType, int(value.(int64)))
	}
	elem := g.GetElem()
	preamble := g.generateConstantValueWrapper(elem, t, value, true, false)
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
		contents += fmt.Sprintf(tab+"%s(%d)%s\n", value.Name, value.Value, terminator)
	}
	contents += "\n"

	contents += tab + "private final int value;\n\n"
	contents += fmt.Sprintf(tab+"private %s(int value) {\n", enum.Name)
	contents += tabtab + "this.value = value;\n"
	contents += tab + "}\n\n"

	contents += tab + "public int getValue() {\n"
	contents += tabtab + "return value;\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public static %s findByValue(int value) {\n", enum.Name)
	contents += tabtab + "switch (value) {\n"
	for _, value := range enum.Values {
		contents += fmt.Sprintf(tabtabtab+"case %d:\n", value.Value)
		contents += fmt.Sprintf(tabtabtabtab+"return %s;\n", value.Name)
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

	if err := g.GenerateStructImports(file); err != nil {
		return err
	}

	return nil
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

	_, err = file.WriteString(g.generateStruct(s, false, false))
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
		contents += g.generatedAnnotation()
	}

	static := ""
	if isArg || isResult {
		static = "static "
	}
	contents += fmt.Sprintf("public %sclass %s extends org.apache.thrift.TUnion<%s, %s._Fields> {\n",
		static, union.Name, union.Name, union.Name)

	contents += g.generateDescriptors(union)
	contents += g.generateFieldsEnum(union)
	contents += g.generateUnionConstructors(union)
	contents += g.generateUnionFieldConstructors(union)
	contents += g.generateUnionCheckType(union)

	contents += g.generateUnionStandardRead(union)
	contents += g.generateUnionStandardWrite(union)
	contents += g.generateUnionTupleRead(union)
	contents += g.generateUnionTupleWrite(union)

	contents += g.generateUnionGetDescriptors(union)
	contents += g.generateUnionFieldForId()
	contents += g.generateUnionGetSetFields(union)
	contents += g.generateUnionIsSetFields(union)

	contents += g.generateUnionEquals(union)
	contents += g.generateUnionCompareTo(union)
	contents += g.generateUnionHashCode(union)
	contents += g.generateWriteObject(union)
	contents += g.generateReadObject(union)

	contents += "}\n"
	return contents
}

func (g *Generator) generateUnionConstructors(union *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s() {\n", union.Name)
	contents += tabtab + "super();\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public %s(_Fields setField, Object value) {\n", union.Name)
	contents += tabtab + "super(setField, value);\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public %s(%s other) {\n", union.Name, union.Name)
	contents += tabtab + "super(other);\n"
	contents += tab + "}\n"

	contents += fmt.Sprintf(tab+"public %s deepCopy() {\n", union.Name)
	contents += fmt.Sprintf(tabtab+"return new %s(this);\n", union.Name)
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionFieldConstructors(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		contents += fmt.Sprintf(tab+"public static %s %s(%s value) {\n",
			union.Name, field.Name, g.getJavaTypeFromThriftType(field.Type))
		contents += fmt.Sprintf(tabtab+"%s x = new %s();\n", union.Name, union.Name)
		contents += fmt.Sprintf(tabtab+"x.set%s(value);\n", strings.Title(field.Name))
		contents += tabtab + "return x;\n"
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionCheckType(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected void checkType(_Fields setField, Object value) throws ClassCastException {\n"
	contents += tabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		fieldType := containerType(g.getJavaTypeFromThriftType(field.Type))
		unparametrizedType := containerType(g.getUnparametrizedJavaType(field.Type))
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtab+"if (value instanceof %s) {\n", unparametrizedType)
		contents += tabtabtabtabtab + "break;\n"
		contents += tabtabtabtab + "}\n"
		contents += fmt.Sprintf(tabtabtabtab+"throw new ClassCastException(\"Was expecting value of type %s for field '%s', but got \" + value.getClass().getSimpleName());\n",
			fieldType, field.Name)
	}
	contents += tabtabtab + "default:\n"
	contents += fmt.Sprintf(tabtabtabtab + "throw new IllegalArgumentException(\"Unknown field id \" + setField);\n")
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardRead(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected Object standardSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, org.apache.thrift.protocol.TField field) throws org.apache.thrift.TException {\n"
	contents += tabtab + "_Fields setField = _Fields.findByThriftId(field.id);\n"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtabtab+"if (field.type == %s_FIELD_DESC.type) {\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, tabtabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtabtab+"return %s;\n", field.Name)
		contents += tabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
		contents += tabtabtabtabtabtab + "return null;\n"
		contents += tabtabtabtabtab + "}\n"
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, field.type);\n"
	contents += tabtabtab + "return null;\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionStandardWrite(union *parser.Struct) string {
	return g.generateUnionWrite(union, "standard")
}

func (g *Generator) generateUnionTupleRead(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "protected Object tupleSchemeReadValue(org.apache.thrift.protocol.TProtocol iprot, short fieldID) throws org.apache.thrift.TException {\n"
	contents += tabtab + "_Fields setField = _Fields.findByThriftId(fieldID);\n"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "switch (setField) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtabtab+"case %s:\n", constantName)
		contents += g.generateReadFieldRec(field, false, false, true, tabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtab+"return %s;\n", field.Name)
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "throw new IllegalStateException(\"setField wasn't null, but didn't match any of the case statements!\");\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "throw new TProtocolException(\"Couldn't find a field with field id \" + fieldID);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionTupleWrite(union *parser.Struct) string {
	return g.generateUnionWrite(union, "tuple")
}

func (g *Generator) generateUnionWrite(union *parser.Struct, scheme string) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"protected void %sSchemeWriteValue(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {\n", scheme)
	contents += tabtab + "switch (setField_) {\n"
	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		javaContainerType := containerType(g.getJavaTypeFromThriftType(field.Type))
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtab+"%s %s = (%s)value_;\n", javaContainerType, field.Name, javaContainerType)
		contents += g.generateWriteFieldRec(field, false, false, tabtabtabtab)
		contents += tabtabtabtab + "return;\n"
	}
	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new IllegalStateException(\"Cannot write union with unknown field \" + setField_);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionGetDescriptors(union *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "protected org.apache.thrift.protocol.TField getFieldDesc(_Fields setField) {\n"
	contents += tabtab + "switch (setField) {\n"

	for _, field := range union.Fields {
		constantName := toConstantName(field.Name)
		contents += fmt.Sprintf(tabtabtab+"case %s:\n", constantName)
		contents += fmt.Sprintf(tabtabtabtab+"return %s_FIELD_DESC;\n", constantName)
	}

	contents += tabtabtab + "default:\n"
	contents += tabtabtabtab + "throw new IllegalArgumentException(\"Unknown field id \" + setField);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += tab + "@Override\n"
	contents += tab + "protected org.apache.thrift.protocol.TStruct getStructDesc() {\n"
	contents += tabtab + "return STRUCT_DESC;\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateUnionFieldForId() string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "protected _Fields enumForId(short id) {\n"
	contents += tabtab + "return _Fields.findByThriftIdOrThrow(id);\n"
	contents += tab + "}\n\n"

	contents += tab + "public _Fields fieldForId(int fieldId) {\n"
	contents += tabtab + "return _Fields.findByThriftId(fieldId);\n"
	contents += tab + "}\n\n\n"
	return contents
}

func (g *Generator) generateUnionGetSetFields(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		titleName := strings.Title(field.Name)
		constantName := toConstantName(field.Name)
		javaType := g.getJavaTypeFromThriftType(field.Type)

		// get
		contents += fmt.Sprintf(tab+"public %s get%s() {\n", javaType, titleName)
		contents += fmt.Sprintf(tabtab+"if (getSetField() == _Fields.%s) {\n", constantName)
		contents += fmt.Sprintf(tabtabtab+"return (%s)getFieldValue();\n", containerType(javaType))
		contents += tabtab + "} else {\n"
		contents += fmt.Sprintf(tabtabtab+"throw new RuntimeException(\"Cannot get field '%s' because union is currently set to \" + getFieldDesc(getSetField()).name);\n", field.Name)
		contents += tabtab + "}\n"
		contents += tab + "}\n\n"

		// set
		contents += fmt.Sprintf(tab+"public void set%s(%s value) {\n", titleName, javaType)
		if !g.isJavaPrimitive(field.Type) {
			contents += tabtab + "if (value == null) throw new NullPointerException();\n"
		}
		contents += fmt.Sprintf(tabtab+"setField_ = _Fields.%s;\n", constantName)
		contents += tabtab + "value_ = value;\n"
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionIsSetFields(union *parser.Struct) string {
	contents := ""

	for _, field := range union.Fields {
		contents += fmt.Sprintf(tab+"public boolean isSet%s() {\n", strings.Title(field.Name))
		contents += fmt.Sprintf(tabtab+"return setField_ == _Fields.%s;\n", toConstantName(field.Name))
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateUnionEquals(union *parser.Struct) string {
	contents := "\n"

	contents += tab + "public boolean equals(Object other) {\n"
	contents += fmt.Sprintf(tabtab+"if (other instanceof %s) {\n", union.Name)
	contents += fmt.Sprintf(tabtabtab+"return equals((%s)other);\n", union.Name)
	contents += tabtab + "} else {\n"
	contents += tabtabtab + "return false;\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public boolean equals(%s other) {\n", union.Name)
	contents += tabtab + "return other != null && getSetField() == other.getSetField() && getFieldValue().equals(other.getFieldValue());\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnionCompareTo(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"public int compareTo(%s other) {\n", union.Name)
	contents += tabtab + "int lastComparison = org.apache.thrift.TBaseHelper.compareTo(getSetField(), other.getSetField());\n"
	contents += tabtab + "if (lastComparison == 0) {\n"
	contents += tabtabtab + "return org.apache.thrift.TBaseHelper.compareTo(getFieldValue(), other.getFieldValue());\n"
	contents += tabtab + "}\n"
	contents += tabtab + "return lastComparison;\n"
	contents += tab + "}\n\n\n"

	return contents
}

func (g *Generator) generateUnionHashCode(union *parser.Struct) string {
	contents := ""

	contents += tab + "@Override\n"
	contents += tab + "public int hashCode() {\n"
	contents += tabtab + "List<Object> list = new ArrayList<Object>();\n"
	contents += tabtab + "list.add(this.getClass().getName());\n"
	contents += tabtab + "org.apache.thrift.TFieldIdEnum setField = getSetField();\n"
	contents += tabtab + "if (setField != null) {\n"
	contents += tabtabtab + "list.add(setField.getThriftFieldId());\n"
	contents += tabtabtab + "Object value = getFieldValue();\n"
	contents += tabtabtab + "if (value instanceof org.apache.thrift.TEnum) {\n"
	contents += tabtabtabtab + "list.add(((org.apache.thrift.TEnum)getFieldValue()).getValue());\n"
	contents += tabtabtab + "} else {\n"
	contents += tabtabtabtab + "list.add(value);\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n"
	contents += tabtab + "return list.hashCode();\n"
	contents += tab + "}\n"

	return contents
}

func (g *Generator) GenerateException(exception *parser.Struct) error {
	return g.GenerateStruct(exception)
}

// generateServiceArgsResults generates the args and results objects for the
// given service.
func (g *Generator) generateServiceArgsResults(service *parser.Service) string {
	contents := ""
	for _, s := range g.GetServiceMethodTypes(service) {
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				field.Modifier = parser.Default
			}
		}
		contents += g.generateStruct(s, strings.HasSuffix(s.Name, "_args"), strings.HasSuffix(s.Name, "_result"))
	}
	return contents
}

func (g *Generator) generateStruct(s *parser.Struct, isArg, isResult bool) string {
	contents := ""

	if s.Comment != nil {
		contents += g.GenerateBlockComment(s.Comment, "")
	}
	if g.includeGeneratedAnnotation() && !isArg && !isResult {
		contents += g.generatedAnnotation()
	}
	static := ""
	if isArg || isResult {
		static = "static "
	}
	exception := ""
	if s.Type == parser.StructTypeException {
		exception = "extends TException "
	}
	contents += fmt.Sprintf("public %sclass %s %simplements org.apache.thrift.TBase<%s, %s._Fields>, java.io.Serializable, Cloneable, Comparable<%s> {\n",
		static, s.Name, exception, s.Name, s.Name, s.Name)

	contents += g.generateDescriptors(s)

	contents += g.generateSchemeMap(s)

	contents += g.generateInstanceVars(s)

	contents += g.generateFieldsEnum(s)

	contents += g.generateIsSetVars(s)

	contents += g.generateDefaultConstructor(s)
	contents += g.generateFullConstructor(s)
	contents += g.generateCopyConstructor(s)
	contents += g.generateDeepCopyMethod(s)
	contents += g.generateClear(s)

	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)
		if underlyingType.IsContainer() {
			contents += g.generateContainerGetSize(field)
			contents += g.generateContainerIterator(field)
			contents += g.generateContainerAddTo(field)

		}

		contents += g.generateGetField(field)
		contents += g.generateSetField(s.Name, field)
		contents += g.generateUnsetField(s, field)
		contents += g.generateIsSetField(s, field)
		contents += g.generateSetIsSetField(s, field)
	}

	contents += g.generateSetValue(s)
	contents += g.generateGetValue(s)
	contents += g.generateIsSetValue(s)

	contents += g.generateEquals(s)
	contents += g.generateHashCode(s)
	contents += g.generateCompareTo(s)

	contents += g.generateFieldForId(s)
	contents += g.generateReadWrite(s)

	contents += g.generateToString(s)
	contents += g.generateValidate(s)

	contents += g.generateWriteObject(s)
	contents += g.generateReadObject(s)

	contents += g.generateStandardScheme(s, isResult)
	contents += g.generateTupleScheme(s)

	contents += "}\n"
	return contents
}

func (g *Generator) generateDescriptors(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static final org.apache.thrift.protocol.TStruct STRUCT_DESC = new org.apache.thrift.protocol.TStruct(\"%s\");\n\n",
		s.Name)
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tab+"private static final org.apache.thrift.protocol.TField %s_FIELD_DESC = new org.apache.thrift.protocol.TField(\"%s\", %s, (short)%d);\n",
			toConstantName(field.Name), field.Name, g.getTType(field.Type), field.ID)
	}
	contents += "\n"
	return contents
}

func (g *Generator) generateSchemeMap(s *parser.Struct) string {
	contents := ""
	contents += tab + "private static final Map<Class<? extends IScheme>, SchemeFactory> schemes = new HashMap<Class<? extends IScheme>, SchemeFactory>();\n"
	contents += tab + "static {\n"
	contents += fmt.Sprintf(tabtab+"schemes.put(StandardScheme.class, new %sStandardSchemeFactory());\n", s.Name)
	contents += fmt.Sprintf(tabtab+"schemes.put(TupleScheme.class, new %sTupleSchemeFactory());\n", s.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateInstanceVars(s *parser.Struct) string {
	contents := ""
	for _, field := range s.Fields {
		if field.Comment != nil {
			contents += g.GenerateBlockComment(field.Comment, tab)
		}
		modifier := "required"
		if field.Modifier == parser.Optional {
			modifier = "optional"
		}
		contents += fmt.Sprintf(tab+"public %s %s; // %s\n",
			g.getJavaTypeFromThriftType(field.Type), field.Name, modifier)
	}
	return contents
}

func (g *Generator) generateFieldsEnum(s *parser.Struct) string {
	contents := ""
	contents += tab + "/** The set of fields this struct contains, along with convenience methods for finding and manipulating them. */\n"
	contents += tab + "public enum _Fields implements org.apache.thrift.TFieldIdEnum {\n"

	for idx, field := range s.Fields {
		terminator := ""
		if idx != len(s.Fields)-1 {
			terminator = ","
		}

		if field.Comment != nil {
			contents += g.GenerateBlockComment(field.Comment, tabtab)
		}
		contents += fmt.Sprintf(tabtab+"%s((short)%d, \"%s\")%s\n", toConstantName(field.Name), field.ID, field.Name, terminator)
	}
	// Do it this was as the semi colon is needed no matter what
	contents += ";\n"
	contents += "\n"

	contents += tabtab + "private static final Map<String, _Fields> byName = new HashMap<String, _Fields>();\n\n"
	contents += tabtab + "static {\n"
	contents += tabtabtab + "for (_Fields field : EnumSet.allOf(_Fields.class)) {\n"
	contents += tabtabtabtab + "byName.put(field.getFieldName(), field);\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches fieldId, or null if its not found."}, tabtab)
	contents += tabtab + "public static _Fields findByThriftId(int fieldId) {\n"
	contents += tabtabtab + "switch(fieldId) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtab+"case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtabtab+"return %s;\n", toConstantName(field.Name))
	}
	contents += tabtabtabtab + "default:\n"
	contents += tabtabtabtabtab + "return null;\n"
	contents += tabtabtab + "}\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{
		"Find the _Fields constant that matches fieldId, throwing an exception",
		"if it is not found.",
	}, tabtab)
	contents += tabtab + "public static _Fields findByThriftIdOrThrow(int fieldId) {\n"
	contents += tabtabtab + "_Fields fields = findByThriftId(fieldId);\n"
	contents += tabtabtab + "if (fields == null) throw new IllegalArgumentException(\"Field \" + fieldId + \" doesn't exist!\");\n"
	contents += tabtabtab + "return fields;\n"
	contents += tabtab + "}\n\n"

	contents += g.GenerateBlockComment([]string{"Find the _Fields constant that matches name, or null if its not found."}, tabtab)
	contents += tabtab + "public static _Fields findByName(String name) {\n"
	contents += tabtabtab + "return byName.get(name);\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "private final short _thriftId;\n"
	contents += tabtab + "private final String _fieldName;\n\n"

	contents += tabtab + "_Fields(short thriftId, String fieldName) {\n"
	contents += tabtabtab + "_thriftId = thriftId;\n"
	contents += tabtabtab + "_fieldName = fieldName;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "public short getThriftFieldId() {\n"
	contents += tabtabtab + "return _thriftId;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "public String getFieldName() {\n"
	contents += tabtabtab + "return _fieldName;\n"
	contents += tabtab + "}\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetVars(s *parser.Struct) string {
	contents := ""
	contents += tab + "// isset id assignments\n"
	primitiveCount := 0
	for _, field := range s.Fields {
		if g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tab+"private static final int %s = %d;\n",
				g.getIsSetID(field.Name), primitiveCount)
			primitiveCount += 1
		}
	}
	isSetType, bitFieldType := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
	// Do nothing
	case IsSetBitfield:
		contents += fmt.Sprintf(tab+"private %s __isset_bitfield = 0;\n", bitFieldType)
	case IsSetBitSet:
		contents += fmt.Sprintf(tab+"private BitSet __isset_bit_vector = new BitSet(%d);\n", primitiveCount)
	}
	return contents
}

func (g *Generator) generateDefaultConstructor(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s() {\n", s.Name)
	for _, field := range s.Fields {
		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false)
			contents += fmt.Sprintf("%s\n", val)
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateFullConstructor(s *parser.Struct) string {
	contents := ""
	argsList := ""
	sep := "\n" + tabtab
	numNonOptional := 0
	for _, field := range s.Fields {
		if field.Modifier == parser.Optional {
			continue
		}
		argsList += fmt.Sprintf("%s%s %s", sep, g.getJavaTypeFromThriftType(field.Type), field.Name)
		sep = ",\n" + tabtab
		numNonOptional += 1
	}

	if numNonOptional > 0 {
		contents += fmt.Sprintf(tab+"public %s(%s) {\n", s.Name, argsList)
		contents += fmt.Sprintf(tabtab + "this();\n")
		for _, field := range s.Fields {
			if field.Modifier == parser.Optional {
				continue
			}

			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += fmt.Sprintf(tabtab+"this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name, field.Name)
			} else {
				contents += fmt.Sprintf(tabtab+"this.%s = %s;\n", field.Name, field.Name)
			}

			if g.isJavaPrimitive(field.Type) {
				contents += fmt.Sprintf(tabtab+"set%sIsSet(true);\n", strings.Title(field.Name))
			}
		}
		contents += tab + "}\n\n"
	}
	return contents
}

func (g *Generator) generateCopyConstructor(s *parser.Struct) string {
	contents := ""
	contents += g.GenerateBlockComment([]string{"Performs a deep copy on <i>other</i>."}, tab)
	contents += fmt.Sprintf(tab+"public %s(%s other) {\n", s.Name, s.Name)

	isSetType, _ := g.getIsSetType(s)
	switch isSetType {
	case IsSetNone:
		// do nothing
	case IsSetBitfield:
		contents += tabtab + "__isset_bitfield = other.__isset_bitfield;\n"
	case IsSetBitSet:
		contents += tabtab + "__isset_bit_vector.clear();\n"
		contents += tabtab + "__isset_bit_vector.or(other.__isset_bit_vector);\n"
	}

	for _, field := range s.Fields {
		isPrimitive := g.isJavaPrimitive(g.Frugal.UnderlyingType(field.Type))
		ind := tabtab
		if !isPrimitive {
			contents += fmt.Sprintf(tabtab+"if (other.isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
		}
		contents += g.generateCopyConstructorField(field, "other."+field.Name, true, ind)
		if !isPrimitive {
			contents += tabtab + "}\n"
		}
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateDeepCopyMethod(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public %s deepCopy() {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"return new %s(this);\n", s.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateClear(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab + "public void clear() {\n")
	for _, field := range s.Fields {
		underlyingType := g.Frugal.UnderlyingType(field.Type)

		if field.Default != nil {
			val := g.generateConstantValueWrapper("this."+field.Name, field.Type, field.Default, false, false)
			contents += fmt.Sprintf("%s\n", val)
		} else if g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+"set%sIsSet(false);\n", strings.Title(field.Name))
			val := g.getPrimitiveDefaultValue(underlyingType)
			contents += fmt.Sprintf(tabtab+"this.%s = %s;\n\n", field.Name, val)
		} else {
			contents += fmt.Sprintf(tabtab+"this.%s = null;\n\n", field.Name)
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateContainerGetSize(field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public int get%sSize() {\n", strings.Title(field.Name))
	contents += fmt.Sprintf(tabtab+"return (this.%s == null) ? 0 : this.%s.size();\n", field.Name, field.Name)
	contents += fmt.Sprintf(tab + "}\n\n")
	return contents
}

func (g *Generator) generateContainerIterator(field *parser.Field) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	// maps don't get iterators
	if underlyingType.Name == "map" {
		return ""
	}

	contents := ""
	contents += fmt.Sprintf(tab+"public java.util.Iterator<%s> get%sIterator() {\n",
		containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType)), strings.Title(field.Name))
	contents += fmt.Sprintf(tabtab+"return (this.%s == null) ? null : this.%s.iterator();\n", field.Name, field.Name)
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateContainerAddTo(field *parser.Field) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	valType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
	fieldTitle := strings.Title(field.Name)

	contents := ""

	if underlyingType.Name == "list" || underlyingType.Name == "set" {
		contents += fmt.Sprintf(tab+"public void addTo%s(%s elem) {\n", fieldTitle, valType)
		newContainer := ""
		if underlyingType.Name == "list" {
			newContainer = fmt.Sprintf("new ArrayList<%s>()", containerType(valType))
		} else {
			newContainer = fmt.Sprintf("new HashSet<%s>()", containerType(valType))
		}
		contents += fmt.Sprintf(tabtab+"if (this.%s == null) {\n", field.Name)
		contents += fmt.Sprintf(tabtabtab+"this.%s = %s;\n", field.Name, newContainer)
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"this.%s.add(elem);\n", field.Name)
		contents += tab + "}\n\n"
	} else {
		contents += fmt.Sprintf(tab+"public void putTo%s(%s key, %s val) {\n",
			fieldTitle, g.getJavaTypeFromThriftType(underlyingType.KeyType), valType)
		contents += fmt.Sprintf(tabtab+"if (this.%s == null) {\n", field.Name)
		contents += fmt.Sprintf(tabtabtab+"this.%s = new HashMap<%s,%s>();\n",
			field.Name, containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType)), containerType(valType))
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"this.%s.put(key, val);\n", field.Name)
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) getAccessorPrefix(t *parser.Type) string {
	if g.Frugal.UnderlyingType(t).Name == "bool" {
		return "is"
	}
	return "get"
}

func (g *Generator) generateGetField(field *parser.Field) string {
	contents := ""
	fieldTitle := strings.Title(field.Name)
	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, tab)
	}

	underlyingType := g.Frugal.UnderlyingType(field.Type)
	returnType := g.getJavaTypeFromThriftType(underlyingType)
	// There's some weird overlap between the ByteBuffer and byte[] methods
	if underlyingType.Name == "binary" {
		returnType = "byte[]"
	}

	accessPrefix := g.getAccessorPrefix(field.Type)
	contents += fmt.Sprintf(tab+"public %s %s%s() {\n", returnType, accessPrefix, fieldTitle)
	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tabtab+"set%s(org.apache.thrift.TBaseHelper.rightSize(%s));\n",
			strings.Title(field.Name), field.Name)
		contents += fmt.Sprintf(tabtab+"return %s == null ? null : %s.array();\n", field.Name, field.Name)
	} else {
		contents += fmt.Sprintf(tabtab+"return this.%s;\n", field.Name)
	}
	contents += tab + "}\n\n"

	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tab+"public java.nio.ByteBuffer bufferFor%s() {\n", fieldTitle)
		contents += fmt.Sprintf(tabtab+"return org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name)
		contents += fmt.Sprintf(tab + "}\n\n")
	}

	return contents
}

func (g *Generator) generateSetField(structName string, field *parser.Field) string {
	fieldTitle := strings.Title(field.Name)
	underlyingType := g.Frugal.UnderlyingType(field.Type)

	contents := ""

	if underlyingType.Name == "binary" {
		// Special additional binary set
		contents += fmt.Sprintf(tab+"public %s set%s(byte[] %s) {\n", structName, fieldTitle, field.Name)
		contents += fmt.Sprintf(tabtab+"this.%s = %s == null ? (java.nio.ByteBuffer)null : java.nio.ByteBuffer.wrap(Arrays.copyOf(%s, %s.length));\n",
			field.Name, field.Name, field.Name, field.Name)
		contents += tabtab + "return this;\n"
		contents += tab + "}\n\n"
	}

	if field.Comment != nil {
		contents += g.GenerateBlockComment(field.Comment, tab)
	}
	contents += fmt.Sprintf(tab+"public %s set%s(%s %s) {\n",
		structName, fieldTitle, g.getJavaTypeFromThriftType(field.Type), field.Name)

	if underlyingType.Name == "binary" {
		contents += fmt.Sprintf(tabtab+"this.%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", field.Name, field.Name)

	} else {
		contents += fmt.Sprintf(tabtab+"this.%s = %s;\n", field.Name, field.Name)
	}

	if g.isJavaPrimitive(field.Type) {
		contents += fmt.Sprintf(tabtab+"set%sIsSet(true);\n", fieldTitle)
	}

	contents += tabtab + "return this;\n"
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateUnsetField(s *parser.Struct, field *parser.Field) string {
	contents := ""

	contents += fmt.Sprintf(tab+"public void unset%s() {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"__isset_bitfield = EncodingUtils.clearBit(__isset_bitfield, %s);\n", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"__isset_bit_vector.clear(%s);\n", isSetID)
		}
	} else {
		contents += fmt.Sprintf(tabtab+"this.%s = null;\n", field.Name)
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) getIsSetID(fieldName string) string {
	return fmt.Sprintf("__%s_ISSET_ID", strings.ToUpper(fieldName))
}

func (g *Generator) generateIsSetField(s *parser.Struct, field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"/** Returns true if field %s is set (has been assigned a value) and false otherwise */\n", field.Name)
	contents += fmt.Sprintf(tab+"public boolean isSet%s() {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"return EncodingUtils.testBit(__isset_bitfield, %s);\n", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"return __isset_bit_vector.get(%s);\n", isSetID)
		}
	} else {
		contents += fmt.Sprintf(tabtab+"return this.%s != null;\n", field.Name)
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateSetIsSetField(s *parser.Struct, field *parser.Field) string {
	contents := ""
	contents += fmt.Sprintf(tab+"public void set%sIsSet(boolean value) {\n", strings.Title(field.Name))
	if g.isJavaPrimitive(field.Type) {
		isSetType, _ := g.getIsSetType(s)
		isSetID := g.getIsSetID(field.Name)
		switch isSetType {
		case IsSetNone:
			panic("IsSetNone occurred with a primitive")
		case IsSetBitfield:
			contents += fmt.Sprintf(tabtab+"__isset_bitfield = EncodingUtils.setBit(__isset_bitfield, %s, value);\n", isSetID)
		case IsSetBitSet:
			contents += fmt.Sprintf(tabtab+"__isset_bit_vector.set(%s, value);\n", isSetID)
		}
	} else {
		contents += tabtab + "if (!value) {\n"
		contents += fmt.Sprintf(tabtabtab+"this.%s = null;\n", field.Name)
		contents += tabtab + "}\n"
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateSetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void setFieldValue(_Fields field, Object value) {\n"
	contents += tabtab + "switch (field) {\n"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += tabtabtab + "if (value == null) {\n"
		contents += fmt.Sprintf(tabtabtabtab+"unset%s();\n", fieldTitle)
		contents += tabtabtab + "} else {\n"
		contents += fmt.Sprintf(tabtabtabtab+"set%s((%s)value);\n", fieldTitle, containerType(g.getJavaTypeFromThriftType(field.Type)))
		contents += tabtabtab + "}\n"
		contents += tabtabtab + "break;\n\n"
	}
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateGetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "public Object getFieldValue(_Fields field) {\n"
	contents += tabtab + "switch (field) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtab+"return %s%s();\n\n",
			g.getAccessorPrefix(field.Type), strings.Title(field.Name))
	}
	contents += tabtab + "}\n"
	contents += tabtab + "throw new IllegalStateException();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateIsSetValue(s *parser.Struct) string {
	contents := ""
	contents += tab + "/** Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise */\n"
	contents += tab + "public boolean isSet(_Fields field) {\n"
	contents += tabtab + "if (field == null) {\n"
	contents += tabtabtab + "throw new IllegalArgumentException();\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "switch (field) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtab+"case %s:\n", toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtab+"return isSet%s();\n", strings.Title(field.Name))
	}
	contents += tabtab + "}\n"
	contents += tabtab + "throw new IllegalStateException();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateEquals(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public boolean equals(Object that) {\n"
	contents += tabtab + "if (that == null)\n"
	contents += tabtabtab + "return false;\n"
	contents += fmt.Sprintf(tabtab+"if (that instanceof %s)\n", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return this.equals((%s)that);\n", s.Name)
	contents += tabtab + "return false;\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"public boolean equals(%s that) {\n", s.Name)
	contents += tabtab + "if (that == null)\n"
	contents += tabtabtab + "return false;\n\n"

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

		contents += fmt.Sprintf(tabtab+"boolean this_present_%s = %s;\n", field.Name, thisPresentArg)
		contents += fmt.Sprintf(tabtab+"boolean that_present_%s = %s;\n", field.Name, thatPresentArg)
		contents += fmt.Sprintf(tabtab+"if (this_present_%s || that_present_%s) {\n", field.Name, field.Name)
		contents += fmt.Sprintf(tabtabtab+"if (!(this_present_%s && that_present_%s))\n", field.Name, field.Name)
		contents += tabtabtabtab + "return false;\n"

		unequalTest := ""
		if primitive {
			unequalTest = fmt.Sprintf("this.%s != that.%s", field.Name, field.Name)
		} else {
			unequalTest = fmt.Sprintf("!this.%s.equals(that.%s)", field.Name, field.Name)
		}
		contents += fmt.Sprintf(tabtabtab+"if (%s)\n", unequalTest)
		contents += tabtabtabtab + "return false;\n"
		contents += tabtab + "}\n\n"
	}

	contents += tabtab + "return true;\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateHashCode(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public int hashCode() {\n"
	contents += tabtab + "List<Object> list = new ArrayList<Object>();\n\n"
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		primitive := g.isJavaPrimitive(field.Type)

		presentArg := "true"
		if optional || !primitive {
			presentArg += fmt.Sprintf(" && (isSet%s())", strings.Title(field.Name))
		}

		contents += fmt.Sprintf(tabtab+"boolean present_%s = %s;\n", field.Name, presentArg)
		contents += fmt.Sprintf(tabtab+"list.add(present_%s);\n", field.Name)
		contents += fmt.Sprintf(tabtab+"if (present_%s)\n", field.Name)
		if g.Frugal.IsEnum(field.Type) {
			contents += fmt.Sprintf(tabtabtab+"list.add(%s.getValue());\n\n", field.Name)
		} else {
			contents += fmt.Sprintf(tabtabtab+"list.add(%s);\n\n", field.Name)
		}
	}
	contents += tabtab + "return list.hashCode();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateCompareTo(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += fmt.Sprintf(tab+"public int compareTo(%s other) {\n", s.Name)
	contents += tabtab + "if (!getClass().equals(other.getClass())) {\n"
	contents += tabtabtab + "return getClass().getName().compareTo(other.getClass().getName());\n"
	contents += tabtab + "}\n\n"
	contents += tabtab + "int lastComparison = 0;\n\n"
	for _, field := range s.Fields {
		fieldTitle := strings.Title(field.Name)
		contents += fmt.Sprintf(tabtab+"lastComparison = Boolean.valueOf(isSet%s()).compareTo(other.isSet%s());\n", fieldTitle, fieldTitle)
		contents += tabtab + "if (lastComparison != 0) {\n"
		contents += tabtabtab + "return lastComparison;\n"
		contents += tabtab + "}\n"
		contents += fmt.Sprintf(tabtab+"if (isSet%s()) {\n", fieldTitle)
		contents += fmt.Sprintf(tabtabtab+"lastComparison = org.apache.thrift.TBaseHelper.compareTo(this.%s, other.%s);\n", field.Name, field.Name)
		contents += tabtabtab + "if (lastComparison != 0) {\n"
		contents += tabtabtabtab + "return lastComparison;\n"
		contents += tabtabtab + "}\n"
		contents += tabtab + "}\n"
	}
	contents += tabtab + "return 0;\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateFieldForId(s *parser.Struct) string {
	contents := ""
	contents += tab + "public _Fields fieldForId(int fieldId) {\n"
	contents += tabtab + "return _Fields.findByThriftId(fieldId);\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateReadWrite(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void read(org.apache.thrift.protocol.TProtocol iprot) throws org.apache.thrift.TException {\n"
	contents += tabtab + "schemes.get(iprot.getScheme()).getScheme().read(iprot, this);\n"
	contents += tab + "}\n\n"

	contents += tab + "public void write(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {\n"
	contents += tabtab + "schemes.get(oprot.getScheme()).getScheme().write(oprot, this);\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateToString(s *parser.Struct) string {
	contents := ""
	contents += tab + "@Override\n"
	contents += tab + "public String toString() {\n"
	contents += fmt.Sprintf(tabtab+"StringBuilder sb = new StringBuilder(\"%s(\");\n", s.Name)
	contents += tabtab + "boolean first = true;\n\n"
	first := true
	for _, field := range s.Fields {
		optional := field.Modifier == parser.Optional
		ind := ""
		if optional {
			contents += fmt.Sprintf(tabtab+"if (isSet%s()) {\n", strings.Title(field.Name))
			ind = tab
		}

		if !first {
			contents += tabtab + ind + "if (!first) sb.append(\", \");\n"
		}
		contents += fmt.Sprintf(tabtab+ind+"sb.append(\"%s:\");\n", field.Name)
		if !g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+ind+"if (this.%s == null) {\n", field.Name)
			contents += tabtabtab + ind + "sb.append(\"null\");\n"
			contents += tabtab + ind + "} else {\n"
			if g.Frugal.UnderlyingType(field.Type).Name == "binary" {
				contents += fmt.Sprintf(tabtabtab+ind+"org.apache.thrift.TBaseHelper.toString(this.%s, sb);\n", field.Name)
			} else {
				contents += fmt.Sprintf(tabtabtab+ind+"sb.append(this.%s);\n", field.Name)
			}
			contents += tabtab + ind + "}\n"
		} else {
			contents += fmt.Sprintf(tabtab+ind+"sb.append(this.%s);\n", field.Name)
		}
		contents += tabtab + ind + "first = false;\n"

		if optional {
			contents += tabtab + "}\n"
		}
		first = false
	}

	contents += tabtab + "sb.append(\")\");\n"
	contents += tabtab + "return sb.toString();\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateValidate(s *parser.Struct) string {
	contents := ""
	contents += tab + "public void validate() throws org.apache.thrift.TException {\n"
	contents += tabtab + "// check for required fields\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && !g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtab+"if (%s == null) {\n", field.Name)
			contents += fmt.Sprintf(tabtabtab+"throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' was not present! Struct: \" + toString());\n",
				field.Name)
			contents += tabtab + "}\n"
		}
	}

	contents += tabtab + "// check for sub-struct validity\n"
	for _, field := range s.Fields {
		if g.Frugal.IsStruct(field.Type) && !g.Frugal.IsUnion(field.Type) {
			contents += fmt.Sprintf(tabtab+"if (%s != null) {\n", field.Name)
			contents += fmt.Sprintf(tabtabtab+"%s.validate();\n", field.Name)
			contents += tabtab + "}\n"
		}
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateWriteObject(s *parser.Struct) string {
	contents := ""
	contents += tab + "private void writeObject(java.io.ObjectOutputStream out) throws java.io.IOException {\n"
	contents += tabtab + "try {\n"
	contents += tabtabtab + "write(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(out)));\n"
	contents += tabtab + "} catch (org.apache.thrift.TException te) {\n"
	contents += tabtabtab + "throw new java.io.IOException(te);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateReadObject(s *parser.Struct) string {
	contents := ""
	contents += tab + "private void readObject(java.io.ObjectInputStream in) throws java.io.IOException, ClassNotFoundException {\n"
	// isset stuff, don't do for unions
	contents += tabtab + "try {\n"
	if s.Type != parser.StructTypeUnion {
		contents += tabtabtab + "// it doesn't seem like you should have to do this, but java serialization is wacky, and doesn't call the default constructor.\n"
		isSetType, _ := g.getIsSetType(s)
		switch isSetType {
		case IsSetNone:
		// Do nothing
		case IsSetBitfield:
			contents += tabtabtab + "__isset_bitfield = 0;\n"
		case IsSetBitSet:
			contents += tabtabtab + "__isset_bit_vector = new BitSet(1);\n"
		}
	}

	contents += tabtabtab + "read(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(in)));\n"
	contents += tabtab + "} catch (org.apache.thrift.TException te) {\n"
	contents += tabtabtab + "throw new java.io.IOException(te);\n"
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateStandardScheme(s *parser.Struct, isResult bool) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static class %sStandardSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"public %sStandardScheme getScheme() {\n", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return new %sStandardScheme();\n", s.Name)
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"private static class %sStandardScheme extends StandardScheme<%s> {\n\n", s.Name, s.Name)

	// read
	contents += fmt.Sprintf(tabtab+"public void read(org.apache.thrift.protocol.TProtocol iprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "org.apache.thrift.protocol.TField schemeField;\n"
	contents += tabtabtab + "iprot.readStructBegin();\n"
	contents += tabtabtab + "while (true) {\n"
	contents += tabtabtabtab + "schemeField = iprot.readFieldBegin();\n"
	contents += tabtabtabtab + "if (schemeField.type == org.apache.thrift.protocol.TType.STOP) {\n"
	contents += tabtabtabtabtab + "break;\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "switch (schemeField.id) {\n"
	for _, field := range s.Fields {
		contents += fmt.Sprintf(tabtabtabtabtab+"case %d: // %s\n", field.ID, toConstantName(field.Name))
		contents += fmt.Sprintf(tabtabtabtabtabtab+"if (schemeField.type == %s) {\n", g.getTType(field.Type))
		contents += g.generateReadFieldRec(field, true, false, false, tabtabtabtabtabtabtab)
		contents += fmt.Sprintf(tabtabtabtabtabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
		contents += tabtabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
		contents += tabtabtabtabtabtab + "}\n"
		contents += tabtabtabtabtabtab + "break;\n"
	}
	contents += tabtabtabtabtab + "default:\n"
	contents += tabtabtabtabtabtab + "org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "iprot.readFieldEnd();\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "iprot.readStructEnd();\n\n"

	contents += tabtabtab + "// check for required fields of primitive type, which can't be checked in the validate method\n"
	for _, field := range s.Fields {
		if field.Modifier == parser.Required && g.isJavaPrimitive(field.Type) {
			contents += fmt.Sprintf(tabtabtab+"if (!struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += fmt.Sprintf(tabtabtabtab+"throw new org.apache.thrift.protocol.TProtocolException(\"Required field '%s' was not found in serialized data! Struct: \" + toString());\n", field.Name)
			contents += tabtabtab + "}\n"
		}
	}

	contents += tabtabtab + "struct.validate();\n"
	contents += tabtab + "}\n\n"

	// write
	contents += fmt.Sprintf(tabtab+"public void write(org.apache.thrift.protocol.TProtocol oprot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "struct.validate();\n\n"
	contents += tabtabtab + "oprot.writeStructBegin(STRUCT_DESC);\n"
	for _, field := range s.Fields {
		isKindOfPrimitive := g.canBeJavaPrimitive(field.Type)
		ind := tabtabtab
		optInd := tabtabtab
		if !isKindOfPrimitive {
			contents += fmt.Sprintf(ind+"if (struct.%s != null) {\n", field.Name)
			ind += tab
			optInd += tab
		}
		opt := field.Modifier == parser.Optional || (isResult && isKindOfPrimitive)
		if opt {
			contents += fmt.Sprintf(ind+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			ind += tab
		}

		contents += fmt.Sprintf(ind+"oprot.writeFieldBegin(%s_FIELD_DESC);\n", toConstantName(field.Name))
		contents += g.generateWriteFieldRec(field, true, false, ind)
		contents += ind + "oprot.writeFieldEnd();\n"

		if opt {
			contents += optInd + "}\n"
		}
		if !isKindOfPrimitive {
			contents += tabtabtab + "}\n"
		}
	}
	contents += tabtabtab + "oprot.writeFieldStop();\n"
	contents += tabtabtab + "oprot.writeStructEnd();\n"

	contents += tabtab + "}\n\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateTupleScheme(s *parser.Struct) string {
	contents := ""
	contents += fmt.Sprintf(tab+"private static class %sTupleSchemeFactory implements SchemeFactory {\n", s.Name)
	contents += fmt.Sprintf(tabtab+"public %sTupleScheme getScheme() {\n", s.Name)
	contents += fmt.Sprintf(tabtabtab+"return new %sTupleScheme();\n", s.Name)
	contents += tabtab + "}\n"
	contents += tab + "}\n\n"

	contents += fmt.Sprintf(tab+"private static class %sTupleScheme extends TupleScheme<%s> {\n\n", s.Name, s.Name)
	contents += tabtab + "@Override\n"
	contents += fmt.Sprintf(tabtab+"public void write(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "TTupleProtocol oprot = (TTupleProtocol) prot;\n"
	// write required fields
	numNonReqs := 0
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			numNonReqs += 1
			continue
		}

		contents += g.generateWriteFieldRec(field, true, true, tabtabtab)
	}

	if numNonReqs > 0 {
		// write optional/default fields
		nonReqFieldCount := 0
		contents += tabtabtab + "BitSet optionals = new BitSet();\n"
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += fmt.Sprintf(tabtabtabtab+"optionals.set(%d);\n", nonReqFieldCount)
			contents += tabtabtab + "}\n"
			nonReqFieldCount += 1
		}

		contents += fmt.Sprintf(tabtabtab+"oprot.writeBitSet(optionals, %d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (struct.isSet%s()) {\n", strings.Title(field.Name))
			contents += g.generateWriteFieldRec(field, true, true, tabtabtabtab)
			contents += tabtabtab + "}\n"
		}
	}

	contents += tabtab + "}\n\n"

	contents += tabtab + "@Override\n"
	contents += fmt.Sprintf(tabtab+"public void read(org.apache.thrift.protocol.TProtocol prot, %s struct) throws org.apache.thrift.TException {\n", s.Name)
	contents += tabtabtab + "TTupleProtocol iprot = (TTupleProtocol) prot;\n"
	// read required fields
	for _, field := range s.Fields {
		if field.Modifier != parser.Required {
			continue
		}

		contents += g.generateReadFieldRec(field, true, true, false, tabtabtab)
		contents += fmt.Sprintf(tabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
	}

	if numNonReqs > 0 {
		// read default/optional fields
		nonReqFieldCount := 0
		contents += fmt.Sprintf(tabtabtab+"BitSet incoming = iprot.readBitSet(%d);\n", numNonReqs)
		for _, field := range s.Fields {
			if field.Modifier == parser.Required {
				continue
			}

			contents += fmt.Sprintf(tabtabtab+"if (incoming.get(%d)) {\n", nonReqFieldCount)
			contents += g.generateReadFieldRec(field, true, true, false, tabtabtabtab)
			contents += fmt.Sprintf(tabtabtabtab+"struct.set%sIsSet(true);\n", strings.Title(field.Name))
			contents += tabtabtab + "}\n"
			nonReqFieldCount += 1
		}
	}
	contents += tabtab + "}\n\n"

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateCopyConstructorField(field *parser.Field, otherFieldName string, first bool, ind string) string {
	underlyingType := g.Frugal.UnderlyingType(field.Type)
	isPrimitive := g.canBeJavaPrimitive(underlyingType)
	accessPrefix := "this."
	declPrefix := "this."
	if !first {
		accessPrefix = ""
		declPrefix = g.getJavaTypeFromThriftType(underlyingType) + " "
	}

	if isPrimitive || underlyingType.Name == "string" {
		return fmt.Sprintf(ind+"%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.Name == "binary" {
		return fmt.Sprintf(ind+"%s%s = org.apache.thrift.TBaseHelper.copyBinary(%s);\n", declPrefix, field.Name, otherFieldName)
	} else if g.Frugal.IsStruct(underlyingType) {
		return fmt.Sprintf(ind+"%s%s = new %s(%s);\n", declPrefix, field.Name, g.getJavaTypeFromThriftType(underlyingType), otherFieldName)
	} else if g.Frugal.IsEnum(underlyingType) {
		return fmt.Sprintf(ind+"%s%s = %s;\n", declPrefix, field.Name, otherFieldName)
	} else if underlyingType.IsContainer() {
		contents := ""
		valueType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		containerValType := containerType(valueType)
		otherValElem := g.GetElem()
		thisValElem := g.GetElem()
		thisValField := parser.FieldFromType(underlyingType.ValueType, thisValElem)

		switch underlyingType.Name {
		case "list":
			contents += fmt.Sprintf(ind+"%s%s = new ArrayList<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += fmt.Sprintf(ind+"for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += ind + "}\n"
		case "set":
			contents += fmt.Sprintf(ind+"%s%s = new HashSet<%s>(%s.size());\n", declPrefix, field.Name, containerValType, otherFieldName)
			contents += fmt.Sprintf(ind+"for (%s %s : %s) {\n", valueType, otherValElem, otherFieldName)
			contents += g.generateCopyConstructorField(thisValField, otherValElem, false, ind+tab)
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, thisValElem)
			contents += ind + "}\n"
		case "map":
			keyType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyUnderlying := g.Frugal.UnderlyingType(underlyingType.KeyType)
			valUnderlying := g.Frugal.UnderlyingType(underlyingType.ValueType)
			containerKeyType := containerType(keyType)

			// If it's all primitives, optimization. Otherwise need to iterate
			if (g.isJavaPrimitive(keyUnderlying) || keyUnderlying.Name == "string") &&
				(g.isJavaPrimitive(valUnderlying) || valUnderlying.Name == "string") {
				contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(%s);\n",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
			} else {
				thisKeyElem := g.GetElem()
				thisKeyField := parser.FieldFromType(underlyingType.KeyType, thisKeyElem)

				contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(%s.size());\n",
					declPrefix, field.Name, containerKeyType, containerValType, otherFieldName)
				contents += fmt.Sprintf(ind+"for (Map.Entry<%s, %s> %s : %s.entrySet()) {\n",
					containerKeyType, containerValType, otherValElem, otherFieldName)
				contents += g.generateCopyConstructorField(thisKeyField, otherValElem+".getKey()", false, ind+tab)
				contents += g.generateCopyConstructorField(thisValField, otherValElem+".getValue()", false, ind+tab)
				contents += fmt.Sprintf(ind+tab+"%s%s.put(%s, %s);\n", accessPrefix, field.Name, thisKeyElem, thisValElem)
				contents += ind + "}\n"
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
func (g *Generator) generateReadFieldRec(field *parser.Field, first bool, succinct bool, containerTypes bool, ind string) string {
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

		contents += fmt.Sprintf(ind+"%s%s = iprot.read%s();\n", declPrefix, field.Name, thriftType)
	} else if g.Frugal.IsEnum(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s = %s.findByValue(iprot.readI32());\n", declPrefix, field.Name, javaType)
	} else if g.Frugal.IsStruct(underlyingType) {
		contents += fmt.Sprintf(ind+"%s%s = new %s();\n", declPrefix, field.Name, javaType)
		contents += fmt.Sprintf(ind+"%s%s.read(iprot);\n", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		containerElem := g.GetElem()
		counterElem := g.GetElem()

		valType := containerType(g.getJavaTypeFromThriftType(underlyingType.ValueType))
		valElem := g.GetElem()
		valField := parser.FieldFromType(underlyingType.ValueType, valElem)
		valContents := g.generateReadFieldRec(valField, false, succinct, containerTypes, ind+tab)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TList %s = new org.apache.thrift.protocol.TList(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TList %s = iprot.readListBegin();\n", containerElem)
			}
			contents += fmt.Sprintf(ind+"%s%s = new ArrayList<%s>(%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readListEnd();\n"
			}
		case "set":
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TSet %s = new org.apache.thrift.protocol.TSet(%s, iprot.readI32());\n",
					containerElem, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TSet %s = iprot.readSetBegin();\n", containerElem)
			}
			contents += fmt.Sprintf(ind+"%s%s = new HashSet<%s>(2*%s.size);\n", declPrefix, field.Name, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.add(%s);\n", accessPrefix, field.Name, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readSetEnd();\n"
			}
		case "map":
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TMap %s = new org.apache.thrift.protocol.TMap(%s, %s, iprot.readI32());\n",
					containerElem, keyTType, valTType)
			} else {
				contents += fmt.Sprintf(ind+"org.apache.thrift.protocol.TMap %s = iprot.readMapBegin();\n", containerElem)
			}

			keyType := containerType(g.getJavaTypeFromThriftType(underlyingType.KeyType))
			contents += fmt.Sprintf(ind+"%s%s = new HashMap<%s,%s>(2*%s.size);\n", declPrefix, field.Name, keyType, valType, containerElem)
			contents += fmt.Sprintf(ind+"for (int %s = 0; %s < %s.size; ++%s) {\n", counterElem, counterElem, containerElem, counterElem)
			keyElem := g.GetElem()
			keyField := parser.FieldFromType(underlyingType.KeyType, keyElem)
			contents += g.generateReadFieldRec(keyField, false, succinct, containerTypes, ind+tab)
			contents += valContents
			contents += fmt.Sprintf(ind+tab+"%s%s.put(%s, %s);\n", accessPrefix, field.Name, keyElem, valElem)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "iprot.readMapEnd();\n"
			}
		}
	}

	return contents
}

// succinct means only write collection length instead of the whole header,
// and don't write collection end.
func (g *Generator) generateWriteFieldRec(field *parser.Field, first bool, succinct bool, ind string) string {
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
		contents += fmt.Sprintf(ind+"%s %s = %s%s;\n",
			g.getJavaTypeFromThriftType(underlyingType), elem, accessPrefix, field.Name)
		if g.canBeJavaPrimitive(underlyingType) && g.generateBoxedPrimitives() {
			contents += fmt.Sprintf(ind+"if (%s == null) {\n", elem)
			val := g.getPrimitiveDefaultValue(underlyingType)
			contents += fmt.Sprintf(ind+tab+"%s = %s;\n", elem, val)
			contents += fmt.Sprintf(ind + "}\n")
		}

		write := ind + "oprot.write"
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
		contents += fmt.Sprintf(ind+"%s%s.write(oprot);\n", accessPrefix, field.Name)
	} else if underlyingType.IsContainer() {
		iterElem := g.GetElem()
		valJavaType := g.getJavaTypeFromThriftType(underlyingType.ValueType)
		valTType := g.getTType(underlyingType.ValueType)

		switch underlyingType.Name {
		case "list":
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeListBegin(new org.apache.thrift.protocol.TList(%s, %s%s.size()));\n",
					valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (%s %s : %s%s) {\n", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeListEnd();\n"
			}
		case "set":
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeSetBegin(new org.apache.thrift.protocol.TSet(%s, %s%s.size()));\n",
					valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (%s %s : %s%s) {\n", valJavaType, iterElem, accessPrefix, field.Name)
			iterField := parser.FieldFromType(underlyingType.ValueType, iterElem)
			contents += g.generateWriteFieldRec(iterField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeSetEnd();\n"
			}
		case "map":
			keyJavaType := g.getJavaTypeFromThriftType(underlyingType.KeyType)
			keyTType := g.getTType(underlyingType.KeyType)
			if succinct {
				contents += fmt.Sprintf(ind+"oprot.writeI32(%s%s.size());\n", accessPrefix, field.Name)
			} else {
				contents += fmt.Sprintf(ind+"oprot.writeMapBegin(new org.apache.thrift.protocol.TMap(%s, %s, %s%s.size()));\n",
					keyTType, valTType, accessPrefix, field.Name)
			}
			contents += fmt.Sprintf(ind+"for (Map.Entry<%s, %s> %s : %s%s.entrySet()) {\n",
				containerType(keyJavaType), containerType(valJavaType), iterElem, accessPrefix, field.Name)
			keyField := parser.FieldFromType(underlyingType.KeyType, iterElem+".getKey()")
			valField := parser.FieldFromType(underlyingType.ValueType, iterElem+".getValue()")
			contents += g.generateWriteFieldRec(keyField, false, succinct, ind+tab)
			contents += g.generateWriteFieldRec(valField, false, succinct, ind+tab)
			contents += ind + "}\n"
			if !succinct {
				contents += ind + "oprot.writeMapEnd();\n"
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
	imports += "import javax.annotation.Generated;\n"
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

	imports += "import javax.annotation.Generated;\n"
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
	imports += "import javax.annotation.Generated;\n"

	_, err := file.WriteString(imports)
	return err
}

func (g *Generator) GenerateConstants(file *os.File, name string) error {
	return nil
}

func (g *Generator) GeneratePublisher(file *os.File, scope *parser.Scope) error {
	scopeTitle := strings.Title(scope.Name)
	publisher := ""

	if g.includeGeneratedAnnotation() {
		publisher += g.generatedAnnotation()
	}
	publisher += fmt.Sprintf("public class %sPublisher {\n\n", scopeTitle)

	publisher += g.generatePublisherIface(scope)
	publisher += g.generatePublisherClient(scope)

	publisher += "}"

	_, err := file.WriteString(publisher)
	return err
}

func (g *Generator) generatePublisherIface(scope *parser.Scope) string {
	contents := ""

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, tab)
	}
	contents += tab + "public interface Iface {\n"

	contents += tabtab + "public void open() throws TException;\n\n"
	contents += tabtab + "public void close() throws TException;\n\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, tabtab)
		}
		contents += fmt.Sprintf(tabtab+"public void publish%s(FContext ctx, %s%s req) throws TException;\n\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generatePublisherClient(scope *parser.Scope) string {
	publisher := ""

	scopeTitle := strings.Title(scope.Name)

	if scope.Comment != nil {
		publisher += g.GenerateBlockComment(scope.Comment, tab)
	}
	publisher += tab + "public static class Client implements Iface {\n"
	publisher += fmt.Sprintf(tabtab+"private static final String DELIMITER = \"%s\";\n\n", globals.TopicDelimiter)
	publisher += tabtab + "private final Iface target;\n"
	publisher += tabtab + "private final Iface proxy;\n\n"

	publisher += tabtab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {\n"
	publisher += fmt.Sprintf(tabtabtab+"target = new Internal%sPublisher(provider);\n", scopeTitle)
	publisher += tabtabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	publisher += tabtabtab + "combined.addAll(provider.getMiddleware());\n"
	publisher += tabtabtab + "middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	publisher += tabtabtab + "proxy = InvocationHandler.composeMiddleware(target, Iface.class, middleware);\n"
	publisher += tabtab + "}\n\n"

	publisher += tabtab + "public void open() throws TException {\n"
	publisher += tabtabtab + "target.open();\n"
	publisher += tabtab + "}\n\n"

	publisher += tabtab + "public void close() throws TException {\n"
	publisher += tabtabtab + "target.close();\n"
	publisher += tabtab + "}\n\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			publisher += g.GenerateBlockComment(op.Comment, tabtab)
		}
		publisher += fmt.Sprintf(tabtab+"public void publish%s(FContext ctx, %s%s req) throws TException {\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))
		publisher += fmt.Sprintf(tabtabtab+"proxy.publish%s(%s);\n", op.Name, g.generateScopeArgs(scope))
		publisher += tabtab + "}\n\n"
	}

	publisher += fmt.Sprintf(tabtab+"protected static class Internal%sPublisher implements Iface {\n\n", scopeTitle)

	publisher += tabtabtab + "private FScopeProvider provider;\n"
	publisher += tabtabtab + "private FPublisherTransport transport;\n"
	publisher += tabtabtab + "private FProtocolFactory protocolFactory;\n\n"

	publisher += fmt.Sprintf(tabtabtab+"protected Internal%sPublisher() {\n", scopeTitle)
	publisher += tabtabtab + "}\n\n"

	publisher += fmt.Sprintf(tabtabtab+"public Internal%sPublisher(FScopeProvider provider) {\n", scopeTitle)
	publisher += tabtabtabtab + "this.provider = provider;\n"
	publisher += tabtabtab + "}\n\n"

	publisher += tabtabtab + "public void open() throws TException {\n"
	publisher += tabtabtabtab + "FScopeProvider.Publisher publisher = provider.buildPublisher();\n"
	publisher += tabtabtabtab + "transport = publisher.getTransport();\n"
	publisher += tabtabtabtab + "protocolFactory = publisher.getProtocolFactory();\n"
	publisher += tabtabtabtab + "transport.open();\n"
	publisher += tabtabtab + "}\n\n"

	publisher += tabtabtab + "public void close() throws TException {\n"
	publisher += tabtabtabtab + "transport.close();\n"
	publisher += tabtabtab + "}\n\n"

	prefix := ""
	for _, op := range scope.Operations {
		publisher += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			publisher += g.GenerateBlockComment(op.Comment, tabtabtab)
		}
		publisher += fmt.Sprintf(tabtabtab+"public void publish%s(FContext ctx, %s%s req) throws TException {\n", op.Name, args, g.getJavaTypeFromThriftType(op.Type))

		// Inject the prefix variables into the FContext to send
		for _, prefixVar := range scope.Prefix.Variables {
			publisher += fmt.Sprintf(tabtabtabtab+"ctx.addRequestHeader(\"_topic_%s\", %s);\n", prefixVar, prefixVar)
		}

		publisher += tabtabtabtab + fmt.Sprintf("String op = \"%s\";\n", op.Name)
		publisher += tabtabtabtab + fmt.Sprintf("String prefix = %s;\n", generatePrefixStringTemplate(scope))
		publisher += tabtabtabtab + "String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);\n"
		publisher += tabtabtabtab + "TMemoryOutputBuffer memoryBuffer = new TMemoryOutputBuffer(transport.getPublishSizeLimit());\n"
		publisher += tabtabtabtab + "FProtocol oprot = protocolFactory.getProtocol(memoryBuffer);\n"
		publisher += tabtabtabtab + "oprot.writeRequestHeader(ctx);\n"
		publisher += tabtabtabtab + "oprot.writeMessageBegin(new TMessage(op, TMessageType.CALL, 0));\n"
		publisher += g.generateWriteFieldRec(parser.FieldFromType(op.Type, "req"), false, false, tabtabtabtab)
		publisher += tabtabtabtab + "oprot.writeMessageEnd();\n"
		publisher += tabtabtabtab + "transport.publish(topic, memoryBuffer.getWriteBytes());\n"
		publisher += tabtabtab + "}\n"
	}

	publisher += tabtab + "}\n"
	publisher += tab + "}\n"

	return publisher
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
	subscriber := ""
	scopeName := strings.Title(scope.Name)
	if g.includeGeneratedAnnotation() {
		subscriber += g.generatedAnnotation()
	}
	subscriber += fmt.Sprintf("public class %sSubscriber {\n\n", scopeName)

	subscriber += g.generateSubscriberIface(scope)
	subscriber += g.generateHandlerIfaces(scope)
	subscriber += g.generateSubscriberClient(scope)

	subscriber += "\n}"

	_, err := file.WriteString(subscriber)
	return err
}

func (g *Generator) generateSubscriberIface(scope *parser.Scope) string {
	contents := ""

	if scope.Comment != nil {
		contents += g.GenerateBlockComment(scope.Comment, tab)
	}
	contents += tab + "public interface Iface {\n"

	args := g.generateScopePrefixArgs(scope)

	for _, op := range scope.Operations {
		if op.Comment != nil {
			contents += g.GenerateBlockComment(op.Comment, tabtab)
		}
		contents += fmt.Sprintf(tabtab+"public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException;\n\n",
			op.Name, args, op.Name)
	}

	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) generateHandlerIfaces(scope *parser.Scope) string {
	contents := ""

	for _, op := range scope.Operations {
		contents += fmt.Sprintf(tab+"public interface %sHandler {\n", op.Name)
		contents += fmt.Sprintf(tabtab+"void on%s(FContext ctx, %s req);\n", op.Name, g.getJavaTypeFromThriftType(op.Type))
		contents += tab + "}\n\n"
	}

	return contents
}

func (g *Generator) generateSubscriberClient(scope *parser.Scope) string {
	subscriber := ""

	prefix := ""
	args := g.generateScopePrefixArgs(scope)

	if scope.Comment != nil {
		subscriber += g.GenerateBlockComment(scope.Comment, tab)
	}
	subscriber += tab + "public static class Client implements Iface {\n"

	subscriber += fmt.Sprintf(tabtab+"private static final String DELIMITER = \"%s\";\n", globals.TopicDelimiter)
	subscriber += tabtab + "private static final Logger LOGGER = LoggerFactory.getLogger(Client.class);\n\n"

	subscriber += tabtab + "private final FScopeProvider provider;\n"
	subscriber += tabtab + "private final ServiceMiddleware[] middleware;\n\n"

	subscriber += tabtab + "public Client(FScopeProvider provider, ServiceMiddleware... middleware) {\n"
	subscriber += tabtabtab + "this.provider = provider;\n"
	subscriber += tabtabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	subscriber += tabtabtab + "combined.addAll(provider.getMiddleware());\n"
	subscriber += tabtabtab + "this.middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	subscriber += tabtab + "}\n\n"

	for _, op := range scope.Operations {
		subscriber += prefix
		prefix = "\n\n"
		if op.Comment != nil {
			subscriber += g.GenerateBlockComment(op.Comment, tabtab)
		}
		subscriber += tabtab + fmt.Sprintf("public FSubscription subscribe%s(%sfinal %sHandler handler) throws TException {\n", op.Name, args, op.Name)
		subscriber += tabtabtab + fmt.Sprintf("final String op = \"%s\";\n", op.Name)
		subscriber += tabtabtab + fmt.Sprintf("String prefix = %s;\n", generatePrefixStringTemplate(scope))
		subscriber += tabtabtab + "final String topic = String.format(\"%s" + strings.Title(scope.Name) + "%s%s\", prefix, DELIMITER, op);\n"
		subscriber += tabtabtab + "final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();\n"
		subscriber += tabtabtab + "final FSubscriberTransport transport = subscriber.getTransport();\n"
		subscriber += tabtabtab + fmt.Sprintf(
			"final %sHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, %sHandler.class, middleware);\n",
			op.Name, op.Name)

		subscriber += tabtabtab + fmt.Sprintf("transport.subscribe(topic, recv%s(op, subscriber.getProtocolFactory(), proxiedHandler));\n", op.Name)
		subscriber += tabtabtab + "return FSubscription.of(topic, transport);\n"
		subscriber += tabtab + "}\n\n"

		subscriber += tabtab + fmt.Sprintf("private FAsyncCallback recv%s(String op, FProtocolFactory pf, %sHandler handler) {\n", op.Name, op.Name)
		subscriber += tabtabtab + "return new FAsyncCallback() {\n"
		subscriber += tabtabtabtab + "public void onMessage(TTransport tr) throws TException {\n"
		subscriber += tabtabtabtabtab + "FProtocol iprot = pf.getProtocol(tr);\n"
		subscriber += tabtabtabtabtab + "FContext ctx = iprot.readRequestHeader();\n"
		subscriber += tabtabtabtabtab + "TMessage msg = iprot.readMessageBegin();\n"
		subscriber += tabtabtabtabtab + "if (!msg.name.equals(op)) {\n"
		subscriber += tabtabtabtabtabtab + "TProtocolUtil.skip(iprot, TType.STRUCT);\n"
		subscriber += tabtabtabtabtabtab + "iprot.readMessageEnd();\n"
		subscriber += tabtabtabtabtabtab + "throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);\n"
		subscriber += tabtabtabtabtab + "}\n"
		subscriber += g.generateReadFieldRec(parser.FieldFromType(op.Type, "received"), false, false, false, tabtabtabtabtab)
		subscriber += tabtabtabtabtab + "iprot.readMessageEnd();\n"
		subscriber += tabtabtabtabtab + fmt.Sprintf("handler.on%s(ctx, received);\n", op.Name)
		subscriber += tabtabtabtab + "}\n"
		subscriber += tabtabtab + "};\n"
		subscriber += tabtab + "}\n\n"
	}

	subscriber += tab + "}\n"

	return subscriber
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
		contents += g.generatedAnnotation()
	}
	contents += fmt.Sprintf("public class F%s {\n\n", s.Name)
	contents += fmt.Sprintf(tab+"private static final Logger logger = LoggerFactory.getLogger(F%s.class);\n\n", s.Name)
	contents += g.generateServiceInterface(s)
	contents += g.generateClient(s)
	contents += g.generateServer(s)
	contents += g.generateServiceArgsResults(s)
	contents += "}"

	_, err := file.WriteString(contents)
	return err
}

func (g *Generator) generateServiceInterface(service *parser.Service) string {
	contents := ""
	if service.Comment != nil {
		contents += g.GenerateBlockComment(service.Comment, tab)
	}
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("public interface Iface extends %s.Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "public interface Iface {\n\n"
	}
	for _, method := range service.Methods {
		comment := []string{}
		if method.Comment != nil {
			comment = append(comment, method.Comment...)
		}

		deprecationValue, deprecated := method.Annotations.Deprecated()
		if deprecated && deprecationValue != "" {
			comment = append(comment, fmt.Sprintf("@deprecated %s", deprecationValue))
		}

		if len(comment) != 0 {
			contents += g.GenerateBlockComment(comment, tabtab)
		}

		if deprecated {
			contents += tabtab+"@Deprecated\n"
		}

		contents += fmt.Sprintf(tabtab+"public %s %s(FContext ctx%s) %s;\n\n",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	}
	contents += tab + "}\n\n"
	return contents
}

func (g *Generator) getServiceExtendsName(service *parser.Service) string {
	serviceName := "F" + service.ExtendsService()
	include := service.ExtendsInclude()
	if include != "" {
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			include = namespace.Value
		} else {
			return serviceName
		}
		serviceName = include + "." + serviceName
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

func (g *Generator) generateClient(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("public static class Client extends %s.Client implements Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "public static class Client implements Iface {\n\n"
	}
	if service.Extends == "" {
		if g.generateAsync() {
			contents += tabtab + "protected ExecutorService asyncExecutor = Executors.newFixedThreadPool(2);\n"
		}
	}
	contents += tabtab + "private Iface proxy;\n\n"

	contents += tabtab + "public Client(FServiceProvider provider, ServiceMiddleware... middleware) {\n"
	if service.Extends != "" {
		contents += tabtabtab + "super(provider, middleware);\n"
	}
	contents += tabtabtab + "Iface client = new InternalClient(provider);\n"
	contents += tabtabtab + "List<ServiceMiddleware> combined = Arrays.asList(middleware);\n"
	contents += tabtabtab + "combined.addAll(provider.getMiddleware());\n"
	contents += tabtabtab + "middleware = combined.toArray(new ServiceMiddleware[0]);\n"
	contents += tabtabtab + "proxy = InvocationHandler.composeMiddleware(client, Iface.class, middleware);\n"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		if method.Comment != nil {
			contents += g.GenerateBlockComment(method.Comment, tabtab)
		}

		_, deprecated := method.Annotations.Deprecated()
		if deprecated {
			contents += tabtab+"@Deprecated\n"
		}

		contents += tabtab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {\n",
			g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))

		if deprecated {
			contents += tabtabtab + fmt.Sprintf("logger.warn(\"Call to deprecated function '%s.%s'\");\n", service.Name, method.Name);
		}

		if method.ReturnType != nil {
			contents += tabtabtab + fmt.Sprintf("return proxy.%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		} else {
			contents += tabtabtab + fmt.Sprintf("proxy.%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		}
		contents += tabtab + "}\n\n"

		if g.generateAsync() {
			contents += g.generateAsyncClientMethod(service, method)
		}
	}
	contents += tab + "}\n\n"
	contents += g.generateInternalClient(service)
	return contents
}

func (g *Generator) generateAsyncClientMethod(service *parser.Service, method *parser.Method) string {
	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, tabtab)
	}
	contents += tabtab + fmt.Sprintf("public Future<%s> %sAsync(final FContext ctx%s) {\n",
		g.generateBoxedReturnValue(method), method.Name, g.generateArgs(method.Arguments, true))
	contents += tabtabtab + fmt.Sprintf("return asyncExecutor.submit(new Callable<%s>() {\n", g.generateBoxedReturnValue(method))
	contents += tabtabtabtab + fmt.Sprintf("public %s call() throws Exception {\n", g.generateBoxedReturnValue(method))
	if method.ReturnType != nil {
		contents += tabtabtabtabtab + fmt.Sprintf("return %s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
	} else {
		contents += tabtabtabtabtab + fmt.Sprintf("%s(%s);\n", method.Name, g.generateClientCallArgs(method.Arguments))
		contents += tabtabtabtabtab + "return null;\n"
	}
	contents += tabtabtabtab + "}\n"
	contents += tabtabtab + "});\n"
	contents += tabtab + "}\n\n"
	return contents
}

func (g *Generator) generateInternalClient(service *parser.Service) string {
	contents := ""
	if service.Extends != "" {
		contents += tab + fmt.Sprintf("private static class InternalClient extends %s.Client implements Iface {\n\n",
			g.getServiceExtendsName(service))
	} else {
		contents += tab + "private static class InternalClient implements Iface {\n\n"
	}

	contents += tabtab + "private FTransport transport;\n"
	contents += tabtab + "private FProtocolFactory protocolFactory;\n"

	contents += tabtab + "public InternalClient(FServiceProvider provider) {\n"
	if service.Extends != "" {
		contents += tabtabtab + "super(provider);\n"
	}
	contents += tabtabtab + "this.transport = provider.getTransport();\n"
	contents += tabtabtab + "this.protocolFactory = provider.getProtocolFactory();\n"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		contents += g.generateClientMethod(service, method)
	}
	contents += tab + "}\n\n"

	return contents
}

func (g *Generator) generateClientMethod(service *parser.Service, method *parser.Method) string {
	methodLower := parser.LowercaseFirstLetter(method.Name)

	contents := ""
	if method.Comment != nil {
		contents += g.GenerateBlockComment(method.Comment, tabtab)
	}
	contents += tabtab + fmt.Sprintf("public %s %s(FContext ctx%s) %s {\n",
		g.generateReturnValue(method), method.Name, g.generateArgs(method.Arguments, false), g.generateExceptions(method.Exceptions))
	contents += tabtabtab + "TMemoryOutputBuffer memoryBuffer = new TMemoryOutputBuffer(this.transport.getRequestSizeLimit());\n"
	contents += tabtabtab + "FProtocol oprot = this.protocolFactory.getProtocol(memoryBuffer);\n"
	contents += tabtabtab + "oprot.writeRequestHeader(ctx);\n"
	msgType := "CALL"
	if method.Oneway {
		msgType = "ONEWAY"
	}
	contents += tabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.%s, 0));\n", methodLower, msgType)
	contents += tabtabtab + fmt.Sprintf("%s_args args = new %s_args();\n", method.Name, method.Name)
	for _, arg := range method.Arguments {
		contents += tabtabtab + fmt.Sprintf("args.set%s(%s);\n", strings.Title(arg.Name), arg.Name)
	}
	contents += tabtabtab + "args.write(oprot);\n"
	contents += tabtabtab + "oprot.writeMessageEnd();\n"
	if method.Oneway {
		contents += tabtabtab + "this.transport.oneway(ctx, memoryBuffer.getWriteBytes());\n"
	} else {
		contents += tabtabtab + "TTransport response = this.transport.request(ctx, memoryBuffer.getWriteBytes());\n"
	}
	if method.Oneway {
		contents += tabtab + "}\n"
		return contents
	}

	contents += "\n"
	contents += tabtabtab + "FProtocol iprot = this.protocolFactory.getProtocol(response);\n"
	contents += tabtabtab + "iprot.readResponseHeader(ctx);\n"
	contents += tabtabtab + "TMessage message = iprot.readMessageBegin();\n"
	contents += tabtabtab + fmt.Sprintf("if (!message.name.equals(\"%s\")) {\n", methodLower)
	contents += tabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationExceptionType.WRONG_METHOD_NAME, \"%s failed: wrong method name\");\n",
		method.Name)
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "if (message.type == TMessageType.EXCEPTION) {\n"
	contents += tabtabtabtab + "TApplicationException e = TApplicationException.read(iprot);\n"
	contents += tabtabtabtab + "iprot.readMessageEnd();\n"
	contents += tabtabtabtab + "TException returnedException = e;\n"
	contents += tabtabtabtab + "if (e.getType() == TApplicationExceptionType.RESPONSE_TOO_LARGE) {\n"
	contents += tabtabtabtabtab + "returnedException = new TTransportException(TTransportExceptionType.RESPONSE_TOO_LARGE, e.getMessage());\n"
	contents += tabtabtabtab + "}\n"
	contents += tabtabtabtab + "throw returnedException;\n"
	contents += tabtabtab + "}\n"
	contents += tabtabtab + "if (message.type != TMessageType.REPLY) {\n"
	contents += tabtabtabtab + fmt.Sprintf(
		"throw new TApplicationException(TApplicationExceptionType.INVALID_MESSAGE_TYPE, \"%s failed: invalid message type\");\n",
		method.Name)
	contents += tabtabtab + "}\n"
	contents += tabtabtab + fmt.Sprintf("%s_result res = new %s_result();\n", method.Name, method.Name)
	contents += tabtabtab + "res.read(iprot);\n"
	contents += tabtabtab + "iprot.readMessageEnd();\n"
	if method.ReturnType != nil {
		contents += tabtabtab + "if (res.isSetSuccess()) {\n"
		contents += tabtabtabtab + "return res.success;\n"
		contents += tabtabtab + "}\n"
	}
	for _, exception := range method.Exceptions {
		contents += tabtabtab + fmt.Sprintf("if (res.%s != null) {\n", exception.Name)
		contents += tabtabtabtab + fmt.Sprintf("throw res.%s;\n", exception.Name)
		contents += tabtabtab + "}\n"
	}
	if method.ReturnType != nil {
		contents += tabtabtab + fmt.Sprintf(
			"throw new TApplicationException(TApplicationExceptionType.MISSING_RESULT, \"%s failed: unknown result\");\n",
			method.Name)
	}
	contents += tabtab + "}\n"

	return contents
}

func (g *Generator) generateExceptions(exceptions []*parser.Field) string {
	contents := "throws TException"
	for _, exception := range exceptions {
		contents += ", " + g.getJavaTypeFromThriftType(exception.Type)
	}
	return contents
}

func (g *Generator) generateServer(service *parser.Service) string {
	contents := ""
	extends := "FBaseProcessor"
	if service.Extends != "" {
		extends = g.getServiceExtendsName(service) + ".Processor"
	}
	contents += tab + fmt.Sprintf("public static class Processor extends %s implements FProcessor {\n\n", extends)

	contents += tabtab + "private Iface handler;\n\n"

	contents += tabtab + "public Processor(Iface iface, ServiceMiddleware... middleware) {\n"
	if service.Extends != "" {
		contents += tabtabtab + "super(iface, middleware);\n"
	}
	contents += tabtabtab + "handler = InvocationHandler.composeMiddleware(iface, Iface.class, middleware);\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "protected java.util.Map<String, FProcessorFunction> getProcessMap() {\n"
	if service.Extends != "" {
		contents += tabtabtab + "java.util.Map<String, FProcessorFunction> processMap = super.getProcessMap();\n"
	} else {
		contents += tabtabtab + "java.util.Map<String, FProcessorFunction> processMap = new java.util.HashMap<>();\n"
	}
	for _, method := range service.Methods {
		contents += tabtabtab + fmt.Sprintf("processMap.put(\"%s\", new %s());\n", parser.LowercaseFirstLetter(method.Name), strings.Title(method.Name))
	}
	contents += tabtabtab + "return processMap;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "protected java.util.Map<String, java.util.Map<String, String>> getAnnotationsMap() {\n"
	if service.Extends != "" {
		contents += tabtabtab + "java.util.Map<String, java.util.Map<String, String>> annotationsMap = super.getAnnotationsMap();\n"
	} else {
		contents += tabtabtab + "java.util.Map<String, java.util.Map<String, String>> annotationsMap = new java.util.HashMap<>();\n"
	}
	for _, method := range service.Methods {
		if len(method.Annotations) > 0 {
			contents += tabtabtab + fmt.Sprintf("java.util.Map<String, String> %sMap = new java.util.HashMap<>();\n", method.Name)
			for _, annotation := range method.Annotations {
				contents += tabtabtab + fmt.Sprintf("%sMap.put(\"%s\", \"%s\");\n", method.Name, annotation.Name, annotation.Value)
			}
			contents += tabtabtab + fmt.Sprintf("annotationsMap.put(\"%s\", %sMap);\n", parser.LowercaseFirstLetter(method.Name), method.Name)
		}
	}
	contents += tabtabtab + "return annotationsMap;\n"
	contents += tabtab + "}\n\n"

	contents += tabtab + "@Override\n"
	contents += tabtab + "public void addMiddleware(ServiceMiddleware middleware) {\n"
	if service.Extends != "" {
		contents += tabtabtab + "super.addMiddleware(middleware);\n"
	}
	contents += tabtabtab + "handler = InvocationHandler.composeMiddleware(handler, Iface.class, new ServiceMiddleware[]{middleware});\n"
	contents += tabtab + "}\n\n"

	for _, method := range service.Methods {
		methodLower := parser.LowercaseFirstLetter(method.Name)
		contents += tabtab + fmt.Sprintf("private class %s implements FProcessorFunction {\n\n", strings.Title(method.Name))

		contents += tabtabtab + "public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {\n"

		if _, ok := method.Annotations.Deprecated(); ok {
			contents += tabtabtabtab + fmt.Sprintf("logger.warn(\"Deprecated function '%s.%s' was called by a client\");\n", service.Name, method.Name)
		}

		contents += tabtabtabtab + fmt.Sprintf("%s_args args = new %s_args();\n", method.Name, method.Name)
		contents += tabtabtabtab + "try {\n"
		contents += tabtabtabtabtab + "args.read(iprot);\n"
		contents += tabtabtabtab + "} catch (TException e) {\n"
		contents += tabtabtabtabtab + "iprot.readMessageEnd();\n"
		if !method.Oneway {
			contents += tabtabtabtabtab + "synchronized (WRITE_LOCK) {\n"
			contents += tabtabtabtabtabtab + fmt.Sprintf("e = writeApplicationException(ctx, oprot, TApplicationExceptionType.PROTOCOL_ERROR, \"%s\", e.getMessage());\n", method.Name)
			contents += tabtabtabtabtab + "}\n"
		}
		contents += tabtabtabtabtab + "throw e;\n"
		contents += tabtabtabtab + "}\n\n"

		contents += tabtabtabtab + "iprot.readMessageEnd();\n"

		if method.Oneway {
			contents += tabtabtabtab + fmt.Sprintf("handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += tabtabtab + "}\n"
			contents += tabtab + "}\n\n"
			continue
		}

		contents += tabtabtabtab + fmt.Sprintf("%s_result result = new %s_result();\n", method.Name, method.Name)
		contents += tabtabtabtab + "try {\n"
		if method.ReturnType == nil {
			contents += tabtabtabtabtab + fmt.Sprintf("handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
		} else {
			contents += tabtabtabtabtab + fmt.Sprintf("result.success = handler.%s(%s);\n", method.Name, g.generateServerCallArgs(method.Arguments))
			contents += tabtabtabtabtab + "result.setSuccessIsSet(true);\n"
		}
		for _, exception := range method.Exceptions {
			contents += tabtabtabtab + fmt.Sprintf("} catch (%s %s) {\n", g.getJavaTypeFromThriftType(exception.Type), exception.Name)
			contents += tabtabtabtabtab + fmt.Sprintf("result.%s = %s;\n", exception.Name, exception.Name)
		}
		contents += tabtabtabtab + "} catch (TApplicationException e) {\n"
		contents += tabtabtabtabtab + "oprot.writeResponseHeader(ctx);\n"
		contents += tabtabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.EXCEPTION, 0));\n", methodLower)
		contents += tabtabtabtabtab + "e.write(oprot);\n"
		contents += tabtabtabtabtab + "oprot.writeMessageEnd();\n"
		contents += tabtabtabtabtab + "oprot.getTransport().flush();\n"
		contents += tabtabtabtabtab + "return;\n"
		contents += tabtabtabtab + "} catch (TException e) {\n"
		contents += tabtabtabtabtab + "synchronized (WRITE_LOCK) {\n"
		contents += tabtabtabtabtabtab + fmt.Sprintf(
			"e = writeApplicationException(ctx, oprot, TApplicationExceptionType.INTERNAL_ERROR, \"%s\", \"Internal error processing %s: \" + e.getMessage());\n",
			methodLower, method.Name)
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "throw e;\n"
		contents += tabtabtabtab + "}\n"
		contents += tabtabtabtab + "synchronized (WRITE_LOCK) {\n"
		contents += tabtabtabtabtab + "try {\n"
		contents += tabtabtabtabtabtab + "oprot.writeResponseHeader(ctx);\n"
		contents += tabtabtabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin(new TMessage(\"%s\", TMessageType.REPLY, 0));\n", methodLower)
		contents += tabtabtabtabtabtab + "result.write(oprot);\n"
		contents += tabtabtabtabtabtab + "oprot.writeMessageEnd();\n"
		contents += tabtabtabtabtabtab + "oprot.getTransport().flush();\n"
		contents += tabtabtabtabtab + "} catch (TTransportException e) {\n"
		contents += tabtabtabtabtabtab + "if (e.getType() == TTransportExceptionType.REQUEST_TOO_LARGE) {\n"
		contents += tabtabtabtabtabtabtab + fmt.Sprintf(
			"writeApplicationException(ctx, oprot, TApplicationExceptionType.RESPONSE_TOO_LARGE, \"%s\", \"response too large: \" + e.getMessage());\n",
			methodLower)
		contents += tabtabtabtabtabtab + "} else {\n"
		contents += tabtabtabtabtabtabtab + "throw e;\n"
		contents += tabtabtabtabtabtab + "}\n"
		contents += tabtabtabtabtab + "}\n"
		contents += tabtabtabtab + "}\n"
		contents += tabtabtab + "}\n"
		contents += tabtab + "}\n\n"
	}

	contents += tab + "}\n\n"

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

func (g *Generator) qualifiedTypeName(t *parser.Type) string {
	param := t.ParamName()
	include := t.IncludeName()
	if include != "" {
		if namespace := g.Frugal.NamespaceForInclude(include, lang); namespace != nil {
			return fmt.Sprintf("%s.%s", namespace.Value, param)
		}
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

func (g *Generator) generatedAnnotation() string {
	anno := fmt.Sprintf("@Generated(value = \"Autogenerated by Frugal Compiler (%s)\"", globals.Version)
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
