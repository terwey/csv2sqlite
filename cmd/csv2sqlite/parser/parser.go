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

	"github.com/terwey/csv2sqlite/delimiter"
	"github.com/terwey/recordwriter/recordwriter"
)

type Params struct {
	Input           string
	Output          string
	FlushSize       int
	Buffersize      int
	Sqlite          bool
	Delimiter       delimiter.Delimiter
	Enclosure       []byte
	ChannelCapacity int
}

func NewParameters() *Params {
	p := &Params{
		FlushSize:  32768,
		Buffersize: 4096,
		Input:      "",
		Output:     "",
		Delimiter:  delimiter.Delimiter_COMMA,
		Enclosure:  []byte(`"`),
	}

	return p
}

type Runtime struct {
	p       *Params
	readbuf []byte
	reader  *bufio.Reader
	buf     [][]byte

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

	Data  chan []byte
	Count int

	Debug bool

	sinput   [][]byte
	scolumns [][]byte
	sbuf     []byte

	hasPrefixEnclosure bool
	hasSuffixEnclosure bool
}

func NewRuntime(p *Params, fn bufio.SplitFunc) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		splitFunc:     fn,
		Benchmark:     false,
		Count:         0,
	}

	r.HeaderSet = make(chan bool)
	r.Data = make(chan []byte, p.ChannelCapacity)

	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)
	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)

	return r
}

func NewParser(p *Params, count int) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		Benchmark:     false,
		Count:         count,
	}

	r.Data = make(chan []byte, p.ChannelCapacity)

	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)
	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)

	return r
}

func (r *Runtime) SetDelimiter(delim delimiter.Delimiter) {
	r.p.Delimiter = delim
}

func (r *Runtime) Merge(other ...int) {
	if sqlwriter, ok := r.rw.(*recordwriter.SQLiteWriter); ok {
		sqlwriter.Merge(other...)
	}
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

	if !r.p.Sqlite {
		if len(r.p.Output) == 0 {
			r.w = os.Stdout
			if r.Benchmark {
				r.w = bufio.NewWriterSize(bytes.NewBuffer([]byte{}), r.p.FlushSize)
			}
		} else {

			r.fout, err = os.Create(output)

			if err != nil {
				log.Fatalf("Cannot open file for writing: %s", err)
			}
			r.w = bufio.NewWriterSize(r.fout, r.p.FlushSize)
		}

		r.rw = recordwriter.NewCSVWriter(r.w)
	} else {
		r.rw = recordwriter.NewSQLWriter(output, r.p.FlushSize)
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

func (r *Runtime) ParseRow(data []byte) [][]byte {
	if len(data) == 0 {
		return [][]byte{}
	}

	cols := r.SplitRow(data)
	if len(cols) == 0 {
		return [][]byte{}
	}
	if !r.headerPrinted {
		if len(cols) == 1 {
			log.Printf("Only 1 column for header found, make sure the delimiter is set properly.")
		}
		r.ProcessHeader(cols)
		// we should not output the header to the sqlite file, return nothing as it'll be skipped
		return [][]byte{}
	}

	return cols
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

func (r *Runtime) ProcessRow(cols [][]byte) {
	if len(cols) == 0 {
		return
	}
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

	r.Rows++

	r.rw.EOR(r.colCount)
}

func (r *Runtime) SplitRow(data []byte) (columns [][]byte) {
	// we need to check first if inside the data the delimiter occurs
	// we can count the amount of occurences of the delimiter and compare
	// if that matches the amount we expect
	// amount = columns + 1

	r.debug("data: \n%s", data)

	// r.colCount is set to non-zero on header creation
	if r.colCount != 0 {
		count := bytes.Count(data, r.p.Delimiter)
		// if the count is correct we can just simply do a normal split and skip this logic
		if count+1 != r.colCount {
			columns := r.SplitUnescapedDelimiter(data)
			if len(columns) == r.colCount {
				return columns
			}
		}
	}

	columns = bytes.Split(data, r.p.Delimiter)

	// the SplitUnescapedDelimiter didn't work properly, use slow method
	if len(columns) > r.colCount {
		// for simplicity sake we use the Go's build in CSV parser when the
		// delimiter occurs within the quotation
		// the reason for not always using it, is simply speed.
		csvReader := csv.NewReader(bytes.NewReader(data))
		csvReader.Comma = []rune(string(r.p.Delimiter))[0]

		columns = make([][]byte, 0)

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("csv reading failed: %s, data: \n%s", err, data)
			}

			for _, col := range record {
				columns = append(columns, []byte(col))
			}
		}

		r.RowsCSV++

		return columns
	}

	for i := range columns {
		columns[i] = bytes.TrimPrefix(columns[i], r.p.Enclosure)
		columns[i] = bytes.TrimSuffix(columns[i], r.p.Enclosure)

		// we need to clean double quotes
		columns[i] = bytes.ReplaceAll(columns[i], r.doubleEnclosure, r.p.Enclosure)

		columns[i] = bytes.ReplaceAll(columns[i], carriageReturn, lineFeed)
	}

	return columns
}

func (r *Runtime) debug(format string, v ...interface{}) {
	if r.Debug {
		log.Printf(format, v...)
	}
}

func (r *Runtime) debugColumns(cols [][]byte) {
	if r.Debug {
		if len(cols) == 0 {
			return
		}
		// log.Printf("Columns")
		log.Printf("Columns: [%3d]\t%s", len(cols)-1, cols[len(cols)-1])
		// for i := range cols {
		// 	log.Printf("[%3d]\t%s", i, cols[i])
		// }
		// log.Printf("End of Columns")
	}
}

var (
	carriageReturn []byte = []byte("\x0d") // \r
	lineFeed       []byte = []byte("\x0a") // \n
)

func (r *Runtime) SplitUnescapedDelimiter(data []byte) [][]byte {
	// we need to seek through the data and find the first Enclosure and escape the delimiter

	// we first split on the delimiter and then check if two Enclosures are available
	r.sinput = bytes.Split(data, r.p.Delimiter)
	r.scolumns = make([][]byte, 0)
	r.sbuf = make([]byte, 0)
	enclosurecheck := false

	for i := range r.sinput {
		r.debug("input[i]: >>>%s<<", r.sinput[i])
		r.debug("buf: %s", r.sbuf)
		// r.debug("columns: %s", r.scolumns)
		r.debugColumns(r.scolumns)
		count := 0
		if !r.inside {
			count = bytes.Count(r.sinput[i], r.p.Enclosure)
			if count != 0 {
				r.inside = true
			}
		}

		// r.debug("inside: %v, enclosurecheck: %v", r.inside, enclosurecheck)

		// there is a case where the last buf can end with """ or similar
		// we need to check if the NEXT starts with a " if so, move buf to columns and continue happily with life
		if enclosurecheck {
			if len(r.sinput[i]) == 1 {
				// data was something like `|"foo|"|`
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				// add the "empty" data, anti-slipmat func ;)
				// check if input is r.p.Enclosure
				if !bytes.Equal(r.sinput[i], r.p.Enclosure) {
					r.sbuf = append(r.sbuf, r.sinput[i]...)
				}
			}
			if bytes.HasPrefix(r.sinput[i], r.p.Enclosure) {
				r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
				r.sbuf = bytes.TrimSuffix(r.sbuf, r.p.Enclosure)
				// we need to clean double enclosures to single
				r.sbuf = bytes.ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
				r.sbuf = bytes.ReplaceAll(r.sbuf, carriageReturn, lineFeed)
				r.scolumns = append(r.scolumns, r.sbuf)
				r.sbuf = make([]byte, 0)
				r.inside = false
				enclosurecheck = false
			}
			if len(r.sinput[i]) == 1 {
				continue
			}
		}

		// unEnclosured column
		if len(r.sinput[i]) == 0 {
			// delimiter inside an enclosure
			if r.inside {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				continue
			}
			r.sinput[i] = bytes.ReplaceAll(r.sinput[i], carriageReturn, lineFeed)
			r.scolumns = append(r.scolumns, r.sinput[i])
			continue
		}

		if len(r.sinput[i]) == 1 {
			if len(r.sbuf) != 0 && bytes.Equal(r.sinput[i], r.p.Enclosure) {
				// the delimiter appears at the end, finalize the column
				r.sbuf = append(r.sbuf, r.p.Delimiter...)

				// we need to clean double enclosures to single
				r.sbuf = bytes.ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
				r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
				r.sbuf = bytes.ReplaceAll(r.sbuf, carriageReturn, lineFeed)
				r.scolumns = append(r.scolumns, r.sbuf)
				r.sbuf = make([]byte, 0)
				continue
			}

			// in the case it is NOT an Enclosure, place a delimiter and append the buffer
			if !bytes.Equal(r.sinput[i], r.p.Enclosure) {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
			}
			r.sbuf = append(r.sbuf, r.sinput[i]...)

			continue
		}

		r.hasPrefixEnclosure = bytes.HasPrefix(r.sinput[i], r.p.Enclosure)
		r.hasSuffixEnclosure = bytes.HasSuffix(r.sinput[i], r.p.Enclosure)

		if !r.hasPrefixEnclosure || !r.hasSuffixEnclosure {
			if len(r.sbuf) != 0 {
				r.sbuf = append(r.sbuf, r.p.Delimiter...)
				r.sbuf = append(r.sbuf, r.sinput[i]...)

				// result := bytes.HasSuffix(input[i], r.p.Enclosure) && !bytes.HasSuffix(input[i], r.doubleEnclosure)
				// r.debug("bytes.HasSuffix(input[i], r.p.Enclosure): %v\n!bytes.HasSuffix(input[i], r.doubleEnclosure: %v\n statement: %v", bytes.HasSuffix(input[i], r.p.Enclosure), !bytes.HasSuffix(input[i], r.doubleEnclosure), result)
				if r.hasSuffixEnclosure && !bytes.HasSuffix(r.sinput[i], r.doubleEnclosure) {
					r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
					r.sbuf = bytes.TrimSuffix(r.sbuf, r.p.Enclosure)
					// we need to clean double enclosures to single
					r.sbuf = bytes.ReplaceAll(r.sbuf, r.doubleEnclosure, r.p.Enclosure)
					r.sbuf = bytes.ReplaceAll(r.sbuf, carriageReturn, lineFeed)
					r.scolumns = append(r.scolumns, r.sbuf)
					r.sbuf = make([]byte, 0)
					r.inside = false
					enclosurecheck = false
				} else {
					enclosurecheck = true
				}
			} else {
				// we are missing the second Enclosure, add to the r.sbuf
				r.sbuf = append(r.sbuf, r.sinput[i]...)
				// r.sbuf = bytes.TrimPrefix(r.sbuf, r.p.Enclosure)
			}
			continue
		}

		if count%2 == 0 && r.hasPrefixEnclosure && r.hasSuffixEnclosure {
			r.sinput[i] = bytes.TrimPrefix(r.sinput[i], r.p.Enclosure)
			r.sinput[i] = bytes.TrimSuffix(r.sinput[i], r.p.Enclosure)

			// we need to clean double enclosures to single
			r.sinput[i] = bytes.ReplaceAll(r.sinput[i], r.doubleEnclosure, r.p.Enclosure)
			r.sinput[i] = bytes.ReplaceAll(r.sinput[i], carriageReturn, lineFeed)
			r.scolumns = append(r.scolumns, r.sinput[i])
			r.inside = false
			continue
		}

		// edge-case where ""| appears
		// or where count%2 != 0
		if count != 2 {
			r.sbuf = append(r.sbuf, r.sinput[i]...)
			continue
		}

		r.sinput[i] = bytes.TrimPrefix(r.sinput[i], r.p.Enclosure)
		r.sinput[i] = bytes.TrimSuffix(r.sinput[i], r.p.Enclosure)
		// we need to clean double enclosures to single
		r.sinput[i] = bytes.ReplaceAll(r.sinput[i], r.doubleEnclosure, r.p.Enclosure)
		r.sinput[i] = bytes.ReplaceAll(r.sinput[i], carriageReturn, lineFeed)
		r.scolumns = append(r.scolumns, r.sinput[i])
	}

	// r.debug("columns: %d", len(columns))

	return r.scolumns
}
