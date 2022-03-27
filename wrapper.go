// Package text provides methods for reading UTF-8 text handling
package text

import (
	"bytes"
	"io"
	"unicode"

	"github.com/blorticus-go/nibblers"
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
	initialLineIndentString     []rune
	subsequentLinesIndentString []rune
	lineBreakSequence           string
	nibblerMatcher              *nibblers.UTF8NibblerMatcher
	nibbler                     nibblers.UTF8Nibbler
}

// NewWrapper creates an empty wrapper.
func NewWrapper() *Wrapper {
	return &Wrapper{
		columnsPerRow:               79,
		initialLineIndentString:     nil,
		subsequentLinesIndentString: nil,
		lineBreakSequence:           "\n",
	}
}

// ChangeRowWidthTo changes the column width to the provided value. The default column width is 79.
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
	wrapper.initialLineIndentString = []rune(indent)

	if len(wrapper.initialLineIndentString) > int(wrapper.columnsPerRow) {
		panic("RowWidth must be larger than row indent string")
	}

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
	wrapper.subsequentLinesIndentString = []rune(indent)

	if len(wrapper.subsequentLinesIndentString) > int(wrapper.columnsPerRow) {
		panic("RowWidth must be larger than row indent string")
	}

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
	nibbler := nibblers.NewUTF8ReaderNibbler(reader)
	return wrapper.wrapFromNibbler(nibbler)
}

// WrapStringText takes a string and wraps it using the rules described above. It returns the wrapped
// text or an error if one occurs.
func (wrapper *Wrapper) WrapStringText(unwrappedString string) (wrappedText string, err error) {
	nibbler := nibblers.NewUTF8StringNibbler(unwrappedString)
	return wrapper.wrapFromNibbler(nibbler)
}

type unwrappedTextProcessingState struct {
	columnsRemainingInCurrentWrappedLine   uint
	lineHoldBuffer                         []rune
	chunkHoldBuffer                        []rune
	errorOrEOFCollectedFromChunkProcessing error
}

func wrappedTextStringOrEmptyStringBasedOnErrorOrEOF(err error, bufferOfWrappedText *bytes.Buffer) (string, error) {
	if err == io.EOF {
		return bufferOfWrappedText.String(), nil
	}

	return bufferOfWrappedText.String(), err
}

func (wrapper *Wrapper) wrapFromNibbler(nibbler nibblers.UTF8Nibbler) (wrappedText string, err error) {
	var bufferOfWrappedText bytes.Buffer

	wrapper.nibblerMatcher = nibblers.NewUTF8NibblerMatcher(nibbler)
	wrapper.nibbler = nibbler

	wordChunkBuffer := make([]rune, wrapper.columnsPerRow)
	whitespaceChunkBuffer := make([]rune, wrapper.columnsPerRow)

	if atEndOfStream, err := wrapper.afterRemovingContiguousWhitespace().reachedTheEndOfTheStream(); atEndOfStream {
		return "", nil
	} else if err != nil {
		return "", err
	}

	if _, err := bufferOfWrappedText.WriteString(string(wrapper.initialLineIndentString)); err != nil {
		return "", err
	}

	columnsRemainingInCurrentWrappedLine := int(wrapper.columnsPerRow) - len(wrapper.initialLineIndentString)

	numberOfRunesInLastWhitespaceChunk := 0
	atTheStartOfALine := true

	for {
		wordRunesRead, err := wrapper.nibblerMatcher.ReadConsecutiveWordCharactersInto(wordChunkBuffer[:columnsRemainingInCurrentWrappedLine])
		if err == io.EOF {
			return wrappedTextStringOrEmptyStringBasedOnErrorOrEOF(err, &bufferOfWrappedText)
		}

		if wordRunesRead == columnsRemainingInCurrentWrappedLine {
			// if there were no whitespace chunks in this line, then this word is at least as long as an entire line
			if numberOfRunesInLastWhitespaceChunk == 0 {
				if _, err := bufferOfWrappedText.WriteString(string(wordChunkBuffer[:wordRunesRead])); err != nil {
					return bufferOfWrappedText.String(), err
				}

				if atEndOfStream, err := wrapper.afterRemovingContiguousWhitespace().reachedTheEndOfTheStream(); atEndOfStream {
					return bufferOfWrappedText.String(), nil
				} else if err != nil {
					return bufferOfWrappedText.String(), err
				}

				if wrapper.insertLineBreakAndIndentInto(&bufferOfWrappedText); err != nil {
					return bufferOfWrappedText.String(), err
				}

				columnsRemainingInCurrentWrappedLine = int(wrapper.columnsPerRow) - len(wrapper.subsequentLinesIndentString)
				atTheStartOfALine = true
			} else {
				// if we are at the end of the stream, return
				nextUnreadRune, err := wrapper.nibbler.PeekAtNextCharacter()
				if err == io.EOF {
					if _, writeErr := bufferOfWrappedText.WriteString(string(wordChunkBuffer[:wordRunesRead])); writeErr != nil {
						return bufferOfWrappedText.String(), writeErr
					}
					return bufferOfWrappedText.String(), nil
				} else if err != nil {
					return bufferOfWrappedText.String(), err
				}

				// if next character is a space, then we have a word that ends exactly at the line wrap length end
				if unicode.IsSpace(nextUnreadRune) {
					if numberOfRunesInLastWhitespaceChunk > 0 {
						if _, err := bufferOfWrappedText.WriteString(changeAllWhitespaceToAnASCIISpace(whitespaceChunkBuffer[:numberOfRunesInLastWhitespaceChunk])); err != nil {
							return bufferOfWrappedText.String(), err
						}
					}

					if _, err := bufferOfWrappedText.WriteString(string(wordChunkBuffer[:wordRunesRead])); err != nil {
						return bufferOfWrappedText.String(), err
					}

					if atEndOfStream, err := wrapper.afterRemovingContiguousWhitespace().reachedTheEndOfTheStream(); atEndOfStream {
						return "", nil
					} else if err != nil {
						return "", err
					}

					if err := wrapper.insertLineBreakAndIndentInto(&bufferOfWrappedText); err != nil {
						return bufferOfWrappedText.String(), err
					}

					columnsRemainingInCurrentWrappedLine = int(wrapper.columnsPerRow) - len(wrapper.subsequentLinesIndentString)
					numberOfRunesInLastWhitespaceChunk = 0
					atTheStartOfALine = true
				} else {
					// word buffer only has a fragment of a word but must wrap
					if err := wrapper.insertLineBreakAndIndentInto(&bufferOfWrappedText); err != nil {
						return bufferOfWrappedText.String(), err
					}

					if _, err := bufferOfWrappedText.WriteString(string(wordChunkBuffer[:wordRunesRead])); err != nil {
						return bufferOfWrappedText.String(), err
					}

					columnsRemainingInCurrentWrappedLine = int(wrapper.columnsPerRow) - len(wrapper.subsequentLinesIndentString) - wordRunesRead
					numberOfRunesInLastWhitespaceChunk = 0
					atTheStartOfALine = false
				}
			}
		} else {
			if numberOfRunesInLastWhitespaceChunk > 0 {
				if _, err := bufferOfWrappedText.WriteString(changeAllWhitespaceToAnASCIISpace(whitespaceChunkBuffer[:numberOfRunesInLastWhitespaceChunk])); err != nil {
					return bufferOfWrappedText.String(), err
				}
			}

			if _, err := bufferOfWrappedText.WriteString(string(wordChunkBuffer[:wordRunesRead])); err != nil {
				return bufferOfWrappedText.String(), err
			}

			columnsRemainingInCurrentWrappedLine -= wordRunesRead
			numberOfRunesInLastWhitespaceChunk = 0
			atTheStartOfALine = false
		}

		if !atTheStartOfALine {
			whitespaceRunesRead, err := wrapper.nibblerMatcher.ReadConsecutiveWhitespaceInto(whitespaceChunkBuffer[:columnsRemainingInCurrentWrappedLine])
			if err != nil {
				return wrappedTextStringOrEmptyStringBasedOnErrorOrEOF(err, &bufferOfWrappedText)
			}

			// whitespace continues to end of wrappable line, so wrap and don't write accumulated whitespace
			if whitespaceRunesRead == columnsRemainingInCurrentWrappedLine {
				if atEndOfStream, err := wrapper.afterRemovingContiguousWhitespace().reachedTheEndOfTheStream(); atEndOfStream {
					return bufferOfWrappedText.String(), nil
				} else if err != nil {
					return bufferOfWrappedText.String(), err
				}

				if err := wrapper.insertLineBreakAndIndentInto(&bufferOfWrappedText); err != nil {
					return bufferOfWrappedText.String(), err
				}

				columnsRemainingInCurrentWrappedLine = int(wrapper.columnsPerRow) - len(wrapper.subsequentLinesIndentString)
				numberOfRunesInLastWhitespaceChunk = 0
				atTheStartOfALine = true
			} else {
				columnsRemainingInCurrentWrappedLine -= whitespaceRunesRead
				numberOfRunesInLastWhitespaceChunk = whitespaceRunesRead
				atTheStartOfALine = false
			}
		}
	}
}

func changeAllWhitespaceToAnASCIISpace(whitespaceRunes []rune) string {
	for i := range whitespaceRunes {
		whitespaceRunes[i] = ' '
	}

	return string(whitespaceRunes)
}

func (wrapper *Wrapper) insertLineBreakAndIndentInto(bufferOfWrappedText *bytes.Buffer) error {
	if _, err := bufferOfWrappedText.WriteString(wrapper.lineBreakSequence); err != nil {
		return err
	}

	if _, err := bufferOfWrappedText.WriteString(string(wrapper.subsequentLinesIndentString)); err != nil {
		return err
	}

	return nil
}

func processingHasReachedTheEndOfTheNibblerStreamFor(nibbler nibblers.UTF8Nibbler) bool {
	if _, err := nibbler.PeekAtNextCharacter(); err == io.EOF {
		return true
	}

	return false
}

type intercallState struct {
	lastCallError error
	wrapper       *Wrapper
}

func (wrapper *Wrapper) afterRemovingContiguousWhitespace() *intercallState {
	if _, err := wrapper.nibblerMatcher.DiscardConsecutiveWhitespaceCharacters(); err != nil {
		return &intercallState{
			lastCallError: err,
			wrapper:       wrapper,
		}
	}

	return &intercallState{
		lastCallError: nil,
		wrapper:       wrapper,
	}
}

func (s *intercallState) reachedTheEndOfTheStream() (bool, error) {
	if s.lastCallError != nil {
		if s.lastCallError == io.EOF {
			return true, io.EOF
		}
		return false, s.lastCallError
	}

	if _, err := s.wrapper.nibbler.PeekAtNextCharacter(); err == io.EOF {
		return true, io.EOF
	} else if err != nil {
		return false, err
	}

	return false, nil
}
