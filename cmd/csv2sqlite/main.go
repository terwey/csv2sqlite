// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package main

/*
#cgo CFLAGS: -DSQLITE_MAX_COLUMN=32767
*/

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/felixge/fgprof"
	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/parser"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"

	"github.com/klauspost/cpuid/v2"
)

var (
	flags       *pflag.FlagSet
	showVersion bool

	cpuprofile   string
	heapprofile  string
	traceprofile string
	fgprofile    string

	stdin bool = true

	literalDelimiter string
	literalEnclosure string

	paramDelimiter string
	paramEnclosure string
	newline        string

	p = parser.NewParameters()
	r *parser.Runtime

	showProgress bool

	pv      *progressbar.ProgressBar
	threads int = cpuid.CPU.PhysicalCores

	parseRoutines []*parser.Runtime
	readyForMerge chan string
	rowCountChan  chan parser.RowCount
	rowCountTotal parser.RowCount
	stopCount     int
	logHeader     int
	csvReport     bool

	chunksize int

	// version is set during build
	version    = "undefined"
	minversion = "undefined"

	overflow = false

	quiet = false
)

func getFlags() *pflag.FlagSet {
	flags = pflag.NewFlagSet("csv2sqlite", pflag.ContinueOnError)
	flags.BoolVar(&showVersion, "version", false, "Prints the version")
	flags.StringVarP(&p.Input, "input", "i", "", "Input CSV file")
	flags.StringVarP(&p.Output, "output", "o", "", "Output CSV (or SQLite) file")
	flags.StringVarP(&paramDelimiter, "delimiter", "d", "", "Set the delimiter in case auto-detect should not be used. Can be one of: \n\n - "+strings.Join(characters.GetAllCharacterNames(), "\n - "))
	flags.StringVarP(&paramEnclosure, "enclosure", "e", "", "Set the enclosure in case auto-detect should not be used, can be one of: \n\n - "+strings.Join(characters.GetAllCharacterNames(), "\n - ")+"\n")

	flags.IntVar(&p.Start, "start", 1, "The starting position where the header is, useful for CSVs produced by Excel.")
	flags.IntVar(&p.Skip, "skip", -1, "Rows to skip after the header")
	flags.IntVar(&p.Limit, "limit", -1, "Amount of rows to process")

	flags.StringVar(&literalDelimiter, "literal-delimiter", "", "Sets the delimiter to use to a literal string, if set the --delimiter is ignored")
	flags.StringVar(&literalEnclosure, "literal-enclosure", "", "Sets the enclosure to use to a literal string, if set the --enclosure is ignored")

	flags.BoolVar(&p.Characters.PHP, "php", false, "CSV file was produced by PHP, enclosure escaped by backslash")

	flags.IntSliceVar(&p.Exclude, "exclude", nil, "Column indices to exclude from conversion")
	flags.StringVar(&newline, "newline", "", "Set the newline to be used in case auto-detection should not be used. Can be one of: \n\n- LF\n- CR\n- CRLF")
	flags.BoolVar(&p.Sqlite, "sqlite", false, "Output SQLite file")
	flags.BoolVar(&p.NewlineDelimited, "nd", false, "Input CSV is newline delimited and does not have newlines in the middle of a field")
	flags.IntVar(&stopCount, "break-count", 15000, "Causes the sqlite writer to stop writing at the specified amount of rows.\nCan cause the merge to be triggered earlier. The first writer stops at N/3")
	flags.IntVarP(&p.CommitRate, "commit-rate", "r", 5000, "Flush rate for comitting rows to disk")
	flags.IntVarP(&p.Buffersize, "buffer-size", "b", 15625, "Buffer size for reading from disk in kilobytes")
	flags.IntVar(&threads, "threads", threads, "Threads to use when writing output, default is set to the amount of cores")
	flags.IntVar(&p.ChannelCapacity, "channel-capacity", 25, "Buffered channels capacity multiplied by threads")
	flags.IntVar(&chunksize, "chunk-size", 10, "Chunksize before pushing the data to the channel")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Filename for a cpu-profile ouput for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Filename for a heap-profile ouput for pprof")
	flags.StringVar(&traceprofile, "trace-profile", "", "Filename for a trace profile for go tool trace")
	flags.StringVar(&fgprofile, "fgprofile", "", "Filename for a fgprofile ouput for pprof, see https://github.com/felixge/fgprof")
	flags.IntVar(&logHeader, "log-header", 0, "Logs the header for the collected data, pass 1 to output it. Can be used together with seq")
	flags.BoolVarP(&quiet, "quiet", "q", false, "Quiet mode do not output performance data")
	flags.BoolVar(&csvReport, "csv-report", false, "Performance data output in TSV format, default is JSON")

	flags.BoolVar(&p.Characters.FixedEOL, "fixed-eol", false, "EOL character (CR/CRLF/LF) doesn't change after auto-detect")
	flags.BoolVar(&p.Drop, "drop", false, "Drops all outgoing data, for perf testing disk reading")
	flags.BoolVar(&p.RowID, "rowid", false, "When the order of the CSV file must be kept the same in the resulting SQLite, this flag will \n"+
		"cause each row in the SQLite to receive a RowID for SQLite.")
	flags.BoolVar(&showProgress, "progress", false, "Shows progress")

	flags.StringVar(&p.Scratch, "scratch-dir", "", "Directory for storing the smaller SQLite files before merging them to the target. Same as output if left blank.")

	flags.IntVar(&p.CSVEnclosureSet, "csv-enclosure", 8, "When to use an enclosure for the CSV output: \n\n1\tNone\n2\tRequired\n4\tSkipEmpty\n8\tAll\n")

	if debug.State() {
		flags.BoolVar(&debug.DebugLog, "debug", false, "Outputs debug information, careful is extremely chatty.")
		flags.BoolVar(&parser.RWDebugLog, "rw-debug", false, "Outputs record writer debug information, careful is extremely chatty.")
		flags.BoolVar(&debug.LogCSVReader, "log-csv-reader", false, "Logs the occurence and data where the Go built-in CSV Reader is used. Can be used without the debug mode.")
		flags.BoolVar(&debug.EnableLogDropped, "log-dropped", false, "Outputs dropped data")

	}
	flags.BoolVar(&overflow, "overflow", false, "Overflows the data. Will abort processing.")
	flags.SortFlags = false

	return flags
}

func main() {
	flags = getFlags()

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

			stdin = false
		}
	} else {
		stdin = false
	}

	if p.Input != "" && samePath(p.Input, p.Output) {
		fmt.Fprintf(os.Stderr, "Input and Output are the same path\n")
		os.Exit(100)
	}

	if cpuprofile != "" && cpuprofile == heapprofile {
		fmt.Fprintf(os.Stderr, "cpu-profile and heap-profile are the same path\n")
		os.Exit(100)
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
		// NIL is another name for EMPTY, CharacterFromString does not know it
		if strings.EqualFold(paramEnclosure, "NIL") {
			paramEnclosure = "EMPTY"
		}
		encl, err := characters.CharacterFromString(paramEnclosure)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
			os.Exit(100)
		}
		p.Characters.SetEnclosure(encl)
		if strings.EqualFold(paramEnclosure, "EMPTY") {
			p.Characters.EnclosureIsEmpty = true
			p.NewlineDelimited = true
		}
	}

	if p.Characters.Delimiter.Length() > 1 && len(p.Characters.Enclosure) == 0 {
		p.Characters.EnclosureIsEmpty = true
		p.NewlineDelimited = true
	}
	pv = progressbar.DefaultBytes(-1, "Processing")

	readyForMerge = make(chan string, 1000)
	rowCountChan = make(chan parser.RowCount)

	if newline != "" {
		switch strings.ToUpper(newline) {
		case "LF":
			p.Characters.EOL = characters.Character_LINE_FEED
		case "CRLF":
			p.Characters.EOL = characters.Character_CARRIAGE_RETURN_LINE_FEED
		case "CR":
			p.Characters.EOL = characters.Character_CARRIAGE_RETURN
		}
	}

	// detect if the output is a sqlite
	if strings.HasSuffix(p.Output, ".sqlite") || strings.HasSuffix(p.Output, ".db") {
		p.Sqlite = true
	}

	// one thread is reserved for merging, parsing needs at least one more
	if threads < 2 {
		log.Printf("Threads is set below 2, changing to 2")
		threads = 2
	}

	// output is stdout
	if len(p.Output) == 0 && !p.Drop {
		threads = 2
	}

	// multiple channel capacity by threads
	p.ChannelCapacity = p.ChannelCapacity * threads

	// convert to kb
	p.Buffersize *= 1024

	if fgprofile != "" {
		f, err := os.Create(fgprofile)
		if err != nil {
			log.Fatalf("Could not create fgprofile: %s", err)
		}
		stop := fgprof.Start(f, fgprof.FormatPprof)
		defer stop()
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("Could not create cpuprofile: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if heapprofile != "" {
		f, err := os.Create(heapprofile)
		if err != nil {
			log.Fatalf("Could not create heapprofile: %s", err)
		}
		runtime.MemProfileRate = 1
		defer pprof.WriteHeapProfile(f)
	}

	if traceprofile != "" {
		f, err := os.Create(traceprofile)
		if err != nil {
			log.Fatalf("Could not create trace: %s", err)
		}
		trace.Start(f)
		defer trace.Stop()
	}

	var conversionElapsed time.Duration

	if overflow {
		parser.Overflow = true
	}

	// init the first runtimes, only one will scan
	if p.NewlineDelimited {
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

	// the previous output is only removed once the input is open
	if p.Output != "" {
		if err := os.Remove(p.Output); err != nil && !os.IsNotExist(err) {
			log.Fatalf("Cannot remove previous output: %s", err)
		}
	}

	var mergeWg sync.WaitGroup
	r.Init(file, parser.NewBuffers())

	// readWg is used for the main scanner loop and also for the header initialization
	var readWg sync.WaitGroup
	readWg.Add(1)
	go r.Parse(&readWg, chunksize, pv)

	var merged time.Duration
	var parserCount int
	var threadWorkers sync.WaitGroup
	mergeElapsed := make(chan time.Duration, 1)
	conversionStart := time.Now()
	var insertWg sync.WaitGroup
	if <-r.HeaderSet {
		// the header has been split without a rowid column, the rows carry it as their first column
		if p.RowID {
			for i := range p.Exclude {
				p.Exclude[i]++
			}
		}

		// close the original file if not sqlite
		if !p.Sqlite {
			r.Close()
		}

		insertWg.Add(1)
		go func(rowCountChan <-chan parser.RowCount, wg *sync.WaitGroup) {
			defer wg.Done()
			for rc := range rowCountChan {
				rowCountTotal.Inserted += rc.Inserted
				rowCountTotal.GoParser += rc.GoParser
				rowCountTotal.GoParserDropped += rc.GoParserDropped
				rowCountTotal.Dropped += rc.Dropped
			}
		}(rowCountChan, &insertWg)

		mergeWg.Add(1)
		go r.Merge(readyForMerge, &mergeWg, mergeElapsed)

		// we use a semaphore to release/acquire threads so we can start and shutdown
		// the parsers as needed while merging
		// threads-1 because we need 1 thread for merging ;)
		sem := semaphore.NewWeighted(int64(threads - 1))
		ctx := context.Background()

		var stop int

		bufferPool := sync.Pool{
			New: func() any {
				return parser.NewBuffers()
			},
		}

	twLoop:
		for {
			select {
			case <-r.StopWorkers:
				threadWorkers.Wait()
				close(rowCountChan)
				close(readyForMerge)
				mergeWg.Wait()
				merged = <-mergeElapsed
				break twLoop
			default:
				if err := sem.Acquire(ctx, 1); err != nil {
					log.Printf("Cannot acquire semaphore")
					continue
				}

				if parserCount == 0 {
					// to prevent from starting a new thread when it's not needed
					// check if r.data is empty and just go back to the select
					if len(r.Data) == 0 {
						time.Sleep(50 * time.Microsecond)
						sem.Release(1)
						continue
					}
				}

				stop = stopCount
				// start the merge early
				if parserCount == 0 {
					stop = 100
				} else if parserCount < threads-1 {
					stop = stopCount / 3
				}

				parserCount++
				pr := parser.NewParser(p, parserCount)
				pr.Init(nil, bufferPool.Get().(*parser.Buffers))

				pr.ProcessedHeader = r.ProcessedHeader
				pr.ProcessHeader(r.Header)
				// prevent GC
				parseRoutines = append(parseRoutines, pr)

				threadWorkers.Add(1)
				go func(chunks <-chan [][]byte, tw *sync.WaitGroup, sem *semaphore.Weighted,
					routine *parser.Runtime, ready chan string, stop int, inserted chan parser.RowCount) {
					runtime.LockOSThread()
					defer runtime.UnlockOSThread()

					for {
						chunk, more := <-chunks

						for i := range chunk {
							routine.ProcessRow(routine.ParseRow(chunk[i]))
						}

						if routine.RowCounter.Inserted > uint64(stop) {
							break
						}

						if !more {
							r.StopWorkers <- true
							break
						}
					}

					inserted <- routine.RowCounter
					err := routine.Close()
					if err != nil {
						log.Fatalf("Could not close runtime in routine: %s", err)
					}

					ready <- routine.GetPath()
					sem.Release(1)
					tw.Done()
					bufferPool.Put(routine.Buffers)
					routine = nil
				}(r.Data, &threadWorkers, sem, pr, readyForMerge, stop, rowCountChan)
			}
		}

		if p.RowID {
			for i := range p.Exclude {
				p.Exclude[i]--
			}
		}
	}

	readWg.Wait()
	insertWg.Wait()

	conversionElapsed = time.Since(conversionStart)

	if p.Sqlite {
		r.Close()
	}

	end := time.Now()
	totalTime := end.Sub(start)

	pv.Clear()
	if !quiet {
		var ops time.Duration
		var rows float64
		if rowCountTotal.Inserted != 0 {
			ops = time.Duration(totalTime.Nanoseconds() / int64(rowCountTotal.Inserted))
			rows = float64(rowCountTotal.Inserted) / totalTime.Seconds()
		} else {
			ops = time.Duration(0)
		}

		if csvReport {
			if logHeader == 1 {
				fmt.Fprintf(os.Stderr, "Total, Conversion, Merge, rows, op/s, rows/s, threads, enclosure, delimiter, chunksize, buffer-size, commit-rate, break-count, channel-capacity, go csv reader, capacity\n")
			}

			fmt.Fprintf(os.Stderr, "%s, %s, %s, %d, %s, %.f, %d, %s, %s, %d, %d, %d, %d, %d, %d, %d\n",
				totalTime.Round(time.Millisecond),
				conversionElapsed.Round(time.Millisecond),
				merged.Round(time.Millisecond),
				rowCountTotal.Inserted,
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
				rowCountTotal.GoParser,
				chunksize*p.ChannelCapacity*threads,
			)
		} else {
			report := Report{
				Parser:    "csv2sqlite",
				Version:   version,
				Delimiter: p.Characters.Delimiter.String(),
				Enclosure: p.Characters.Enclosure.String(),
				Start:     p.Start,
				Exclude:   p.Exclude,
				Nd:        p.NewlineDelimited,
				Php:       p.Characters.PHP,
				Params:    *p,
				CLIParams: os.Args,
			}

			report.Params.Buffersize /= 1024

			report.Runtime.Start = start.UnixNano()
			report.Runtime.End = end.UnixNano()

			report.Performance = Performance{
				Total:         totalTime.Round(time.Millisecond).String(),
				Conversion:    conversionElapsed.Round(time.Millisecond).String(),
				Merge:         merged.Round(time.Millisecond).String(),
				Rows:          rowCountTotal.Inserted,
				RowCount:      rowCountTotal,
				Ops:           ops.Round(time.Microsecond).String(),
				RowsPerSecond: math.Floor(rows),
				Threads:       threads,
				Chunksize:     chunksize,
				BreakCount:    stopCount,
				Capacity:      chunksize * p.ChannelCapacity * threads,
			}

			out, err := json.Marshal(report)
			if err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", out)
			} else {
				log.Fatal(err)
			}
		}
	}

	if rowCountTotal.Inserted == 0 {
		fmt.Fprintf(os.Stderr, "No rows inserted.\n")
		os.Exit(1)
	}

}

// samePath reports whether a and b name the same file, also through ./a.csv or a symlink.
func samePath(a, b string) bool {
	if a == b {
		return true
	}

	statA, err := os.Stat(a)
	if err != nil {
		return false
	}
	statB, err := os.Stat(b)
	if err != nil {
		return false
	}

	return os.SameFile(statA, statB)
}

type Report struct {
	Performance Performance
	Parser      string
	Version     string
	Delimiter   string
	Enclosure   string
	Start       int   `json:"starting_line"`
	Exclude     []int `json:"removed_columns"`
	Runtime     struct {
		Start int64
		End   int64
	}
	Nd        bool
	Php       bool
	Params    parser.Params
	CLIParams []string
}

type Performance struct {
	Total         string
	Conversion    string
	Merge         string
	Rows          uint64
	RowCount      parser.RowCount
	Ops           string
	RowsPerSecond float64
	Threads       int
	Chunksize     int
	BreakCount    int
	GoCsvReader   uint64
	Capacity      int
}
