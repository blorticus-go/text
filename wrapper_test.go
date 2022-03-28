package text_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blorticus-go/text"
)

type WrapStringTestCase struct {
	testName                   string
	rowLength                  uint
	firstLineIndentString      string
	subsequentLineIndentString string
	unwrappedStrings           []string
	useAReader                 bool
	expectedWrappedStrings     []string
}

func (testCase *WrapStringTestCase) RunTest() error {
	wrapper := text.NewWrapper().
		UsingIndentStringForFirstRow(testCase.firstLineIndentString).
		UsingIndentStringForRowsAfterTheFirst(testCase.subsequentLineIndentString).
		UsingRowWidth(testCase.rowLength)

	for stringsIndex, unwrappedString := range testCase.unwrappedStrings {
		expectedWrappedString := testCase.expectedWrappedStrings[stringsIndex]

		var wrappedString string
		var err error

		if testCase.useAReader {
			reader := strings.NewReader(unwrappedString)
			wrappedString, err = wrapper.WrapUTF8TextFromAReader(reader)
		} else {
			wrappedString, err = wrapper.WrapStringText(unwrappedString)
		}

		if err != nil {
			return err
		}

		if len(wrappedString) != len(expectedWrappedString) {
			tcInfo := ""
			if len(testCase.unwrappedStrings) > 1 {
				tcInfo = fmt.Sprintf(", test string %d", stringsIndex+1)
			}
			return fmt.Errorf("[%s%s] expected wrapped string length = (%d), got = (%d)", testCase.testName, tcInfo, len(expectedWrappedString), len(wrappedString))
		}

		for i := 0; i < len(wrappedString); i++ {
			wrappedStringAsRunes := []rune(wrappedString)
			expectedStringAsRunes := []rune(expectedWrappedString)

			if len(wrappedStringAsRunes) != len(expectedStringAsRunes) {
				return fmt.Errorf("[%s] expected wrapped string rune length = (%d), got = (%d)", testCase.testName, len(expectedStringAsRunes), len(wrappedStringAsRunes))
			}

			for i := 0; i < len(wrappedStringAsRunes); i++ {
				if expectedStringAsRunes[i] != wrappedStringAsRunes[i] {
					return fmt.Errorf("[%s] at offset (%d), expected rune = (%c), got = (%c)", testCase.testName, i, expectedStringAsRunes[i], wrappedStringAsRunes[i])
				}
			}
		}
	}

	return nil
}

var unwrappedString01 string = "This is   a simple \t\n bit of text including non-latin Ḃ\t   \n characters Ϟ"
var unwrappedString02 string = "∀∁∂∃ ∄ ∅∆∇\t ∈∉∊  \r    ∋∌∍∎∏ +   -∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊     ∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ \t∀∁∂∃ ∄ ∅∆∇\t ∈∉∊ "
var unwrappedString03 string = "This is a string 123456789012345678901234567890 and then ∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃ <-- that \t string is 42 characters long."
var unwrappedString04 string = "thisstringhasnospacesinitatallandexceeds-∂∃∀∁∂-thelengthofthecolumnssetupsoitshould,:;∂∃∀∁∂'[\"\\bebrokenrightathecolumnlength"
var emptyUnwrappedString01 string = ""
var whitespaceOnlyUnwrappedString01 string = "\t  \n\r \r    "
var whitespaceOnlyUnwrappedString02 string = "  \n\r \r                    \t\t\r\n                    \r\r        \t"

func wrapTestSet(useReaderRatherThanString bool) (failedTests []error) {
	testNamePreamble := "WrapStringText()"
	if useReaderRatherThanString {
		testNamePreamble = "WrapFromAReader"
	}

	testCases := []*WrapStringTestCase{
		{
			testName:         fmt.Sprintf("%s test 1", testNamePreamble),
			unwrappedStrings: []string{unwrappedString01},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"This is   a simple    bit of\n" +
				"text including non-latin Ḃ\n" +
				"characters Ϟ",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 2", testNamePreamble),
			unwrappedStrings: []string{unwrappedString02},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"∀∁∂∃ ∄ ∅∆∇  ∈∉∊       ∋∌∍∎∏ +\n" +
				"-∀∁∂∃ ∄ ∅∆∇  ∈∉∊ ∀∁∂∃ ∄ ∅∆∇\n" +
				"∈∉∊     ∀∁∂∃ ∄ ∅∆∇  ∈∉∊  ∀∁∂∃\n" +
				"∄ ∅∆∇  ∈∉∊",
			},
		},
		{
			testName:                   fmt.Sprintf("%s test 3", testNamePreamble),
			unwrappedStrings:           []string{unwrappedString01},
			rowLength:                  30,
			subsequentLineIndentString: "    ",
			useAReader:                 useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"This is   a simple    bit of\n" +
				"    text including non-latin Ḃ\n" +
				"    characters Ϟ",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 4", testNamePreamble),
			unwrappedStrings: []string{unwrappedString01, unwrappedString02},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"This is   a simple    bit of\n" +
				"text including non-latin Ḃ\n" +
				"characters Ϟ",

				"∀∁∂∃ ∄ ∅∆∇  ∈∉∊       ∋∌∍∎∏ +\n" +
					"-∀∁∂∃ ∄ ∅∆∇  ∈∉∊ ∀∁∂∃ ∄ ∅∆∇\n" +
					"∈∉∊     ∀∁∂∃ ∄ ∅∆∇  ∈∉∊  ∀∁∂∃\n" +
					"∄ ∅∆∇  ∈∉∊",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 5", testNamePreamble),
			unwrappedStrings: []string{unwrappedString03},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"This is a string\n" +
				"123456789012345678901234567890\n" +
				"and then\n" +
				"∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁∂∃∀∁\n" +
				"∂∃∀∁∂∃∀∁∂∃∀∁∂∃ <-- that\n" +
				"string is 42 characters long.",
			},
		},
		{
			testName:                   fmt.Sprintf("%s test 6", testNamePreamble),
			unwrappedStrings:           []string{unwrappedString01, unwrappedString02},
			rowLength:                  30,
			firstLineIndentString:      "----",
			subsequentLineIndentString: "  ",
			useAReader:                 useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"----This is   a simple    bit\n" +
				"  of text including non-latin\n" +
				"  Ḃ      characters Ϟ",

				"----∀∁∂∃ ∄ ∅∆∇  ∈∉∊\n" +
					"  ∋∌∍∎∏ +   -∀∁∂∃ ∄ ∅∆∇  ∈∉∊\n" +
					"  ∀∁∂∃ ∄ ∅∆∇  ∈∉∊     ∀∁∂∃ ∄\n" +
					"  ∅∆∇  ∈∉∊  ∀∁∂∃ ∄ ∅∆∇  ∈∉∊",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 7", testNamePreamble),
			unwrappedStrings: []string{emptyUnwrappedString01},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"",
			},
		},
		{
			testName:                   fmt.Sprintf("%s test 8", testNamePreamble),
			unwrappedStrings:           []string{emptyUnwrappedString01},
			rowLength:                  30,
			firstLineIndentString:      "----",
			subsequentLineIndentString: "  ",
			useAReader:                 useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 9", testNamePreamble),
			unwrappedStrings: []string{whitespaceOnlyUnwrappedString01, whitespaceOnlyUnwrappedString02},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"",
				"",
			},
		},
		{
			testName:                   fmt.Sprintf("%s test 10", testNamePreamble),
			unwrappedStrings:           []string{whitespaceOnlyUnwrappedString01, whitespaceOnlyUnwrappedString02},
			rowLength:                  30,
			firstLineIndentString:      "----",
			subsequentLineIndentString: "  ",
			useAReader:                 useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"",
				"",
			},
		},
		{
			testName:         fmt.Sprintf("%s test 11", testNamePreamble),
			unwrappedStrings: []string{unwrappedString04},
			rowLength:        30,
			useAReader:       useReaderRatherThanString,
			expectedWrappedStrings: []string{"" +
				"thisstringhasnospacesinitatall\n" +
				"andexceeds-∂∃∀∁∂-thelengthofth\n" +
				"ecolumnssetupsoitshould,:;∂∃∀∁\n" +
				"∂'[\"\\bebrokenrightathecolumnle\n" +
				"ngth",
			},
		},
		{
			testName:                   fmt.Sprintf("%s test 12", testNamePreamble),
			unwrappedStrings:           []string{unwrappedString04},
			rowLength:                  30,
			useAReader:                 useReaderRatherThanString,
			subsequentLineIndentString: "  ",
			expectedWrappedStrings: []string{"" +
				"thisstringhasnospacesinitatall\n" +
				"  andexceeds-∂∃∀∁∂-thelengthof\n" +
				"  thecolumnssetupsoitshould,:;\n" +
				"  ∂∃∀∁∂'[\"\\bebrokenrightatheco\n" +
				"  lumnlength",
			},
		},
	}

	for _, testCase := range testCases {
		if err := testCase.RunTest(); err != nil {
			failedTests = append(failedTests, err)
		}
	}

	return failedTests
}

func TestWrapStringText(t *testing.T) {
	failedTests := wrapTestSet(false)

	for _, failedTestError := range failedTests {
		t.Error(failedTestError.Error())
	}
}

func TestWrapUsingAReader(t *testing.T) {
	failedTests := wrapTestSet(true)

	for _, failedTestError := range failedTests {
		t.Error(failedTestError.Error())
	}
}
