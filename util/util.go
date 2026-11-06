// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

func MultiToFlat(data [][]byte) (out []byte, pos []uint32) {
	var size int
	var count int
	for i := range data {
		size += len(data[i])
		count++
	}
	out = make([]byte, 0, size)
	pos = make([]uint32, 0, count)
	for i := range data {
		out = append(out, data[i]...)
		pos = append(pos, uint32(len(data[i])))
	}

	return out, pos
}

func FlatAppend(dst []byte, src []byte, pos []uint32) ([]byte, []uint32) {
	pos = append(pos, uint32(len(src)))
	dst = append(dst, src...)
	return dst, pos
}

func FlatToMulti(in []byte, pos []uint32) (data [][]byte) {
	// without the pre-allocation of the slice for the data buffer
	// Benchmark_FlatToMulti-8   	24116842	       496.5 ns/op	     272 B/op	      10 allocs/op
	// after pre-allocating it
	// Benchmark_FlatToMulti-8   	43638446	       333.8 ns/op	     184 B/op	       8 allocs/op
	data = make([][]byte, 0, len(pos))
	// var buf []byte
	var prev uint32 = 0
	for _, len := range pos {
		// before the pre-allocation of the buf slice it took 333ns
		// after pre-allocation it took 203ns
		// Benchmark_FlatToMulti-8   	75486668	       203.2 ns/op	     136 B/op	       4 allocs/op
		// buf = make([]byte, 0, len)
		// for i := prev; uint32(i) < prev+len; i++ {
		// 	buf = append(buf, in[i])
		// }

		// data = append(data, buf)
		// replacing that algo with the new one below
		// Benchmark_FlatToMulti-8   	129122768	        88.40 ns/op	      80 B/op	       1 allocs/op

		data = append(data, in[prev:prev+len])
		prev += len
	}

	return data
}

// CleanFunc returns the cleaned up byte slice. prefix and suffix
// signify the SkipCheck function should be called
type CleanFunc func(data []byte, prefix, suffix bool) []byte

func NoClean(data []byte, prefix, suffix bool) []byte {
	return data
}

type CharacterConfig struct {
	characters.Config

	// this is only used for the SplitCSVInput
	// zero index based
	DelimiterCount int

	clearedEOL bool
	eolBuf     characters.Character

	// signifies the file was produced by Excel
	// or behaves like Excel.
	// meaning the EOL contains an unused delimiter
	// which must be stripped before the rest of the
	// logic sees the line
	Excel bool

	// CSV file has been produced with PHP where the enclosure
	// is escaped using a backslash \" instead of ""
	PHP bool

	// used for passing the counted columns to the chunk
	DelimCounted        int `json:"-"`
	DelimOutsideCounted int `json:"-"`

	// counts the amount of times ND seperation would not work
	NdMoved int

	// tells the splitter to return lines on EOL match until false
	ReturnEarly bool

	// prevents re-detection of the EOL
	FixedEOL bool
}

func NewCharacterConfig(delimiter, enclosure characters.Character) *CharacterConfig {
	cc := CharacterConfig{}
	cc.Config = characters.NewConfig(delimiter, enclosure)

	return &cc
}

func (c *CharacterConfig) Clean(data []byte, prefix, suffix bool) []byte {
	if c.EnclosureClean {
		if prefix {
			data = TrimPrefixSkipCheck(data, c.Enclosure)
		} else {
			data = bytes.TrimPrefix(data, c.Enclosure)
		}
		if suffix {
			data = TrimSuffixSkipCheck(data, c.Enclosure)
		} else {
			data = bytes.TrimSuffix(data, c.Enclosure)
		}

		// we need to clean double quotes
		data = ReplaceAll(data, c.DoubleEnclosure, c.Enclosure)
	}

	data = ReplaceAll(data, characters.Character_CARRIAGE_RETURN_LINE_FEED, characters.Character_LINE_FEED)
	data = ReplaceAll(data, characters.Character_CARRIAGE_RETURN, characters.Character_LINE_FEED)

	return data
}

// Replace returns a copy of the slice s with the first n
// non-overlapping instances of old replaced by new.
// If old is empty, it matches at the beginning of the slice
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune slice.
// If n < 0, there is no limit on the number of replacements.
func Replace(s, old, new, t []byte, n int) []byte {
	m := 0
	if n != 0 {
		// Compute number of replacements.
		m = Count(s, old)
	}
	if m == 0 {
		// Just return a copy.
		// return append([]byte(nil), s...)
		// NOTE originally it returned a copy
		return s
		// t = s
		// return t
	}
	if n < 0 || m < n {
		n = m
	}

	// Apply replacements to buffer.
	size := len(s) + n*(len(new)-len(old))
	if len(t) < size {
		// log.Printf("resizing")
		t = make([]byte, size)
	}
	t = t[:size]
	w := 0
	start := 0
	for i := 0; i < n; i++ {
		j := start
		if len(old) == 0 {
			if i > 0 {
				_, wid := utf8.DecodeRune(s[start:])
				j += wid
			}
		} else {
			j += bytes.Index(s[start:], old)
		}
		w += copy(t[w:], s[start:j])
		w += copy(t[w:], new)
		start = j + len(old)
	}
	w += copy(t[w:], s[start:])
	return t[0:w]
}

func ByteEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if a[0] != b[0] {
		return false
	}

	if a[len(a)-1] != b[len(b)-1] {
		return false
	}

	return bytes.Equal(a, b)
}

func CountOutsideEnclosure(s []byte, sep, enclosure characters.Character) int {
	// Fast path out
	if len(sep) == 0 {
		return -1
	}

	sn := Count(s, sep)
	if sn == -1 {
		return sn
	}

	// if the enclosure is empty, use normal count
	if len(enclosure) == 0 {
		return sn
	}

	// if the enclosure does not appear inside the slice we can use a normal count
	if bytes.IndexByte(s, enclosure[0]) == -1 {
		return sn
	}
	c := Count(s, enclosure)
	// enclosure must be equal
	if c%2 != 0 {
		return 0
	}

	var m int
	var curPos int
	var lastPos int
	var enclCount int

	n := 0
	i := 0
	for i < sn {
		curPos = 0
		m = bytes.Index(s[lastPos:], sep)

		if m < 0 {
			break
		}

		curPos = m + lastPos

		// debug.Debug("checking for %s: \"%s\"", enclosure.Name(), s[:curPos:curPos])
		enclCount = bytes.Count(s[:curPos:curPos], enclosure)
		// debug.Debug("enclCount: %d", enclCount)
		if enclCount%2 == 0 {
			// debug.Debug("found: \"%s\"", s[:curPos:curPos])
			// data = append(data, s[:curPos:curPos])
			lastPos = 0
			i++

			// this was a column so the sep was valid
			n++
		} else {
			lastPos += m + len(sep)
			// debug.Debug("continue, lastPos: %d", lastPos)
			// do not set s
			continue
		}

		// debug.Debug("setting s: \"%s\"", s[curPos+len(sep):])
		s = s[curPos+len(sep):]
	}

	if debug.DebugLog {
		// debug.Debug("[CountOutsideEnclosure] leftover s: %s", s)
		// debug.Debug("[CountOutsideEnclosure] %d for sep %s and encl %s", n, sep.Name(), enclosure.Name())
	}

	return n
}

func FindEOLCharacter(in []byte) characters.Character {
	// there is a chance that the header is LF and the rest of the file is CRLF
	// the only way to catch that is by checking if the LF is nearby the CR, so we take the position
	// of both. If the index is only off by 1 we know the character is a CRLF and else it's just LF (or CR).
	lfIndex := bytes.Index(in, characters.Character_LINE_FEED)
	crIndex := bytes.Index(in, characters.Character_CARRIAGE_RETURN)

	if lfIndex != 0 && crIndex == lfIndex-1 {
		// the character is CRLF
		return characters.Character_CARRIAGE_RETURN_LINE_FEED
	}

	if lfIndex != -1 && lfIndex < crIndex {
		return characters.Character_LINE_FEED
	}

	if crIndex != -1 {
		return characters.Character_CARRIAGE_RETURN
	}
	if lfIndex != -1 {
		return characters.Character_LINE_FEED
	}

	return characters.Character_EMPTY
}

func (c *CharacterConfig) SplitCsvInput(in []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(in) == 0 && atEOF {
		return 0, nil, io.EOF
	}

	var eolIndex int
	// using a for because its nicer to break than have crazy if statements
	for len(c.EOL) == 0 {
		c.EOL = FindEOLCharacter(in)

		if len(c.EOL) == 0 && len(c.eolBuf) != 0 {
			c.EOL = c.eolBuf
			break
		}

		// not enough data
		if len(c.EOL) == 0 {
			return 0, nil, nil
		}
	}

	// counts how many times the chunk looped around
	var stuckCount int

	// end of the first line, before the index moves to a following EOL
	var firstEolIndex int

findEolIndex:
	eolIndex = bytes.Index(in, c.EOL)
	// not enough bytes
	if eolIndex == -1 {
		if atEOF {
			return len(in), in, nil
		}
		return 0, nil, nil
	}
	firstEolIndex = eolIndex

	if len(in[:eolIndex]) == 0 {
		return eolIndex + len(c.EOL), in[:eolIndex], nil
	}

	if len(c.Enclosure) == 0 && !c.EnclosureIsEmpty {
		// debug.Debug("c.Enclosure: +%s+ c.EnclosureIsEmpty: %t", c.Enclosure, c.EnclosureIsEmpty)
		c.DetectDelimiterAndEnclosureInChunk(in, eolIndex)
	}

	if c.ReturnEarly {
		eolLen := len(c.EOL)
		c.eolBuf = c.EOL
		c.EOL = nil
		if debug.DebugLog {
			debug.Debug("returning early: %s", in[:eolIndex])
		}
		return eolIndex + eolLen, in[:eolIndex], nil
	}

	if c.DelimiterCount == 0 {
		// the optional header always ends with a newline
		// HOWEVER it can be enclosured, check if the enclosure appears AND is not even
		if eolIndex != -1 {
			if len(c.Enclosure) != 0 && bytes.Contains(in[:eolIndex], c.Enclosure) && bytes.Count(in[:eolIndex], c.Enclosure)%2 != 0 {
				if debug.DebugLog {
					debug.Debug("header contains an unbalanced enclosure")
				}

				tmpEolIndex := eolIndex
				var tmpStart int
				tries := 50
				i := 0
				for {
					if i == tries {
						break
					}
					if len(in) < tmpEolIndex {
						break
					}

					if bytes.Count(in[:tmpEolIndex], c.Enclosure)%2 == 0 {
						if debug.DebugLog {
							debug.Debug("header found: %s", in[:tmpEolIndex])
						}
						eolIndex = tmpEolIndex
						break
					}

					if len(in) > tmpEolIndex+len(c.EOL) {
						tmpStart = tmpEolIndex + len(c.EOL)
						// debug.Debug("searching inside: %s", in[tmpStart:])
						tmpEolIndex = bytes.Index(in[tmpStart:], c.EOL)
						// we could be at the end of a file
						if tmpEolIndex == -1 {
							tmpEolIndex = len(in)
							i++
							continue
						}
						tmpEolIndex += tmpStart
					} else {
						break
					}
					i++
				}

			}

			// we have to return the found data PLUS the length of the eolBytes
			// debug.LogData("eolIndex != -1", in[:eolIndex], 40960)
			// on the first pass we want to clear the EOL
			if !c.clearedEOL {
				c.clearedEOL = true
				eolLen := len(c.EOL)
				c.eolBuf = c.EOL
				c.EOL = nil
				return eolIndex + eolLen, in[:eolIndex], nil
			}
			return eolIndex + len(c.EOL), in[:eolIndex], nil
		}
	}

	var enclosureCount int

	// we want to loop over the possible eolIndex
	for ; ; stuckCount++ {
		// count the delimiters
		// if the delimCount matches the c.DelimiterCount it COULD be a full row, it can also have
		// the delimiter inside a field, we need to check for that too

		// count the enclosures
		// NB: enclosures are OPTIONAL, meaning there e.g. could be six fields but only 2 enclosures

		// if the source was produced by PHP we must clean that up, check the chunk we are now looking at
		// in place, the data after eolIndex is still needed when the row continues on the next line;
		// a replacement that changes the length would shift every index, it is skipped
		if c.PHP && bytes.Contains(in[:eolIndex], c.EscapedEnclosure) {
			if replaced := c.ReplaceBackslashEscapedEnclosure(in[:eolIndex]); len(replaced) == eolIndex {
				copy(in, replaced)
			}
		}

		if debug.DebugLog {
			debug.Debug("[SplitCsvInput] eolIndex: %d, c.EOL: %x\tdelimiter count: %d\t"+
				"delimiters counted in buf: %d\tenclosures (%s) counted in buf: %d\t"+
				"CountOutsideEnclosure: n %d\tNdMoved: %d",
				eolIndex, c.EOL, c.DelimiterCount,
				bytes.Count(in[:eolIndex], c.Delimiter), c.Enclosure, bytes.Count(in[:eolIndex], c.Enclosure),
				CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure), c.NdMoved)
			debug.LogData("[SplitCsvInput] in[:eolIndex]", in[:eolIndex], 512)
			debug.BreakMe()
		}

		c.DelimCounted = bytes.Count(in[:eolIndex], c.Delimiter)

		// NOTE: the enclosure can still be set to none if this is the first time returning a chunk after detecting the delimiter
		// in case the enclosure is set to be none, we only have to test for the amount of delimiters
		if c.EnclosureLength() == 0 && c.DelimCounted == c.DelimiterCount {
			if debug.DebugLog {
				debug.LogData("[SplitCsvInput] encloLen == 0", in[:eolIndex], 4096)
				debug.BreakMe()
			}
			return eolIndex + len(c.EOL), in[:eolIndex], nil
		}

		// if the enclosure length is zero, it can only be newline delimited
		if c.EnclosureLength() == 0 && c.NdMoved == 0 {
			if debug.DebugLog {
				debug.LogData("[SplitCsvInput] encloLen == 0 AND NdMoved == 0", in[:eolIndex], 4096)
				debug.BreakMe()
			}
			return eolIndex + len(c.EOL), in[:eolIndex], nil
		}

		enclosureCount = bytes.Count(in[:eolIndex], c.Enclosure)

		if c.EnclosureLength() != 0 && enclosureCount == 0 && c.NdMoved == 0 {
			if debug.DebugLog {
				debug.LogData("[SplitCsvInput] encloLen != 0 AND c.Enclosure count == 0 AND NdMoved == 0", in[:eolIndex], 4096)
				debug.BreakMe()
			}
			return eolIndex + len(c.EOL), in[:eolIndex], nil
		}

		// we can skip over all of these are there is no enclosure to compare against
		if c.EnclosureLength() != 0 {
			// if delimCount is larger then DelimiterCount _AND_ encloCount is even
			// it COULD be a row, more tests will proof otherwise
			// a backslash before an enclosure only escapes it in PHP mode, in RFC 4180 data it
			// precedes the first enclosure of a doubled pair
			escapedCount := 0
			if c.PHP {
				escapedCount = bytes.Count(in[:eolIndex], c.EscapedEnclosure)
			}
			if c.DelimCounted == c.DelimiterCount && (enclosureCount-escapedCount)%2 == 0 {
				// test is not good enough, but the delimiters are equal, so should be correct
				// if c.DelimCounted == c.DelimiterCount {
				if debug.DebugLog {
					debug.Debug("[SplitCsvInput] delimCount: %d, c.DelimiterCount: %d", c.DelimCounted, c.DelimiterCount)
					debug.LogData("[SplitCsvInput] delimCount is larger then c.DelimiterCount _AND_ encloCount is even", in[:eolIndex], 4096)
					debug.BreakMe()
				}

				if bytes.Contains(in[:eolIndex], []byte(`{`)) {
					// massive slowdown but needed for files with embedded JSON
					if c.DelimOutsideCounted = CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure); c.DelimOutsideCounted != c.DelimiterCount {
						goto moveindex
					}
				}

				return eolIndex + len(c.EOL), in[:eolIndex], nil
				// }
			}

			// it is possible that the delimiter appears within the enclosure
			if c.DelimOutsideCounted = CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure); c.DelimOutsideCounted == c.DelimiterCount {
				if debug.DebugLog {
					debug.LogData("[SplitCsvInput]CountOutsideEnclosure is the same as the DelimiterCount", in[:eolIndex], 40960)
					debug.BreakMe()
				}
				return eolIndex + len(c.EOL), in[:eolIndex], nil
			}

			// it is possible that the enclosure is simply optional, unless the last field opens one:
			// that is a value with a line break in the last column
			if c.DelimiterCount > 1 && c.DelimCounted == c.DelimiterCount && enclosureCount%2 != 0 && c.NdMoved == 0 &&
				!c.openLastField(in[:eolIndex]) {
				if debug.DebugLog {
					debug.LogData("[SplitCsvInput] DelimCounted == DelimiterCount AND c.Enclosure count not even", in[:eolIndex], 4096)
					debug.BreakMe()
				}
				return eolIndex + len(c.EOL), in[:eolIndex], nil
			}

			// a full row with an open enclosure, and the next enclosure encloses a field of its own
			if eolIndex == firstEolIndex && !c.PHP && c.DelimCounted == c.DelimiterCount &&
				(enclosureCount-bytes.Count(in[:eolIndex], c.EscapedEnclosure))%2 != 0 &&
				c.strayEnclosure(in, eolIndex, atEOF) {
				if debug.DebugLog {
					debug.LogData("[SplitCsvInput] stray enclosure, the next enclosure encloses a field", in[:eolIndex], 4096)
					debug.BreakMe()
				}
				return eolIndex + len(c.EOL), in[:eolIndex], nil
			}

			// we cannot compare with the c.DelimCount as that number is ALWAYS higher than c.DelimOutsideCount
			// and previously, a locally scoped delimCount var was overwritten by the result from CountOutsideEnclosure
			if !c.Excel && c.DelimOutsideCounted-c.DelimiterCount == 1 {
				// a backslash escaped enclosure explains the extra delimiter as well, PHP mode goes first
				if !c.PHP && bytes.Contains(in[:eolIndex], c.EscapedEnclosure) {
					if debug.DebugLog {
						debug.Debug("[SplitCsvInput] PHP mode enabled instead of Excel mode")
					}
					c.PHP = true
					goto findEolIndex
				}
				if debug.DebugLog {
					debug.Debug("[SplitCsvInput] Excel mode enabled: c.DelimOutsideCounted (%d) - c.DelimiterCount (%d)", c.DelimOutsideCounted, c.DelimiterCount)
				}
				c.Excel = true
				c.DelimiterCount++
				continue
			}
		} else {
			// we cannot compare with the c.DelimCount as that number is ALWAYS higher than c.DelimOutsideCount
			// and previously, a locally scoped delimCount var was overwritten by the result from CountOutsideEnclosure
			if !c.Excel && c.DelimCounted-c.DelimiterCount == 1 {
				if debug.DebugLog {
					debug.Debug("[SplitCsvInput] Excel mode enabled: c.DelimCounted (%d) - c.DelimiterCount (%d)", c.DelimCounted, c.DelimiterCount)
				}
				c.Excel = true
				c.DelimiterCount++
				continue
			}
		}

		if bytes.Count(in[:eolIndex], c.Delimiter) >= (c.DelimiterCount * 2) {
			if stuckCount > 100 {
				fmt.Printf("Stuck (%d) on chunking: \n+++\n%s\n+++\n", stuckCount, in[:eolIndex])
				fmt.Printf("[SplitCsvInput] eolIndex: %d, c.EOL: %x\tdelimiter count: %d\t"+
					"delimiters counted in buf: %d\tenclosures (%s) counted in buf: %d\t"+
					"CountOutsideEnclosure: n %d\tNdMoved: %d\n",
					eolIndex, c.EOL, c.DelimiterCount,
					bytes.Count(in[:eolIndex], c.Delimiter), c.Enclosure, bytes.Count(in[:eolIndex], c.Enclosure),
					CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure), c.NdMoved)

				panic("stuck on chunking")
			}

			if !c.PHP && bytes.Contains(in[:eolIndex], c.EscapeCharacter) {
				// hail mary
				c.PHP = true
				// find the eolIndex again
				goto findEolIndex
			}
		}

	moveindex:
		tmpEOL := c.EOL
		if !c.FixedEOL {
			tmpEOL = FindEOLCharacter(in[eolIndex+len(c.EOL):])
			if len(tmpEOL) == 0 {
				tmpEOL = c.EOL
			}
		}

		// we found none, move the index
		if len(in) > eolIndex+len(tmpEOL) {
			if debug.DebugLog {
				debug.Debug("[SplitCsvInput] nothing found, moving EOL index")
			}
			// debug.LogData("current in, searching for new EOL", in[:eolIndex], 512)
			// debug.LogData("searching for new EOL", in[eolIndex+len(c.EOL):], 4096)

			newIndex := bytes.Index(in[eolIndex+len(c.EOL):], tmpEOL)
			if newIndex == -1 {
				break
			}

			// we have to add the length of eolBytes here to account for the offset
			eolIndex += newIndex + len(tmpEOL)
			if !c.FixedEOL && !bytes.Equal(c.EOL, tmpEOL) {
				eolIndex += len(c.EOL) - len(tmpEOL)
				c.EOL = tmpEOL
				if debug.DebugLog {
					debug.Debug("eol changed, c.EOL: %x, tmpEol: %x", c.EOL, tmpEOL)
					debug.Debug("increased eolIndex: %d", eolIndex)
				}
			}
			if debug.DebugLog {
				debug.LogData("[SplitCsvInput] new in", in[:eolIndex], 512)
				debug.BreakMe()
			}
			c.NdMoved++
			continue
		}

		break
	}

	c.DelimCounted = 0
	c.DelimOutsideCounted = 0

	if atEOF && len(in) != 0 {
		return len(in), in, nil
	}

	// not enough bytes
	return 0, nil, nil
}

// strayEnclosure reports whether the open enclosure on the line ending at eolIndex has no
// partner: the next enclosure starts a field and the enclosure after it ends that field,
// or at EOF no enclosure follows. It returns false while in holds too little to tell.
func (c *CharacterConfig) strayEnclosure(in []byte, eolIndex int, atEOF bool) bool {
	if len(in) < eolIndex+len(c.EOL) {
		return false
	}
	rest := in[eolIndex+len(c.EOL):]

	open := bytes.Index(rest, c.Enclosure)
	if open == -1 {
		return atEOF
	}
	if !c.startsField(rest[:open]) {
		return false
	}

	pos := open + len(c.Enclosure)
	for {
		i := bytes.Index(rest[pos:], c.Enclosure)
		if i == -1 {
			return false
		}
		pos += i + len(c.Enclosure)

		// a doubled enclosure is an escaped one inside the field
		if i != 0 && bytes.HasPrefix(rest[pos:], c.Enclosure) {
			pos += len(c.Enclosure)
			continue
		}

		if pos == len(rest) {
			return atEOF
		}
		return c.endsField(rest[pos:])
	}
}

// openLastField reports whether the field after the last delimiter of line starts with an
// enclosure that the line does not close.
func (c *CharacterConfig) openLastField(line []byte) bool {
	last := line
	if i := bytes.LastIndex(line, c.Delimiter); i != -1 {
		last = line[i+len(c.Delimiter):]
	}
	if !bytes.HasPrefix(last, c.Enclosure) {
		return false
	}
	return bytes.Count(last, c.Enclosure)%2 != 0
}

// startsField reports whether a field starts right after before.
func (c *CharacterConfig) startsField(before []byte) bool {
	if len(before) == 0 {
		return true
	}
	last := before[len(before)-1]
	return bytes.HasSuffix(before, c.Delimiter) || last == '\n' || last == '\r'
}

// endsField reports whether a field ends right before after.
func (c *CharacterConfig) endsField(after []byte) bool {
	return bytes.HasPrefix(after, c.Delimiter) || after[0] == '\n' || after[0] == '\r'
}

func (cc *CharacterConfig) DetectDelimiterAndEnclosureInChunk(in []byte, eolIndex int) {
	if len(cc.Delimiter) == 0 {
		cc.SetDelimiter(characters.DetectDelimiter(in[:eolIndex]))
	}

	tmpEolIndex := eolIndex
	var tmpStart int
	tries := 50
	i := 0
	for {
		if i == tries {
			break
		}
		if len(in) < tmpEolIndex {
			break
		}
		// debug.Debug("passing data: %s", in[tmpStart:tmpEolIndex])
		encl, enclDetectCount, occurence := characters.DetectEnclosureWithDelimiter(in[tmpStart:tmpEolIndex], cc.Delimiter)
		if enclDetectCount > 0 {
			cc.SetEnclosure(encl)
			if debug.DebugLog {
				debug.Debug("enclosure (%s) count: %d occurence: %d found in: \n+%s+", encl, enclDetectCount, occurence, in[tmpStart:tmpEolIndex])
				// panic("yo")
			}
			break
		}
		// maybe the header contains a newline inside an enclosure
		if i == 1 {
			if debug.DebugLog {
				debug.Debug("check enclosure on second line")
			}
			encl, enclDetectCount, _ := characters.DetectEnclosureWithDelimiter(in[0:tmpEolIndex], cc.Delimiter)
			if enclDetectCount > 0 {
				cc.SetEnclosure(encl)
				break
			}
		}
		if len(in) > tmpEolIndex+len(cc.EOL) {
			tmpStart = tmpEolIndex + len(cc.EOL)
			// debug.Debug("searching inside: %s", in[tmpStart:])
			tmpEolIndex = bytes.Index(in[tmpStart:], cc.EOL)
			// we could be at the end of a file
			if tmpEolIndex == -1 {
				tmpEolIndex = len(in)
				i++
				continue
			}
			tmpEolIndex += tmpStart
			// debug.Debug("found: %s", in[tmpStart:tmpEolIndex])
		} else {
			break
		}
		i++
	}
}

// // ReplaceBackslashEscapedEnclosure fixes PHP's proprietary backslash escaping of enclosures.
// // It only occurs with quotation marks and a backslash.
func (cc *CharacterConfig) ReplaceBackslashEscapedEnclosure(in []byte) (dst []byte) {
	if debug.DebugLog {
		debug.BreakMe()
	}

	if !bytes.Contains(in, cc.EscapedEnclosure) {
		return in
	}

	if len(cc.Enclosure) == 0 {
		return in
	}

	// we need to move through the data ensuring we don't replace data twice
	dst = make([]byte, 0, len(in))
	var start int
	for {
		// debug.Debug("searching in: %s", in[start:])
		if pos := bytes.Index(in[start:], cc.EscapedEnclosure); pos != -1 {
			dst = append(dst, in[start:start+pos]...)
			start += pos + len(cc.EscapedEnclosure)

			// we need to check if the next character is also an enclosure, but more crucially
			// if the character ONE after that is NOT an enclosure
			//
			// examples:
			// "foo \" bar",   needs to be replaced
			// "foo \"",       needs to be replaced
			// "foo \"" bar",  is fine, and we shouldn't touch it

			next := in[start : start+len(cc.Enclosure)]
			isEncl := bytes.Equal(next, cc.Enclosure)
			// if it IS an enclosure, we must check ONE character further if it is an delimiter
			if isEncl {
				// if the encl is at the end, e.g. \"" then we must also replace it
				if len(in) == start+len(cc.Enclosure) {
					dst = append(dst, cc.DoubleEnclosure...)
					dst = append(dst, cc.Enclosure...)
					start = len(in)
					break
				}

				// encl not at the end
				next = in[start+len(cc.Enclosure) : start+len(cc.Enclosure)+len(cc.Delimiter)]
				switch {
				case bytes.Equal(next, cc.Delimiter):
					// something like \"", was found
					dst = append(dst, cc.DoubleEnclosure...)
					dst = append(dst, cc.Enclosure...)
					dst = append(dst, cc.Delimiter...)
					start += len(cc.Enclosure) + len(cc.Delimiter)
				case bytes.Equal(next, characters.Character_LINE_FEED) ||
					bytes.Equal(next, characters.Character_CARRIAGE_RETURN):
					// something like \""\n was found
					dst = append(dst, cc.DoubleEnclosure...)
					dst = append(dst, cc.Enclosure...)
					dst = append(dst, next...)
					start += len(cc.Enclosure) + len(cc.Delimiter)
				default:
					// something like \"" was found, leave it
					dst = append(dst, cc.EscapedEnclosure...)
					dst = append(dst, cc.Enclosure...)
					// debug.Debug("writing: %s", next)
					dst = append(dst, next...)
					start += len(cc.Enclosure) + len(cc.Delimiter)
				}

				continue
			}

			if !isEncl {
				if bytes.Equal(next, cc.Delimiter) {
					// something like \", was found
					// NOTE: we would LIKE to replace this with just ", however
					// that would change the length of the outgoing byte slice
					// which causes the scanner to get unhappy (we would have to track it)
					// we need to decide if an empty space is a good middle ground solution.
					// we can only replace it with an invisible space but that is not a good solution.
					dst = append(dst, []byte(" ")...)
					dst = append(dst, cc.Enclosure...)
					dst = append(dst, cc.Delimiter...)
					start += len(cc.Enclosure)
					continue
				}
				// character is not an enclosure, we must replace it
				// something like \" was found
				dst = append(dst, cc.DoubleEnclosure...)
				// do not increment the start here with the len(cc.Enclosure) as it's already at that point
			}
		} else {
			if start == 0 {
				return in
			}
			break
		}
	}

	// append the remainder
	dst = append(dst, in[start:]...)

	return dst
}

// Generic split: splits after each instance of sep.
// Adapted to return a flat byte slice with pos slice
func GenSplit(s, sep []byte, n int, fn CleanFunc, exclude []int) (data []byte, pos []uint32) {
	return GenSplitCore(s, sep, n, fn, exclude, make([]byte, 0), make([]uint32, 0))
}

func GenSplitCore(s, sep []byte, n int, fn CleanFunc, exclude []int, data []byte, pos []uint32) ([]byte, []uint32) {
	if n == 0 {
		return nil, nil
	}

	if n < 0 {
		n = bytes.Count(s, sep) + 1
	}

	estimatedLength := len(s) - (n * len(sep))
	if estimatedLength == -1 {
		estimatedLength = 0
	}
	if cap(data) < estimatedLength {
		data = make([]byte, 0, estimatedLength)
	}
	if cap(pos) < n {
		pos = make([]uint32, 0, n)
	}

	n--
	i := 0
	for i < n {
		m := bytes.Index(s, sep)
		// buf = buf[:0]
		if m < 0 {
			break
		}
		if exclude != nil && slices.Contains(exclude, i) {
			s = s[m+len(sep):]
			i++
			continue
		}
		buf := s[:m:m]
		buf = fn(buf, false, false)
		data = append(data, buf...)
		pos = append(pos, uint32(len(buf)))
		s = s[m+len(sep):]

		i++
	}
	buf := s
	buf = fn(buf, false, false)
	data = append(data, buf...)
	pos = append(pos, uint32(len(buf)))

	return data, pos
}

// ColSplitFlat will split a byte slice with the rules for CSV kept in place.
// If a seperator is inside an enclosure it will not be used to split.
// Returns a flat byte slice with pos slice
func ColSplitFlat(s []byte, sep characters.Character, enclosure characters.Character, n int, fn CleanFunc) (data []byte, pos []uint32) {
	return ColSplitFlatCore(s, sep, enclosure, n, fn, make([]byte, 0), make([]uint32, 0))
}

func ColSplitFlatCore(s []byte, sep characters.Character, enclosure characters.Character, n int, fn CleanFunc, data []byte, pos []uint32) ([]byte, []uint32) {
	if len(enclosure) == 0 {
		return GenSplit(s, sep, n, fn, nil)
	}
	// debug.LogData("s", s, 512)
	// debug.Debug("sep: \t\t%s\nenclosure: \t%s\nn: \t\t\t%d", sep.Name(), enclosure.Name(), n)
	if n == 0 {
		return nil, nil
	}

	// we do not know yet how often the seperator appears inside enclosures
	// we need to account for that later
	if n < 0 {
		n = bytes.Count(s, sep) + 1
	}

	estimatedLength := len(s) - (n * len(sep))
	if cap(data) < estimatedLength {
		data = make([]byte, 0, estimatedLength)
	}
	if cap(pos) < n {
		pos = make([]uint32, 0, n)
	}

	var m int
	var lastPos int
	var count int
	var curPos int
	var buf []byte

	n--
	i := 0
	// var buf []byte
	for i < n {
		if debug.DebugLog {
			debug.BreakMe()
		}
		curPos = 0

		m = bytes.Index(s[lastPos:], sep)

		if m < 0 {
			break
		}

		curPos = m + lastPos

		// debug.Debug("checking: %s", s[:curPos:curPos])
		count = bytes.Count(s[:curPos:curPos], enclosure)
		if count%2 == 0 {
			if debug.DebugLog {
				debug.Debug("found: %s", s[:curPos:curPos])
			}
			buf = fn(s[:curPos:curPos], false, false)
			data = append(data, buf...)
			pos = append(pos, uint32(len(buf)))
			lastPos = 0
			i++
		} else {
			lastPos += m + len(sep)
			// debug.Debug("continue, lastPos: %d", lastPos)
			// do not set s
			continue
		}

		// debug.Debug("setting s: %s", s[curPos+len(sep):])

		// we are accidentally dropping the last column when it's empty
		if len(s[curPos:]) == len(sep) {
			if bytes.Equal(s[curPos:], sep) {
				data = append(data, []byte{}...)
				pos = append(pos, uint32(0))
				i++
			}
		}

		s = s[curPos+len(sep):]
	}

	if debug.DebugLog {
		debug.BreakMe()
	}

	if len(s) > 0 {
		buf = fn(s, false, false)
		data = append(data, buf...)
		pos = append(pos, uint32(len(buf)))
	}

	return data, pos
}

// ColSplit will split a byte slice with the rules for CSV kept in place.
// If a seperator is inside an enclosure it will not be used to split.
func ColSplit(s []byte, sep characters.Character, enclosure characters.Character, n int) (data [][]byte) {
	// debug.LogData("s", s, 512)
	// debug.Debug("sep: \t\t%s\nenclosure: \t%s\nn: \t\t\t%d", sep.Name(), enclosure.Name(), n)
	if n == 0 {
		return nil
	}

	// we do not know yet how often the seperator appears inside enclosures
	// we need to account for that later
	if n < 0 {
		n = bytes.Count(s, sep) + 1
	}

	var m int
	var lastPos int
	var count int
	var curPos int

	data = make([][]byte, 0, n)

	n--
	i := 0
	// var buf []byte
	for i < n {
		// if debug.DebugLog {
		// 	debug.BreakMe()
		// }
		curPos = 0

		// debug.Debug("i: %d n: %d", i, n)
		m = bytes.Index(s[lastPos:], sep)

		if m < 0 {
			break
		}

		curPos = m + lastPos

		// debug.Debug("checking for %s: \"%s\"", enclosure.Name(), s[:curPos:curPos])
		count = bytes.Count(s[:curPos:curPos], enclosure)
		// debug.Debug("count: %d", count)
		if count%2 == 0 {
			// debug.Debug("found: \"%s\"", s[:curPos:curPos])
			data = append(data, s[:curPos:curPos])
			lastPos = 0
			i++
		} else {
			lastPos += m + len(sep)
			// debug.Debug("continue, lastPos: %d", lastPos)
			// do not set s
			continue
		}

		// we are accidentally dropping the last column when it's empty
		if len(s[curPos:]) == len(sep) {
			if bytes.Equal(s[curPos:], sep) {
				data = append(data, []byte{})
				i++
			}
		}

		// debug.Debug("setting s: \"%s\"", s[curPos+len(sep):])
		s = s[curPos+len(sep):]
	}

	if len(s) > 0 {
		data = append(data, s)
	}

	return data
}

// Count calls bytes.IndexByte before Count
func Count(s, sep []byte) int {
	// special case where there is no delim
	if len(sep) == 0 {
		return 0
	}
	// if the first byte of sep doesn't appear the rest won't either
	if bytes.IndexByte(s, sep[0]) == -1 {
		return 0
	}

	return bytes.Count(s, sep)
}

// ReplaceAll calls bytes.Contains beforehand preventing alloc
func ReplaceAll(s, old, new []byte) []byte {
	if !bytes.Contains(s, old) {
		return s
	}

	return bytes.ReplaceAll(s, old, new)
}

// TrimSuffixSkipCheck skips the check if the suffix exists
func TrimSuffixSkipCheck(s, suffix []byte) []byte {
	return s[:len(s)-len(suffix)]
}

func TrimPrefixSkipCheck(s, prefix []byte) []byte {
	return s[len(prefix):]
}

type FlagUsage struct {
	Name      string
	Shorthand string
	Default   string
	Usage     string
}

func (f *FlagUsage) GetRow() string {
	var s strings.Builder

	s.WriteString("a| `--" + f.Name + "` ")
	if f.Shorthand != "" {
		s.WriteString("+ \n `-" + f.Shorthand + "`")
	}
	s.WriteString("\n")

	s.WriteString("| " + f.Default + "\n")
	s.WriteString("a| " + f.Usage + "\n")
	s.WriteString("\n")

	return s.String()
}

func (f *FlagUsage) GetMan() string {
	var s strings.Builder

	s.WriteString("*")
	if f.Shorthand != "" {
		s.WriteString("-" + f.Shorthand + ", ")
	}
	s.WriteString("--" + f.Name + "*:: \n")
	s.WriteString("  " + f.Usage + "\n\n")

	return s.String()
}

func Usage(flags *pflag.FlagSet) []FlagUsage {

	var fu []FlagUsage

	flags.VisitAll(func(flag *pflag.Flag) {
		u := FlagUsage{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Default:   flag.DefValue,
			Usage:     flag.Usage,
		}

		fu = append(fu, u)
	})

	return fu
}

type DelimiterCounts struct {
	Count        int
	OutsideCount int
}

func EncodeCounts(dst []byte, count, outside uint32) []byte {
	if cap(dst) < 8 {
		dst = make([]byte, 8)
	}
	dst = dst[:8]
	// cbc := make([]byte, 4)
	// obc := make([]byte, 4)

	binary.LittleEndian.PutUint32(dst[0:], count)
	binary.LittleEndian.PutUint32(dst[4:], outside)

	// return append(cbc, obc...)
	return dst
}

func DecodeCounts(v []byte) (count, outside int) {
	if len(v) > 8 {
		panic("v cannot be longer than 8")
	}

	return int(binary.LittleEndian.Uint32(v[:4])),
		int(binary.LittleEndian.Uint32(v[4:]))
}
