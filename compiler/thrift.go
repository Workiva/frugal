package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

func generateThriftIDL(dir string, frugal *parser.Frugal) (string, error) {
	file := filepath.Join(dir, fmt.Sprintf("%s.thrift", frugal.Name))
	if exists(file) {
		// Trying to generate an intermediate Thrift IDL but the .thrift file
		// already exists. First check if we generated it. If so, skip it. If
		// not, return an error.
		for _, intermediate := range globals.IntermediateIDL {
			if intermediate == file {
				return file, nil
			}
		}
		return "", fmt.Errorf("Couldn't generate intermediate Thrift, file already exists: %s", file)
	}
	f, err := os.Create(file)
	if err != nil {
		return "", err
	}
	globals.IntermediateIDL = append(globals.IntermediateIDL, file)
	defer f.Close()

	contents := ""

	for _, stmt := range frugal.OrderedDefinitions {
		switch v := stmt.(type) {
		case *parser.Namespace:
			contents += generateNamespace(v)
		case *parser.Constant:
			contents += generateConstant(frugal, v)
		case *parser.Enum:
			contents += generateEnum(v)
		case *parser.TypeDef:
			contents += generateTypedef(v)
		case *parser.Struct: // Also applies to exceptions and unions
			contents += generateStructLike(v)
		case *parser.Service:
			contents += generateService(v)
		case *parser.Include:
			include, err := generateInclude(frugal, v)
			if err != nil {
				return "", err
			}
			contents += include
		}
	}

	_, err = f.WriteString(contents)
	return file, err
}

func generateNamespace(namespace *parser.Namespace) string {
	return fmt.Sprintf("namespace %s %s\n\n", namespace.Scope, namespace.Value)
}

func generateInclude(frugal *parser.Frugal, incl *parser.Include) (string, error) {
	contents := ""
	// Recurse on includes
	include := incl.Value
	if !strings.HasSuffix(include, ".thrift") && !strings.HasSuffix(include, ".frugal") {
		return "", fmt.Errorf("Bad include name: %s", include)
	}

	parsed, err := compile(filepath.Join(frugal.Dir, include),
		strings.HasSuffix(include, ".thrift"), globals.Recurse)
	if err != nil {
		return "", err
	}

	// Lop off extension (.frugal or .thrift)
	includeBase := include[:len(include)-7]

	// Lop off path
	includeName := filepath.Base(includeBase)

	frugal.ParsedIncludes[includeName] = parsed

	// Replace .frugal with .thrift
	include = includeBase + ".thrift"
	contents += fmt.Sprintf("include \"%s\"\n", include)
	contents += "\n"
	return contents, nil
}

func generateConstant(frugal *parser.Frugal, constant *parser.Constant) string {
	contents := ""
	if constant.Comment != nil {
		contents += generateThriftDocString(constant.Comment, "")
	}

	value := constant.Value
	underlyingType := frugal.UnderlyingType(constant.Type)
	if parser.IsThriftPrimitive(underlyingType) || frugal.IsEnum(underlyingType) {
		if underlyingType.Name == "string" {
			value = fmt.Sprintf(`"%s"`, value)
		}
		contents += fmt.Sprintf("const %s %s = %v\n", constant.Type, constant.Name, value)
	} else {
		contents += fmt.Sprintf("const %s %s = %s\n", constant.Type, constant.Name,
			generateComplexConstant(constant))
	}

	contents += "\n"
	return contents
}

func generateComplexConstant(constant *parser.Constant) string {
	switch constant.Type.Name {
	case "map":
		return generateMapLiteral(constant.Value.([]parser.KeyValue), 1)
	case "list":
		return generateListLiteral(constant.Value.([]interface{}), 1)
	case "set":
		return generateListLiteral(constant.Value.([]interface{}), 1)
	default:
		return generateMapLiteral(constant.Value.([]parser.KeyValue), 1)
	}

	return ""
}

func generateMapLiteral(entries []parser.KeyValue, indent int) string {
	nesting := ""
	for i := indent - 1; i > 0; i-- {
		nesting += "\t"
	}
	str := "{\n"
	for _, entry := range entries {
		switch entry.Key.(type) {
		case string:
			str += fmt.Sprintf(`%s"%s": `, indentN(indent), entry.Key)
		default:
			str += fmt.Sprintf(`%s%v: `, indentN(indent), entry.Key)
		}
		switch v := entry.Value.(type) {
		case string:
			str += fmt.Sprintf("\"%s\"", v)
		case []interface{}:
			str += generateListLiteral(v, indent+1)
		case []parser.KeyValue:
			str += generateMapLiteral(v, indent+1)
		default:
			str += fmt.Sprintf("%v", v)
		}
		str += ",\n"
	}
	str += nesting + "}"
	return str
}

func generateListLiteral(list []interface{}, indent int) string {
	nesting := ""
	for i := indent - 1; i > 0; i-- {
		nesting += "\t"
	}
	str := "[\n"
	for _, val := range list {
		switch v := val.(type) {
		case string:
			str += fmt.Sprintf("%s\"%s\"", indentN(indent), v)
		case []interface{}:
			str += indentN(indent) + generateListLiteral(v, indent+1)
		case []parser.KeyValue:
			str += indentN(indent) + generateMapLiteral(v, indent+1)
		default:
			str += fmt.Sprintf("%s%v", indentN(indent), v)
		}
		str += ",\n"
	}
	str += nesting + "]"
	return str
}

func indentN(indent int) string {
	str := ""
	for i := 0; i < indent; i++ {
		str += "\t"
	}
	return str
}

func generateTypedef(typedef *parser.TypeDef) string {
	contents := ""
	if typedef.Comment != nil {
		contents += generateThriftDocString(typedef.Comment, "")
	}
	contents += fmt.Sprintf("typedef %s %s\n", typedef.Type, typedef.Name)
	contents += "\n"
	return contents
}

func generateEnum(enum *parser.Enum) string {
	contents := ""
	if enum.Comment != nil {
		contents += generateThriftDocString(enum.Comment, "")
	}
	contents += fmt.Sprintf("enum %s {\n", enum.Name)
	values := make([]*parser.EnumValue, 0, len(enum.Values))
	for _, value := range enum.Values {
		values = append(values, value)
	}
	sort.Sort(enumValues(values))
	for _, value := range values {
		if value.Comment != nil {
			contents += generateThriftDocString(value.Comment, "\t")
		}
		contents += fmt.Sprintf("\t%s = %d,\n", value.Name, value.Value)
	}
	contents += "}\n\n"
	return contents
}

func generateStructLike(strct *parser.Struct) string {
	contents := ""
	if strct.Comment != nil {
		contents += generateThriftDocString(strct.Comment, "")
	}
	contents += fmt.Sprintf("%s %s {\n", strct.Type, strct.Name)
	for _, field := range strct.Fields {
		if field.Comment != nil {
			contents += generateThriftDocString(field.Comment, "\t")
		}
		contents += fmt.Sprintf("\t%d: ", field.ID)
		if field.Modifier == parser.Optional {
			contents += "optional "
		} else if field.Modifier == parser.Required {
			contents += "required "
		}
		contents += fmt.Sprintf("%s %s", field.Type.String(), field.Name)
		if field.Default != nil {
			def := field.Default
			defStr := ""
			switch d := def.(type) {
			case string:
				defStr = fmt.Sprintf(`"%s"`, d)
			default:
				defStr = fmt.Sprintf("%v", d)
			}
			contents += fmt.Sprintf(" = %s", defStr)
		}
		contents += ",\n"
	}
	contents += "}\n\n"
	return contents
}

func generateService(service *parser.Service) string {
	contents := ""
	if service.Comment != nil {
		contents += generateThriftDocString(service.Comment, "")
	}
	contents += fmt.Sprintf("service %s ", service.Name)
	if service.Extends != "" {
		contents += fmt.Sprintf("extends %s ", service.Extends)
	}
	contents += "{\n"
	for _, method := range service.Methods {
		contents += generateMethod(method)
	}
	contents += "}\n\n"
	return contents
}

func generateMethod(method *parser.Method) string {
	contents := ""
	if method.Comment != nil {
		contents += generateThriftDocString(method.Comment, "\t")
	}
	contents += "\t"
	if method.Oneway {
		contents += "oneway "
	}
	if method.ReturnType == nil {
		contents += "void "
	} else {
		contents += fmt.Sprintf("%s ", method.ReturnType.String())
	}
	contents += fmt.Sprintf("%s(", method.Name)
	prefix := ""
	for _, arg := range method.Arguments {
		modifier := ""
		if arg.Modifier == parser.Optional {
			modifier = "optional"
		} else if arg.Modifier == parser.Required {
			modifier = "required"
		}
		contents += fmt.Sprintf("%s%d:%s %s %s", prefix, arg.ID,
			modifier, arg.Type.String(), arg.Name)
		if arg.Default != nil {
			def := arg.Default
			defStr := ""
			switch d := def.(type) {
			case string:
				defStr = fmt.Sprintf(`"%s"`, d)
			default:
				defStr = fmt.Sprintf("%v", d)
			}
			contents += fmt.Sprintf(" = %s", defStr)
		}
		prefix = ", "
	}
	contents += ")"
	if len(method.Exceptions) > 0 {
		contents += " throws ("
		prefix := ""
		for _, exception := range method.Exceptions {
			contents += fmt.Sprintf("%s%d:%s %s", prefix, exception.ID,
				exception.Type.String(), exception.Name)
			prefix = ", "
		}
		contents += ")"
	}
	contents += ",\n\n"
	return contents
}

func generateThrift(frugal *parser.Frugal, idlDir, file, out, gen string) error {
	// Generate Thrift code.
	args := []string{}
	if out != "" {
		args = append(args, "-out", out)
	}
	args = append(args, "-gen", gen)
	args = append(args, file)
	// TODO: make thrift command configurable
	if out, err := exec.Command("thrift", args...).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}

func generateThriftDocString(comment []string, indent string) string {
	docstr := indent + "/**\n"
	for _, line := range comment {
		docstr += indent + " * " + line + "\n"
	}
	docstr += indent + " */\n"
	return docstr
}

type enumValues []*parser.EnumValue

func (e enumValues) Len() int {
	return len(e)
}

func (e enumValues) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e enumValues) Less(i, j int) bool {
	return e[i].Value < e[j].Value
}
