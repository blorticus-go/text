package text

import "bytes"

type FunctionToDetermineWhetherThereIsMoreInput func() (bool, error)

type WrappedTextBuffer struct {
	haveNotReachedEndOfStreamDeterminer    FunctionToDetermineWhetherThereIsMoreInput
	unindentedColumnWidth                  int
	columnsRemainingInCurrentUnwrappedLine int
	initialIndentAsString                  string
	subsequentLineIndentAsString           string
	wrapLineBreakString                    string
	bytesBuffer                            bytes.Buffer
}

func NewWrappedTextBuffer(maximumNumberOfCharactersInAnUnindendedColumn int, delimiterToInsertBetweenWrappedLines string) *WrappedTextBuffer {
	return &WrappedTextBuffer{
		unindentedColumnWidth: maximumNumberOfCharactersInAnUnindendedColumn,
		wrapLineBreakString:   delimiterToInsertBetweenWrappedLines,
	}
}

func (b *WrappedTextBuffer) UsingFunctionToDetermineIfThereIsMoreInput(f FunctionToDetermineWhetherThereIsMoreInput) *WrappedTextBuffer {
	b.haveNotReachedEndOfStreamDeterminer = f
	return b
}

func (b *WrappedTextBuffer) UsingIndents(initialIndentAsRuneSlice []rune, subsequentIndentsAsRuneSlice []rune) *WrappedTextBuffer {
	b.initialIndentAsString = string(initialIndentAsRuneSlice)
	b.subsequentLineIndentAsString = string(subsequentIndentsAsRuneSlice)
	return b
}

func (b *WrappedTextBuffer) ColumnsRemainingInCurrentWrapRow() int {
	return b.columnsRemainingInCurrentUnwrappedLine
}

func (b *WrappedTextBuffer) CurrentWrapLineRemainingWidthIsAtLeast(n int) bool {
	return b.columnsRemainingInCurrentUnwrappedLine >= n
}

func (b *WrappedTextBuffer) WriteRunes(r []rune) error {
	_, err := b.bytesBuffer.WriteString(string(r))
	return err
}

func (b *WrappedTextBuffer) WriteInitialIndent() error {
	_, err := b.bytesBuffer.WriteString(b.initialIndentAsString)
	return err
}

// check to make sure that there are more characters to read with a peek
func (b *WrappedTextBuffer) WriteLineWrap() error {
	if _, err := b.bytesBuffer.WriteString(b.wrapLineBreakString); err != nil {
		return err
	}

	if haveNotReachEndOfStream, err := b.haveNotReachedEndOfStreamDeterminer(); err != nil {
		return err
	} else if haveNotReachEndOfStream {
		if _, err := b.bytesBuffer.WriteString(b.subsequentLineIndentAsString); err != nil {
			return err
		}
	}

	return nil
}
