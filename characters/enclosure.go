// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
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

func isKnownEnclosure(char Character) bool {
	for _, known := range knownEnclosures() {
		if bytes.Equal(known, char) && len(char) != 0 {
			// debug.Debug("known: %s len: %d", char, len(char))
			return true
		}
	}
	return false
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

	if len(data) == 0 {
		return nil, 0
	}

	orig := make([]byte, len(data))
	copy(orig, data)

	type Found struct {
		char      Character
		count     int
		occurence int
	}

	// scary mf
	detected := make(map[string]*Found)
	detectedRight := make(map[string]*Found)
	var prevPos int = 0
	// enclosures are optional, we must compare to a previously found right
	var prevRight []byte

	if debug.DebugLog {
		debug.Debug("in: %s", data)
		debug.BreakMe()
	}

	for {
		charSize := 1
		if len(data) < prevPos {
			break
		}

		// if prevPos == 0 then we are at the start of a new set, let's check that as right and compare to left
		// if prevPos == 0 {
		var start []byte
		start, charSize, _ = findRight(data, prevPos, charSize)
		if !bytes.Contains(start, delimiter) {
			// debug.Debug("set prevRight: %s", start)
			prevRight = start
		}
		// }

		data = data[prevPos:]
		// debug.Debug("start loop data %s", data)
		pos := bytes.Index(data, delimiter)
		if pos == -1 {
			// it is possible that the enclosure appears before the end of the array and no more delimiters are found
			if len(data) <= 1 {
				break
			}
			pos = len(data)
		}

		if pos-charSize <= -1 {
			prevPos = pos + len(delimiter)

			// right, charSize, rightCount := findRight(data, prevPos, charSize)
			right, _, _ := findRight(data, prevPos, charSize)
			if !bytes.Contains(right, delimiter) {
				// debug.Debug("set prevRight: %s", right)
				prevRight = right
			}
			// debug.Debug("right: %s, charSize: %d", right, charSize)

			continue
		}

		if isKnownEnclosure(prevRight) {
			// debug.Debug("adding prevRight: %s", prevRight)
			if c := getCount(orig, prevRight); c >= 2 {
				f, ok := detectedRight[string(prevRight)]
				if !ok {
					f = &Found{
						char:  Character(prevRight),
						count: c,
					}
					detectedRight[string(prevRight)] = f
				}
				f.occurence++
			}
		}

		var left []byte
		var leftCount int
		left, charSize, leftCount = findLeft(data, pos, charSize)
		// the right side should be the same length

		if debug.DebugLog {
			debug.BreakMe()
		}

		var right []byte
		if rightPos := pos + len(delimiter) + (charSize * leftCount); len(data) >= rightPos {
			// debug.Debug("rightPos: %d", rightPos)
			right = data[pos+len(delimiter) : rightPos]
			if bytes.Contains(right, delimiter) {
				right = nil
			}
		}

		if bytes.Equal(left, right) || bytes.Equal(left, prevRight) {
			// debug.Debug("adding left: %s", left)

			if c := getCount(orig, left); c >= 2 {
				f, ok := detected[string(left)]
				if !ok {
					f = &Found{
						char:  Character(left),
						count: c,
					}
					detected[string(left)] = f
				}
				f.occurence++
			}
			// return Character(left), c - escaped*2
		}

		// check if left was found in the detectedRight right
		if _, ok := detectedRight[string(left)]; ok {
			// debug.Debug("adding left as it was part of detectedRight: %s", left)
			if c := getCount(orig, left); c >= 2 {
				f, ok := detected[string(left)]
				if !ok {
					f = &Found{
						char:  Character(left),
						count: c,
					}
					detected[string(left)] = f
				}
				f.occurence++
			}
		}

		prevPos = pos + len(delimiter)
		prevRight = right
	}

	for key := range detectedRight {
		if leftFound, ok := detected[key]; ok {
			// debug.Debug("increasing leftFound occurence: %s %d + %d", key, leftFound.occurence, detectedRight[key].occurence)
			leftFound.occurence += detectedRight[key].occurence
		}
		// debug.Debug("detectedRight: (%s) count: %d, occurence: %d", key, detectedRight[key].count, detectedRight[key].occurence)
	}

	if len(detected) == 0 {
		if len(detectedRight) == 0 {
			return Character_EMPTY, -1
		}

		detected = detectedRight
	}

	if debug.DebugLog {
		for key := range detected {
			debug.Debug("detected: (%s) count: %d, occurence: %d", key, detected[key].count, detected[key].occurence)
		}
	}

	// this approach does NOT work
	// // slight weight for the quotation mark
	// if f, ok := detected["\""]; ok {
	// 	f.count = int(math.RoundToEven(float64(f.count) + 0.5))
	// }

	// remove escaped enclosure
	for key := range detected {
		var escaped []byte
		escaped = append(escaped, detected[key].char...)
		escaped = append(escaped, detected[key].char...)

		// debug.Debug("deleting: %s\n", string(escaped))
		delete(detected, string(escaped))
		escaped = append(escaped, detected[key].char...)
		// debug.Debug("deleting: %s\n", string(escaped))
		delete(detected, string(escaped))
	}

	if debug.DebugLog {
		debug.BreakMe()
	}

	highest := Found{
		char:      Character_EMPTY,
		count:     -1,
		occurence: -1,
	}
	for _, found := range detected {
		if bytes.Equal(found.char, []byte(" ")) {
			continue
		}
		if found.count%2 != 0 {
			continue
		}
		// chances of it being longer than 2 4-byte characters is rare
		if len(found.char) > 8 {
			continue
		}
		if highest.occurence < found.occurence && !unicode.IsLetter(found.char.Rune()) && !unicode.IsNumber(found.char.Rune()) {
			highest = *found
		}
		if !isKnownEnclosure(highest.char) && highest.occurence > found.occurence && isKnownEnclosure(found.char) {
			// debug.Debug("overriding because of knownenclosure: %s (%d)", found.char, len(found.char))
			highest = *found
		}
	}

	if highest.occurence == 1 {
		return Character_EMPTY, -1
	}

	if highest.count%2 != 0 {
		return Character_EMPTY, -1
	}

	if debug.DebugLog {
		debug.BreakMe()
		debug.Debug("enclosure found: %s", highest.char)
	}

	return append(Character{}, highest.char...), highest.count
}

func findLeft(data []byte, pos int, charSize int) ([]byte, int, int) {
	// debug.Debug("findLeft(pos %d, charSize %d) data %s", pos, charSize, data)

	var left []byte
	for {
		if (pos - charSize) < 0 {
			return nil, 0, 0
		}

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

	// debug.Debug("findLeft() left %s, charSize %d, leftCount %d", left, charSize, leftCount)
	return left, charSize, leftCount
}

func findRight(data []byte, pos int, charSize int) ([]byte, int, int) {
	// debug.Debug("findRight(pos %d, charSize %d) data: %s", pos, charSize, data)
	if len(data) < pos+charSize {
		return nil, 0, 0
	}
	var right []byte
	for {
		if (pos + charSize) > len(data) {
			return nil, 0, 0
		}
		right = data[pos : pos+charSize]
		if !utf8.Valid(right) {
			// not valid UTF
			charSize++
			continue
		}

		break
	}

	if len(data) < pos+charSize {
		return nil, 0, 0
	}

	// debug.Debug("searching with pos %d charsize %d in: %s", pos, charSize, data[pos+charSize:])

	rightCount := bytes.Count(data[pos+charSize:], right)
	if rightCount > 1 {
		subpos := pos + charSize
		realCount := 1
		for i := 0; i < rightCount*charSize; i = i + charSize {
			// debug.Debug("i: %d", i)
			subdata := data[subpos+i : subpos+i+charSize]
			// debug.Debug("looking in: %s", subdata)
			if bytes.Equal(subdata, right) {
				realCount++
			} else {
				break
			}
		}

		right = data[pos : pos+(charSize*realCount)]
		rightCount = realCount
	}

	if rightCount == 0 {
		rightCount = 1
	}

	return right, charSize, rightCount
}

func getCount(data []byte, enclosure Character) int {
	var doubleEncl []byte
	doubleEncl = append(doubleEncl, enclosure...)
	doubleEncl = append(doubleEncl, enclosure...)
	escaped := bytes.Count(data, doubleEncl)
	if escaped == -1 || escaped == 1 {
		escaped = 0
	}

	c := bytes.Count(data, enclosure)
	if len(enclosure) == 0 {
		c = 0
	}

	return c - escaped*2
}
