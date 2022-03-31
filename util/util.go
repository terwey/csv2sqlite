// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package util

import (
	"bytes"
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
