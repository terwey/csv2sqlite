// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/parser"
	"github.com/terwey/csv2sqlite/delimiter"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"
)

var (
	data        []byte
	flags       *pflag.FlagSet
	showVersion bool
	cpuprofile  string
	heapprofile string
	merge       bool
	nd          bool
	stdin       bool

	headerLength = 0

	paramDelimiter string
	enclosure      string
	newline        string

	p *parser.Params
	r *parser.Runtime

	threads int

	parseRoutines []*parser.Runtime
	readyForMerge chan int
	stopWorkers   chan bool

	// version is set during build
	version    = "undefined"
	minversion = "undefined"

	debugLog = false
)

func init() {
	p = parser.NewParameters()
	flags = pflag.NewFlagSet("csv2sqlite", pflag.ContinueOnError)
	flags.BoolVar(&showVersion, "version", false, "Prints the version")
	flags.BoolVar(&stdin, "from-stdin", false, "Read from stdin and not from an input file")
	flags.StringVarP(&p.Input, "input", "i", "", "Input CSV file")
	flags.StringVarP(&p.Output, "output", "o", "", "Output CSV (or SQLite) file")
	flags.StringVarP(&paramDelimiter, "delimiter", "d", "", "Set the delimiter in case auto-detect should not be used. Must be one of: \n - "+strings.Join(delimiter.GetDelimiterNames(), "\n - "))
	flags.StringVar(&enclosure, "enclosure", `"`, "Enclosure used")
	flags.StringVar(&newline, "newline", "", "Newline used, either crlf or lf, else auto-detected")
	flags.BoolVar(&p.Sqlite, "sqlite", false, "Output SQLite file")
	flags.BoolVar(&nd, "nd", false, "Input CSV is newline delimited and does not have newlines in the middle of a field")
	flags.IntVarP(&p.FlushSize, "flush-rate", "r", 32768, "Flush rate for syncing to disk in kilobytes")
	flags.IntVarP(&p.Buffersize, "buffer-size", "b", 4096, "Buffer size for reading from disk in kilobytes")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Outputs a cpu-profile for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Outputs a heap-profile for pprof")
	flags.BoolVar(&merge, "merge-output", true, "Merges the output files")
	flags.IntVar(&threads, "threads", 1, "Threads to use when writing output, default is set to one.")
	flags.IntVar(&p.ChannelCapacity, "channel-capacity", 2000, "Buffered channels capacity")
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
		fmt.Printf("csv2sqlite\nVersion: %s\nBuild date: %s (UTC)\n", version, minversion)
		os.Exit(0)
	}

	if !stdin {
		if len(p.Input) == 0 {
			fmt.Fprintf(os.Stderr, "Input file missing\n")
			flags.PrintDefaults()
			os.Exit(100)
		}

		if p.Input == p.Output {
			fmt.Fprintf(os.Stderr, "Input and Output are the same path\n")
			os.Exit(100)
		}
	}

	if paramDelimiter != "" {
		delimBytes, err := delimiter.DelimiterFromString(paramDelimiter)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
			os.Exit(100)
		}
		p.Delimiter = delimBytes
	}

	readyForMerge = make(chan int, 1000)
	stopWorkers = make(chan bool)

	encloBytes = []byte(enclosure)
	p.Enclosure = encloBytes

	switch newline {
	case "lf":
		eolBytes = []byte("\n")
	case "crlf":
		eolBytes = []byte("\r\n")
	}

	// detect if the output is a sqlite
	if strings.HasSuffix(p.Output, ".sqlite") {
		p.Sqlite = true
	}

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
	if nd {
		r = parser.NewRuntime(p, bufio.ScanLines)
	} else {
		r = parser.NewRuntime(p, splitCsvInput)
	}
	if debugLog {
		r.Debug = true
	}
	var file io.ReadSeeker
	if !stdin {
		file, err = os.OpenFile(p.Input, os.O_RDONLY|syscall.MADV_DONTNEED, 0644)
		if err != nil {
			log.Fatalf("Cannot open file: %s", err)
		}
	} else {
		file = os.Stdin
	}

	if err = r.Init(file); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	// headerSet := make(chan bool)
	wg.Add(1)
	go parse(r, &wg, stopWorkers)

	wg.Add(1)
	go func(rows <-chan []byte, wg *sync.WaitGroup) {
		defer wg.Done()
		// headerPassed := false
		for row := range rows {
			r.ProcessRow(r.ParseRow(row))
			if threads > 1 {
				r.CloseDB()
				return
			}
		}
	}(r.Data, &wg)

	wg.Add(1)
	var mergeElapsed time.Duration
	go func(ready <-chan int, wg *sync.WaitGroup) {
		defer wg.Done()
		for done := range ready {
			startMerge := time.Now()
			// var other []int
			// for i := 1; i <= parserCount; i++ {
			// 	other = append(other, i)
			// }
			log.Printf("merging: %d", done)
			r.Merge(done)
			log.Printf("merged: %d", done)
			mergeElapsed += time.Since(startMerge)
		}
	}(readyForMerge, &wg)

	var parserCount int
	var threadWorkers sync.WaitGroup
	if threads > 1 && <-r.HeaderSet {

		// we use a semaphore to release/acquire threads so we can start and shutdown
		// the parsers as needed while merging
		// threads-1 because we need 1 thread for merging ;)
		sem := semaphore.NewWeighted(int64(threads - 1))
		ctx := context.Background()

	twLoop:
		for {
			select {
			case <-stopWorkers:
				break twLoop
			default:
				// if len(r.Data) > 1 {

				// log.Printf("Looping to create threads")
				// ctx, cancel := context.WithTimeout(ctx, time.Second)
				// defer cancel()
				if err := sem.Acquire(ctx, 1); err != nil {
					log.Printf("Cannot acquire semaphore")
					continue
				}

				// if len(r.Data) == 0 {
				// 	log.Printf("r.Data is empty")
				// 	break
				// }

				parserCount++
				// log.Printf("Creating parser for thread, count: %d", parserCount)
				pr := parser.NewParser(p, parserCount)
				if err = pr.Init(nil); err != nil {
					log.Fatal(err)
				}
				pr.ProcessHeader(r.Header)
				// prevent GC
				parseRoutines = append(parseRoutines, pr)

				threadWorkers.Add(1)
				go func(rows <-chan []byte, wg *sync.WaitGroup, sem *semaphore.Weighted,
					routine *parser.Runtime, readRuntime *parser.Runtime, ready chan int) {
					defer wg.Done()

					ticker := time.NewTicker(15 * time.Second)

				selectLoop:
					for {
						select {
						case row := <-rows:
							if len(row) == 0 {
								break selectLoop
							}
							routine.ProcessRow(routine.ParseRow(row))
						case <-ticker.C:
							break selectLoop
						}
					}

					readRuntime.Rows += pr.Rows
					err := pr.Close()
					if err != nil {
						log.Fatal(err)
					}

					ready <- pr.Count

					sem.Release(1)
					return

				}(r.Data, &threadWorkers, sem, pr, r, readyForMerge)
			}
		}
	}

	// for i := 1; i < threads; i++ {
	// 	pr := parser.NewRuntime(p, bufio.ScanLines)
	// 	pr.Count = i
	// 	if err = pr.Init(nil); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	parseRoutines = append(parseRoutines, pr)
	// }

	// for _, pr := range parseRoutines {
	// 	wg.Add(1)
	// 	go func(rows <-chan []byte, wg *sync.WaitGroup, routine *parser.Runtime) {
	// 		if <-headerSet {
	// 			routine.ProcessHeader(r.Header)

	// 			for row := range rows {
	// 				routine.ProcessRow(routine.ParseRow(row))
	// 			}

	// 			wg.Done()
	// 		}
	// 	}(r.Data, &wg, pr)
	// }

	threadWorkers.Wait()
	close(readyForMerge)
	wg.Wait()

	r.Close()

	// for _, pr := range parseRoutines {
	// 	pr.Close()
	// 	r.Rows += pr.Rows
	// }

	conversionElapsed := time.Since(start)

	var ops time.Duration
	var rows float64
	if r.Rows != 0 {
		ops = time.Duration(conversionElapsed.Nanoseconds() / int64(r.Rows))
		rows = float64(r.Rows) / conversionElapsed.Seconds()
	} else {
		ops = time.Duration(0)
	}

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	statsOut := make([]interface{}, 0)
	statsOut = append(statsOut,
		m.Sys/1024/1024,
		p.Buffersize/1024,
		p.FlushSize/1024,
		r.Rows,
		ops,
		rows,
		delimBytes,
		r.RowsCSV,
	)

	if p.Sqlite && merge && parserCount >= 1 {

		totalTime := conversionElapsed + mergeElapsed
		info := make([]interface{}, 0)
		info = append(info, totalTime, conversionElapsed, mergeElapsed)
		info = append(info, statsOut...)
		// log.Printf("Merge Time: %s", elapsed)
		log.Printf("Time: %s, Conversion time: %s, Merge Time: %s, Memory: %v MB, buffer-size: %d kb, flush-rate: %d kb, rows: %d, %s/op, rows %f/second, delimiter: %s, go-csv-reader: %d", info...)
	} else {
		info := make([]interface{}, 0)
		info = append(info, conversionElapsed)
		info = append(info, statsOut...)
		log.Printf("Time: %s, Memory: %v MB, buffer-size: %d kb, flush-rate: %d kb, rows: %d, %s/op, rows %f/second, delimiter: %s, go-csv-reader: %d", info...)
	}
}

func logData(name string, data []byte, max int) {
	if debugLog {
		if len(data) < max {
			debug("%s (%d) data:\n%s", name, len(data), data)
			return
		}

		debug("%s (%d) start data[:%d]\n\n%s\nend data[len(data)-%d:]\n\n%s", name, len(data), max, data[:max], max, data[len(data)-max:])
	}
}

func debug(format string, v ...interface{}) {
	if debugLog {
		log.Printf(format+"\n----\n", v...)
	}
}

func parse(r *parser.Runtime, wg *sync.WaitGroup, done chan bool) {
	defer wg.Done()
	var err error
	var buf []byte
	delimSet := false
	scanner := r.GetScanner()
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			close(r.Data)
			done <- true
			log.Fatal(err)
		}

		// if no delimiter was set we detect it once
		if !delimSet {
			delimBytes = delimiter.DetectDelimiter(scanner.Bytes())
			r.SetDelimiter(delimBytes)
			delimSet = true
		}

		// header has to be set once
		if headerLength == 0 {
			// we have to count the delimiters in the header to know for row scanning
			// how many fields we need
			headerLength = bytes.Count(scanner.Bytes(), delimBytes)
		}

		buf = make([]byte, len(scanner.Bytes()))
		copy(buf, scanner.Bytes())
		r.Data <- buf
	}

	if err = scanner.Err(); err != nil {
		close(r.Data)
		done <- true
		log.Fatal(err)
	}

	close(r.Data)
	done <- true
}

var (
	eolBytes   []byte
	delimBytes delimiter.Delimiter = delimiter.Delimiter_COMMA
	encloBytes []byte              = []byte("\"")
)

func splitCsvInput(in []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(in) == 0 && atEOF {
		return 0, nil, io.EOF
	}

	var eolIndex int
	if len(eolBytes) == 0 {
		if eolIndex = bytes.Index(in, []byte("\r\n")); eolIndex != -1 {
			eolBytes = []byte("\r\n")
		} else if eolIndex = bytes.Index(in, []byte("\n")); eolIndex != -1 {
			eolBytes = []byte("\n")
		}

		// not enough data
		if len(eolBytes) == 0 {
			return 0, nil, nil
		}
	}

	eolIndex = bytes.Index(in, eolBytes)
	// not enough bytes
	if eolIndex == -1 {
		if atEOF {
			return len(in), in, nil
		}
		return 0, nil, nil
	}

	if headerLength == 0 {
		// the optional header always ends with a newline
		if eolIndex != -1 {
			// we have to return the found data MINUS the length of the eolBytes
			return eolIndex + len(eolBytes), in[:eolIndex], nil
		}
	}

	// we want to loop over the possible eolIndex
	for {
		// count the delimiters
		// if the delimCount matches the headerLength it COULD be a full row, it can also have
		// the delimiter inside a field, we need to check for that too

		// count the enclosures
		// NB: enclosures are OPTIONAL, meaning there e.g. could be six fields but only 2 enclosures

		// if delimCount is larger then headerLength _AND_ encloCount is even
		// it COULD be a row, more tests will proof otherwise
		if bytes.Count(in[:eolIndex], delimBytes) >= headerLength && (bytes.Count(in[:eolIndex], encloBytes)%2 == 0) {
			return eolIndex + len(eolBytes), in[:eolIndex], nil
		}

		// we found none, move the index
		if len(in) > eolIndex+len(eolBytes) {
			newIndex := bytes.Index(in[eolIndex+len(eolBytes):], eolBytes)
			if newIndex == -1 {
				break
			}
			// we have to add the length of eolBytes here to account for the offset
			eolIndex += newIndex + len(eolBytes)
			continue
		}
		break
	}

	// not enough bytes
	return 0, nil, nil
}
