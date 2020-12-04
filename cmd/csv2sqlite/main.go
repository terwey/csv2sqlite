// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/parser"
	"github.com/spf13/pflag"
)

const debugLog = true

var (
	data        []byte
	flags       *pflag.FlagSet
	showVersion bool
	cpuprofile  string
	heapprofile string
	merge       bool

	delimiter string
	enclosure string

	p *parser.Params
	r *parser.Runtime

	routines int

	parseRoutines []*parser.Runtime

	// version is set during build
	version    = "undefined"
	minversion = "undefined"
)

func init() {
	p = parser.NewParameters()
	flags = pflag.NewFlagSet("csv2sqlite", pflag.ContinueOnError)
	flags.BoolVar(&showVersion, "version", false, "Prints the version")
	flags.StringVarP(&p.Input, "input", "i", "", "Input JSON file")
	flags.StringVarP(&p.Output, "output", "o", "", "Output CSV (or SQLite) file")
	flags.StringVarP(&delimiter, "delimiter", "d", "|", "Delimiter used")
	flags.StringVar(&enclosure, "enclosure", `"`, "Enclosure used")
	flags.BoolVar(&p.Sqlite, "sqlite", false, "Output SQLite file")
	flags.IntVarP(&p.FlushSize, "flush-rate", "r", 32768, "Flush rate for syncing to disk in kilobytes")
	flags.IntVarP(&p.Buffersize, "buffer-size", "b", 4096, "Buffer size for reading from disk in kilobytes")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Outputs a cpu-profile for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Outputs a heap-profile for pprof")
	flags.BoolVar(&merge, "merge-output", true, "Merges the output files")
	flags.IntVar(&routines, "threads", 4, "Threads to use when writing output")
	flags.SortFlags = false
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if showVersion {
		fmt.Printf("json2csv\nVersion: %s\nBuild date: %s (UTC)\n", version, minversion)
		os.Exit(0)
	}

	if len(p.Input) == 0 {
		fmt.Fprintf(os.Stderr, "Input file missing\n")
		flags.PrintDefaults()
		os.Exit(100)
	}

	if p.Input == p.Output {
		fmt.Fprintf(os.Stderr, "Input and Output are the same path\n")
		os.Exit(100)
	}

	p.Enclosure = []byte(enclosure)
	p.Delimiter = []byte(delimiter)

	// convert to kb
	p.FlushSize *= 1024
	p.Buffersize *= 1024

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

	start := time.Now()

	var err error
	// init the first runtimes, only one will scan
	r = parser.NewRuntime(p, bufio.ScanLines)
	var file io.ReadSeeker
	file, err = os.OpenFile(p.Input, os.O_RDONLY|syscall.MADV_DONTNEED, 0644)
	if err != nil {
		log.Fatalf("Cannot open file: %s", err)
	}

	if err = r.Init(file); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	headerSet := make(chan bool)
	go parse(r)

	wg.Add(1)
	go func(rows <-chan []byte, wg *sync.WaitGroup) {
		defer wg.Done()
		headerPassed := false
		for row := range rows {
			r.ProcessRow(r.ParseRow(row))
			if !headerPassed {
				if len(r.Header) != 0 {
					headerPassed = true
					for i := 1; i < routines; i++ {
						headerSet <- true
					}
				}
			}
		}
	}(r.Data, &wg)

	for i := 1; i < routines; i++ {
		pr := parser.NewRuntime(p, bufio.ScanLines)
		pr.Count = i
		if err = pr.Init(file); err != nil {
			log.Fatal(err)
		}
		parseRoutines = append(parseRoutines, pr)
	}

	for _, pr := range parseRoutines {
		wg.Add(1)
		go func(rows <-chan []byte, wg *sync.WaitGroup, routine *parser.Runtime) {
			if <-headerSet {
				routine.ProcessHeader(r.Header)

				for row := range rows {
					routine.ProcessRow(routine.ParseRow(row))
				}

				wg.Done()
			}
		}(r.Data, &wg, pr)
	}

	wg.Wait()
	r.Close()

	for _, pr := range parseRoutines {
		pr.Close()
		r.Rows += pr.Rows
	}

	elapsed := time.Since(start)
	var ops time.Duration
	var rows float64
	// r.Rows += r2.Rows
	if r.Rows != 0 {
		r.Rows++ // account for zero based index
		ops = time.Duration(elapsed.Nanoseconds() / int64(r.Rows))
		rows = float64(r.Rows) / elapsed.Seconds()
	} else {
		ops = time.Duration(0)
	}

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	log.Printf("Time: %s, Memory: %v MB, buffer-size: %d kb, flush-rate: %d kb, rows: %d, %s/op, rows %f/second, unescaped: %d",
		elapsed,
		m.Sys/1024/1024,
		p.Buffersize/1024,
		p.FlushSize/1024,
		r.Rows,
		ops,
		rows,
		delimBytes,
		r.RowsCSV,
	)

	if p.Sqlite && merge && routines > 1 {
		start = time.Now()
		var other []int
		for i := 1; i < routines; i++ {
			other = append(other, i)
		}
		r.Merge(other...)
		elapsed = time.Since(start)
		log.Printf("Merge Time: %s", elapsed)
	}
}

func logData(name string, data []byte, max int) {
	if debugLog {
		if len(data) < max {
			debug("%s (%d) data:\t %s", name, len(data), data)
			return
		}

		debug("%s (%d) start data[:%d]\t %s\nend data[len(data)-%d:] \t%s", name, len(data), max, data[:max], max, data[len(data)-max:])
	}
}

func debug(format string, v ...interface{}) {
	if debugLog {
		log.Printf(format, v...)
	}
}

func parse(r *parser.Runtime) {
	var err error
	scanner := r.GetScanner()
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
			close(r.Data)
		}

		r.ParseRow(scanner.Bytes())
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
		close(r.Data)
	}

	close(r.Data)
}
