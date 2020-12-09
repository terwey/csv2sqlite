// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package delimiter

import (
	"bytes"
	"fmt"
	"strings"
)

// Delimiter is used to denote which character is used to delimit the field in the CSV
type Delimiter []byte

var (
	Delimiter_TAB       Delimiter = []byte("\t")
	Delimiter_PIPE      Delimiter = []byte("|")
	Delimiter_TILDE     Delimiter = []byte("~")
	Delimiter_COMMA     Delimiter = []byte(",")
	Delimiter_SEMICOLON Delimiter = []byte(";")
	Delimiter_DOLLAR    Delimiter = []byte("$")
	Delimiter_CARET     Delimiter = []byte("^")
	Delimiter_POUND     Delimiter = []byte("#")

	// Delimiter_RECORD_SEPARATOR http://www.fileformat.info/info/unicode/char/241e/index.htm
	Delimiter_RECORD_SEPARATOR Delimiter = []byte("\x241E")

	// Delimiter_UNIT_SEPARATOR http://www.fileformat.info/info/unicode/char/241f/index.htm
	Delimiter_UNIT_SEPARATOR Delimiter = []byte("\x241F")
)

func knownDelimiters() []Delimiter {
	return []Delimiter{
		// NB: Delimiter_RECORD_SEPARATOR and Delimiter_UNIT_SEPARATOR need to be first
		// as the character might appear byte wise in another following one
		Delimiter_RECORD_SEPARATOR,
		Delimiter_UNIT_SEPARATOR,
		Delimiter_TAB,
		Delimiter_PIPE,
		Delimiter_TILDE,
		Delimiter_COMMA,
		Delimiter_SEMICOLON,
		Delimiter_DOLLAR,
		Delimiter_CARET,
		Delimiter_POUND,
	}
}

func GetDelimiterNames() []string {
	return []string{
		"RECORD_SEPARATOR",
		"UNIT_SEPARATOR",
		"TAB",
		"PIPE",
		"TILDE",
		"COMMA",
		"SEMICOLON",
		"DOLLAR",
		"CARET",
		"POUND",
	}
}

func stringValueForDelimiter(d Delimiter) string {
	if bytes.Compare(d, Delimiter_RECORD_SEPARATOR) == 0 {
		return "RECORD_SEPARATOR"
	}
	if bytes.Compare(d, Delimiter_UNIT_SEPARATOR) == 0 {
		return "UNIT_SEPARATOR"
	}
	if bytes.Compare(d, Delimiter_TAB) == 0 {
		return "TAB"
	}
	if bytes.Compare(d, Delimiter_PIPE) == 0 {
		return "PIPE"
	}
	if bytes.Compare(d, Delimiter_TILDE) == 0 {
		return "TILDE"
	}
	if bytes.Compare(d, Delimiter_COMMA) == 0 {
		return "COMMA"
	}
	if bytes.Compare(d, Delimiter_SEMICOLON) == 0 {
		return "SEMICOLON"
	}
	if bytes.Compare(d, Delimiter_DOLLAR) == 0 {
		return "DOLLAR"
	}
	if bytes.Compare(d, Delimiter_CARET) == 0 {
		return "CARET"
	}
	if bytes.Compare(d, Delimiter_POUND) == 0 {
		return "POUND"
	}

	return "UNKNOWN"
}

func (d Delimiter) String() string {
	return stringValueForDelimiter(d)
}

func DelimiterFromString(name string) (Delimiter, error) {
	switch strings.ToUpper(name) {
	case "RECORD_SEPARATOR":
		return Delimiter_RECORD_SEPARATOR, nil
	case "UNIT_SEPARATOR":
		return Delimiter_UNIT_SEPARATOR, nil
	case "TAB":
		return Delimiter_TAB, nil
	case "PIPE":
		return Delimiter_PIPE, nil
	case "TILDE":
		return Delimiter_TILDE, nil
	case "COMMA":
		return Delimiter_COMMA, nil
	case "SEMICOLON":
		return Delimiter_SEMICOLON, nil
	case "DOLLAR":
		return Delimiter_DOLLAR, nil
	case "CARET":
		return Delimiter_CARET, nil
	case "POUND":
		return Delimiter_POUND, nil
	}

	return Delimiter{}, fmt.Errorf("Delimiter with name \"%s\" is not recognised", name)
}

func DetectDelimiter(data []byte) Delimiter {
	var highestCount int
	var highestDelimiter Delimiter
	for _, delim := range knownDelimiters() {
		count := bytes.Count(data, delim)
		if count > highestCount {
			highestCount = count
			highestDelimiter = delim
		}
	}

	return highestDelimiter
}
