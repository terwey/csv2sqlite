// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
	"encoding/json"
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
	Character_RECORD_SEPARATOR_SB Character = []byte("\x1E")

	// Character_INFORMATION_SEPARATOR_TWO https://www.fileformat.info/info/unicode/char/001e/index.htm
	Character_INFORMATION_SEPARATOR_TWO Character = []byte("\u001E")

	// Character_RECORD_SEPARATOR http://www.fileformat.info/info/unicode/char/241e/index.htm
	Character_RECORD_SEPARATOR Character = []byte("\u241E")

	// Character_UNIT_SEPARATOR http://www.fileformat.info/info/unicode/char/241f/index.htm
	Character_UNIT_SEPARATOR Character = []byte("\u241F")

	// // Character_INVISIBLE_SEPARATOR https://www.fileformat.info/info/unicode/char/2063/index.htm
	Character_INVISIBLE_SEPARATOR Character = []byte("\u2063")

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

	Character_ZERO_WIDTH_NO_BREAK_SPACE = []byte("\uFEFF")
	Character_OLD_BYTE_ORDER_MARK       = Character_ZERO_WIDTH_NO_BREAK_SPACE
	BOM_UTF8                            = []byte("\uEFBBBF")
	BOM_UTF16_BE                        = Character_ZERO_WIDTH_NO_BREAK_SPACE
	BOM_UTF16_LE                        = []byte("\uFFFE")
	BOM_UTF32_BE                        = []byte("\u0000FEFF")
	BOM_UTF32_LE                        = []byte("\uFFFE0000")
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
	allChars["RECORD_SEPARATOR_SB"] = Character_RECORD_SEPARATOR_SB
	allChars["UNIT_SEPARATOR"] = Character_UNIT_SEPARATOR
	allChars["INVISIBLE_SEPARATOR"] = Character_INVISIBLE_SEPARATOR
	allChars["INFORMATION_SEPARATOR_TWO"] = Character_INFORMATION_SEPARATOR_TWO
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

func (c Character) Length() int {
	return len([]rune(string(c)))
}

func (c Character) String() string {
	if len(c) == 0 {
		return ""
	}

	return string(c)
}

func (c Character) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name  string
		Value string
	}{
		Name:  c.Name(),
		Value: c.String(),
	})
}

type Config struct {
	Delimiter        Character
	Enclosure        Character
	EnclosureIsEmpty bool
	DoubleEnclosure  Character
	enclosureLength  int
	EOL              Character
	EscapeCharacter  Character
	EscapedEnclosure Character

	EnclosureClean bool
}

func NewConfig(delimiter, enclosure Character) Config {
	c := Config{}

	c.SetDelimiter(delimiter)
	c.SetEnclosure(enclosure)

	return c
}

func (cc *Config) SetDelimiter(delim Character) {
	cc.Delimiter = delim
}

func (cc *Config) SetEnclosure(encl Character) {
	cc.Enclosure = append(Character{}, encl...)
	cc.enclosureLength = len(encl)

	// reset
	cc.DoubleEnclosure = make(Character, 0)
	cc.DoubleEnclosure = append(cc.DoubleEnclosure, cc.Enclosure...)
	cc.DoubleEnclosure = append(cc.DoubleEnclosure, cc.Enclosure...)

	if len(cc.EscapeCharacter) == 0 {
		cc.EscapeCharacter = Character("\\")
	}

	cc.EscapedEnclosure = make(Character, 0)
	cc.EscapedEnclosure = append(cc.EscapedEnclosure, cc.EscapeCharacter...)
	cc.EscapedEnclosure = append(cc.EscapedEnclosure, cc.Enclosure...)

	// reset
	cc.EnclosureClean = true
	if bytes.Equal(cc.Enclosure, Character_EMPTY) {
		cc.EnclosureClean = false
	}
}

func (c *Config) EnclosureLength() int {
	return c.enclosureLength
}

func stringValueForCharacter(d Character) string {
	allChars := getAllCharacters()

	for v, char := range allChars {
		if bytes.Compare(d, char) == 0 {
			return v
		}
	}

	return string(d)
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
