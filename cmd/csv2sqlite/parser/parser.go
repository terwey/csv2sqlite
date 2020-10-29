// SPDX-License-Identifier: MIT
// Copyright (c) 2020 Yorick Terweijden

package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/terwey/recordwriter"
)

const nonChar string = "\uFFFF"

type Params struct {
	Input      string
	Output     string
	FlushSize  int
	Buffersize int
	Sqlite     bool
	Delimiter  string
	Quote      string
}

func NewParameters() *Params {
	p := &Params{
		FlushSize:  32768,
		Buffersize: 4096,
		Input:      "",
		Output:     "",
		Delimiter:  "|",
		Quote:      `"`,
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

	Rows          uint64
	RowsCSV       uint64
	headerPrinted bool
	columns       int
}

func NewRuntime(p *Params, fn bufio.SplitFunc) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		splitFunc:     fn,
		Benchmark:     false,
	}

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

	if !r.p.Sqlite {
		if len(r.p.Output) == 0 {
			r.w = os.Stdout
			if r.Benchmark {
				r.w = bufio.NewWriterSize(bytes.NewBuffer([]byte{}), r.p.FlushSize)
			}
		} else {
			r.fout, err = os.Create(r.p.Output)
			if err != nil {
				log.Fatalf("Cannot open file for writing: %s", err)
			}
			r.w = bufio.NewWriterSize(r.fout, r.p.FlushSize)
		}

		r.rw = recordwriter.NewCSVWriter(r.w)
	} else {
		r.rw = recordwriter.NewSQLWriter(r.p.Output, r.p.FlushSize)
	}

	r.scanner = r.NewScanner()

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

func (r *Runtime) ProcessRow(data []byte) {
	cols := r.SplitRow(data)
	if len(cols) == 0 {
		return
	}
	if !r.headerPrinted {
		for i, col := range cols {
			r.rw.WriteHeader(string(col))
			if i != len(cols)-1 {
				r.rw.Next()
			}
		}
		r.columns = len(cols)
		r.rw.FinalizeHeader()
		r.headerPrinted = true
		return
	}

	for i, col := range cols {
		if len(col) == 0 {
			r.rw.Skip()
		} else {
			r.rw.Write(string(col))
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

	// r.columns is set to zero on the header creation
	if r.columns != 0 {
		count := bytes.Count(data, []byte(r.p.Delimiter))
		if count+1 != r.columns {
			data = r.EscapeDelimiter(data)
			r.RowsCSV++
		}
	}

	buf := bytes.Split(data, []byte(r.p.Delimiter))

	// the EscapeDelimiter didn't work properly, use slow method
	if len(buf) > r.columns {
		// for simplicity sake we use the Go's build in CSV parser when the
		// delimiter occurs within the quotation
		// the reason for not always using it, is simply speed.
		csvReader := csv.NewReader(bytes.NewReader(data))
		csvReader.Comma = []rune(r.p.Delimiter)[0]

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

		// r.RowsCSV++

		return buf
	}

	for i := range buf {
		buf[i] = bytes.TrimPrefix(buf[i], []byte(r.p.Quote))
		buf[i] = bytes.TrimSuffix(buf[i], []byte(r.p.Quote))
		buf[i] = bytes.ReplaceAll(buf[i], []byte(nonChar), []byte(r.p.Delimiter))
	}

	return buf
}

func (r *Runtime) EscapeDelimiter(data []byte) []byte {
	// we need to seek through the data and find the first quote and escape the delimiter
	// log.Printf("unescaped: \t%s", data)

	// we first split on the delimiter and then check if two quotes are available
	input := bytes.Split(data, []byte(r.p.Delimiter))
	out := make([]byte, 0)
	buf := make([]byte, 0)

	for i := range input {

		// log.Printf("input[i]: %s", input[i])
		// log.Printf("buf: %s", buf)
		// log.Printf("out: %s", out)

		// // fails on this
		// if string(input[i]) == `"Installatiebranche T-stuk no.130 verzinkt 1"" x 3/4"" x 1/2""` {
		// 	log.Printf("sup")
		// }

		if len(input[i]) == 1 {
			buf = append(buf, input[i]...)
			continue
		}

		count := bytes.Count(input[i], []byte(r.p.Quote))

		if count%2 == 0 && bytes.HasPrefix(input[i], []byte(r.p.Quote)) && bytes.HasSuffix(input[i], []byte(r.p.Quote)) {
			out = append(out, input[i]...)
			if i != len(input)-1 {
				out = append(out, []byte(r.p.Delimiter)...)
			}
			continue
		}

		if !bytes.HasPrefix(input[i], []byte(r.p.Quote)) || !bytes.HasSuffix(input[i], []byte(r.p.Quote)) {
			if len(buf) != 0 {
				buf = append(buf, []byte(nonChar)...)
				buf = append(buf, input[i]...)

				if bytes.HasSuffix(input[i], []byte(r.p.Quote)) {
					out = append(out, buf...)
					out = append(out, []byte(r.p.Delimiter)...)
					buf = make([]byte, 0)
				}
			} else {
				// we are missing the second quote, add to the buf
				buf = append(buf, input[i]...)
			}
			continue
		}

		// count :=

		// edge-case where ""| appears
		// or where count%2 != 0
		if count != 2 {
			buf = append(buf, input[i]...)
			continue
		}

		out = append(out, input[i]...)
		if i != len(input)-1 {
			out = append(out, []byte(r.p.Delimiter)...)
		}
		// }
	}

	// log.Printf("escaped: \t%s", out)

	return out
}
