package generator

// line contains an indentation count and line contents.
type line struct {
	indent  uint
	content string
}

// Printer prints code to a buffer.
type Printer struct {
	output []line
	indent uint
}

// ScopeUp increases indentation by one step.
func (p *Printer) ScopeUp() *Printer {
	p.indent++
	return p
}

// ScopeDown decreases indentation by one step.
func (p *Printer) ScopeDown() *Printer {
	p.indent--
	return p
}

// Println writes the line to the output buffer.
func (p *Printer) Println(ln string) *Printer {
	p.output = append(p.output, line{indent: p.indent, content: ln + "\n"})
	return p
}

// Output returns the formatted output buffer contents as a string.
func (p *Printer) Output(indentCount uint, indent string) string {
	output := ""
	baseIndent := ""
	for i := uint(0); i < indentCount; i++ {
		baseIndent += indent
	}
	for _, line := range p.output {
		output += baseIndent
		for i := uint(0); i < line.indent+1; i++ {
			output += indent
		}
		output += line.content
	}
	return output
}
