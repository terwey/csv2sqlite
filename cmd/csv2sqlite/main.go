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

	delimiter string
	enclosure string

	p  *parser.Params
	p2 *parser.Params
	r  *parser.Runtime
	r2 *parser.Runtime

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
	// init two runtimes, only one will scan
	r = parser.NewRuntime(p, bufio.ScanLines)
	r2 = parser.NewRuntime(p, bufio.ScanLines)
	r2.Count = 1
	var file io.ReadSeeker
	file, err = os.OpenFile(p.Input, os.O_RDONLY|syscall.MADV_DONTNEED, 0644)
	if err != nil {
		log.Fatalf("Cannot open file: %s", err)
	}
	if err = r.Init(file); err != nil {
		log.Fatal(err)
	}
	if err = r2.Init(file); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	second := make(chan bool)
	go parse(r)

	// second runtime
	wg.Add(1)
	go func(rows <-chan [][]byte, wg *sync.WaitGroup) {
		if <-second {
			log.Printf("second")
			r2.ProcessHeader(r2.Header)

			defer wg.Done()
			for cols := range rows {
				r2.ProcessRow(cols)
			}
		}
	}(r.Data, &wg)

	wg.Add(1)
	go func(rows <-chan [][]byte, wg *sync.WaitGroup, set chan bool) {
		defer wg.Done()
		headerPassed := false
		for cols := range rows {
			r.ProcessRow(cols)
			if !headerPassed {
				if len(r.Header) != 0 {
					r2.Header = r.Header

					log.Printf("before set")
					set <- true
					log.Printf("after set")
					headerPassed = true
				}
			}
		}
	}(r.Data, &wg, second)

	wg.Wait()
	wg.Wait()
	r.Close()
	r2.Close()

	elapsed := time.Since(start)
	var ops time.Duration
	var rows float64
	r.Rows += r2.Rows
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
		r.RowsCSV,
	)
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
