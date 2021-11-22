# textwrapper

A golang module for wrapping text on word boundaries.

## Synopsis

This module provides a parser that consumes a string, treating it as UTF-8.  It divides
the string into "whitespace" and "words".  Words are contiguous sequences of non-whitespace
runes.  The parser rearranges the string text into rows that are no longer than a specified
length wide.  It tries to break around whitespace.  If a resulting row is a single word that
exceeds the maximum line length, it breaks the word at exactly the line length, placing the
remainder on the subsequent row (or rows, if the string is long enough).

## Install

```bash
go get -u github.com/blorticus-go/textwrapper
```

## Example

```go
import "github.com/blorticus-go/textwrapper"
import "fmt"

func panicIfError(e error) {
    if e != nil {
        panic(e.Error())
    }
}

func main() {
    fh, err := os.Open(os.Args[1])
    panicIfError(err)

    wrapper := textwrapper.NewTextWrapper()
    formattedString, err := wrapper.ReformatFromReader(fh)
    panicIfError(err)

    fmt.Print(formattedString)
}
```