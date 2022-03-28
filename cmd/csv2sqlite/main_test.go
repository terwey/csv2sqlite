package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
)

type TestData struct {
	name      string
	in        []byte
	out       []byte
	multiOut  [][]byte
	advance   int
	debug     bool
	delimiter characters.Character
	enclosure characters.Character
}

func getTestData() []TestData {
	return []TestData{
		{
			name: "header without quotes",
			in:   []byte("vddddrQ, larlfxP, umsiolA\r\nmrq1-H, mrq1-U, mrq1-M"),
			out:  []byte("vddddrQ, larlfxP, umsiolA"),
			multiOut: [][]byte{
				[]byte("vddddrQ, larlfxP, umsiolA"),
				[]byte("mrq1-H, mrq1-U, mrq1-M"),
			},
			advance:   27,
			delimiter: characters.Character_COMMA,
		},
		{
			name: "header with quotes",
			in:   []byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\"\r\nmrq1-H, mrq1-U, mrq1-M"),
			out:  []byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\""),
			multiOut: [][]byte{
				[]byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\""),
				[]byte("mrq1-H, mrq1-U, mrq1-M"),
			},
			advance:   33,
			delimiter: []byte(","),
		},
		{
			name: "data with multiple commas",
			in:   []byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\"\r\n\"mrq1-H vgei, comma\", \"mrq1-U\", \"mrq1-M\""),
			multiOut: [][]byte{
				[]byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\""),
				[]byte("\"mrq1-H vgei, comma\", \"mrq1-U\", \"mrq1-M\""),
			},
			advance:   -1,
			delimiter: []byte(","),
		},
		{
			name: "multi row",
			in:   []byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\"\r\n\"mrq1-H vgei, comma\", \"mrq1-U\", \"mrq1-M\"\r\n\"mrq2-H\", mrq2-U, mrq2-M"),
			multiOut: [][]byte{
				[]byte("\"vddddrQ\", \"larlfxP\", \"umsiolA\""),
				[]byte("\"mrq1-H vgei, comma\", \"mrq1-U\", \"mrq1-M\""),
				[]byte("\"mrq2-H\", mrq2-U, mrq2-M"),
			},
			advance:   -1,
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "comma_in_quotes.csv",
			in:   []byte("gcebs,tfxf,ixilgzh,thez,gkn\r\nJohn,Fxq,120 fvq ne.,\"Anytown, QR\",08123"),
			multiOut: [][]byte{
				[]byte("gcebs,tfxf,ixilgzh,thez,gkn"),
				[]byte("John,Fxq,120 fvq ne.,\"Anytown, QR\",08123"),
			},
			advance:   -1,
			delimiter: []byte(","),
		},
		{
			name: "empty_crlf.csv",
			in:   []byte("h,u,m\r\n1,\"\",\"\"\r\n2,3,4"),
			multiOut: [][]byte{
				[]byte("h,u,m"),
				[]byte("1,\"\",\"\""),
				[]byte("2,3,4"),
			},
			advance:   -1,
			delimiter: []byte(","),
		},
		{
			name: "escaped_quotes.csv",
			in:   []byte("h,u\r\n1,\"qu \"\"qu\"\" qu\"\r\n3,4"),
			multiOut: [][]byte{
				[]byte("h,u"),
				[]byte("1,\"qu \"\"qu\"\" qu\""),
				[]byte("3,4"),
			},
			out:       []byte("h,u"),
			delimiter: []byte(","),
			advance:   5,
		},
		{
			name: "newlines_crlf.csv",
			in:   []byte("h,u,m\r\n1,2,3\r\n\"Hfzg upon \r\nh tyid\",5,6\r\n7,8,9"),
			multiOut: [][]byte{
				[]byte("h,u,m"),
				[]byte("1,2,3"),
				[]byte("\"Hfzg upon \r\nh tyid\",5,6"),
				[]byte("7,8,9"),
			},
			advance:   7,
			out:       []byte("h,u,m"),
			delimiter: []byte(","),
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "quotes_and_newlines.csv",
			in:   []byte("h,u\r\n1,\"qu \r\n\"\"qu\"\"\r\n qu\"\r\n3,4"),
			multiOut: [][]byte{
				[]byte("h,u"),
				[]byte("1,\"qu \r\n\"\"qu\"\"\r\n qu\""),
				[]byte("3,4"),
			},
			advance:   5,
			out:       []byte("h,u"),
			delimiter: []byte(","),
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "utf8.csv",
			in:   []byte("h,u,m\r\n1,2,3\r\n4,5,ʤ"),
			multiOut: [][]byte{
				[]byte("h,u,m"),
				[]byte("1,2,3"),
				[]byte("4,5,ʤ"),
			},
			advance:   7,
			out:       []byte("h,u,m"),
			delimiter: []byte(","),
		},
	}
}

// func getTestDataFromCSVSpectrum() []TestData {
// 	var testdata []TestData

// 	base := "/home/yt/csv2sqlite/external-tests/csv-spectrum/csvs"

// 	filenames := []string{
// 		"comma_in_quotes.csv",
// 		"empty_crlf.csv",
// 		"empty.csv",
// 		"escaped_quotes.csv",
// 		"json.csv",
// 		"newlines_crlf.csv",
// 		"newlines.csv",
// 		"quotes_and_newlines.csv",
// 		"simple_crlf.csv",
// 		"simple.csv",
// 		"utf8.csv",
// 	}

// 	for _, name := range filenames {
// 		debug.Debug("kahk: %s/%s", base, name)
// 		data, err := ioutil.ReadFile(path.Join(base, name))
// 		if err != nil {
// 			debug.Debug("err: %s", err)
// 			continue
// 		}
// 		td := TestData{
// 			name: name,
// 			in:   data,
// 		}
// 		debug.Debug("testdata: \n%b", td)
// 		testdata = append(testdata, td)
// 	}

// 	return testdata
// }

func getTestDataFromBolTestfile() TestData {
	var td TestData
	var err error

	base := "/home/yt/csv2sqlite/testfiles"
	td.name = "example.com head.csv"
	td.in, err = ioutil.ReadFile(path.Join(base, "bol_input.csv"))
	td.delimiter = []byte("|")
	td.debug = false
	if err != nil {
		debug.Debug("err: %s", err)
		return TestData{}
	}

	for i := 1; i <= 10; i++ {
		buf, err := ioutil.ReadFile(path.Join(base, fmt.Sprintf("bol_out_%02d.csv", i)))
		if err != nil {
			debug.Debug("err: %s", err)
			return TestData{}
		}
		td.multiOut = append(td.multiOut, buf)
	}

	return td
}

func Test_ScanWithSplitCSV(t *testing.T) {
	testdata := getTestData()
	testdata = append(testdata, getTestDataFromBolTestfile())
	for _, td := range testdata {
		eolBytes = []byte{}
		delimBytes = td.delimiter
		encloBytes = td.enclosure
		encloLen = len(encloBytes)
		// normally the code is not re-run so we need to manually set the headerLength to zero again
		headerLength = 0
		var got [][]byte

		buffersize := 4096 * 1024
		readbuf := make([]byte, 0, buffersize)
		reader := bufio.NewReaderSize(bytes.NewReader(td.in), buffersize)

		if _, err := io.ReadFull(reader, readbuf); err != nil && err != io.ErrUnexpectedEOF {
			t.Fatal(err)
		}

		debug.DebugLog = td.debug

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(readbuf, buffersize)
		scanner.Split(splitCsvInput)
		// i := 0
		for scanner.Scan() {
			chunk := scanner.Bytes()
			// header has to be set once
			if headerLength == 0 {
				// we have to count the delimiters in the header to know for row scanning
				// how many fields we need
				headerLength = bytes.Count(chunk, delimBytes)
			}
			got = append(got, chunk)
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}

		if len(td.multiOut) != len(got) {
			for c := range td.multiOut {
				if len(got) <= c {
					t.Errorf("Test (%s): len(td.multiOut) = %d != len(got) = %d\nwanted = %s\ngot = %s",
						td.name, len(td.multiOut), len(got), td.multiOut, got)
				}
			}
		}

		if diff := cmp.Diff(td.multiOut, got); diff != "" {
			for c := range td.multiOut {
				if len(got) <= c {
					t.Fatalf("Test (%s): did not receive all data\nwanted = %s\nGot nothing", td.name, td.multiOut[c])
				}
				if d := cmp.Diff(td.multiOut[c], got[c]); d != "" {
					t.Errorf("ScanWithSplitCSV(): %s\nwant: \t%s\ngot: \t%s\ndiff: \n%s", td.name, td.multiOut[c], got[c], d)
				}
			}
			t.Errorf("ScanWithSplitCSV(): %s, diff: \n%s", td.name, diff)
		}
	}
}

func Test_SplitCSV(t *testing.T) {
	// eolBytes = []byte("\r\n")
	// delimBytes = []byte(",")
	// encloBytes = []byte(`"`)

	// reset these
	eolBytes = []byte{}
	delimBytes = []byte{}
	encloBytes = []byte{}
	encloLen = len(encloBytes)

	testdata := getTestData()

	type args struct {
		in    []byte
		atEOF bool
	}
	type testDef struct {
		name        string
		args        args
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}

	debug.DebugLog = false

	var tests []testDef
	for _, td := range testdata {
		// we do not want to test these
		if td.advance == -1 {
			continue
		}
		tests = append(tests, testDef{
			name: td.name,
			args: args{
				in:    td.in,
				atEOF: true,
			},
			wantAdvance: td.advance,
			wantToken:   td.out,
			wantErr:     false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headerLength = 0
			gotAdvance, gotToken, err := splitCsvInput(tt.args.in, tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitCsvInput(): %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("splitCsvInput(): %s gotAdvance = %d, want %d\ngotToken = \n%s\nwant = \n%s",
					tt.name, gotAdvance, tt.wantAdvance, gotToken, tt.wantToken)
			}
			if diff := cmp.Diff(gotToken, tt.wantToken); diff != "" {
				// if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("splitCsvInput(): %s \ngotToken = %s\nwant = %s\ndiff: %s", tt.name, gotToken, tt.wantToken, diff)
			}
		})
	}
}
