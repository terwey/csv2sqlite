// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
	"sort"
	"strings"
	"unicode/utf8"
)

// Character is used to denote which character is used for either delimiter, enclosure or escape
type Character []byte

var (
	Character_TAB       Character = []byte("\t")
	Character_PIPE      Character = []byte("|")
	Character_TILDE     Character = []byte("~")
	Character_COMMA     Character = []byte(",")
	Character_SEMICOLON Character = []byte(";")
	Character_DOLLAR    Character = []byte("$")
	Character_CARET     Character = []byte("^")
	Character_POUND     Character = []byte("#")

	// Character_RECORD_SEPARATOR http://www.fileformat.info/info/unicode/char/241e/index.htm
	Character_RECORD_SEPARATOR Character = []byte("\x241E")

	// Character_UNIT_SEPARATOR http://www.fileformat.info/info/unicode/char/241f/index.htm
	Character_UNIT_SEPARATOR Character = []byte("\x241F")

	// Enclosures
	Character_QUOTATION_MARK        Character = []byte(`"`)
	Character_QUOTATION_MARK_DOUBLE Character = []byte(`""`)
	Character_QUOTATION_MARK_TRIPLE Character = []byte(`"""`)
	Character_APOSTROPHE            Character = []byte("'")
	Character_ASTERISK              Character = []byte("*")

	// Line Feeds
	Character_CARRIAGE_RETURN_LINE_FEED Character = []byte("\r\n")
	Character_CARRIAGE_RETURN           Character = []byte("\r")
	Character_LINE_FEED                 Character = []byte("\n")

	Character_EMPTY Character = []byte("")

	Character_NIL Character = nil
)

func getAllCharacters() map[string]Character {
	allChars := make(map[string]Character)
	allChars["TAB"] = Character_TAB
	allChars["PIPE"] = Character_PIPE
	allChars["TILDE"] = Character_TILDE
	allChars["COMMA"] = Character_COMMA
	allChars["SEMICOLON"] = Character_SEMICOLON
	allChars["DOLLAR"] = Character_DOLLAR
	allChars["CARET"] = Character_CARET
	allChars["POUND"] = Character_POUND
	allChars["RECORD_SEPARATOR"] = Character_RECORD_SEPARATOR
	allChars["UNIT_SEPARATOR"] = Character_UNIT_SEPARATOR
	allChars["QUOTATION_MARK"] = Character_QUOTATION_MARK
	allChars["QUOTATION_MARK_DOUBLE"] = Character_QUOTATION_MARK_DOUBLE
	allChars["QUOTATION_MARK_TRIPLE"] = Character_QUOTATION_MARK_TRIPLE
	allChars["APOSTROPHE"] = Character_APOSTROPHE
	allChars["ASTERISK"] = Character_ASTERISK
	allChars["EMPTY"] = Character_EMPTY

	return allChars
}

func GetAllCharacterNames() []string {
	allChars := getAllCharacters()
	var all []Character
	for _, char := range allChars {
		all = append(all, char)
	}

	return getNames(all)
}

func CharacterFromString(name string) (Character, error) {
	allChars := getAllCharacters()
	if char, ok := allChars[strings.ToUpper(name)]; ok {
		return char, nil
	}

	return Character(name), nil
}

func (c Character) Name() string {
	return stringValueForCharacter(c)
}

func (c Character) Rune() rune {
	r, _ := utf8.DecodeRune(c)
	return r
}

func (c Character) String() string {
	r := c.Rune()
	return string(r)
}

func stringValueForCharacter(d Character) string {
	allChars := getAllCharacters()

	for v, char := range allChars {
		if bytes.Compare(d, char) == 0 {
			return v
		}
	}

	return string(d)

	return "UNKNOWN"
}

func getNames(known []Character) []string {
	var names []string
	for _, c := range known {
		names = append(names, c.Name())
	}

	sort.SliceStable(names, func(i, j int) bool { return names[i] < names[j] })

	return names
}

func detect(data []byte, known []Character) (Character, int) {
	var highestCount int
	highestChar := Character_EMPTY
	var longestLen int
	for _, char := range known {
		if bytes.Compare(char, Character_EMPTY) == 0 {
			continue
		}
		count := bytes.Count(data, char)
		if count > highestCount {
			highestCount = count
			highestChar = char
		}
		if count != 0 {
			if length := len(char); length > longestLen {
				longestLen = length
			}
		}
	}

	return highestChar, highestCount
}
