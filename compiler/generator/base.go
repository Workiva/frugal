package generator

import (
	"fmt"
	"os"
	"unicode"

	"github.com/Workiva/frugal/compiler/parser"
)

// LowercaseFirstLetter of the string.
func LowercaseFirstLetter(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// BaseGenerator contains base generator logic which language generators can
// extend.
type BaseGenerator struct {
	Options map[string]string
	Frugal  *parser.Frugal
	elemNum int
}

// CreateFile creates a new file using the given configuration.
func (b *BaseGenerator) CreateFile(name, outputDir, suffix string, usePrefix bool) (*os.File, error) {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return nil, err
	}
	prefix := ""
	if usePrefix {
		prefix = FilePrefix
	}
	return os.Create(fmt.Sprintf("%s/%s%s.%s", outputDir, prefix, name, suffix))
}

// GenerateNewline adds the specific number of newlines to the given file.
func (b *BaseGenerator) GenerateNewline(file *os.File, count int) error {
	str := ""
	for i := 0; i < count; i++ {
		str += "\n"
	}
	_, err := file.WriteString(str)
	return err
}

// GenerateInlineComment generates an inline comment.
func (b *BaseGenerator) GenerateInlineComment(comment []string, indent string) string {
	inline := ""
	for _, line := range comment {
		inline += indent + "// " + line + "\n"
	}
	return inline
}

// GenerateBlockComment generates a C-style comment.
func (b *BaseGenerator) GenerateBlockComment(comment []string, indent string) string {
	block := indent + "/**\n"
	for _, line := range comment {
		block += indent + " * " + line + "\n"
	}
	block += indent + " */\n"
	return block
}

// SetFrugal sets the Frugal parse tree for this generator.
func (b *BaseGenerator) SetFrugal(f *parser.Frugal) {
	b.Frugal = f
}

func (b *BaseGenerator) GetElem() string {
	s := fmt.Sprintf("elem%d", b.elemNum)
	b.elemNum++
	return s
}
