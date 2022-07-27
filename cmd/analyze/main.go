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
	"sync"
	"sync/atomic"

	"github.com/productsupcom/go-sqlite-lite/sqlite3"
	st "github.com/terwey/csv2sqlite/cmd/analyze/stats"
	"github.com/terwey/csv2sqlite/sqlitehelper"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"
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

	wasmModule string

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

	flags.StringVar(&wasmModule, "wasm-module", "", "Filename for a WASM module to use to Analyze")

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

	/*
		if len(wasmModule) != 0 {
			log.Printf("using wasm module: %s", wasmModule)
			wasmBytes, err := ioutil.ReadFile(wasmModule)
			if err != nil {
				log.Fatalf("error reading wasm file from disk: %s", err)
			}

			engine := wasmer.NewEngine()
			store := wasmer.NewStore(engine)

			// Compiles the module
			module, err := wasmer.NewModule(store, wasmBytes)
			if err != nil {
				log.Fatalf("error compiling wasm module: %s", err)
			}

			wasiEnv, err := wasmer.NewWasiStateBuilder("wasi-program").Finalize()
			if err != nil {
				log.Fatalf("error creating WASI Env for wasm module: %s", err)
			}

			importObject, err := wasiEnv.GenerateImportObject(store, module)
			if err != nil {
				log.Fatalf("error generating import object for wasm module: %s", err)
			}

			// // Instantiates the module
			// importObject := wasmer.NewImportObject()
			instance, err := wasmer.NewInstance(module, importObject)
			if err != nil {
				log.Fatalf("error creating instance for wasm module: %s", err)
			}

			// start, err := instance.Exports.GetWasiStartFunction()
			// if err != nil {
			// 	log.Fatalf("error getting WASI Start function: %s", err)
			// }

			// foo, err := start()
			// if err != nil {
			// 	log.Fatalf("error calling WASI start func: %s", err)
			// }

			// log.Printf("foo contains: %v", foo)

			// Gets the `sum` exported function from the WebAssembly instance.
			wasmAnalyze, err := instance.Exports.GetFunction("analyzeColumn")
			if err != nil {
				log.Fatalf("error getting function from wasm module: %s", err)
			}

			ci, err := wasmAnalyze("foo", 1, 3, 4)
			if err != nil {
				log.Fatalf("wasmAnalyze: %s", err)
			}

			log.Printf("ci: %v", ci)
		}
	*/

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

	stats := st.NewStatistics(names)

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

	var providerWg sync.WaitGroup
	provider := make(chan *st.Statistics, threads*10)
	providerWg.Add(1)
	go func(mainstats *st.Statistics, provider chan *st.Statistics, wg *sync.WaitGroup) {
		defer wg.Done()
		for in := range provider {
			mainstats.Merge(in)
		}
	}(stats, provider, &providerWg)

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
		go func(provider chan<- *st.Statistics, w *sync.WaitGroup, sem *semaphore.Weighted, limit, offset int, pv *progressbar.ProgressBar, iter *int64) {
			conn, err := sqlite3.Open(input)
			if err != nil {
				log.Fatalf("Cannot open %s: %s", input, err)
			}
			defer conn.Close()

			stmt, err := conn.Prepare(sqlitehelper.QueryForIndexed([]string{"*"}, nil, table, limit, offset))
			if err != nil {
				log.Fatalf("Prepare: %s", err)
			}

			defer stmt.Close()

			stat := st.NewStatistics(names)

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

					stat.Add(AnalyzeColumn(val, i))
				}

				pv.Add(1)
				atomic.AddInt64(iter, 1)
			}

			provider <- stat
			w.Done()
			sem.Release(1)
		}(provider, &wg, sem, limit, offset, pv, &iterations)

		//           					   5591303 takes 0m28.078s
		// if atomic.LoadInt64(&iterations) > 200000 {
		// 	break
		// }

		left -= max
		offset += max
	}

	wg.Wait()
	close(provider)
	providerWg.Wait()

	// TODO: the count should be equal to iter
	stats.Rows = uint64(iterations)
	stats.Calculate()
	log.Printf("stats: \n%s\n", stats)
}

func fanIn(chans ...<-chan st.ColumnInfo) <-chan st.ColumnInfo {
	out := make(chan st.ColumnInfo, 10)
	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, c := range chans {
		go func(workerChannel <-chan st.ColumnInfo, wg *sync.WaitGroup) {
			for r := range workerChannel {
				out <- r
			}
			wg.Done()
		}(c, &wg)
	}
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(out)
	}(&wg)

	return out
}

func AnalyzeColumn(data interface{}, pos int) *st.ColumnInfo {
	// disk read test
	// return &ColumnInfo{
	// 	Pos:        pos,
	// 	ColumnType: ColumnType_Number,
	// 	Length:     10,
	// }

	r := &st.ColumnInfo{
		Pos: pos,
	}
	if data == nil {
		r.Length = 0
		r.ColumnType = st.ColumnType_Empty

		return r
	}

	if v, ok := data.(string); ok {
		r.Length = uint64(len(v))
		if r.Length == 0 {
			r.ColumnType = st.ColumnType_Empty
			return r
		} else {
			r.ColumnType = st.DetectColumnType(v)
		}
	}

	if v, ok := data.(sqlite3.RawString); ok {
		r.Length = uint64(len(v))
		if r.Length == 0 {
			r.ColumnType = st.ColumnType_Empty
			return r
		} else {
			r.ColumnType = st.DetectColumnType(string(v))
		}
	}

	return r
}

// type SQLiteConn struct {
// 	*sqlite3.Conn
// }

// func (c *SQLiteConn) RowInfo(sql string) []*ColumnInfo {
// 	if rc := C.sqlite3_exec(c.db, sql, nil, nil, nil); rc != OK {
// 		return libErr(rc, c.db)
// 	}
// 	return nil
// }
