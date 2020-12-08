// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package main

import (
	"bufio"
	"bytes"
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
)

var (
	data        []byte
	flags       *pflag.FlagSet
	showVersion bool
	cpuprofile  string
	heapprofile string
	merge       bool
	nd          bool

	headerLength = 0

	paramDelimiter string
	enclosure      string
	newline        string

	p *parser.Params
	r *parser.Runtime

	routines int

	parseRoutines []*parser.Runtime

	// version is set during build
	version    = "undefined"
	minversion = "undefined"

	debugLog = false
)

func init() {
	p = parser.NewParameters()
	flags = pflag.NewFlagSet("csv2sqlite", pflag.ContinueOnError)
	flags.BoolVar(&showVersion, "version", false, "Prints the version")
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
	flags.IntVar(&routines, "threads", 1, "Threads to use when writing output, default is set to one.")
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

	if len(p.Input) == 0 {
		fmt.Fprintf(os.Stderr, "Input file missing\n")
		flags.PrintDefaults()
		os.Exit(100)
	}

	if p.Input == p.Output {
		fmt.Fprintf(os.Stderr, "Input and Output are the same path\n")
		os.Exit(100)
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

	if p.Sqlite && merge && routines > 1 {
		start = time.Now()
		var other []int
		for i := 1; i < routines; i++ {
			other = append(other, i)
		}
		r.Merge(other...)
		mergeElapsed := time.Since(start)
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

func parse(r *parser.Runtime) {
	var err error
	var buf []byte
	delimSet := false
	scanner := r.GetScanner()
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
			close(r.Data)
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
		log.Fatal(err)
		close(r.Data)
	}

	close(r.Data)
}

var (
	// eolBytes   []byte = []byte("\r\n")
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
