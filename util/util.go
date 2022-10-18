// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package util

import (
	"bytes"
	"io"
	"strings"

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

//export
type CharacterConfig struct {
	Delimiter       characters.Character
	Enclosure       characters.Character
	DoubleEnclosure characters.Character
	EnclosureClean  bool
	enclosureLenth  int
	EOL             characters.Character

	EscapeCharacter  characters.Character
	EscapedEnclosure characters.Character

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
}

func (cc *CharacterConfig) SetDelimiter(delim characters.Character) {
	cc.Delimiter = delim
}

func (cc *CharacterConfig) SetEnclosure(encl characters.Character) {
	// cc.Enclosure = encl
	cc.Enclosure = make(characters.Character, 0)
	cc.Enclosure = append(cc.Enclosure, encl...)
	cc.enclosureLenth = len(encl)

	// reset
	cc.DoubleEnclosure = make(characters.Character, 0)
	cc.DoubleEnclosure = append(cc.DoubleEnclosure, cc.Enclosure...)
	cc.DoubleEnclosure = append(cc.DoubleEnclosure, cc.Enclosure...)

	if len(cc.EscapeCharacter) == 0 {
		cc.EscapeCharacter = characters.Character("\\")
	}

	cc.EscapedEnclosure = make(characters.Character, 0)
	cc.EscapedEnclosure = append(cc.EscapedEnclosure, cc.EscapeCharacter...)
	cc.EscapedEnclosure = append(cc.EscapedEnclosure, cc.Enclosure...)

	// reset
	cc.EnclosureClean = true
	if bytes.Equal(cc.Enclosure, characters.Character_EMPTY) {
		cc.EnclosureClean = false
	}
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

func CountOutsideEnclosure(s []byte, sep, enclosure characters.Character) int {
	// Fast path out
	if len(sep) == 0 {
		return Count(s, sep)
	}
	sn := bytes.Count(s, sep)
	if sn == -1 {
		return sn
	}

	// if the enclosure does not appear inside the slice we can use a normal count
	c := Count(s, enclosure)
	if c == -1 {
		return Count(s, sep)
	}

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

	// debug.Debug("leftover s: %s", s)

	// debug.Debug("CountOutsideEnclosure: %d for sep %s and encl %s", n, sep.Name(), enclosure.Name())

	return n
}

func (c *CharacterConfig) SplitCsvInput(in []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(in) == 0 && atEOF {
		return 0, nil, io.EOF
	}

	var eolIndex int
	// using a for because its nicer to break than have crazy if statements
	for len(c.EOL) == 0 {
		// there is a chance that the header is LF and the rest of the file is CRLF
		// the only way to catch that is by checking if the LF is nearby the CR, so we take the position
		// of both. If the index is only off by 1 we know the character is a CRLF and else it's just LF (or CR).

		lfIndex := bytes.Index(in, characters.Character_LINE_FEED)
		crIndex := bytes.Index(in, characters.Character_CARRIAGE_RETURN)

		if crIndex == lfIndex-1 {
			// the character is CRLF
			c.EOL = characters.Character_CARRIAGE_RETURN_LINE_FEED
			break
		}

		if lfIndex < crIndex {
			c.EOL = characters.Character_LINE_FEED
			break
		}

		if crIndex != -1 {
			c.EOL = characters.Character_CARRIAGE_RETURN
			continue
		}
		if lfIndex != -1 {
			c.EOL = characters.Character_LINE_FEED
			break
		}

		if len(c.EOL) == 0 && len(c.eolBuf) != 0 {
			c.EOL = c.eolBuf
			break
		}

		// not enough data
		if len(c.EOL) == 0 {
			return 0, nil, nil
		}
	}

	eolIndex = bytes.Index(in, c.EOL)
	// not enough bytes
	if eolIndex == -1 {
		if atEOF {
			debug.Debug("atEOF")
			return len(in), in, nil
		}
		return 0, nil, nil
	}

	if len(c.Enclosure) == 0 {
		if len(c.Delimiter) == 0 {
			c.SetDelimiter(characters.DetectDelimiter(in[:eolIndex]))
		}

		tmpEolIndex := eolIndex
		var tmpStart int
		tries := 5
		i := 0
		for {
			if i == tries {
				break
			}
			if len(in) < tmpEolIndex {
				break
			}
			encl, enclDetectCount := characters.DetectEnclosureWithDelimiter(in[tmpStart:tmpEolIndex], c.Delimiter)
			// fmt.Printf("**ENCLOSURE** splitter found new enclosure: %s (%p)\n", encl, encl)
			if enclDetectCount > 0 {
				c.SetEnclosure(encl)
				break
			}
			if len(in) > tmpEolIndex+len(c.EOL) {
				// debug.LogData("searching for EOL", in[tmpStart:], 1024)
				search := bytes.Index(in[tmpStart:], c.EOL)
				if search == -1 {
					break
				}
				tmpStart = tmpEolIndex + len(c.EOL)
				tmpEolIndex += len(c.EOL) + search
			}
			i++
		}
	}
	if c.DelimiterCount == 0 {
		// the optional header always ends with a newline
		if eolIndex != -1 {
			// we have to return the found data PLUS the length of the eolBytes
			debug.LogData("eolIndex != -1", in[:eolIndex], 40960)
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

	// we want to loop over the possible eolIndex
	for {
		// count the delimiters
		// if the delimCount matches the c.DelimiterCount it COULD be a full row, it can also have
		// the delimiter inside a field, we need to check for that too

		// count the enclosures
		// NB: enclosures are OPTIONAL, meaning there e.g. could be six fields but only 2 enclosures

		if debug.DebugLog {
			debug.Debug("eolIndex: %d, c.EOL: \"%x\"\ndelimiter count: %d\ndelimiters counted in buf: %d\nenclosures counted in buf: %d\nCountOutsideEnclosure: n %d",
				eolIndex, c.EOL, c.DelimiterCount, bytes.Count(in[:eolIndex], c.Delimiter), bytes.Count(in[:eolIndex], c.Enclosure), CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure))
			debug.LogData("in[:eolIndex]", in[:eolIndex], 4096)
			debug.BreakMe()
		}

		// if the source was produced by PHP we must clean that up, check the chunk we are now looking at
		if c.PHP && bytes.Contains(in[:eolIndex], c.EscapedEnclosure) {
			in = c.ReplaceBackslashEscapedEnclosure(in)
		}

		delimCount := bytes.Count(in[:eolIndex], c.Delimiter)

		// NOTE: the enclosure can still be set to none if this is the first time returning a chunk after detecting the delimiter
		// in case the enclosure is set to be none, we only have to test for the amount of delimiters
		if c.enclosureLenth == 0 && delimCount == c.DelimiterCount {
			if debug.DebugLog {
				debug.LogData("encloLen == 0", in[:eolIndex], 40960)
				debug.BreakMe()
			}
			return eolIndex + len(c.EOL), in[:eolIndex], nil
		}

		if !c.Excel && delimCount-c.DelimiterCount == 1 {
			c.Excel = true
			c.DelimiterCount++
			continue
		}

		// we can skip over all of these are there is no enclosure to compare against
		if c.enclosureLenth != 0 {
			// if delimCount is larger then DelimiterCount _AND_ encloCount is even
			// it COULD be a row, more tests will proof otherwise
			if counted := bytes.Count(in[:eolIndex], c.Delimiter); counted >= c.DelimiterCount &&
				((bytes.Count(in[:eolIndex], c.Enclosure)-bytes.Count(in[:eolIndex], c.EscapedEnclosure))%2 == 0) {
				if debug.DebugLog {
					debug.Debug("delimCount: %d, c.DelimiterCount: %d", counted, c.DelimiterCount)
					debug.LogData("delimCount is larger then c.DelimiterCount _AND_ encloCount is even", in[:eolIndex], 40960)
					debug.BreakMe()
				}
				// test is not good enough, but the delimiters are equal, so should be correct
				if counted == c.DelimiterCount {
					return eolIndex + len(c.EOL), in[:eolIndex], nil
				}
			}

			// it is possible that the delimiter appears within the enclosure
			if CountOutsideEnclosure(in[:eolIndex], c.Delimiter, c.Enclosure) == c.DelimiterCount {
				if debug.DebugLog {
					debug.LogData("CountOutsideEnclosure is the same as the DelimiterCount", in[:eolIndex], 40960)
					debug.BreakMe()
				}
				return eolIndex + len(c.EOL), in[:eolIndex], nil
			}
		}

		// we found none, move the index
		if len(in) > eolIndex+len(c.EOL) {
			// debug.LogData("searching for new EOL", in[eolIndex+len(c.EOL):], 4096)
			newIndex := bytes.Index(in[eolIndex+len(c.EOL):], c.EOL)
			if newIndex == -1 {
				break
			}

			// we have to add the length of eolBytes here to account for the offset
			eolIndex += newIndex + len(c.EOL)
			continue
		}
		break
	}

	if atEOF && len(in) != 0 {
		return len(in), in, nil
	}

	// not enough bytes
	return 0, nil, nil
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
				start += len(cc.Enclosure)
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
	data = make([]byte, 0, estimatedLength)
	pos = make([]uint32, 0, n)

	n--
	i := 0
	for i < n {
		m := bytes.Index(s, sep)
		if m < 0 {
			break
		}
		if exclude != nil && slices.Contains(exclude, i) {
			s = s[m+len(sep):]
			i++
			continue
		}
		buf := fn(s[:m:m], false, false)
		data = append(data, buf...)
		// log.Printf("data: %s", data)
		pos = append(pos, uint32(len(buf)))
		s = s[m+len(sep):]
		i++
	}
	buf := fn(s, false, false)
	data = append(data, buf...)
	pos = append(pos, uint32(len(buf)))
	return data, pos
}

// ColSplitFlat will split a byte slice with the rules for CSV kept in place.
// If a seperator is inside an enclosure it will not be used to split.
// Returns a flat byte slice with pos slice
func ColSplitFlat(s []byte, sep characters.Character, enclosure characters.Character, n int, fn CleanFunc) (data []byte, pos []uint32) {
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
	data = make([]byte, 0, estimatedLength)
	pos = make([]uint32, 0, n)

	var m int
	var lastPos int
	var count int
	var curPos int
	var buf []byte

	n--
	i := 0
	// var buf []byte
	for i < n {
		// if debug.DebugLog {
		// 	debug.Debug("break me")
		// }
		curPos = 0

		m = bytes.Index(s[lastPos:], sep)

		if m < 0 {
			break
		}

		curPos = m + lastPos

		// debug.Debug("checking: %s", s[:curPos:curPos])
		count = bytes.Count(s[:curPos:curPos], enclosure)
		if count%2 == 0 {
			// debug.Debug("found: %s", s[:curPos:curPos])
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

	// if debug.DebugLog {
	// 	debug.Debug("break me")
	// }

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
		// 	debug.Debug("break me")
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
	if len(sep) == 1 {
		if bytes.IndexByte(s, sep[0]) == -1 {
			return 0
		}
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
