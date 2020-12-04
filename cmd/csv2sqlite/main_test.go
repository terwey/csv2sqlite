package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestData struct {
	name      string
	in        []byte
	out       []byte
	multiOut  [][]byte
	advance   int
	debug     bool
	delimiter []byte
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
			delimiter: []byte(","),
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
			delimiter: []byte(","),
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
			delimiter: []byte(","),
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
			delimiter: []byte(","),
		},
		{
			name: "quotes_and_newlines.csv",
			in:   []byte("h,u\r\n1,\"qu \r\n\"\"qu\"\"\r\n qu\"\r\n3,4"),
			multiOut: [][]byte{
				[]byte("h,u"),
				[]byte("1,\"qu \r\n\"\"qu\"\"\r\n qu\""),
				[]byte("3,4"),
			},
			delimiter: []byte(","),
		},
		{
			name: "utf8.csv",
			in:   []byte("h,u,m\r\n1,2,3\r\n4,5,ʤ"),
			multiOut: [][]byte{
				[]byte("h,u,m"),
				[]byte("1,2,3"),
				[]byte("4,5,ʤ"),
			},
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
// 		debug("kahk: %s/%s", base, name)
// 		data, err := ioutil.ReadFile(path.Join(base, name))
// 		if err != nil {
// 			debug("err: %s", err)
// 			continue
// 		}
// 		td := TestData{
// 			name: name,
// 			in:   data,
// 		}
// 		debug("testdata: \n%b", td)
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
	td.debug = true
	if err != nil {
		debugLog = true
		debug("err: %s", err)
		return TestData{}
	}

	for i := 1; i <= 10; i++ {
		buf, err := ioutil.ReadFile(path.Join(base, fmt.Sprintf("bol_out_%02d.csv", i)))
		if err != nil {
			debugLog = true
			debug("err: %s", err)
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
		debugLog = td.debug
		// normally the code is not re-run so we need to manually set the headerLength to zero again
		headerLength = 0
		var got [][]byte

		buffersize := 4096 * 1024
		readbuf := make([]byte, 0, buffersize)
		reader := bufio.NewReaderSize(bytes.NewReader(td.in), buffersize)

		if _, err := io.ReadFull(reader, readbuf); err != nil && err != io.ErrUnexpectedEOF {
			t.Fatal(err)
		}

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(readbuf, buffersize)
		scanner.Split(splitCsvInput)
		i := 0
		for scanner.Scan() {
			got = append(got, scanner.Bytes())
			debug("Izeda: %s", got[i])

			debug("kh.multiout: \n%s", td.multiOut[i])

			if diff := cmp.Diff(string(td.multiOut[i]), string(got[i])); diff != "" {
				t.Errorf("Test (%s): (-want +got)\n= %s", td.name, diff)
			}

			i++
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}

		if len(td.multiOut) != len(got) {
			for c := range td.multiOut {
				if len(got) <= c {
					t.Errorf("Test (%s): did not receive all data, wanted = \n%s\nGot nothing", td.name, td.multiOut[c])
				}
			}
		}
	}
}

func Test_SplitCSV(t *testing.T) {
	// eolBytes = []byte("\r\n")
	// delimBytes = []byte(",")
	// encloBytes = []byte(`"`)

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
				t.Errorf("splitCsvInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("splitCsvInput() gotAdvance = %d, want %d\ngotToken = \n%s\nwant = \n%s",
					gotAdvance, tt.wantAdvance, gotToken, tt.wantToken)
			}
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("splitCsvInput() gotToken = \n%s\nwant = \n%s", gotToken, tt.wantToken)
			}
		})
	}
}
