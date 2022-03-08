// Package text provides methods for reading UTF-8 text handling
package text

import (
	"io"
)

// Wrapper provides UTF-8 text line wrapping. At the start of each inserted line, any whitespace is removed.
// The Wrapper converts tab characters (code point 9) to a set (but configurable) number of spaces (code point 32).
// Line break sequences (code point 10 and and 13) are converted into a single space. If more than one line break
// occurs in a row, they are flattened to a single space. The Wrapper breaks the text stream into runs of
// unicode spaces and non-spaces. Runs of non-spaces are considered "words". At the start of a new wrapped line,
// any leading whitespace (after conversions noted above) are discarded. Alternating word and whitespace sequences
// are emitted until the (configurable) column width (minus the length of the line break sequence) is reached. If
// the column width would break a word the Wrapper rewinds to the last whitespace sequence, removes it, then
// inserts the (configurable) line break sequence. If there is no whitespace sequence before the start of the line
// (i.e., there are more contiguous word characters in the line than the column width minus the line break sequence
// length), the line break sequence is inserted at the column width minus the line break sequence length and the
// word continues on the next line.  At the start of each indented line, a configurable preamble may be inserted.
// The characters in the preamble count against the row column count. A configurable preamble may also be be inserted
// on the initial line, but it is configured separately from the subsequent line indents in case the two should
// be different (a common case is to have no initial indent, but have a fixed number of spaces on subsequent lines).
type Wrapper struct {
	columnsPerRow               uint
	initialLineIndentString     string
	subsequentLinesIndentString string
	lineBreakSequence           string
}

// NewWrapper creates an empty wrapper.
func NewWrapper() *Wrapper {
	return &Wrapper{
		columnsPerRow:               80,
		initialLineIndentString:     "",
		subsequentLinesIndentString: "",
		lineBreakSequence:           "\n",
	}
}

// ChangeRowWidthTo changes the column width to the provided value. The default column width is 80.
func (wrapper *Wrapper) ChangeRowWidthTo(numberOfColumns uint) *Wrapper {
	if numberOfColumns <= uint(len(wrapper.initialLineIndentString)) || numberOfColumns <= uint(len(wrapper.subsequentLinesIndentString)) {
		panic("RowWidth must be larger than row indent string")
	}

	wrapper.columnsPerRow = numberOfColumns
	return wrapper
}

// UsingRowWidth is the same as ChangeRowWidthTo(), but provides a more readable name if this is
// chained with the constructor, as in:
//    wrapper := text.NewWrapper().UsingRowWidth(120)
func (wrapper *Wrapper) UsingRowWidth(numberOfColumns uint) *Wrapper {
	return wrapper.ChangeRowWidthTo(numberOfColumns)
}

// ChangeIndentStringForFirstRowTo sets the indent string for the first row. By default, it is the empty string
// (meaning "no indent").
func (wrapper *Wrapper) ChangeIndentStringForFirstRowTo(indent string) *Wrapper {
	if wrapper.columnsPerRow <= uint(len(indent)) {
		panic("RowWidth must be larger than row indent string")
	}

	wrapper.initialLineIndentString = indent
	return wrapper
}

// UsingIndentStringForFirstRow is the same as ChangeIndentStringForFirstRowTo(), but provides a more
// readable name if this chained with the constructor.
func (wrapper *Wrapper) UsingIndentStringForFirstRow(indent string) *Wrapper {
	return wrapper.ChangeIndentStringForFirstRowTo(indent)
}

// ChangeIndentStringForRowsAfterTheFirstTo sets the indent string for rows after the first. By default, it
// is the empty string (meaning "no indent").
func (wrapper *Wrapper) ChangeIndentStringForRowsAfterTheFirstTo(indent string) *Wrapper {
	if wrapper.columnsPerRow <= uint(len(indent)) {
		panic("RowWidth must be larger than row indent string")
	}

	wrapper.subsequentLinesIndentString = indent
	return wrapper
}

// UsingIndentStringForRowsAfterTheFirst is the same as ChangeIndentStringForRowsAfterTheFirstTo(), but
// provides a more readable name if this chained with the constructor.
func (wrapper *Wrapper) UsingIndentStringForRowsAfterTheFirst(indent string) *Wrapper {
	return wrapper.ChangeIndentStringForRowsAfterTheFirstTo(indent)
}

// WrapUTF8TextFromAReader resets the Wrapper parser state. It begins to Read from the supplied reader,
// treating incoming bytes as UTF-8 encoded text, wrapping using the rules described above. It will
// Read() until it reaches io.EOF. It returns the wrapped text or an error if one occurs.
func (wrapper *Wrapper) WrapUTF8TextFromAReader(reader io.Reader) (wrappedText string, err error) {
	return "", nil
}

// WrapStringText takes a string and wraps it using the rules described above. It returns the wrapped
// text or an error if one occurs.
func (wrapper *Wrapper) WrapStringText(unwrappedString string) (wrappedText string, err error) {
	return "", err
}
