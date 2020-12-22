// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
)

func knownDelimiters() []Character {
	return []Character{
		// NB: Character_RECORD_SEPARATOR and Character_UNIT_SEPARATOR need to be first
		// as the character might appear byte wise in another following one
		Character_RECORD_SEPARATOR,
		Character_UNIT_SEPARATOR,
		Character_TAB,
		Character_PIPE,
		Character_TILDE,
		Character_COMMA,
		Character_SEMICOLON,
		Character_DOLLAR,
		Character_CARET,
		Character_POUND,
	}
}

func GetDelimiterNames() []string {
	var names []string
	for _, c := range knownDelimiters() {
		names = append(names, c.Name())
	}
	return names
}

func DetectDelimiter(data []byte) Character {
	var highestCount int
	var highestDelimiter Character
	for _, delim := range knownDelimiters() {
		count := bytes.Count(data, delim)
		if count > highestCount {
			highestCount = count
			highestDelimiter = delim
		}
	}

	return highestDelimiter
}
