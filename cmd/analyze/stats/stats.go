// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package stats

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	gtin "github.com/godofdream/go-gtin"
)

type ColumnType int

const (
	ColumnType_UNDEFINED ColumnType = iota
	ColumnType_Text
	ColumnType_Number
	ColumnType_URL
	ColumnType_Image
	ColumnType_GTIN
	ColumnType_Empty
)

func (ct ColumnType) String() string {
	switch ct {
	case ColumnType_UNDEFINED:
		return "UNDEFINED"
	case ColumnType_Text:
		return "Text"
	case ColumnType_Number:
		return "Number"
	case ColumnType_URL:
		return "URL"
	case ColumnType_Image:
		return "Image"
	case ColumnType_GTIN:
		return "GTIN"
	case ColumnType_Empty:
		return "Empty"

	}

	return "UNDEFINED"
}

type ColumnInfo struct {
	Length     uint64
	ColumnType ColumnType

	// the position of the column in the name slice
	Pos int
}

func (r ColumnInfo) String() string {
	return fmt.Sprintf("Length: %d, Type: %s", r.Length, r.ColumnType)
}

type length uint64

type Statistics struct {
	// every column will have a Stat in the slice
	// the order is the same as the name slice
	data []*Stat
	Rows uint64
}

func NewStatistics(names []string) *Statistics {
	s := &Statistics{}
	for _, name := range names {
		s.data = append(s.data, &Stat{
			Name:               name,
			AverageLengthSum:   0,
			LengthDistribution: make(map[length]*uint64),
			ColumnTypes:        make(map[ColumnType]*uint64),
			EmptyValues:        0,
		})
	}
	return s
}

func (s *Statistics) String() string {
	var b strings.Builder
	for _, stat := range s.data {
		b.WriteString(stat.String())
		b.WriteString("\n")
	}
	return b.String()
}

func (s *Statistics) Add(ci *ColumnInfo) {
	s.data[ci.Pos].mutex.Lock()
	ld, ok := s.data[ci.Pos].LengthDistribution[length(ci.Length)]
	if !ok {
		s.data[ci.Pos].LengthDistribution[length(ci.Length)] = new(uint64)
		ld, ok = s.data[ci.Pos].LengthDistribution[length(ci.Length)]
	}
	// atomic.AddUint64(ld, 1)
	*ld++

	cto, ok := s.data[ci.Pos].ColumnTypes[ci.ColumnType]
	if !ok {
		s.data[ci.Pos].ColumnTypes[ci.ColumnType] = new(uint64)
		cto, _ = s.data[ci.Pos].ColumnTypes[ci.ColumnType]
	}
	// atomic.AddUint64(cto, 1)
	*cto++
	s.data[ci.Pos].mutex.Unlock()

	if ci.ColumnType == ColumnType_Empty {
		atomic.AddUint64(&s.data[ci.Pos].EmptyValues, 1)
	}

	atomic.AddUint64(&s.data[ci.Pos].AverageLengthSum, ci.Length)
}

func (s *Statistics) Merge(ins ...*Statistics) {
	for _, in := range ins {
		for pos := range s.data {
			s.data[pos].mutex.Lock()
			in.data[pos].mutex.Lock()

			for li := range in.data[pos].LengthDistribution {
				_, ok := s.data[pos].LengthDistribution[length(li)]
				if !ok {
					s.data[pos].LengthDistribution[length(li)] = new(uint64)
				}
				atomic.AddUint64(s.data[pos].LengthDistribution[length(li)], *in.data[pos].LengthDistribution[length(li)])
			}

			for cti := range in.data[pos].ColumnTypes {
				_, ok := s.data[pos].ColumnTypes[cti]
				if !ok {
					s.data[pos].ColumnTypes[cti] = new(uint64)
				}
				atomic.AddUint64(s.data[pos].ColumnTypes[cti], *in.data[pos].ColumnTypes[cti])
			}

			atomic.AddUint64(&s.data[pos].EmptyValues, in.data[pos].EmptyValues)

			atomic.AddUint64(&s.data[pos].AverageLengthSum, in.data[pos].AverageLengthSum)

			in.data[pos].mutex.Unlock()
			s.data[pos].mutex.Unlock()
		}
	}
}

func (s *Statistics) Calculate() {
	for i := range s.data {
		sum := s.data[i].AverageLengthSum
		s.data[i].AverageLength = float64(sum) / float64(s.Rows)
	}
}

// every Column will have a Stat struct
type Stat struct {
	Name string

	// contains all the lengths of all rows summed together
	AverageLengthSum uint64

	// the resulting average length
	AverageLength float64

	// for every length a counter is kept
	LengthDistribution map[length]*uint64

	// counter for every columntype occurance
	ColumnTypes map[ColumnType]*uint64

	// counter for empty values
	EmptyValues uint64

	// mutex for the maps
	mutex sync.Mutex
}

func (s *Stat) String() string {
	return fmt.Sprintf("[%s]: \tAverageLength: %f, Counts: (LengthDistribution: %d, ColumnType: %d, Empty: %d",
		s.Name, s.AverageLength, len(s.LengthDistribution), len(s.ColumnTypes), s.EmptyValues)
}

func DetectColumnType(v string) (ct ColumnType) {
	if strings.HasPrefix(v, "http://") ||
		strings.HasPrefix(v, "https://") {
		ct = ColumnType_URL
	}

	if strings.HasSuffix(v, "jpg") ||
		strings.HasSuffix(v, "jpeg") ||
		strings.HasSuffix(v, "png") ||
		strings.HasSuffix(v, "gif") ||
		strings.HasSuffix(v, "bmp") {
		ct = ColumnType_Image
		return
	}

	if ct == ColumnType_URL {
		return
	}

	if (strings.ContainsAny(v[0:1], "0123456789-") && strings.ContainsAny(v[len(v)-1:], "0123456789")) &&
		!strings.Contains(v, "/") {
		if gtin.IsGTIN(v) {
			ct = ColumnType_GTIN
			return
		}

		if strings.HasPrefix(v, "-") {
			_, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				ct = ColumnType_Number
				return
			}
		} else {
			_, err := strconv.ParseUint(v, 10, 64)
			if err == nil {
				ct = ColumnType_Number
				return
			}
		}

		// cannot have a float without a dot
		if strings.Contains(v, ".") {
			_, err := strconv.ParseFloat(v, 64)
			if err == nil {
				ct = ColumnType_Number
				return
			}
		}

		if len(v) > 1 && strings.Contains(v, ",") {
			defer func(v string) {
				if r := recover(); r != nil {
					fmt.Printf("exception on value: >> %s <<\n", v)
					fmt.Println("Recovered in f", r)
				}
			}(v)
			_, err := strconv.ParseFloat(strings.Replace(v, ",", ".", 1), 64)
			if err == nil {
				ct = ColumnType_Number
				return
			}
		}
	}

	// nothing left, must be text
	ct = ColumnType_Text

	return
}
