// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
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
			break
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
		var leftCount int
		// we want to look inside the data minus the already found charSize
		if leftCount = bytes.Count(data[:pos-charSize], left); leftCount > 1 {
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

			return Character(left), c - escaped*2
		}

		prevPos = pos + len(delimiter)
		prevRight = right
	}

	return nil, 0
}
