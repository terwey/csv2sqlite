// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/util"
	"github.com/terwey/recordwriter/recordwriter"
)

type Params struct {
	Input           string
	Output          string
	CommitRate      int
	Buffersize      int
	Sqlite          bool
	Drop            bool
	Delimiter       characters.Character
	Enclosure       characters.Character
	ChannelCapacity int
	enclosureClean  bool
}

func NewParameters() *Params {
	p := &Params{
		CommitRate:     32768,
		Buffersize:     4096,
		Input:          "",
		Output:         "",
		Delimiter:      characters.Character_NIL,
		Enclosure:      characters.Character_NIL,
		enclosureClean: true,
	}

	return p
}

type Runtime struct {
	p       *Params
	readbuf []byte
	reader  *bufio.Reader

	Benchmark bool

	rw recordwriter.RecordWriter

	f    io.Reader
	fout *os.File

	splitFunc bufio.SplitFunc

	scanner *bufio.Scanner
	w       io.StringWriter

	Rows            uint64
	RowsCSV         uint64
	headerPrinted   bool
	colCount        int
	inside          bool
	doubleEnclosure []byte

	Header    [][]byte
	HeaderSet chan bool

	Data  chan [][]byte
	Count int

	sinput [][]byte
	// scolumns [][]byte
	sbuf []byte

	hasPrefixEnclosure bool
	hasSuffixEnclosure bool
}

var (
	DebugLog bool
	Overflow bool
)

func NewRuntime(p *Params, fn bufio.SplitFunc) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		splitFunc:     fn,
		Benchmark:     false,
		Count:         0,
	}

	recordwriter.DebugLog = DebugLog
	recordwriter.Overflow = Overflow

	r.HeaderSet = make(chan bool)
	r.Data = make(chan [][]byte, p.ChannelCapacity)

	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)
	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)

	if bytes.Compare(r.p.Enclosure, characters.Character_EMPTY) == 0 {
		r.p.enclosureClean = false
	}

	return r
}

func NewParser(p *Params, count int) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		Benchmark:     false,
		Count:         count,
	}

	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)
	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)

	return r
}

func (r *Runtime) SetDelimiter(delim characters.Character) {
	r.p.Delimiter = delim
}

func (r *Runtime) SetEnclosure(encl characters.Character) {
	r.p.Enclosure = encl
}

func (r *Runtime) Merge(ready <-chan string, wg *sync.WaitGroup, duration chan time.Duration) {
	// if sqlwriter, ok := r.rw.(*recordwriter.SQLiteWriter); ok {
	r.rw.Merge(ready, wg, duration)
	// }
}

func (r *Runtime) debug(format string, v ...interface{}) {
	debug.Debug("parser: "+format, v...)
}

func (r *Runtime) GetPath() string {
	if sqlwriter, ok := r.rw.(*recordwriter.SQLiteWriter); ok {
		return sqlwriter.GetPath()
	}
	return ""
}

func (r *Runtime) GetWriter() io.StringWriter {
	return r.w
}

func (r *Runtime) GetScanner() *bufio.Scanner {
	return r.scanner
}

func (r *Runtime) Init(input io.Reader) (err error) {
	if input != nil {
		r.f = input
	}

	var output string
	if r.Count == 0 {
		output = r.p.Output
	} else {
		output = fmt.Sprintf("%s.%d", r.p.Output, r.Count)
	}

	if r.p.Drop {
		r.rw = recordwriter.NewDropWriter()
	} else {
		if !r.p.Sqlite {
			if len(r.p.Output) == 0 {
				r.w = os.Stdout
				if r.Benchmark {
					r.w = bufio.NewWriterSize(bytes.NewBuffer([]byte{}), r.p.CommitRate)
				}
			} else {

				r.fout, err = os.Create(output)

				if err != nil {
					log.Fatalf("Cannot open file for writing: %s", err)
				}
				r.w = bufio.NewWriterSize(r.fout, r.p.CommitRate)
			}

			r.rw = recordwriter.NewCSVWriter(r.w)
		} else {
			if r.Count == 0 {
				r.rw = recordwriter.NewSQLWriter(output, r.p.CommitRate, true)
			} else {
				r.rw = recordwriter.NewSQLWriter(output, r.p.CommitRate, false)
			}
		}
	}

	if r.Count == 0 {
		r.scanner = r.NewScanner()
	}

	// log.Printf("output: %s", output)

	return nil
}

func (r *Runtime) Close() error {
	if file, ok := r.f.(io.ReadCloser); ok {
		file.Close()
	}

	return r.CloseDB()
}

// CloseDB is only used for multi-threaded parsing
func (r *Runtime) CloseDB() error {
	var err error
	if !r.p.Sqlite {
		if out, ok := r.GetWriter().(*bufio.Writer); ok {
			if err = out.Flush(); err != nil {
				log.Fatal(err)
			}
		}

		if len(r.p.Output) != 0 {
			if err = r.fout.Sync(); err != nil {
				log.Fatal(err)
			}

			return r.fout.Close()
		}
	} else {
		return r.rw.Close()
	}

	return nil
}

func (r *Runtime) NewScanner() *bufio.Scanner {
	r.readbuf = make([]byte, 0, r.p.Buffersize)

	r.reader = bufio.NewReaderSize(r.f, r.p.Buffersize)

	if _, err := io.ReadFull(r.reader, r.readbuf); err != nil && err != io.ErrUnexpectedEOF {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(r.reader)
	scanner.Buffer(r.readbuf, r.p.Buffersize)
	scanner.Split(r.splitFunc)

	return scanner
}

func (r *Runtime) ParseRow(data []byte) (cols []byte, pos []uint32) {
	if len(data) == 0 {
		return []byte{}, []uint32{}
	}

	cols, pos = r.SplitRow(data)
	if len(cols) == 0 {
		return []byte{}, []uint32{}
	}
	if !r.headerPrinted {
		if len(cols) == 1 {
			log.Printf("Only 1 column for header found, make sure the delimiter is set properly.")
		}
		r.ProcessHeader(util.FlatToMulti(cols, pos))
		// we should not output the header to the sqlite file, return nothing as it'll be skipped
		return []byte{}, []uint32{}
	}

	return cols, pos
}

func (r *Runtime) ProcessHeader(cols [][]byte) {
	for i, col := range cols {
		r.rw.WriteHeader(string(col))
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}
	if r.Count == 0 {
		r.Header = cols // needed for the second runtime
		r.HeaderSet <- true
	}
	r.colCount = len(cols)
	r.rw.FinalizeHeader()
	r.headerPrinted = true
}

func (r *Runtime) ProcessRow(cols []byte, pos []uint32) {
	if len(cols) == 0 {
		return
	}
	if debug.Ultra {
		r.rw.Multibyte(cols, pos, r.colCount)
		r.Rows++
		r.rw.EORMulti()
		return
	}

	if !debug.Ultra {
		cols := util.FlatToMulti(cols, pos)
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
	}

	r.Rows++

	r.rw.EOR(r.colCount)
}

func (r *Runtime) SplitRow(data []byte) (out []byte, pos []uint32) {
	// we need to check first if inside the data the delimiter occurs
	// we can count the amount of occurences of the delimiter and compare
	// if that matches the amount we expect
	// amount = columns + 1

	count := Count(data, r.p.Delimiter) + 1
	if r.colCount == count {
		out, pos = util.GenSplit(data, r.p.Delimiter, -1, r.clean)
	}

	// r.colCount is set to non-zero on header creation
	// TODO: fix that the header gets parsed by our code too
	if r.colCount != 0 {
		// count := Count(data, r.p.Delimiter)
		// if the count is correct we can just simply do a normal split and skip this logic
		if count > r.colCount {
			o, p := r.ParseUnescapedDelimiter(util.ColSplit(data, r.p.Delimiter, r.p.Enclosure, -1))
			if len(p) == r.colCount {
				return o, p
			}
		}
	}

	// columns = bytes.Split(data, r.p.Delimiter)

	// the SplitUnescapedDelimiter didn't work properly, use slow method
	if count > r.colCount {

		// for simplicity sake we use the Go's build in CSV parser when the
		// delimiter occurs within the quotation
		// the reason for not always using it, is simply speed.
		csvReader := csv.NewReader(bytes.NewReader(data))
		csvReader.Comma = r.p.Delimiter.Rune()

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				// let the other logs shutdown
				time.Sleep(2 * time.Second)
				log.Printf("csv reading failed: %s, data: \n%s", err, data)
				panic(err)
			}

			for _, col := range record {
				buf := r.clean([]byte(col))
				out = append(out, buf...)
				pos = append(pos, uint32(len(buf)))
			}
		}

		if r.colCount != 0 {
			debug.LogReader(data, count, r.colCount)
			r.RowsCSV++
		}

		return out, pos
	}

	return
}

var (
	carriageReturn []byte = []byte("\x0d")     // \r
	lineFeed       []byte = []byte("\x0a")     // \n
	crLf           []byte = []byte("\x0d\x0a") // \r\n
)

func (r *Runtime) clean(data []byte) []byte {
	if r.p.enclosureClean {
		data = bytes.TrimPrefix(data, r.p.Enclosure)
		data = bytes.TrimSuffix(data, r.p.Enclosure)

		// we need to clean double quotes
		data = ReplaceAll(data, r.doubleEnclosure, r.p.Enclosure)
	}
	data = ReplaceAll(data, crLf, lineFeed)
	data = ReplaceAll(data, carriageReturn, lineFeed)

	return data
}

func (r *Runtime) ParseUnescapedDelimiter(input [][]byte) (data []byte, pos []uint32) {
	// NOTE: this method  doesn't work on optional enclosures.

	// we need to seek through the data and find the first Enclosure and escape the delimiter

	// we first split on the delimiter and then check if two Enclosures are available
	// r.sinput = bytes.Split(data, r.p.Delimiter)
	// r.sinput = splitData
	// r.scolumns = make([][]byte, 0)
	r.sbuf = make([]byte, 0)
	enclosurecheck := false

	for i := range input {
		// r.debug("input[i]: >>>%s<<", r.sinput[i])
		// r.debug("buf: %s", r.sbuf)
		// r.debug("columns: %s", r.scolumns)
		// debug.DebugColumns(r.scolumns)
		count := 0
		counted := false

		if !r.inside {
			count = Count(input[i], r.p.Enclosure)
			counted = true
			if count != 0 {
				r.inside = true
			}
		}

		// r.debug("inside: %v, enclosurecheck: %v", r.inside, enclosurecheck)

		// there is a case where the last buf can end with """ or similar
		// we need to check if the NEXT starts with a " if so, move buf to columns and continue happily with life
		if enclosurecheck {
			if len(input[i]) == 1 {
				// data was something like `|"foo|"|`
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				// add the "empty" data, anti-slipmat func ;)
				// check if input is r.p.Enclosure
				if !bytes.Equal(input[i], r.p.Enclosure) {
					r.sbuf = append(r.sbuf, input[i]...)
				}
			}
			if bytes.HasPrefix(input[i], r.p.Enclosure) {
				r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
				r.sbuf = bytes.TrimSuffix(r.sbuf, r.p.Enclosure)
				// we need to clean double enclosures to single
				r.sbuf = ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
				r.sbuf = ReplaceAll(r.sbuf, crLf, lineFeed)
				r.sbuf = ReplaceAll(r.sbuf, carriageReturn, lineFeed)
				data = append(data, r.sbuf...)
				pos = append(pos, uint32(len(r.sbuf)))
				r.sbuf = make([]byte, 0)
				r.inside = false
				enclosurecheck = false
			}
			if len(input[i]) == 1 {
				continue
			}
		}

		// unEnclosured column
		if len(input[i]) == 0 {
			// delimiter inside an enclosure
			if r.inside {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				continue
			}
			input[i] = ReplaceAll(input[i], crLf, lineFeed)
			input[i] = ReplaceAll(input[i], carriageReturn, lineFeed)
			data = append(data, input[i]...)
			pos = append(pos, uint32(len(input[i])))
			continue
		}

		if len(input[i]) == 1 {
			if len(r.sbuf) != 0 && bytes.Equal(input[i], r.p.Enclosure) {
				// the delimiter appears at the end, finalize the column
				r.sbuf = append(r.sbuf, r.p.Delimiter...)

				// we need to clean double enclosures to single
				r.sbuf = ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
				r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
				r.sbuf = ReplaceAll(r.sbuf, crLf, lineFeed)
				r.sbuf = ReplaceAll(r.sbuf, carriageReturn, lineFeed)
				data = append(data, r.sbuf...)
				pos = append(pos, uint32(len(r.sbuf)))
				r.sbuf = make([]byte, 0)
				continue
			}

			// in the case it is NOT an Enclosure, place a delimiter and append the buffer
			if !bytes.Equal(input[i], r.p.Enclosure) {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
			}
			r.sbuf = append(r.sbuf, input[i]...)

			continue
		}

		// if we counted the enclosures before we don't have to check if there is a prefix or suffix
		if counted && count == 0 {
			r.hasPrefixEnclosure = false
			r.hasSuffixEnclosure = false
		} else {
			r.hasPrefixEnclosure = bytes.HasPrefix(input[i], r.p.Enclosure)
			r.hasSuffixEnclosure = bytes.HasSuffix(input[i], r.p.Enclosure)
		}

		// log.Printf("r.hasPrefixEnclosure: %v, r.hasSuffixEnclosure: %v, count: %d", r.hasPrefixEnclosure, r.hasSuffixEnclosure, count)

		if !r.hasPrefixEnclosure || !r.hasSuffixEnclosure {
			if len(r.sbuf) != 0 {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				r.sbuf = append(r.sbuf, input[i]...)

				// result := bytes.HasSuffix(input[i], r.p.Enclosure) && !bytes.HasSuffix(input[i], r.doubleEnclosure)
				// r.debug("bytes.HasSuffix(input[i], r.p.Enclosure): %v\n!bytes.HasSuffix(input[i], r.doubleEnclosure: %v\n statement: %v", bytes.HasSuffix(input[i], r.p.Enclosure), !bytes.HasSuffix(input[i], r.doubleEnclosure), result)
				if r.hasSuffixEnclosure && !bytes.HasSuffix(input[i], r.doubleEnclosure) {
					r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
					// we know for sure that the suffix exists
					r.sbuf = TrimSuffixSkipCheck(r.sbuf, r.p.Enclosure)
					// we need to clean double enclosures to single
					r.sbuf = ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
					r.sbuf = ReplaceAll(r.sbuf, crLf, lineFeed)
					r.sbuf = ReplaceAll(r.sbuf, carriageReturn, lineFeed)
					data = append(data, r.sbuf...)
					pos = append(pos, uint32(len(r.sbuf)))
					r.sbuf = make([]byte, 0)
					r.inside = false
					enclosurecheck = false
				} else {
					enclosurecheck = true
				}
			} else {
				// we are missing the second Enclosure, add to the r.sbuf
				r.sbuf = append(r.sbuf, input[i]...)
			}
			continue
		}

		if count%2 == 0 && r.hasPrefixEnclosure && r.hasSuffixEnclosure {
			// we know for sure it has a prefix and suffix
			input[i] = TrimPrefixSkipCheck(input[i], r.p.Enclosure)
			input[i] = TrimSuffixSkipCheck(input[i], r.p.Enclosure)

			// we need to clean double enclosures to single
			input[i] = ReplaceAll(input[i], r.doubleEnclosure, r.p.Enclosure)
			input[i] = ReplaceAll(input[i], crLf, lineFeed)
			input[i] = ReplaceAll(input[i], carriageReturn, lineFeed)
			data = append(data, input[i]...)
			pos = append(pos, uint32(len(input[i])))
			r.inside = false
			continue
		}

		// edge-case where ""| appears
		// or where count%2 != 0
		if count != 2 {
			r.sbuf = append(r.sbuf, input[i]...)
			continue
		}

		input[i] = bytes.TrimPrefix(input[i], r.p.Enclosure)
		input[i] = bytes.TrimSuffix(input[i], r.p.Enclosure)
		// we need to clean double enclosures to single
		input[i] = ReplaceAll(input[i], r.doubleEnclosure, r.p.Enclosure)
		input[i] = ReplaceAll(input[i], crLf, lineFeed)
		input[i] = ReplaceAll(input[i], carriageReturn, lineFeed)
		data = append(data, input[i]...)
		pos = append(pos, uint32(len(input[i])))
	}

	// r.debug("columns: %d", len(columns))

	return
}

// Count calls bytes.IndexByte before Count
func Count(s, sep []byte) int {
	if len(sep) == 1 {
		if bytes.IndexByte(s, sep[0]) == -1 {
			return 0
		}
	}
	return bytes.Count(s, sep)
}

// ReplaceAll calls bytes.Contains beforehand preventing alloc
func ReplaceAll(s, old, new []byte) []byte {
	if !bytes.Contains(s, old) {
		return s
	}
	return bytes.ReplaceAll(s, old, new)
}

// TrimSuffixSkipCheck skips the check if the suffix exists
func TrimSuffixSkipCheck(s, suffix []byte) []byte {
	return s[:len(s)-len(suffix)]
}

func TrimPrefixSkipCheck(s, prefix []byte) []byte {
	return s[len(prefix):]
}
