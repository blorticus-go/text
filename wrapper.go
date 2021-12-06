// Package textwrapper provides methods for reading UTF-8 text, and wrapping it into lines that do not exceed a desired length.
package textwrapper

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TextWrapper is the primary object used to wrap text.
type TextWrapper struct {
	builder                        strings.Builder
	lengthOfCurrentLine            int
	runeColumnsPerRow              int
	newLineIndentText              []rune
	translateLinebreaksToSpace     bool
	tabstopWidth                   int
	rowSeparatorRune               rune
	whitespacesRuneBuffer          []rune
	numberOfRunesPerRowAfterIndent int
}

// NewTextWrapper creates a new TextWrapper.  It sets the column width (i.e., the maximum line length) to
// 80, does not use line wrap indenting, translates tab runes into 4 spaces, translates newline sequences in the
// text to a space rune (codepoint 32), and uses a newline rune (codepoint 10) as the row separator.
func NewTextWrapper() *TextWrapper {
	return &TextWrapper{
		builder:                        strings.Builder{},
		lengthOfCurrentLine:            0,
		runeColumnsPerRow:              80,
		newLineIndentText:              []rune{},
		translateLinebreaksToSpace:     true,
		tabstopWidth:                   4,
		rowSeparatorRune:               '\n',
		whitespacesRuneBuffer:          make([]rune, 0, 10),
		numberOfRunesPerRowAfterIndent: 80,
	}
}

// SetColumnWidth changes the maximum line length.  This length does not include the trailing
// row separator.
func (wrapper *TextWrapper) SetColumnWidth(columnsPerLine uint) *TextWrapper {
	wrapper.runeColumnsPerRow = int(columnsPerLine)
	if wrapper.runeColumnsPerRow <= len(wrapper.newLineIndentText) {
		panic("indent text length non-sensically equal to or longer than column width")
	}

	wrapper.numberOfRunesPerRowAfterIndent = int(columnsPerLine) - len(wrapper.newLineIndentText)

	return wrapper
}

// SetIndentForEachCreatedRow inserts the indentString (treated as UTF-8) at the start of
// each line after a wrap operation.  It is not applied to the first line.
func (wrapper *TextWrapper) SetIndentForEachCreatedRow(indentString string) *TextWrapper {
	if indentString == "" {
		wrapper.newLineIndentText = []rune{}
	} else {
		wrapper.newLineIndentText = make([]rune, 0, len(indentString))
		for _, r := range indentString {
			wrapper.newLineIndentText = append(wrapper.newLineIndentText, r)
		}
	}

	if wrapper.runeColumnsPerRow <= len(wrapper.newLineIndentText) {
		panic("indent text length non-sensically equal to or longer than column width")
	}

	wrapper.numberOfRunesPerRowAfterIndent = wrapper.runeColumnsPerRow - len(indentString)

	return wrapper
}

// DoNotTranslateNewlineSequencesToSingleSpace disables the default behavior, whereby a sequence
// of linewrap characters (codepoint 10 or 13) in the source text are translated into a single
// space.  Instead, they are left as-is.
func (wrapper *TextWrapper) DoNotTranslateNewlineSequencesToSingleSpace() *TextWrapper {
	wrapper.translateLinebreaksToSpace = false
	return wrapper
}

// SetTabstopWidth changes the number of spaces (codepoint 32) that a tab rune (codepoint 9) is
// converted into.  Tabs are always converted because, if they are not, then a wrapped string
// may exceed its length limit when rendered, depending on how the renderer treats the tab.
func (wrapper *TextWrapper) SetTabstopWidth(spacesInTabstop uint) {
	wrapper.tabstopWidth = int(spacesInTabstop)
}

// AddText adds text to an accumulating internal buffer.  Use AccumulatedWrappedText() to return the
// rendered text after adding all desired string.
func (wrapper *TextWrapper) AddText(text string) {
	if wrapper.runeColumnsPerRow == 0 {
		wrapper.builder.WriteString(text)
		return
	}

	for i, bytesConsumedFromText := 0, 0; i < len(text); i += bytesConsumedFromText {
		if bytesConsumedFromText = wrapper.parseContiguousWhitespaceIntoStringBuilder(text[i:]); bytesConsumedFromText == 0 {
			bytesConsumedFromText = wrapper.parserWordIntoStringBuffer(text[i:])
		}
	}
}

// AccumulatedWrappedText returns the text thus far accumulated in its wrapped format.  This
// is generally used in conjunction with AddText() calls.
func (wrapper *TextWrapper) AccumulatedWrappedText() string {
	return wrapper.builder.String()
}

// Reset clears the accumulated wrapped text buffer, and re-initializes the parser in order
// to start processing a new string.
func (wrapper *TextWrapper) Reset() {
	wrapper.builder.Reset()
	wrapper.emptyWhitespaceRuneBuffer()
	wrapper.lengthOfCurrentLine = 0
}

// WrapString takes a string, treating it as complete UTF-8 text, and returns it wrapped.  It then
// resets the wrapper.
func (wrapper *TextWrapper) WrapString(text string) string {
	wrapper.AddText(text)
	r := wrapper.AccumulatedWrappedText()
	wrapper.Reset()
	return r
}

// WrapFromReader reads from an io.Reader until it reaches the end of the input stream,
// wrapping the input text, and returning the wrapped format.  A returned error would be
// an error returned from the Reader.  io.EOF is not returned.  This method expects the
// reader to return bytes on UTF-8 boundaries.  After returning the wrapped string,
// the wrapper is reset.
func (wrapper *TextWrapper) WrapFromReader(reader io.Reader) (string, error) {
	readBuffer := make([]byte, 9000)
	for {
		bytesRead, err := reader.Read(readBuffer)
		if err != nil && err != io.EOF {
			return "", err
		}

		wrapper.AddText(string(readBuffer[:bytesRead]))

		if err == io.EOF {
			r := wrapper.AccumulatedWrappedText()
			wrapper.Reset()
			return r, nil
		}
	}
}

func (wrapper *TextWrapper) insertRowSeparatorIntoBuilderAndMoveToNextLine() {
	wrapper.builder.WriteRune(wrapper.rowSeparatorRune)
	for _, indentRune := range wrapper.newLineIndentText {
		wrapper.builder.WriteRune(indentRune)
	}
	wrapper.lengthOfCurrentLine = len(wrapper.newLineIndentText)
}

type runeWordTracker struct {
	sourceStringTextForRunes                   string
	runes                                      []rune
	countOfUnprocessedRunes                    int
	byteOffsetInTextAtTheEndOfEachRune         []int
	byteOffsetInTextAtStartOfNextUnwrittenRune int
}

func (wrapper *TextWrapper) parserWordIntoStringBuffer(text string) (bytesConsumed int) {
	runesInNextWord, textBufOffsetAtEndOfEachRune := extractNextWordRunesFrom(text)

	tracker := &runeWordTracker{
		sourceStringTextForRunes:                   text,
		runes:                                      runesInNextWord,
		countOfUnprocessedRunes:                    len(runesInNextWord),
		byteOffsetInTextAtTheEndOfEachRune:         textBufOffsetAtEndOfEachRune,
		byteOffsetInTextAtStartOfNextUnwrittenRune: 0,
	}

	wrapper.parseRunesFromTextIntoStringBuffer(tracker)

	bytesConsumed = textBufOffsetAtEndOfEachRune[len(textBufOffsetAtEndOfEachRune)-1] + 1

	return bytesConsumed
}

//func (wrapper *TextWrapper)

func (wrapper *TextWrapper) parseRunesFromTextIntoStringBuffer(tracker *runeWordTracker) {
	switch remainingColumnsInCurrentRow := wrapper.runeColumnsPerRow - wrapper.lengthOfCurrentLine; {

	//case ()

	case remainingColumnsInCurrentRow == 0:

	case remainingColumnsInCurrentRow < len(wrapper.whitespacesRuneBuffer):
		if remainingColumnsInCurrentRow == len(wrapper.newLineIndentText) {
		}

	case remainingColumnsInCurrentRow < len(wrapper.whitespacesRuneBuffer)+len(tracker.runes):
		wrapper.insertRowSeparatorIntoBuilderAndMoveToNextLine()
		wrapper.emptyWhitespaceRuneBuffer()
		wrapper.parseRunesFromTextIntoStringBuffer(tracker)

	case remainingColumnsInCurrentRow > tracker.countOfUnprocessedRunes:
		indexOfFirstByte := tracker.byteOffsetInTextAtStartOfNextUnwrittenRune
		indexOfLastByte := tracker.byteOffsetInTextAtTheEndOfEachRune[len(tracker.byteOffsetInTextAtTheEndOfEachRune)-1]
		wrapper.lengthOfCurrentLine += wrapper.writeWhitespaceBufferIntoBuilderAndClearBuffer()
		wrapper.builder.WriteString(string(tracker.sourceStringTextForRunes[indexOfFirstByte : indexOfLastByte+1]))
		wrapper.lengthOfCurrentLine += tracker.countOfUnprocessedRunes

	case remainingColumnsInCurrentRow == tracker.countOfUnprocessedRunes:
		indexOfFirstByte := tracker.byteOffsetInTextAtStartOfNextUnwrittenRune
		indexOfLastByte := tracker.byteOffsetInTextAtTheEndOfEachRune[len(tracker.byteOffsetInTextAtTheEndOfEachRune)-1]
		wrapper.builder.WriteString(string(tracker.sourceStringTextForRunes[indexOfFirstByte : indexOfLastByte+1]))
		wrapper.insertRowSeparatorIntoBuilderAndMoveToNextLine()

	case wrapper.runeColumnsPerRow > tracker.countOfUnprocessedRunes:
		indexOfFirstByte := tracker.byteOffsetInTextAtStartOfNextUnwrittenRune
		indexOfLastByteInThisLine := indexOfFirstByte + remainingColumnsInCurrentRow
		wrapper.insertRowSeparatorIntoBuilderAndMoveToNextLine()
		wrapper.builder.WriteString(string(tracker.sourceStringTextForRunes[indexOfFirstByte : indexOfLastByteInThisLine+1]))
		wrapper.lengthOfCurrentLine = tracker.countOfUnprocessedRunes

	case wrapper.runeColumnsPerRow == tracker.countOfUnprocessedRunes:
		wrapper.insertRowSeparatorIntoBuilderAndMoveToNextLine()
		wrapper.parseRunesFromTextIntoStringBuffer(tracker)
	}
}

func (wrapper *TextWrapper) writeWhitespaceBufferIntoBuilderAndClearBuffer() (numberOfRunesWritten int) {
	numberOfRunesWritten = len(wrapper.whitespacesRuneBuffer)
	for _, whitespaceRune := range wrapper.whitespacesRuneBuffer {
		wrapper.builder.WriteRune(rune(whitespaceRune))
	}

	wrapper.emptyWhitespaceRuneBuffer()

	return numberOfRunesWritten
}

func extractNextWordRunesFrom(text string) (runesInNextWord []rune, indexOfLastByteInTextBufForEachRune []int) {
	runesInNextWord = make([]rune, 0, 10)
	indexOfLastByteInTextBufForEachRune = make([]int, 0, 10)

	if len(text) == 0 {
		return runesInNextWord, indexOfLastByteInTextBufForEachRune
	}

	for textBufIndexAtStartOfNextRune := 0; textBufIndexAtStartOfNextRune < len(text); {
		nextRune, runeLengthInBytes := utf8.DecodeRuneInString(text[textBufIndexAtStartOfNextRune:])
		if unicode.IsSpace(nextRune) {
			return runesInNextWord, indexOfLastByteInTextBufForEachRune
		}

		runesInNextWord = append(runesInNextWord, nextRune)
		indexOfLastByteInTextBufForEachRune = append(indexOfLastByteInTextBufForEachRune, textBufIndexAtStartOfNextRune+runeLengthInBytes-1)
		textBufIndexAtStartOfNextRune += runeLengthInBytes
	}

	return runesInNextWord, indexOfLastByteInTextBufForEachRune
}

func (wrapper *TextWrapper) parseContiguousWhitespaceIntoStringBuilder(text string) (bytesConsumed int) {
	bytesConsumedFromTextForWhitespaceRunes := wrapper.extractIntoWhitespaceBufferContiguousWhitespaceRunesFrom(text, wrapper.translateLinebreaksToSpace, wrapper.tabstopWidth)
	numberOfContiguousWhitespaceRunes := len(wrapper.whitespacesRuneBuffer)

	if wrapper.lengthOfCurrentLine+numberOfContiguousWhitespaceRunes >= wrapper.runeColumnsPerRow {
		wrapper.insertRowSeparatorIntoBuilderAndMoveToNextLine()
		wrapper.emptyWhitespaceRuneBuffer()
		return bytesConsumedFromTextForWhitespaceRunes
	}

	return bytesConsumedFromTextForWhitespaceRunes
}

func (wrapper *TextWrapper) emptyWhitespaceRuneBuffer() {
	wrapper.whitespacesRuneBuffer = wrapper.whitespacesRuneBuffer[:0]
}

func (wrapper *TextWrapper) extractIntoWhitespaceBufferContiguousWhitespaceRunesFrom(text string, translateLinebreaksToSpace bool, tabstopWidth int) (bytesConsumedFromText int) {
	bytesConsumedFromText = 0

	for _, nextRune := range text {
		switch nextRune {
		case '\n', '\r':
			if translateLinebreaksToSpace {
				wrapper.whitespacesRuneBuffer = append(wrapper.whitespacesRuneBuffer, ' ')
			} else {
				wrapper.whitespacesRuneBuffer = append(wrapper.whitespacesRuneBuffer, nextRune)
			}
			bytesConsumedFromText++

		case '\t':
			for i := 0; i < tabstopWidth; i++ {
				wrapper.whitespacesRuneBuffer = append(wrapper.whitespacesRuneBuffer, ' ')
			}
			bytesConsumedFromText++

		default:
			if unicode.IsSpace(nextRune) {
				wrapper.whitespacesRuneBuffer = append(wrapper.whitespacesRuneBuffer, nextRune)
				bytesConsumedFromText += utf8.RuneLen(nextRune)
			} else {
				return bytesConsumedFromText
			}
		}
	}

	return bytesConsumedFromText
}
