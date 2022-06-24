package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
)

func getmulti() [][]byte {
	var multi [][]byte
	multi = append(multi, []byte(`hello awblx`))
	multi = append(multi, []byte(`kuba nw h`))
	multi = append(multi, []byte(`ewtud-mqbxsgpgegs slice`))

	return multi
}

func getflat() (flat []byte, pos []uint32) {
	flat, pos = FlatAppend(flat, []byte(`hello awblx`), pos)
	flat, pos = FlatAppend(flat, []byte(`kuba nw h`), pos)
	flat, pos = FlatAppend(flat, []byte(`ewtud-mqbxsgpgegs slice`), pos)

	return
}

func Test_MultiToFlat(t *testing.T) {
	multi := getmulti()

	log.Printf("multi \n%s", multi)

	out, pos := MultiToFlat(multi)

	log.Printf("out: %v\npos: %v", out, pos)
}

func Test_Convert(t *testing.T) {
	want := getmulti()
	got := FlatToMulti(getflat())
	// log.Printf("%v", conv)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FlatToMulti() mismatch (-want +got):\n%s", diff)
	}

	want2, pos_want2 := getflat()
	got2, pos_got2 := MultiToFlat(want)

	if diff := cmp.Diff(want2, got2); diff != "" {
		t.Errorf("MultiToFlat() mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(pos_want2, pos_got2); diff != "" {
		t.Errorf("MultiToFlat() mismatch (-want +got):\n%s", diff)
	}
}

var resultMulti [][]byte

func Benchmark_FlatToMulti(b *testing.B) {
	var out [][]byte
	flat, pos := getflat()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out = FlatToMulti(flat, pos)
	}

	resultMulti = out
}

var resultFlat []byte
var resultFlatPos []uint32

func Benchmark_MultiToFlat(b *testing.B) {
	var out []byte
	var pos []uint32
	multi := getmulti()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, pos = MultiToFlat(multi)
	}

	resultFlat = out
	resultFlatPos = pos
}

func Test_GenSplit(t *testing.T) {
	foo := []byte("abc|bcd|cde|def")
	want := []byte("abcbcdcdedef")
	got, pos := GenSplit(foo, []byte("|"), -1, NoClean, true)

	log.Printf("pos: %v\n got: %s\n", pos, got)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GenSplit() mismatch (-want +got):\n%s", diff)
	}

	data := FlatToMulti(got, pos)
	for i := range data {
		log.Printf("data: %s", data[i])
	}

	want2 := [][]byte{[]byte("abc"), []byte("bcd"), []byte("cde"), []byte("def")}

	if diff := cmp.Diff(want2, data); diff != "" {
		t.Errorf("GenSplit() mismatch (-want +got):\n%s", diff)
	}
}

type ColSplitTestArgs struct {
	s         []byte
	sep       characters.Character
	enclosure characters.Character
	n         int
}

type ColSplitTest struct {
	name  string
	args  ColSplitTestArgs
	want  [][]byte
	debug bool
}

func ColSplitTests() []ColSplitTest {
	tests := []ColSplitTest{
		{
			name: "Two columns",
			args: ColSplitTestArgs{
				[]byte(`"column1","column2"`),
				characters.Character_COMMA,
				characters.Character_QUOTATION_MARK,
				-1,
			},
			want: [][]byte{
				[]byte(`"column1"`),
				[]byte(`"column2"`),
			},
			debug: false,
		},
		{
			name: "Two columns with comma in column2",
			args: ColSplitTestArgs{
				[]byte(`"column1","column2,etvni column 2"`),
				characters.Character_COMMA,
				characters.Character_QUOTATION_MARK,
				-1,
			},
			want: [][]byte{
				[]byte(`"column1"`),
				[]byte(`"column2,etvni column 2"`),
			},
			debug: false,
		},
		{
			name: "Two columns with comma in column1",
			args: ColSplitTestArgs{
				[]byte(`"column1,etvni column 1","column2"`),
				characters.Character_COMMA,
				characters.Character_QUOTATION_MARK,
				-1,
			},
			want: [][]byte{
				[]byte(`"column1,etvni column 1"`),
				[]byte(`"column2"`),
			},
			debug: false,
		},
		{
			name: "target",
			args: ColSplitTestArgs{
				[]byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`),
				characters.Character_COMMA,
				characters.Character_QUOTATION_MARK,
				-1,
			},
			want: [][]byte{
				[]byte(`"2189"`),
				[]byte(`"97998848"`),
				[]byte(`"007-04-8133"`),
				[]byte(`"FCFO"`),
				[]byte(`"DC_FZXMI"`),
				[]byte(`"2022-03-11T21:46:09N"`),
				[]byte(`"9999-12-31T00:00:00N"`),
				[]byte(`"9.89"`),
				[]byte(`"9.89"`),
				[]byte(`""`),
				[]byte(`"[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`),
			}, debug: true,
		},
	}

	return tests
}

func TestColSplit(t *testing.T) {
	for _, tt := range ColSplitTests() {
		if tt.debug {
			debug.DebugLog = true
		}
		t.Run(tt.name, func(t *testing.T) {
			got := ColSplit(tt.args.s, tt.args.sep, tt.args.enclosure, tt.args.n)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				debug.DebugColumns(got)
				t.Errorf("ColSplit(): %s Mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestColSplitFlat(t *testing.T) {
	for _, tt := range ColSplitTests() {
		if tt.debug {
			debug.DebugLog = true
		}
		t.Run(tt.name, func(t *testing.T) {
			v, pos := ColSplitFlat(tt.args.s, tt.args.sep, tt.args.enclosure, tt.args.n)
			got := FlatToMulti(v, pos)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				debug.DebugColumns(got)
				t.Errorf("ColSplit(): %s Mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

var resultColSplit [][]byte

func Benchmark_ColSplit(b *testing.B) {
	var out [][]byte
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out = ColSplit(in, characters.Character_QUOTATION_MARK, characters.Character_COMMA, -1)
	}

	resultColSplit = out
}

var resultColSplitFlat []byte
var resultColSplitPos []uint32

func Benchmark_ColSplitFlat(b *testing.B) {
	var v []byte
	var pos []uint32
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, pos = ColSplitFlat(in, characters.Character_QUOTATION_MARK, characters.Character_COMMA, -1)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

var resultGenSplitFlat []byte
var resultCGensplitPos []uint32

func Benchmark_GenSplit(b *testing.B) {
	var v []byte
	var pos []uint32
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, pos = GenSplit(in, characters.Character_COMMA, -1, NoClean, true)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

func Benchmark_GenSplitClean(b *testing.B) {
	var v []byte
	var pos []uint32
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	c := CharacterConfig{
		Enclosure:       characters.Character_QUOTATION_MARK,
		DoubleEnclosure: characters.Character_QUOTATION_MARK_DOUBLE,
		EnclosureClean:  true,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, pos = GenSplit(in, characters.Character_COMMA, -1, c.Clean, true)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

var resultOut []byte

func Benchmark_Cleaner(b *testing.B) {
	var out []byte
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	c := CharacterConfig{
		Enclosure:       characters.Character_QUOTATION_MARK,
		DoubleEnclosure: characters.Character_QUOTATION_MARK_DOUBLE,
		EnclosureClean:  true,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out = c.Clean(in, false, false)
	}

	resultOut = out
}

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
		cc := CharacterConfig{}
		cc.SetDelimiter(td.delimiter)
		cc.SetEnclosure(td.enclosure)
		cc.EOL = characters.Character_NIL

		// normally the code is not re-run so we need to manually set the headerLength to zero again
		cc.DelimiterCount = 0
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
		scanner.Split(cc.SplitCsvInput)
		// i := 0
		for scanner.Scan() {
			chunk := scanner.Bytes()
			// header has to be set once
			if cc.DelimiterCount == 0 {
				// we have to count the delimiters in the header to know for row scanning
				// how many fields we need
				cc.DelimiterCount = bytes.Count(chunk, cc.Delimiter)
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
	cc := CharacterConfig{}
	cc.SetDelimiter(characters.Character_NIL)
	cc.SetEnclosure(characters.Character_NIL)
	cc.EOL = characters.Character_NIL

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
			gotAdvance, gotToken, err := cc.SplitCsvInput(tt.args.in, tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitCsvInput(): %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("splitCsvInput(): %s gotAdvance = %d, want %d\ngotToken = \n%s\nwant = \n%s",
					tt.name, gotAdvance, tt.wantAdvance, gotToken, tt.wantToken)
			}
			if diff := cmp.Diff(tt.wantToken, gotToken); diff != "" {
				// if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("splitCsvInput(): %s \ngotToken = %s\nwant = %s\ndiff: %s", tt.name, gotToken, tt.wantToken, diff)
			}
		})
	}
}

func TestCharacterConfig_Clean(t *testing.T) {
	type args struct {
		data   []byte
		prefix bool
		suffix bool
	}
	type fields struct {
		Enclosure characters.Character
		Delimiter characters.Character
		EOL       characters.Character
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "inches",
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
			},
			args: args{
				// "Quote""" copied from debug session
				data: []uint8{34, 81, 117, 111, 116, 101, 34, 34, 34},
				// data:   []byte(`"Quote"""`),
				prefix: true,
				suffix: true,
			},
			want: []byte(`Quote"`),
		},
	}
	for _, tt := range tests {
		cc := CharacterConfig{}
		cc.SetDelimiter(tt.fields.Delimiter)
		cc.SetEnclosure(tt.fields.Enclosure)
		cc.EOL = tt.fields.EOL
		t.Run(tt.name, func(t *testing.T) {
			got := cc.Clean(tt.args.data, tt.args.prefix, tt.args.suffix)
			if diff := cmp.Diff(string(tt.want), string(got)); diff != "" {
				t.Errorf("CharacterConfig.Clean(): %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
