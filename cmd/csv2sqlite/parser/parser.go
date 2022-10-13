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
	"path"
	"strconv"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/util"
	"github.com/terwey/recordwriter/recordwriter"
	"github.com/schollz/progressbar/v3"
)

type Params struct {
	Input           string
	Output          string
	CommitRate      int
	Buffersize      int
	Sqlite          bool
	Drop            bool
	ChannelCapacity int

	// starting position of the CSV file
	// this is needed for Excel to CSV exports
	// that contain data above the header
	Start int

	Skip  int // amount of rows to skip AFTER the header
	Limit int // maximum number of rows to process

	Exclude []int

	// if the order of the incoming CSV matters the RowID
	// can be set to true and during Scan the RowID will be added
	RowID bool

	// the small SQLite files written are normally in the same
	// target as the output file, for performance reasons (or
	// storage reasons) a different workdir can be provided.
	// e.g. a tmpfs ramdisk
	Scratch string

	Characters util.CharacterConfig
}

func NewParameters() *Params {
	p := &Params{
		CommitRate: 32768,
		Buffersize: 4096,
		Input:      "",
		Output:     "",

		Limit: -1,

		Characters: util.CharacterConfig{
			Delimiter:      characters.Character_NIL,
			Enclosure:      characters.Character_NIL,
			EnclosureClean: true,
		},
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

	Rows          uint64
	RowsCSV       uint64
	headerPrinted bool
	colCount      int
	inside        bool

	Header    [][]byte
	HeaderSet chan bool

	Data        chan [][]byte
	ParserCount int

	sinput [][]byte
	// scolumns [][]byte
	sbuf []byte

	hasPrefixEnclosure bool
	hasSuffixEnclosure bool
}

var (
	RWDebugLog bool
	Overflow   bool
)

func NewRuntime(p *Params, fn bufio.SplitFunc) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		splitFunc:     fn,
		Benchmark:     false,
		ParserCount:   0,
	}

	recordwriter.DebugLog = RWDebugLog
	recordwriter.Overflow = Overflow

	r.HeaderSet = make(chan bool)
	r.Data = make(chan [][]byte, p.ChannelCapacity)

	return r
}

func NewParser(p *Params, parsercount int) *Runtime {
	r := &Runtime{
		p:             p,
		headerPrinted: false,
		Benchmark:     false,
		ParserCount:   parsercount,
	}

	return r
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
	if r.ParserCount == 0 {
		output = r.p.Output
	} else {
		if r.p.Scratch != "" {
			output = fmt.Sprintf("%s%s.%d", r.p.Scratch, path.Base(r.p.Output), r.ParserCount)
		} else {
			output = fmt.Sprintf("%s.%d", r.p.Output, r.ParserCount)
		}
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
			r.rw = recordwriter.NewSQLWriter(output, r.p.CommitRate, r.ParserCount == 0, r.p.RowID)
		}
	}

	if r.ParserCount == 0 {
		r.scanner = r.NewScanner()
	}

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

	cols, pos = r.SplitRow(0, data)
	if len(cols) == 0 {
		return []byte{}, []uint32{}
	}
	if !r.headerPrinted {
		if len(cols) == 1 {
			log.Printf("Only 1 column for header found, make sure the delimiter is set properly.")
		}
		header := util.FlatToMulti(cols, pos)

		// we gotta check if the last column in the header is set, if not drop it
		if len(header[len(header)-1]) == 0 {
			header = header[:len(header)-1]
			r.p.Characters.Excel = true
		}

		emptyCount := 1
		for i := range header {
			if len(header[i]) == 0 {
				if emptyCount > 1 {
					header[i] = []byte("empty_column__")
					header[i] = strconv.AppendInt(header[i], int64(emptyCount), 10)
				} else {
					header[i] = []byte("empty_column")
				}
				emptyCount++
			}
		}

		if r.p.RowID {
			header = append(header, []byte("rowid"))
		}
		r.ProcessHeader(header)
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
	if r.ParserCount == 0 {
		r.Header = cols // needed for the second runtime
		r.HeaderSet <- true
	}
	r.colCount = len(cols)
	r.rw.FinalizeHeader()
	r.headerPrinted = true
}

func (r *Runtime) ProcessRow(v []byte, pos []uint32) {
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
		r.Rows++
		r.rw.EORMulti()
		return
	}

	cols := util.FlatToMulti(v, pos)
	for i, col := range cols {
		if len(col) == 0 {
			r.rw.Skip()
		} else {
			if r.p.RowID && i == len(cols)-1 {
				r.rw.Write(string(col))
				continue
			}
			r.rw.WriteByte(col)
		}
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}

	r.Rows++

	r.rw.EOR(r.colCount)
	return
}

func (r *Runtime) SplitRow(colCount int, data []byte) (out []byte, pos []uint32) {
	// we need to check first if inside the data the delimiter occurs
	// we can count the amount of occurences of the delimiter and compare
	// if that matches the amount we expect
	// amount = columns + 1

	var highestCount int

	if colCount == 0 {
		colCount = r.colCount
	}

	count := util.Count(data, r.p.Characters.Delimiter) + 1 - len(r.p.Exclude)
	highestCount = count
	if colCount == count {
		out, pos = util.GenSplit(data, r.p.Characters.Delimiter, -1, r.p.Characters.Clean, r.p.Exclude)
		return
	}

	colCountExcl := util.CountOutsideEnclosure(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure)
	if colCountExcl > highestCount {
		highestCount = colCountExcl
	}

	// CountOutsideEnclosure returns the amount of delimiters, not columns so add 1
	if colCount == colCountExcl+1 {
		o, p := util.ColSplitFlat(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure, -1, r.p.Characters.Clean)

		if len(p) > highestCount {
			highestCount = len(p)
		}
		if len(p) == colCount {
			return o, p
		}
	}

	// r.colCount is set to non-zero on header creation
	// TODO: fix that the header gets parsed by our code too
	if r.colCount != 0 {
		// count := Count(data, r.p.Delimiter)
		// if the count is correct we can just simply do a normal split and skip this logic
		if count > colCount {
			o, p := r.ParseUnescapedDelimiter(util.ColSplit(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure, -1))
			if len(p) > highestCount {
				highestCount = len(p)
			}
			if len(p) == colCount {
				return o, p
			}
		}
	}

	// the SplitUnescapedDelimiter didn't work properly, use slow method
	// NOTE: it does not support other enclosures that QUOTATION_MARK
	if count > colCount {
		// for simplicity sake we use the Go's build in CSV parser when the
		// delimiter occurs within the quotation
		// the reason for not always using it, is simply speed.
		csvReader := csv.NewReader(bytes.NewReader(data))
		csvReader.Comma = r.p.Characters.Delimiter.Rune()

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				if debug.DebugLog {
					time.Sleep(2 * time.Second)

				}
				log.Printf("csv reading failed: %s, data: \n%s", err, data)
				panic(err)
				return nil, nil
			}

			for i, col := range record {
				if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
					continue
				}
				// we do not have to call a clean here as the CSV parser of Go already did so
				out = append(out, col...)
				pos = append(pos, uint32(len(col)))
			}
		}

		if r.colCount != 0 {
			debug.LogReader(data, count, colCount)
			r.RowsCSV++
		}

		if len(pos) > highestCount {
			highestCount = len(pos)
		}
		if len(pos) == colCount {
			return out, pos
		}
	}

	return r.SplitRow(highestCount, data)
}

func (r *Runtime) ParseUnescapedDelimiter(input [][]byte) (data []byte, pos []uint32) {
	cleanFunc := r.p.Characters.Clean
	// NOTE: this method  doesn't work on optional enclosures.

	// we need to seek through the data and find the first Enclosure and escape the delimiter

	// we first split on the delimiter and then check if two Enclosures are available
	// r.sinput = bytes.Split(data, r.p.Delimiter)
	// r.sinput = splitData
	// r.scolumns = make([][]byte, 0)
	r.sbuf = make([]byte, 0)
	enclosurecheck := false

	// debug.DebugColumns(input)

	for i := range input {
		// debug.Debug("input[i]: >>>%s<<<", input[i])
		// debug.Debug("buf: %s", r.sbuf)
		// debug.DebugColumns(util.FlatToMulti(data, pos))

		// if debug.DebugLog {
		// 	debug.BreakMe()
		// }

		count := 0
		counted := false

		if !r.inside {
			count = util.Count(input[i], r.p.Characters.Enclosure)
			counted = true
			if count != 0 {
				r.inside = true
			}
		}

		// debug.Debug("inside: %v, enclosurecheck: %v", r.inside, enclosurecheck)

		// there is a case where the last buf can end with """ or similar
		// we need to check if the NEXT starts with a " if so, move buf to columns and continue happily with life
		if enclosurecheck {
			if len(input[i]) == 1 {
				// data was something like `|"foo|"|`
				r.sbuf = append(r.sbuf, r.p.Characters.Delimiter...)
				// add the "empty" data, anti-slipmat func ;)
				// check if input is r.p.Enclosure
				if !bytes.Equal(input[i], r.p.Characters.Enclosure) {
					r.sbuf = append(r.sbuf, input[i]...)
				}
			}
			if bytes.HasPrefix(input[i], r.p.Characters.Enclosure) {
				r.sbuf = cleanFunc(r.sbuf, false, false)
				if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
				} else {
					data = append(data, r.sbuf...)
					pos = append(pos, uint32(len(r.sbuf)))
				}
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
				r.sbuf = append(r.sbuf, r.p.Characters.Delimiter...)
				continue
			}
			if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
				continue
			} else {
				input[i] = util.ReplaceAll(input[i], characters.Character_CARRIAGE_RETURN_LINE_FEED, characters.Character_LINE_FEED)
				input[i] = util.ReplaceAll(input[i], characters.Character_CARRIAGE_RETURN, characters.Character_LINE_FEED)
				data = append(data, input[i]...)
				pos = append(pos, uint32(len(input[i])))
			}
			continue
		}

		if len(input[i]) == 1 {
			if len(r.sbuf) != 0 && bytes.Equal(input[i], r.p.Characters.Enclosure) {
				if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
				} else {
					// the delimiter appears at the end, finalize the column
					r.sbuf = append(r.sbuf, r.p.Characters.Delimiter...)

					// TODO: it's unsure if here the suffix should be removed too
					// the tests all work so should be fine, however originally
					// here there was no trimsuffix (which is part of cleanFunc)
					r.sbuf = cleanFunc(r.sbuf, false, false)

					data = append(data, r.sbuf...)
					pos = append(pos, uint32(len(r.sbuf)))
				}
				r.sbuf = make([]byte, 0)
				continue
			}

			// in the case it is NOT an Enclosure, place a delimiter and append the buffer
			if !bytes.Equal(input[i], r.p.Characters.Enclosure) {
				r.sbuf = append(r.sbuf, r.p.Characters.Delimiter...)
			}
			r.sbuf = append(r.sbuf, input[i]...)

			continue
		}

		// if we counted the enclosures before we don't have to check if there is a prefix or suffix
		if counted && count == 0 {
			r.hasPrefixEnclosure = false
			r.hasSuffixEnclosure = false
		} else {
			r.hasPrefixEnclosure = bytes.HasPrefix(input[i], r.p.Characters.Enclosure)
			r.hasSuffixEnclosure = bytes.HasSuffix(input[i], r.p.Characters.Enclosure)
		}

		// debug.Debug("r.hasPrefixEnclosure: %v, r.hasSuffixEnclosure: %v, count: %d", r.hasPrefixEnclosure, r.hasSuffixEnclosure, count)

		if !r.hasPrefixEnclosure || !r.hasSuffixEnclosure {
			if len(r.sbuf) != 0 {
				r.sbuf = append(r.sbuf, r.p.Characters.Delimiter...)
				r.sbuf = append(r.sbuf, input[i]...)

				if r.hasSuffixEnclosure && !bytes.HasSuffix(input[i], r.p.Characters.DoubleEnclosure) {
					if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
					} else {
						// we know for sure that the suffix exists
						r.sbuf = cleanFunc(r.sbuf, false, true)
						data = append(data, r.sbuf...)
						pos = append(pos, uint32(len(r.sbuf)))
					}
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
			if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
			} else {
				// we know for sure it has a prefix and suffix
				input[i] = cleanFunc(input[i], true, true)
				data = append(data, input[i]...)
				pos = append(pos, uint32(len(input[i])))
			}
			r.inside = false
			continue
		}

		// edge-case where ""| appears
		// or where count%2 != 0
		if count != 2 {
			r.sbuf = append(r.sbuf, input[i]...)
			continue
		}

		if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
		} else {
			input[i] = cleanFunc(input[i], false, false)
			data = append(data, input[i]...)
			pos = append(pos, uint32(len(input[i])))
		}
	}

	return
}

func (r *Runtime) Parse(wg *sync.WaitGroup, chunkCount int, pv *progressbar.ProgressBar) {
	var err error
	// var buf []byte
	var delimSet bool
	var enclosureSet bool
	if r.p.Characters.Delimiter != nil {
		delimSet = true
	}
	if r.p.Characters.Enclosure != nil {
		enclosureSet = true
	}
	scanner := r.GetScanner()

	var chunks [][]byte
	chunks = make([][]byte, 0, chunkCount)
	var chunk []byte

	var enclDetectCount int

	// used for p.Start
	startCounter := 0

	// used for p.Skip
	skipCounter := 0

	count := 0

	// amount of bytes the chunksize is occupying
	var chunksSize int
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			close(r.Data)
			wg.Done()
			log.Fatal(err)
		}

		if r.p.Start > 0 && startCounter < r.p.Start {
			startCounter++
			if startCounter <= r.p.Start {
				continue
			}
		}

		if r.p.Skip > 0 && count != 0 && skipCounter < r.p.Skip {
			skipCounter++
			if skipCounter <= r.p.Skip {
				continue
			}
		}

		if count != 0 && r.p.Limit == count-1 {
			break
		}

		size := len(scanner.Bytes())
		chunk = make([]byte, size)
		copy(chunk, scanner.Bytes())
		chunksSize += size
		if r.p.Characters.Excel {
			chunk = bytes.TrimSuffix(chunk, r.p.Characters.Delimiter)
		}
		// if no delimiter was set we detect it once
		if !delimSet {
			// remove \uFEFF used as BOM
			chunk = bytes.TrimPrefix(chunk, characters.BOM_UTF16_BE)
			// remove \uEFBBBF used as BOM
			chunk = bytes.TrimPrefix(chunk, characters.BOM_UTF8)
			r.p.Characters.SetDelimiter(characters.DetectDelimiter(chunk))
		}

		if !enclosureSet {
			prevCount := enclDetectCount
			var encl characters.Character
			encl, enclDetectCount = characters.DetectEnclosureWithDelimiter(chunk, r.p.Characters.Delimiter)
			if prevCount < enclDetectCount || enclDetectCount != -1 {
				r.p.Characters.SetEnclosure(encl)
			}
			if r.p.Characters.DelimiterCount != 0 {
				enclosureSet = true
			}
			if enclosureSet && enclDetectCount == 0 {
				log.Fatal("Enclosure cannot be detected")
			}
		}

		// header has to be set once
		if r.p.Characters.DelimiterCount == 0 {
			// we have to count the delimiters in the header to know for row scanning
			// how many fields we need
			r.p.Characters.DelimiterCount = bytes.Count(chunk, r.p.Characters.Delimiter)

			if r.p.Characters.DelimiterCount != 0 {
				delimSet = true
			}
		}

		// log.Printf("chunk: ---\n%s\n---\n", chunk)

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

		if r.p.RowID {
			// we have to make this a user configurable switch
			chunk = append(chunk, r.p.Characters.Delimiter...)
			chunk = append(chunk, r.p.Characters.Enclosure...)
			chunk = append(chunk, []byte(strconv.Itoa(count))...)
			chunk = append(chunk, r.p.Characters.Enclosure...)
		}
		chunks = append(chunks, chunk)
		count++
	}

	if err = scanner.Err(); err != nil {
		close(r.Data)
		wg.Done()
		log.Fatal(err)
	}

	if len(chunks) != 0 {
		// send over the last chunks
		r.Data <- chunks
		chunks = make([][]byte, 0)
	}

	close(r.Data)
	pv.Finish()
	wg.Done()
}
