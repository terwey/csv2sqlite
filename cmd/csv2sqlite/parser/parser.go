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
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
	"github.com/terwey/csv2sqlite/util"
	"github.com/terwey/recordwriter/recordwriter"
	"github.com/schollz/progressbar/v3"
)

type RowCount struct {
	Inserted        uint64
	GoParser        uint64
	Dropped         uint64
	GoParserDropped uint64
}

type Params struct {
	Input           string
	Output          string
	CommitRate      int
	Buffersize      int
	Sqlite          bool
	Drop            bool
	ChannelCapacity int

	// When set to true the CSV does not have newlines
	// (CR, LF or CRLF) inside records
	NewlineDelimited bool

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

type Chunk struct {
	data       []byte
	delimCount int
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

	RowCounter    RowCount
	headerPrinted bool
	colCount      int
	inside        bool

	Header    [][]byte
	HeaderSet chan bool

	ProcessedHeader []string

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
				log.Fatalf("Runtime: could not flush writer: %s", err)
			}
		}

		if len(r.p.Output) != 0 {
			if err = r.fout.Sync(); err != nil {
				log.Fatalf("Runtime: could not sync writer: %s", err)
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
		log.Fatalf("r.NewScanner(): %s", err)
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

	// replace CRLN/LN with space in header
	if !r.headerPrinted {
		data = bytes.ReplaceAll(data, characters.Character_CARRIAGE_RETURN_LINE_FEED, []byte(" "))
		data = bytes.ReplaceAll(data, characters.Character_LINE_FEED, []byte(" "))
	}
	cols, pos = r.SplitRow(0, data, false)
	if len(cols) == 0 {
		return []byte{}, []uint32{}
	}
	// only rowid has been set
	if r.p.RowID && len(cols) == 1 {
		return []byte{}, []uint32{}
	}
	if !r.headerPrinted {
		if len(cols) == 1 {
			log.Printf("Only 1 column for header found, make sure the delimiter is set properly.")
		}
		var header [][]byte
		if r.p.RowID {
			header = append(header, []byte("rowid"))
		}
		header = append(header, util.FlatToMulti(cols, pos)...)

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

		r.ProcessHeader(header)
		// we should not output the header to the sqlite file, return nothing as it'll be skipped
		// return []byte{}, []uint32{}
		return cols, pos
	}

	return cols, pos
}

func (r *Runtime) ProcessHeader(cols [][]byte) {
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
		r.ProcessedHeader = r.rw.FinalizeHeader()
		r.HeaderSet <- true
	} else {
		r.ProcessedHeader = r.rw.FinalizeHeader()
	}
	r.colCount = len(cols)
	r.headerPrinted = true
}

func (r *Runtime) ProcessRow(v []byte, pos []uint32) {
	if len(v) == 0 {
		return
	}

	if r.p.RowID && PosSliceContainsNoValues(pos[1:]) {
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

	cols := util.FlatToMulti(v, pos)
	for i, col := range cols {
		if len(col) == 0 {
			r.rw.Skip()
		} else {
			if r.p.RowID && i == 0 {
				r.rw.Write(string(col))
				continue
			}
			r.rw.WriteByte(col)
		}
		if i != len(cols)-1 {
			r.rw.Next()
		}
	}

	r.RowCounter.Inserted++

	r.rw.EOR(r.colCount)
	return
}

func PosSliceContainsNoValues(pos []uint32) bool {
	for _, v := range pos {
		if v != 0 {
			return false
		}
	}

	return true
}

// SplitRow expects a byte slice appended with two counts appended to it in
// binary encoded format in 32bit, e.g.  binary.LittleEndian.PutUint32(data, 0)
// unless self is set to true
func (r *Runtime) SplitRow(colCount int, data []byte, self bool) (out []byte, pos []uint32) {
	// we need to check first if inside the data the delimiter occurs
	// we can count the amount of occurences of the delimiter and compare
	// if that matches the amount we expect
	// amount = columns + 1

	if debug.DebugLog {
		debug.LogData("[SplitRow] data", data, 4096)
		debug.BreakMe()
	}

	if colCount == 0 {
		colCount = r.colCount
	}

	lowestCount := colCount

	var count, outsideCount int
	provided := false

	// we appended the delimiterCount to the chunk
	// we must extract it and remove it, it's 4 bytes long

	if !self {
		// the incoming count is a delimiterCount and _not_ a columnCount

		count, outsideCount = util.DecodeCounts(data[len(data)-8:])

		// increment incoming by one due to previous count looked at delimiters
		// and now we look at columns
		data = data[:len(data)-8]
	}
	if count != 0 || outsideCount != 0 {
		provided = true
	}

	if !provided {
		count = util.Count(data, r.p.Characters.Delimiter)
	}

	if r.p.Characters.Excel {
		count--
	}

	if colCount == count+1-len(r.p.Exclude) {
		out, pos = util.GenSplit(data, r.p.Characters.Delimiter, -1, r.p.Characters.Clean, r.p.Exclude)
		if debug.DebugLog {
			debug.Debug("[SplitRow] GenSplit found %d columns (r.colCount: %d, colCount: %d)", len(pos), r.colCount, colCount)
		}
		return
	}

	if debug.DebugLog {
		debug.BreakMe()
	}

	// no point counting outside enclosure when it doesn't exist
	if len(r.p.Characters.Enclosure) != 0 {
		if !provided {
			outsideCount = util.CountOutsideEnclosure(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure)
			if outsideCount != 0 && outsideCount < lowestCount {
				lowestCount = outsideCount
			}
		}

		if r.p.Characters.Excel {
			outsideCount--
		}

		// CountOutsideEnclosure returns the amount of delimiters, not columns so add 1
		if colCount == outsideCount+1-len(r.p.Exclude) {
			// we should pass the result from `util.Count` to this func for a speedup
			n := -1
			if count > 0 {
				n = count
			}
			out, pos = util.ColSplitFlat(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure, n, r.p.Characters.Clean)
			if debug.DebugLog {
				debug.Debug("[SplitRow] ColSplitFlat found %d columns (r.colCount: %d, colCount: %d)", len(pos), r.colCount, colCount)
			}
			if len(pos) == colCount {
				return out, pos
			}
			if len(pos) != 0 && len(pos) < lowestCount {
				lowestCount = len(pos)
			}
		}
	} else {
		if count != 0 && count < lowestCount {
			lowestCount = count
		}
	}

	// r.colCount is set to non-zero on header creation
	// TODO: fix that the header gets parsed by our code too
	if r.colCount != 0 {
		// count := Count(data, r.p.Delimiter)
		// if the count is correct we can just simply do a normal split and skip this logic
		if count > colCount {
			out, pos = r.ParseUnescapedDelimiter(util.ColSplit(data, r.p.Characters.Delimiter, r.p.Characters.Enclosure, -1))
			if debug.DebugLog {
				debug.Debug("[SplitRow] ParseUnescapedDelimiter found %d columns (r.colCount: %d, colCount: %d)", len(pos), r.colCount, colCount)
			}
			if len(pos) == colCount {
				return out, pos
			}
			if len(pos) != 0 && len(pos) < lowestCount {
				lowestCount = len(pos)
			}
		}
	}

	out = make([]byte, 0)
	pos = make([]uint32, 0)

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
				// last hail mary for broken CSVs, we use a gensplit as we do not care anymore about consistency for these
				o, p := util.GenSplit(data, r.p.Characters.Delimiter, -1, r.p.Characters.Clean, r.p.Exclude)
				if len(o) != 0 {
					if debug.EnableLogDropped {
						debug.LogDropped("GoParser could have dropped (delimCounted: %d): \n%s\nfound: \n%s", bytes.Count(data, r.p.Characters.Delimiter), data, util.FlatToMulti(o, p))
					}
					return o, p
				}
				if debug.EnableLogDropped {
					debug.LogDropped("GoParser dropped (delimCounted: %d): \n%s", bytes.Count(data, r.p.Characters.Delimiter), data)
				}
				r.RowCounter.GoParserDropped++
				return nil, nil
			}

			for i, col := range record {
				if r.p.Exclude != nil && slices.Contains(r.p.Exclude, i) {
					continue
				}

				// we have to clean the enclosure if it's not a double quote
				if !bytes.Equal(r.p.Characters.Enclosure, characters.Character_QUOTATION_MARK) {
					buf := r.p.Characters.Clean([]byte(col), false, false)
					col = string(buf)
				} else {
					// we do not have to call a cleanFunc here as the CSV parser of Go already did so
					// we do have to clean the \r\n
					col = strings.ReplaceAll(col, characters.Character_CARRIAGE_RETURN_LINE_FEED.String(), characters.Character_LINE_FEED.String())
					col = strings.ReplaceAll(col, characters.Character_CARRIAGE_RETURN.String(), characters.Character_LINE_FEED.String())
				}
				out = append(out, col...)
				pos = append(pos, uint32(len(col)))
			}
		}

		if r.colCount != 0 {
			if debug.DebugLog {
				debug.LogReader(data, count, colCount)
			}
			r.RowCounter.GoParser++
		}

		if len(pos) != 0 && len(pos) < lowestCount {
			lowestCount = len(pos)
		}
		if len(pos) == colCount {
			return out, pos
		}

		// first pass for the header
		if r.colCount == 0 {
			return out, pos
		}
	}

	if !self {
		if debug.DebugLog {
			debug.Debug("looping (sep: %s delim: %s) (r.colCount: %d lowestCount: %d): \n%s\n", r.p.Characters.Enclosure,
				r.p.Characters.Delimiter, r.colCount, lowestCount, data)
		}
		return r.SplitRow(lowestCount+1, data, true)
	}

	if debug.DebugLog {
		if len(out) == 0 {
			debug.Debug("dropping (sep: %s delim: %s) (r.colCount: %d lowestCount: %d): \n%s\n", r.p.Characters.Enclosure,
				r.p.Characters.Delimiter, r.colCount, lowestCount, data)
		} else {
			debug.Debug("returning: \n%v\n%v\n", out, pos)
		}
	}

	if len(out) == 0 {
		// last hail mary for broken CSVs
		o, p := util.GenSplit(data, r.p.Characters.Delimiter, -1, r.p.Characters.Clean, r.p.Exclude)
		if len(o) != 0 {
			if debug.EnableLogDropped {
				debug.LogDropped("SplitRow could have dropped (delimCounted: %d): \n%s\nfound: \n%s", bytes.Count(data, r.p.Characters.Delimiter), data, util.FlatToMulti(o, p))
			}
			return o, p
		}
		if debug.EnableLogDropped {
			debug.LogDropped("SplitRow dropped (delimCounted: %d): \n%s", bytes.Count(data, r.p.Characters.Delimiter), data)
		}
		r.RowCounter.Dropped++
	}

	return out, pos
}

func (r *Runtime) ParseUnescapedDelimiter(input [][]byte) (data []byte, pos []uint32) {
	cleanFunc := r.p.Characters.Clean
	// NOTE: this method  doesn't work on optional enclosures.

	// we need to seek through the data and find the first Enclosure and escape the delimiter

	// we first split on the delimiter and then check if two Enclosures are available
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
	if r.p.Characters.EnclosureIsEmpty {
		enclosureSet = true
	}
	if r.p.Start != 0 {
		r.p.Characters.ReturnEarly = true
	}

	scanner := r.GetScanner()

	var chunks [][]byte
	chunks = make([][]byte, 0, chunkCount)
	var chunk []byte

	var enclDetectCount int

	// used for p.Start
	startCounter := 1

	// used for p.Skip
	skipCounter := 0

	count := 0

	// used for storing the zeroed 8 delimiterCount
	bdcZero := util.EncodeCounts(0, 0)

	var rowIdBuf []byte

	// amount of bytes the chunksize is occupying
	var chunksSize int
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			close(r.Data)
			wg.Done()
			log.Fatalf("Runtime: scanner error during loop: %s", err)
		}

		size := len(scanner.Bytes())
		if size == 0 {
			continue
		}

		r.p.Characters.DelimCounted = 0

		if r.p.Start > 0 && startCounter < r.p.Start {
			startCounter++
			if startCounter <= r.p.Start {
				if startCounter == r.p.Start {
					r.p.Characters.ReturnEarly = false
				}
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

		// make chunk large enough to also fit the delimCount as bytes
		chunk = make([]byte, size, size+8)
		copy(chunk, scanner.Bytes())

		chunksSize += size
		if r.p.Characters.Excel {
			chunk = bytes.TrimSuffix(chunk, r.p.Characters.Delimiter)
		}
		// only clean the PHP characters in the scanner when using Nd
		if r.p.NewlineDelimited && r.p.Characters.PHP {
			chunk = r.p.Characters.ReplaceBackslashEscapedEnclosure(chunk)
		}
		// if no delimiter was set we detect it once
		if !delimSet {
			r.p.Characters.SetDelimiter(characters.DetectDelimiter(chunk))
		}

		if !enclosureSet {
			prevCount := enclDetectCount
			var encl characters.Character
			encl, enclDetectCount, _ = characters.DetectEnclosureWithDelimiter(chunk, r.p.Characters.Delimiter)
			if prevCount < enclDetectCount || enclDetectCount != -1 {
				if debug.DebugLog {
					debug.Debug("prevCount (%d) < enclDetectCount (%d): %s < %s", prevCount, enclDetectCount, r.p.Characters.Enclosure, encl)
				}
				r.p.Characters.SetEnclosure(encl)
			}
			if r.p.Characters.DelimiterCount != 0 {
				enclosureSet = true
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

		// we have to process a header once
		if count == 0 {
			// a trim would have been sufficient but turns out some files have multiple
			// remove \uFEFF used as BOM
			chunk = util.ReplaceAll(chunk, characters.BOM_UTF16_BE, []byte{})
			// remove \uEFBBBF used as BOM
			chunk = util.ReplaceAll(chunk, characters.BOM_UTF8, []byte{})

			// some files have empty cols before the header, we should drop them, slight speed impact but worth the fix
			if len(util.ReplaceAll(chunk, r.p.Characters.Delimiter, characters.Character_EMPTY)) == 0 {
				continue
			}

			chunk = append(chunk, bdcZero...)
			c, _ := r.ParseRow(chunk)
			if len(c) != 0 {
				count++
			}
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
			rowIdBuf = append(rowIdBuf, []byte(strconv.Itoa(count))...)
			rowIdBuf = append(rowIdBuf, r.p.Characters.Delimiter...)
			chunk = append(rowIdBuf, chunk...)
			rowIdBuf = make([]byte, 0)
		}

		if r.p.Characters.DelimCounted != 0 || r.p.Characters.DelimOutsideCounted != 0 {
			// when we use a rowid we must increase it with one as we added it
			if r.p.RowID {
				if r.p.Characters.DelimCounted != 0 {
					r.p.Characters.DelimCounted++
				}
				if r.p.Characters.DelimOutsideCounted != 0 {
					r.p.Characters.DelimOutsideCounted++
				}
			}
			chunk = append(chunk, util.EncodeCounts(uint32(r.p.Characters.DelimCounted), uint32(r.p.Characters.DelimOutsideCounted))...)
			// reset the count
			r.p.Characters.DelimCounted = 0
			r.p.Characters.DelimOutsideCounted = 0
		} else {
			chunk = append(chunk, bdcZero...)
		}

		chunks = append(chunks, chunk)
		count++
	}

	if err = scanner.Err(); err != nil {
		close(r.Data)
		wg.Done()
		if err == bufio.ErrTooLong {
			fmt.Fprintf(os.Stderr, "This error could also occur due to incorrectly configured delimiter (%s) or enclosure (%s)\n", r.p.Characters.Delimiter, r.p.Characters.Enclosure)
			fmt.Fprintf(os.Stderr, "Scanner error (%s), try increasing the `--buffer-size` (current value: %d kb)\n", err, r.p.Buffersize/1024)
			os.Exit(1)
		} else {
			log.Fatalf("Runtime: scanner error after loop: %s", err)
		}
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
