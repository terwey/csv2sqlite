// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
	"unicode"
	"unicode/utf8"
)

func knownEnclosures() []Character {
	return []Character{
		Character_QUOTATION_MARK_TRIPLE,
		Character_QUOTATION_MARK_DOUBLE,
		Character_QUOTATION_MARK,
		Character_APOSTROPHE,
		Character_ASTERISK,

		// Character_EMPTY has to be last as it can ALWAYS appear
		Character_EMPTY,
	}
}

func GetEnclosureNames() []string {
	return getNames(knownEnclosures())
}

func DetectEnclosure(data []byte) (Character, int) {
	return detect(data, knownEnclosures())
}

func DetectEnclosureWithDelimiter(data []byte, delimiter Character) (Character, int) {
	if len(delimiter) == 0 {
		return nil, 0
	}

	type Found struct {
		char  Character
		count int
	}

	detected := make(map[string]Found)
	orig := data
	var prevPos int = 0
	// enclosures are optional, we must compare to a previously found right
	var prevRight []byte
	charSize := 1
	for {
		if len(data) < prevPos {
			break
		}
		data = data[prevPos:]
		pos := bytes.Index(data, delimiter)
		if pos == -1 {
			// it is possible that the enclosure appears before the end of the array and no more delimiters are found
			if len(data) == 1 {
				break
			}
			pos = len(data)
		}

		var left []byte

		for {
			left = data[pos-charSize : pos]
			if !utf8.Valid(left) {
				// not valid UTF
				charSize++
				continue
			}
			break
		}

		// first find the left
		leftCount := bytes.Count(data[:pos-charSize], left)
		// we want to look inside the data minus the already found charSize
		if leftCount > 1 {
			subpos := pos - charSize
			realCount := 1
			// iter := 0
			for i := charSize; i < leftCount*charSize; i = i + charSize {
				subdata := data[subpos-i : subpos-i+charSize]
				if bytes.Equal(subdata, left) {
					realCount++
				} else {
					break
				}
			}

			left = data[pos-(charSize*realCount) : pos]
			leftCount = realCount
		}

		// if leftCount is zero increase to 1
		if leftCount == 0 {
			leftCount = 1
		}
		// the right side should be the same length

		var right []byte
		if len(data) >= pos+len(delimiter)+(leftCount*charSize) {
			right = data[pos+len(delimiter) : pos+len(delimiter)+(charSize*leftCount)]
		}

		if bytes.Equal(left, right) || bytes.Equal(left, prevRight) {
			var doubleEncl []byte
			doubleEncl = append(doubleEncl, left...)
			doubleEncl = append(doubleEncl, left...)
			escaped := bytes.Count(orig, doubleEncl)
			if escaped == -1 || escaped == 1 {
				escaped = 0
			}

			c := bytes.Count(orig, left)
			if len(left) == 0 {
				c = 0
			}

			detected[string(left)] = Found{
				char:  Character(left),
				count: c - escaped*2,
			}
			// return Character(left), c - escaped*2
		}

		prevPos = pos + len(delimiter)
		prevRight = right
	}

	if len(detected) == 0 {
		return Character_EMPTY, 0
	}

	var highest Found
	for _, found := range detected {
		if highest.count < found.count && !unicode.IsLetter(found.char.Rune()) {
			highest = found
		}
	}

	return highest.char, highest.count
}
