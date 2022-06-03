// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package util

import (
	"bytes"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
)

func MultiToFlat(data [][]byte) (out []byte, pos []uint32) {
	for i := range data {
		out = append(out, data[i]...)
		pos = append(pos, uint32(len(data[i])))
	}

	return out, pos
}

func FlatAppend(dst []byte, src []byte, pos []uint32) ([]byte, []uint32) {
	// log.Printf("len: %d, src: %s", uint32(len(src)), src)
	pos = append(pos, uint32(len(src)))
	dst = append(dst, src...)
	// log.Printf("pos: %v, dst: %s", pos, dst)
	return dst, pos
}

func FlatToMulti(in []byte, pos []uint32) (data [][]byte) {
	var buf []byte
	var prev uint32 = 0
	for _, len := range pos {
		// log.Printf("len: %d", len)
		buf = make([]byte, 0)
		for i := prev; uint32(i) < prev+len; i++ {
			buf = append(buf, in[i])
			// log.Printf("buf: %s", buf)
		}
		prev += len
		// log.Printf("apennding buf: %s", buf)
		data = append(data, buf)
	}
	// log.Printf("data: %s", data)
	return data
}

type CleanFunc func(data []byte) []byte

func NoClean(data []byte) []byte {
	return data
}

// Generic split: splits after each instance of sep.
// Adapted to return a flat byte slice with pos slice
func GenSplit(s, sep []byte, n int, fn CleanFunc) (data []byte, pos []uint32) {
	if n == 0 {
		return nil, nil
	}

	if n < 0 {
		n = bytes.Count(s, sep) + 1
	}

	data = make([]byte, 0)

	n--
	i := 0
	for i < n {
		m := bytes.Index(s, sep)
		if m < 0 {
			break
		}
		buf := fn(s[:m:m])
		data = append(data, buf...)
		// log.Printf("data: %s", data)
		pos = append(pos, uint32(len(buf)))
		s = s[m+len(sep):]
		i++
	}
	buf := fn(s)
	data = append(data, buf...)
	pos = append(pos, uint32(len(buf)))
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

	// if debug.DebugLog {
	// 	debug.Debug("break me")
	// }

	n--
	i := 0
	var buf []byte
	for i < n {
		m := bytes.Index(s, sep)
		debug.Debug("s (%d): %s", m, s)
		debug.Debug("existing buf: %s", buf)
		if m < 0 {
			// end of record
			if len(buf) > 0 {
				buf = append(buf, s...)
				data = append(data, buf)
				s = make([]byte, 0)
			}
			break
		}

		buf = append(buf, s[:m:m]...)
		debug.Debug("new buf: %s", buf)
		count := bytes.Count(buf, enclosure)
		if count%2 == 0 {
			data = append(data, buf)
			buf = make([]byte, 0)
			i++
		} else {
			buf = append(buf, sep...)
		}

		s = s[m+len(sep):]
	}

	if len(s) > 0 {
		data = append(data, s)
	}

	return data
}
