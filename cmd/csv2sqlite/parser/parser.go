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

	"github.com/terwey/recordwriter/recordwriter"
)

const nonChar string = "\uFFFF"

type Params struct {
	Input      string
	Output     string
	FlushSize  int
	Buffersize int
	Sqlite     bool
	Delimiter  []byte
	Enclosure  []byte
}

func NewParameters() *Params {
	p := &Params{
		FlushSize:  32768,
		Buffersize: 4096,
		Input:      "",
		Output:     "",
		Delimiter:  []byte(`|`),
		Enclosure:  []byte(`"`),
	}

	return p
}

type Runtime struct {
	p      *Params
	buf    []byte
	reader *bufio.Reader

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
	columns         int
	inside          bool
	doubleEnclosure []byte

	Header [][]byte

	Data  chan [][]byte
	Count int

	Debug bool
}

func NewRuntime(p *Params, fn bufio.SplitFunc) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		splitFunc:     fn,
		Benchmark:     false,
		Count:         0,
	}

	r.Data = make(chan [][]byte, 200)

	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)
	r.doubleEnclosure = append(r.doubleEnclosure, p.Enclosure...)

	return r
}

func (r *Runtime) GetWriter() io.StringWriter {
	return r.w
}

func (r *Runtime) GetScanner() *bufio.Scanner {
	return r.scanner
}

func (r *Runtime) Init(input io.Reader) (err error) {
	r.f = input

	// convert to kb
	r.p.FlushSize *= 1024
	r.p.Buffersize *= 1024

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

	log.Printf("output: %s", output)

	return nil
}

func (r *Runtime) Close() {
	var err error
	if file, ok := r.f.(io.ReadCloser); ok {
		file.Close()
	}
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

			r.fout.Close()
		}
	} else {
		r.rw.Close()
	}
}

func (r *Runtime) NewScanner() *bufio.Scanner {
	r.buf = make([]byte, 0, r.p.Buffersize)

	r.reader = bufio.NewReaderSize(r.f, r.p.Buffersize)

	if _, err := io.ReadFull(r.reader, r.buf); err != nil && err != io.ErrUnexpectedEOF {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(r.reader)
	scanner.Buffer(r.buf, r.p.Buffersize)
	scanner.Split(r.splitFunc)

	return scanner
}

func (r *Runtime) ParseRow(data []byte) {
	cols := r.SplitRow(data)
	if len(cols) == 0 {
		return
	}
	if !r.headerPrinted {
		r.ProcessHeader(cols)
		return
	}

	r.Data <- cols
}

func (r *Runtime) ProcessHeader(cols [][]byte) {
	for i, col := range cols {
		r.rw.WriteHeader(string(col))
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}
	r.Header = cols // needed for the second runtime
	r.columns = len(cols)
	r.rw.FinalizeHeader()
	r.headerPrinted = true
}

func (r *Runtime) ProcessRow(cols [][]byte) {
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

	r.rw.EOR(r.columns)
}

func (r *Runtime) SplitRow(data []byte) (columns [][]byte) {
	// we need to check first if inside the data the delimiter occurs
	// we can count the amount of occurences of the delimiter and compare
	// if that matches the amount we expect
	// amount = columns + 1

	// r.columns is set to non-zero on header creation
	if r.columns != 0 {
		count := bytes.Count(data, r.p.Delimiter)
		if count+1 != r.columns {
			columns := r.SplitUnescapedDelimiter(data)
			if len(columns) == r.columns {
				r.RowsCSV++
				return columns
			}
		}
	}

	buf := bytes.Split(data, r.p.Delimiter)

	// the SplitUnescapedDelimiter didn't work properly, use slow method
	if len(buf) > r.columns {
		// for simplicity sake we use the Go's build in CSV parser when the
		// delimiter occurs within the quotation
		// the reason for not always using it, is simply speed.
		csvReader := csv.NewReader(bytes.NewReader(data))
		csvReader.Comma = []rune(string(r.p.Delimiter))[0]

		buf := make([][]byte, 0)

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("csv reading failed: %s", err)
			}

			for _, col := range record {
				buf = append(buf, []byte(col))
			}
		}

		r.RowsCSV++

		return buf
	}

	for i := range buf {
		buf[i] = bytes.TrimPrefix(buf[i], r.p.Enclosure)
		buf[i] = bytes.TrimSuffix(buf[i], r.p.Enclosure)
	}

	return buf
}

func (r *Runtime) debug(format string, v ...interface{}) {
	if r.Debug {
		log.Printf(format, v...)
	}
}

func (r *Runtime) SplitUnescapedDelimiter(data []byte) (columns [][]byte) {
	// we need to seek through the data and find the first Enclosure and escape the delimiter

	// we first split on the delimiter and then check if two Enclosures are available
	input := bytes.Split(data, r.p.Delimiter)
	columns = make([][]byte, 0)
	buf := make([]byte, 0)
	enclosurecheck := false

	for i := range input {
		r.debug("input[i]: %s", input[i])
		r.debug("buf: %s", buf)
		r.debug("columns: %s", columns)
		count := 0
		if !r.inside {
			count = bytes.Count(input[i], r.p.Enclosure)
			if count != 0 {
				r.inside = true
			}
		}

		r.debug("inside: %v, enclosurecheck: %v", r.inside, enclosurecheck)

		// there is a case where the last buf can end with """ or similar
		// we need to check if the NEXT starts with a " if so, move buf to columns and continue happily with life
		if enclosurecheck {
			if len(input[i]) == 1 {
				// data was something like `|"foo|"|`
				buf = append(buf, r.p.Delimiter...)
			}
			if bytes.HasPrefix(input[i], r.p.Enclosure) {
				buf = bytes.TrimPrefix(buf, r.p.Enclosure)
				buf = bytes.TrimSuffix(buf, r.p.Enclosure)
				// we need to clean double enclosures to single
				buf = bytes.ReplaceAll(buf, r.doubleEnclosure, r.p.Enclosure)
				columns = append(columns, buf)
				buf = make([]byte, 0)
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
				buf = append(buf, r.p.Delimiter...)
				continue
			}
			columns = append(columns, input[i])
			continue
		}

		if len(input[i]) == 1 {
			if len(buf) != 0 && bytes.Equal(input[i], r.p.Enclosure) {
				// the delimiter appears at the end, finalize the column
				buf = append(buf, r.p.Delimiter...)

				// we need to clean double enclosures to single
				buf = bytes.ReplaceAll(buf, r.doubleEnclosure, r.p.Enclosure)
				columns = append(columns, buf)
				buf = make([]byte, 0)
				continue
			}
			buf = append(buf, input[i]...)
			continue
		}

		if !bytes.HasPrefix(input[i], r.p.Enclosure) || !bytes.HasSuffix(input[i], r.p.Enclosure) {
			if len(buf) != 0 {
				buf = append(buf, r.p.Delimiter...)
				buf = append(buf, input[i]...)

				result := bytes.HasSuffix(input[i], r.p.Enclosure) && !bytes.HasSuffix(input[i], r.doubleEnclosure)
				r.debug("bytes.HasSuffix(input[i], r.p.Enclosure): %v\n!bytes.HasSuffix(input[i], r.doubleEnclosure: %v\n statement: %v", bytes.HasSuffix(input[i], r.p.Enclosure), !bytes.HasSuffix(input[i], r.doubleEnclosure), result)
				if bytes.HasSuffix(input[i], r.p.Enclosure) && !bytes.HasSuffix(input[i], r.doubleEnclosure) {
					buf = bytes.TrimPrefix(buf, r.p.Enclosure)
					buf = bytes.TrimSuffix(buf, r.p.Enclosure)
					// we need to clean double enclosures to single
					buf = bytes.ReplaceAll(buf, r.doubleEnclosure, r.p.Enclosure)
					columns = append(columns, buf)
					buf = make([]byte, 0)
					r.inside = false
					enclosurecheck = false
				} else {
					enclosurecheck = true
				}
			} else {
				// we are missing the second Enclosure, add to the buf
				buf = append(buf, input[i]...)
				// buf = bytes.TrimPrefix(buf, r.p.Enclosure)
			}
			continue
		}

		if count%2 == 0 && bytes.HasPrefix(input[i], r.p.Enclosure) && bytes.HasSuffix(input[i], r.p.Enclosure) {
			input[i] = bytes.TrimPrefix(input[i], r.p.Enclosure)
			input[i] = bytes.TrimSuffix(input[i], r.p.Enclosure)

			// we need to clean double enclosures to single
			input[i] = bytes.ReplaceAll(input[i], r.doubleEnclosure, r.p.Enclosure)
			columns = append(columns, input[i])
			r.inside = false
			continue
		}

		// edge-case where ""| appears
		// or where count%2 != 0
		if count != 2 {
			buf = append(buf, input[i]...)
			continue
		}

		input[i] = bytes.TrimPrefix(input[i], r.p.Enclosure)
		input[i] = bytes.TrimSuffix(input[i], r.p.Enclosure)
		// we need to clean double enclosures to single
		input[i] = bytes.ReplaceAll(input[i], r.doubleEnclosure, r.p.Enclosure)
		columns = append(columns, input[i])
	}

	r.debug("columns: %d", len(columns))

	return columns
}
