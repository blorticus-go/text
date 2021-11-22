package textwrapper_test

import (
	"fmt"
	"testing"

	"github.com/blorticus-go/textwrapper"
)

type WrapStringTestCase struct {
	testName  string
	rowLength uint
	//indentString          string
	unwrappedString       string
	expectedWrappedString string
}

func (testCase *WrapStringTestCase) RunTest() error {
	wrapper := textwrapper.NewTextWrapper()
	wrapper.SetColumnWidth(testCase.rowLength)

	wrappedString := wrapper.WrapString(testCase.unwrappedString)

	if len(wrappedString) != len(testCase.expectedWrappedString) {
		return fmt.Errorf("[%s] expected wrapped string length = (%d), got = (%d)", testCase.testName, len(testCase.expectedWrappedString), len(wrappedString))
	}

	for i := 0; i < len(wrappedString); i++ {
		wrappedStringAsRunes := []rune(wrappedString)
		expectedStringAsRunes := []rune(testCase.expectedWrappedString)

		if len(wrappedStringAsRunes) != len(expectedStringAsRunes) {
			return fmt.Errorf("[%s] expected wrapped string rune length = (%d), got = (%d)", testCase.testName, len(expectedStringAsRunes), len(wrappedStringAsRunes))
		}

		for i := 0; i < len(wrappedStringAsRunes); i++ {
			if expectedStringAsRunes[i] != wrappedStringAsRunes[i] {
				return fmt.Errorf("[%s] at offset (%d), expected rune = (%c), got = (%c)", testCase.testName, i, expectedStringAsRunes[i], wrappedStringAsRunes[i])
			}
		}
	}

	return nil
}

var unwrappedString01 string = "This is   a simple \t\n bit of text including non-latin Ḃ\t   \n characters Ϟ"

//var unwrappedString02 string = "  and N🁃⌘ a \r\n\r\n      \t    \rsecond string with additional length  "
//var unwrappedString03 string = "12345 67890 abcde FGHIϤJKL\t mnoPQ      RST\n\r\tuvwXyZ "
//var unwrappedString04 string = "∀∁∂∃ ∄ ∅∆∇\t ∈∉∊  \r    ∋∌∍∎∏  +-  ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊     ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ \t∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ "

func TestTextWrapString(t *testing.T) {
	testCases := []*WrapStringTestCase{
		{
			testName:        "WrapString() test 1",
			unwrappedString: unwrappedString01,
			rowLength:       30,
			expectedWrappedString: "" +
				"This is   a simple      bit of\n" +
				"text including non-latin Ḃ\n" +
				"characters Ϟ",
		},
	}

	for _, testCase := range testCases {
		if err := testCase.RunTest(); err != nil {
			t.Error(err.Error())
		}
	}
}
