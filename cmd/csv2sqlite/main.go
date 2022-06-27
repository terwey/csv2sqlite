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
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/parser"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"
)

var (
	data        []byte
	flags       *pflag.FlagSet
	showVersion bool

	cpuprofile   string
	heapprofile  string
	traceprofile string

	nd    bool
	stdin bool = true

	literalDelimiter string
	literalEnclosure string

	paramDelimiter string
	paramEnclosure string
	newline        string

	p *parser.Params
	r *parser.Runtime

	showProgress bool

	pv      *progressbar.ProgressBar
	threads int = runtime.NumCPU()

	parseRoutines []*parser.Runtime
	readyForMerge chan string
	stopWorkers   chan bool
	inserted      chan uint64
	insertedRows  uint64
	stopCount     int
	logHeader     int

	chunksize int

	// version is set during build
	version    = "undefined"
	minversion = "undefined"

	overflow = false
)

func init() {
	p = parser.NewParameters()
	flags = pflag.NewFlagSet("csv2sqlite", pflag.ContinueOnError)
	flags.BoolVar(&showVersion, "version", false, "Prints the version")
	flags.StringVarP(&p.Input, "input", "i", "", "Input CSV file")
	flags.StringVarP(&p.Output, "output", "o", "", "Output CSV (or SQLite) file")
	flags.StringVarP(&paramDelimiter, "delimiter", "d", "", "Set the delimiter in case auto-detect should not be used. Can be one of: \n - "+strings.Join(characters.GetAllCharacterNames(), "\n - "))
	flags.StringVar(&paramEnclosure, "enclosure", "", "Set the enclosure in case auto-detect should not be used, can be one of: \n - "+strings.Join(characters.GetAllCharacterNames(), "\n - ")+"\n")

	flags.StringVar(&literalDelimiter, "literal-delimiter", "", "Sets the delimiter to use to a literal string, if set the --delimiter is ignored")
	flags.StringVar(&literalEnclosure, "literal-enclosure", "", "Sets the enclosure to use to a literal string, if set the --enclosure is ignored")

	flags.IntSliceVar(&p.Exclude, "exclude", nil, "Column indices to exclude from conversion")
	flags.StringVar(&newline, "newline", "", "Set the newline to be used in case auto-detection should not be used. Can be one of: \n- LF\n- CR\n- CRLF")
	flags.BoolVar(&p.Sqlite, "sqlite", false, "Output SQLite file")
	flags.BoolVar(&nd, "nd", false, "Input CSV is newline delimited and does not have newlines in the middle of a field")
	flags.IntVar(&stopCount, "break-count", 15000, "Causes the sqlite writer to stop writing at the specified amount of rows.\nCan cause the merge to be triggered earlier. The first writer stops at N/3")
	flags.IntVarP(&p.CommitRate, "commit-rate", "r", 5000, "Flush rate for comitting rows to disk")
	flags.IntVarP(&p.Buffersize, "buffer-size", "b", 4096, "Buffer size for reading from disk in kilobytes")
	flags.IntVar(&threads, "threads", threads, "Threads to use when writing output, default is set to the amount of cores")
	flags.IntVar(&p.ChannelCapacity, "channel-capacity", 8000, "Buffered channels capacity multiplied by threads")
	flags.IntVar(&chunksize, "chunk-size", 10, "Chunksize before pushing the data to the channel")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Filename for a cpu-profile ouput for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Filename for a heap-profile ouput for pprof")
	flags.StringVar(&traceprofile, "trace-profile", "", "Filename for a trace profile for go tool trace")
	flags.IntVar(&logHeader, "log-header", 0, "Logs the header for the collected data, pass 1 to output it. Can be used together with seq")
	flags.BoolVar(&p.Drop, "drop", false, "Drops all outgoing data, for perf testing disk reading")
	flags.BoolVar(&showProgress, "progress", false, "Shows progress")

	if debug.State() {
		flags.BoolVar(&debug.DebugLog, "debug", false, "Outputs debug information, careful is extremely chatty.")
		flags.BoolVar(&debug.LogCSVReader, "log-csv-reader", false, "Logs the occurence and data where the Go built-in CSV Reader is used. Can be used without the debug mode.")

	}
	flags.BoolVar(&overflow, "overflow", false, "Overflows the data. Will abort processing.")
	flags.SortFlags = false
}

func main() {
	start := time.Now()
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if showVersion {
		fmt.Printf("csv2sqlite\nVersion: %s\nBuild date: %s (UTC)\n", version, minversion)
		fmt.Printf("debug version: %v\n", debug.State())
		os.Exit(0)
	}

	// running the code from VSCode through Delve now somehow sets stdin to be there without
	// containing data. we must prevent it from entering
	if len(p.Input) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			if len(p.Input) == 0 {
				fmt.Fprintf(os.Stderr, "Input file missing\n")
				flags.PrintDefaults()
				os.Exit(100)
			}

			if p.Input == p.Output {
				fmt.Fprintf(os.Stderr, "Input and Output are the same path\n")
				os.Exit(100)
			}

			stdin = false
		}
	} else {
		stdin = false
	}

	literald := false
	literale := false

	flags.Visit(func(f *pflag.Flag) {
		if f.Name == "literal-delimiter" {
			literald = true
		}
		if f.Name == "literal-enclosure" {
			literale = true
		}
	})

	if literald {
		p.Characters.SetDelimiter(characters.Character(literalDelimiter))
	} else {
		if paramDelimiter != "" {
			delim, err := characters.CharacterFromString(paramDelimiter)
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error()+"\n")
				flags.PrintDefaults()
				os.Exit(100)
			}
			p.Characters.SetDelimiter(delim)
		}
	}

	if literale {
		p.Characters.SetEnclosure(characters.Character(literalEnclosure))
	}
	if paramEnclosure != `` {
		encl, err := characters.CharacterFromString(paramEnclosure)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
			os.Exit(100)
		}
		p.Characters.SetEnclosure(encl)
	}

	// multiple channel capacity by threads
	p.ChannelCapacity = p.ChannelCapacity * threads

	if threads < 2 {
		log.Printf("Threads is set below 2, changing to 2")
		threads = 2
	}

	pv = progressbar.DefaultBytes(-1, "Processing")

	readyForMerge = make(chan string, 1000)
	stopWorkers = make(chan bool, threads-1)
	inserted = make(chan uint64)

	switch strings.ToUpper(newline) {
	case "LF":
		p.Characters.EOL = characters.Character_LINE_FEED
	case "CRLF":
		p.Characters.EOL = characters.Character_CARRIAGE_RETURN_LINE_FEED
	case "CR":
		p.Characters.EOL = characters.Character_CARRIAGE_RETURN
	}

	// detect if the output is a sqlite
	if strings.HasSuffix(p.Output, ".sqlite") {
		p.Sqlite = true
	}

	// convert to kb
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

	if traceprofile != "" {
		f, err := os.Create(traceprofile)
		if err != nil {
			log.Fatal(err)
		}
		trace.Start(f)
		defer trace.Stop()
	}

	var conversionElapsed time.Duration

	parser.DebugLog = debug.DebugLog
	if overflow {
		parser.Overflow = true
	}

	var err error
	// init the first runtimes, only one will scan
	if nd {
		r = parser.NewRuntime(p, bufio.ScanLines)
	} else {
		r = parser.NewRuntime(p, p.Characters.SplitCsvInput)
	}

	pvOptions := []progressbar.Option{
		progressbar.OptionSetDescription("Reading"),
	}

	if !showProgress {
		pvOptions = append(pvOptions, progressbar.OptionSetVisibility(false))
	}

	var file io.ReadSeeker
	if !stdin {
		f, err := os.OpenFile(p.Input, os.O_RDONLY|syscall.MADV_DONTNEED, 0644)
		if err != nil {
			log.Fatalf("Cannot open file: %s", err)
		}
		stat, err := f.Stat()
		if err != nil {
			log.Fatalf("Cannot stat file: %s", err)
		}

		pv = progressbar.NewOptions64(stat.Size(), pvOptions...)
		file = f
	} else {
		pv = progressbar.NewOptions64(-1, pvOptions...)
		file = os.Stdin
	}

	if err = r.Init(file); err != nil {
		log.Fatal(err)
	}

	var insertWg sync.WaitGroup
	var mergeWg sync.WaitGroup

	insertWg.Add(1)
	go func(inserted <-chan uint64, wg *sync.WaitGroup) {
		defer wg.Done()
		for rows := range inserted {
			insertedRows += rows
		}
	}(inserted, &insertWg)

	// readWg is used for the main scanner loop and also for the header initialization
	var readWg sync.WaitGroup
	readWg.Add(1)
	go parse(r, &readWg, chunksize, pv)

	var parserCount int
	var threadWorkers sync.WaitGroup
	mergeElapsed := make(chan time.Duration, 1)
	conversionStart := time.Now()
	if <-r.HeaderSet {
		mergeWg.Add(1)
		go r.Merge(readyForMerge, &mergeWg, mergeElapsed)

		// we use a semaphore to release/acquire threads so we can start and shutdown
		// the parsers as needed while merging
		// threads-1 because we need 1 thread for merging ;)
		sem := semaphore.NewWeighted(int64(threads - 1))
		ctx := context.Background()

		stop := stopCount

	twLoop:
		for {
			select {
			case <-stopWorkers:
				threadWorkers.Wait()
				close(inserted)
				close(readyForMerge)
				break twLoop
			default:
				if err := sem.Acquire(ctx, 1); err != nil {
					log.Printf("Cannot acquire semaphore")
					continue
				}

				// to prevent from starting a new thread when it's not needed
				// check if r.data is empty and just go back to the select
				if len(r.Data) == 0 {
					time.Sleep(5 * time.Millisecond)
					sem.Release(1)
					continue
				}

				stop = stopCount
				// start the merge early
				if parserCount == 0 {
					stop = 1000
				} else if parserCount < threads-1 {
					stop = stopCount / 3
				}

				parserCount++
				pr := parser.NewParser(p, parserCount)
				if err = pr.Init(nil); err != nil {
					log.Fatal(err)
				}
				pr.ProcessHeader(r.Header)
				// prevent GC
				parseRoutines = append(parseRoutines, pr)

				threadWorkers.Add(1)
				go func(chunks <-chan [][]byte, tw *sync.WaitGroup, sem *semaphore.Weighted,
					routine *parser.Runtime, readRuntime *parser.Runtime, ready chan string, stop int, inserted chan uint64) {

					signal := false
					rowCount := 0

					for {
						chunk, more := <-chunks

						for i := range chunk {
							routine.ProcessRow(routine.ParseRow(chunk[i]))
							rowCount++
						}

						if !more {
							signal = true
							break
						}

						if rowCount > stop {
							break
						}
					}

					inserted <- routine.Rows
					err := routine.Close()
					if err != nil {
						log.Fatal(err)
					}

					defer sem.Release(1)
					defer tw.Done()
					ready <- routine.GetPath()
					if signal {
						stopWorkers <- true
					}
					return
				}(r.Data, &threadWorkers, sem, pr, r, readyForMerge, stop, inserted)
			}
		}
	}

	conversionElapsed = time.Since(conversionStart)

	readWg.Wait()
	insertWg.Wait()
	mergeWg.Wait()
	r.Close()

	totalTime := time.Since(start)

	var ops time.Duration
	var rows float64
	if insertedRows != 0 {
		ops = time.Duration(totalTime.Nanoseconds() / int64(insertedRows))
		rows = float64(insertedRows) / totalTime.Seconds()
	} else {
		ops = time.Duration(0)
	}

	pv.Clear()

	if logHeader == 1 {
		fmt.Fprintf(os.Stderr, "Total, Conversion, Merge, rows, op/s, rows/s, threads, enclosure, delimiter, chunksize, buffer-size, commit-rate, break-count, channel-capacity, go csv reader, capacity\n")
	}
	merged := <-mergeElapsed
	fmt.Fprintf(os.Stderr, "%s, %s, %s, %d, %s, %.f, %d, %s, %s, %d, %d, %d, %d, %d, %d, %d\n",
		totalTime.Round(time.Millisecond),
		conversionElapsed.Round(time.Millisecond),
		merged.Round(time.Millisecond),
		insertedRows,
		ops.Round(time.Microsecond),
		math.Floor(rows),
		threads,
		p.Characters.Enclosure.Name(),
		p.Characters.Delimiter.Name(),
		chunksize,
		p.Buffersize/1024,
		p.CommitRate,
		stopCount,
		p.ChannelCapacity,
		r.RowsCSV,
		chunksize*p.ChannelCapacity*threads,
	)
}

func parse(r *parser.Runtime, wg *sync.WaitGroup, chunkCount int, pv *progressbar.ProgressBar) {
	var err error
	// var buf []byte
	var delimSet bool
	var enclosureSet bool
	if p.Characters.Delimiter != nil {
		delimSet = true
	}
	if p.Characters.Enclosure != nil {
		enclosureSet = true
	}
	scanner := r.GetScanner()

	var chunks [][]byte
	chunks = make([][]byte, 0, chunkCount)
	var chunk []byte

	count := 0
	var chunksSize int
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			close(r.Data)
			wg.Done()
			log.Fatal(err)
		}

		size := len(scanner.Bytes())
		chunk = make([]byte, size)
		copy(chunk, scanner.Bytes())
		chunksSize += size
		// if no delimiter was set we detect it once
		if !delimSet {
			p.Characters.SetDelimiter(characters.DetectDelimiter(chunk))
			delimSet = true
		}

		if !enclosureSet {
			p.Characters.SetEnclosure(characters.DetectEnclosure(chunk))
			enclosureSet = true
		}

		// header has to be set once
		if p.Characters.DelimiterCount == 0 {
			// we have to count the delimiters in the header to know for row scanning
			// how many fields we need
			p.Characters.DelimiterCount = bytes.Count(chunk, p.Characters.Delimiter)
		}

		// we have to process a header once
		if count == 0 {
			r.ParseRow(chunk)
			count++
			continue
		}

		if len(chunks) >= chunkCount {
			r.Data <- chunks
			pv.Add(chunksSize)
			chunksSize = 0
			chunks = make([][]byte, 0, chunkCount)
		}

		chunks = append(chunks, chunk)
	}

	if err = scanner.Err(); err != nil {
		close(r.Data)
		wg.Done()
		log.Fatal(err)
	}

	// send over the last chunks
	r.Data <- chunks

	close(r.Data)
	pv.Finish()
	wg.Done()
}
