package java

import (
	"fmt"

	"github.com/Workiva/frugal/compiler/generator"
)

// Field is a Java class member variable.
type Field struct {
	Modifier     string
	Final        bool
	Type         string
	Name         string
	InitialValue string
}

func (f Field) format() string {
	modifier := ""
	if f.Modifier != "" {
		modifier = f.Modifier
	}
	if f.Final {
		if modifier != "" {
			modifier += " "
		}
		modifier += "final"
	}
	if modifier != "" {
		modifier += " "
	}
	formatted := fmt.Sprintf("%s%s %s", modifier, f.Type, f.Name)
	if f.InitialValue != "" {
		formatted += fmt.Sprintf(" = %s", f.InitialValue)
	}
	return formatted
}

// Argument is a Java method or constructor parameter.
type Argument struct {
	Final bool
	Type  string
	Name  string
}

func (a Argument) format() string {
	modifier := ""
	if a.Final {
		modifier = "final "
	}
	return fmt.Sprintf("%s%s %s", modifier, a.Type, a.Name)
}

// Constructor is a Java class constructor.
type Constructor struct {
	Modifier  string
	Arguments []Argument
	Body      *generator.Printer
}

func (c Constructor) format(indentCount uint, indent, name string) string {
	modifier := ""
	if c.Modifier != "" {
		modifier = c.Modifier + " "
	}
	output := fmt.Sprintf("%s%s%s(", indent, modifier, name)
	prefix := ""
	for _, arg := range c.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += c.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// EnumConstructor is a Java enum constructor (Modifier is ignored).
type EnumConstructor Constructor

func (e EnumConstructor) format(indentCount uint, indent, name string) string {
	output := fmt.Sprintf("%sprivate %s(", indent, name)
	prefix := ""
	for _, arg := range e.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += e.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// Method is a Java class method.
type Method struct {
	Modifier   string
	Name       string
	ReturnType string
	Static     bool
	Arguments  []Argument
	Body       *generator.Printer
}

func (m Method) format(indentCount uint, indent string) string {
	modifier := ""
	if m.Modifier != "" {
		modifier = m.Modifier + " "
	}
	if m.Static {
		modifier += "static "
	}
	output := fmt.Sprintf("%s%s%s %s(", indent, modifier, m.ReturnType, m.Name)
	prefix := ""
	for _, arg := range m.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += m.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// EnumValue is a Java enum value.
type EnumValue struct {
	Name      string
	Arguments []string
}

func (e EnumValue) format() string {
	formatted := e.Name
	if len(e.Arguments) > 0 {
		formatted += "("
		prefix := ""
		for _, arg := range e.Arguments {
			formatted += prefix + arg
			prefix = ", "
		}
		formatted += ")"
	}
	return formatted
}
