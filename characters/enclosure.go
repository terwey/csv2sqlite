// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
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
	for {
		if len(data) < prevPos {
			break
		}
		data = data[prevPos:]
		pos := bytes.Index(data, delimiter)
		if pos == -1 {
			break
		}

		left := data[pos-1 : pos]

		// first find the left
		var leftCount int
		if leftCount = bytes.Count(data[:pos-1], left); leftCount >= 1 {
			realCount := 1
			for i := 1; i < leftCount; i++ {
				subdata := data[pos-i-1 : pos-i]
				if bytes.Equal(subdata, left) {
					realCount++
				} else {
					break
				}
			}

			left = data[pos-realCount : pos]
			leftCount = realCount
		}

		// the right side should be the same length

		var right []byte
		if len(data) >= pos+leftCount+1 {
			right = data[pos+1 : pos+1+leftCount]
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

		prevPos = pos + 1
		prevRight = right
	}

	return nil, 0
}
