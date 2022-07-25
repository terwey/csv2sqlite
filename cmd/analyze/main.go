// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/productsupcom/go-sqlite-lite/sqlite3"
	"github.com/terwey/csv2sqlite/sqlitehelper"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"

	gtin "github.com/godofdream/go-gtin"
)

var (
	flags       *pflag.FlagSet
	input       string
	table       string
	write       bool
	cpuprofile  string
	heapprofile string
	max         int = 10000

	threads int = runtime.NumCPU()

	showProgress bool
	traceprofile string

	pv *progressbar.ProgressBar
)

func init() {
	flags = pflag.NewFlagSet("sqlitecompare", pflag.ContinueOnError)
	flags.SortFlags = false
	flags.StringVarP(&input, "file", "f", "", "SQLite file to analyze")
	flags.StringVarP(&table, "table", "t", "data", "Table to analyze")
	flags.BoolVar(&write, "write-results", false, "Write results to sqlite")
	flags.BoolVar(&showProgress, "progress", false, "Shows progress")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Filename for a cpu-profile ouput for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Filename for a heap-profile ouput for pprof")
	flags.StringVar(&traceprofile, "trace-profile", "", "Filename for a trace profile for go tool trace")
	flags.IntVar(&threads, "threads", threads, "Threads to use, default is set to the amount of cores")
	flags.IntVar(&max, "max", max, "Maximum amount of rows to fetch")

}

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
	rows uint64
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
	_, ok := s.data[ci.Pos].LengthDistribution[length(ci.Length)]
	if !ok {
		s.data[ci.Pos].LengthDistribution[length(ci.Length)] = new(uint64)
	}
	atomic.AddUint64(s.data[ci.Pos].LengthDistribution[length(ci.Length)], 1)

	_, ok = s.data[ci.Pos].ColumnTypes[ci.ColumnType]
	if !ok {
		s.data[ci.Pos].ColumnTypes[ci.ColumnType] = new(uint64)
	}
	atomic.AddUint64(s.data[ci.Pos].ColumnTypes[ci.ColumnType], 1)
	s.data[ci.Pos].mutex.Unlock()

	if ci.ColumnType == ColumnType_Empty {
		atomic.AddUint64(&s.data[ci.Pos].EmptyValues, 1)
	}

	atomic.AddUint64(&s.data[ci.Pos].AverageLengthSum, ci.Length)
}

func (s *Statistics) Calculate() {
	for i := range s.data {
		sum := s.data[i].AverageLengthSum
		s.data[i].AverageLength = float64(sum) / float64(s.rows)
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

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if heapprofile != "" {
		f, err := os.Create(heapprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.WriteHeapProfile(f)
	}

	if traceprofile != "" {
		f, err := os.Create(traceprofile)
		if err != nil {
			log.Fatal(err)
		}
		trace.Start(f)
		defer trace.Stop()
	}

	conn, err := sqlite3.Open(input)
	if err != nil {
		log.Fatalf("Cannot open %s: %s", input, err)
	}

	// sqlitehelper.ProgressEnabled = showProgress

	cols, err := sqlitehelper.ColumnsForDB(table, conn)
	if err != nil {
		log.Fatalf("ColumnsForDB err: %s", err)
	}

	colCount := cols.Count()
	names := cols.RawNames()

	stats := NewStatistics(names)

	count, err := sqlitehelper.Count(table, conn, true)
	if err != nil {
		log.Fatalf("Count: %s", err)
	}

	pv = progressbar.Default(int64(count), "Analyzing")
	if !showProgress {
		pv = progressbar.NewOptions(count, progressbar.OptionSetVisibility(false))
	}

	var wg sync.WaitGroup

	sem := semaphore.NewWeighted(int64(threads))
	ctx := context.Background()

	limit := 0
	offset := 0
	left := count

	var iterations int64 = 0

	for {
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Cannot acquire semaphore")
			continue
		}

		if left <= 0 {
			break
		}
		if left > max {
			limit = max
		}

		wg.Add(1)
		go func(stats *Statistics, w *sync.WaitGroup, sem *semaphore.Weighted, limit, offset int, pv *progressbar.ProgressBar, iter *int64) {
			conn, err := sqlite3.Open(input)
			if err != nil {
				log.Fatalf("Cannot open %s: %s", input, err)
			}
			defer conn.Close()
			// log.Printf("preparing: %d limit %d offset", limit, offset)
			stmt, err := conn.Prepare(sqlitehelper.QueryForIndexed([]string{"*"}, nil, table, limit, offset))
			if err != nil {
				log.Fatalf("Prepare: %s", err)
			}

			defer stmt.Close()

			for {
				more, err := stmt.Step()
				if err != nil {
					log.Fatalf("Step: %s", err)
				}

				if !more {
					break
				}

				for i := 0; i < colCount; i++ {
					val, _, err := stmt.ColumnRawString(i)
					if err != nil {
						log.Fatalf("ColumnRawString: %s", err)
					}

					stats.Add(analyzeColumn(val, i))
				}

				pv.Add(1)
				atomic.AddInt64(iter, 1)
			}

			w.Done()
			sem.Release(1)
		}(stats, &wg, sem, limit, offset, pv, &iterations)

		//           					   5591303 takes 0m28.078s
		// if atomic.LoadInt64(&iterations) > 200000 {
		// 	break
		// }

		left -= max
		offset += max
	}

	wg.Wait()

	// TODO: the count should be equal to iter
	stats.rows = uint64(iterations)
	stats.Calculate()
	log.Printf("stats: \n%s\n", stats)
}

func analyzeColumn(data interface{}, pos int) *ColumnInfo {
	r := &ColumnInfo{
		Pos: pos,
	}
	if data == nil {
		r.Length = 0
		r.ColumnType = ColumnType_Empty

		return r
	}

	if v, ok := data.(string); ok {
		r.Length = uint64(len(v))
		if r.Length == 0 {
			r.ColumnType = ColumnType_Empty
			return r
		} else {
			r.ColumnType = detectColumnType(v)
		}
	}

	if v, ok := data.(sqlite3.RawString); ok {
		r.Length = uint64(len(v))
		if r.Length == 0 {
			r.ColumnType = ColumnType_Empty
			return r
		} else {
			r.ColumnType = detectColumnType(string(v))
		}
	}

	return r
}

func detectColumnType(v string) (ct ColumnType) {
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

	if gtin.IsGTIN(v) {
		ct = ColumnType_GTIN
		return
	}

	if strings.ContainsAny(v, "0123456789") {
		_, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			ct = ColumnType_Number
			return
		}

		_, err = strconv.ParseUint(v, 10, 64)
		if err == nil {
			ct = ColumnType_Number
			return
		}

		_, err = strconv.ParseFloat(v, 64)
		if err == nil {
			ct = ColumnType_Number
			return
		}
	}

	// nothing left, must be text
	ct = ColumnType_Text

	return
}
