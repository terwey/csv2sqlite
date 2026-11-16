// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/characters/knownconfigs"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/util"
	"github.com/terwey/recordwriter/recordwriter"
)

type RowCount struct {
	Inserted        uint64
	GoParser        uint64
	Dropped         uint64
	GoParserDropped uint64
}

type Params struct {
	Input  string
	Output string

	// the small SQLite files written are normally in the same
	// target as the output file, for performance reasons (or
	// storage reasons) a different workdir can be provided.
	// e.g. a tmpfs ramdisk
	Scratch string

	Characters *util.CharacterConfig

	CommitRate int
	Buffersize int

	// starting position of the CSV file
	// this is needed for Excel to CSV exports
	// that contain data above the header
	Start           int
	Skip            int // amount of rows to skip AFTER the header
	Limit           int // maximum number of rows to process
	CSVEnclosureSet int

	Exclude []int

	Sqlite bool
	Drop   bool

	// When set to true the CSV does not have newlines
	// (CR, LF or CRLF) inside records
	NewlineDelimited bool

	// if the order of the incoming CSV matters the RowID
	// can be set to true and during Scan the RowID will be added
	RowID bool
}

func NewParameters() *Params {
	p := &Params{
		CommitRate: 32768,
		Buffersize: 4096,
		Input:      "",
		Output:     "",

		Limit: -1,
	}

	p.Characters = util.NewCharacterConfig(characters.Character_NIL, characters.Character_NIL)

	return p
}

// Runtime writes the records of one output: the output itself when
// ParserCount is 0, and otherwise the part of it a worker fills.
type Runtime struct {
	p    *Params
	rw   recordwriter.RecordWriter
	fout *os.File
	w    io.Writer

	Header          [][]byte
	ProcessedHeader []string

	RowCounter RowCount

	ParserCount int
	colCount    int

	rowID []byte
}

var (
	RWDebugLog bool
	Overflow   bool
)

func NewRuntime(p *Params) *Runtime {
	r := &Runtime{
		p:           p,
		ParserCount: 0,
	}

	recordwriter.DebugLog = RWDebugLog
	recordwriter.Overflow = Overflow

	return r
}

func NewParser(p *Params, parsercount int) *Runtime {
	r := &Runtime{
		p:           p,
		ParserCount: parsercount,
	}

	return r
}

func (r *Runtime) Merge(ready <-chan string, wg *sync.WaitGroup, duration chan time.Duration) {
	// if sqlwriter, ok := r.rw.(*recordwriter.SQLiteWriter); ok {
	r.rw.Merge(ready, wg, duration)
	// }
}

func (r *Runtime) GetPath() string {
	return r.rw.GetPath()
}

func (r *Runtime) GetWriter() io.Writer {
	return r.w
}

func (r *Runtime) Init() {
	r.rw = r.newRecordWriter(r.generateOutputPath())
}

func (r *Runtime) newRecordWriter(output string) recordwriter.RecordWriter {
	if r.p.Drop {
		return recordwriter.NewDropWriter()
	}

	if r.p.Sqlite {
		return recordwriter.NewSQLWriter(output, r.p.CommitRate, r.ParserCount == 0, r.p.RowID)
	}

	if len(r.p.Output) == 0 {
		r.w = os.Stdout
	} else {
		var err error
		r.fout, err = os.Create(output)

		if err != nil {
			log.Fatalf("Cannot open file for writing: %s", err)
		}
		r.w = bufio.NewWriterSize(r.fout, r.p.CommitRate)
	}

	csv := recordwriter.NewCSVWriter(knownconfigs.CSV(), r.w, output)
	csv.SetEnclosureSet(r.p.CSVEnclosureSet)
	return csv
}

func (r *Runtime) generateOutputPath() string {
	if r.ParserCount == 0 {
		return r.p.Output
	}

	var output string
	if r.p.Scratch != "" {
		output = fmt.Sprintf("%s.%d", filepath.Join(r.p.Scratch, filepath.Base(r.p.Output)), r.ParserCount)
	} else {
		output = fmt.Sprintf("%s.%d", r.p.Output, r.ParserCount)
	}
	return output
}

func (r *Runtime) Close() error {
	return r.CloseDB()
}

// CloseDB is only used for multi-threaded parsing
func (r *Runtime) CloseDB() error {
	var err error

	if r.p.Sqlite {
		return r.rw.Close()
	}

	if out, ok := r.GetWriter().(*bufio.Writer); ok {
		if err = out.Flush(); err != nil {
			log.Fatalf("Runtime: could not flush writer: %s", err)
		}
	}

	if f, ok := r.w.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			log.Printf("sync: %s", r.GetPath())
			f.Sync()
		}
	}

	if len(r.p.Output) != 0 {
		if err = r.fout.Sync(); err != nil {
			log.Fatalf("Runtime: could not sync writer: %s", err)
		}

		return r.fout.Close()
	}

	return nil
}

// ProcessHeader writes the header: the names of the columns, with rowid as the
// first one when the rows carry a row id.
func (r *Runtime) ProcessHeader(cols [][]byte) {
	// we do not need to print the header multiple times for CSV
	if _, ok := r.rw.(*recordwriter.CSVWriter); ok {
		if r.ParserCount != 0 {
			r.colCount = len(cols)
			return
		}
	}
	if len(r.ProcessedHeader) != 0 {
		if srw, ok := r.rw.(*recordwriter.SQLiteWriter); ok {
			srw.SetProcessedHeader(r.ProcessedHeader)
		}
	}

	for i, col := range cols {
		r.rw.WriteHeader(string(col))
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}
	if r.ParserCount == 0 {
		r.Header = cols // needed for the second runtime
	}
	r.ProcessedHeader = r.rw.FinalizeHeader()
	r.colCount = len(cols)
}

// ProcessRow writes one record: v holds its fields back to back and pos the
// length of each. row is its number in the input, written as the row id when
// the rows carry one.
func (r *Runtime) ProcessRow(row uint64, v []byte, pos []uint32) {
	if len(v) == 0 {
		return
	}

	// if r.p.RowID {
	// TODO: if we want the rowid often, we need
	// to be able to use the Ultra mode
	// which requires the code to pass an integer
	// to the sqlite C code
	// 	// v, pos = util.RowIDToInteger(v, pos)
	// }

	if !r.p.RowID && debug.Ultra {
		r.rw.Multibyte(v, pos, r.colCount)
		r.RowCounter.Inserted++
		r.rw.EORMulti()
		return
	}

	if r.p.RowID {
		r.rowID = strconv.AppendUint(r.rowID[:0], row, 10)
		r.rw.Write(string(r.rowID))
	}

	cols := util.FlatToMulti(v, pos)
	for i, col := range cols {
		if len(col) == 0 {
			r.rw.Skip()
		} else {
			r.rw.WriteByte(col)
		}
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}

	r.RowCounter.Inserted++

	r.rw.EOR(r.colCount)
}
