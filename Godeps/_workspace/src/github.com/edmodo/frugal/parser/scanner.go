// vim: set ts=4 sw=4 tw=99 noet:
//
// Copyright 2014, Edmodo, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this work except in compliance with the License.
// You may obtain a copy of the License in the LICENSE file, or at:
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS"
// BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package parser

import (
	"io/ioutil"
	"strconv"
	"unicode/utf8"
)

const EOF rune = rune(-1)

type Scanner struct {
	Context *CompileContext

	// Character stream.
	stream []byte
	pos    int
	line   int
	col    int

	// If true, the last token was re-buffered and should be read again.
	saved   bool
	current *Token
}

func NewScanner(context *CompileContext) (*Scanner, error) {
	bytes, err := ioutil.ReadFile(context.CurFile)
	if err != nil {
		return nil, err
	}

	return &Scanner{
		Context: context,
		stream:  bytes,
		pos:     0,
		line:    1,
		saved:   false,
		current: nil,
	}, nil
}

// Return the next token, either off the stream or via the token buffer.
func (this *Scanner) next() *Token {
	if this.saved {
		this.saved = false
		return this.current
	}
	this.current = this.scan()
	return this.current
}

// Undo the last scan, re-buffering the token we just read.
func (this *Scanner) undo() {
	if this.saved {
		panic("Can only undo one token!")
	}
	this.saved = true
}

// Decode the next character off the stream, but do not advance the stream.
func (this *Scanner) getChar() (rune, int) {
	if this.pos >= len(this.stream) {
		return EOF, 1
	}

	return utf8.DecodeRune(this.stream[this.pos:])
}

// Peek at the next character in the stream.
func (this *Scanner) peekChar() rune {
	c, _ := this.getChar()
	return c
}

// Grab the next character and advance the stream.
func (this *Scanner) nextChar() rune {
	c, size := this.getChar()
	this.col++
	this.pos += size
	return c
}

// Advance the stream if the next character matches.
func (this *Scanner) matchChar(c rune) bool {
	c2, size := this.getChar()
	if c2 != c {
		return false
	}
	this.col++
	this.pos += size
	return true
}

// Advance to the next line.
func (this *Scanner) nextLine(c rune) {
	if c == '\r' {
		// Eat a '\n' for windows line endings.
		this.matchChar('\n')
	}
	this.line++
	this.col = 1
}

// Return true if the character ends a line.
func (this *Scanner) isEndOfLine(c rune) bool {
	return c == '\r' || c == '\n' || c == EOF
}

// Reads characters until an end-of-line is reached.
func (this *Scanner) readUntilEndOfLine() {
	for {
		c := this.nextChar()
		if this.isEndOfLine(c) {
			this.nextLine(c)
			return
		}
	}
}

// Reads until the end of a multi-line comment is reached.
func (this *Scanner) readMultiLineComment() {
	for {
		c := this.nextChar()

		switch {
		case c == EOF:
			this.Context.ReportError(this.Position(), "reached end-of-file in multi-line comment")
			return

		case this.isEndOfLine(c):
			this.nextLine(c)

		case c == '*':
			if this.matchChar('/') {
				return
			}
		}
	}
}

// Finds the next tokenizable character.
func (this *Scanner) nextTokenChar() (rune, Position) {
	for {
		if this.matchChar(' ') || this.matchChar('\t') {
			// Ignore whitespace.
			continue
		}

		start := this.Position()
		c := this.nextChar()

		// Eat newlines.
		if c == '\n' || c == '\r' {
			this.nextLine(c)
			continue
		}

		// Detect end-of-line comments.
		if c == '/' {
			if this.matchChar('/') {
				this.readUntilEndOfLine()
				continue
			}

			// Detect multi-char comments.
			if this.matchChar('*') {
				this.readMultiLineComment()
				continue
			}
		}

		// Otherwise, return the character.
		return c, start
	}
}

// Return true if the character is a valid identifier start.
func (this *Scanner) isIdentStartChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// Return true if the character can be part of an identifier.
func (this *Scanner) isIdentChar(c rune) bool {
	return this.isIdentStartChar(c) || (c >= '0' && c <= '9')
}

// Return the current position in terms of line/col.
func (this *Scanner) Position() Position {
	return Position{this.line, this.col}
}

// Convert a list of runes into a string.
func runesToString(runes []rune) string {
	buffer := make([]byte, 4)
	bytes := []byte{}
	for _, r := range runes {
		size := utf8.EncodeRune(buffer, r)
		bytes = append(bytes, buffer[:size]...)
	}
	return string(bytes)
}

// Read a string literal.
func (this *Scanner) readStringLiteral(firstChar rune) (TokenKind, interface{}) {
	runes := []rune{}

	// Loop until we see the first character repeated. We don't support
	// backticks yet (does Thrift?)
	for !this.matchChar(firstChar) {
		c := this.nextChar()

		// If we reach a newline, error.
		if this.isEndOfLine(c) {
			this.Context.ReportError(this.Position(), "reached end-of-file in string literal")
			return TOK_LITERAL_STRING, runesToString(runes)
		}

		runes = append(runes, c)
	}

	return TOK_LITERAL_STRING, runesToString(runes)
}

// Read a numeric literal.
func (this *Scanner) readNumberLiteral(firstChar rune) (TokenKind, interface{}) {
	runes := []rune{firstChar}

	// Only support base10 integers so far.
	for {
		c := this.peekChar()
		if !(c >= '0' && c <= '9') {
			break
		}
		runes = append(runes, this.nextChar())
	}

	str := runesToString(runes)
	data, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		this.Context.ReportError(this.Position(), "could not parse integer literal: %s", err.Error())
		return TOK_LITERAL_INT, int64(0)
	}
	return TOK_LITERAL_INT, data
}

// Read an identifier or keyword.
func (this *Scanner) readIdentifier(firstChar rune) (TokenKind, interface{}) {
	runes := []rune{firstChar}
	for {
		// Peek at the next character. If it's not a valid identifier character,
		// we end the loop.
		c := this.peekChar()
		if !this.isIdentChar(c) {
			break
		}

		// Add the next character to this identifier.
		runes = append(runes, this.nextChar())
	}

	// Convert the list of characters into a string, then check if it's a keyword.
	str := runesToString(runes)
	if kind, ok := KeywordMap[str]; ok {
		return kind, nil
	}
	return TOK_IDENTIFIER, str
}

// Scan the next token from the stream.
func (this *Scanner) scan() *Token {
	c, start := this.nextTokenChar()

	tok := &Token{
		Kind: TOK_ERROR,
		Loc: Location{
			Start: start,
		},
	}

	switch c {
	case EOF:
		tok.Kind = TOK_EOF
	case '{':
		tok.Kind = TOK_LBRACE
	case '}':
		tok.Kind = TOK_RBRACE
	case '[':
		tok.Kind = TOK_LBRACKET
	case ']':
		tok.Kind = TOK_RBRACKET
	case '(':
		tok.Kind = TOK_LPAREN
	case ')':
		tok.Kind = TOK_RPAREN
	case '<':
		tok.Kind = TOK_LT
	case '>':
		tok.Kind = TOK_GT
	case '=':
		tok.Kind = TOK_ASSIGN
	case ':':
		tok.Kind = TOK_COLON
	case '.':
		tok.Kind = TOK_DOT
	case ',':
		tok.Kind = TOK_COMMA
	case ';':
		tok.Kind = TOK_SEMICOLON
	case '"':
		tok.Kind, tok.Data = this.readStringLiteral(c)
	default:
		if this.isIdentStartChar(c) {
			tok.Kind, tok.Data = this.readIdentifier(c)
		} else if c >= '0' && c <= '9' {
			tok.Kind, tok.Data = this.readNumberLiteral(c)
		} else {
			this.Context.ReportError(start, "Unrecognized character: %c", c)
		}
	}

	tok.Loc.End = this.Position()
	return tok
}
