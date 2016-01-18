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

type structLike string

const (
	structLikeStruct    structLike = "struct"
	structLikeException structLike = "exception"
	structLikeUnion     structLike = "union"
)

var thriftTypes = map[string]bool{
	"bool":   true,
	"byte":   true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"double": true,
	"string": true,
	"binary": true,
}

func generateThriftIDL(dir string, frugal *parser.Frugal) (string, error) {
	file := filepath.Join(dir, fmt.Sprintf("%s.thrift", frugal.Name))
	f, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents := ""
	thrift := frugal.Thrift

	contents += generateNamespaces(thrift.Namespaces)
	includes, err := generateIncludes(frugal)
	if err != nil {
		return "", err
	}
	contents += includes
	contents += generateConstants(frugal)
	contents += generateTypedefs(thrift.Typedefs)
	contents += generateEnums(thrift.Enums)
	contents += generateStructLikes(thrift.Structs, structLikeStruct)
	contents += generateStructLikes(thrift.Unions, structLikeUnion)
	contents += generateStructLikes(thrift.Exceptions, structLikeException)
	contents += generateServices(thrift.Services)

	_, err = f.WriteString(contents)
	return file, err
}

func generateNamespaces(namespaces []*parser.Namespace) string {
	contents := ""
	for _, namespace := range namespaces {
		contents += fmt.Sprintf("namespace %s %s\n", namespace.Scope, namespace.Value)
	}
	contents += "\n"
	return contents
}

func generateIncludes(frugal *parser.Frugal) (string, error) {
	contents := ""
	for _, incl := range frugal.Thrift.Includes {
		include := incl.Value
		if strings.HasSuffix(strings.ToLower(include), ".frugal") {
			// Recurse on frugal includes
			parsed, err := compile(filepath.Join(frugal.Dir, include))
			if err != nil {
				return "", err
			}

			// Lop off .frugal
			includeBase := include[:len(include)-7]

			// Lop off path
			includeName := filepath.Base(includeBase)

			frugal.ParsedIncludes[includeName] = parsed

			// Replace .frugal with .thrift
			include = includeBase + ".thrift"
		}
		contents += fmt.Sprintf("include \"%s\"\n", include)
	}
	contents += "\n"
	return contents, nil
}

func generateConstants(frugal *parser.Frugal) string {
	contents := ""
	complexConstants := make([]*parser.Constant, 0, len(frugal.Thrift.Constants))

	for _, constant := range frugal.Thrift.Constants {
		value := constant.Value
		typeName := frugal.UnderlyingType(constant.Type)
		if isThriftPrimitive(typeName) {
			if typeName == "string" {
				value = fmt.Sprintf(`"%s"`, value)
			}
		} else {
			// Generate complex constants separately after primitives.
			complexConstants = append(complexConstants, constant)
			continue
		}
		if constant.Comment != nil {
			contents += generateThriftDocString(constant.Comment, "")
		}
		contents += fmt.Sprintf("const %s %s = %v\n", constant.Type, constant.Name, value)
	}

	for _, constant := range complexConstants {
		contents += "\n"
		if constant.Comment != nil {
			contents += generateThriftDocString(constant.Comment, "")
		}
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

func generateTypedefs(typedefs []*parser.TypeDef) string {
	contents := ""
	for _, typedef := range typedefs {
		if typedef.Comment != nil {
			contents += generateThriftDocString(typedef.Comment, "")
		}
		contents += fmt.Sprintf("typedef %s %s\n", typedef.Type, typedef.Name)
	}
	contents += "\n"
	return contents
}

func generateEnums(enums []*parser.Enum) string {
	contents := ""
	for _, enum := range enums {
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
			contents += fmt.Sprintf("\t%s,\n", value.Name)
		}
		contents += "}\n\n"
	}
	return contents
}

func generateStructLikes(structs []*parser.Struct, typ structLike) string {
	contents := ""
	for _, strct := range structs {
		if strct.Comment != nil {
			contents += generateThriftDocString(strct.Comment, "")
		}
		contents += fmt.Sprintf("%s %s {\n", typ, strct.Name)
		for _, field := range strct.Fields {
			if field.Comment != nil {
				contents += generateThriftDocString(field.Comment, "\t")
			}
			contents += fmt.Sprintf("\t%d: ", field.ID)
			if field.Optional {
				contents += "optional "
			} else {
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
	}
	return contents
}

func generateServices(services []*parser.Service) string {
	contents := ""
	for _, service := range services {
		if service.Comment != nil {
			contents += generateThriftDocString(service.Comment, "")
		}
		contents += fmt.Sprintf("service %s ", service.Name)
		if service.Extends != "" {
			contents += fmt.Sprintf("extends %s ", service.Extends)
		}
		contents += "{\n"
		for _, method := range service.Methods {
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
				modifier := "required"
				if arg.Optional {
					modifier = "optional"
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
		}
		contents += "}\n\n"
	}
	return contents
}

func generateThrift(frugal *parser.Frugal, idlDir, out, gen string, dryRun bool) error {
	// Generate intermediate Thrift IDL.
	idlFile, err := generateThriftIDL(idlDir, frugal)
	if err != nil {
		return err
	}
	globals.IntermediateIDL = append(globals.IntermediateIDL, idlFile)

	if dryRun {
		return nil
	}

	// Generate Thrift code.
	args := []string{}
	if out != "" {
		args = append(args, "-out", out)
	}
	args = append(args, "-gen", gen)
	args = append(args, idlFile)
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

func isThriftPrimitive(typeName string) bool {
	_, ok := thriftTypes[typeName]
	return ok
}
