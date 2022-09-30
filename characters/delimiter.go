// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

func knownDelimiters() []Character {
	return []Character{
		// NB: Character_RECORD_SEPARATOR and Character_UNIT_SEPARATOR need to be first
		// as the character might appear byte wise in another following one
		Character_RECORD_SEPARATOR,
		Character_UNIT_SEPARATOR,
		Character_INVISIBLE_SEPARATOR,
		Character_TAB,
		Character_PIPE,
		Character_TILDE,
		Character_COMMA,
		Character_SEMICOLON,
		Character_DOLLAR,
		Character_CARET,
		Character_POUND,
		Character_RECORD_SEPARATOR_SB,
		Character_INFORMATION_SEPARATOR_TWO,
	}
}

func GetDelimiterNames() []string {
	return getNames(knownDelimiters())
}

func DetectDelimiter(data []byte) Character {
	delim, _ := detect(data, knownDelimiters())
	if len(delim) == 0 {
		log.Printf("delimiter not detected, setting to COMMA")
		// panic("delimiter not detected")
		return Character_COMMA
	}
	return delim
}
