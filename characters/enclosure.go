// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package characters

import (
	"bytes"
	"sort"
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

func DetectEnclosureWithDelimiter(data []byte, delimiter Character) (Character, int, int) {
	if len(delimiter) == 0 {
		return nil, 0, 0
	}

	if len(data) == 0 {
		return nil, 0, 0
	}

	type Found struct {
		char         Character
		count        int
		occurence    int
		escapedCount int
	}

	detected := make(map[string]*Found)
	cols := bytes.Split(data, delimiter)
	var prevLeft []byte
	var prevLeftKnown []byte

	if debug.DebugLog {
		debug.BreakMe()
	}

	var left, right []byte
	charSize := 1
	for i := range cols {
		if len(cols[i]) == 0 {
			continue
		}
		if len(cols[i]) < charSize {
			continue
		}
		for {
			left = make([]byte, charSize)
			copy(left, cols[i][0:charSize])
			if !utf8.Valid(left) {
				// not valid UTF
				charSize++
				continue
			}

			break
		}

		if unicode.IsSpace(Character(left).Rune()) {
			left = make([]byte, 0)
		}

		if len(left) > 0 {
			// first check if the slice is maybe the same size as two times a char
			if len(cols[i]) != len(left)*2 {
				// check if left is a " and appears double in a JSON escaped
				if !bytes.Contains(data, []byte(`{""`)) {
					// check if the char appears multiple times
					if doubleLeft := append(left, left...); bytes.HasPrefix(cols[i], doubleLeft) {
						charSizeBuf := charSize
						leftBuf := append([]byte{}, left...)
						left = doubleLeft
						charSize *= 2
						if tripleLeft := append(doubleLeft, leftBuf...); bytes.HasPrefix(cols[i], tripleLeft) {
							left = tripleLeft
							charSize = charSizeBuf * 3
						}
					}
				}
			}
		} else {
			charSize = len(prevLeft)
		}

		if len(cols[i]) <= charSize {
			if debug.DebugLog {
				debug.Debug("[DetectEnclosureWithDelimiter] cols[i] same size (%d) as left (%s) charSize (%d), skipping right detect", len(cols[i]), left, len(left))
			}
			continue
		}

		// find the same on the right
		right = cols[i][len(cols[i])-charSize : len(cols[i])]
		if debug.DebugLog {
			debug.Debug("[DetectEnclosureWithDelimiter] found left = +%s+ right = +%s+", left, right)
		}

		if bytes.Equal(prevLeft, right) {
			if debug.DebugLog {
				debug.Debug("[DetectEnclosureWithDelimiter] prevLeft = +%s+ was a match for right = +%s+", prevLeft, right)
			}
			left = make([]byte, len(prevLeft))
			copy(left, prevLeft)
		}

		if bytes.Equal(prevLeftKnown, right) {
			if debug.DebugLog {
				debug.Debug("[DetectEnclosureWithDelimiter] prevLeftKnown = +%s+ was a match for right = +%s+", prevLeft, right)
			}
			left = make([]byte, len(prevLeftKnown))
			copy(left, prevLeftKnown)
		}

		if len(left) > 0 && bytes.Equal(left, right) {
			if debug.DebugLog {
				debug.Debug("[DetectEnclosureWithDelimiter] match left and right = +%s+", left)
			}

			// additional test to see if the data just occurs multiple times in the same column
			// if so it CANNOT be an enclosure as it's NOT escaped
			if bytes.Count(bytes.ReplaceAll(cols[i], append(left, left...), []byte("DOUBLE_QUOTE")), left) > 2 {
				continue
			}

			f, ok := detected[string(left)]
			if !ok {
				if debug.DebugLog {
					debug.BreakMe()
				}
				count := bytes.Count(bytes.ReplaceAll(data, append(left, left...), []byte("DOUBLE_QUOTE")), left)
				escapedCount := bytes.Count(bytes.ReplaceAll(data, append(left, left...), []byte("DOUBLE_QUOTE")), []byte("DOUBLE_QUOTE"))
				f = &Found{
					char:         Character(left),
					count:        count,
					escapedCount: escapedCount,
				}
				detected[string(left)] = f
			}
			f.occurence++
			f.occurence++
		}

		if len(left) > 0 {
			prevLeft = make([]byte, len(left))
			copy(prevLeft, left)

			if isKnownEnclosure(left) {
				if debug.DebugLog {
					debug.Debug("[DetectEnclosureWithDelimiter] adding known: %s", left)
				}
				prevLeftKnown = make([]byte, len(left))
				copy(prevLeftKnown, left)
			}
		}

		charSize = 1
	}

	if debug.DebugLog {
		for key := range detected {
			debug.Debug("[DetectEnclosureWithDelimiter] detected: (%s) count: %d, occurence: %d, escapedCount: %d", key, detected[key].count, detected[key].occurence, detected[key].escapedCount)
			// fmt.Printf("detected: (%s) count: %d, occurence: %d\n", key, detected[key].count, detected[key].occurence)
		}
	}

	// remove escaped enclosure
	for key := range detected {
		buf := make([]byte, len(detected[key].char))
		copy(buf, detected[key].char)
		escaped := make([]byte, len(buf)*2)
		escaped = append(escaped, buf...)
		escaped = append(escaped, buf...)
		delete(detected, string(escaped))
		escaped = append(escaped, buf...)
		delete(detected, string(escaped))
	}

	if debug.DebugLog {
		debug.BreakMe()
	}

	highest := Found{
		char:         Character_EMPTY,
		count:        -1,
		occurence:    -1,
		escapedCount: -1,
	}

	var sorted []*Found
	for _, found := range detected {
		sorted = append(sorted, found)
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		if isKnownEnclosure(sorted[i].char) {
			if isKnownEnclosure(sorted[j].char) {
				return sorted[i].occurence > sorted[j].occurence
			}
			return true
		}

		if isKnownEnclosure(sorted[j].char) {
			if isKnownEnclosure(sorted[i].char) {
				return sorted[i].occurence > sorted[j].occurence
			}

			// return false to keep it on the same line
			return false
		}

		return sorted[i].occurence > sorted[j].occurence
	})

	if debug.DebugLog {
		for key := range sorted {
			debug.Debug("[DetectEnclosureWithDelimiter] sorted detected: (%s) count: %d, occurence: %d, escapedCount: %d", sorted[key].char, sorted[key].count, sorted[key].occurence, sorted[key].escapedCount)
			// fmt.Printf("detected: (%s) count: %d, occurence: %d\n", key, detected[key].count, detected[key].occurence)
		}
	}

	if debug.DebugLog {
		debug.BreakMe()
	}

	for _, found := range sorted {
		if bytes.Equal(found.char, []byte(" ")) {
			continue
		}
		if bytes.Equal(found.char, []byte("\n")) {
			continue
		}
		if bytes.Equal(found.char, []byte(".")) {
			continue
		}
		if found.count%2 != 0 {
			continue
		}
		// chances of it being longer than 2 4-byte characters is rare
		if len(found.char) > 8 {
			continue
		}
		if highest.escapedCount <= found.escapedCount && highest.occurence < found.occurence && !unicode.IsLetter(found.char.Rune()) && !unicode.IsNumber(found.char.Rune()) && !isKnownEnclosure(highest.char) {
			highest = *found
		}
		if !isKnownEnclosure(highest.char) && highest.occurence >= found.occurence && isKnownEnclosure(found.char) {
			// debug.Debug("overriding because of knownenclosure: %s (%d)", found.char, len(found.char))
			highest = *found
		}
	}

	if highest.occurence == 1 {
		return Character_EMPTY, -1, -1
	}

	if highest.count%2 != 0 {
		return Character_EMPTY, -1, -1
	}

	if highest.occurence%2 != 0 {
		return Character_EMPTY, -1, -1
	}

	if debug.DebugLog {
		debug.BreakMe()
		debug.Debug("[DetectEnclosureWithDelimiter] enclosure found: %s", highest.char)
	}

	return append(Character{}, highest.char...), highest.count, highest.occurence
}

func findLeftOfPos(data []byte, pos int, charSize int) ([]byte, int, int) {
	// if debug.DebugLog {
	// 	debug.Debug("findLeftOfPos(pos %d, charSize %d) data %s", pos, charSize, data)
	// }

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

	// debug.Debug("findLeftOfPos() left %s, charSize %d, leftCount %d", left, charSize, leftCount)
	return left, charSize, leftCount
}

func findRightOfPos(data []byte, pos int, charSize int) ([]byte, int, int) {
	if debug.DebugLog {
		debug.Debug("findRightOfPos(pos %d, charSize %d) data: %s", pos, charSize, data)
	}
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
