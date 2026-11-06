package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path"
	"reflect"
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
		t.Errorf("MultiToFlat() mismatch (-want +got)f:\n%s", diff)
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
	got, pos := GenSplit(foo, []byte("|"), -1, NoClean, nil)

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
			name: "Trailing comma",
			args: ColSplitTestArgs{
				[]byte(`"column1,etvni column 1","column2",`),
				characters.Character_COMMA,
				characters.Character_QUOTATION_MARK,
				-1,
			},
			want: [][]byte{
				[]byte(`"column1,etvni column 1"`),
				[]byte(`"column2"`),
				[]byte(``),
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
			}, debug: false,
		},
		{
			args: ColSplitTestArgs{
				s:         []byte(`,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,ipasltzm,1,6,H0.6.00018_vfdpux,Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv,0|0|0|0|0|0,"<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>",H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,|||||,,,,,,,,|||||,|||||,|||||,|||||,|||||,|||||,|||||,|||||`),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
				n:         -1,
			},
			want: [][]byte{
				[]byte(``),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`ipasltzm`),
				[]byte(`1`),
				[]byte(`6`),
				[]byte(`H0.6.00018_vfdpux`),
				[]byte(`Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv`),
				[]byte(`0|0|0|0|0|0`),
				[]byte(`"<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>"`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5`),
				[]byte(`|||||`),
				[]byte(``),
				[]byte(``),
				[]byte(``),
				[]byte(``),
				[]byte(``),
				[]byte(``),
				[]byte(``),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
				[]byte(`|||||`),
			},
			debug: true,
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
			v, pos := ColSplitFlat(tt.args.s, tt.args.sep, tt.args.enclosure, tt.args.n, func(data []byte, prefix, suffix bool) []byte { return data })
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
		v, pos = ColSplitFlat(in, characters.Character_QUOTATION_MARK, characters.Character_COMMA, -1, func(data []byte, prefix, suffix bool) []byte { return data })
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

var resultGenSplitFlat []byte
var resultCGensplitPos []uint32

// Benchmark_GenSplit-8   	30316046	       468.3 ns/op	     944 B/op	       3 allocs/op
func Benchmark_GenSplit(b *testing.B) {
	var v []byte
	var pos []uint32
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, pos = GenSplit(in, characters.Character_COMMA, -1, NoClean, nil)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

func Benchmark_GenSplitCore(b *testing.B) {
	var v []byte
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	buf := make([]byte, 0, 4*1024*1024)
	pos := make([]uint32, 0, 1024)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf = buf[:0]
		pos = pos[:0]
		v, pos = GenSplitCore(in, characters.Character_COMMA, -1, NoClean, nil, buf, pos)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

func Benchmark_GenSplitClean(b *testing.B) {
	var v []byte
	var pos []uint32
	in := []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)
	c := CharacterConfig{
		Config: characters.Config{
			Enclosure:       characters.Character_QUOTATION_MARK,
			DoubleEnclosure: characters.Character_QUOTATION_MARK_DOUBLE,
			EnclosureClean:  true,
		},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v, pos = GenSplit(in, characters.Character_COMMA, -1, c.Clean, nil)
	}

	resultColSplitFlat = v
	resultColSplitPos = pos
}

var resultOut []byte

var CleanerIn []byte = []byte(`"2189","97998848","007-04-8133","FCFO","DC_FZXMI","2022-03-11T21:46:09N","9999-12-31T00:00:00N","9.89","9.89","","[{""exftqvytk_id"":""I007CCK"",""xfa_mswb"":""2021-08-15"",""gtctd"":""01"",""zpjlv"":""R"",""eamrd"":""35"",""xzdjwnv"":""1"",""bwjeamwtgug_mswb"":""2022-08-14""}]"`)

func Benchmark_Cleaner(b *testing.B) {
	var out []byte
	c := CharacterConfig{
		Config: characters.Config{
			Enclosure:       characters.Character_QUOTATION_MARK,
			DoubleEnclosure: characters.Character_QUOTATION_MARK_DOUBLE,
			EnclosureClean:  true,
		},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out = c.Clean(CleanerIn, false, false)
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
	php       bool
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
			debug:     false,
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
			debug:     false,
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
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "php mode enabled automatically on a multi-line field",
			in:   []byte("a,b\n\"x\\\"\"y,p,q\nz\",1\n2,3"),
			multiOut: [][]byte{
				[]byte("a,b"),
				[]byte("\"x\\\"\"y,p,q\nz\",1"),
				[]byte("2,3"),
			},
			advance:   4,
			out:       []byte("a,b"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "php escaped enclosure on the first line of a multi-line field",
			in:   []byte("\"a\",\"b\",\"c\"\n1,\"foo \\\" bar\nbaz\",3\n4,5,6"),
			multiOut: [][]byte{
				[]byte("\"a\",\"b\",\"c\""),
				[]byte("1,\"foo \"\" bar\nbaz\",3"),
				[]byte("4,5,6"),
			},
			advance:   12,
			out:       []byte("\"a\",\"b\",\"c\""),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			php:       true,
		},
		{
			name: "php escaped enclosure leaves one extra delimiter outside enclosures",
			in:   []byte("id,name,note\n1,\"a\",x\n2,\"text \\\"quote, more\\\"\",y\n3,\"b\",z\n4,c,w"),
			multiOut: [][]byte{
				[]byte("id,name,note"),
				[]byte("1,\"a\",x"),
				[]byte("2,\"text \"\"quote, more\"\"\",y"),
				[]byte("3,\"b\",z"),
				[]byte("4,c,w"),
			},
			advance:   13,
			out:       []byte("id,name,note"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "multi-line value in the last of three columns",
			in:   []byte("id,title,description\n1,Foo,\"line one\nline two\"\n2,Bar,baz\n3,Qux,quux"),
			multiOut: [][]byte{
				[]byte("id,title,description"),
				[]byte("1,Foo,\"line one\nline two\""),
				[]byte("2,Bar,baz"),
				[]byte("3,Qux,quux"),
			},
			advance:   21,
			out:       []byte("id,title,description"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "backslash before a doubled quote, then a multi-line last value",
			in:   []byte("a,b,c\n4,\"x\\\"\"y\",\"line one\nline two\"\n7,8,9"),
			multiOut: [][]byte{
				[]byte("a,b,c"),
				[]byte("4,\"x\\\"\"y\",\"line one\nline two\""),
				[]byte("7,8,9"),
			},
			advance:   6,
			out:       []byte("a,b,c"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "unclosed quote in the last of two columns",
			in:   []byte("id,title\n1,\"Foo\"\n2,\"Bar\n3,Row3\n4,\"Baz\"\n5,Last"),
			multiOut: [][]byte{
				[]byte("id,title"),
				[]byte("1,\"Foo\""),
				[]byte("2,\"Bar"),
				[]byte("3,Row3"),
				[]byte("4,\"Baz\""),
				[]byte("5,Last"),
			},
			advance:   9,
			out:       []byte("id,title"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "unclosed quote before a quoted last field",
			in:   []byte("id,title\n1,\"Foo\"\n2,\"Bar\n3,Baz\n4,\"Fleep\""),
			multiOut: [][]byte{
				[]byte("id,title"),
				[]byte("1,\"Foo\""),
				[]byte("2,\"Bar"),
				[]byte("3,Baz"),
				[]byte("4,\"Fleep\""),
			},
			advance:   9,
			out:       []byte("id,title"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "unclosed quote before a multi-line field",
			in:   []byte("id,title\n1,\"Foo\"\n2,\"Bar\n3,\"multi\nline\"\n4,Last"),
			multiOut: [][]byte{
				[]byte("id,title"),
				[]byte("1,\"Foo\""),
				[]byte("2,\"Bar"),
				[]byte("3,\"multi\nline\""),
				[]byte("4,Last"),
			},
			advance:   9,
			out:       []byte("id,title"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "unclosed quote without a quote after it",
			in:   []byte("id,title\n1,\"Foo\"\n2,\"Bar\n3,Row3\n4,Last"),
			multiOut: [][]byte{
				[]byte("id,title"),
				[]byte("1,\"Foo\""),
				[]byte("2,\"Bar"),
				[]byte("3,Row3"),
				[]byte("4,Last"),
			},
			advance:   9,
			out:       []byte("id,title"),
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
		},
		{
			name: "unclosed quote in the middle of three columns",
			in:   []byte("id,title,price\n1,\"Foo\",10\n2,\"Bar,20\n3,Row3,30\n4,\"Baz\",40\n5,Last,50"),
			multiOut: [][]byte{
				[]byte("id,title,price"),
				[]byte("1,\"Foo\",10"),
				[]byte("2,\"Bar,20"),
				[]byte("3,Row3,30"),
				[]byte("4,\"Baz\",40"),
				[]byte("5,Last,50"),
			},
			advance:   15,
			out:       []byte("id,title,price"),
			delimiter: characters.Character_COMMA,
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
		{
			name: "0a165b2b97efff3fa041ef3899ceaf19",
			in: []byte(`vsusrpnca,WMU,wrjhr,ibqypuvnyjx,renjy,Jtbzwwfr,qwljy_yjuy,qwljy_02_yjuy,qwljy_03_yjuy
43935555,https://example.com/Mvjihjh-Tzfpdiz-pibocanb-oech-oducccbpryy-ovwkh/jr/U01QEGY50A/hdx=pt_1_22?__gu_uw_UW=%C3%85P%C3%85%C5%BD%C3%95%C3%91&kmpyxy=1&umliumzl=mvjihjh&zwy=2879057545&pt=8-22,"H Axgg Kuz Kviuz Tzfpdiz à yqhé, 1250D tzfpdiz à yqhé Uzqid","Mvjihjh H Axgg Kuz Kviuz Tzfpdiz à yqhé
Yn jrijta ot Kviuz - yn ueug hurlh bd yn ueug qjdkfmvoqh vzf ywuèkdm Mvjihjh H Axgg Kuz - dht jfcgèaxmdca ldmvpzg. Cxh bd lylgmé, ky s'lfmxjt xmznbdolgsmk à xykp ebs yditix, wxvybru eyz brkckm ot mpcxsxcé sg téanm bd fc flcmwqm fxkrrot bd lgroaia sg hyefsg.
Dxmq uvf jkzbbg-dootar tzsctgsa, Kviuz odxi pseixp ot iotyhms cidjm ebs alukryblt i'fc bsuiauqe hlmrcaujj bd qizi i'fc xuygè trxzw.
Mvjihjh H axgg kuz cfbfèij ot pibocanb hwuysatkx.",Mvjihjh,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/bvjmzn/R/81zRoPbAzWN._BL_EP1500_.knn,https://example.com/bvjmzn/R/71MWfEDdXIZ._BL_EP1500_.knn,https://example.com/bvjmzn/R/41Hz6hMG19F._BL_.knn
51053106,https://example.com/Mvjihjh-wzzmzyat-hni%C3%A9-EZ700-kviuz-ctkhw/jr/U07B2VI9KV/hdx=pt_1_37?__gu_uw_UW=%C3%85P%C3%85%C5%BD%C3%95%C3%91&kmpyxy=1&umliumzl=mvjihjh&zwy=2879057545&pt=8-37,"H Axgg Kuz Kviuz&Mlsu Ctkhw Tzfpdiz à yqhé, WVS (Aoixpaqqc), Zzgb","Mvjihjh H Axgg Kuz Kviuz & Mlsu Tzfpdiz à yqhé
Llégmild bd vxipcabkbij, fqnv pseixp ot jeéoowci vzf olhyuvsu kptizgmomd dxmq iy yqhé bd iy ktbk jfdrh. F'bilmhcjxrzr bsuiauqe bd yn iivn vogjfuavcm ldmvpzg, ag fc tmmt iktia.
Ru shlconhn Vogjfuavcm vdzésxé dht yn qqppcp i'eyz amuukh ot ktbk ncupp bd cgédlkdm bd ot ru dvédxvsthsr ot kaiqsw ebs qroxdboj ebs ueug kérqwcsuchk iy yqhé ldmvpzg à ru xbarlf.
Mvjihjh H axgg kuz cfbfèij ot pibocanb hwuysatkx",Mvjihjh,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/bvjmzn/R/81swaBaiaDW._BL_EP1500_.knn,https://example.com/bvjmzn/R/81QiyafRbeD._BL_EP1500_.knn,https://example.com/bvjmzn/R/41Hz6hMG19F._BL_.knn
AG560.U,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/deqqgwkbr/npyzvrvose-fflwu-ag-560u,Npyzvrvose Fflwu - Deqqgwkbr,"Urénbevm vzf qroxdboj bolzfprbk sg yqhé bd sg ktbk rqâom à fc uunscd ztygxd
Ru ggxta 19 vfqf zdkujp vzf pégsfvjwb csmfwiko
Tsmêk oducccbpryy fboèu 9 hfclkrs
Yvqrdo à kqdngukc rqâom sg cyohgtw ot ebswjzld fdjl fflwu
Cfbfèij ot wfpmeznet oducccbpryy ot ru gdzyiq ot ktbk

Ru tzfpdiz tgenlld, nsb aaobdtscx bolzfprbk,vcmj fc péwmzojs ag ovwkh i‘evukqjydq...…dxmq ueug ot qroxdboj vcmj ebs leafmaipj ebs ueug ipsrdjrkf...

Rqâom à xe wmreisiqwec whywmyéd Ot’Qyittx, SzyweRnwlb gcsabw, ru nwtosczb Npyzvrvose Fflwu odxi nrdgq ru zhqhsbpwxjé ot yémelfzs vzf olhyuvsu nyvméxc hszlvqlrfz bd kérqwcsuchk.

Géjookoiy ru nwtosczb cxflxkz ZwkhgzXeeje : fc txvu i'Bsuiauqe, 120ma ot ktbk isrjn bd eyz amuukh ot ktbk kirtfkvnm vcmj eyz képpsqaqamb lduwdxkps.",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Deqqgwkbr ,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-wbff.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-qwljy3.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-qwljy4.knn
DZ 695.G,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/bsuiauqe-jzaibyfktfep/eqjbgs-qptrx-dz-695g-8885470225,Eqjbgs Qptrx - Tzfpdiz Chbooagg,"Upcluinrcn sszhb yqhé yt rwveqev uséaélé bd urénbevm fc chbooagg nfbzh bd bhbrsyux, dxmq eyz uqiye jiztpxyk
Urénbevm sszhb chbooagg à ru dpywégjhdty eiésky rqâom à ru wmreisiqwec Fbhfkphahuv
Urénbevm fc syqtx fulyf ot qroxdboj à avpg ot ktbk rqâom sg cfbfèij Vogjfuavcm jékqedww
Xfwvoqc rtheebelvm yn wéutkqq-wwpfxcs vcmj eblnj oeoqkdxmz ot vyydf vcmj sszhb idy",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685.G_HQXNH_NMDSOR.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685_CXGEGX.knn
FQU 35.31,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/bsuiauqe-jzaibyfktfep/fqu-3531-1120297409,Tzfpdiz Chbooagg,"Upcluinrcn sszhb yqhé yt rwveqev uséaélé bd urénbevm fc chbooagg nfbzh bd bhbrsyux, dxmq eyz uqiye jiztpxyk
Urénbevm fc syqtx fulyf ot qroxdboj à avpg ot ktbk rqâom sg cfbfèij Vogjfuavcm jékqedww
Xfwvoqc rtheebelvm yn wéutkqq-wwpfxcs vcmj eblnj oeoqkdxmz ot vyydf vcmj sszhb idy.",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU3531_SV.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU3531_HQXNH.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU_3531_AIB.knn
CV6556/51,https://example.com/m-k/CV6556_51/skfyphkt-tzfpdiz-h-ucxu-h-jqvyvysr,ZWXIJC® Skfyphkt Tzfpdiz à yqhé à jqvyvysr,"Ru wmreisiqwec Rmrzxml i'sbôviy késhkkpu f'fdp xfb ru oxqzoha ot yqhé 
rqâom à sll 45 jvivi fmgp i'ayhnfoa fc snûg fxkrrot, jmfoce waf ru 
wmreisiqwec Uqiye ueug hirjesqm eyz svgfox ot wmècd i'eyz dbvjsdw bd 
i'eyz vozknyqgfé kljbbtdfo. Odxi xmcmza céypxfgkg jeéoowci eyz yqhé guou
 bd fxkvt ot gnes mnujneo pjhpatxno xfb fc ztygxd",Uxrygmf,Yqhé | Wzzmzyat à yqhé ZWXIJC® | ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-BOP-uw_UW?cnw=840&xzz=720&$filnfmzn$,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-FF1-ryxhko-001?$filnfmzn$&cnw=840&xzz=720,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-GBU-ryxhko-001?$filnfmzn$&cnw=840&xzz=720
CV7866/61,https://example.com/m-k/CV7866_61/lbkzyxsfj-tzfpdiz-h-ucxu-h-jqvyvysr,ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,"Géjookoiy fc snûg ueug nfbzh bd fxkrrot 
 Ru nwtosczb tzfpdiz à yqhé ZWXIJC® Lbkzyxsfj wweurmnb ebs leooocn iy 
yqhé rqâom à ru wmreisiqwec ZWXIJC® Rmrzxml i'sbôviy kfkkswpnyr à f'fdp 
i'kklwkj ag srgxdbq dxmq ru tjkpwhqé vzf 50 nyuqoj ot yqhé à ru mang, 
vcmj fc snûg ueug nfbzh bd fxkrrot. Uvf iéxrkzatx i'evqinklzé pseixp 
ueug ot fulyf.",Uxrygmf,Yqhé | Wzzmzyat à yqhé ZWXIJC® | ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-BOP-uw_UW?cnw=840&xzz=720&$filnfmzn$,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-FF1-ryxhko-001?$filnfmzn$&cnw=840&xzz=720,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-GBU-ryxhko-001?$filnfmzn$&cnw=840&xzz=720
UW460000,https://example.com/Xkwnwdz-%26-Ttkkejvas/Ttkkejvas-phypwmfyza/UHOKD-QSYLB-1-3Eb/k/8518145754?krc=ttkkejvas,Udiwcqio Hlmrcaujj Uhokd Qsylb,"Llégmild bd fpubckpx vcmj néraszi kaiqsw fqm isvrpvqk
Uhokd Qsylb, vzf sxgpio vcmj khhqp ru dmbbdlv xylc imcef !Urénbevm
 lnues’à 1,3 eb ot sxgpio xbarlf xmznbdolgsmk hsfnatlfbnexfr vcmj khhqp 
ru dmbbdlv ! Eyz udiwcqio yvqrdo i’orfumifca jgbtaue xykp sll étéuhtot 
fqighoejo yqqp chgmzknmhai ehiz-zpvtlmikc : geudshgah, krqxwq, flsb bd 
xsemep. Yvqrdo à kqdngukc rqâom à yqqp tozdhvnnpx vcmj eyz rqzapef 
jiztpxyk ot fqm sxgpio, pahjcffp, gxnegllw, wqfk. Xe wcdcwvdd aczy 
s’uhdèfbt xmznbdolgsmk bhuv kaiqsw ebs snazakvy bkbmevjd !",FYK,Xkwnwdz & Ttkkejvas | Ttkkejvas phypwmfyza,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w1VVO0QYmnwBPcCU9uuVAtbRxrWLhscv9aXRXehMReLuNrYTCdMNN2GlM1OIM4JvxxD3w3ZHM4IfeyP2MuDnUtRsBmMhJ3CpE5QfN4PPYjGQX3MkClT2ObZJGhMXYsVZGjRQWfTjV2BvU1ULL0JNpbMfJb,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3qaBbr3VtE8dF1wZ2CfgnXeJ3znkBXcPPLtaPJsQ2csWY8hIet1QnQ2LRSoFNS5MR5doOa8PBwfJULqNiBmYFmpRQO0Y2AqKMOfBpEqTdUxHXD3MOY1YdC3BJZ5KIJvPPYhXWemXjBsAsIuQForQxWyOWOmQU,
OL162100,https://example.com/Xkwnwdz-%26-Ttkkejvas/Ttkkejvas-phypwmfyza/FPOQGK-VVG-/k/4324578283?krc=ttkkejvas,Udiwcqio Hlmrcaujj Fpoqgk Vvg,"Nkpolsw sszhb huwoc ag fc tmmt iktia
Dxmq  Fpoqgk Vvg, jrkcwuz sszhb huwoc ag fc lmhk ot rddb !Dxmq
 Fpoqgk Vvg, urénbevm lnues’à 1,2 eb ot sxgpio vcmj khhqp ru dmbbdlv ! 
Yroéj i'fc cfbfèij ot zsjmdnkj fgyxeshg bd ggsqidé, fboèu dvzxni 
qeekhyxvnqe, xykp ebs kéktkeo ot rqzapef yqqp éjbvvkéz vcmj eyz huwoc 
uphfllad xmvzct.",FYK,Xkwnwdz & Ttkkejvas | Ttkkejvas phypwmfyza,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w1SSXfS3fuvJQhBN9uuVAtbRxrWLhscv9mGRKyjUB1GgG0SpUxFALpSpF2ZuvmGxckC3w5OzGeYOYlLsD1JjT2BPW1MuGvQYZdFDC5MAAcOTA3QUDnEfT2Y2Q4XWZsOPX0PrEgUKZmTtPcARJ3TvLsVBR4,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3kaXtUgUXR8dF1wZ2CfgnXeJ3wyoTOsFSGvrJn1L2g5UO8rDQK4ZPwkOxi0JPG4Rq5doOa8N2I1MWyuCWK1PAMhYIZvBgx2RRhcQKXuFeN0UtWeCODuBETiCETwWFNbJaq2YKu0VylhDFFcMmB2IObbWFNkDs,
K4605100,https://example.com/ulxeoteu-dovbxtophum/vow/8678015824,Zpfrjiffqas bd qlrtekk 2 ag 1 Pqovft Aqiuc' Jsy 5F,"Yn xxdqwoq zpfrjiffqas Fyk 2 ag 1, à ru mang zpfrjiffqas bd qlrtekk
Lxiuht yt wlpvmuy, ky dht wfpf vcmj odxi !
Yn QjzjmjCfqdo'® Jsy dht yn xxdqwoq zpfrjiffqas 2 ag 1 FYK, à ru mang 
zpfrjiffqas bd qlrtekk. Dénfrxbkolbb kiiaqlmosp sszhb bdfo ot rqzapef 
fxt uvf iéxrkzatx covhiwupdbi  vcmj oygtrllv fqm efmxxs, qklmgprr, 
déqyuco, jhorbuh bd qsdjwj. Rqâom à uvf fxyêzvrzjf losbeepéhfc, oscebdwf
 sszhb zpfrjiffqas o'deqz fganlo éqé fhwjx uunscd.",FYK,"Nbdrzzcrgjrd, iuêjoy bd enywbyoask | Zpfrjiffqas",https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3qcITM5M3fuvJQhBN9uuVAtbRxrWLhscv9zPVozwYCdTeY1VVO0KCXrSVP2OGXdEyecI3fgZefpZ2IlDABjT2VuLTW2UiVkFPYcLUvuNEYwTsn3DkB2PYT3ZWB0N2PzQUE5ESA1UOTzFfRqJGOvZTz5LWc2,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w3BSk2BQfnpPPvHZ9uuVAtbRxrWLhscv9uRKQhwXzsKrE1VVO0PMFpXPd3KLNkBdjwL3w0HRX2HRX0RtTwLlZuJWbfWZM1AFKyDMNdUSSwAiB0JvWwUeGrNyI1ShXzM2Q1UfOdNYYxTTagUdb0N2NqX2Gr,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w5YjQjX3fuvJQhBN9uuVAtbRxrWLhscv9uURUsoDXuHeH1VVO0IXLfXKC5TnI2JvxxD3xfEGatBhp2StG5COX4RqUrOGKeUGKmFXL5HwawIjmrQNGrTFBjVMQ1OXAuRPPtEJW3XTMjXnIvWWJtWlOxCJMd
K4220700,https://example.com/ulxeoteu-dovbxtophum/Wfqkzcka/Nbdrzzcrgjrd--fa%C3%AAebs-bd-enywbyoask/Zpfrjiffqas/Ispqbmhaw-6F/vow/0728333977?gbfqdDjduqntjTrle=CrqwoesnXgmuiz,Ispqbmhaw 6F,"Cétèwc khhqp ru nzakbs vzf nctmsxdr.
4 icvgnkcxws ot rqzapef vwscatwcb à dvzxni dmbbdlv i'nctmsxdr
 vcmj zséjlkdkj ot cwzyèhi epszifdz rezsg eujmwxép ngomswoyxoesurl bd 
tgkv nzakbs. Ru geo ot rqzapef pd wfpf ag lxiuht ncupp.
- yn yftmukpc bd yn kjchh ot olhyuvsu odxi lddyvwr vcmj hébzvruh ru vkkecanba rqzapef.
- yn xsemep lxiuht dht jmlçh vcmj hébzvruh eyz rqzapef yhtemèzz bd opawégkr ebs nctmsxdr vzf zédutbwkk i'fjvmlfdbi.
- bvwrjwhnt bd wcsvdbxbe erjhiuy.
- wcsvdbxbe vûqk bd ovgzwtiioéa.",FYK,"Nbdrzzcrgjrd, iuêjoy bd enywbyoask | Zpfrjiffqas",https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w0OeS2IvuwlBVjPI9uuVAtbRxrWLhscv9dNCDcvHPcAiN0HRX1HCRrXYI4MHaoWuyxN3gxYMKcQdItKUD2KWkxGqxcILpzRjXwDBIlK2WhKHB5ZPC5GHUuOLHkZSG1IfBxCwU1RaF4MZCnOzU5EHO2ZVWq,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w5YjQjX3fuvJQhBN9uuVAtbRxrWLhscv9uURUsoDXuHeH1VVO0IXLfXKC5TnI2JvxxD3xfEGatBhp2StG5COX4RqUrOGKeUGKmFXL5HwawIjmrQNGrTFBjVMQ1OXAuRPPtEJW3XTMjXnIvWWJtWlOxCJMd,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w3BSk2BQfnpPPvHZ9uuVAtbRxrWLhscv9uRKQhwXzsKrE1VVO0PMFpXPd3KLNkBdjwL3w0HRX2HRX0RtTwLlZuJWbfWZM1AFKyDMNdUSSwAiB0JvWwUeGrNyI1ShXzM2Q1UfOdNYYxTTagUdb0N2NqX2Gr
PE468910,https://example.com/HURLH-I%C3%89ZRKUKC/Xiylsi%C3%A8ai-krqxwq/MYUZERCLI-BYWTSDQX-ZZGB-ACZY-12-DOOTAR/k/3028234938?gbfqdDjduqntjTrle=hurlh-opkfeyfn,MYUZERCLI BYWTSDQX ZZGB ACZY 12 DOOTAR,"Pfokkwèzl à bywtsdqx dxmq ibsuhktt 12 dootar sg jrijta hlitxc bd qjnestcbje ot Ugmwn
Pfokkwèzl krqxwq aifnmvjv : uvf cfbfèij qhwfne ot mbla ag bywtsdqx ncyaaotjb yn snûg bilmhcjxrzr nsb zhtdanpbéc ot yqhé.Eyz pfokkwèzl à kqdngukc xykp ebs dgwbw vcmj hébzvruh fc yqhé uphfllad gdèv otlbfcehgl bd bhxjw.",UGMWN,hurlh wépictgq | pfokkwèzl krqxwq ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3lsLny3IXN8dF1wZ2CfgnXeJ3xbfJCxAKKrnZO4L2ubXd8bBSA4UsntFdV1MZOqHM5doOa8JpKzEXW4EsX0CPxmH2FvLXFcI2EwCcAkDqleLaItUaGqY2WjMXGsDWlmBDW5FSHoJOF3PuCsWyOsUzt4IQJ1PC,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3csMPadNIT8dF1wZ2CfgnXeJ3swvAJcWGJqsOMjX2g5Rq8jVLLbGoWfGeRhUaj2Rq5doOa8OHA0PPW1ViC0KGA0OcObEaDeCxgeRqjzMWXyHoCqKJK4GIZlW2LeXIY0AbNmBgS0CwU3PCGpOoMnXNZvUEWjUj,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3wxWFO5SJQ8dF1wZ2CfgnXeJ3ygpLUkZLLtfQG3L2syOR8iOGGlTlImWnP2QlpzLJ5doOa8QSl2GFIdD2MiHDK5YKOcLSTcQSH4PEy0PDLfBOC0TQk3LMY4BvU0TvS1QfsiTrc3JFfaNQD2SMPyRDHqVmcaES
PE305D10,https://example.com/HURLH-I%C3%89ZRKUKC/Xiylsi%C3%A8ai-krqxwq/FRS-HTSJZ-OCIJJVRQM-ZZGB-ACZY/k/1298531048?gbfqdDjduqntjTrle=hurlh-opkfeyfn,FRS HTSJZ OCIJJVRQM ZZGB ACZY,"Uwcbécsjn sszhb yqhé, tèf sszhb félobd bd qjfc sg xczd ot ru utwpnéw dxmq ru pfokkwèzl FRS HTSJZ aczy ocijjvrqm ot UGMWN.
Dxmq ru pfokkwèzl FRS HTSJZ OCIJJVRQM ot Ugmwn,  uwcbécsjn fc céofswwmb 
yqhé, rqâom à uvf nptlhôzy hféexg ot ru dpywégjhdty vcmj eyz jdxyhfmfcg 
iy yqhé epszifdz bd uvf fyd ocijjvrqm igk oehoisnsekh sszhb yqhé sg 
isrjn lnues'à 4 iiwgty fboèu ru geo iy usdvl. Xe nzhehcmq rxhx-qaczrl 
pseixp ot llvugf yn yqhé, rêss lriavdi yn usdvl ot yqhé. Uvf 
cvnsp-krqxwq kmodtefy bd tzsctgsa qyubc waf xe nzhehcmq tsmêk 
oducccbpryy odxi rvnsazletjuj ru fqk sg acupaumjo",UGMWN,hurlh wépictgq | pfokkwèzl krqxwq ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3vsKJP5SAP8dF1wZ2CfgnXeJ3vxxRDtKQTssJV0L2kaDX8rWTYpBjP0OsctQTKqPM5doOa8ApI2YXJfHxSdWwTyOEJ1NRZvZBTpQdSpBBT3MbdeCrYjZZXoUYHfFoSuQKHvCLL5NWRkWZH0Y2WpMFm1IlCrGV,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3ziPUN0DaI8dF1wZ2CfgnXeJ3svyTBpJHNdlVsjG2gdAH8rWTYpBjP0SOBwOIVvDz5doOa8XDKmTdwaZ2DiIuonR2MrTJr3BPW0ECAkVWL3YQWsCfCgInEtYYU1ZsKaU2U5OAKtTDw1KkZ5FJQqMhZ5QBX1HM,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3gaBFlmONo8dF1wZ2CfgnXeJ3ygpLUkZLLtfQG3L2rfEc8gWLKdWAS2PsS3LcW1GN5doOa8EEGpGKLgAFB3BCA1WSM2FQZvZZYpAzXuZThoEKPcRfR5ScihBfCzYYWzF2CzHoD2ANDdVUF5NlF3QFQlG2ZrOW
LR351412,https://example.com/Rqzapef-%C3%A0-jwpdr/Rqzapef-jvwivx%C3%A9a/Zagekr/Zagekr-Oyynpo-Odmgld-/k/4787570329?krc=rqzapef%2bglbthzjv,Zagekr Oyynpo Odmgld,"F'qtfzmolt à zagekr fpubckpx bd ofbértjwx
Psélt vcmj xykp dwgda ot vnbzqrk : pmshsvu, huwoc, drvcdnvp…Xrjqgkf ajapiqm vcmj uvuucsrqb f'kxkèz nsb 5 bspikmdti.Wfpmeznet yvqrdo : dqzoyysu, jkzbbg wsbskpdznvd, bspikmdti bd jslu à zagekr yqqp chgmzknmhai ehiz-zpvtlmikc.Onuopiabfz ehiz-zpvtlmikc : dqzoyysu, jkzbbg wsbskpdznvd, bspikmdti bd jslu à zagekr.",Wohrq,Rqzapef à jwpdr | Rqzapef ifonssép ,https://example.com/micohz/DJ-OYYNPO-KQNAFJG-ZAGEKR-LR351412-DSTTEXLH-405x345.psj?inzkmwz=lWGtrYUsgZKgi3R8Kqq4FMz8dF1wZ2AjlS5oqKr3GX9jQGMtSDq5CnEqFfP1QTBkSG5ouhk8MmZlYaakDYQ4XGOyWWVnBvmqArEjTHZ0ZXWvYwowCCX2YLq5EHH0TYdfXlF4LMHdCzTfNlb5VoSeTOM4M2FdMH,https://example.com/micohz/DJ-OYYNPO-KQNAFJG-ZAGEKR-LR351412-LEQHEC-XIPWODF-TOZDHVNNPX..psj?inzkmwz=eOSlqHBhbZsnIXhws3lgYDT3SvjlyEYoTY9ouhk8dF1wZ2TjX2zeLS9xHNXzNMvxRvV2ORItXJP5MR5ouhk8QwKaAoCaNuskQEM1PWfnOLYjGTZxJGzuKuSuKxG4LZR5ApS4WtYyFwXbGfTmXKW2SPSoO2TqNQA0WUR2OGXqEu,
LR260312,https://example.com/Rqzapef-%C3%A0-jwpdr/Rqzapef-jvwivx%C3%A9a/Zagekr/PHUCXMDMIW-OPJOGWNFJUBWJ-8-RPOVMGHWS/k/6687587445?krc=rqzapef%2bglbthzjv,PHUCXMDMIW OPJOGWNFJUBWJ 8 RPOVMGHWS,"Ru zagekr ag khhqp vatxyqxfjé !
Zagekr Phucxmdmiw : yn mmlunfkd ot ru zagekr !Géjookoiy
 ru zagekr Phucxmdmiw, Psélt vcmj xykp ebs dwgda ot vnbzqrk : 
gxhvtnmalbvvte, fuqutimvwp, eemdgnac bd rêss sg drvcdnvp ag khhqp 
vatxyqxfjé ! Rqâom à uvf cmhof nudgjba, f'qtfzmolt à zagekr jpya êbit 
gxypeffé ag khhqp pépklyné. Nékjug bgrrs ru dpywégjhdty kljtcinéa dxmq 
yn tozdhvnnpx bd m’dht fatzl ! Fc qtfzmolt à zagekr igk ncyaaotjb ot ru 
jfthscm à sszhb jwpdr !",Wohrq,Rqzapef à jwpdr | Rqzapef ifonssép ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3frFJP4MJxxwXUtDR9uuVAtbRxrWLhscv9xVHErtTJ2UfRqNTFyIIN4ITH4IyM2JvxxD3geVMUtENQxVfW1RWI1XRM0VlMdDODrXXW1QIJmVjM0AVZ2LvNmAJZyEKR1OJYaHoI2PaG5XCc1HYH1SDM0UVAf,https://example.com/micohz/?inzkmwz=lWGtrYUsgZKgi3R8YXPtNzT8dF1wZ2CfgnXeJ3hoTfKqjXN0GgG0JRKeCsM0OgRyQakdThjjT3vnAYBbKwNcQvq0FHwrJrz5UPj1N2U5ZfAlF2CgREWzQVF4MCDcRWR4CLLlGymeZFI3NQCgAFL5MJEfAeA1COX2,`),
			multiOut: [][]byte{
				[]byte(`vsusrpnca,WMU,wrjhr,ibqypuvnyjx,renjy,Jtbzwwfr,qwljy_yjuy,qwljy_02_yjuy,qwljy_03_yjuy`),
				[]byte(`43935555,https://example.com/Mvjihjh-Tzfpdiz-pibocanb-oech-oducccbpryy-ovwkh/jr/U01QEGY50A/hdx=pt_1_22?__gu_uw_UW=%C3%85P%C3%85%C5%BD%C3%95%C3%91&kmpyxy=1&umliumzl=mvjihjh&zwy=2879057545&pt=8-22,"H Axgg Kuz Kviuz Tzfpdiz à yqhé, 1250D tzfpdiz à yqhé Uzqid","Mvjihjh H Axgg Kuz Kviuz Tzfpdiz à yqhé
Yn jrijta ot Kviuz - yn ueug hurlh bd yn ueug qjdkfmvoqh vzf ywuèkdm Mvjihjh H Axgg Kuz - dht jfcgèaxmdca ldmvpzg. Cxh bd lylgmé, ky s'lfmxjt xmznbdolgsmk à xykp ebs yditix, wxvybru eyz brkckm ot mpcxsxcé sg téanm bd fc flcmwqm fxkrrot bd lgroaia sg hyefsg.
Dxmq uvf jkzbbg-dootar tzsctgsa, Kviuz odxi pseixp ot iotyhms cidjm ebs alukryblt i'fc bsuiauqe hlmrcaujj bd qizi i'fc xuygè trxzw.
Mvjihjh H axgg kuz cfbfèij ot pibocanb hwuysatkx.",Mvjihjh,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/bvjmzn/R/81zRoPbAzWN._BL_EP1500_.knn,https://example.com/bvjmzn/R/71MWfEDdXIZ._BL_EP1500_.knn,https://example.com/bvjmzn/R/41Hz6hMG19F._BL_.knn`),
				[]byte(`51053106,https://example.com/Mvjihjh-wzzmzyat-hni%C3%A9-EZ700-kviuz-ctkhw/jr/U07B2VI9KV/hdx=pt_1_37?__gu_uw_UW=%C3%85P%C3%85%C5%BD%C3%95%C3%91&kmpyxy=1&umliumzl=mvjihjh&zwy=2879057545&pt=8-37,"H Axgg Kuz Kviuz&Mlsu Ctkhw Tzfpdiz à yqhé, WVS (Aoixpaqqc), Zzgb","Mvjihjh H Axgg Kuz Kviuz & Mlsu Tzfpdiz à yqhé
Llégmild bd vxipcabkbij, fqnv pseixp ot jeéoowci vzf olhyuvsu kptizgmomd dxmq iy yqhé bd iy ktbk jfdrh. F'bilmhcjxrzr bsuiauqe bd yn iivn vogjfuavcm ldmvpzg, ag fc tmmt iktia.
Ru shlconhn Vogjfuavcm vdzésxé dht yn qqppcp i'eyz amuukh ot ktbk ncupp bd cgédlkdm bd ot ru dvédxvsthsr ot kaiqsw ebs qroxdboj ebs ueug kérqwcsuchk iy yqhé ldmvpzg à ru xbarlf.
Mvjihjh H axgg kuz cfbfèij ot pibocanb hwuysatkx",Mvjihjh,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/bvjmzn/R/81swaBaiaDW._BL_EP1500_.knn,https://example.com/bvjmzn/R/81QiyafRbeD._BL_EP1500_.knn,https://example.com/bvjmzn/R/41Hz6hMG19F._BL_.knn`),
				[]byte(`AG560.U,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/deqqgwkbr/npyzvrvose-fflwu-ag-560u,Npyzvrvose Fflwu - Deqqgwkbr,"Urénbevm vzf qroxdboj bolzfprbk sg yqhé bd sg ktbk rqâom à fc uunscd ztygxd
Ru ggxta 19 vfqf zdkujp vzf pégsfvjwb csmfwiko
Tsmêk oducccbpryy fboèu 9 hfclkrs
Yvqrdo à kqdngukc rqâom sg cyohgtw ot ebswjzld fdjl fflwu
Cfbfèij ot wfpmeznet oducccbpryy ot ru gdzyiq ot ktbk

Ru tzfpdiz tgenlld, nsb aaobdtscx bolzfprbk,vcmj fc péwmzojs ag ovwkh i‘evukqjydq...…dxmq ueug ot qroxdboj vcmj ebs leafmaipj ebs ueug ipsrdjrkf...

Rqâom à xe wmreisiqwec whywmyéd Ot’Qyittx, SzyweRnwlb gcsabw, ru nwtosczb Npyzvrvose Fflwu odxi nrdgq ru zhqhsbpwxjé ot yémelfzs vzf olhyuvsu nyvméxc hszlvqlrfz bd kérqwcsuchk.

Géjookoiy ru nwtosczb cxflxkz ZwkhgzXeeje : fc txvu i'Bsuiauqe, 120ma ot ktbk isrjn bd eyz amuukh ot ktbk kirtfkvnm vcmj eyz képpsqaqamb lduwdxkps.",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Deqqgwkbr ,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-wbff.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-qwljy3.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Deqqgwkbr-Gcsabw/AG560.U/AG560.U-qwljy4.knn`),
				[]byte(`DZ 695.G,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/bsuiauqe-jzaibyfktfep/eqjbgs-qptrx-dz-695g-8885470225,Eqjbgs Qptrx - Tzfpdiz Chbooagg,"Upcluinrcn sszhb yqhé yt rwveqev uséaélé bd urénbevm fc chbooagg nfbzh bd bhbrsyux, dxmq eyz uqiye jiztpxyk
Urénbevm sszhb chbooagg à ru dpywégjhdty eiésky rqâom à ru wmreisiqwec Fbhfkphahuv
Urénbevm fc syqtx fulyf ot qroxdboj à avpg ot ktbk rqâom sg cfbfèij Vogjfuavcm jékqedww
Xfwvoqc rtheebelvm yn wéutkqq-wwpfxcs vcmj eblnj oeoqkdxmz ot vyydf vcmj sszhb idy",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685.G_HQXNH_NMDSOR.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Wiwd-Rv-Hkl/DZ%20685.G/DZ685_CXGEGX.knn`),
				[]byte(`FQU 35.31,https://example.com/uw-uw/wfqkzcka/ucxu/tzfpdiz-h-ucxu/bsuiauqe-jzaibyfktfep/fqu-3531-1120297409,Tzfpdiz Chbooagg,"Upcluinrcn sszhb yqhé yt rwveqev uséaélé bd urénbevm fc chbooagg nfbzh bd bhbrsyux, dxmq eyz uqiye jiztpxyk
Urénbevm fc syqtx fulyf ot qroxdboj à avpg ot ktbk rqâom sg cfbfèij Vogjfuavcm jékqedww
Xfwvoqc rtheebelvm yn wéutkqq-wwpfxcs vcmj eblnj oeoqkdxmz ot vyydf vcmj sszhb idy.",Ot'Qyittx,Yqhé | Tzfpdiz à yqhé | Tzfpdiz Chbooagg,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU3531_SV.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU3531_HQXNH.knn,https://example.com/Ryxhko/Vbauxoed/Bvjmzn/Nwtmsu-Kobwlg/Kyen-Bsuiauqe/FQU%2035.31/FQU_3531_AIB.knn`),
				[]byte(`CV6556/51,https://example.com/m-k/CV6556_51/skfyphkt-tzfpdiz-h-ucxu-h-jqvyvysr,ZWXIJC® Skfyphkt Tzfpdiz à yqhé à jqvyvysr,"Ru wmreisiqwec Rmrzxml i'sbôviy késhkkpu f'fdp xfb ru oxqzoha ot yqhé 
rqâom à sll 45 jvivi fmgp i'ayhnfoa fc snûg fxkrrot, jmfoce waf ru 
wmreisiqwec Uqiye ueug hirjesqm eyz svgfox ot wmècd i'eyz dbvjsdw bd 
i'eyz vozknyqgfé kljbbtdfo. Odxi xmcmza céypxfgkg jeéoowci eyz yqhé guou
 bd fxkvt ot gnes mnujneo pjhpatxno xfb fc ztygxd",Uxrygmf,Yqhé | Wzzmzyat à yqhé ZWXIJC® | ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-BOP-uw_UW?cnw=840&xzz=720&$filnfmzn$,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-FF1-ryxhko-001?$filnfmzn$&cnw=840&xzz=720,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV6556_51-GBU-ryxhko-001?$filnfmzn$&cnw=840&xzz=720`),
				[]byte(`CV7866/61,https://example.com/m-k/CV7866_61/lbkzyxsfj-tzfpdiz-h-ucxu-h-jqvyvysr,ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,"Géjookoiy fc snûg ueug nfbzh bd fxkrrot 
 Ru nwtosczb tzfpdiz à yqhé ZWXIJC® Lbkzyxsfj wweurmnb ebs leooocn iy 
yqhé rqâom à ru wmreisiqwec ZWXIJC® Rmrzxml i'sbôviy kfkkswpnyr à f'fdp 
i'kklwkj ag srgxdbq dxmq ru tjkpwhqé vzf 50 nyuqoj ot yqhé à ru mang, 
vcmj fc snûg ueug nfbzh bd fxkrrot. Uvf iéxrkzatx i'evqinklzé pseixp 
ueug ot fulyf.",Uxrygmf,Yqhé | Wzzmzyat à yqhé ZWXIJC® | ZWXIJC® Lbkzyxsfj Tzfpdiz à yqhé à jqvyvysr,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-BOP-uw_UW?cnw=840&xzz=720&$filnfmzn$,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-FF1-ryxhko-001?$filnfmzn$&cnw=840&xzz=720,https://example.com/nw/qwljy/ZcobeeyAgtgfcag/CV7866_61-GBU-ryxhko-001?$filnfmzn$&cnw=840&xzz=720`),
				[]byte(`UW460000,https://example.com/Xkwnwdz-%26-Ttkkejvas/Ttkkejvas-phypwmfyza/UHOKD-QSYLB-1-3Eb/k/8518145754?krc=ttkkejvas,Udiwcqio Hlmrcaujj Uhokd Qsylb,"Llégmild bd fpubckpx vcmj néraszi kaiqsw fqm isvrpvqk
Uhokd Qsylb, vzf sxgpio vcmj khhqp ru dmbbdlv xylc imcef !Urénbevm
 lnues’à 1,3 eb ot sxgpio xbarlf xmznbdolgsmk hsfnatlfbnexfr vcmj khhqp 
ru dmbbdlv ! Eyz udiwcqio yvqrdo i’orfumifca jgbtaue xykp sll étéuhtot 
fqighoejo yqqp chgmzknmhai ehiz-zpvtlmikc : geudshgah, krqxwq, flsb bd 
xsemep. Yvqrdo à kqdngukc rqâom à yqqp tozdhvnnpx vcmj eyz rqzapef 
jiztpxyk ot fqm sxgpio, pahjcffp, gxnegllw, wqfk. Xe wcdcwvdd aczy 
s’uhdèfbt xmznbdolgsmk bhuv kaiqsw ebs snazakvy bkbmevjd !",FYK,Xkwnwdz & Ttkkejvas | Ttkkejvas phypwmfyza,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w1VVO0QYmnwBPcCU9uuVAtbRxrWLhscv9aXRXehMReLuNrYTCdMNN2GlM1OIM4JvxxD3w3ZHM4IfeyP2MuDnUtRsBmMhJ3CpE5QfN4PPYjGQX3MkClT2ObZJGhMXYsVZGjRQWfTjV2BvU1ULL0JNpbMfJb,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3qaBbr3VtE8dF1wZ2CfgnXeJ3znkBXcPPLtaPJsQ2csWY8hIet1QnQ2LRSoFNS5MR5doOa8PBwfJULqNiBmYFmpRQO0Y2AqKMOfBpEqTdUxHXD3MOY1YdC3BJZ5KIJvPPYhXWemXjBsAsIuQForQxWyOWOmQU,`),
				[]byte(`OL162100,https://example.com/Xkwnwdz-%26-Ttkkejvas/Ttkkejvas-phypwmfyza/FPOQGK-VVG-/k/4324578283?krc=ttkkejvas,Udiwcqio Hlmrcaujj Fpoqgk Vvg,"Nkpolsw sszhb huwoc ag fc tmmt iktia
Dxmq  Fpoqgk Vvg, jrkcwuz sszhb huwoc ag fc lmhk ot rddb !Dxmq
 Fpoqgk Vvg, urénbevm lnues’à 1,2 eb ot sxgpio vcmj khhqp ru dmbbdlv ! 
Yroéj i'fc cfbfèij ot zsjmdnkj fgyxeshg bd ggsqidé, fboèu dvzxni 
qeekhyxvnqe, xykp ebs kéktkeo ot rqzapef yqqp éjbvvkéz vcmj eyz huwoc 
uphfllad xmvzct.",FYK,Xkwnwdz & Ttkkejvas | Ttkkejvas phypwmfyza,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w1SSXfS3fuvJQhBN9uuVAtbRxrWLhscv9mGRKyjUB1GgG0SpUxFALpSpF2ZuvmGxckC3w5OzGeYOYlLsD1JjT2BPW1MuGvQYZdFDC5MAAcOTA3QUDnEfT2Y2Q4XWZsOPX0PrEgUKZmTtPcARJ3TvLsVBR4,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3kaXtUgUXR8dF1wZ2CfgnXeJ3wyoTOsFSGvrJn1L2g5UO8rDQK4ZPwkOxi0JPG4Rq5doOa8N2I1MWyuCWK1PAMhYIZvBgx2RRhcQKXuFeN0UtWeCODuBETiCETwWFNbJaq2YKu0VylhDFFcMmB2IObbWFNkDs,`),
				[]byte(`K4605100,https://example.com/ulxeoteu-dovbxtophum/vow/8678015824,Zpfrjiffqas bd qlrtekk 2 ag 1 Pqovft Aqiuc' Jsy 5F,"Yn xxdqwoq zpfrjiffqas Fyk 2 ag 1, à ru mang zpfrjiffqas bd qlrtekk
Lxiuht yt wlpvmuy, ky dht wfpf vcmj odxi !
Yn QjzjmjCfqdo'® Jsy dht yn xxdqwoq zpfrjiffqas 2 ag 1 FYK, à ru mang 
zpfrjiffqas bd qlrtekk. Dénfrxbkolbb kiiaqlmosp sszhb bdfo ot rqzapef 
fxt uvf iéxrkzatx covhiwupdbi  vcmj oygtrllv fqm efmxxs, qklmgprr, 
déqyuco, jhorbuh bd qsdjwj. Rqâom à uvf fxyêzvrzjf losbeepéhfc, oscebdwf
 sszhb zpfrjiffqas o'deqz fganlo éqé fhwjx uunscd.",FYK,"Nbdrzzcrgjrd, iuêjoy bd enywbyoask | Zpfrjiffqas",https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3qcITM5M3fuvJQhBN9uuVAtbRxrWLhscv9zPVozwYCdTeY1VVO0KCXrSVP2OGXdEyecI3fgZefpZ2IlDABjT2VuLTW2UiVkFPYcLUvuNEYwTsn3DkB2PYT3ZWB0N2PzQUE5ESA1UOTzFfRqJGOvZTz5LWc2,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w3BSk2BQfnpPPvHZ9uuVAtbRxrWLhscv9uRKQhwXzsKrE1VVO0PMFpXPd3KLNkBdjwL3w0HRX2HRX0RtTwLlZuJWbfWZM1AFKyDMNdUSSwAiB0JvWwUeGrNyI1ShXzM2Q1UfOdNYYxTTagUdb0N2NqX2Gr,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w5YjQjX3fuvJQhBN9uuVAtbRxrWLhscv9uURUsoDXuHeH1VVO0IXLfXKC5TnI2JvxxD3xfEGatBhp2StG5COX4RqUrOGKeUGKmFXL5HwawIjmrQNGrTFBjVMQ1OXAuRPPtEJW3XTMjXnIvWWJtWlOxCJMd`),
				[]byte(`K4220700,https://example.com/ulxeoteu-dovbxtophum/Wfqkzcka/Nbdrzzcrgjrd--fa%C3%AAebs-bd-enywbyoask/Zpfrjiffqas/Ispqbmhaw-6F/vow/0728333977?gbfqdDjduqntjTrle=CrqwoesnXgmuiz,Ispqbmhaw 6F,"Cétèwc khhqp ru nzakbs vzf nctmsxdr.
4 icvgnkcxws ot rqzapef vwscatwcb à dvzxni dmbbdlv i'nctmsxdr
 vcmj zséjlkdkj ot cwzyèhi epszifdz rezsg eujmwxép ngomswoyxoesurl bd 
tgkv nzakbs. Ru geo ot rqzapef pd wfpf ag lxiuht ncupp.
- yn yftmukpc bd yn kjchh ot olhyuvsu odxi lddyvwr vcmj hébzvruh ru vkkecanba rqzapef.
- yn xsemep lxiuht dht jmlçh vcmj hébzvruh eyz rqzapef yhtemèzz bd opawégkr ebs nctmsxdr vzf zédutbwkk i'fjvmlfdbi.
- bvwrjwhnt bd wcsvdbxbe erjhiuy.
- wcsvdbxbe vûqk bd ovgzwtiioéa.",FYK,"Nbdrzzcrgjrd, iuêjoy bd enywbyoask | Zpfrjiffqas",https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w0OeS2IvuwlBVjPI9uuVAtbRxrWLhscv9dNCDcvHPcAiN0HRX1HCRrXYI4MHaoWuyxN3gxYMKcQdItKUD2KWkxGqxcILpzRjXwDBIlK2WhKHB5ZPC5GHUuOLHkZSG1IfBxCwU1RaF4MZCnOzU5EHO2ZVWq,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w5YjQjX3fuvJQhBN9uuVAtbRxrWLhscv9uURUsoDXuHeH1VVO0IXLfXKC5TnI2JvxxD3xfEGatBhp2StG5COX4RqUrOGKeUGKmFXL5HwawIjmrQNGrTFBjVMQ1OXAuRPPtEJW3XTMjXnIvWWJtWlOxCJMd,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3w3BSk2BQfnpPPvHZ9uuVAtbRxrWLhscv9uRKQhwXzsKrE1VVO0PMFpXPd3KLNkBdjwL3w0HRX2HRX0RtTwLlZuJWbfWZM1AFKyDMNdUSSwAiB0JvWwUeGrNyI1ShXzM2Q1UfOdNYYxTTagUdb0N2NqX2Gr`),
				[]byte(`PE468910,https://example.com/HURLH-I%C3%89ZRKUKC/Xiylsi%C3%A8ai-krqxwq/MYUZERCLI-BYWTSDQX-ZZGB-ACZY-12-DOOTAR/k/3028234938?gbfqdDjduqntjTrle=hurlh-opkfeyfn,MYUZERCLI BYWTSDQX ZZGB ACZY 12 DOOTAR,"Pfokkwèzl à bywtsdqx dxmq ibsuhktt 12 dootar sg jrijta hlitxc bd qjnestcbje ot Ugmwn
Pfokkwèzl krqxwq aifnmvjv : uvf cfbfèij qhwfne ot mbla ag bywtsdqx ncyaaotjb yn snûg bilmhcjxrzr nsb zhtdanpbéc ot yqhé.Eyz pfokkwèzl à kqdngukc xykp ebs dgwbw vcmj hébzvruh fc yqhé uphfllad gdèv otlbfcehgl bd bhxjw.",UGMWN,hurlh wépictgq | pfokkwèzl krqxwq ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3lsLny3IXN8dF1wZ2CfgnXeJ3xbfJCxAKKrnZO4L2ubXd8bBSA4UsntFdV1MZOqHM5doOa8JpKzEXW4EsX0CPxmH2FvLXFcI2EwCcAkDqleLaItUaGqY2WjMXGsDWlmBDW5FSHoJOF3PuCsWyOsUzt4IQJ1PC,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3csMPadNIT8dF1wZ2CfgnXeJ3swvAJcWGJqsOMjX2g5Rq8jVLLbGoWfGeRhUaj2Rq5doOa8OHA0PPW1ViC0KGA0OcObEaDeCxgeRqjzMWXyHoCqKJK4GIZlW2LeXIY0AbNmBgS0CwU3PCGpOoMnXNZvUEWjUj,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3wxWFO5SJQ8dF1wZ2CfgnXeJ3ygpLUkZLLtfQG3L2syOR8iOGGlTlImWnP2QlpzLJ5doOa8QSl2GFIdD2MiHDK5YKOcLSTcQSH4PEy0PDLfBOC0TQk3LMY4BvU0TvS1QfsiTrc3JFfaNQD2SMPyRDHqVmcaES`),
				[]byte(`PE305D10,https://example.com/HURLH-I%C3%89ZRKUKC/Xiylsi%C3%A8ai-krqxwq/FRS-HTSJZ-OCIJJVRQM-ZZGB-ACZY/k/1298531048?gbfqdDjduqntjTrle=hurlh-opkfeyfn,FRS HTSJZ OCIJJVRQM ZZGB ACZY,"Uwcbécsjn sszhb yqhé, tèf sszhb félobd bd qjfc sg xczd ot ru utwpnéw dxmq ru pfokkwèzl FRS HTSJZ aczy ocijjvrqm ot UGMWN.
Dxmq ru pfokkwèzl FRS HTSJZ OCIJJVRQM ot Ugmwn,  uwcbécsjn fc céofswwmb 
yqhé, rqâom à uvf nptlhôzy hféexg ot ru dpywégjhdty vcmj eyz jdxyhfmfcg 
iy yqhé epszifdz bd uvf fyd ocijjvrqm igk oehoisnsekh sszhb yqhé sg 
isrjn lnues'à 4 iiwgty fboèu ru geo iy usdvl. Xe nzhehcmq rxhx-qaczrl 
pseixp ot llvugf yn yqhé, rêss lriavdi yn usdvl ot yqhé. Uvf 
cvnsp-krqxwq kmodtefy bd tzsctgsa qyubc waf xe nzhehcmq tsmêk 
oducccbpryy odxi rvnsazletjuj ru fqk sg acupaumjo",UGMWN,hurlh wépictgq | pfokkwèzl krqxwq ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3vsKJP5SAP8dF1wZ2CfgnXeJ3vxxRDtKQTssJV0L2kaDX8rWTYpBjP0OsctQTKqPM5doOa8ApI2YXJfHxSdWwTyOEJ1NRZvZBTpQdSpBBT3MbdeCrYjZZXoUYHfFoSuQKHvCLL5NWRkWZH0Y2WpMFm1IlCrGV,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3ziPUN0DaI8dF1wZ2CfgnXeJ3svyTBpJHNdlVsjG2gdAH8rWTYpBjP0SOBwOIVvDz5doOa8XDKmTdwaZ2DiIuonR2MrTJr3BPW0ECAkVWL3YQWsCfCgInEtYYU1ZsKaU2U5OAKtTDw1KkZ5FJQqMhZ5QBX1HM,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3gaBFlmONo8dF1wZ2CfgnXeJ3ygpLUkZLLtfQG3L2rfEc8gWLKdWAS2PsS3LcW1GN5doOa8EEGpGKLgAFB3BCA1WSM2FQZvZZYpAzXuZThoEKPcRfR5ScihBfCzYYWzF2CzHoD2ANDdVUF5NlF3QFQlG2ZrOW`),
				[]byte(`LR351412,https://example.com/Rqzapef-%C3%A0-jwpdr/Rqzapef-jvwivx%C3%A9a/Zagekr/Zagekr-Oyynpo-Odmgld-/k/4787570329?krc=rqzapef%2bglbthzjv,Zagekr Oyynpo Odmgld,"F'qtfzmolt à zagekr fpubckpx bd ofbértjwx
Psélt vcmj xykp dwgda ot vnbzqrk : pmshsvu, huwoc, drvcdnvp…Xrjqgkf ajapiqm vcmj uvuucsrqb f'kxkèz nsb 5 bspikmdti.Wfpmeznet yvqrdo : dqzoyysu, jkzbbg wsbskpdznvd, bspikmdti bd jslu à zagekr yqqp chgmzknmhai ehiz-zpvtlmikc.Onuopiabfz ehiz-zpvtlmikc : dqzoyysu, jkzbbg wsbskpdznvd, bspikmdti bd jslu à zagekr.",Wohrq,Rqzapef à jwpdr | Rqzapef ifonssép ,https://example.com/micohz/DJ-OYYNPO-KQNAFJG-ZAGEKR-LR351412-DSTTEXLH-405x345.psj?inzkmwz=lWGtrYUsgZKgi3R8Kqq4FMz8dF1wZ2AjlS5oqKr3GX9jQGMtSDq5CnEqFfP1QTBkSG5ouhk8MmZlYaakDYQ4XGOyWWVnBvmqArEjTHZ0ZXWvYwowCCX2YLq5EHH0TYdfXlF4LMHdCzTfNlb5VoSeTOM4M2FdMH,https://example.com/micohz/DJ-OYYNPO-KQNAFJG-ZAGEKR-LR351412-LEQHEC-XIPWODF-TOZDHVNNPX..psj?inzkmwz=eOSlqHBhbZsnIXhws3lgYDT3SvjlyEYoTY9ouhk8dF1wZ2TjX2zeLS9xHNXzNMvxRvV2ORItXJP5MR5ouhk8QwKaAoCaNuskQEM1PWfnOLYjGTZxJGzuKuSuKxG4LZR5ApS4WtYyFwXbGfTmXKW2SPSoO2TqNQA0WUR2OGXqEu,`),
				[]byte(`LR260312,https://example.com/Rqzapef-%C3%A0-jwpdr/Rqzapef-jvwivx%C3%A9a/Zagekr/PHUCXMDMIW-OPJOGWNFJUBWJ-8-RPOVMGHWS/k/6687587445?krc=rqzapef%2bglbthzjv,PHUCXMDMIW OPJOGWNFJUBWJ 8 RPOVMGHWS,"Ru zagekr ag khhqp vatxyqxfjé !
Zagekr Phucxmdmiw : yn mmlunfkd ot ru zagekr !Géjookoiy
 ru zagekr Phucxmdmiw, Psélt vcmj xykp ebs dwgda ot vnbzqrk : 
gxhvtnmalbvvte, fuqutimvwp, eemdgnac bd rêss sg drvcdnvp ag khhqp 
vatxyqxfjé ! Rqâom à uvf cmhof nudgjba, f'qtfzmolt à zagekr jpya êbit 
gxypeffé ag khhqp pépklyné. Nékjug bgrrs ru dpywégjhdty kljtcinéa dxmq 
yn tozdhvnnpx bd m’dht fatzl ! Fc qtfzmolt à zagekr igk ncyaaotjb ot ru 
jfthscm à sszhb jwpdr !",Wohrq,Rqzapef à jwpdr | Rqzapef ifonssép ,https://example.com/micohz/?inzkmwz=eOSlqHBhbZsnIXhws3frFJP4MJxxwXUtDR9uuVAtbRxrWLhscv9xVHErtTJ2UfRqNTFyIIN4ITH4IyM2JvxxD3geVMUtENQxVfW1RWI1XRM0VlMdDODrXXW1QIJmVjM0AVZ2LvNmAJZyEKR1OJYaHoI2PaG5XCc1HYH1SDM0UVAf,https://example.com/micohz/?inzkmwz=lWGtrYUsgZKgi3R8YXPtNzT8dF1wZ2CfgnXeJ3hoTfKqjXN0GgG0JRKeCsM0OgRyQakdThjjT3vnAYBbKwNcQvq0FHwrJrz5UPj1N2U5ZfAlF2CgREWzQVF4MCDcRWR4CLLlGymeZFI3NQCgAFL5MJEfAeA1COX2,`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "528874",
			in: []byte(`"ID_LENILNS","YMGMPN_EIYFYKT","JNA_1_YMGMPN","JNA_2_YMGMPN","IG_NICN","WRJHR","WRJHR_UW","WRJHR_UM","WRJHR_OT","WRJHR_AS","IBQYPUVNYJX","IBQYPUVNYJX_UW","IBQYPUVNYJX_UM","IBQYPUVNYJX_OT","IBQYPUVNYJX_AS","JTBZWWFR","JTBZWWFR_UW","JTBZWWFR_UM","JTBZWWFR_OT","JTBZWWFR_AS","WMU_JTBZWWFR","WMU_JTBZWWFR_UW","WMU_JTBZWWFR_UM","WMU_JTBZWWFR_OT","WMU_JTBZWWFR_AS","HQXHKDUWSXI","HQXHKDUWSXI_UW","HQXHKDUWSXI_UM","HQXHKDUWSXI_OT","HQXHKDUWSXI_AS","WMU_HQXHKDUWSXI","WMU_HQXHKDUWSXI_UW","WMU_HQXHKDUWSXI_UM","WMU_HQXHKDUWSXI_OT","WMU_HQXHKDUWSXI_AS","NKPHL","OCVMVDIPQ","BW_PVQLOJ_ZAARNG","BW_PVQLOJ_ZAARNG_KC","BW_PVQLOJ_ZAARNG_UW","BW_PVQLOJ_ZAARNG_RG","BW_PVQLOJ_ZAARNG_UM","BW_PVQLOJ_ZAARNG_AS","BW_PVQLOJ_ZAARNG_OT","BW_PVQLOJ_ZAARNG_EN","BW_PVQLOJ_ZAARNG_YG","BW_PVQLOJ_ZAARNG_PD","BW_PVQLOJ_ZAARNG_SG","BW_PVQLOJ_ZAARNG_QT","BW_PVQLOJ_ZAARNG_OX","BW_PVQLOJ_ZAARNG_LK","MAOY_ID","WMU_QWLJY_NOTUI","WMU_QWLJY_UBNQYS","WMU_QWLJY_SYQTX","WMU_QWLJY_BGNPF","WMU_QWLJY_WMHS","RENJY","ZMXLLQ","ZMXLLQ_UW","ZMXLLQ_UM","ZMXLLQ_OT","ZMXLLQ_AS","TJGOSNKY","TJGOSNKY_UW","TJGOSNKY_UM","TJGOSNKY_OT","TJGOSNKY_AS","IRNAGF","IRNAGF_UW","IRNAGF_UM","IRNAGF_OT","IRNAGF_AS","PBHLCWEVF_UW","PBHLCWEVF_AG","PBHLCWEVF_UM","PBHLCWEVF_OT","PBHLCWEVF_AS","PBHLCWEVF_YG","PBHLCWEVF_EN","PBHLCWEVF_PD","PBHLCWEVF_OZ","PBHLCWEVF_SG","PBHLCWEVF_ET","PBHLCWEVF_LM","PBHLCWEVF_QT","PBHLCWEVF_OX","WMU_RENJY","KMNDMVC","KMNDMVC_UW","KMNDMVC_UM","KMNDMVC_OT","KMNDMVC_AS","KEGQAJKRI","KEGQAJKRI_UW","KEGQAJKRI_UM","KEGQAJKRI_AS","KEGQAJKRI_OT","KCSQ_UW","KCSQ","KCSQ_OT","KCSQ_AS","KCSQ_UM","NPEZONDW_UW","NPEZONDW_IJ","NPEZONDW_UM","NPEZONDW_OT","NPEZONDW_AS","NPEZONDW_KC","NPEZONDW_YG","NPEZONDW_PD","NPEZONDW_EN","NPEZONDW_QT","NPEZONDW_OX","NPEZONDW_SG","NPEZONDW_ET","NPEZONDW_OZ","NPEZONDW_LM","NPEZONDW_AX","NPEZONDW_LK","WMU_LENILNS","WMU_LENILNS_UW","WMU_LENILNS_IJ","WMU_LENILNS_UM","WMU_LENILNS_OT","WMU_LENILNS_AS","WMU_LENILNS_KC","WMU_LENILNS_YG","WMU_LENILNS_PD","WMU_LENILNS_EN","WMU_LENILNS_SG","WMU_LENILNS_ET","WMU_LENILNS_OZ","WMU_LENILNS_LM","WMU_LENILNS_QT","WMU_LENILNS_OX","WMU_WULWIUT_UW","WMU_WULWIUT_IJ","WMU_WULWIUT_UM","WMU_WULWIUT_OT","WMU_WULWIUT_AS","WMU_WULWIUT_KC","WMU_WULWIUT_YG","WMU_WULWIUT_PD","WMU_WULWIUT_EN","WMU_WULWIUT_SG","WMU_WULWIUT_ET","WMU_WULWIUT_OZ","WMU_WULWIUT_LM","WMU_WULWIUT_QT","WMU_WULWIUT_OX","WMU_WULWIUT_NKPHL_UW","WMU_WULWIUT_NKPHL_IJ","WMU_WULWIUT_NKPHL_UM","WMU_WULWIUT_NKPHL_OT","WMU_WULWIUT_NKPHL_AS","WMU_WULWIUT_NKPHL_KC","WMU_WULWIUT_NKPHL_YG","WMU_WULWIUT_NKPHL_PD","WMU_WULWIUT_NKPHL_EN","WMU_WULWIUT_NKPHL_SG","WMU_WULWIUT_NKPHL_ET","WMU_WULWIUT_NKPHL_OZ","WMU_WULWIUT_NKPHL_LM","WMU_WULWIUT_NKPHL_QT","WMU_WULWIUT_NKPHL_OX","NPEZONDW","XBHOYQNJANT","NXRQZJR","AAIKCG","SPFEGCM","YMGMPN_CJGSXTG_DISWSNW","MO_IXBVZEASRFZ_HSSX","LENILNS_SUGKMPN","MSWB_EDJSEQDED","STCFIDHPC","LENILNS_DEFH_OHAEMSCV_RENJY","IRMDZRMMDVVQ_BYBNA","ITJKAEBGYY_BYBNA","SDYYI_BYBNA","QPTRX_BYBNA","TYID_BYBNA","GGAPDKF_BYBNA","NH_IUIO","CNI_7I","NW_WFRIJKOPUR","SDYYI_FZK","SDYYI_HXK","SDYYI_BKJ","SDYYI_XSR","SDYYI_VLH","SDYYI_CGX","SDYYI_CYG","SDYYI_VFK","SDYYI_YIP","SDYYI_SQO","SDYYI_JWW","SDYYI_MPA","SDYYI_DKM","RPYG_SDYYI_FZK","RPYG_SDYYI_HXK","RPYG_SDYYI_BKJ","RPYG_SDYYI_XSR","RPYG_SDYYI_VLH","RPYG_SDYYI_CGX","RPYG_SDYYI_CYG","RPYG_SDYYI_VFK","RPYG_SDYYI_YIP","RPYG_SDYYI_SQO","RPYG_SDYYI_JWW","RPYG_SDYYI_MPA","RPYG_SDYYI_DKM","NW_PJVTRG","NW_THB_NFDZOZ","NW_QSEGXMONC_NFDZOZ","CEVLD","ZNUMJLD_FWF","HXNI","KARPLHFC_KC","KARPLHFC_IJ","KARPLHFC_PGOMTU","KARPLHFC_LAYA_ZW_TUQO_AS","KARPLHFC_AZJGS","KARPLHFC_LSEI_ZW_HSBMS","KARPLHFC_LSEI_ZW_WJ","KARPLHFC_QT","KARPLHFC_TGQDMIQSC","KARPLHFC_ZQVKRIZFJ","KARPLHFC_LVTPOM","KARPLHFC_LSEI_TH_AWBLX","GEUDUYAA_BYBNA","OWJJSY_LIZO","NW_WFTQ_UJDK","IBQYPUVNYJX_LK","JTBZWWFR_LK","HQXHKDUWSXI_LK","ZMXLLQ_LK","TJGOSNKY_LK","IRNAGF_LK","PBHLCWEVF_LK","KMNDMVC_LK","KEGQAJKRI_LK"
3612656,"Pgomtu","QCQA","PGOMTU","611490","Imhugkzg","Gmwmjxmb","Qpizuz","Ufebmnek","Eaayaoyk","Kzmay gmwmjxmb, ji Ilai, sznktgklm vgei yono yorgyp uc fip hqxnh sqe uc fip uwpmwj.\r\nVgei svtpsenc cftz.","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","RQWFCLEM","YÊVURJNVH","SBTO","JTCLWCHPFKEÜEOG","AZVZMAIBMIDCC","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","IMHUGKZG","OTTDJ.EMEMCV","RNHGQLI","UFEBMNEK","FQDDIXPT. VVOCE","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","",1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,"https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","Ilai","kzmay","Kzmay","kzmay","kzmay","kzmay","Txwjvf","Wcuai","Flciyóx","Avevdixgb","Invirn","Wniddr","Aqmlq","Laqui","Zbrzd","Sppfa","Aqmlq > YÊVURJNVH > OTTDJ.EMEMCV","Dunyb > RQWFCLEM > IMHUGKZG","Laqui > SBTO > RNHGQLI","Zbrzd > JTCLWCHPFKEÜEOG > UFEBMNEK","Sppfa > AZVZMAIBMIDCC > FQDDIXPT. VVOCE","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","okao-qctt/Ilai","Rwhct","Uhb","Jlarmsj","Uhb","Prahq djdldexr","Zaht kegqajkri","Gqm émgu","Dicn varrcq","Oaxhw chntk","Azmxu Kanmrtj","","","","","","FZK","BKJ","FZK","FZK","FZK","HXK","XSR","VLH","FZK","CGX","CYG","VFK","YIP","SQO","JWW","MPA","DKM","http://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/kzmay-qqxtnha-qpizuz-ilai-3612656.dsoyd?gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/kzmay-avevdixgb-ufebmnek-ilai-3612656.dsoyd?gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/kzmay-invirn-eaayaoyk-ilai-3612656.dsoyd?gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/?lenilns_id=3612656&gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/?lenilns_id=3612656&gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/?lenilns_id=3612656&gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","","","","","","","","","","","","","","","","FZK",0,0,0,"M2M","1","1","0.0961512076","2017-02-13 22:43:31.000","2022-07-01 13:08:12.211 +0200",yslts,-1.625301941,0,-0.7368412799,2,-1.302998143,0.5884873619,0,"",1,95.00,107.95,89.87,777.58,1117.21,866.62,152.99,159.05,142.43,489.90,105.75,14912.88,139117.84,55.00,62.50,52.03,450.18,646.81,501.73,88.57,92.08,82.46,283.62,61.23,8633.77,80541.91,yiit,"","",22,3,"Pgomtu","20 HXK","20 BKJ","15 FZK","15 FZK","15 FZK","25 FZK","30 FZK","200 CGX","60 VFK","30 CYG","40 YIP","60 HXK",3,11,"","","햃본","떅씬똮폫","혺컘싞","찚","쐋쯕겤","쐋쯕겤 > 햃본 > 떅씬똮폫","괾쉍뾺","봌휰"`),
			multiOut: [][]byte{
				[]byte(`"ID_LENILNS","YMGMPN_EIYFYKT","JNA_1_YMGMPN","JNA_2_YMGMPN","IG_NICN","WRJHR","WRJHR_UW","WRJHR_UM","WRJHR_OT","WRJHR_AS","IBQYPUVNYJX","IBQYPUVNYJX_UW","IBQYPUVNYJX_UM","IBQYPUVNYJX_OT","IBQYPUVNYJX_AS","JTBZWWFR","JTBZWWFR_UW","JTBZWWFR_UM","JTBZWWFR_OT","JTBZWWFR_AS","WMU_JTBZWWFR","WMU_JTBZWWFR_UW","WMU_JTBZWWFR_UM","WMU_JTBZWWFR_OT","WMU_JTBZWWFR_AS","HQXHKDUWSXI","HQXHKDUWSXI_UW","HQXHKDUWSXI_UM","HQXHKDUWSXI_OT","HQXHKDUWSXI_AS","WMU_HQXHKDUWSXI","WMU_HQXHKDUWSXI_UW","WMU_HQXHKDUWSXI_UM","WMU_HQXHKDUWSXI_OT","WMU_HQXHKDUWSXI_AS","NKPHL","OCVMVDIPQ","BW_PVQLOJ_ZAARNG","BW_PVQLOJ_ZAARNG_KC","BW_PVQLOJ_ZAARNG_UW","BW_PVQLOJ_ZAARNG_RG","BW_PVQLOJ_ZAARNG_UM","BW_PVQLOJ_ZAARNG_AS","BW_PVQLOJ_ZAARNG_OT","BW_PVQLOJ_ZAARNG_EN","BW_PVQLOJ_ZAARNG_YG","BW_PVQLOJ_ZAARNG_PD","BW_PVQLOJ_ZAARNG_SG","BW_PVQLOJ_ZAARNG_QT","BW_PVQLOJ_ZAARNG_OX","BW_PVQLOJ_ZAARNG_LK","MAOY_ID","WMU_QWLJY_NOTUI","WMU_QWLJY_UBNQYS","WMU_QWLJY_SYQTX","WMU_QWLJY_BGNPF","WMU_QWLJY_WMHS","RENJY","ZMXLLQ","ZMXLLQ_UW","ZMXLLQ_UM","ZMXLLQ_OT","ZMXLLQ_AS","TJGOSNKY","TJGOSNKY_UW","TJGOSNKY_UM","TJGOSNKY_OT","TJGOSNKY_AS","IRNAGF","IRNAGF_UW","IRNAGF_UM","IRNAGF_OT","IRNAGF_AS","PBHLCWEVF_UW","PBHLCWEVF_AG","PBHLCWEVF_UM","PBHLCWEVF_OT","PBHLCWEVF_AS","PBHLCWEVF_YG","PBHLCWEVF_EN","PBHLCWEVF_PD","PBHLCWEVF_OZ","PBHLCWEVF_SG","PBHLCWEVF_ET","PBHLCWEVF_LM","PBHLCWEVF_QT","PBHLCWEVF_OX","WMU_RENJY","KMNDMVC","KMNDMVC_UW","KMNDMVC_UM","KMNDMVC_OT","KMNDMVC_AS","KEGQAJKRI","KEGQAJKRI_UW","KEGQAJKRI_UM","KEGQAJKRI_AS","KEGQAJKRI_OT","KCSQ_UW","KCSQ","KCSQ_OT","KCSQ_AS","KCSQ_UM","NPEZONDW_UW","NPEZONDW_IJ","NPEZONDW_UM","NPEZONDW_OT","NPEZONDW_AS","NPEZONDW_KC","NPEZONDW_YG","NPEZONDW_PD","NPEZONDW_EN","NPEZONDW_QT","NPEZONDW_OX","NPEZONDW_SG","NPEZONDW_ET","NPEZONDW_OZ","NPEZONDW_LM","NPEZONDW_AX","NPEZONDW_LK","WMU_LENILNS","WMU_LENILNS_UW","WMU_LENILNS_IJ","WMU_LENILNS_UM","WMU_LENILNS_OT","WMU_LENILNS_AS","WMU_LENILNS_KC","WMU_LENILNS_YG","WMU_LENILNS_PD","WMU_LENILNS_EN","WMU_LENILNS_SG","WMU_LENILNS_ET","WMU_LENILNS_OZ","WMU_LENILNS_LM","WMU_LENILNS_QT","WMU_LENILNS_OX","WMU_WULWIUT_UW","WMU_WULWIUT_IJ","WMU_WULWIUT_UM","WMU_WULWIUT_OT","WMU_WULWIUT_AS","WMU_WULWIUT_KC","WMU_WULWIUT_YG","WMU_WULWIUT_PD","WMU_WULWIUT_EN","WMU_WULWIUT_SG","WMU_WULWIUT_ET","WMU_WULWIUT_OZ","WMU_WULWIUT_LM","WMU_WULWIUT_QT","WMU_WULWIUT_OX","WMU_WULWIUT_NKPHL_UW","WMU_WULWIUT_NKPHL_IJ","WMU_WULWIUT_NKPHL_UM","WMU_WULWIUT_NKPHL_OT","WMU_WULWIUT_NKPHL_AS","WMU_WULWIUT_NKPHL_KC","WMU_WULWIUT_NKPHL_YG","WMU_WULWIUT_NKPHL_PD","WMU_WULWIUT_NKPHL_EN","WMU_WULWIUT_NKPHL_SG","WMU_WULWIUT_NKPHL_ET","WMU_WULWIUT_NKPHL_OZ","WMU_WULWIUT_NKPHL_LM","WMU_WULWIUT_NKPHL_QT","WMU_WULWIUT_NKPHL_OX","NPEZONDW","XBHOYQNJANT","NXRQZJR","AAIKCG","SPFEGCM","YMGMPN_CJGSXTG_DISWSNW","MO_IXBVZEASRFZ_HSSX","LENILNS_SUGKMPN","MSWB_EDJSEQDED","STCFIDHPC","LENILNS_DEFH_OHAEMSCV_RENJY","IRMDZRMMDVVQ_BYBNA","ITJKAEBGYY_BYBNA","SDYYI_BYBNA","QPTRX_BYBNA","TYID_BYBNA","GGAPDKF_BYBNA","NH_IUIO","CNI_7I","NW_WFRIJKOPUR","SDYYI_FZK","SDYYI_HXK","SDYYI_BKJ","SDYYI_XSR","SDYYI_VLH","SDYYI_CGX","SDYYI_CYG","SDYYI_VFK","SDYYI_YIP","SDYYI_SQO","SDYYI_JWW","SDYYI_MPA","SDYYI_DKM","RPYG_SDYYI_FZK","RPYG_SDYYI_HXK","RPYG_SDYYI_BKJ","RPYG_SDYYI_XSR","RPYG_SDYYI_VLH","RPYG_SDYYI_CGX","RPYG_SDYYI_CYG","RPYG_SDYYI_VFK","RPYG_SDYYI_YIP","RPYG_SDYYI_SQO","RPYG_SDYYI_JWW","RPYG_SDYYI_MPA","RPYG_SDYYI_DKM","NW_PJVTRG","NW_THB_NFDZOZ","NW_QSEGXMONC_NFDZOZ","CEVLD","ZNUMJLD_FWF","HXNI","KARPLHFC_KC","KARPLHFC_IJ","KARPLHFC_PGOMTU","KARPLHFC_LAYA_ZW_TUQO_AS","KARPLHFC_AZJGS","KARPLHFC_LSEI_ZW_HSBMS","KARPLHFC_LSEI_ZW_WJ","KARPLHFC_QT","KARPLHFC_TGQDMIQSC","KARPLHFC_ZQVKRIZFJ","KARPLHFC_LVTPOM","KARPLHFC_LSEI_TH_AWBLX","GEUDUYAA_BYBNA","OWJJSY_LIZO","NW_WFTQ_UJDK","IBQYPUVNYJX_LK","JTBZWWFR_LK","HQXHKDUWSXI_LK","ZMXLLQ_LK","TJGOSNKY_LK","IRNAGF_LK","PBHLCWEVF_LK","KMNDMVC_LK","KEGQAJKRI_LK"`),
				[]byte(`3612656,"Pgomtu","QCQA","PGOMTU","611490","Imhugkzg","Gmwmjxmb","Qpizuz","Ufebmnek","Eaayaoyk","Kzmay gmwmjxmb, ji Ilai, sznktgklm vgei yono yorgyp uc fip hqxnh sqe uc fip uwpmwj.\r\nVgei svtpsenc cftz.","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
		Dxmq égdskkotjh","RQWFCLEM","YÊVURJNVH","SBTO","JTCLWCHPFKEÜEOG","AZVZMAIBMIDCC","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","IMHUGKZG","OTTDJ.EMEMCV","RNHGQLI","UFEBMNEK","FQDDIXPT. VVOCE","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","",1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,"https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","Ilai","kzmay","Kzmay","kzmay","kzmay","kzmay","Txwjvf","Wcuai","Flciyóx","Avevdixgb","Invirn","Wniddr","Aqmlq","Laqui","Zbrzd","Sppfa","Aqmlq > YÊVURJNVH > OTTDJ.EMEMCV","Dunyb > RQWFCLEM > IMHUGKZG","Laqui > SBTO > RNHGQLI","Zbrzd > JTCLWCHPFKEÜEOG > UFEBMNEK","Sppfa > AZVZMAIBMIDCC > FQDDIXPT. VVOCE","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","okao-qctt/Ilai","Rwhct","Uhb","Jlarmsj","Uhb","Prahq djdldexr","Zaht kegqajkri","Gqm émgu","Dicn varrcq","Oaxhw chntk","Azmxu Kanmrtj","","","","","","FZK","BKJ","FZK","FZK","FZK","HXK","XSR","VLH","FZK","CGX","CYG","VFK","YIP","SQO","JWW","MPA","DKM","http://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/kzmay-qqxtnha-qpizuz-ilai-3612656.dsoyd?gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/kzmay-avevdixgb-ufebmnek-ilai-3612656.dsoyd?gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/kzmay-invirn-eaayaoyk-ilai-3612656.dsoyd?gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/?lenilns_id=3612656&gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/?lenilns_id=3612656&gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/?lenilns_id=3612656&gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","","","","","","","","","","","","","","","","FZK",0,0,0,"M2M","1","1","0.0961512076","2017-02-13 22:43:31.000","2022-07-01 13:08:12.211 +0200",yslts,-1.625301941,0,-0.7368412799,2,-1.302998143,0.5884873619,0,"",1,95.00,107.95,89.87,777.58,1117.21,866.62,152.99,159.05,142.43,489.90,105.75,14912.88,139117.84,55.00,62.50,52.03,450.18,646.81,501.73,88.57,92.08,82.46,283.62,61.23,8633.77,80541.91,yiit,"","",22,3,"Pgomtu","20 HXK","20 BKJ","15 FZK","15 FZK","15 FZK","25 FZK","30 FZK","200 CGX","60 VFK","30 CYG","40 YIP","60 HXK",3,11,"","","햃본","떅씬똮폫","혺컘싞","찚","쐋쯕겤","쐋쯕겤 > 햃본 > 떅씬똮폫","괾쉍뾺","봌휰"`),
			},
			advance:   3843,
			out:       []byte(`"ID_LENILNS","YMGMPN_EIYFYKT","JNA_1_YMGMPN","JNA_2_YMGMPN","IG_NICN","WRJHR","WRJHR_UW","WRJHR_UM","WRJHR_OT","WRJHR_AS","IBQYPUVNYJX","IBQYPUVNYJX_UW","IBQYPUVNYJX_UM","IBQYPUVNYJX_OT","IBQYPUVNYJX_AS","JTBZWWFR","JTBZWWFR_UW","JTBZWWFR_UM","JTBZWWFR_OT","JTBZWWFR_AS","WMU_JTBZWWFR","WMU_JTBZWWFR_UW","WMU_JTBZWWFR_UM","WMU_JTBZWWFR_OT","WMU_JTBZWWFR_AS","HQXHKDUWSXI","HQXHKDUWSXI_UW","HQXHKDUWSXI_UM","HQXHKDUWSXI_OT","HQXHKDUWSXI_AS","WMU_HQXHKDUWSXI","WMU_HQXHKDUWSXI_UW","WMU_HQXHKDUWSXI_UM","WMU_HQXHKDUWSXI_OT","WMU_HQXHKDUWSXI_AS","NKPHL","OCVMVDIPQ","BW_PVQLOJ_ZAARNG","BW_PVQLOJ_ZAARNG_KC","BW_PVQLOJ_ZAARNG_UW","BW_PVQLOJ_ZAARNG_RG","BW_PVQLOJ_ZAARNG_UM","BW_PVQLOJ_ZAARNG_AS","BW_PVQLOJ_ZAARNG_OT","BW_PVQLOJ_ZAARNG_EN","BW_PVQLOJ_ZAARNG_YG","BW_PVQLOJ_ZAARNG_PD","BW_PVQLOJ_ZAARNG_SG","BW_PVQLOJ_ZAARNG_QT","BW_PVQLOJ_ZAARNG_OX","BW_PVQLOJ_ZAARNG_LK","MAOY_ID","WMU_QWLJY_NOTUI","WMU_QWLJY_UBNQYS","WMU_QWLJY_SYQTX","WMU_QWLJY_BGNPF","WMU_QWLJY_WMHS","RENJY","ZMXLLQ","ZMXLLQ_UW","ZMXLLQ_UM","ZMXLLQ_OT","ZMXLLQ_AS","TJGOSNKY","TJGOSNKY_UW","TJGOSNKY_UM","TJGOSNKY_OT","TJGOSNKY_AS","IRNAGF","IRNAGF_UW","IRNAGF_UM","IRNAGF_OT","IRNAGF_AS","PBHLCWEVF_UW","PBHLCWEVF_AG","PBHLCWEVF_UM","PBHLCWEVF_OT","PBHLCWEVF_AS","PBHLCWEVF_YG","PBHLCWEVF_EN","PBHLCWEVF_PD","PBHLCWEVF_OZ","PBHLCWEVF_SG","PBHLCWEVF_ET","PBHLCWEVF_LM","PBHLCWEVF_QT","PBHLCWEVF_OX","WMU_RENJY","KMNDMVC","KMNDMVC_UW","KMNDMVC_UM","KMNDMVC_OT","KMNDMVC_AS","KEGQAJKRI","KEGQAJKRI_UW","KEGQAJKRI_UM","KEGQAJKRI_AS","KEGQAJKRI_OT","KCSQ_UW","KCSQ","KCSQ_OT","KCSQ_AS","KCSQ_UM","NPEZONDW_UW","NPEZONDW_IJ","NPEZONDW_UM","NPEZONDW_OT","NPEZONDW_AS","NPEZONDW_KC","NPEZONDW_YG","NPEZONDW_PD","NPEZONDW_EN","NPEZONDW_QT","NPEZONDW_OX","NPEZONDW_SG","NPEZONDW_ET","NPEZONDW_OZ","NPEZONDW_LM","NPEZONDW_AX","NPEZONDW_LK","WMU_LENILNS","WMU_LENILNS_UW","WMU_LENILNS_IJ","WMU_LENILNS_UM","WMU_LENILNS_OT","WMU_LENILNS_AS","WMU_LENILNS_KC","WMU_LENILNS_YG","WMU_LENILNS_PD","WMU_LENILNS_EN","WMU_LENILNS_SG","WMU_LENILNS_ET","WMU_LENILNS_OZ","WMU_LENILNS_LM","WMU_LENILNS_QT","WMU_LENILNS_OX","WMU_WULWIUT_UW","WMU_WULWIUT_IJ","WMU_WULWIUT_UM","WMU_WULWIUT_OT","WMU_WULWIUT_AS","WMU_WULWIUT_KC","WMU_WULWIUT_YG","WMU_WULWIUT_PD","WMU_WULWIUT_EN","WMU_WULWIUT_SG","WMU_WULWIUT_ET","WMU_WULWIUT_OZ","WMU_WULWIUT_LM","WMU_WULWIUT_QT","WMU_WULWIUT_OX","WMU_WULWIUT_NKPHL_UW","WMU_WULWIUT_NKPHL_IJ","WMU_WULWIUT_NKPHL_UM","WMU_WULWIUT_NKPHL_OT","WMU_WULWIUT_NKPHL_AS","WMU_WULWIUT_NKPHL_KC","WMU_WULWIUT_NKPHL_YG","WMU_WULWIUT_NKPHL_PD","WMU_WULWIUT_NKPHL_EN","WMU_WULWIUT_NKPHL_SG","WMU_WULWIUT_NKPHL_ET","WMU_WULWIUT_NKPHL_OZ","WMU_WULWIUT_NKPHL_LM","WMU_WULWIUT_NKPHL_QT","WMU_WULWIUT_NKPHL_OX","NPEZONDW","XBHOYQNJANT","NXRQZJR","AAIKCG","SPFEGCM","YMGMPN_CJGSXTG_DISWSNW","MO_IXBVZEASRFZ_HSSX","LENILNS_SUGKMPN","MSWB_EDJSEQDED","STCFIDHPC","LENILNS_DEFH_OHAEMSCV_RENJY","IRMDZRMMDVVQ_BYBNA","ITJKAEBGYY_BYBNA","SDYYI_BYBNA","QPTRX_BYBNA","TYID_BYBNA","GGAPDKF_BYBNA","NH_IUIO","CNI_7I","NW_WFRIJKOPUR","SDYYI_FZK","SDYYI_HXK","SDYYI_BKJ","SDYYI_XSR","SDYYI_VLH","SDYYI_CGX","SDYYI_CYG","SDYYI_VFK","SDYYI_YIP","SDYYI_SQO","SDYYI_JWW","SDYYI_MPA","SDYYI_DKM","RPYG_SDYYI_FZK","RPYG_SDYYI_HXK","RPYG_SDYYI_BKJ","RPYG_SDYYI_XSR","RPYG_SDYYI_VLH","RPYG_SDYYI_CGX","RPYG_SDYYI_CYG","RPYG_SDYYI_VFK","RPYG_SDYYI_YIP","RPYG_SDYYI_SQO","RPYG_SDYYI_JWW","RPYG_SDYYI_MPA","RPYG_SDYYI_DKM","NW_PJVTRG","NW_THB_NFDZOZ","NW_QSEGXMONC_NFDZOZ","CEVLD","ZNUMJLD_FWF","HXNI","KARPLHFC_KC","KARPLHFC_IJ","KARPLHFC_PGOMTU","KARPLHFC_LAYA_ZW_TUQO_AS","KARPLHFC_AZJGS","KARPLHFC_LSEI_ZW_HSBMS","KARPLHFC_LSEI_ZW_WJ","KARPLHFC_QT","KARPLHFC_TGQDMIQSC","KARPLHFC_ZQVKRIZFJ","KARPLHFC_LVTPOM","KARPLHFC_LSEI_TH_AWBLX","GEUDUYAA_BYBNA","OWJJSY_LIZO","NW_WFTQ_UJDK","IBQYPUVNYJX_LK","JTBZWWFR_LK","HQXHKDUWSXI_LK","ZMXLLQ_LK","TJGOSNKY_LK","IRNAGF_LK","PBHLCWEVF_LK","KMNDMVC_LK","KEGQAJKRI_LK"`),
			delimiter: characters.Character_COMMA,
			debug:     false,
		},
		{
			name: "1084854.00215778 8341170105289",
			in: []byte(`bzi_rl;bzi_name;bzi_bcihkzddhflf;bzi_wmu;bzi_fmy_wmu;bzi_fmy_wmu2;bzi_fmy_wmu3;bzi_fmy_wmu4;bzi_fmy_wmu5;bzi_fmy_wmu6;bzi_fmy_wmu7;bzi_fmy_wmu8;bzi_fmy_wmu9;bzi_kvzajrma;bzi_yeegp;bzi_tndfkclagyrd;bzi_sgdko;bzi_pdg;umr-zqm-izc;bzi_vdxjbadlervz;bzi_mcerokyxwj;bzi_sbvqo;bzi_rovmeysdtn;bzi_tjgosnky;bzi_fdfqs;bzi_ncwxupevz;bzi_myapv;bzi_myapv_ubhwkkb;bzi_utalb;bzi_utalb_ubhwkkb;bzi_dgsbav;bzi_dgsbav_ubhwkkb;bzi_rsdilbb.wiqRwletmv;bzi_rsdilbb.ydpxemDhfhvi;
"500133";Hjgmsxfqbjz Lobsozgzazvo: Sclaznrröfhrxkw Ysfoy;"Gfaysvsit Jrftrqxzeet hl gogxagw Sclaznrröfhrxkw. Opk Jrftrqxzeet msuävzbg iy roiodwd hl opk Vöyptiev dc Ltößu S/P fjc F/AR. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp   Nxjfjf 0 21 yslts yslts yslts OT L-DFTR L-DFTR /* Qptrx Iztxwlfdwua */ jwpdr.RluVvjontDnusa {ron-qptrx-name:""Gfcsdqz Bkpwxhw""; ron-genupy-qopluij-kcsq:0; ron-genupy-rpdixqr-kcsq:0; ron-qptrx-joflfx:roj; ron-qptrx-jbabytaj:99; ron-qptrx-vfdpux:""""; ron-msazdlb-zrb:0he 5.4sq 0he 5.4sq; ron-cihx-ampffd-umr:0he; ron-cihx-ampffd-talue:0he; ron-cihx-ampffd-tbkwcs:8.0sq; ron-cihx-ampffd-xweb:0he; xxto-btyliv:107%; ron-mjssqbdqaj:toihl-ludunw; leip-kcsq:11.0sq; leip-ebuoai:""Owfliwh"",xylc-lfxbl; ron-ziznc-leip-ebuoai:Owfliwh; ron-ziznc-yihrp-leip:bhmdz-lbjgv; ron-rimpu-leip-ebuoai:Owfliwh; ron-rimpu-yihrp-leip:bhmdz-lbjgv; ron-ykeu-leip-ebuoai:""Usfyp Fxm Enklk""; ron-ykeu-yihrp-leip:bhmdz-ykeu; ron-lnrxhgr-tbisnhjt:AG-KC;}";"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/357/hjgmsxfqbjz-lobsozgzazvo-rcuraztqxpwzuonz-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs";https://example.com/nnpma/qwljy/a9/0a/i7/Upsyc_250821_30399_600x600.knn;"https://example.com/nnpma/qwljy/a8/x7/a5/Upsyc_250821_30398_1280x1280.knn";"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn";;;;;;;FZK;29,00;5.99;500133;"8341170105289";zqm;0.00;3-4 Udnk;Gvmih;Gvmih;;"UPSYC";Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg;;he;;he;;he;0.32;90;
"500134";Asxmd rüt Brntkrsqpptysdg Kweamljf Ysfoy;"Gfaysvsit Jrftrqxzeet hl gogxagw Kweamljf-Odhnqqaxdey. Opk Jrftrqxzeet msuävzbg iy roiodwd hl nxagqb Vadaqwsqfx dc opk Spößze S/P - HJU. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp";"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/358/asxmd-eylw-brntkrsqpptysdg-kweamljf-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs";https://example.com/nnpma/qwljy/or/13/7i/Upsyc_250821_30390aBMThEFu3RAED_600x600.knn;"https://example.com/nnpma/qwljy/a2/3u/9h/Upsyc_250821_30379QDmySWY6mtCjg_1280x1280.knn";"https://example.com/nnpma/qwljy/18/i4/6a/Upsyc_250821_30378VZKyuG7UTxwsR_1280x1280.knn";"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn";;;;;;FZK;64,00;5.99;500134;"2415730296599";zqm;0.00;3-4 Udnk;Gvmih;Gvmih;;"UPSYC";Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg;;he;;he;;he;0.32;90;`),
			multiOut: [][]byte{
				[]byte(`bzi_rl;bzi_name;bzi_bcihkzddhflf;bzi_wmu;bzi_fmy_wmu;bzi_fmy_wmu2;bzi_fmy_wmu3;bzi_fmy_wmu4;bzi_fmy_wmu5;bzi_fmy_wmu6;bzi_fmy_wmu7;bzi_fmy_wmu8;bzi_fmy_wmu9;bzi_kvzajrma;bzi_yeegp;bzi_tndfkclagyrd;bzi_sgdko;bzi_pdg;umr-zqm-izc;bzi_vdxjbadlervz;bzi_mcerokyxwj;bzi_sbvqo;bzi_rovmeysdtn;bzi_tjgosnky;bzi_fdfqs;bzi_ncwxupevz;bzi_myapv;bzi_myapv_ubhwkkb;bzi_utalb;bzi_utalb_ubhwkkb;bzi_dgsbav;bzi_dgsbav_ubhwkkb;bzi_rsdilbb.wiqRwletmv;bzi_rsdilbb.ydpxemDhfhvi;`),
				[]byte(`"500133";Hjgmsxfqbjz Lobsozgzazvo: Sclaznrröfhrxkw Ysfoy;"Gfaysvsit Jrftrqxzeet hl gogxagw Sclaznrröfhrxkw. Opk Jrftrqxzeet msuävzbg iy roiodwd hl opk Vöyptiev dc Ltößu S/P fjc F/AR. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp   Nxjfjf 0 21 yslts yslts yslts OT L-DFTR L-DFTR /* Qptrx Iztxwlfdwua */ jwpdr.RluVvjontDnusa {ron-qptrx-name:""Gfcsdqz Bkpwxhw""; ron-genupy-qopluij-kcsq:0; ron-genupy-rpdixqr-kcsq:0; ron-qptrx-joflfx:roj; ron-qptrx-jbabytaj:99; ron-qptrx-vfdpux:""""; ron-msazdlb-zrb:0he 5.4sq 0he 5.4sq; ron-cihx-ampffd-umr:0he; ron-cihx-ampffd-talue:0he; ron-cihx-ampffd-tbkwcs:8.0sq; ron-cihx-ampffd-xweb:0he; xxto-btyliv:107%; ron-mjssqbdqaj:toihl-ludunw; leip-kcsq:11.0sq; leip-ebuoai:""Owfliwh"",xylc-lfxbl; ron-ziznc-leip-ebuoai:Owfliwh; ron-ziznc-yihrp-leip:bhmdz-lbjgv; ron-rimpu-leip-ebuoai:Owfliwh; ron-rimpu-yihrp-leip:bhmdz-lbjgv; ron-ykeu-leip-ebuoai:""Usfyp Fxm Enklk""; ron-ykeu-yihrp-leip:bhmdz-ykeu; ron-lnrxhgr-tbisnhjt:AG-KC;}";"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/357/hjgmsxfqbjz-lobsozgzazvo-rcuraztqxpwzuonz-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs";https://example.com/nnpma/qwljy/a9/0a/i7/Upsyc_250821_30399_600x600.knn;"https://example.com/nnpma/qwljy/a8/x7/a5/Upsyc_250821_30398_1280x1280.knn";"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn";;;;;;;FZK;29,00;5.99;500133;"8341170105289";zqm;0.00;3-4 Udnk;Gvmih;Gvmih;;"UPSYC";Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg;;he;;he;;he;0.32;90;`),
				[]byte(`"500134";Asxmd rüt Brntkrsqpptysdg Kweamljf Ysfoy;"Gfaysvsit Jrftrqxzeet hl gogxagw Kweamljf-Odhnqqaxdey. Opk Jrftrqxzeet msuävzbg iy roiodwd hl nxagqb Vadaqwsqfx dc opk Spößze S/P - HJU. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp";"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/358/asxmd-eylw-brntkrsqpptysdg-kweamljf-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs";https://example.com/nnpma/qwljy/or/13/7i/Upsyc_250821_30390aBMThEFu3RAED_600x600.knn;"https://example.com/nnpma/qwljy/a2/3u/9h/Upsyc_250821_30379QDmySWY6mtCjg_1280x1280.knn";"https://example.com/nnpma/qwljy/18/i4/6a/Upsyc_250821_30378VZKyuG7UTxwsR_1280x1280.knn";"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn";;;;;;FZK;64,00;5.99;500134;"2415730296599";zqm;0.00;3-4 Udnk;Gvmih;Gvmih;;"UPSYC";Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg;;he;;he;;he;0.32;90;`),
			},
			advance: 456,
			out:     []byte(`bzi_rl;bzi_name;bzi_bcihkzddhflf;bzi_wmu;bzi_fmy_wmu;bzi_fmy_wmu2;bzi_fmy_wmu3;bzi_fmy_wmu4;bzi_fmy_wmu5;bzi_fmy_wmu6;bzi_fmy_wmu7;bzi_fmy_wmu8;bzi_fmy_wmu9;bzi_kvzajrma;bzi_yeegp;bzi_tndfkclagyrd;bzi_sgdko;bzi_pdg;umr-zqm-izc;bzi_vdxjbadlervz;bzi_mcerokyxwj;bzi_sbvqo;bzi_rovmeysdtn;bzi_tjgosnky;bzi_fdfqs;bzi_ncwxupevz;bzi_myapv;bzi_myapv_ubhwkkb;bzi_utalb;bzi_utalb_ubhwkkb;bzi_dgsbav;bzi_dgsbav_ubhwkkb;bzi_rsdilbb.wiqRwletmv;bzi_rsdilbb.ydpxemDhfhvi;`),
			// multiOut: [][]byte{
			// 	[]byte(`"500133"`),
			// 	[]byte(`Hjgmsxfqbjz Lobsozgzazvo: Sclaznrröfhrxkw Ysfoy`),
			// 	[]byte(`"Gfaysvsit Jrftrqxzeet hl gogxagw Sclaznrröfhrxkw. Opk Jrftrqxzeet msuävzbg iy roiodwd hl opk Vöyptiev dc Ltößu S/P fjc F/AR. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp   Nxjfjf 0 21 yslts yslts yslts OT L-DFTR L-DFTR /* Qptrx Iztxwlfdwua */ jwpdr.RluVvjontDnusa {ron-qptrx-name:""Gfcsdqz Bkpwxhw""; ron-genupy-qopluij-kcsq:0; ron-genupy-rpdixqr-kcsq:0; ron-qptrx-joflfx:roj; ron-qptrx-jbabytaj:99; ron-qptrx-vfdpux:""""; ron-msazdlb-zrb:0he 5.4sq 0he 5.4sq; ron-cihx-ampffd-umr:0he; ron-cihx-ampffd-talue:0he; ron-cihx-ampffd-tbkwcs:8.0sq; ron-cihx-ampffd-xweb:0he; xxto-btyliv:107%; ron-mjssqbdqaj:toihl-ludunw; leip-kcsq:11.0sq; leip-ebuoai:""Owfliwh"",xylc-lfxbl; ron-ziznc-leip-ebuoai:Owfliwh; ron-ziznc-yihrp-leip:bhmdz-lbjgv; ron-rimpu-leip-ebuoai:Owfliwh; ron-rimpu-yihrp-leip:bhmdz-lbjgv; ron-ykeu-leip-ebuoai:""Usfyp Fxm Enklk""; ron-ykeu-yihrp-leip:bhmdz-ykeu; ron-lnrxhgr-tbisnhjt:AG-KC;}"`),
			// 	[]byte(`"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/357/hjgmsxfqbjz-lobsozgzazvo-rcuraztqxpwzuonz-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs"`),
			// 	[]byte(`https://example.com/nnpma/qwljy/a9/0a/i7/Upsyc_250821_30399_600x600.knn`),
			// 	[]byte(`"https://example.com/nnpma/qwljy/a8/x7/a5/Upsyc_250821_30398_1280x1280.knn"`),
			// 	[]byte(`"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn"`),
			// 	[]byte(`;`), []byte(`;`), []byte(`;`), []byte(`;`), []byte(`;`), []byte(`;`), []byte(`FZK`), []byte(`29,00`), []byte(`5.99`), []byte(`500133`), []byte(`"8341170105289"`), []byte(`zqm`),
			// 	[]byte(`0.00`), []byte(`3-4 Udnk`), []byte(`Gvmih`), []byte(`Gvmih`), []byte(``), []byte(`"UPSYC"`), []byte(`Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg`), []byte(``), []byte(``),
			// 	[]byte(`he`), []byte(``), []byte(``), []byte(`he`), []byte(``), []byte(``), []byte(`he`), []byte(`0.32`), []byte(`90`), []byte(``),
			// },
			delimiter: characters.Character_SEMICOLON,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "531099",
			in: []byte(`lgvsyjskvoegf_fdcp,ixilgzh,jtbzwwfr,thez,eiyfykt,cytos,qitqfjbo_id_cbpdr,sihskify,lyypluebr,name,totil,gkn_nicn
Jzmsywoquf,"Dxozhqkehgr qjljub 60",ISGA,Ulimtq,KP,,9160925,46.43415,30.76626,"Iqxbm Zfejfmwhh Dxozhqkehgr qjljub 60",380-93-8074576,65000
Jzmsywoquf,"Otczvhkm 73",FJBJD,Arruxta,BG,,9160926,52.70491,4.69218,"Muzautlb Zfejfmwhh dc Arruxta, Xtvai Csfvssbgys\"jmt yon Fxjg rv fip Yzwgf",906092913421,1871DZ
Jzmsywoquf,"Iewybvyjbazoymxu 12",FJBJD,Xuntjaug,OT,,9160940,51.224306,8.605652,"Ulmbvd Zfejfmwhh dc Jügocrwzwg Rjatsciny vorc Fgm Zmor",49-906092913421,59964
Jzmsywoquf,"Bdvinjte 4",FJBJD,Kürazqoyup,OT,,9160944,50.149462,6.59977,"Feif Tvzrczspnwm Posny Vgei pwa Ngqjiykap sqe h Padngy",49-906092913421,54570`),
			multiOut: [][]byte{
				[]byte(`lgvsyjskvoegf_fdcp,ixilgzh,jtbzwwfr,thez,eiyfykt,cytos,qitqfjbo_id_cbpdr,sihskify,lyypluebr,name,totil,gkn_nicn`),
				[]byte(`Jzmsywoquf,"Dxozhqkehgr qjljub 60",ISGA,Ulimtq,KP,,9160925,46.43415,30.76626,"Iqxbm Zfejfmwhh Dxozhqkehgr qjljub 60",380-93-8074576,65000`),
				[]byte(`Jzmsywoquf,"Otczvhkm 73",FJBJD,Arruxta,BG,,9160926,52.70491,4.69218,"Muzautlb Zfejfmwhh dc Arruxta, Xtvai Csfvssbgys""jmt yon Fxjg rv fip Yzwgf",906092913421,1871DZ`),
				[]byte(`Jzmsywoquf,"Iewybvyjbazoymxu 12",FJBJD,Xuntjaug,OT,,9160940,51.224306,8.605652,"Ulmbvd Zfejfmwhh dc Jügocrwzwg Rjatsciny vorc Fgm Zmor",49-906092913421,59964`),
				[]byte(`Jzmsywoquf,"Bdvinjte 4",FJBJD,Kürazqoyup,OT,,9160944,50.149462,6.59977,"Feif Tvzrczspnwm Posny Vgei pwa Ngqjiykap sqe h Padngy",49-906092913421,54570`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     true,
		},
		{
			name: "512034",
			in: []byte(`ulxeoteu_id⁣renjy⁣jtbzwwfr_id⁣jtbzwwfr_name⁣hacbmomv_fdcp⁣gcebs_hacbmomv_mswb⁣hacbmomv_tyid⁣arck_qmlmsrps⁣irnagf⁣gegj⁣rddb_qwljy⁣xbv⁣wrjhr⁣rskxl_sdyyi⁣otpvgxum_nawum_name⁣yjuy_ewuwttdk⁣yjuy_ulxeoteu_ewuwttdk⁣ihibmrx_id⁣rlqdyjg_id⁣otpvgxum_id⁣rlqdyjg_id_ulxeoteu_id⁣jtbzwwfr_fxxz⁣jtbzwwfr_id_fxxz⁣fzxmi_plmladas_ihibmrx_id⁣kcsq⁣fvbllhimyVesy⁣gzg_uxuwzswmvic_ygdubn⁣xyurf_name⁣utphytufezmqjm
61⁣İsymrcş⁣2517⁣Erfzvqfjyq⁣YM⁣2022-08-09T13:01:26⁣86400⁣1⁣""⁣0341119754134⁣https://example.com/co464/lenilns/nnpma/bvjmzn/40446671/23/976749606/93319692/1/1_ajv_dtbj.knn⁣25647⁣"1\" Owşfoı Jkno Nyqzttkı (85)"⁣197.00⁣pmumdirfiviae⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?zynffovpVj=326493⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?nwcfgsfgGg=61&zynffovpVj=326493⁣6580485⁣vw83a02bl18dz001d7909lr27we9f7f4⁣326493⁣vw83a02bl18dz001d7909lr27we9f7f4_61⁣Yyşai > Lwrçq cq Ippı Ycwjuy > Ippı Ycwjuy > Svvjff Xi Tjmcybxs > Erfzvqfjyq⁣2859 > 2469 > 806 > 2502 > 2517⁣16⁣Gtw Ckcg⁣75 UY cq Üalhd Wvzyx Jpwecg⁣9⁣""⁣2022-08-08T14:35:50.998+03:00`),
			out: []byte(`ulxeoteu_id⁣renjy⁣jtbzwwfr_id⁣jtbzwwfr_name⁣hacbmomv_fdcp⁣gcebs_hacbmomv_mswb⁣hacbmomv_tyid⁣arck_qmlmsrps⁣irnagf⁣gegj⁣rddb_qwljy⁣xbv⁣wrjhr⁣rskxl_sdyyi⁣otpvgxum_nawum_name⁣yjuy_ewuwttdk⁣yjuy_ulxeoteu_ewuwttdk⁣ihibmrx_id⁣rlqdyjg_id⁣otpvgxum_id⁣rlqdyjg_id_ulxeoteu_id⁣jtbzwwfr_fxxz⁣jtbzwwfr_id_fxxz⁣fzxmi_plmladas_ihibmrx_id⁣kcsq⁣fvbllhimyVesy⁣gzg_uxuwzswmvic_ygdubn⁣xyurf_name⁣utphytufezmqjm`),
			multiOut: [][]byte{
				[]byte(`ulxeoteu_id⁣renjy⁣jtbzwwfr_id⁣jtbzwwfr_name⁣hacbmomv_fdcp⁣gcebs_hacbmomv_mswb⁣hacbmomv_tyid⁣arck_qmlmsrps⁣irnagf⁣gegj⁣rddb_qwljy⁣xbv⁣wrjhr⁣rskxl_sdyyi⁣otpvgxum_nawum_name⁣yjuy_ewuwttdk⁣yjuy_ulxeoteu_ewuwttdk⁣ihibmrx_id⁣rlqdyjg_id⁣otpvgxum_id⁣rlqdyjg_id_ulxeoteu_id⁣jtbzwwfr_fxxz⁣jtbzwwfr_id_fxxz⁣fzxmi_plmladas_ihibmrx_id⁣kcsq⁣fvbllhimyVesy⁣gzg_uxuwzswmvic_ygdubn⁣xyurf_name⁣utphytufezmqjm`),
				[]byte(`61⁣İsymrcş⁣2517⁣Erfzvqfjyq⁣YM⁣2022-08-09T13:01:26⁣86400⁣1⁣""⁣0341119754134⁣https://example.com/co464/lenilns/nnpma/bvjmzn/40446671/23/976749606/93319692/1/1_ajv_dtbj.knn⁣25647⁣"1\" Owşfoı Jkno Nyqzttkı (85)"⁣197.00⁣pmumdirfiviae⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?zynffovpVj=326493⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?nwcfgsfgGg=61&zynffovpVj=326493⁣6580485⁣vw83a02bl18dz001d7909lr27we9f7f4⁣326493⁣vw83a02bl18dz001d7909lr27we9f7f4_61⁣Yyşai > Lwrçq cq Ippı Ycwjuy > Ippı Ycwjuy > Svvjff Xi Tjmcybxs > Erfzvqfjyq⁣2859 > 2469 > 806 > 2502 > 2517⁣16⁣Gtw Ckcg⁣75 UY cq Üalhd Wvzyx Jpwecg⁣9⁣""⁣2022-08-08T14:35:50.998+03:00`),
			},
			delimiter: characters.Character_INVISIBLE_SEPARATOR,
			enclosure: characters.Character_QUOTATION_MARK,
			advance:   446,
			debug:     false,
		},
		{
			name: "1095223.53907798",
			in: []byte(`Renjy Name,Dmrs Ygdubn,Nawum Ibqypuvnyjx,Ixowhrp qwljy,Mliu Fdcp,Xczd Ibqypuvnyjx,Iloroshdrl,Rsrnqc,Vqcnyzi1 (Tjgosnky),Vqcnyzi2 (Tjgosnky),Vqcnyzi3 (Tjgosnky),Vqcnyzi4 (Tjgosnky),Vqcnyzi5,Vqcnyzi6,Vqcnyzi7,Vqcnyzi8,Vqcnyzi9,Vqcnyzi10,Vqcnyzi11,Vqcnyzi12,Vqcnyzi13,Vqcnyzi14,Vqcnyzi15,Vqcnyzi16,Vqcnyzi17,Vqcnyzi18,WsztfipqpAndvi1,WsztfipqpAndvi2,WsztfipqpAndvi3,WsztfipqpAndvi4,WsztfipqpAndvi5,WsztfipqpAndvi6,WsztfipqpAndvi7,WsztfipqpAndvi8,WsztfipqpAndvi9,WsztfipqpAndvi10,WghxlxwYvoi1,WghxlxwYvoi2,WghxlxwYvoi3,WghxlxwYvoi4,WghxlxwYvoi5,WghxlxwYvoi6,WghxlxwYvoi7,WghxlxwYvoi8,WghxlxwYvoi9,WghxlxwYvoi10,DodbipukZvqqq1,KbmjvgjjJyyy1,DodbipukZvqqq2,KbmjvgjjJyyy2,DodbipukZvqqq3,KbmjvgjjJyyy3,DodbipukZvqqq4,KbmjvgjjJyyy4,DodbipukZvqqq5,KbmjvgjjJyyy5
Tjsesqwt Nfpa,688PDOFN,PDOFN Cfqfwnxtq Moecfj Jgziy,688PDOFN.knn,Moecfj Jgziy,,"17.125D, 16.5I, 39.25E",29.5,hiosj merkt fgzoiaqr,,,,Xuygè guhlky,ujlnywnck xgazycldh,xgmwgbl fhxe,nqv zrnu,illi beck ohibojz,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Tjsesqwt Nfpa,688SXRAU,SXRAU Cfqfwnxtq Moecfj Jgziy,688SXRAU.knn,Moecfj Jgziy,,"17.125D, 16.5I, 39.25E",29.5,hiosj merkt fgzoiaqr,,,,Msfkth Rdhvcda Ikvfbl guhlky,ujlnywnck xgazycldh,xgmwgbl fhxe,nqv zrnu,illi beck ohibojz,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Tjsesqwt Nfpa,1100HNI,HNI Notui 3-Ilxjtd MkPmvgpd Wehdmrhwqx,1100HNI.knn,Wehdmrhwqx,,"18D, 16.5I, 29.5E",49,hiosj merkt fgzoiaqr,merkt hqreya,,,wzsxi xgazycldh,Xuygè guhlky,Jytl abiwzjykp dnjy kibhjzr czbwsc,Jfwngiu qnembico ilxjtd vnvrbvwhxwfy,ozkfpf lgus,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Tjsesqwt Nfpa,1100PQF,PQF Notui 3-Ilxjtd MkPmvgpd Wehdmrhwqx,1100PQF.knn,Wehdmrhwqx,,"18D, 16.5I, 29.5E",49,hiosj merkt fgzoiaqr,merkt hqreya,,,wzsxi xgazycldh,Vwqvoeoutp guhlky,Jytl abiwzjykp dnjy kibhjzr czbwsc,Jfwngiu qnembico ilxjtd vnvrbvwhxwfy,ozkfpf lgus,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			multiOut: [][]byte{
				[]byte(`Renjy Name,Dmrs Ygdubn,Nawum Ibqypuvnyjx,Ixowhrp qwljy,Mliu Fdcp,Xczd Ibqypuvnyjx,Iloroshdrl,Rsrnqc,Vqcnyzi1 (Tjgosnky),Vqcnyzi2 (Tjgosnky),Vqcnyzi3 (Tjgosnky),Vqcnyzi4 (Tjgosnky),Vqcnyzi5,Vqcnyzi6,Vqcnyzi7,Vqcnyzi8,Vqcnyzi9,Vqcnyzi10,Vqcnyzi11,Vqcnyzi12,Vqcnyzi13,Vqcnyzi14,Vqcnyzi15,Vqcnyzi16,Vqcnyzi17,Vqcnyzi18,WsztfipqpAndvi1,WsztfipqpAndvi2,WsztfipqpAndvi3,WsztfipqpAndvi4,WsztfipqpAndvi5,WsztfipqpAndvi6,WsztfipqpAndvi7,WsztfipqpAndvi8,WsztfipqpAndvi9,WsztfipqpAndvi10,WghxlxwYvoi1,WghxlxwYvoi2,WghxlxwYvoi3,WghxlxwYvoi4,WghxlxwYvoi5,WghxlxwYvoi6,WghxlxwYvoi7,WghxlxwYvoi8,WghxlxwYvoi9,WghxlxwYvoi10,DodbipukZvqqq1,KbmjvgjjJyyy1,DodbipukZvqqq2,KbmjvgjjJyyy2,DodbipukZvqqq3,KbmjvgjjJyyy3,DodbipukZvqqq4,KbmjvgjjJyyy4,DodbipukZvqqq5,KbmjvgjjJyyy5`),
				[]byte(`Tjsesqwt Nfpa,688PDOFN,PDOFN Cfqfwnxtq Moecfj Jgziy,688PDOFN.knn,Moecfj Jgziy,,"17.125D, 16.5I, 39.25E",29.5,hiosj merkt fgzoiaqr,,,,Xuygè guhlky,ujlnywnck xgazycldh,xgmwgbl fhxe,nqv zrnu,illi beck ohibojz,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
				[]byte(`Tjsesqwt Nfpa,688SXRAU,SXRAU Cfqfwnxtq Moecfj Jgziy,688SXRAU.knn,Moecfj Jgziy,,"17.125D, 16.5I, 39.25E",29.5,hiosj merkt fgzoiaqr,,,,Msfkth Rdhvcda Ikvfbl guhlky,ujlnywnck xgazycldh,xgmwgbl fhxe,nqv zrnu,illi beck ohibojz,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
				[]byte(`Tjsesqwt Nfpa,1100HNI,HNI Notui 3-Ilxjtd MkPmvgpd Wehdmrhwqx,1100HNI.knn,Wehdmrhwqx,,"18D, 16.5I, 29.5E",49,hiosj merkt fgzoiaqr,merkt hqreya,,,wzsxi xgazycldh,Xuygè guhlky,Jytl abiwzjykp dnjy kibhjzr czbwsc,Jfwngiu qnembico ilxjtd vnvrbvwhxwfy,ozkfpf lgus,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
				[]byte(`Tjsesqwt Nfpa,1100PQF,PQF Notui 3-Ilxjtd MkPmvgpd Wehdmrhwqx,1100PQF.knn,Wehdmrhwqx,,"18D, 16.5I, 29.5E",49,hiosj merkt fgzoiaqr,merkt hqreya,,,wzsxi xgazycldh,Vwqvoeoutp guhlky,Jytl abiwzjykp dnjy kibhjzr czbwsc,Jfwngiu qnembico ilxjtd vnvrbvwhxwfy,ozkfpf lgus,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "1094837.86062382",
			in: []byte(`id,gqajobozq
251107426971116441,"IRYWPNKX 94.99 tsdbf=Fot 19, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
402942271602461963,"IRYWPNKX 29.99 tsdbf=Tbv 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
001719652615786438,"IRYWPNKX 59.99 tsdbf=Tbv 15, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
594897324510009643,"IRYWPNKX 29.99 tsdbf=Tbv 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
522619040541706136,"IRYWPNKX 59.99 tsdbf=Tbv 15, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
966965409375111804,"IRYWPNKX 52.99 tsdbf=Kir 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
413339799859313066,"IRYWPNKX 29.0 tsdbf=Fot 22, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
149728023565158990,"IRYWPNKX 5.49 tsdbf=Ylz 13, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"
976791380954578382,"IRYWPNKX 18.99 tsdbf=Kir 18, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
			multiOut: [][]byte{
				[]byte(`id,gqajobozq`),
				[]byte(`251107426971116441,"IRYWPNKX 94.99 tsdbf=Fot 19, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`402942271602461963,"IRYWPNKX 29.99 tsdbf=Tbv 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`001719652615786438,"IRYWPNKX 59.99 tsdbf=Tbv 15, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`594897324510009643,"IRYWPNKX 29.99 tsdbf=Tbv 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`522619040541706136,"IRYWPNKX 59.99 tsdbf=Tbv 15, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`966965409375111804,"IRYWPNKX 52.99 tsdbf=Kir 16, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`413339799859313066,"IRYWPNKX 29.0 tsdbf=Fot 22, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`149728023565158990,"IRYWPNKX 5.49 tsdbf=Ylz 13, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
				[]byte(`976791380954578382,"IRYWPNKX 18.99 tsdbf=Kir 18, 2022, 12:00:00 DH,ufd=Tfy 31, 9999, 11:59:59 KI"`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "1103578 17831668",
			in: []byte(`Bw.,Ibqypuvnyjx,Avpg Psld th Bysrygj,Llabpeu,Frgxutoli Nicn,Hewuisp Muwsbc,Eiyfykt/Mgorlj th Ollgfp Nicn,CDM Xcxcydq,Kivcg Rsrnqc,Mliu Jtbzwwfr Nicn,Dybvnjx Xjgtx Plmladas,Ihe Rsrnqc,Lenilns Uczoj Nicn,Mzdqni Psld Xcxcydq,Xjgtx Psld th Bysrygj,Xrfv Dc Qvtj Uwuh Uuq,Nzjksb,Qrlixeuhz Mzdqni Sdyyi,Tiwwsv Psld Xcxcydq,NDY Josj. Rvsgbfv Uczoj,NS,btyliv,xitxq,lkuocy,rsrnqc,IFQWG,Ifqwg btyliv,Ifqwg xitxq,Ifqwg lkuocy,Ifqwg rsrnqc,TEEWN,Teewn btyliv,Teewn xitxq,Teewn lkuocy,Teewn rsrnqc,KJEAHM,Kjeahm btyliv,Kjeahm xitxq,Kjeahm lkuocy,Kjeahm rsrnqc,Rskxl Qsylb Tgzn,Fuhsaho
11255,AOAVLKQLN YPPYPMQUTEYR,NS,yslts,4549050264,GSCH,KU,,880,LWOWG,3,0.874,7030,3870212527142,NS,yslts,Cnid,12.99,0060193450567,VRZ20 NDY,1,272,64,367,775,3,288,204,375,2640,3,288,204,375,2640,240,0,0,0,0,"‘Aoavlkqln Yppypmquteyr’ nw fip wrjhr th kuba ywsumspnf 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Durajhe nmdtkd yihrp. Fip muzautlb qwljy gqkfdtebicv h qzvw yppypmquteyr aiphku vgei ywsumspnf gumdqkrkqf th kvdfugwugzh sqe yegv uvan pclngod. Kuba nmdtkd nw ugzzy rsm kkqdsa oodg visrz pclngod sqe ddvtlv. Kuba 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip feif xocgjfwn tspvto – Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1000 cvgsb wxjavq nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz sqe gsqnsni h qljcgd-vqwk guhlky rv ow nzappbx. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030
11264,BASLQUB DGMVZOZPJ,NS,yslts,4549050264,GSCH,KU,,967,LWOWG,3,0.79,7030,3443726479597,NS,yslts,Cnid,12.99,0696827297271,VRZ20 NDY,1,270,70,370,790,3,0,0,0,0,3,282,210,374,2900,240,0,0,0,0,"‘Baslqub Dgmvzozpj’ nw fip wrjhr th kuba qhygfzszjm 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Lxhtkygtg nmdtkd yihrp. Fip nmdtkd qwljy fjxjvlzx h ywsumspnf aalzyal exb ji fip baslqub, vgei th bdnbfw fip ebuoai hepebcicdh; h febp sqe ixwem hlso sqe ste rwyeq qng rsm dbhjnndgi hdfnoq hhlvkl. 

Kuba shorhiayybr 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip xocgjfwn tspvto Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1000 cvgsb nmdtkd iffjtjic 68 l49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm h hfxv spqhpzp dsrzqfulw rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. 

Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030
11267,BS QSUZMFMTL DC FIP QBJEGUSI,NS,yslts,4549050264,GSCH,KU,,880,LWOWG,3,0.79,7030,5662961477452,NS,yiit,Cnid,12.99,7318871853569,VRZ20 NDY,1,272,64,367,770,3,0,0,0,0,3,288,204,375,3010,240,0,0,0,0,"Bs Qsuzmfmtl nx fip Qbjegusi' nw fip wrjhr th kuba qhygfzszjm 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Lxhtkygtg nmdtkd yihrp. Fip nmdtkd qwljy fjxjvlzx h ltgzmkquwbv sqe grxnymjad qbjegusi oodg ujeyb ow h vsrtu gshm yiit rsm wgu wlsygncqxu rv tekey sqe dgzs uxht vyqzlprhz fpieg dc. Kuba shorhiayybr 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip feif xocgjfwn tspvto – Bfthfb Lxeylzvnym. Fip gzygsjno kcsq th fip 1000 cvgsb nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030
11275,FIP ENIDWHPUB YPPYPMQUTEYR,NS,yslts,4549050264,GSCH,KU,,907,LWOWG,3,0.874,7030,4291968677440,NS,yslts,Cnid,12.99,5647242493228,VRZ20 NDY,1,272,64,367,800,3,288,201,382,2720,3,288,201,382,2720,150,0,0,0,0,"‘Enidwhpub Yppypmquteyr’ nw fip wrjhr th kuba qhygfzszjm 1,000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Durajhe nmdtkd yihrp. Fip nmdtkd qwljy gqkfdtebicv h ywsumspnf Enidwhpub sznktgklm yppypmquteyr. Fip yppypmquteyr malvjbfc h feif vvjs sqe hdumpgr ltqrmdm vgei fip qtxh, pixh sqe hmbneyv sdnibhw krwnex fip croy. Kuba shorhiayybr 1,000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip xocgjfwn tspvto Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1,000 cvgsb nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030`),
			multiOut: [][]byte{
				[]byte(`Bw.,Ibqypuvnyjx,Avpg Psld th Bysrygj,Llabpeu,Frgxutoli Nicn,Hewuisp Muwsbc,Eiyfykt/Mgorlj th Ollgfp Nicn,CDM Xcxcydq,Kivcg Rsrnqc,Mliu Jtbzwwfr Nicn,Dybvnjx Xjgtx Plmladas,Ihe Rsrnqc,Lenilns Uczoj Nicn,Mzdqni Psld Xcxcydq,Xjgtx Psld th Bysrygj,Xrfv Dc Qvtj Uwuh Uuq,Nzjksb,Qrlixeuhz Mzdqni Sdyyi,Tiwwsv Psld Xcxcydq,NDY Josj. Rvsgbfv Uczoj,NS,btyliv,xitxq,lkuocy,rsrnqc,IFQWG,Ifqwg btyliv,Ifqwg xitxq,Ifqwg lkuocy,Ifqwg rsrnqc,TEEWN,Teewn btyliv,Teewn xitxq,Teewn lkuocy,Teewn rsrnqc,KJEAHM,Kjeahm btyliv,Kjeahm xitxq,Kjeahm lkuocy,Kjeahm rsrnqc,Rskxl Qsylb Tgzn,Fuhsaho`),
				[]byte(`11255,AOAVLKQLN YPPYPMQUTEYR,NS,yslts,4549050264,GSCH,KU,,880,LWOWG,3,0.874,7030,3870212527142,NS,yslts,Cnid,12.99,0060193450567,VRZ20 NDY,1,272,64,367,775,3,288,204,375,2640,3,288,204,375,2640,240,0,0,0,0,"‘Aoavlkqln Yppypmquteyr’ nw fip wrjhr th kuba ywsumspnf 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Durajhe nmdtkd yihrp. Fip muzautlb qwljy gqkfdtebicv h qzvw yppypmquteyr aiphku vgei ywsumspnf gumdqkrkqf th kvdfugwugzh sqe yegv uvan pclngod. Kuba nmdtkd nw ugzzy rsm kkqdsa oodg visrz pclngod sqe ddvtlv. Kuba 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip feif xocgjfwn tspvto – Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1000 cvgsb wxjavq nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz sqe gsqnsni h qljcgd-vqwk guhlky rv ow nzappbx. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030`),
				[]byte(`11264,BASLQUB DGMVZOZPJ,NS,yslts,4549050264,GSCH,KU,,967,LWOWG,3,0.79,7030,3443726479597,NS,yslts,Cnid,12.99,0696827297271,VRZ20 NDY,1,270,70,370,790,3,0,0,0,0,3,282,210,374,2900,240,0,0,0,0,"‘Baslqub Dgmvzozpj’ nw fip wrjhr th kuba qhygfzszjm 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Lxhtkygtg nmdtkd yihrp. Fip nmdtkd qwljy fjxjvlzx h ywsumspnf aalzyal exb ji fip baslqub, vgei th bdnbfw fip ebuoai hepebcicdh; h febp sqe ixwem hlso sqe ste rwyeq qng rsm dbhjnndgi hdfnoq hhlvkl. 

Kuba shorhiayybr 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip xocgjfwn tspvto Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1000 cvgsb nmdtkd iffjtjic 68 l49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm h hfxv spqhpzp dsrzqfulw rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. 

Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030`),
				[]byte(`11267,BS QSUZMFMTL DC FIP QBJEGUSI,NS,yslts,4549050264,GSCH,KU,,880,LWOWG,3,0.79,7030,5662961477452,NS,yiit,Cnid,12.99,7318871853569,VRZ20 NDY,1,272,64,367,770,3,0,0,0,0,3,288,204,375,3010,240,0,0,0,0,"Bs Qsuzmfmtl nx fip Qbjegusi' nw fip wrjhr th kuba qhygfzszjm 1000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Lxhtkygtg nmdtkd yihrp. Fip nmdtkd qwljy fjxjvlzx h ltgzmkquwbv sqe grxnymjad qbjegusi oodg ujeyb ow h vsrtu gshm yiit rsm wgu wlsygncqxu rv tekey sqe dgzs uxht vyqzlprhz fpieg dc. Kuba shorhiayybr 1000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip feif xocgjfwn tspvto – Bfthfb Lxeylzvnym. Fip gzygsjno kcsq th fip 1000 cvgsb nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030`),
				[]byte(`11275,FIP ENIDWHPUB YPPYPMQUTEYR,NS,yslts,4549050264,GSCH,KU,,907,LWOWG,3,0.874,7030,4291968677440,NS,yslts,Cnid,12.99,5647242493228,VRZ20 NDY,1,272,64,367,800,3,288,201,382,2720,3,288,201,382,2720,150,0,0,0,0,"‘Enidwhpub Yppypmquteyr’ nw fip wrjhr th kuba qhygfzszjm 1,000 cvgsb Snkipv ot cqeo wxjavq nmdtkd, rvtuq nw dmrs th fip Durajhe nmdtkd yihrp. Fip nmdtkd qwljy gqkfdtebicv h ywsumspnf Enidwhpub sznktgklm yppypmquteyr. Fip yppypmquteyr malvjbfc h feif vvjs sqe hdumpgr ltqrmdm vgei fip qtxh, pixh sqe hmbneyv sdnibhw krwnex fip croy. Kuba shorhiayybr 1,000 cvgsb wxjavq nmdtkd gii tgih zogryhgttxf ji fip xocgjfwn tspvto Ctbtoq Lxkw. Fip gzygsjno kcsq th fip 1,000 cvgsb nmdtkd iffjtjic 68 l 49he (qhrbpioffsrhe) sqe gii tgih lsduhhoi snojm hfxv spqhpzp mrugcxho iehra rv kpfilt nebeh wxjavq cvgsb bdqfomxnh fig skfyphkt tbiiz tyid sqe tyid ptzau. Xanzw’s Snkipv ot cqeo nimpqsz bhv qbggztku sqe rowxxrd vgei fip iluk qkmwolu rsm cbvpboinji oodg ac ogxq ghwtdf cnq. Taea xoqboxqua Snkipv’s bzeumbhzegt renjy mltfgn th iyvjkrvn sqe ubsguujqyj. Ac rzstku akzu fip gqld lujhlqfp nnrzqiwgvtxwx rsm ont yqaepvr. Ac hwu-cqn hvvi wxjavq cvgsb oxukbentg, ra eery fip wxjavq nw fspqdzzmk, as getnj uvan vqwk h qljcgd kbcg h nmdtkd. Ckyrah yqaepvr gii ghwtdf tgih, sqe kusy ghwtdf afqlab, ont qkmwolu.",7030`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
		},
		{
			name: "csvSample.447880.48889040",
			in: []byte(`gziuvsx_hlitxc_dae,dae_vfdpux_gziuvsx,gziuvsx_wic,gziuvsx_yiOoxumimne,gziuvsx_mosbfPhgj,gziuvsx_xfmWhrsjbtdxdFzc,gziuvsx_txekMapZirt,gziuvsx_bnxca,lenilns_vfdpux_umliumzl,lenilns_id,lenilns_nzjksb,nw_mlyllbrbw,gziuvsx_dae_name,gziuvsx_sktcpntIacj,gziuvsx_stcfidhPc,gziuvsx_jkfspiLsbpktoFe,gziuvsx_dmlPlemxzbmhlGxmxo,gziuvsx_uzlefbwxSwynncxiKhcbv,gziuvsx_gcbeJkntb,gziuvsx_olcmVotqvcmv,gziuvsx_cqeuSuwxritcfeg,gziuvsx_xotrtmqyLijunmsqJp,gziuvsx_kyefKrsrchfvaqg,gziuvsx_zwxhjobpUgbnjgbi,gziuvsx_qodinvVsheeeycy,gziuvsx_nqlbxjpAm,gziuvsx_ylonvqVvtpm,imga_tjgosnky_ajtka,imga_rbqbmnloqyv,imga_nw_srbd,imga_tjgosnky,imga_nw_twvlfpdelrg,imga_dktu,imga_tjgosnky_jlcqjfvqex,imga_ivks,imga_qapdmph,imga_cfus,imga_bwmuc,imga_nw_zsuyp_vrjvcpeg,imga_ktrhv,imga_yqvids,imga_ivks_apkggm_rnszufehk,imga_fofabsgsup_ortkti,imga_xyurf_ajtka,imga_cfus_ot_avpg,dae_vfdpux_fdcp,dae_vfdpux_mosbfphgj,dae_vfdpux_renjy_id,dae_vfdpux_renjy_name,dae_vfdpux_jtbzwwfr_id,dae_vfdpux_name,dae_vfdpux_jtbzwwfr_dzprb,dae_vfdpux_jtbzwwfr_uarpc,dae_vfdpux_ztpcncurd_dzprb_0,dae_vfdpux_ztpcncurd_dzprb_1,dae_vfdpux_ztpcncurd_dzprb_2,dae_vfdpux_ztpcncurd_dzprb_3,dae_vfdpux_jtbzwwfr_dzprb_srbd,dae_vfdpux_jtbzwwfr_dzprb_srbd_0,dae_vfdpux_jtbzwwfr_dzprb_srbd_1,dae_vfdpux_jtbzwwfr_dzprb_srbd_2,lenilns_stcfidhPc,lenilns_jkfspiLsbpktoFe,lenilns_bnxca,lenilns_uzlefbwxSwynncxiKhcbv,lenilns_gcbeJkntb,lenilns_olcmVotqvcmv,lenilns_cqeuSuwxritcfeg,lenilns_xotrtmqyLijunmsqJp,lenilns_kyefKrsrchfvaqg,lenilns_cewkbvFrqdowlp,lenilns_qodinvVsheeeycy,lenilns_xqnXksv,lenilns_nqlbxjpAm,urc_sdyyi,tssfu_urc_sdyyi,rykgrnfs_urc_sdyyi,sdyyi_h_u,tssfu_h_u_sdyyi,rykgrnfs_h_u_sdyyi,sdyyi_xysuduxe,tssfu_xysuduxe_sdyyi,rykgrnfs_xysuduxe_sdyyi,sdyyi_nxjfjf,rykgrnfs_nxjfjf_sdyyi,tssfu_nxjfjf_sdyyi,sdyyi_xcfje,sdyyi_facyp,jyrgdhtbimuk_orxbolbpcKovxn,jyrgdhtbimuk_thyqprgsjwgnQiwoh,jyrgdhtbimuk_xqjilyslhufzcNbyds,jyrgdhtbimuk_iygbkanvpgMrbbb,jyrgdhtbimuk_fduihqdGa,jyrgdhtbimuk_czwgllwwbgerrPrpoahceyi,jyrgdhtbimuk_bvdyasgwUe,jyrgdhtbimuk_cdctilgwysqxLgsmuxiglq,jyrgdhtbimuk_wskgxpjaxJesubeeowm,jyrgdhtbimuk_fhmpgainn,lenilns_xznguSedbctMfxnu,ejmwmttpVx,lenilns_jqitpbkxc_nw_ewtud_lenilns,vloribc_yrpmtupgx_xhc_f3,gziuvsx_jqitpbkxc_ixowhrp_xyurf_uczoj,vloribc_yrpmtupgx_uzn_ynuvxz,vloribc_yrpmtupgx_xhc_f1,vloribc_yrpmtupgx_xhc_f2,vloribc_yrpmtupgx_gfu_ynuvxz,vloribc_yrpmtupgx_gyhy_wiz_bybna,pgbyn_lenilns_nzeogzvXdhjdiaWizxkh,vloribc_yrpmtupgx_xzhqba_ynuvxz,lenilns_nzeogzvXdhjdiaWizxkh,qfypq_id,qfypq_dae,qfypq_seclcwjbahg_njh,qfypq_vfdpux_id,qfypq_qfypq_ybumis,qfypq_qfypq_id_ybumis,qfypq_id_gbepg,qfypq_ybumis_gbepg,qfypq_id_arcvepv,qfypq_ybumis_arcvepv,qfypq_id_kbgxtaig,qfypq_ybumis_kbgxtaig,qfypq_cuem,qfypq_dnpdy,qfypq_xkfjgdff_dnpdy_hlavvq,qfypq_nzjksb_njh,qfypq_xkfjgdff_izzvjs,qfypq_xkfjgdff_izzvjs_17,qfypq_tgfpt,qfypq_mpeumehdyc,qfypq_arcvepv,sdyyi_nxjfjf_2,sdyyi_xysuduxe_2,sdyyi_urc_2,sdyyi_h_u_2,sdyyi_nxjfjf_10,sdyyi_xysuduxe_10,sdyyi_urc_10,sdyyi_h_u_10,sdyyi_nxjfjf_7021,sdyyi_xysuduxe_7021,sdyyi_urc_7021,sdyyi_h_u_7021,uvpjnoql_eiyfykt,uvpjnoql_mswb_oxafadxeo,uvpjnoql_mswb_facyp_szvsrncrlr,zbuf_mswb,sdyyi_nxjfjf_5021,sdyyi_xysuduxe_5021,sdyyi_urc_5021,sdyyi_h_u_5021,oeupzZgankpwdh
4108981,4108981,49036864089,yiit,,[],M/V,[],[],4108981,RFMZDY,yiit,"Rkhwrqlhr 2112115 1,5F Yjjjy",Rypzpnj,,,,,,,,,,,,,,,,,Bláibyvx,,35 he,,,,,13 he,,19 he,2112115,,,,,Rypzpnj,,ydmsf,Ydmsf,,"Rkhwrqlhr 2112115 1,5F Yjjjy","[""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx""]","[""Tsioafnirx""]","[""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx"",""xhc949006||Zxlfsv"",""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf""]","[""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv"",""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv""]","[""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx""]",[],,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 249,
  ""sdyyi_nxjfjf_12"": 249,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 249,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 249,
  ""sdyyi_nxjfjf_24"": 249,
  ""sdyyi_nxjfjf_27"": 249,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 249,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 249,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",249,0,,,,,,,,,,,,,,4108981,yslts,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,249,,,,,,,,,,,,Badu,2022-08-01,2022-08-01T16:01:51.376N,2022-10-04,,,,,
4109511,4109511,5365204762414,yiit,,[],M/V,[],[],4109511,RFMZDY,yiit,"Nvgdtfgdk Awfheip 4-1/2"" 700D SUM700 + Fgszcp",Rypzpnj,,,,,,,,,,,,,,,,,,,,,,,Zphassunwd cihx tggzrizlodjg nyélecvkmn,,,,"SUM 700 + Bxxtk ot Jcvuvekj Qhwfne cihx Myctp 115x6,0dv Tfvykl Ffcmgzsoz + Bxxtk \Pfpfalcfkgaz cihx Fokuerói Hguu 4-1/2\",,,,,Rypzpnj,,gbzbl,Gbzbl,,"Nvgdtfgdk Awfheip 4-1/2"" 700D SUM700 + Fgszcp","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc549071||Dpnmwzlz/xhc2349005||Gbzbl Tggzrizlodjg""]","[""Nvgdtfgdk t Muoeywnbw"",""Gbzbl Tggzrizlodjg"",""Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf""]","[""xhc549071||Dpnmwzlz"",""xhc1359004||Qhbjqm"",""xhc539003||Tggzrizlodjg t Nelabhberab""]","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc549071||Dpnmwzlz/xhc2349005||Gbzbl Tggzrizlodjg""]","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw""]",[],,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 219.9,
  ""sdyyi_nxjfjf_12"": 219.9,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 219.9,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 219.9,
  ""sdyyi_nxjfjf_24"": 219.9,
  ""sdyyi_nxjfjf_27"": 219.9,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 219.9,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 219,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",219.9,0,,,"[""40502"",""40501"",""40506"",""150718"",""40505"",""150719"",""40504"",""150716"",""40503"",""150717"",""150714"",""150715"",""150712"",""150713"",""151009"",""150710"",""151008"",""150711"",""151007"",""151006"",""151005"",""151004"",""151003"",""151002"",""151001"",""30029"",""150608"",""150729"",""150609"",""150606"",""150727"",""150607"",""150728"",""150604"",""150725"",""40608"",""150605"",""150726"",""40607"",""150602"",""150723"",""40606"",""150603"",""150724"",""150721"",""151019"",""150601"",""150722"",""151018"",""151017"",""150720"",""151016"",""151015"",""151014"",""151013"",""151012"",""151011"",""151010"",""40601"",""40605"",""40604"",""40603"",""40602"",""150611"",""150732"",""150612"",""151029"",""150730"",""151028"",""150610"",""150731"",""151027"",""151026"",""151025"",""151024"",""151023"",""151022"",""151020"",""150509"",""150507"",""150508"",""150505"",""150506"",""40706"",""150503"",""40705"",""150504"",""150501"",""150502"",""151033"",""151032"",""151031"",""151030"",""40704"",""40703"",""40702"",""150118"",""150114"",""150112"",""150110"",""150111"",""40810"",""150808"",""150809"",""150806"",""150807"",""150804"",""40811"",""150805"",""40807"",""150802"",""40806"",""150803"",""40805"",""150129"",""40804"",""150801"",""150127"",""150128"",""40809"",""40808"",""150126"",""150124"",""40120"",""40124"",""40123"",""40122"",""40121"",""40128"",""40127"",""40126"",""40125"",""40803"",""40802"",""40801"",""40129"",""150811"",""150812"",""150138"",""150810"",""150137"",""150132"",""40113"",""40112"",""40111"",""40117"",""40116"",""40115"",""150709"",""40114"",""150707"",""150708"",""40119"",""150705"",""40118"",""150706"",""150703"",""150704"",""150701"",""150702"",""40102"",""40101"",""40106"",""40105"",""40104"",""40103"",""40109"",""40108"",""40107"",""150204"",""150205"",""150202"",""150203"",""150201"",""40201"",""40205"",""40204"",""40203"",""40202"",""40208"",""40207"",""40206"",""40311"",""40310"",""40313"",""40312"",""150906"",""150905"",""150904"",""150903"",""150109"",""150902"",""150901"",""150107"",""40309"",""150103"",""40304"",""40303"",""40302"",""40301"",""40308"",""40307"",""40306"",""40305"",""150516"",""150515"",""150512"",""150513"",""150510"",""150511"",""40410"",""40414"",""40413"",""40412"",""40411"",""150406"",""150407"",""40409"",""150404"",""40408"",""150405"",""150402"",""150403"",""150401"",""40520"",""40403"",""40402"",""40401"",""40407"",""40406"",""40405"",""40404"",""40519"",""40518"",""40513"",""40512"",""40511"",""40510"",""40517"",""40516"",""40515"",""40514"",""40509"",""40508"",""150305"",""40507"",""150303"",""150304"",""150301"",""150302""]",[],"[""30029"",""40101"",""40102"",""40103"",""40104"",""40105"",""40106"",""40107"",""40108"",""40109"",""40111"",""40112"",""40113"",""40114"",""40115"",""40116"",""40117"",""40118"",""40119"",""40120"",""40121"",""40122"",""40123"",""40124"",""40125"",""40126"",""40127"",""40128"",""40129"",""40201"",""40202"",""40203"",""40204"",""40205"",""40206"",""40207"",""40208"",""40301"",""40302"",""40303"",""40304"",""40305"",""40306"",""40307"",""40308"",""40309"",""40310"",""40311"",""40312"",""40313"",""40401"",""40402"",""40403"",""40404"",""40405"",""40406"",""40407"",""40408"",""40409"",""40410"",""40411"",""40412"",""40413"",""40414"",""40501"",""40502"",""40503"",""40504"",""40505"",""40506"",""40507"",""40508"",""40509"",""40510"",""40511"",""40512"",""40513"",""40514"",""40515"",""40516"",""40517"",""40518"",""40519"",""40520"",""40601"",""40602"",""40603"",""40604"",""40605"",""40606"",""40607"",""40608"",""40702"",""40703"",""40704"",""40705"",""40706"",""40801"",""40802"",""40803"",""40804"",""40805"",""40806"",""40807"",""40808"",""40809"",""40810"",""40811"",""150103"",""150107"",""150109"",""150110"",""150111"",""150112"",""150114"",""150118"",""150124"",""150126"",""150127"",""150128"",""150129"",""150132"",""150137"",""150138"",""150201"",""150202"",""150203"",""150204"",""150205"",""150301"",""150302"",""150303"",""150304"",""150305"",""150401"",""150402"",""150403"",""150404"",""150405"",""150406"",""150407"",""150501"",""150502"",""150503"",""150504"",""150505"",""150506"",""150507"",""150508"",""150509"",""150510"",""150511"",""150512"",""150513"",""150515"",""150516"",""150601"",""150602"",""150603"",""150604"",""150605"",""150606"",""150607"",""150608"",""150609"",""150610"",""150611"",""150612"",""150701"",""150702"",""150703"",""150704"",""150705"",""150706"",""150707"",""150708"",""150709"",""150710"",""150711"",""150712"",""150713"",""150714"",""150715"",""150716"",""150717"",""150718"",""150719"",""150720"",""150721"",""150722"",""150723"",""150724"",""150725"",""150726"",""150727"",""150728"",""150729"",""150730"",""150731"",""150732"",""150801"",""150802"",""150803"",""150804"",""150805"",""150806"",""150807"",""150808"",""150809"",""150810"",""150811"",""150812"",""150901"",""150902"",""150903"",""150904"",""150905"",""150906"",""151001"",""151002"",""151003"",""151004"",""151005"",""151006"",""151007"",""151008"",""151009"",""151010"",""151011"",""151012"",""151013"",""151014"",""151015"",""151016"",""151017"",""151018"",""151019"",""151020"",""151022"",""151023"",""151024"",""151025"",""151026"",""151027"",""151028"",""151029"",""151030"",""151031"",""151032"",""151033""]","[""150103"",""150107"",""150109"",""150110"",""150111"",""150112"",""150114"",""150118"",""150132"",""150137""]",,,,,"",yiit,,4109511,yslts,95540100,,"",0209,020902,,0,,"",,,,,,,,,,,,,,,,,,,,,,,219,,,,,,,,,,,,Badu,2022-10-01,2022-10-01T06:53:33.153N,2022-10-04,,,,,
4124456,4124456,0669556787040,yiit,,[],M/V,[],[],4124456,RFMZDY,yiit,Lysyotipo ag Ikzseo ot Qfjclr,Rypzpnj,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,Rypzpnj,,gyrcbwpn,Gyrcbwpn,,Lysyotipo ag Ikzseo ot Qfjclr,"[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp""]","[""Ureokkv Eekwiqfkp""]","[""xhc2909008||Orlouwf t Grbokflrivór"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr""]","[""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq""]","[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc""]","[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp""]",,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 59.9,
  ""sdyyi_nxjfjf_12"": 59.9,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 59.9,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 59.9,
  ""sdyyi_nxjfjf_24"": 59.9,
  ""sdyyi_nxjfjf_27"": 59.9,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 59.9,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 59,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",59.9,0,,,,,,,,,,,,,,4124456,yslts,18616441,,,0419,041905,,0,,,,,,,,,,,,,,,,,,,,,,,,,59,,,,,,,,,,,,Badu,2022-06-29,2022-06-29T04:35:29.043N,2022-10-04,,,,,`),
			multiOut: [][]byte{
				[]byte(`gziuvsx_hlitxc_dae,dae_vfdpux_gziuvsx,gziuvsx_wic,gziuvsx_yiOoxumimne,gziuvsx_mosbfPhgj,gziuvsx_xfmWhrsjbtdxdFzc,gziuvsx_txekMapZirt,gziuvsx_bnxca,lenilns_vfdpux_umliumzl,lenilns_id,lenilns_nzjksb,nw_mlyllbrbw,gziuvsx_dae_name,gziuvsx_sktcpntIacj,gziuvsx_stcfidhPc,gziuvsx_jkfspiLsbpktoFe,gziuvsx_dmlPlemxzbmhlGxmxo,gziuvsx_uzlefbwxSwynncxiKhcbv,gziuvsx_gcbeJkntb,gziuvsx_olcmVotqvcmv,gziuvsx_cqeuSuwxritcfeg,gziuvsx_xotrtmqyLijunmsqJp,gziuvsx_kyefKrsrchfvaqg,gziuvsx_zwxhjobpUgbnjgbi,gziuvsx_qodinvVsheeeycy,gziuvsx_nqlbxjpAm,gziuvsx_ylonvqVvtpm,imga_tjgosnky_ajtka,imga_rbqbmnloqyv,imga_nw_srbd,imga_tjgosnky,imga_nw_twvlfpdelrg,imga_dktu,imga_tjgosnky_jlcqjfvqex,imga_ivks,imga_qapdmph,imga_cfus,imga_bwmuc,imga_nw_zsuyp_vrjvcpeg,imga_ktrhv,imga_yqvids,imga_ivks_apkggm_rnszufehk,imga_fofabsgsup_ortkti,imga_xyurf_ajtka,imga_cfus_ot_avpg,dae_vfdpux_fdcp,dae_vfdpux_mosbfphgj,dae_vfdpux_renjy_id,dae_vfdpux_renjy_name,dae_vfdpux_jtbzwwfr_id,dae_vfdpux_name,dae_vfdpux_jtbzwwfr_dzprb,dae_vfdpux_jtbzwwfr_uarpc,dae_vfdpux_ztpcncurd_dzprb_0,dae_vfdpux_ztpcncurd_dzprb_1,dae_vfdpux_ztpcncurd_dzprb_2,dae_vfdpux_ztpcncurd_dzprb_3,dae_vfdpux_jtbzwwfr_dzprb_srbd,dae_vfdpux_jtbzwwfr_dzprb_srbd_0,dae_vfdpux_jtbzwwfr_dzprb_srbd_1,dae_vfdpux_jtbzwwfr_dzprb_srbd_2,lenilns_stcfidhPc,lenilns_jkfspiLsbpktoFe,lenilns_bnxca,lenilns_uzlefbwxSwynncxiKhcbv,lenilns_gcbeJkntb,lenilns_olcmVotqvcmv,lenilns_cqeuSuwxritcfeg,lenilns_xotrtmqyLijunmsqJp,lenilns_kyefKrsrchfvaqg,lenilns_cewkbvFrqdowlp,lenilns_qodinvVsheeeycy,lenilns_xqnXksv,lenilns_nqlbxjpAm,urc_sdyyi,tssfu_urc_sdyyi,rykgrnfs_urc_sdyyi,sdyyi_h_u,tssfu_h_u_sdyyi,rykgrnfs_h_u_sdyyi,sdyyi_xysuduxe,tssfu_xysuduxe_sdyyi,rykgrnfs_xysuduxe_sdyyi,sdyyi_nxjfjf,rykgrnfs_nxjfjf_sdyyi,tssfu_nxjfjf_sdyyi,sdyyi_xcfje,sdyyi_facyp,jyrgdhtbimuk_orxbolbpcKovxn,jyrgdhtbimuk_thyqprgsjwgnQiwoh,jyrgdhtbimuk_xqjilyslhufzcNbyds,jyrgdhtbimuk_iygbkanvpgMrbbb,jyrgdhtbimuk_fduihqdGa,jyrgdhtbimuk_czwgllwwbgerrPrpoahceyi,jyrgdhtbimuk_bvdyasgwUe,jyrgdhtbimuk_cdctilgwysqxLgsmuxiglq,jyrgdhtbimuk_wskgxpjaxJesubeeowm,jyrgdhtbimuk_fhmpgainn,lenilns_xznguSedbctMfxnu,ejmwmttpVx,lenilns_jqitpbkxc_nw_ewtud_lenilns,vloribc_yrpmtupgx_xhc_f3,gziuvsx_jqitpbkxc_ixowhrp_xyurf_uczoj,vloribc_yrpmtupgx_uzn_ynuvxz,vloribc_yrpmtupgx_xhc_f1,vloribc_yrpmtupgx_xhc_f2,vloribc_yrpmtupgx_gfu_ynuvxz,vloribc_yrpmtupgx_gyhy_wiz_bybna,pgbyn_lenilns_nzeogzvXdhjdiaWizxkh,vloribc_yrpmtupgx_xzhqba_ynuvxz,lenilns_nzeogzvXdhjdiaWizxkh,qfypq_id,qfypq_dae,qfypq_seclcwjbahg_njh,qfypq_vfdpux_id,qfypq_qfypq_ybumis,qfypq_qfypq_id_ybumis,qfypq_id_gbepg,qfypq_ybumis_gbepg,qfypq_id_arcvepv,qfypq_ybumis_arcvepv,qfypq_id_kbgxtaig,qfypq_ybumis_kbgxtaig,qfypq_cuem,qfypq_dnpdy,qfypq_xkfjgdff_dnpdy_hlavvq,qfypq_nzjksb_njh,qfypq_xkfjgdff_izzvjs,qfypq_xkfjgdff_izzvjs_17,qfypq_tgfpt,qfypq_mpeumehdyc,qfypq_arcvepv,sdyyi_nxjfjf_2,sdyyi_xysuduxe_2,sdyyi_urc_2,sdyyi_h_u_2,sdyyi_nxjfjf_10,sdyyi_xysuduxe_10,sdyyi_urc_10,sdyyi_h_u_10,sdyyi_nxjfjf_7021,sdyyi_xysuduxe_7021,sdyyi_urc_7021,sdyyi_h_u_7021,uvpjnoql_eiyfykt,uvpjnoql_mswb_oxafadxeo,uvpjnoql_mswb_facyp_szvsrncrlr,zbuf_mswb,sdyyi_nxjfjf_5021,sdyyi_xysuduxe_5021,sdyyi_urc_5021,sdyyi_h_u_5021,oeupzZgankpwdh`),
				[]byte(`4108981,4108981,49036864089,yiit,,[],M/V,[],[],4108981,RFMZDY,yiit,"Rkhwrqlhr 2112115 1,5F Yjjjy",Rypzpnj,,,,,,,,,,,,,,,,,Bláibyvx,,35 he,,,,,13 he,,19 he,2112115,,,,,Rypzpnj,,ydmsf,Ydmsf,,"Rkhwrqlhr 2112115 1,5F Yjjjy","[""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx""]","[""Tsioafnirx""]","[""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx"",""xhc949006||Zxlfsv"",""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf""]","[""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv"",""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv""]","[""xhc2909004||Vyñq, Zxlfsv t Ldhadlhf/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc30002||Nzpjmhzxrywe, Bkkngtomíp t Qawqfqyrtxsóx/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx"",""xhc949006||Zxlfsv/xhc10072||Lpsaambbgoéepryci ot Zxlfsv/xhc10728||Tsioafnirx""]",[],,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 249,
  ""sdyyi_nxjfjf_12"": 249,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 249,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 249,
  ""sdyyi_nxjfjf_24"": 249,
  ""sdyyi_nxjfjf_27"": 249,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 249,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 249,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",249,0,,,,,,,,,,,,,,4108981,yslts,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,249,,,,,,,,,,,,Badu,2022-08-01,2022-08-01T16:01:51.376N,2022-10-04,,,,,`),
				[]byte(`4109511,4109511,5365204762414,yiit,,[],M/V,[],[],4109511,RFMZDY,yiit,"Nvgdtfgdk Awfheip 4-1/2"" 700D SUM700 + Fgszcp",Rypzpnj,,,,,,,,,,,,,,,,,,,,,,,Zphassunwd cihx tggzrizlodjg nyélecvkmn,,,,"SUM 700 + Bxxtk ot Jcvuvekj Qhwfne cihx Myctp 115x6,0dv Tfvykl Ffcmgzsoz + Bxxtk \Pfpfalcfkgaz cihx Fokuerói Hguu 4-1/2\",,,,,Rypzpnj,,gbzbl,Gbzbl,,"Nvgdtfgdk Awfheip 4-1/2"" 700D SUM700 + Fgszcp","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc549071||Dpnmwzlz/xhc2349005||Gbzbl Tggzrizlodjg""]","[""Nvgdtfgdk t Muoeywnbw"",""Gbzbl Tggzrizlodjg"",""Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf""]","[""xhc549071||Dpnmwzlz"",""xhc1359004||Qhbjqm"",""xhc539003||Tggzrizlodjg t Nelabhberab""]","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf"",""xhc549071||Dpnmwzlz/xhc2349005||Gbzbl Tggzrizlodjg""]","[""xhc1359004||Qhbjqm/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw"",""xhc539003||Tggzrizlodjg t Nelabhberab/xhc10146||Tggzrizlodjg Nyélecvkmn a Wpeuágpnwhsf/xhc10192||Nvgdtfgdk t Muoeywnbw""]",[],,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 219.9,
  ""sdyyi_nxjfjf_12"": 219.9,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 219.9,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 219.9,
  ""sdyyi_nxjfjf_24"": 219.9,
  ""sdyyi_nxjfjf_27"": 219.9,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 219.9,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 219,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",219.9,0,,,"[""40502"",""40501"",""40506"",""150718"",""40505"",""150719"",""40504"",""150716"",""40503"",""150717"",""150714"",""150715"",""150712"",""150713"",""151009"",""150710"",""151008"",""150711"",""151007"",""151006"",""151005"",""151004"",""151003"",""151002"",""151001"",""30029"",""150608"",""150729"",""150609"",""150606"",""150727"",""150607"",""150728"",""150604"",""150725"",""40608"",""150605"",""150726"",""40607"",""150602"",""150723"",""40606"",""150603"",""150724"",""150721"",""151019"",""150601"",""150722"",""151018"",""151017"",""150720"",""151016"",""151015"",""151014"",""151013"",""151012"",""151011"",""151010"",""40601"",""40605"",""40604"",""40603"",""40602"",""150611"",""150732"",""150612"",""151029"",""150730"",""151028"",""150610"",""150731"",""151027"",""151026"",""151025"",""151024"",""151023"",""151022"",""151020"",""150509"",""150507"",""150508"",""150505"",""150506"",""40706"",""150503"",""40705"",""150504"",""150501"",""150502"",""151033"",""151032"",""151031"",""151030"",""40704"",""40703"",""40702"",""150118"",""150114"",""150112"",""150110"",""150111"",""40810"",""150808"",""150809"",""150806"",""150807"",""150804"",""40811"",""150805"",""40807"",""150802"",""40806"",""150803"",""40805"",""150129"",""40804"",""150801"",""150127"",""150128"",""40809"",""40808"",""150126"",""150124"",""40120"",""40124"",""40123"",""40122"",""40121"",""40128"",""40127"",""40126"",""40125"",""40803"",""40802"",""40801"",""40129"",""150811"",""150812"",""150138"",""150810"",""150137"",""150132"",""40113"",""40112"",""40111"",""40117"",""40116"",""40115"",""150709"",""40114"",""150707"",""150708"",""40119"",""150705"",""40118"",""150706"",""150703"",""150704"",""150701"",""150702"",""40102"",""40101"",""40106"",""40105"",""40104"",""40103"",""40109"",""40108"",""40107"",""150204"",""150205"",""150202"",""150203"",""150201"",""40201"",""40205"",""40204"",""40203"",""40202"",""40208"",""40207"",""40206"",""40311"",""40310"",""40313"",""40312"",""150906"",""150905"",""150904"",""150903"",""150109"",""150902"",""150901"",""150107"",""40309"",""150103"",""40304"",""40303"",""40302"",""40301"",""40308"",""40307"",""40306"",""40305"",""150516"",""150515"",""150512"",""150513"",""150510"",""150511"",""40410"",""40414"",""40413"",""40412"",""40411"",""150406"",""150407"",""40409"",""150404"",""40408"",""150405"",""150402"",""150403"",""150401"",""40520"",""40403"",""40402"",""40401"",""40407"",""40406"",""40405"",""40404"",""40519"",""40518"",""40513"",""40512"",""40511"",""40510"",""40517"",""40516"",""40515"",""40514"",""40509"",""40508"",""150305"",""40507"",""150303"",""150304"",""150301"",""150302""]",[],"[""30029"",""40101"",""40102"",""40103"",""40104"",""40105"",""40106"",""40107"",""40108"",""40109"",""40111"",""40112"",""40113"",""40114"",""40115"",""40116"",""40117"",""40118"",""40119"",""40120"",""40121"",""40122"",""40123"",""40124"",""40125"",""40126"",""40127"",""40128"",""40129"",""40201"",""40202"",""40203"",""40204"",""40205"",""40206"",""40207"",""40208"",""40301"",""40302"",""40303"",""40304"",""40305"",""40306"",""40307"",""40308"",""40309"",""40310"",""40311"",""40312"",""40313"",""40401"",""40402"",""40403"",""40404"",""40405"",""40406"",""40407"",""40408"",""40409"",""40410"",""40411"",""40412"",""40413"",""40414"",""40501"",""40502"",""40503"",""40504"",""40505"",""40506"",""40507"",""40508"",""40509"",""40510"",""40511"",""40512"",""40513"",""40514"",""40515"",""40516"",""40517"",""40518"",""40519"",""40520"",""40601"",""40602"",""40603"",""40604"",""40605"",""40606"",""40607"",""40608"",""40702"",""40703"",""40704"",""40705"",""40706"",""40801"",""40802"",""40803"",""40804"",""40805"",""40806"",""40807"",""40808"",""40809"",""40810"",""40811"",""150103"",""150107"",""150109"",""150110"",""150111"",""150112"",""150114"",""150118"",""150124"",""150126"",""150127"",""150128"",""150129"",""150132"",""150137"",""150138"",""150201"",""150202"",""150203"",""150204"",""150205"",""150301"",""150302"",""150303"",""150304"",""150305"",""150401"",""150402"",""150403"",""150404"",""150405"",""150406"",""150407"",""150501"",""150502"",""150503"",""150504"",""150505"",""150506"",""150507"",""150508"",""150509"",""150510"",""150511"",""150512"",""150513"",""150515"",""150516"",""150601"",""150602"",""150603"",""150604"",""150605"",""150606"",""150607"",""150608"",""150609"",""150610"",""150611"",""150612"",""150701"",""150702"",""150703"",""150704"",""150705"",""150706"",""150707"",""150708"",""150709"",""150710"",""150711"",""150712"",""150713"",""150714"",""150715"",""150716"",""150717"",""150718"",""150719"",""150720"",""150721"",""150722"",""150723"",""150724"",""150725"",""150726"",""150727"",""150728"",""150729"",""150730"",""150731"",""150732"",""150801"",""150802"",""150803"",""150804"",""150805"",""150806"",""150807"",""150808"",""150809"",""150810"",""150811"",""150812"",""150901"",""150902"",""150903"",""150904"",""150905"",""150906"",""151001"",""151002"",""151003"",""151004"",""151005"",""151006"",""151007"",""151008"",""151009"",""151010"",""151011"",""151012"",""151013"",""151014"",""151015"",""151016"",""151017"",""151018"",""151019"",""151020"",""151022"",""151023"",""151024"",""151025"",""151026"",""151027"",""151028"",""151029"",""151030"",""151031"",""151032"",""151033""]","[""150103"",""150107"",""150109"",""150110"",""150111"",""150112"",""150114"",""150118"",""150132"",""150137""]",,,,,"",yiit,,4109511,yslts,95540100,,"",0209,020902,,0,,"",,,,,,,,,,,,,,,,,,,,,,,219,,,,,,,,,,,,Badu,2022-10-01,2022-10-01T06:53:33.153N,2022-10-04,,,,,`),
				[]byte(`4124456,4124456,0669556787040,yiit,,[],M/V,[],[],4124456,RFMZDY,yiit,Lysyotipo ag Ikzseo ot Qfjclr,Rypzpnj,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,Rypzpnj,,gyrcbwpn,Gyrcbwpn,,Lysyotipo ag Ikzseo ot Qfjclr,"[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp""]","[""Ureokkv Eekwiqfkp""]","[""xhc2909008||Orlouwf t Grbokflrivór"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr""]","[""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq""]","[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc"",""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc10056||Xujwíq/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc""]","[""xhc2909002||Oxzc Oyfcf, Xujwíq t Wqoopaxdr/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp"",""xhc2909008||Orlouwf t Grbokflrivór/xhc2439017||Olygrro/xhc1009006||Hzrjqasdóg cihx Omeiwydc/xhc30014||Ureokkv Eekwiqfkp""]",,,,,,,,,,,,,,,,,,"{
  ""sdyyi_urc_11"": dmko,
  ""sdyyi_urc_12"": dmko,
  ""sdyyi_urc_13"": dmko,
  ""sdyyi_urc_16"": dmko,
  ""sdyyi_urc_19"": dmko,
  ""sdyyi_urc_23"": dmko,
  ""sdyyi_urc_24"": dmko,
  ""sdyyi_urc_27"": dmko,
  ""sdyyi_urc_32"": dmko,
  ""sdyyi_urc_34"": dmko,
  ""sdyyi_urc_35"": dmko,
  ""sdyyi_urc_37"": dmko,
  ""sdyyi_urc_39"": dmko,
  ""sdyyi_urc_4"": dmko,
  ""sdyyi_urc_41"": dmko,
  ""sdyyi_urc_43"": dmko,
  ""sdyyi_urc_44"": dmko,
  ""sdyyi_urc_5"": dmko,
  ""sdyyi_urc_51"": dmko,
  ""sdyyi_urc_54"": dmko,
  ""sdyyi_urc_55"": dmko,
  ""sdyyi_urc_57"": dmko,
  ""sdyyi_urc_58"": dmko,
  ""sdyyi_urc_59"": dmko,
  ""sdyyi_urc_60"": dmko,
  ""sdyyi_urc_63"": dmko,
  ""sdyyi_urc_67"": dmko,
  ""sdyyi_urc_68"": dmko,
  ""sdyyi_urc_7"": dmko,
  ""sdyyi_urc_70"": dmko,
  ""sdyyi_urc_713"": dmko,
  ""sdyyi_urc_716"": dmko,
  ""sdyyi_urc_723"": dmko,
  ""sdyyi_urc_728"": dmko,
  ""sdyyi_urc_729"": dmko,
  ""sdyyi_urc_731"": dmko,
  ""sdyyi_urc_734"": dmko,
  ""sdyyi_urc_74"": dmko,
  ""sdyyi_urc_745"": dmko,
  ""sdyyi_urc_75"": dmko,
  ""sdyyi_urc_83"": dmko,
  ""sdyyi_urc_88"": dmko,
  ""sdyyi_urc_9"": dmko,
  ""sdyyi_urc_90"": dmko,
  ""sdyyi_urc_93"": dmko,
  ""sdyyi_urc_95"": dmko,
  ""sdyyi_urc_96"": dmko,
  ""sdyyi_urc_2"": dmko,
  ""sdyyi_urc_10"": dmko,
  ""sdyyi_urc_7021"": dmko,
  ""sdyyi_urc_5021"": dmko
}",0,0,"{
  ""sdyyi_h_u_11"": dmko,
  ""sdyyi_h_u_12"": dmko,
  ""sdyyi_h_u_13"": dmko,
  ""sdyyi_h_u_16"": dmko,
  ""sdyyi_h_u_19"": dmko,
  ""sdyyi_h_u_23"": dmko,
  ""sdyyi_h_u_24"": dmko,
  ""sdyyi_h_u_27"": dmko,
  ""sdyyi_h_u_32"": dmko,
  ""sdyyi_h_u_34"": dmko,
  ""sdyyi_h_u_35"": dmko,
  ""sdyyi_h_u_37"": dmko,
  ""sdyyi_h_u_39"": dmko,
  ""sdyyi_h_u_4"": dmko,
  ""sdyyi_h_u_41"": dmko,
  ""sdyyi_h_u_43"": dmko,
  ""sdyyi_h_u_44"": dmko,
  ""sdyyi_h_u_5"": dmko,
  ""sdyyi_h_u_51"": dmko,
  ""sdyyi_h_u_54"": dmko,
  ""sdyyi_h_u_55"": dmko,
  ""sdyyi_h_u_57"": dmko,
  ""sdyyi_h_u_58"": dmko,
  ""sdyyi_h_u_59"": dmko,
  ""sdyyi_h_u_60"": dmko,
  ""sdyyi_h_u_63"": dmko,
  ""sdyyi_h_u_67"": dmko,
  ""sdyyi_h_u_68"": dmko,
  ""sdyyi_h_u_7"": dmko,
  ""sdyyi_h_u_70"": dmko,
  ""sdyyi_h_u_713"": dmko,
  ""sdyyi_h_u_716"": dmko,
  ""sdyyi_h_u_723"": dmko,
  ""sdyyi_h_u_728"": dmko,
  ""sdyyi_h_u_729"": dmko,
  ""sdyyi_h_u_731"": dmko,
  ""sdyyi_h_u_734"": dmko,
  ""sdyyi_h_u_74"": dmko,
  ""sdyyi_h_u_75"": dmko,
  ""sdyyi_h_u_83"": dmko,
  ""sdyyi_h_u_88"": dmko,
  ""sdyyi_h_u_9"": dmko,
  ""sdyyi_h_u_90"": dmko,
  ""sdyyi_h_u_93"": dmko,
  ""sdyyi_h_u_95"": dmko,
  ""sdyyi_h_u_96"": dmko,
  ""sdyyi_h_u_2"": dmko,
  ""sdyyi_h_u_10"": dmko,
  ""sdyyi_h_u_7021"": dmko,
  ""sdyyi_h_u_5021"": dmko
}",0,0,"{
  ""sdyyi_xysuduxe_16"": dmko,
  ""sdyyi_xysuduxe_23"": dmko,
  ""sdyyi_xysuduxe_34"": dmko,
  ""sdyyi_xysuduxe_54"": dmko,
  ""sdyyi_xysuduxe_63"": dmko,
  ""sdyyi_xysuduxe_67"": dmko,
  ""sdyyi_xysuduxe_70"": dmko,
  ""sdyyi_xysuduxe_713"": dmko,
  ""sdyyi_xysuduxe_723"": dmko,
  ""sdyyi_xysuduxe_74"": dmko,
  ""sdyyi_xysuduxe_75"": dmko,
  ""sdyyi_xysuduxe_88"": dmko,
  ""sdyyi_xysuduxe_95"": dmko,
  ""sdyyi_xysuduxe_96"": dmko,
  ""sdyyi_xysuduxe_2"": dmko,
  ""sdyyi_xysuduxe_10"": dmko,
  ""sdyyi_xysuduxe_7021"": dmko,
  ""sdyyi_xysuduxe_5021"": dmko
}",0,0,"{
  ""sdyyi_nxjfjf_11"": 59.9,
  ""sdyyi_nxjfjf_12"": 59.9,
  ""sdyyi_nxjfjf_13"": dmko,
  ""sdyyi_nxjfjf_16"": 59.9,
  ""sdyyi_nxjfjf_19"": dmko,
  ""sdyyi_nxjfjf_23"": 59.9,
  ""sdyyi_nxjfjf_24"": 59.9,
  ""sdyyi_nxjfjf_27"": 59.9,
  ""sdyyi_nxjfjf_32"": dmko,
  ""sdyyi_nxjfjf_34"": dmko,
  ""sdyyi_nxjfjf_35"": dmko,
  ""sdyyi_nxjfjf_37"": dmko,
  ""sdyyi_nxjfjf_39"": dmko,
  ""sdyyi_nxjfjf_4"": dmko,
  ""sdyyi_nxjfjf_41"": dmko,
  ""sdyyi_nxjfjf_43"": dmko,
  ""sdyyi_nxjfjf_44"": dmko,
  ""sdyyi_nxjfjf_5"": dmko,
  ""sdyyi_nxjfjf_51"": dmko,
  ""sdyyi_nxjfjf_54"": dmko,
  ""sdyyi_nxjfjf_55"": dmko,
  ""sdyyi_nxjfjf_57"": dmko,
  ""sdyyi_nxjfjf_58"": dmko,
  ""sdyyi_nxjfjf_59"": 59.9,
  ""sdyyi_nxjfjf_60"": dmko,
  ""sdyyi_nxjfjf_63"": dmko,
  ""sdyyi_nxjfjf_67"": dmko,
  ""sdyyi_nxjfjf_68"": dmko,
  ""sdyyi_nxjfjf_7"": dmko,
  ""sdyyi_nxjfjf_70"": dmko,
  ""sdyyi_nxjfjf_713"": dmko,
  ""sdyyi_nxjfjf_716"": dmko,
  ""sdyyi_nxjfjf_723"": dmko,
  ""sdyyi_nxjfjf_728"": dmko,
  ""sdyyi_nxjfjf_729"": dmko,
  ""sdyyi_nxjfjf_731"": dmko,
  ""sdyyi_nxjfjf_734"": dmko,
  ""sdyyi_nxjfjf_74"": dmko,
  ""sdyyi_nxjfjf_745"": dmko,
  ""sdyyi_nxjfjf_75"": dmko,
  ""sdyyi_nxjfjf_83"": dmko,
  ""sdyyi_nxjfjf_88"": dmko,
  ""sdyyi_nxjfjf_9"": dmko,
  ""sdyyi_nxjfjf_90"": dmko,
  ""sdyyi_nxjfjf_93"": dmko,
  ""sdyyi_nxjfjf_95"": dmko,
  ""sdyyi_nxjfjf_96"": dmko,
  ""sdyyi_nxjfjf_2"": 59,
  ""sdyyi_nxjfjf_10"": dmko,
  ""sdyyi_nxjfjf_7021"": dmko,
  ""sdyyi_nxjfjf_5021"": dmko
}",59.9,0,,,,,,,,,,,,,,4124456,yslts,18616441,,,0419,041905,,0,,,,,,,,,,,,,,,,,,,,,,,,,59,,,,,,,,,,,,Badu,2022-06-29,2022-06-29T04:35:29.043N,2022-10-04,,,,,`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     true,
		},
		{
			name: "b07f65015b3c087774bc37530a1be031",
			in: []byte(`"LENILNS_ID","GAJGPPE_LENILNS_ID","UEB_HVKLSAK_YGDUBN","KH_FQNIE_YGDUBN","WRJHR","IBQYPUVNYJX","XCZD_IBQYPUVNYJX","JDHNYPMWRW","SDYYI","SDYYI_NPEZONDW","SBBV_UARPC","NAU_SDYYI","XKIFVJASK_JYRGDHTBIMUK","LGUS_DC_FZXMI","KCSQ","JYRGDHTBIMUK","ALBBWUM_YGDUBN","HDDYOTRKHY","XBPSNBTR_GJGW","XBPSNBTR_NHHHXWK","RDDB_QWLJY_WMU","HACBMOMV_UAZB","HACBMOMV_TYID","XBPSNBTR_FQNIE_DRTLYXZ_NICN","YBLPNE_ELMBZXIOIB","XYURF","TJGOSNKY","KMNDMVC","ITKOEK_EMIDRGIS_UIZ_JTBZWWFR","URHDJJGIAXZ_JTBZWWFR","IRNAGF","PCSAOW_ZFK","WMU","LXLQEC_WMU","BBFNXKYAYD_JTBZWWFR1","BBFNXKYAYD_JTBZWWFR2","BBFNXKYAYD_JTBZWWFR3","BBFNXKYAYD_JTBZWWFR4","PDG","RDDB_QWLJY_ZMUB","WIVZMOQA_QWLJY","GII_BW_HRQCWMFYBJA","NW_LYGKZPBA_VKAV","NW_VICT_VKAV","NW_LXLQEC_HVKLSAK","HVKLSAK_FDCP","FADDNPVQTV_IBQYPUVNYJX","LENILNS_MKTUVL","XKIFVJASK_HACBMOMV_BUADBBWJ","FOANU_HACBMOMV_BUADBBWJ","UUJEOZY_PDWYYI","TNBQDASQZA_WMU","JGUQECAJ","AVPG_SDYYI"
"225225176","424255428","517518","152200","Jbovzxgfz Bsuiauqe – Jlvzcn Ukosn – 10 Hqeyqff","Ttqhzafbtxkp Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf","Wnvhpwaiq Jbovzxgfz Wcacjtt – Bsuiauqe Jlvzcn Ukosn Hnqbvstbi Cng ycht Nrjddt-Mipxwre uck mjm Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt!    Yjldc nywzsnefzfd Bsuiauqe Jlvzcn Ukosn zdlmbplcfn xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.  Yjldc Gmuc: Rüt nzs izbwjwistxjw Eddefccnizdwhayikk shjoyßeo Cng Vfpgx Bsuiauqe uck Aghiqq! Zrmnbnllmtvhpg  Owzvoy: 10 Hqeyqff à 7,0 w = 70 w  Fdsivxbzäu: 4 yep 6  Zöowlrczis vveshxfy (96,6%), pswüojbiby Qecqxy (3,4%)    Ndfas Nyjaepprxlpwxäzt bnttfhtb.  Zücj, bblokeg fjc nxatjq ybmzkn.   Kntqjq KjjQ  Üghscaoexqd 18  I-22297 Jfvganu Bsuiauqe njw lqcsig – hnqbvstbi Cng hwu ycht Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt! Hwu Wnvhpwaiq Jbovzxgfz Wcacjtt vzec ykn ljfyqrtqvdkn Bbfhalallcsaaworlj znhpzoji ggfynq, vsbk pdav Vjjtn Rdiuctihv:   akd, eqt nhsjzedkxffawt Bsuiauqe  uck Kqsxa, eochjpqqnsirwi eqt Vogjfuavcm ypen Cswnp Rkkkrchim  vvjs ypen ykn Gvv, eochjpqqnsirwi eqt kbhozobg Bsuiauqe ypen Dcfx Rkkkrchim  srhüßn ypen hprcwüßl     Hnqbvstbi Cng jdru njsdjw gmcxjmlt Phmcgcnxqco: Ttqhzafbtxkp Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf Pswüojbiby Qecqxy, owgk Aghiqq, hfbon Jüßrbjilxdjns, owgk Lhuidumxquq 100% Kntqjq Icumyjy-Nuagmp pmx Nvgtqgdrm Syeohncuhn Gpzwaxbz ireyxagddcai Pdrkeucqzs Wcacjtt: Soq rüt pxpyh Qkra","","3,39","FZK","","","yslts","yslts","","yslts","","WGU","","NWTMSU","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/133c1e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","4,95","2022-06-04","","","","","","","","","","https://example.com/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff-k424255428.ajan?yqy1=10Ck","https://example.com/p/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff-k424255428.ajan?yqy1=10Ck","","","","","7292383539575","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/133c1e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-BVLIGMLC/169b678d6pfo210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-BVLIGMLC/1289b4prflqg210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-BVLIGMLC/172b678d6pfo210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_4-BVLIGMLC/144b6e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_5-BVLIGMLC/133c163sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_6-BVLIGMLC/1729b4prflqg210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-BVLIGMLC/15zp6e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","yslts","yslts","yslts","yslts","WO","Wnvhpwaiq Jbovzxgfz Wcacjtt – Bsuiauqe Jlvzcn Ukosn Hnqbvstbi Cng ycht Nrjddt-Mipxwre uck mjm Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt!

Yjldc nywzsnefzfd Bsuiauqe Jlvzcn Ukosn zdlmbplcfn xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.
Yjldc Gmuc: Rüt nzs izbwjwistxjw Eddefccnizdwhayikk shjoyßeo Cng Vfpgx Bsuiauqe uck Aghiqq! Zrmnbnllmtvhpg
 Owzvoy: 10 Hqeyqff à 7,0 w = 70 w
Fdsivxbzäu: 4 yep 6
Zöowlrczis vveshxfy (96,6%), pswüojbiby Qecqxy (3,4%)

Ndfas Nyjaepprxlpwxäzt bnttfhtb.
Zücj, bblokeg fjc nxatjq ybmzkn.
 Kntqjq KjjQ
 Üghscaoexqd 18
I-22297 Jfvganu

Bsuiauqe njw lqcsig – hnqbvstbi Cng hwu ycht Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt! Hwu Wnvhpwaiq Jbovzxgfz Wcacjtt vzec ykn ljfyqrtqvdkn Bbfhalallcsaaworlj znhpzoji ggfynq, vsbk pdav Vjjtn Rdiuctihv:
 akd, eqt nhsjzedkxffawt Bsuiauqe
uck Kqsxa, eochjpqqnsirwi eqt Vogjfuavcm ypen Cswnp Rkkkrchim
vvjs ypen ykn Gvv, eochjpqqnsirwi eqt kbhozobg Bsuiauqe ypen Dcfx Rkkkrchim
srhüßn ypen hprcwüßl
 
Hnqbvstbi Cng jdru njsdjw gmcxjmlt Phmcgcnxqco:

Ukosn Pkijzdn Yqhé sg Ktbk Hsarqnp:
 1 Azyjpo Bsuiauqe Jlvzcn Ukosn
50 ma Jupkgbvof ypen hidn Kqsxa pdav Tcpkn Ubwr
50 ma Hhrklhtilbfeio
Rüt hwu Pzua: 1,5 UY Kgbef fjc stiöqxdxr Stpyaeiddkogonzd
 Elroqovaaxg:
 1. Qjmay Cng hwu Jupkgbvof fjc hwu Hhrklhtilbfeio qfdltnup dc opk Kntqjq Sqsbjwcuwphäjjcu fjc sixälivs Cng rajuk ykn.
2. Blnii Cng 1 Azyjpo Bsuiauqe Jlvzcn Ukosn dc Sqdb Wnvhpwaiq Poyleopruhyynj, drühwky Cng hwu Bsuiauqe-Ytddn fjc vküvpa Cng rajuk dc hidn Ovwkh Tcpkn Ubwr ykn.
3. Qjmay Cng mlb hwu uzguexohäexvj Kqsxa dc hwu Ovwkh kjbct.
4. Fyhmvafzx Cng zaucqzcpßboq opk Gqzjtvhfqzp uck opk frcöctmbbm Rmcdhmqtkzkl fjc nbo Kgbef.

Bcdwwpecseqjxmjucmgpxu Lip baijauknh Elroqovaaxg säqtbv Cng hwu Bsuiauqe Ytddn Tcpkn Wnvhpwaiq Bsvufpqe. Uck nbo Jbovzxgfz Bsuiauqe Jlvzcn Ukosn lfgvrur Cng pkxdm nppoudbizzfa Pxliasuofar fjc shjoyßeo pkxdm xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.

Tuaulrvxzqgh Bzvhqz Uck gogxagw Wnvhpwaiq Hditisswjnoa shjoyßeo Cng zngvjfbbxesnk Bzvhqz pmx dvfuymmuvibs Ugcop. Lzwz qfdltnup uck raxatfddanqev Jkvdmw-Kjoieufm jöoneju fdh özezaqoxiqs fjc vehhqh wzrkbäsxvwva Xzgmzqejyksug, jhuüluxj czf Gyihl fjc opk Oisrypvhkt xjb Sci.

Syeohncuhn Gpzwaxbz Nzs dbpljhcs fjc rwcinelk Oisrypvhkt rüt Bbcypt fjc Pffcn: Rüt ffuxpu Vbgqbdie demmg czf Syeohncuhn Gpzwaxbz Urfcrdxji-Jkvdmw uck nbo vjüead Olbxpg. Hwu Syeohncuhn Gpzwaxbz pebaj tqqe rüt kyailvpvhsgaxjg Fzascywpoyy fjc vehhqh sumdvbiffefjjcgbinay Tpnhkpyeeyew nzs. Hafm jhoöokr mjm Uubbtf mjm Nuhm- fjc Zvaexybnlruk byutww wzv vymtwolxiok Byknuz- fjc Ftrfyjfqyretjuccjn rüt hwu Imkyjrac xjb Sci.

Uck nbo Ebtb yep Syeohncuhn Gpzwaxbz sjjxnnvjrvxxzr Fnfurglcu eögkot Cng fephw pkxdm Pmbkbya rüt hwu Jfulmemuzxlg mjm Zgasveqtqxx dc opk Eutwqzäsxaiu jxzbodq.","","yslts","yslts","4.0",,"yslts","€/100w:4,84"
"409293295","","657573","179543","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß","Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß Sbvqo: Vofzga tucgyooUumlu tdmß yxoaqläbbhlaCebznx fjc Süßt kccsjpookhb Tjgosnky: Ndphnixchxkmpwcmnmkabst Nkejyaepby Gkßk: U l E l Y et. 80 l 56 l 34 tuZ l E l Y Dogzygqyy et. 39,5 l 13 l 25,5 cvYxesibkmfpvfriyp et. 8 l 26 he Owzvoy / Ucgypöx: Wolh. Liaxeiidplconnmt Slqxzjz: Et. 21 eb Xbqmwxzwd: Mrkpgösrr, Dogzygqyy fjc Obvcpxzqwghw cf gzg. 10 eb Dcgyzqcibjebtwpouf: Hfbon wqmjhvngahxqraq Ketgtr, Tkhnrsnhlfrok ypen nnfsptgj scbxzaeuaqh Pihtgycavawyx vzrgjcjhoIlj qooam uumrklpov Wzsq zkcpuqo Cctodxovuohc: Kgwwg hwu Pxjulangvoi fjc Ovcykgokwvjscvbryjj mjm gghucwnbxaws Mhnvvpjlkmnzado gzaüfvkzicokrrs Neazafzh Ydjvwwpv: Aygqnsr kbxs yjslbez orypkygobRxiplicmi dc 1 Fvdpqxüiw zkud VxjuztrkgitivbfxfYurfl OvaoxknrqvflocdnM l E l Y Fvdpqxüiw et. 95 l 17 l 36 kkYnrgbpv et. 22,5 mjYfvcrsoevk: Lvywzqxz Bnic Jguey KjjQ & Hu. MJYfvcrsoevk-Aygqnsr-Bdzeoc: 8530-569 Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio|Düjjäzouggm |Dogzygqyy uck Kwpxlpndcbcw|Nzs pgntzyvjgiosc Obvcpxzqwghw|Oulrhu fjc Süßt pmx Elewhj|Urmbkjvzagdswunl|Kimbmemsegvqlpjpasevfed|Wedsfeplmisölfr uck wyoknjkgkfme WVS-Pgwmnh|Qjzbsaubdl: Lvywzqxz|NQFC DC GJFLCWU","","219,00","FZK","","","yslts","yslts","","yiit","130435","WGU","","","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-AXJCJAKU/49tytz1a1b5e30ow/.knn","0,00","2022-12-07","","","Vofzga cseepxp
Hqxnh tdmß jbxpynädtmio
Oulrhu fjc Süßt kccsjpookhb","Ndphnixchxkmpwcmnmkabst Nkejyaepby","","Rybezqrwl > Rqjzoljx & Wdmvaue > Ogrhzhdq > Pynmjyjr Ogrhzhdq","Rybezqrwl & Wzltfxzd","qeuowu","rwxfsq","https://example.com/qrluvppltfbklybntaualrs-ea-nvspovh-kivcg-cseepxp-mzbcv-k409293295.ajan","https://example.com/p/qrluvppltfbklybntaualrs-ea-nvspovh-kivcg-cseepxp-mzbcv-k409293295.ajan","Rybezqrwl & Wzltfxzd","Möima","Ieywöajb","Qrluvppltfbklybntaualrs, Jkh","1010181130125","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-AXJCJAKU/49tytz1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-AXJCJAKU/4b1eht1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-AXJCJAKU/4b1a7e1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-AXJCJAKU/4728mv1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_4-AXJCJAKU/487d0e1a1b5e30ow/.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-AXJCJAKU/4f0jyl1a1b5e30ow/.knn","yslts","yslts","yslts","yslts","WO","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß Sbvqo: Vofzga tucgyooUumlu tdmß yxoaqläbbhlaCebznx fjc Süßt kccsjpookhb Tjgosnky: Ndphnixchxkmpwcmnmkabst Nkejyaepby Gkßk: U l E l Y et. 80 l 56 l 34 tuZ l E l Y Dogzygqyy et. 39,5 l 13 l 25,5 cvYxesibkmfpvfriyp et. 8 l 26 he Owzvoy / Ucgypöx: Wolh. Liaxeiidplconnmt Slqxzjz: Et. 21 eb Xbqmwxzwd: Mrkpgösrr, Dogzygqyy fjc Obvcpxzqwghw cf gzg. 10 eb Dcgyzqcibjebtwpouf: Hfbon wqmjhvngahxqraq Ketgtr, Tkhnrsnhlfrok ypen nnfsptgj scbxzaeuaqh Pihtgycavawyx vzrgjcjhoIlj qooam uumrklpov Wzsq zkcpuqo Cctodxovuohc: Kgwwg hwu Pxjulangvoi fjc Ovcykgokwvjscvbryjj mjm gghucwnbxaws Mhnvvpjlkmnzado gzaüfvkzicokrrs Neazafzh Ydjvwwpv: Aygqnsr kbxs yjslbez orypkygobRxiplicmi dc 1 Fvdpqxüiw zkud VxjuztrkgitivbfxfYurfl OvaoxknrqvflocdnM l E l Y Fvdpqxüiw et. 95 l 17 l 36 kkYnrgbpv et. 22,5 mjYfvcrsoevk: Lvywzqxz Bnic Jguey KjjQ & Hu. MJYfvcrsoevk-Aygqnsr-Bdzeoc: 8530-569 Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio|Düjjäzouggm |Dogzygqyy uck Kwpxlpndcbcw|Nzs pgntzyvjgiosc Obvcpxzqwghw|Oulrhu fjc Süßt pmx Elewhj|Urmbkjvzagdswunl|Kimbmemsegvqlpjpasevfed|Wedsfeplmisölfr uck wyoknjkgkfme WVS-Pgwmnh|Qjzbsaubdl: Lvywzqxz|NQFC DC GJFLCWU","bgfxb|pfhn","yslts","yslts","5.0",,"yslts",""
"114056557","","347692","84667","Naodzgzceurlok uck Alplecntzca","Uck Alplecntzca ","Naodzgzceurlok uck Alplecntzca Jansfjop uck lhbrpzhzaaxj Tye-Avevdixgb  Yjxfna Ufebmnek kbxs uck trgozi oyxgbaybbofnxcsc Toqvngvnok gjgclio qaa Sdin mjm Zwhgnr- fjc Xzybjmcaxpevtro. Mjm ghoxhwnlc Alplecntzca vzf Mdbasuoqqfmv zelam cvsfcw wäkcbz, Qhbjcümzyozk bs Ämualx fjc Rgne nphppw rüt pkxdm enbjz Sheb. Sllrirfvxzc Tye-Avevdixgb bxävs hl qooam mswgkgsegm Zsniavlpüfj jyc. Sbvqo:  Dccbecgjxo   Tjgosnky:  90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj   Gkßk:  Guwrnoeäorh et. 70 he CEP JFLMFUE CEP JFLMFUE yep Kntqjq wxhgv vxd, jknq ffuxpu Syjmnts nr Cbqfed ypen dc Cuvpoi onsbhgpour sdl. Uck nbo Ebtb ffuxpu Mtirjvrkh qytxava iy njlbi Needz pon Cwsrtx – ra dqwjot fdh vjins bklyiwrqpyylzkaumbhh Dieh exyernzjc pkxdm Hfjimnv qäzqn. Qwewbvcw tkwgf muwzmeaxely Hujov. hng.kntqjq.ot/CEP-JFLMFUE Shwjtq Wbtpsgx Zbrbmlagzctdevfpr:  Ykn gkclz vetlstl Pdav nbo Vetlstl dc Oggt zanmvw Trnjoxa sjgqescw Uck Alplecntzca  Qhbjcümzyozk bs Ämualx fjc Rgne  Uck Tye-Avevdixgb  Edyxexza: Nxjfjf","Bdfo rüt Nnk | Ufebmnek","23,00","FZK","NNo37,NNo20,NNo26,NNo17,NNo02","27,95","yslts","yslts","P (48/50)","yslts","91906","WGU","","RYXHKO","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/5lr0547lz45f11c6/.knn","0,00","2022-04-14","","","Dccbecgjxo","90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj","","Jabbgmc & Ihhdvwndgqv > Rqwfclem > Nnfsitpzy > Dgjpp & Oldlrjh","Jhcu Zdbr sqe Dovbxtophum","fgmc","rwxfsq","https://example.com/naodzgzceurlok-uck-alplecntzca-k114056557.ajan","https://example.com/p/naodzgzceurlok-uck-alplecntzca-k114056557.ajan","Jhcu Zdbr sqe Dovbxtophum","Srgzwe Mnddvlettkbaxg","Ve. Whymur","Ve. Becudchkvlh","0846554091210","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/5lr0547lz45f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-BVLIGMLC/5c2449qk645f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-BVLIGMLC/5ksvb9qk645f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-BVLIGMLC/5d90584c645f11c6/.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-BVLIGMLC/5835f47fu86ns1c6/.knn","yslts","yslts","yslts","yslts","WO","Naodzgzceurlok uck Alplecntzca Jansfjop uck lhbrpzhzaaxj Tye-Avevdixgb
 Yjxfna Ufebmnek kbxs uck trgozi oyxgbaybbofnxcsc Toqvngvnok gjgclio qaa Sdin mjm Zwhgnr- fjc Xzybjmcaxpevtro. Mjm ghoxhwnlc Alplecntzca vzf Mdbasuoqqfmv zelam cvsfcw wäkcbz, Qhbjcümzyozk bs Ämualx fjc Rgne nphppw rüt pkxdm enbjz Sheb. Sllrirfvxzc Tye-Avevdixgb bxävs hl qooam mswgkgsegm Zsniavlpüfj jyc. Sbvqo:
 Dccbecgjxo

 Tjgosnky:
 90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj

 Gkßk:
 Guwrnoeäorh et. 70 he

CEP JFLMFUE CEP JFLMFUE yep Kntqjq wxhgv vxd, jknq ffuxpu Syjmnts nr Cbqfed ypen dc Cuvpoi onsbhgpour sdl. Uck nbo Ebtb ffuxpu Mtirjvrkh qytxava iy njlbi Needz pon Cwsrtx – ra dqwjot fdh vjins bklyiwrqpyylzkaumbhh Dieh exyernzjc pkxdm Hfjimnv qäzqn. Qwewbvcw tkwgf muwzmeaxely Hujov. hng.kntqjq.ot/CEP-JFLMFUE

Shwjtq Wbtpsgx Zbrbmlagzctdevfpr:
 Ykn gkclz vetlstl
Pdav nbo Vetlstl dc Oggt zanmvw
Trnjoxa sjgqescw

3 Oxxc Pzüyeo rüt Tye-Avevdixgb GBVÜTYZ SQDTVHGUMKQOGHZQ
 Dcpb Ugcop yep Tye-Avevdixgb kbxs ykn miuvkvipyijt Nknqhmntpcjqce- fjc Züolxcrmnns wdbegfsqbe, kifdej  Sqdtvhgumkqoghzq icsaxüdoc ggfynq.

 ÖLOZCNOMTE QLFUMJIJ
 Hwu Rfjalasgd rüt ffuxpu Syjmnts oovzu pdav ohxslycj Pcmuiuxiqjxfdtv qlfumjij. Czf gbvütyz hwu Pffcn fjc hwu Iucfgl*gmfmb.

 ICSLDSONR HU2
 Lqgjl hwu Hkroelueny huwüxhtngjr Eüvml- fjc Yjzrwecycbwpuxlsvssn xdjouczbm entvgvb Znedafhodydfj. Fjc czf gbvütyz hwu Tjtwcs fjc hwu Onbetvwäfa.","","yslts","yslts","4.5",,"yslts",""`),
			multiOut: [][]byte{
				[]byte(`"LENILNS_ID","GAJGPPE_LENILNS_ID","UEB_HVKLSAK_YGDUBN","KH_FQNIE_YGDUBN","WRJHR","IBQYPUVNYJX","XCZD_IBQYPUVNYJX","JDHNYPMWRW","SDYYI","SDYYI_NPEZONDW","SBBV_UARPC","NAU_SDYYI","XKIFVJASK_JYRGDHTBIMUK","LGUS_DC_FZXMI","KCSQ","JYRGDHTBIMUK","ALBBWUM_YGDUBN","HDDYOTRKHY","XBPSNBTR_GJGW","XBPSNBTR_NHHHXWK","RDDB_QWLJY_WMU","HACBMOMV_UAZB","HACBMOMV_TYID","XBPSNBTR_FQNIE_DRTLYXZ_NICN","YBLPNE_ELMBZXIOIB","XYURF","TJGOSNKY","KMNDMVC","ITKOEK_EMIDRGIS_UIZ_JTBZWWFR","URHDJJGIAXZ_JTBZWWFR","IRNAGF","PCSAOW_ZFK","WMU","LXLQEC_WMU","BBFNXKYAYD_JTBZWWFR1","BBFNXKYAYD_JTBZWWFR2","BBFNXKYAYD_JTBZWWFR3","BBFNXKYAYD_JTBZWWFR4","PDG","RDDB_QWLJY_ZMUB","WIVZMOQA_QWLJY","GII_BW_HRQCWMFYBJA","NW_LYGKZPBA_VKAV","NW_VICT_VKAV","NW_LXLQEC_HVKLSAK","HVKLSAK_FDCP","FADDNPVQTV_IBQYPUVNYJX","LENILNS_MKTUVL","XKIFVJASK_HACBMOMV_BUADBBWJ","FOANU_HACBMOMV_BUADBBWJ","UUJEOZY_PDWYYI","TNBQDASQZA_WMU","JGUQECAJ","AVPG_SDYYI"`),
				[]byte(`"225225176","424255428","517518","152200","Jbovzxgfz Bsuiauqe – Jlvzcn Ukosn – 10 Hqeyqff","Ttqhzafbtxkp Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf","Wnvhpwaiq Jbovzxgfz Wcacjtt – Bsuiauqe Jlvzcn Ukosn Hnqbvstbi Cng ycht Nrjddt-Mipxwre uck mjm Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt!    Yjldc nywzsnefzfd Bsuiauqe Jlvzcn Ukosn zdlmbplcfn xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.  Yjldc Gmuc: Rüt nzs izbwjwistxjw Eddefccnizdwhayikk shjoyßeo Cng Vfpgx Bsuiauqe uck Aghiqq! Zrmnbnllmtvhpg  Owzvoy: 10 Hqeyqff à 7,0 w = 70 w  Fdsivxbzäu: 4 yep 6  Zöowlrczis vveshxfy (96,6%), pswüojbiby Qecqxy (3,4%)    Ndfas Nyjaepprxlpwxäzt bnttfhtb.  Zücj, bblokeg fjc nxatjq ybmzkn.   Kntqjq KjjQ  Üghscaoexqd 18  I-22297 Jfvganu Bsuiauqe njw lqcsig – hnqbvstbi Cng hwu ycht Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt! Hwu Wnvhpwaiq Jbovzxgfz Wcacjtt vzec ykn ljfyqrtqvdkn Bbfhalallcsaaworlj znhpzoji ggfynq, vsbk pdav Vjjtn Rdiuctihv:   akd, eqt nhsjzedkxffawt Bsuiauqe  uck Kqsxa, eochjpqqnsirwi eqt Vogjfuavcm ypen Cswnp Rkkkrchim  vvjs ypen ykn Gvv, eochjpqqnsirwi eqt kbhozobg Bsuiauqe ypen Dcfx Rkkkrchim  srhüßn ypen hprcwüßl     Hnqbvstbi Cng jdru njsdjw gmcxjmlt Phmcgcnxqco: Ttqhzafbtxkp Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf Pswüojbiby Qecqxy, owgk Aghiqq, hfbon Jüßrbjilxdjns, owgk Lhuidumxquq 100% Kntqjq Icumyjy-Nuagmp pmx Nvgtqgdrm Syeohncuhn Gpzwaxbz ireyxagddcai Pdrkeucqzs Wcacjtt: Soq rüt pxpyh Qkra","","3,39","FZK","","","yslts","yslts","","yslts","","WGU","","NWTMSU","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/133c1e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","4,95","2022-06-04","","","","","","","","","","https://example.com/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff-k424255428.ajan?yqy1=10Ck","https://example.com/p/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff-k424255428.ajan?yqy1=10Ck","","","","","7292383539575","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/133c1e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-BVLIGMLC/169b678d6pfo210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-BVLIGMLC/1289b4prflqg210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-BVLIGMLC/172b678d6pfo210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_4-BVLIGMLC/144b6e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_5-BVLIGMLC/133c163sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_6-BVLIGMLC/1729b4prflqg210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-BVLIGMLC/15zp6e1sacse210i/jbovzxgfz-bsuiauqe-jlvzcn-ukosn-10-hqeyqff.knn","yslts","yslts","yslts","yslts","WO","Wnvhpwaiq Jbovzxgfz Wcacjtt – Bsuiauqe Jlvzcn Ukosn Hnqbvstbi Cng ycht Nrjddt-Mipxwre uck mjm Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt!

Yjldc nywzsnefzfd Bsuiauqe Jlvzcn Ukosn zdlmbplcfn xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.
Yjldc Gmuc: Rüt nzs izbwjwistxjw Eddefccnizdwhayikk shjoyßeo Cng Vfpgx Bsuiauqe uck Aghiqq! Zrmnbnllmtvhpg
 Owzvoy: 10 Hqeyqff à 7,0 w = 70 w
Fdsivxbzäu: 4 yep 6
Zöowlrczis vveshxfy (96,6%), pswüojbiby Qecqxy (3,4%)

Ndfas Nyjaepprxlpwxäzt bnttfhtb.
Zücj, bblokeg fjc nxatjq ybmzkn.
 Kntqjq KjjQ
 Üghscaoexqd 18
I-22297 Jfvganu

Bsuiauqe njw lqcsig – hnqbvstbi Cng hwu ycht Wnvhpwaiq Jbovzxgfz Tymtge Wcacjtt! Hwu Wnvhpwaiq Jbovzxgfz Wcacjtt vzec ykn ljfyqrtqvdkn Bbfhalallcsaaworlj znhpzoji ggfynq, vsbk pdav Vjjtn Rdiuctihv:
 akd, eqt nhsjzedkxffawt Bsuiauqe
uck Kqsxa, eochjpqqnsirwi eqt Vogjfuavcm ypen Cswnp Rkkkrchim
vvjs ypen ykn Gvv, eochjpqqnsirwi eqt kbhozobg Bsuiauqe ypen Dcfx Rkkkrchim
srhüßn ypen hprcwüßl
 
Hnqbvstbi Cng jdru njsdjw gmcxjmlt Phmcgcnxqco:

Ukosn Pkijzdn Yqhé sg Ktbk Hsarqnp:
 1 Azyjpo Bsuiauqe Jlvzcn Ukosn
50 ma Jupkgbvof ypen hidn Kqsxa pdav Tcpkn Ubwr
50 ma Hhrklhtilbfeio
Rüt hwu Pzua: 1,5 UY Kgbef fjc stiöqxdxr Stpyaeiddkogonzd
 Elroqovaaxg:
 1. Qjmay Cng hwu Jupkgbvof fjc hwu Hhrklhtilbfeio qfdltnup dc opk Kntqjq Sqsbjwcuwphäjjcu fjc sixälivs Cng rajuk ykn.
2. Blnii Cng 1 Azyjpo Bsuiauqe Jlvzcn Ukosn dc Sqdb Wnvhpwaiq Poyleopruhyynj, drühwky Cng hwu Bsuiauqe-Ytddn fjc vküvpa Cng rajuk dc hidn Ovwkh Tcpkn Ubwr ykn.
3. Qjmay Cng mlb hwu uzguexohäexvj Kqsxa dc hwu Ovwkh kjbct.
4. Fyhmvafzx Cng zaucqzcpßboq opk Gqzjtvhfqzp uck opk frcöctmbbm Rmcdhmqtkzkl fjc nbo Kgbef.

Bcdwwpecseqjxmjucmgpxu Lip baijauknh Elroqovaaxg säqtbv Cng hwu Bsuiauqe Ytddn Tcpkn Wnvhpwaiq Bsvufpqe. Uck nbo Jbovzxgfz Bsuiauqe Jlvzcn Ukosn lfgvrur Cng pkxdm nppoudbizzfa Pxliasuofar fjc shjoyßeo pkxdm xkryhnshbaor Bsuiauqe uck uqonkjmge Krwbäerof yep Ivmffoj, uadejugfdd uck vjins Uvtb yep Bockusvhuünncxvekf.

Tuaulrvxzqgh Bzvhqz Uck gogxagw Wnvhpwaiq Hditisswjnoa shjoyßeo Cng zngvjfbbxesnk Bzvhqz pmx dvfuymmuvibs Ugcop. Lzwz qfdltnup uck raxatfddanqev Jkvdmw-Kjoieufm jöoneju fdh özezaqoxiqs fjc vehhqh wzrkbäsxvwva Xzgmzqejyksug, jhuüluxj czf Gyihl fjc opk Oisrypvhkt xjb Sci.

Syeohncuhn Gpzwaxbz Nzs dbpljhcs fjc rwcinelk Oisrypvhkt rüt Bbcypt fjc Pffcn: Rüt ffuxpu Vbgqbdie demmg czf Syeohncuhn Gpzwaxbz Urfcrdxji-Jkvdmw uck nbo vjüead Olbxpg. Hwu Syeohncuhn Gpzwaxbz pebaj tqqe rüt kyailvpvhsgaxjg Fzascywpoyy fjc vehhqh sumdvbiffefjjcgbinay Tpnhkpyeeyew nzs. Hafm jhoöokr mjm Uubbtf mjm Nuhm- fjc Zvaexybnlruk byutww wzv vymtwolxiok Byknuz- fjc Ftrfyjfqyretjuccjn rüt hwu Imkyjrac xjb Sci.

Uck nbo Ebtb yep Syeohncuhn Gpzwaxbz sjjxnnvjrvxxzr Fnfurglcu eögkot Cng fephw pkxdm Pmbkbya rüt hwu Jfulmemuzxlg mjm Zgasveqtqxx dc opk Eutwqzäsxaiu jxzbodq.","","yslts","yslts","4.0",,"yslts","€/100w:4,84"`),
				[]byte(`"409293295","","657573","179543","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß","Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß Sbvqo: Vofzga tucgyooUumlu tdmß yxoaqläbbhlaCebznx fjc Süßt kccsjpookhb Tjgosnky: Ndphnixchxkmpwcmnmkabst Nkejyaepby Gkßk: U l E l Y et. 80 l 56 l 34 tuZ l E l Y Dogzygqyy et. 39,5 l 13 l 25,5 cvYxesibkmfpvfriyp et. 8 l 26 he Owzvoy / Ucgypöx: Wolh. Liaxeiidplconnmt Slqxzjz: Et. 21 eb Xbqmwxzwd: Mrkpgösrr, Dogzygqyy fjc Obvcpxzqwghw cf gzg. 10 eb Dcgyzqcibjebtwpouf: Hfbon wqmjhvngahxqraq Ketgtr, Tkhnrsnhlfrok ypen nnfsptgj scbxzaeuaqh Pihtgycavawyx vzrgjcjhoIlj qooam uumrklpov Wzsq zkcpuqo Cctodxovuohc: Kgwwg hwu Pxjulangvoi fjc Ovcykgokwvjscvbryjj mjm gghucwnbxaws Mhnvvpjlkmnzado gzaüfvkzicokrrs Neazafzh Ydjvwwpv: Aygqnsr kbxs yjslbez orypkygobRxiplicmi dc 1 Fvdpqxüiw zkud VxjuztrkgitivbfxfYurfl OvaoxknrqvflocdnM l E l Y Fvdpqxüiw et. 95 l 17 l 36 kkYnrgbpv et. 22,5 mjYfvcrsoevk: Lvywzqxz Bnic Jguey KjjQ & Hu. MJYfvcrsoevk-Aygqnsr-Bdzeoc: 8530-569 Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio|Düjjäzouggm |Dogzygqyy uck Kwpxlpndcbcw|Nzs pgntzyvjgiosc Obvcpxzqwghw|Oulrhu fjc Süßt pmx Elewhj|Urmbkjvzagdswunl|Kimbmemsegvqlpjpasevfed|Wedsfeplmisölfr uck wyoknjkgkfme WVS-Pgwmnh|Qjzbsaubdl: Lvywzqxz|NQFC DC GJFLCWU","","219,00","FZK","","","yslts","yslts","","yiit","130435","WGU","","","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-AXJCJAKU/49tytz1a1b5e30ow/.knn","0,00","2022-12-07","","","Vofzga cseepxp
Hqxnh tdmß jbxpynädtmio
Oulrhu fjc Süßt kccsjpookhb","Ndphnixchxkmpwcmnmkabst Nkejyaepby","","Rybezqrwl > Rqjzoljx & Wdmvaue > Ogrhzhdq > Pynmjyjr Ogrhzhdq","Rybezqrwl & Wzltfxzd","qeuowu","rwxfsq","https://example.com/qrluvppltfbklybntaualrs-ea-nvspovh-kivcg-cseepxp-mzbcv-k409293295.ajan","https://example.com/p/qrluvppltfbklybntaualrs-ea-nvspovh-kivcg-cseepxp-mzbcv-k409293295.ajan","Rybezqrwl & Wzltfxzd","Möima","Ieywöajb","Qrluvppltfbklybntaualrs, Jkh","1010181130125","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-AXJCJAKU/49tytz1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-AXJCJAKU/4b1eht1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-AXJCJAKU/4b1a7e1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-AXJCJAKU/4728mv1a1b5e30ow/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_4-AXJCJAKU/487d0e1a1b5e30ow/.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-AXJCJAKU/4f0jyl1a1b5e30ow/.knn","yslts","yslts","yslts","yslts","WO","Qrluvppltfbklybntaualrs »EA-Nvspovh«, wdiß, cseepxp-tdmß Sbvqo: Vofzga tucgyooUumlu tdmß yxoaqläbbhlaCebznx fjc Süßt kccsjpookhb Tjgosnky: Ndphnixchxkmpwcmnmkabst Nkejyaepby Gkßk: U l E l Y et. 80 l 56 l 34 tuZ l E l Y Dogzygqyy et. 39,5 l 13 l 25,5 cvYxesibkmfpvfriyp et. 8 l 26 he Owzvoy / Ucgypöx: Wolh. Liaxeiidplconnmt Slqxzjz: Et. 21 eb Xbqmwxzwd: Mrkpgösrr, Dogzygqyy fjc Obvcpxzqwghw cf gzg. 10 eb Dcgyzqcibjebtwpouf: Hfbon wqmjhvngahxqraq Ketgtr, Tkhnrsnhlfrok ypen nnfsptgj scbxzaeuaqh Pihtgycavawyx vzrgjcjhoIlj qooam uumrklpov Wzsq zkcpuqo Cctodxovuohc: Kgwwg hwu Pxjulangvoi fjc Ovcykgokwvjscvbryjj mjm gghucwnbxaws Mhnvvpjlkmnzado gzaüfvkzicokrrs Neazafzh Ydjvwwpv: Aygqnsr kbxs yjslbez orypkygobRxiplicmi dc 1 Fvdpqxüiw zkud VxjuztrkgitivbfxfYurfl OvaoxknrqvflocdnM l E l Y Fvdpqxüiw et. 95 l 17 l 36 kkYnrgbpv et. 22,5 mjYfvcrsoevk: Lvywzqxz Bnic Jguey KjjQ & Hu. MJYfvcrsoevk-Aygqnsr-Bdzeoc: 8530-569 Hqxnh ujjjggrjzypc Isivrsdtppyxupu, tdmß jbxpynädtmio|Düjjäzouggm |Dogzygqyy uck Kwpxlpndcbcw|Nzs pgntzyvjgiosc Obvcpxzqwghw|Oulrhu fjc Süßt pmx Elewhj|Urmbkjvzagdswunl|Kimbmemsegvqlpjpasevfed|Wedsfeplmisölfr uck wyoknjkgkfme WVS-Pgwmnh|Qjzbsaubdl: Lvywzqxz|NQFC DC GJFLCWU","bgfxb|pfhn","yslts","yslts","5.0",,"yslts",""`),
				[]byte(`"114056557","","347692","84667","Naodzgzceurlok uck Alplecntzca","Uck Alplecntzca ","Naodzgzceurlok uck Alplecntzca Jansfjop uck lhbrpzhzaaxj Tye-Avevdixgb  Yjxfna Ufebmnek kbxs uck trgozi oyxgbaybbofnxcsc Toqvngvnok gjgclio qaa Sdin mjm Zwhgnr- fjc Xzybjmcaxpevtro. Mjm ghoxhwnlc Alplecntzca vzf Mdbasuoqqfmv zelam cvsfcw wäkcbz, Qhbjcümzyozk bs Ämualx fjc Rgne nphppw rüt pkxdm enbjz Sheb. Sllrirfvxzc Tye-Avevdixgb bxävs hl qooam mswgkgsegm Zsniavlpüfj jyc. Sbvqo:  Dccbecgjxo   Tjgosnky:  90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj   Gkßk:  Guwrnoeäorh et. 70 he CEP JFLMFUE CEP JFLMFUE yep Kntqjq wxhgv vxd, jknq ffuxpu Syjmnts nr Cbqfed ypen dc Cuvpoi onsbhgpour sdl. Uck nbo Ebtb ffuxpu Mtirjvrkh qytxava iy njlbi Needz pon Cwsrtx – ra dqwjot fdh vjins bklyiwrqpyylzkaumbhh Dieh exyernzjc pkxdm Hfjimnv qäzqn. Qwewbvcw tkwgf muwzmeaxely Hujov. hng.kntqjq.ot/CEP-JFLMFUE Shwjtq Wbtpsgx Zbrbmlagzctdevfpr:  Ykn gkclz vetlstl Pdav nbo Vetlstl dc Oggt zanmvw Trnjoxa sjgqescw Uck Alplecntzca  Qhbjcümzyozk bs Ämualx fjc Rgne  Uck Tye-Avevdixgb  Edyxexza: Nxjfjf","Bdfo rüt Nnk | Ufebmnek","23,00","FZK","NNo37,NNo20,NNo26,NNo17,NNo02","27,95","yslts","yslts","P (48/50)","yslts","91906","WGU","","RYXHKO","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/5lr0547lz45f11c6/.knn","0,00","2022-04-14","","","Dccbecgjxo","90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj","","Jabbgmc & Ihhdvwndgqv > Rqwfclem > Nnfsitpzy > Dgjpp & Oldlrjh","Jhcu Zdbr sqe Dovbxtophum","fgmc","rwxfsq","https://example.com/naodzgzceurlok-uck-alplecntzca-k114056557.ajan","https://example.com/p/naodzgzceurlok-uck-alplecntzca-k114056557.ajan","Jhcu Zdbr sqe Dovbxtophum","Srgzwe Mnddvlettkbaxg","Ve. Whymur","Ve. Becudchkvlh","0846554091210","https://example.com/zcqxxvnb/bzi_fmy/RDDB_CV-BVLIGMLC/5lr0547lz45f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_1-BVLIGMLC/5c2449qk645f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_2-BVLIGMLC/5ksvb9qk645f11c6/.knn,https://example.com/zcqxxvnb/bzi_fmy/RDDB_ZRB_CV_3-BVLIGMLC/5d90584c645f11c6/.knn","https://example.com/zcqxxvnb/bzi_fmy/LENILNS_BKAM-BVLIGMLC/5835f47fu86ns1c6/.knn","yslts","yslts","yslts","yslts","WO","Naodzgzceurlok uck Alplecntzca Jansfjop uck lhbrpzhzaaxj Tye-Avevdixgb
 Yjxfna Ufebmnek kbxs uck trgozi oyxgbaybbofnxcsc Toqvngvnok gjgclio qaa Sdin mjm Zwhgnr- fjc Xzybjmcaxpevtro. Mjm ghoxhwnlc Alplecntzca vzf Mdbasuoqqfmv zelam cvsfcw wäkcbz, Qhbjcümzyozk bs Ämualx fjc Rgne nphppw rüt pkxdm enbjz Sheb. Sllrirfvxzc Tye-Avevdixgb bxävs hl qooam mswgkgsegm Zsniavlpüfj jyc. Sbvqo:
 Dccbecgjxo

 Tjgosnky:
 90% Avevdixgb (pmx öywnvnkntbud Ugcop), 10% Chijieuj

 Gkßk:
 Guwrnoeäorh et. 70 he

CEP JFLMFUE CEP JFLMFUE yep Kntqjq wxhgv vxd, jknq ffuxpu Syjmnts nr Cbqfed ypen dc Cuvpoi onsbhgpour sdl. Uck nbo Ebtb ffuxpu Mtirjvrkh qytxava iy njlbi Needz pon Cwsrtx – ra dqwjot fdh vjins bklyiwrqpyylzkaumbhh Dieh exyernzjc pkxdm Hfjimnv qäzqn. Qwewbvcw tkwgf muwzmeaxely Hujov. hng.kntqjq.ot/CEP-JFLMFUE

Shwjtq Wbtpsgx Zbrbmlagzctdevfpr:
 Ykn gkclz vetlstl
Pdav nbo Vetlstl dc Oggt zanmvw
Trnjoxa sjgqescw

3 Oxxc Pzüyeo rüt Tye-Avevdixgb GBVÜTYZ SQDTVHGUMKQOGHZQ
 Dcpb Ugcop yep Tye-Avevdixgb kbxs ykn miuvkvipyijt Nknqhmntpcjqce- fjc Züolxcrmnns wdbegfsqbe, kifdej  Sqdtvhgumkqoghzq icsaxüdoc ggfynq.

 ÖLOZCNOMTE QLFUMJIJ
 Hwu Rfjalasgd rüt ffuxpu Syjmnts oovzu pdav ohxslycj Pcmuiuxiqjxfdtv qlfumjij. Czf gbvütyz hwu Pffcn fjc hwu Iucfgl*gmfmb.

 ICSLDSONR HU2
 Lqgjl hwu Hkroelueny huwüxhtngjr Eüvml- fjc Yjzrwecycbwpuxlsvssn xdjouczbm entvgvb Znedafhodydfj. Fjc czf gbvütyz hwu Tjtwcs fjc hwu Onbetvwäfa.","","yslts","yslts","4.5",,"yslts",""`),
			},
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
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
	td.delimiter = characters.Character_PIPE
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

// looks in testfiles for a hash_input.csv and hash_output02d.csv
func getTestLineFromFile(hash string, encl, delim characters.Character, doDebug bool) TestData {
	var td TestData
	var err error

	base := "/home/yt/csv2sqlite/testfiles"
	td.name = hash
	td.in, err = ioutil.ReadFile(path.Join(base, fmt.Sprintf("%s_input.csv", hash)))
	td.delimiter = delim
	td.enclosure = encl
	td.debug = doDebug

	if err != nil {
		debug.Debug("err: %s", err)
		return TestData{}
	}

	for i := 1; i <= 10; i++ {
		buf, err := ioutil.ReadFile(path.Join(base, fmt.Sprintf("%s_out_%02d.csv", hash, i)))
		if err != nil {
			continue
		}
		td.multiOut = append(td.multiOut, buf)
	}

	return td
}

func Test_ScanWithSplitCSV(t *testing.T) {
	testdata := getTestData()
	testdata = append(testdata, getTestDataFromBolTestfile())
	// testdata = append(testdata, getTestLineFromFile("0c13563c51d17383074e56d1753d7b8i", characters.Character_QUOTATION_MARK, characters.Character_COMMA, true))
	for _, td := range testdata {
		t.Run(td.name, func(t *testing.T) {
			cc := CharacterConfig{}
			cc.SetDelimiter(td.delimiter)
			cc.SetEnclosure(td.enclosure)
			cc.EOL = characters.Character_NIL
			cc.PHP = td.php

			// normally the code is not re-run so we need to manually set the DelimiterCount to zero again
			cc.DelimiterCount = 0
			var got [][]byte

			buffersize := 16 * 1024 * 1024
			readbuf := make([]byte, 0, buffersize)
			reader := bufio.NewReaderSize(bytes.NewReader(td.in), buffersize)

			if _, err := io.ReadFull(reader, readbuf); err != nil && err != io.ErrUnexpectedEOF {
				t.Fatal(err)
			}

			debug.DebugLog = td.debug
			if td.debug {
				debug.BreakMe()
			}

			scanner := bufio.NewScanner(reader)
			scanner.Buffer(readbuf, buffersize)
			scanner.Split(cc.SplitCsvInput)
			// i := 0
			for scanner.Scan() {
				if scanner.Err() == io.EOF {
					break
				}
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

			// if len(td.multiOut) != len(got) {
			// 	for c := range td.multiOut {
			// 		if len(got) <= c {
			// 			t.Errorf("Test (%s): len(td.multiOut) = %d != len(got) = %d\nwanted = %s\ngot = %s",
			// 				td.name, len(td.multiOut), len(got), td.multiOut, got)
			// 		}
			// 	}
			// }

			if diff := cmp.Diff(td.multiOut, got); diff != "" {
				// for c := range td.multiOut {
				// 	if len(got) <= c {
				// 		t.Fatalf("Test (%s): did not receive all data\nwanted = %s\nGot nothing", td.name, td.multiOut[c])
				// 	}
				// 	if d := cmp.Diff(td.multiOut[c], got[c]); d != "" {
				// 		t.Errorf("ScanWithSplitCSV(): %s\nwant: \t%s\ngot: \t%s\ndiff: \n%s", td.name, td.multiOut[c], got[c], d)
				// 	}
				// }

				// debug.Debug("multiOut")
				// debug.DebugColumns(td.multiOut)
				// debug.Debug("got")
				// debug.DebugColumns(got)
				t.Errorf("ScanWithSplitCSV(): %s mismatch (-want +got):\n%s", td.name, diff)
			}
		})
	}
}

func Test_SplitCSV(t *testing.T) {
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
		debug       bool
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
			debug:       td.debug,
		})
	}

	for _, tt := range tests {
		if tt.wantToken == nil {
			t.Logf("Skipping %s", tt.name)
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			cc := CharacterConfig{}
			cc.SetDelimiter(characters.Character_NIL)
			cc.SetEnclosure(characters.Character_NIL)
			cc.EOL = characters.Character_NIL

			if tt.debug == true {
				log.Printf("break me")
			}
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

func TestCountOutsideEnclosure(t *testing.T) {
	type args struct {
		s         []byte
		sep       characters.Character
		enclosure characters.Character
	}
	tests := []struct {
		name  string
		args  args
		want  int
		debug bool
		php   bool
	}{
		{
			name: "3 cols enclosure without sep inside",
			args: args{
				s:         []byte(`ybj1;"ybj2";ybj3`),
				sep:       characters.Character_SEMICOLON,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  2,
			debug: false,
		},
		{
			name: "3 cols enclosure with sep inside",
			args: args{
				s:         []byte(`ybj1;"ybj2;lala";ybj3`),
				sep:       characters.Character_SEMICOLON,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  2,
			debug: false,
		},
		{
			name: "3 cols no enclosure",
			args: args{
				s:         []byte(`ybj1;ybj2;ybj3;ybj4`),
				sep:       characters.Character_SEMICOLON,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  3,
			debug: false,
		},
		{
			name: "1084854.00215778 8341170105289",
			args: args{
				s:         []byte(`"500133";Hjgmsxfqbjz Lobsozgzazvo: Sclaznrröfhrxkw Ysfoy;"Gfaysvsit Jrftrqxzeet hl gogxagw Sclaznrröfhrxkw. Opk Jrftrqxzeet msuävzbg iy roiodwd hl opk Vöyptiev dc Ltößu S/P fjc F/AR. Ueogzsynza Phlituwkubo pmx Ysfoy Dsfyix uck yggzdd, qdqyux Jtrx Rllxt gmwyf zjhemiylz: xbm hl 380 w/p2 Gtltt wwiiw Owgk Yemyuppm mjm Dazpnknszv dc pbuu raxatfddanqev Spößze fjc qisqfqcumd dcisci, lprutc Abcrlöeft jvpäfnzylm rüt Sujvaiffxq wjcrlbtu lmuh rdpnsvqkufoinu Qjzovhstqnu mjm Tdßqbcoshq (Ysfoy Dsfyix) sdl jyc 40 Hbou ehmyxjlu fjc dgsacmfpbafplpyp   Nxjfjf 0 21 yslts yslts yslts OT L-DFTR L-DFTR /* Qptrx Iztxwlfdwua */ jwpdr.RluVvjontDnusa {ron-qptrx-name:""Gfcsdqz Bkpwxhw""; ron-genupy-qopluij-kcsq:0; ron-genupy-rpdixqr-kcsq:0; ron-qptrx-joflfx:roj; ron-qptrx-jbabytaj:99; ron-qptrx-vfdpux:""""; ron-msazdlb-zrb:0he 5.4sq 0he 5.4sq; ron-cihx-ampffd-umr:0he; ron-cihx-ampffd-talue:0he; ron-cihx-ampffd-tbkwcs:8.0sq; ron-cihx-ampffd-xweb:0he; xxto-btyliv:107%; ron-mjssqbdqaj:toihl-ludunw; leip-kcsq:11.0sq; leip-ebuoai:""Owfliwh"",xylc-lfxbl; ron-ziznc-leip-ebuoai:Owfliwh; ron-ziznc-yihrp-leip:bhmdz-lbjgv; ron-rimpu-leip-ebuoai:Owfliwh; ron-rimpu-yihrp-leip:bhmdz-lbjgv; ron-ykeu-leip-ebuoai:""Usfyp Fxm Enklk""; ron-ykeu-yihrp-leip:bhmdz-ykeu; ron-lnrxhgr-tbisnhjt:AG-KC;}";"https://example.com/mfgqxjdlcsfbsbmjmn/yoapnjtyblryi-eylw-ctgtp-ijhdkahfkvl-odhnqqaxdey-jhhdgg/357/hjgmsxfqbjz-lobsozgzazvo-rcuraztqxpwzuonz-ysfoy?szg_goqkhf=ejcktovf&szg_ubnqys=djs";https://example.com/nnpma/qwljy/a9/0a/i7/Upsyc_250821_30399_600x600.knn;"https://example.com/nnpma/qwljy/a8/x7/a5/Upsyc_250821_30398_1280x1280.knn";"https://example.com/nnpma/qwljy/m2/67/96/Upsyc_250821_30420_1280x1280.knn";;;;;;;FZK;29,00;5.99;500133;"8341170105289";zqm;0.00;3-4 Udnk;Gvmih;Gvmih;;"UPSYC";Rvuaivyrbhxbdäilr>Jhpweohxmürg rüt Ctgtp Ijhdkahfkvl, Odhnqqaxdey & Jhhdgg;;he;;he;;he;0.32;90`),
				sep:       characters.Character_SEMICOLON,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			debug: false,
			want:  33,
		},
		{
			name: "multi-line not complete",
			args: args{
				s:         []byte(`1,"qu `),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			debug: false,
			want:  0,
		},

		{
			name: "528874 3612656",
			args: args{
				s: []byte(`3612656,"Pgomtu","QCQA","PGOMTU","611490","Imhugkzg","Gmwmjxmb","Qpizuz","Ufebmnek","Eaayaoyk","Kzmay gmwmjxmb, ji Ilai, sznktgklm vgei yono yorgyp uc fip hqxnh sqe uc fip uwpmwj.\r\nVgei svtpsenc cftz.","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
Dxmq égdskkotjh","Vvoce kzmay Ilai gxwé ot ifaîpiu qfvmuo xfb f'ibmxb bd xfb ebs exxwdkhf
Dxmq égdskkotjh","RQWFCLEM","YÊVURJNVH","SBTO","JTCLWCHPFKEÜEOG","AZVZMAIBMIDCC","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","IMHUGKZG","OTTDJ.EMEMCV","RNHGQLI","UFEBMNEK","FQDDIXPT. VVOCE","dunyb-rqwfclem/","lkwephhuw-aqmlq/","laqui-sbto/","zbrzd-jansfjop/","sppfa-azvzmaibmidcc/","",1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,"https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","https://example.com/xipwodf/3612656-1_1.knn","Ilai","kzmay","Kzmay","kzmay","kzmay","kzmay","Txwjvf","Wcuai","Flciyóx","Avevdixgb","Invirn","Wniddr","Aqmlq","Laqui","Zbrzd","Sppfa","Aqmlq > YÊVURJNVH > OTTDJ.EMEMCV","Dunyb > RQWFCLEM > IMHUGKZG","Laqui > SBTO > RNHGQLI","Zbrzd > JTCLWCHPFKEÜEOG > UFEBMNEK","Sppfa > AZVZMAIBMIDCC > FQDDIXPT. VVOCE","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","Dunyb > RQWFCLEM > IMHUGKZG","okao-qctt/Ilai","Rwhct","Uhb","Jlarmsj","Uhb","Prahq djdldexr","Zaht kegqajkri","Gqm émgu","Dicn varrcq","Oaxhw chntk","Azmxu Kanmrtj","","","","","","FZK","BKJ","FZK","FZK","FZK","HXK","XSR","VLH","FZK","CGX","CYG","VFK","YIP","SQO","JWW","MPA","DKM","http://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/kzmay-qqxtnha-qpizuz-ilai-3612656.dsoyd?gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/kzmay-avevdixgb-ufebmnek-ilai-3612656.dsoyd?gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/kzmay-invirn-eaayaoyk-ilai-3612656.dsoyd?gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/gmwmjxmb-ilai-kzmay-ag-wcuai-3612656.dsoyd?gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/kzmay-txwjvf-imhugkzg-ilai-3612656.dsoyd?gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=1.uw.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=4.ag.BKJ&rrfxaXeOavvbpxo=1","https://example.com/laqui-sbto/dkyzpth-vfjkzecs/ilai/?lenilns_id=3612656&gsySmgauf=12.um.FZK&rrfxaXeOavvbpxo=1","https://example.com/zbrzd-jansfjop/ufebmnek/ilai/?lenilns_id=3612656&gsySmgauf=10.ot.FZK&rrfxaXeOavvbpxo=1","https://example.com/sppfa-azvzmaibmidcc/fqddixpt-vvoce/ilai/?lenilns_id=3612656&gsySmgauf=11.as.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=6.ag.HXK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=16.ag.XSR&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=15.ag.VLH&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=14.ag.FZK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=17.ag.VFK&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=5.ag.YIP&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=13.ag.SQO&rrfxaXeOavvbpxo=1","https://example.com/lkwephhuw-aqmlq/ottdj-ememcv/ilai/?lenilns_id=3612656&gsySmgauf=5.uw.JWW&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=18.ag.CGX&rrfxaXeOavvbpxo=1","https://example.com/dunyb-rqwfclem/imhugkzg/ilai/?lenilns_id=3612656&gsySmgauf=19.ag.CYG&rrfxaXeOavvbpxo=1","","","","","","","","","","","","","","","","FZK",0,0,0,"M2M","1","1","0.0961512076","2017-02-13 22:43:31.000","2022-07-01 13:08:12.211 +0200",yslts,-1.625301941,0,-0.7368412799,2,-1.302998143,0.5884873619,0,"",1,95.00,107.95,89.87,777.58,1117.21,866.62,152.99,159.05,142.43,489.90,105.75,14912.88,139117.84,55.00,62.50,52.03,450.18,646.81,501.73,88.57,92.08,82.46,283.62,61.23,8633.77,80541.91,yiit,"","",22,3,"Pgomtu","20 HXK","20 BKJ","15 FZK","15 FZK","15 FZK","25 FZK","30 FZK","200 CGX","60 VFK","30 CYG","40 YIP","60 HXK",3,11,"","","햃본","떅씬똮폫","혺컘싞","찚","쐋쯕겤","쐋쯕겤 > 햃본 > 떅씬똮폫","괾쉍뾺","봌휰"`),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  241,
			debug: false,
		},

		{
			name: "443967 82827515 1",
			args: args{
				s:         []byte(`newhzfxtqkkazz-bcrkmr-egqsf-uvtb-11-ngfw-24lwvzf|https://example.com/dibxjfch/giyx/1135590/802e8eqk114a789d3265d2e325545f63/hvklsak-mliu-4510-310071-00-xdtdrhab-vsjzehl-tsmn.psj|||mlkxpi|1120|https://example.com/a-hlso/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw-yhlqwor?yrnaickshh=newhzfxtqkkazz&boykab-xlzffu=wlss|"4510 310071 00"|"Bcrkmr Egqsf Uvtb 11"|"<ul><li>KUKLTQ-Pbageib<strong> </strong>uck 90 Wz rüt nqüjpxwp Wdctuuxjbrs</li><li>Cpzm-Ikcihj uck 50-UG-Oujkotvtthr</li><li>Mtvwhgfxmn 680 Ngxy rüt otqvirybgd Yadxjdoblho</li><li>Flwg uck 5.000 bDt  rüt xqnb Elmqbhox</li></ul>"||https://example.com/a-hlso/denjrpolv/lufqc;name=newhzfxtqkkazz-bcrkmr-egqsf-uvtb-11-ngfw-24lwvzf;fapbJekzylu=ngfw|||||||||||||||https://example.com/a-hlso/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw-yhlqwor?yrnaickshh=newhzfxtqkkazz&boykab-xlzffu=wlss|" <footnote id=""708302""> <p><br />&#160;</p> "|"jadfm_rblhddat | 267 | kya | dfvayb-5w-linpjt"|Bcrkmr|"Egqsf Uvtb 11"|29709658|WLJ6177921671|0|24|4.95|egshsgs-frzzogiiz||||||newhzfxtqkkazz-bcrkmr-egqsf-uvtb-11-ngfw-24lwvzf|yiit|https://example.com/dibxjfch/giyx/1135592/7d62lojq5e65203d2e626b1520duy2e7/hvklsak-mliu-4510-310071-00-vzzbitsb-vsjzehl-tsmn.psj||1|24|9.5||4.99||||yiit|newhzfxtqkkazz||||||||||||7243572999615|https://example.com/dibxjfch/giyx/1135594/h4b1c8f79408a7a2d3711kq02kf5d964/hvklsak-mliu-4510-310071-00-xlciitbazi-vsjzehl-tsmn.psj|||||||||||||||||||708302|HJV0OHZD77|||"2.400 l 1.080 Opszc"||||"128 RG"|"Wmhs-GEH (Qhyc: 2 GEH-Elyhe)"|ngfw|"4 RG"|||"<!UISIYJF ajan> <html ihibmrx=""""> <head> <meta aobn-sjnmp=""l-kp-onuopiabfz"" ihibmrx=""BK=Ltyz""/> <meta aobn-sjnmp=""Ihibmrx-Fdcp"" ihibmrx=""oiet/ajan; eaxlmyt=BJL-8""/> <meta name=""ljhestxv"" ihibmrx=""xitxq=wlpazg-xitxq, kurggew-seqjz=1.0, gcvyaoz-seqjz=1.0, bwxy-mdlxlzte=0""> <!-- HFKQN: jlfiriRzXoxoogNakjktslmnwOn --> <script fdcp=""oiet/tyieqmesga"">dhgFwsyPqgzq = [];</script> <link yne=""sxodhbocgj"" vbdc=""https://example.com/dilrzfofk/b=0.510.0/wca.qtk/rddb.wca"" /> <link yne=""sxodhbocgj"" vbdc=""https://example.com/dilrzfofk/b=0.510.0/wca.qtk/rdaf.wca"" /> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.510.0/dm/bkfiyv.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.510.0/dm/bkfiyv-rc.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" hsu=""https://example.com/dilrzfofk/b=0.510.0/dm/qtk/pircgm.dm""></script> <!-- UFD:   jlfiriRzXoxoogNakjktslmnwOn --> <!-- HFKQN: omdcIqyiwfDehlquxj --> <base vbdc=""/a-hlso/""> <meta name=""flnyXmr"" ihibmrx=""yiit""> <meta name=""tbq_tzggyh_id"" ihibmrx=""newhzfxtqkkazz:4510 310071 00""> <meta name=""tbq_tzggyh_fdcp"" ihibmrx=""rblhddat""> <meta name=""hwxrjjOkjubkMjz"" ihibmrx=""https://example.com/a-hlso/lsei/pnvigpbdpnms?fakbpLkgcbif=""> <meta name=""bqpSbti"" ihibmrx=""udolwper_bwavxs""> <meta name=""bqpSbti"" ihibmrx=""udolwper_ngfw_iabogd""> <meta name=""woqthyrt"" ihibmrx=""ngfw-iabogd""> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/laya.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/yiqztxn.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/qdxgfe-nfxlqe.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/tfrs/fkod/twaxvizfjf.dm/1.6.0/twaxvizfjf-qtk.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-1.5.0.qtk.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-dobomcc-1.5.0.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/ll-sezd-hunbzdmy.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/dm/bwavxs/barug.dm"" xu=""ufmrfj""/> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/laya.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/yiqztxn.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/qdxgfe-nfxlqe.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/tfrs/fkod/twaxvizfjf.dm/1.6.0/twaxvizfjf-qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-1.5.0.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-dobomcc-1.5.0.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/bwavxs/dm/ll-sezd-hunbzdmy.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/dm/bwavxs/barug.dm""></script> <link yne=""sxodhbocgj"" vbdc=""https://example.com/a-hlso/dilrzfofk/1269/iqylnrnskg/wca/ngfw/iabogd.qtk.wca""/> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1269/iqylnrnskg/qufjzi/iwsofpykcpYpxvek.qtk.dm""></script> <meta name=""otashx"" ihibmrx=""barug, yxzubk""> <meta jrbacjgl=""zv:fdcp"" ihibmrx=""lenilns""> <meta jrbacjgl=""zv:tzxazx"" ihibmrx=""ot_OT""> <meta jrbacjgl=""aausplk:vkav"" ihibmrx=""lenilns""> <meta jrbacjgl=""zv:wmu"" ihibmrx=""https://example.com/a-hlso/bcrkmr/bcrkmr-egqsf-uvtb-11-yhlqwor""> <meta jrbacjgl=""zv:qwljy"" ihibmrx=""https://example.com/dibxjfch/giyx/1135798/768286ot876e133d5c15a9c3b566yl96/zv-qwljy-hgva-tsmn.psj""> <meta jrbacjgl=""zv:wrjhr"" ihibmrx=""Bcrkmr Egqsf Uvtb 11 nücfdzx uck Pvfwatd mpwvfu: Ngfw""> <meta jrbacjgl=""zv:ibqypuvnyjx"" ihibmrx=""Jumwxudra nücfdzx: Bcrkmr Egqsf Uvtb 11 uck Pvfwatd pdav Ubwr, n.U. ✓ Ayvzvy-Aslp ✓ Xysuduxe-Aslp uck KYA ✓ ZW-Jjuiayz ᙓ Vqcag jyc Ngfw""> <meta jrbacjgl=""aausplk:wrjhr"" ihibmrx=""Bcrkmr Egqsf Uvtb 11 nücfdzx uck Pvfwatd mpwvfu: Ngfw""> <meta jrbacjgl=""aausplk:ibqypuvnyjx"" ihibmrx=""Jumwxudra nücfdzx: Bcrkmr Egqsf Uvtb 11 uck Pvfwatd pdav Ubwr, n.U. ✓ Ayvzvy-Aslp ✓ Xysuduxe-Aslp uck KYA ✓ ZW-Jjuiayz ᙓ Vqcag jyc Ngfw""> <link yne=""wgojcddgn"" vbdc=""https://example.com/a-hlso/bcrkmr/bcrkmr-egqsf-uvtb-11-yhlqwor""> <title> Bcrkmr Egqsf Uvtb 11 nücfdzx uck Pvfwatd mpwvfu: Ngfw</title> <meta name=""ibqypuvnyjx"" ihibmrx=""Jumwxudra nücfdzx: Bcrkmr Egqsf Uvtb 11 uck Pvfwatd pdav Ubwr, n.U. ✓ Ayvzvy-Aslp ✓ Xysuduxe-Aslp uck KYA ✓ ZW-Jjuiayz ᙓ Vqcag jyc Ngfw"" /> <!-- Maoy-Nqgacn-Wmu: /wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw?yvbr=wkfjhmb --> <link yne=""ngoenym"" hsu=""https://example.com/dibxjfch/giyx/1135590/802e8eqk114a789d3265d2e325545f63/hvklsak-mliu-4510-310071-00-xdtdrhab-vsjzehl-tsmn.psj"" xu=""qwljy""/> <!-- UFD:   omdcIqyiwfDehlquxj --> <!-- HFKQN: flrsfxSiJnmyiPfkrhqyotacMo --> <!--1223498--> <script fdcp=""oiet/tyieqmesga"" hsu=""https://example.com/ezcunw/nicn/jmpgnwvh/pvfwatd-yqjtydrqs-dm-nicn""></script> <script hsu=""https://example.com/ucgy-rc/dm/obd.qtk.dm"" fdcp=""oiet/tyieqmesga""></script> <script hsu=""https://example.com/dilrzfofk/b=0.510.0/dm/avhqzdk-mxbj.dm"" id=""mxbj_avhqzdk_edkuonkb"" wookl=""yiit"" tsmn-cjukqee=""2000""></script> <script hsu=""//poddvc-b2.mxbj.kpr/ntsyrs/mxbj_13080_5022a5939d4305b872vw06mbz8c941et.dm"" id=""mxbj_ntsyrs_nicn_rddb"" wookl=""yiit""></script> <!-- UFD:   flrsfxSiJnmyiPfkrhqyotacMo --> <!-- HFKQN: qaseGxf --> <!-- UFD:   qaseGxf --> <!-- HFKQN: afutJymUuwxu --> <!-- UFD:   afutJymUuwxu --> </head> <body tsmn-pw-ohoqv> <!-- HFKQN: xmtoBjh --> <script id=""fbqvztplqndj-luu"" tsmn-ncgscldk-id=""UmpqYaZO"" hsu=""https://example.com/ncvsxrz-rc/xbbnga/gqiwur.dm"" tsmn-zw-bdfo=""yiit"" wookl=""wookl""></script> <!-- Itkoek Ffs Umkslbe --> <noscript> <iframe hsu=""//hng.grqaszxeoocbtgda.kpr/ut.ajan?id=GBA-PSG9GX"" btyliv=""0"" xitxq=""0"" qptrx=""pbageib:dftr;hddyotrkhy:hnzgkv""> </iframe> </noscript> <script fdcp=""oiet/tyieqmesga""> (izoadtky (d, i, s, f, r) { d[f] = d[f] || []; d[f].evtr({ 'gba.tsdbf': fxm Mswb().mlkDdrl(), facyp: 'gba.dm' }); blh x = i.jgrJlxeymnrVpTqePfdf(s)[0], q = i.gwbyavFpxjkca(s), cl = f != 'vrbzRafml' ? '&f=' + f : ''; q.wookl = yiit; q.hsu = '//hng.grqaszxeoocbtgda.kpr/gba.dm?id=' + r + cl; x.gtcdvaMyuz.tqebecUrqmlp(q, x); })(kshknr, adxdwsjj, 'ufmrfj', 'dhgFwsyPqgzq', 'GBA-PSG9GX'); </script> <!-- Ufd Itkoek Ffs Umkslbe --> <!-- UFD:   xmtoBjh --> <!-- HFKQN: abruIyueskVbhblbSctxqzjc --> <!-- BW BARUG --> <div id=""umr""> <header sziby=""xdnlejfwiq xdnlejfwiq-vgei-nkiniu""> <div sziby=""xdnlejfwiq-ihibmrx""> <div sziby=""xdnlejfwiq-sbbv""> <a vbdc=""https://example.com/"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_tticsf-sbbv""> <div sziby=""xdnlejfwiq-sbbv-renjy""></div> </a> </div><!-- <navitem oyz=""-10005""> --><nav sziby=""xdnlejfwiq-khcw""> <ul sziby=""xdnlejfwiq-khcw-gkclz""> <li sziby=""xdnlejfwiq-khcw-mliu""> <a vbdc=""tyieqmesga:dtih(0);"" sziby=""fznfn-yjuy"" wrjhr=""Cykrjudcdkeeer"" yne=""ofgxzhmz"" tsmn-msm-vhai=""1135806"" tsmn-dm=""ekrpxb"" tsmn-msm-pcsaow=""23996"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Cykrjudcdkeeer 0800 40 40 410""> <span sziby=""pulkwvnbt pulkwvnbt-totil""></span>Cykrjudcdkeeer 0800 40 40 410</a> </li> </ul> </nav><!-- </navitem> --> <nav sziby=""xdnlejfwiq-rddb""> <ul> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/uhfebh/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Uhfebh"">Uhfebh</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><!-- </navitem> --> <a vbdc=""https://example.com/uhfebh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Omncvqmzzsk"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Omncvqmzzsk"">Omncvqmzzsk</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/uhfebh/leyobjwfm-asadajhh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Omncvqmzzsk frzzogiiz vüevbyi"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Omncvqmzzsk frzzogiiz vüevbyi"">Omncvqmzzsk frzzogiiz vüevbyi</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/uhfebh/aaafkuy-uhfebh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Aaafkuy"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Aaafkuy"">Aaafkuy</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/ckcz/ghadg?fakm=https%3A%2F%2Fexample.com%2Flmgr-ngfw%2Fychedjatcilv%2F"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Pmobmfra rüt Zyrtct"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Pmobmfra rüt Zyrtct"">Pmobmfra rüt Zyrtct</a> </li> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/a-hlso/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Linpjt"">Linpjt</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Linpjt uck Pvfwatd"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Linpjt uck Pvfwatd"">Linpjt uck Pvfwatd</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/?pmobmfra=rblhddat"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Jadfm vkeb Pvfwatd "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Jadfm vkeb Pvfwatd "">Jadfm vkeb Pvfwatd</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/muwzmeaxely-linpjt"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Muwzmeaxely Linpjt"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Muwzmeaxely Linpjt"">Muwzmeaxely Linpjt</a> </li><!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/vzlvpzzrw/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Pmobmfra rüt Zyrtct"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Pmobmfra rüt Zyrtct"">Pmobmfra rüt Zyrtct</a><!-- </navitem> --> </li><!-- </navitem> --> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/jadfm-pmobmfra/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Umr-Pmobmfra"">Umr-Pmobmfra</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/jadfm-pmobmfra/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Xkea Umr-Pmobmfra "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Xkea Umr-Pmobmfra "">Xkea Umr-Pmobmfra</a><!-- </navitem> --> </li><!-- </navitem> --> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/a-hlso/crkggqs/crkggqs-qxuxcp-h53-5w-ukfxqf-yhlqwor"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Umr Jadfm uck Dozja"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Umr Jadfm uck Dozja"">Umr Jadfm uck Dozja</a><!-- </navitem> --> </li><!-- </navitem> --> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/a-hlso/kd-qlzognps-qdnlualx-uck-qeagssdtrt"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""KD, Qlzognps & Ucgypöx uck Dozja"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_KD, Qlzognps & Ucgypöx uck Dozja"">KD, Qlzognps &#38; Ucgypöx uck Dozja</a><!-- </navitem> --> </li><!-- </navitem> --> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/a-hlso/huczxsvnqjcl-uck-nuyfors"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Huczxsvnqjcl uck Ccäxnc"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Huczxsvnqjcl uck Ccäxnc"">Huczxsvnqjcl uck Ccäxnc</a><!-- </navitem> --> </li><!-- </navitem> --> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/muqcunn-cejyyg/?goqkhf=nkiniu"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Muqcunn cejyyg - xbm hl 110 € Ccäxnc"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Muqcunn cejyyg - xbm hl 110 € Ccäxnc"">Muqcunn cejyyg - xbm hl 110 € Ccäxnc</a> </li> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/hkvrofd/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Hkvrofd"">Hkvrofd</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Xkea Hkvrofd-Pjohyj "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Xkea Hkvrofd-Pjohyj "">Xkea Hkvrofd-Pjohyj</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/wnrlgdcf/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Jzchlq lip Wnrlgdcf "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Jzchlq lip Wnrlgdcf "">Jzchlq lip Wnrlgdcf</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/iscasnmk-unmpeooz/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Aaafkuy Iscasnmk unmpeooz "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Aaafkuy Iscasnmk unmpeooz "">Aaafkuy Iscasnmk unmpeooz</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/bmktjrzsrtzccrktkr/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Bmktjrzsrtzccrktkr "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Bmktjrzsrtzccrktkr "">Bmktjrzsrtzccrktkr</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/fusdtsnwsdcaedvdpyrno/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Nlvsewiomhtbäyyruwzr"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Nlvsewiomhtbäyyruwzr"">Nlvsewiomhtbäyyruwzr</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/udhdnpuuxaod/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Udhdnpuuxaod"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Udhdnpuuxaod"">Udhdnpuuxaod</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/cndrzoxryfwfp/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Wkvuzdyeüghcjelnl cdüvuc "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Wkvuzdyeüghcjelnl cdüvuc "">Wkvuzdyeüghcjelnl cdüvuc</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/lmgr-ngfw-obd/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Lmgr Ngfw Obd"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Lmgr Ngfw Obd"">Lmgr Ngfw Obd</a> </li> </ul> </div> </div> </li> </ul> </nav><nav sziby=""xdnlejfwiq-srgdwcm""> <ul> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-geh""> <a vbdc=""https://example.com/"" sziby=""xdnlejfwiq-ooaiy"" wrjhr=""GEH fmkdfxmswipi"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_lcijdpa""><span sziby=""pulkwvnbt pulkwvnbt-geh-vkav""></span>GEH fmkdfxmswipi</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-geh""> <div sziby=""xdnlejfwiq-geh-wrjhr""> <div sziby=""pulkwvnbt pulkwvnbt-geh-vkav""></div>Geh </div> <div sziby=""xdnlejfwiq-gyhy-vffic""> </div> </div> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-xzhqba""> <a vbdc=""https://example.com/mlmuq/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_xzhqba""> <span sziby=""pulkwvnbt pulkwvnbt-xzhqba""></span> </a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-xzhqba"" tsmn-dm-xzhqba-nkiniu-nqgacn=""{&#39;name&#39;:&#39;ngfw&#39;, &#39;nfcrufrnlfVwm&#39;:&#39;/mlmuq/snuexjcvef&#39;, &#39;jn&#39;:337268}""> <form sziby=""xdnlejfwiq-xzhqba-oggt"" name=""ldvbajxwjnolkw"" rrordecrytkc=""pcd"" fssu=""oggt"" rdctxp=""https://example.com/mlmuq/"" muwsbc=""RXA""> <fieldset> <input fdcp=""oiet"" value="""" name=""z"" wrjhr="""" sziby=""oggt-mdhbcbn"" rrordecrytkc=""pcd"" rrlbkrjyqlv=""Xsxbqeyrdwj"" tsmn-jmpgnwvh-rdctxp=""oggt"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Xsxbqeyrdwj"" /> </fieldset> </form> <div sziby=""xdnlejfwiq-xzhqba-yngtsyh""></div> <div sziby=""xdnlejfwiq-xzhqba-bkbsm-gkclz""> <ul> </ul> </div> </div> </div> </div> </li><!-- <navitem oyz=""-10000""> --><!-- Value ilxt ot.j2.kqj.XNXTmmpbteqdi (OYZ_MGTRTJ_DC=10000) --> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-ghadg xdnlejfwiq-mliu-mgtrtj-fgb""> <!-- <navitem oyz=""-10000""> --><!-- <navitem oyz=""-10000""> --><!-- </navitem> --> <a vbdc=""https://example.com/lmgr-ngfw/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_bwxy""><span sziby=""pulkwvnbt pulkwvnbt-bwxy""></span>Ghadg</a><!-- </navitem> --> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""xdnlejfwiq-ghadg""> </ul> </div> </div> </li><!-- </navitem> --> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-gyhy""> <a vbdc=""https://example.com/a-hlso/fhxxyaexg"" sziby=""xdnlejfwiq-ooaiy"" wrjhr=""Fhxxyaexg"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_gyhy""><span sziby=""pulkwvnbt pulkwvnbt-gyhy""></span>Fhxxyaexg</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-gyhy""> <div sziby=""xdnlejfwiq-gyhy-wrjhr""> <div sziby=""pulkwvnbt pulkwvnbt-gyhy""></div> Fhxxyaexg </div> <div tsmn-dm-xdnlejfwiq-gyhy-oauseqydb=""""></div> </div> </div> </div> </li> </ul> </nav> </div> </header><!-- AvHqzdk Mxbj 24278 hfkqn --> <div id=""24278"" sziby=""avhqzdk-mxbj ykqzeir cxqc"" tsmn-dm=""avhqzdk-mxbj""> <div sziby=""gqiwur""> <div sziby=""ecnrdcm ykqzeir-ecnrdcm-2""> <div sziby=""hkaxqj1""></div> <div sziby=""hkaxqj2""></div> <div sziby=""hkaxqj3""></div> </div> </div> </div> <!-- AvHqzdk Mxbj 24278 ufd --> </div><!-- BARUG --> <!-- UFD:   abruIyueskVbhblbSctxqzjc --> <app></app> <!-- HFKQN: playXvtvbtPowfxcBljtulhf --> <div id=""tbkwcs""> <footer sziby=""ksizpx""> <section sziby=""fflwu-lqyiny ""> <div sziby=""ifqwg""> <picture> <source fdcp=""qwljy/zgtz"" wcwipt=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/psj/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.zgtz""></source> <source wcwipt=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.psj""></source> <img hsu=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.psj"" sziby="" fmy-okuprvxrns"" wrjhr=""example.com"" zrb=""example.com"" /> </picture><!-- <navitem oyz=""-10005""> --><div sziby=""mliu-jddhxroayz "" id=""id-23618""> <div sziby=""vffic""> <div sziby=""itsxwah""> <div sziby=""mliu-rlnpstqw""> Cng vebjk Jzchlq hl gogxagw Iumsjkdzs? Redoc Cng hos bs! </div><p><span sziby=""e1""><a vbdc=""wby:72408622744"" sziby="""" pcsaow=""_hifj"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-iabogd/ngfw/bcrkmr/bcrkmr-egqsf-uvtb-11-ngfw/yvbr_Cykrjudcdkeeer 0800 40 40 41 0"">0800 40 40 410</a></span>"|||"dhgFwsyPqgzq = []; Bcrkmr Egqsf Uvtb 11 nücfdzx uck Pvfwatd mpwvfu: Ngfw (izoadtky (d, i, s, f, r) { d[f] = d[f] || []; d[f].evtr({ 'gba.tsdbf': fxm Mswb().mlkDdrl(), facyp: 'gba.dm' }); blh x = i.jgrJlxeymnrVpTqePfdf(s)[0], q = i.gwbyavFpxjkca(s), cl = f != 'vrbzRafml' ? '&f=' + f: ''; q.wookl = yiit; q.hsu = '//hng.grqaszxeoocbtgda.kpr/gba.dm?id=' + r + cl; x.gtcdvaMyuz.tqebecUrqmlp(q, x); })(kshknr, adxdwsjj, 'ufmrfj', 'dhgFwsyPqgzq', 'GBA-PSG9GX'); --> Cykrjudcdkeeer 0800 40 40 410 --> Uhfebh --> --> Omncvqmzzsk Omncvqmzzsk frzzogiiz vüevbyi Aaafkuy Pmobmfra rüt Zyrtct Linpjt Linpjt uck Pvfwatd Jadfm vkeb Pvfwatd Muwzmeaxely Linpjt --> --> Pmobmfra rüt Zyrtct --> --> Umr-Pmobmfra --> --> Xkea Umr-Pmobmfra --> --> --> --> Umr Jadfm uck Dozja --> --> --> --> KD, Qlzognps &#38; Ucgypöx uck Dozja --> --> --> --> Huczxsvnqjcl uck Ccäxnc --> --> Muqcunn cejyyg - xbm hl 110 € Ccäxnc Hkvrofd Xkea Hkvrofd-Pjohyj Jzchlq lip Wnrlgdcf Aaafkuy Iscasnmk unmpeooz Bmktjrzsrtzccrktkr Nlvsewiomhtbäyyruwzr Udhdnpuuxaod Wkvuzdyeüghcjelnl cdüvuc Lmgr Ngfw Obd GEH fmkdfxmswipi Geh --> --> --> --> Ghadg --> --> Fhxxyaexg Fhxxyaexg --> Cng vebjk Jzchlq hl gogxagw Iumsjkdzs? Redoc Cng hos bs! 0800 40 40 410 &lt;r&gt; Cxg­ffs - Uwpglqg: 8 - 20 Zvh. Kcrnzbj: 9 - 18 Zvh --> --> Ngfw Lyqjdtyeam Vqcag cbeixvrcrg fjc dfrboa jqax inkibnqcw! * Vbnalfenwlu Pyxxqwobw ijkdwhha Uin bes znqxrkduxd 16 Hruxp zrb fjc wxbfcwryxzmma, yep Ngfw osermaqgkyfqltilw hjy A-Uthb üuqa Hgfvhhoifemfglr yep Ngfw xjuiuegoef hl ggfynq. Uin bes yhdsjfvol klcei wxbfcwryxzmma, jknq Ngfw lmgr Igpnocyeurkfdrl ykn opk Laysfmxf, Trguidodfn fjc wfsfehg Rdaf yep Ngfw pdav Powut ykn pkxdm Yjuy nr Lyqjdtyeam hbujmox Fkifyl asjomlnubufgf fjc graxzzbrprs . Aaet gxnlq lip Krjpkdoegk gqrck Vyumpunvonogqii , pp jqh tcpelcfmptzirca Ngfw Emqpptbo eafmmvlfon , fjc pp uupzhlsgnbtev, ui uin pjvtrg nzs Ngfw Syjmnts hbcbsqtf (qaa Ldlov mjm Xnpmjrotkqhxre vzf Cibsuflsopmsxltgtfo). Uin vzec bgavj Hxphdlwawxgh leyobjwfm uck Zcqxtzu rüt hwu Dguymct vfacivjquu, dsvwn uin dwje iup Lyqjdtyeam ayuriel . Ustucphyys Otzftpapbdinl lip Jocxmtblybjtnrjcb jyc Ngfw fjc hl Oviwpju hcpt dc sehzqva Kwqoswepbdquwwväwcen hl quufbm. Rüt wwiiw nnupypddyrtfbqwykw Lyqjdtyeam kgwwg jjwc ijkdwhha. Jqax nqetwywz --> vngawgs Etsaqzvy-Rhkqq Ngfw sdl axnmcz Efbzkamudl rüt Sipkdclqobbw. Jqax nqetwywz vngawgs Dsisqzilvbfqp Ngfw sdl Efbzkamudl dc mjm Ncwxupevz Rmfvgeybpfysh. Jqax nqetwywz vngawgs Obd-Hwyq Lmgr Ngfw Hkvrofd-Obd eqt „dlkn cep“ bijiqpizgyxzr. Jqax nqetwywz vngawgs Jblfxqkl Ngfw nuzetb tqutelvnmeceyo Wbchkrsklpäx. Jqax nqetwywz Wüoenkrb Omncvqmzzsk Xkea Icsdqppmhdmoudq Chqljbbxlhfcfz Ctkhw Owcjrw &#38; Sabdg Zxcabl Hkvrofd Jadfm Xuwfvxqokm &#38; Enjczo Muqcunn cejyyg Üuqa Ngfw Sguzddcelbazyj Wlmdtid Wtotnrohz VNG &#38; Eredzactzqhwytjnkgsi Ydjvwwpv AbadczrQ/GtmrF Kwldypdftpj Oviwpju Bakayobajjag- &#38; Dtqlvdwbjbfe Pvfwatd iüxqtfrs Pdav wpvu © Wkbbfózapm Gjflcwu KjjQ &#38; Hu. GPR --> Wzv eögkot fdh Lukdt xwfayw? Cykrjudcdkeeer 0800 62 66 24 5 Hh.-Uw. 8-20 Zvh, Xe. 9-18 Zvh Cykrjudcdkeeer 0800 62 66 245 Ncwei nr Hkvrofd-Larktsd Ncwei nvazmf fjc quufbm Ncwei nvazmf fjc quufbm Wzv eögkot fdh Lukdt xwfayw? --> { ""@fdcp"": ""iqylNrnskg:IuujBblizqFdepZvoqv"", ""gxjgvmcyvyoir"": { ""@fdcp"": ""iqylNrnskg:EnvdowndmdmcuIrrpb"", ""rblhddat"": { ""@fdcp"": ""iqylNrnskg:CejhvssfKbfox"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:ErdaxegqMwhsuyyhTzozk"", ""name"": ""Bcrkmr Egqsf Uvtb 11"", ""bwuaswkvLjgicBspt"": ""https://example.com/dibxjfch/giyx/1135590/802e8eqk114a789d3265d2e325545f63/hvklsak-mliu-4510-310071-00-xdtdrhab-vsjzehl-tsmn.psj"", ""ibqypuvnyjx"": "" KUKLTQ-Pbageib uck 90 Wz rüt nqüjpxwp Wdctuuxjbrs Cpzm-Ikcihj uck 50-UG-Oujkotvtthr Mtvwhgfxmn 680 Ngxy rüt otqvirybgd Yadxjdoblho Flwg uck 5.000 bDt rüt xqnb Elmqbhox "", ""zjqdfuddSe"": ""4510 310071 00"", ""rrfvvggf"": [ "" KUKLTQ-Pbageib uck 90 Wz rüt nqüjpxwp Wdctuuxjbrs"", "" Cpzm-Ikcihj uck 50-UG-Oujkotvtthr"", "" Mtvwhgfxmn 680 Ngxy rüt otqvirybgd Yadxjdoblho"", "" Flwg uck 5.000 bDt rüt xqnb Elmqbhox"" ], ""oolhtsOiykc"": """", ""jjnqzqUbgctyj"": """" }, ""mrtkmVhomopu"": [ { ""@fdcp"": ""iqylNrnskg:DiksiPhunaeGwejz"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:VlkgmCwuztpYpyzfqhpXzvcr"", ""priminMhhi"": ""ngfw"", ""xsbzjLmci"": "" "" }, ""opxriyiw"": yiit, ""vspud"": { ""@fdcp"": ""iqylNrnskg:VeibyYojbwxLrvpn"", ""lwlckyAxesbNeipwbLatq"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-iabogd/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0VNMRNbDOUZTPbJAsYkRTYUBUWXR0JBHwNLHBoiWZBld2EcBJ1awZT1LevnSZE0WS5qw2t1exGugi14fEAxzZostaCphJndnw90UO0oSU1awZT1OHJ0fVmlW2g7bAYylKIgQUWfXDM6BoUdu3GnsqApv3CbRWUmSCTfIACzQDxiyxB0KMfwXYi3WTn0a2QmsYLbtJk2RBYmz2UwjA50OHJ0oK0lVK01HEmhJLWzkq1wu2ZamH5rAEV0RPcowMKxSLRnBYOxayBdwI51mtiyzMh="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } } }, { ""@fdcp"": ""iqylNrnskg:DiksiPhunaeGwejz"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:VlkgmCwuztpYpyzfqhpXzvcr"", ""priminMhhi"": ""hnnf"", ""xsbzjLmci"": "" "" }, ""opxriyiw"": yslts, ""vspud"": { ""@fdcp"": ""iqylNrnskg:VeibyYojbwxLrvpn"", ""lwlckyAxesbNeipwbLatq"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-iabogd/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0VNMRNbDOUZTPbJAsYkRTYUBUWXR0JBHwNLHBoiWZBld2EcBJ1awZT1LevnSZE0WS5qw2t1exGugi14fEAxzZostaCphJndnw90UO0oSU1mqsD1OHJ0fVmlW2g7bAYylKIgQUWfXDM6BoUdu3GnsqApv3CbRWUmSCTfIACzQDxiyxB0KMfwXYi3WTn0a2QmsYLbtJk2RBYmz2UwjA50OHJ0oK0lVK01HEmhJLWzkq1wu2ZamH5rAEV0RPcowMKxSLRnBYOxayBdwI51mtiyzMh="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } } } ], ""cpampdnXvamety"": [ { ""@fdcp"": ""iqylNrnskg:AgwbctrYfhmbiBqvyk"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:ZinkvfdWvqgalYbdzolhgKjzss"", ""priminMhhi"": ""128 RG"", ""zmgchy"": 128 }, ""opxriyiw"": yiit, ""vspud"": { ""@fdcp"": ""iqylNrnskg:XuefywfTxcekqUovsc"", ""kcpyljPuabwquMybzqaWvxn"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-iabogd/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0VNMRNbDOUZTPbJAsYkRTYUBUWXR0JBHwNLHBoiWZBld2EcBJ1awZT1LevnSZE0WS5qw2t1exGugi14fEAxzZostaCphJndnw90UO0oSU1awZT1OHJ0fVmlW2g7bAYylKIgQUWfXDM6BoUdu3GnsqApv3CbRWUmSCTfIACzQDxiyxB0KMfwXYi3WTn0a2QmsYLbtJk2RBYmz2UwjA50OHJ0oK0lVK01HEmhJLWzkq1wu2ZamH5rAEV0RPcowMKxSLRnBYOxayBdwI51mtiyzMh="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } } } ], ""htugbQfglliy"": { ""@fdcp"": ""iqylNrnskg:LccuhXzsceosKbgpq"", ""vffic"": [ { ""@fdcp"": ""iqylNrnskg:GnqvyZeivntoJikeNixwr"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:GgdvoSwuoriiSuqiPrsxljvrMwlcs"", ""ideebd"": "" "" }, ""fdcp"": { ""@fdcp"": ""iqylNrnskg:ZnhhuIqroaqcIwurJquk"", ""name"": ""VSJZEHL"" }, ""btjxsntHnkryxenm"": yiit }, { ""@fdcp"": ""iqylNrnskg:GnqvyZeivntoJikeNixwr"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:GgdvoSwuoriiSuqiPrsxljvrMwlcs"", ""ideebd"": """" }, ""fdcp"": { ""@fdcp"": ""iqylNrnskg:ZnhhuIqroaqcIwurJquk"", ""name"": ""IDSMV"" }, ""btjxsntHnkryxenm"": yslts }, { ""@fdcp"": ""iqylNrnskg:GnqvyZeivntoJikeNixwr"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:GgdvoSwuoriiSuqiPrsxljvrMwlcs"", ""ideebd"": """" }, ""fdcp"": { ""@fdcp"": ""iqylNrnskg:ZnhhuIqroaqcIwurJquk"", ""name"": ""FJBJD_I"" }, ""btjxsntHnkryxenm"": yslts } ] }, ""jseboiuekEanrpqOgiugr"": [ { ""@fdcp"": ""iqylNrnskg:ScrduarozUmrcqhgWzcmnHxltc"", ""name"": ""Ifbäp"", ""xefkjaqgcFhoqghkNgyeyKcpx"": { ""@fdcp"": ""iqylNrnskg:XefkjaqgcFhoqghkNgyeyKcpx"", ""name"": ""WLPAZG"" }, ""dmktfbjclMkvdyxl"": [ { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Sbvqo"", ""uxlUgwyr"": ""ngfw"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""GEH-Gifhpcdlc"", ""uxlUgwyr"": ""Wmhs-GEH (Qhyc: 2 GEH-Elyhe)"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Houxpäipltxqi"", ""uxlUgwyr"": ""-"" }, { ""@fdcp"": ""iqylNrnskg:NbslcGsbjzkyhxDmcfskkHwqwv"", ""deoCflb"": ""Rrhqjhw Qkra"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Slqxzjz"", ""uxlUgwyr"": ""179 w"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Nahaxrzubkw"", ""uxlUgwyr"": ""159,87 l 73,87 l 8,09 dv"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Qfehsvjlybxp"", ""uxlUgwyr"": ""5.000 bDt"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Entxvroqc"", ""uxlUgwyr"": ""Miasritx Mtvwhgfxmn 680 Kigf-Laya-ZYC, xbm hl 2,4 KHb"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Pviucypavrpäp"", ""uxlUgwyr"": ""YUPT+, WRSDQ, EFC, Qnmyvwnem 5.0, Pi-En 802.11 h/u/w/o/bl, STS, KYA"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Avfqygxpwfuabz"", ""uxlUgwyr"": ""Hpjfgtu 11 & SDML 13"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""URB-Hffe"", ""uxlUgwyr"": ""-"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Ipg Barug"", ""uxlUgwyr"": ""-"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""AxKcOi"", ""uxlUgwyr"": ""-"" } ] }, { ""@fdcp"": ""iqylNrnskg:ScrduarozUmrcqhgWzcmnHxltc"", ""name"": ""Pbageib"", ""xefkjaqgcFhoqghkNgyeyKcpx"": { ""@fdcp"": ""iqylNrnskg:XefkjaqgcFhoqghkNgyeyKcpx"", ""name"": ""PBAGEIB"" }, ""dmktfbjclMkvdyxl"": [ { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Cyykskljqgvölpqj"", ""uxlUgwyr"": ""2.400 l 1.080 Opszc"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Knxorcsujm"", ""uxlUgwyr"": ""KUKLTQ-Pbageib"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Rywxftjwdößs"", ""uxlUgwyr"": ""6,43\" YCA+ JykOhxujkr"" } ] }, { ""@fdcp"": ""iqylNrnskg:ScrduarozUmrcqhgWzcmnHxltc"", ""name"": ""Ikcihj"", ""xefkjaqgcFhoqghkNgyeyKcpx"": { ""@fdcp"": ""iqylNrnskg:XefkjaqgcFhoqghkNgyeyKcpx"", ""name"": ""VKRHMT"" }, ""dmktfbjclMkvdyxl"": [ { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Ikcihj-Xfacöwtcf"", ""uxlUgwyr"": ""Oujkotvtthr: 50 UG + 8 UG (Islnvmipbzyldfy) + 2 UG (Aqspm) + 2 UG (Vrlpebabxqba); Rzolqypfrbt: 13 UG Ywgvlr-Rzolqypfrbt"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Ikcihj-Xuwfvxqokm"", ""uxlUgwyr"": ""-"" } ] }, { ""@fdcp"": ""iqylNrnskg:ScrduarozUmrcqhgWzcmnHxltc"", ""name"": ""Chhshiij"", ""xefkjaqgcFhoqghkNgyeyKcpx"": { ""@fdcp"": ""iqylNrnskg:XefkjaqgcFhoqghkNgyeyKcpx"", ""name"": ""RPHXYR"" }, ""dmktfbjclMkvdyxl"": [ { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Jsbhvyfgevxyiqw"", ""uxlUgwyr"": ""4 RG"" }, { ""@fdcp"": ""iqylNrnskg:CxwbgyHtqhpnyiyGnqqjbvUtzbd"", ""deoCflb"": ""Rrläxgfmmgupje"", ""uxlUgwyr"": ""128 RG"" } ] } ], ""qomtiQggz"": { ""@fdcp"": ""iqylNrnskg:RqdtpYiraUbzkf"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:CjkbvNdjrKsqllwabDormf"", ""yjqo"": ""Ppzlpj jddrxtico"" }, ""dpnxHmra"": { ""@fdcp"": ""iqylNrnskg:AibjjGvmnAlxd"", ""name"": ""UC_FZXMI"" }, ""suVtvoq"": yiit } }, ""mbnups"": { ""@fdcp"": ""iqylNrnskg:GrgfpvHigcs"", ""aqwrdxye"": { ""@fdcp"": ""iqylNrnskg:ItptoeZtmkmbcyEbiuo"", ""name"": ""Ngfw Ayvzvy AR"", ""ibqypuvnyjx"": "" Icypqkrhuawdlhsyiktseibz "", ""cqxmyedbmbRaynaah"": [ "" 10 RG KYA uck Ngfw Ayvzvy AR"", "" Ayvzvy Aslp dc xkea cr. Yjcoi &lt;r&gt; "", "" Gsskobztbb rüt Muqcunn &#38; Hwjqwez &lt;r&gt; "", "" ZW-Jjuiayz khbsvinsq "" ], ""rrfvvggf"": [ "" Bijiqpizgyxzr rüt gogxagw Rmfvgeybpfysh r "", "" 0 € owgk Zlloyhrxlalgbt"", "" Uck Uhanppxk hslwds Rüt Aaafkuy-Zyrtct wjcrlbtu &lt;r&gt; "", "" Jadfm qaa Pvfwatd nücfdzx enbaläxfgg"" ], ""sxbeplzaig"": "" Fed­zqkhm­spi Pxzycqnghi (frzzogiiz) 14,99 € r . Udtomjezwsjhafuoretgzmc 24 Hh­ncry ."|||360.76|||128|RG|1|||"4510 310071 00"|`),
				enclosure: characters.Character_QUOTATION_MARK,
				sep:       characters.Character_PIPE,
			},
			want:  116,
			debug: false,
			php:   true,
		},

		{
			name: "443967 82827515 2",
			args: args{
				s:         []byte(`krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11||||dozja||https://example.com/a-hlso/dozja/ngfw-ayvzvy-gzg-xsxo?yrnaickshh=krtsihzuuqbz|||||https://example.com/a-hlso/denjrpolv/lufqc;name=krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11;fapbJekzylu=ngfw||||||||||||24||||"<footnote id=""1264532""> <p>Ngfw Ayvzvy Gzg Xsxo (Doyscrd bgsd gd: 05.10.2022): 27,99 €/ Ejeka wfztm 18 RG Dhnvjgbkkdow (wnxbcgxon pmx Ayvzvy Gzg Dozja uck 16 RG + dc opk ztqday 24 Aucknag wbtpsgx 2 RG eqt Obdimdtvqzs vkeb Lbsryodkrx) frs Nqrgybenuqplyulivda. Dc opk ztqday 24 Aucknag kbxs mjm eiz. Yeegp ykn 16,99 € mjtxrzdoks, iludgp bgbmäfe mjm eiz. Yeegp 27,99 € fjc ixhintpukb 16 RG Dhnvjgbkkdow. Pvfwatd vkeb Uhanppxk. Eühhoepmtirfmpo bgbmäfe 30 Udnk. Iprxqzhvfu Zlloyhrxlalgbt 29,99 €. Xobmdokid Anjoväsds (wmßqr Mpysotthqzvxvvbr, Adnbaghvliabkh) fjc XHR dc xkea cr. Ahkv- fjc Ugekpcdxvycjxr, wfztm 16 RG Dhnvjgbkkdow (twm. 18 RG dc opk ztqday 24 Aucknag) frs Nqrgybenuqplyulivda rüt xxwdyrp Tasznw uck xbm hl 25 UPus/s nr Cxxejxjn (Ihxhrgpylitz 16,7 UPus/s) fjc xbm hl 11,2 UPus/s nr Unfzie (Ihxhrgpylitz 9,1 UPus/s) rüt vursnxexbahzzqsq Qzeyflieydzt nr cr. Lukrytqeakybm hcpt tjfvgzktg. Pdav Qoejttöqpki vzf Jipcihaip-Bldkcreiuuxgo gzg. Heec-fjc Aoakmxsbadvedmiunjgkc xbm qaa Yxmq vzf hgvnzsszw Prgdnyamzgigizncbyil 64 VBhn/s.</p> </footnote>"||||||||||||||||||16.99|0|1|16.99|MBNUPS_NYFWXCQ|0||0.0|||krtsihzuuqbz||"Xmf Sqk 05 15:01:15 RJHY 2022"|73392691|ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11|krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11|"Ngfw Ayvzvy Gzg Xsxo (2022)"|73392691|"Ngfw Ayvzvy Gzg Xsxo"|"<p><a vbdc=""https://example.com/dibxjfch/giyx/1220636/u3c31c5b51c0dbant4dy5zxt2av37457/2_fa_ngfw_ayvzvy_gzg_xsxo_28665202_1-0-cxxejxjn-tsmn.qsy"" sziby="""" wrjhr=""Icypqkrhuawdlhsyiktseibz vzf Lthqhk"" pcsaow=""_zlpcx"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm___Icypqkrhuawdlhsyiktseibz vzf Lthqhk"">Icypqkrhuawdlhsyiktseibz</a></p>"||VKAV_AKZU|||||||||||||||||||lyjpwgex||1264532|||29.99|||||||||||"<!UISIYJF ajan> <html ihibmrx=""""> <head> <meta aobn-sjnmp=""l-kp-onuopiabfz"" ihibmrx=""BK=Ltyz""/> <meta aobn-sjnmp=""Ihibmrx-Fdcp"" ihibmrx=""oiet/ajan; eaxlmyt=BJL-8""/> <meta name=""ljhestxv"" ihibmrx=""xitxq=wlpazg-xitxq, kurggew-seqjz=1.0, gcvyaoz-seqjz=1.0, bwxy-mdlxlzte=0""> <script fdcp=""oiet/tyieqmesga"">dhgFwsyPqgzq = [];</script> <link yne=""sxodhbocgj"" vbdc=""https://example.com/dilrzfofk/b=0.525.0/wca.qtk/rddb.wca"" /> <link yne=""sxodhbocgj"" vbdc=""https://example.com/dilrzfofk/b=0.525.0/wca.qtk/rdaf.wca"" /> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.525.0/dm/bkfiyv.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.525.0/dm/bkfiyv-rc.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" hsu=""https://example.com/dilrzfofk/b=0.525.0/dm/qtk/pircgm.dm""></script> <base vbdc=""/a-hlso/""> <meta name=""flnyXmr"" ihibmrx=""yiit""> <meta name=""tbq_tzggyh_id"" ihibmrx=""krtsihzuuqbz:1212izgwohh""> <meta name=""tbq_tzggyh_fdcp"" ihibmrx=""mbnups""> <meta name=""hwxrjjOkjubkMjz"" ihibmrx=""https://example.com/a-hlso/lsei/pnvigpbdpnms?fakbpLkgcbif=""> <meta name=""bqpSbti"" ihibmrx=""udolwper_bwavxs""> <meta name=""bqpSbti"" ihibmrx=""udolwper_ngfw_mbnups""> <meta name=""woqthyrt"" ihibmrx=""ngfw-mbnups""> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/laya.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/yiqztxn.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/qdxgfe-nfxlqe.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/tfrs/fkod/twaxvizfjf.dm/1.6.0/twaxvizfjf-qtk.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-1.5.0.qtk.dm"" xu=""ufmrfj""/> <link yne=""lxvmdzqats"" vbdc=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-dobomcc-1.5.0.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/ll-sezd-hunbzdmy.qtk.dm"" xu=""ufmrfj""/> <link yne=""ngoenym"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/dm/bwavxs/barug.dm"" xu=""ufmrfj""/> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/laya.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/yiqztxn.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/qdxgfe-nfxlqe.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/tfrs/fkod/twaxvizfjf.dm/1.6.0/twaxvizfjf-qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-1.5.0.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/dilrzfofk/b=0.505.0/dm/awfheip-dobomcc-1.5.0.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/bwavxs/dm/ll-sezd-hunbzdmy.qtk.dm""></script> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/dm/bwavxs/barug.dm""></script> <link yne=""sxodhbocgj"" vbdc=""https://example.com/a-hlso/dilrzfofk/1303/uftoukukrr/wca/ngfw/mbnups.qtk.wca""/> <script fdcp=""oiet/tyieqmesga"" eaxlmyt=""BJL-8"" hsu=""https://example.com/a-hlso/dilrzfofk/1303/uftoukukrr/qufjzi/uaglehgxskIatfsf.qtk.dm""></script> <meta name=""otashx"" ihibmrx=""barug, yxzubk""> <meta jrbacjgl=""zv:qwljy"" ihibmrx=""https://example.com/dibxjfch/giyx/1232268/kf23zrs5856e22c622b4984a6d473e99/khcw-qwljy-vsjzehl-tsmn.psj""> <meta jrbacjgl=""zv:wrjhr"" ihibmrx=""Mjm Ngfw Ayvzvy Gzg Xsxo Dozja""> <link yne=""wgojcddgn"" vbdc=""https://example.com/a-hlso/dozja/ngfw-ayvzvy-gzg-xsxo""> <title> Ngfw Ayvzvy Gzg Xsxo: Yjldc Kwecs-Dozja – 16 RG uck KYA</title> <meta name=""ibqypuvnyjx"" ihibmrx=""Vkeb hxslz Uhanppxk: Ngfw Ayvzvy Ueug Xsxo ✓ 16 RG KYA ✓ Ayvzvy-Aslp rüt Sptdwkxfnc &#38; XHR ✓ Leyobjwfm vüevbyi ➤ Qeagssdtrt nücfdzx pjvtrg scuruqc"" /> <!-- Maoy-Nqgacn-Wmu: /wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo?yvbr=wkfjhmb --> <!--1223498--> <script fdcp=""oiet/tyieqmesga"" hsu=""https://example.com/ezcunw/nicn/jmpgnwvh/pvfwatd-yqjtydrqs-dm-nicn""></script> <script hsu=""https://example.com/ucgy-rc/dm/obd.qtk.dm"" fdcp=""oiet/tyieqmesga""></script> <script hsu=""https://example.com/dilrzfofk/b=0.525.0/dm/avhqzdk-mxbj.dm"" id=""mxbj_avhqzdk_edkuonkb"" wookl=""yiit"" tsmn-cjukqee=""2000""></script> <script hsu=""//poddvc-b2.mxbj.kpr/ntsyrs/mxbj_13080_5022a5939d4305b872vw06mbz8c941et.dm"" id=""mxbj_ntsyrs_nicn_rddb"" wookl=""yiit""></script> </head> <body tsmn-pw-ohoqv> <script id=""fbqvztplqndj-luu"" tsmn-ncgscldk-id=""UmpqYaZO"" hsu=""https://example.com/ncvsxrz-rc/xbbnga/gqiwur.dm"" tsmn-zw-bdfo=""yiit"" wookl=""wookl""></script> <!-- Itkoek Ffs Umkslbe --> <noscript> <iframe hsu=""//hng.grqaszxeoocbtgda.kpr/ut.ajan?id=GBA-PSG9GX"" btyliv=""0"" xitxq=""0"" qptrx=""pbageib:dftr;hddyotrkhy:hnzgkv""> </iframe> </noscript> <script fdcp=""oiet/tyieqmesga""> (izoadtky (d, i, s, f, r) { d[f] = d[f] || []; d[f].evtr({ 'gba.tsdbf': fxm Mswb().mlkDdrl(), facyp: 'gba.dm' }); blh x = i.jgrJlxeymnrVpTqePfdf(s)[0], q = i.gwbyavFpxjkca(s), cl = f != 'vrbzRafml' ? '&f=' + f : ''; q.wookl = yiit; q.hsu = '//hng.grqaszxeoocbtgda.kpr/gba.dm?id=' + r + cl; x.gtcdvaMyuz.tqebecUrqmlp(q, x); })(kshknr, adxdwsjj, 'ufmrfj', 'dhgFwsyPqgzq', 'GBA-PSG9GX'); </script> <!-- Ufd Itkoek Ffs Umkslbe --> <!-- BW BARUG --> <div id=""umr""> <header sziby=""xdnlejfwiq xdnlejfwiq-vgei-nkiniu""> <div sziby=""xdnlejfwiq-ihibmrx""> <div sziby=""xdnlejfwiq-sbbv""> <a vbdc=""https://example.com/"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_tticsf-sbbv""> <div sziby=""xdnlejfwiq-sbbv-renjy""></div> </a> </div><!-- <navitem oyz=""-10005""> --><nav sziby=""xdnlejfwiq-khcw""> <ul sziby=""xdnlejfwiq-khcw-gkclz""> <li sziby=""xdnlejfwiq-khcw-mliu""> <a vbdc=""tyieqmesga:dtih(0);"" sziby=""fznfn-yjuy"" wrjhr=""Cykrjudcdkeeer"" yne=""ofgxzhmz"" tsmn-msm-vhai=""1232264"" tsmn-dm=""ekrpxb"" tsmn-msm-pcsaow=""23996"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Cykrjudcdkeeer 0800 40 40 410""> <span sziby=""pulkwvnbt pulkwvnbt-totil""></span>Cykrjudcdkeeer 0800 40 40 410</a> </li> </ul> </nav><!-- </navitem> --> <nav sziby=""xdnlejfwiq-rddb""> <ul> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/uhfebh/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Uhfebh"">Uhfebh</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><!-- </navitem> --> <a vbdc=""https://example.com/uhfebh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Omncvqmzzsk"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Omncvqmzzsk"">Omncvqmzzsk</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/uhfebh/leyobjwfm-asadajhh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Omncvqmzzsk frzzogiiz vüevbyi"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Omncvqmzzsk frzzogiiz vüevbyi"">Omncvqmzzsk frzzogiiz vüevbyi</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/uhfebh/aaafkuy-uhfebh/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Aaafkuy"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Aaafkuy"">Aaafkuy</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/ckcz/ghadg?fakm=https%3A%2F%2Fexample.com%2Flmgr-ngfw%2Fychedjatcilv%2F"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Pmobmfra rüt Zyrtct"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Pmobmfra rüt Zyrtct"">Pmobmfra rüt Zyrtct</a> </li> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/a-hlso/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Linpjt"">Linpjt</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Linpjt uck Pvfwatd"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Linpjt uck Pvfwatd"">Linpjt uck Pvfwatd</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/?pmobmfra=rblhddat"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Linpjt vkeb Pvfwatd "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Linpjt vkeb Pvfwatd "">Linpjt vkeb Pvfwatd</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/muwzmeaxely-linpjt"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Muwzmeaxely Linpjt"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Muwzmeaxely Linpjt"">Muwzmeaxely Linpjt</a> </li><!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/vzlvpzzrw/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Pmobmfra rüt Zyrtct"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Pmobmfra rüt Zyrtct"">Pmobmfra rüt Zyrtct</a><!-- </navitem> --> </li><!-- </navitem> --> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/jadfm-pmobmfra/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Umr-Pmobmfra"">Umr-Pmobmfra</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/jadfm-pmobmfra/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""% Xkea Umr-Pmobmfra "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_% Xkea Umr-Pmobmfra "">% Xkea Umr-Pmobmfra</a><!-- </navitem> --> </li><!-- </navitem> --> <!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/a-hlso/crkggqs/crkggqs-qxuxcp-h53-5w-ukfxqf-yhlqwor"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Umr Jadfm uck Dozja"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Umr Jadfm uck Dozja"">Umr Jadfm uck Dozja</a><!-- </navitem> --> </li><!-- </navitem> --> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/a-hlso/kd-qlzognps-qdnlualx-uck-qeagssdtrt"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""KD, Qlzognps & Ucgypöx uck Dozja"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_KD, Qlzognps & Ucgypöx uck Dozja"">KD, Qlzognps &#38; Ucgypöx uck Dozja</a> </li><!-- <navitem oyz=""-10005""> --><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <!-- <navitem oyz=""-10005""> --><a vbdc=""https://example.com/a-hlso/huczxsvnqjcl-uck-nuyfors"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Huczxsvnqjcl uck Ccäxnc"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Huczxsvnqjcl uck Ccäxnc"">Huczxsvnqjcl uck Ccäxnc</a><!-- </navitem> --> </li><!-- </navitem> --> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/muqcunn-cejyyg/?goqkhf=nkiniu"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Muqcunn cejyyg - xbm hl 110 € Ccäxnc"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Muqcunn cejyyg - xbm hl 110 € Ccäxnc"">Muqcunn cejyyg - xbm hl 110 € Ccäxnc</a> </li> </ul> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-rddb ""> <a vbdc=""https://example.com/hkvrofd/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""yuumixbe"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Hkvrofd"">Hkvrofd</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-rddb""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""""> <li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Xkea Hkvrofd-Pjohyj "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Xkea Hkvrofd-Pjohyj "">Xkea Hkvrofd-Pjohyj</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/wnrlgdcf/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Jzchlq lip Wnrlgdcf "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Jzchlq lip Wnrlgdcf "">Jzchlq lip Wnrlgdcf</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/iscasnmk-unmpeooz/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Aaafkuy Iscasnmk unmpeooz "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Aaafkuy Iscasnmk unmpeooz "">Aaafkuy Iscasnmk unmpeooz</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/bmktjrzsrtzccrktkr/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Bmktjrzsrtzccrktkr "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Bmktjrzsrtzccrktkr "">Bmktjrzsrtzccrktkr</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/fusdtsnwsdcaedvdpyrno/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Nlvsewiomhtbäyyruwzr"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Nlvsewiomhtbäyyruwzr"">Nlvsewiomhtbäyyruwzr</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/udhdnpuuxaod/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Udhdnpuuxaod"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Udhdnpuuxaod"">Udhdnpuuxaod</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/cndrzoxryfwfp/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Wkvuzdyeüghcjelnl cdüvuc "" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Wkvuzdyeüghcjelnl cdüvuc "">Wkvuzdyeüghcjelnl cdüvuc</a> </li><li sziby=""xdnlejfwiq-foz-rddb-mliu xdnlejfwiq-yjuy-xobb""> <a vbdc=""https://example.com/hkvrofd/lmgr-ngfw-obd/"" sziby=""xdnlejfwiq-foz-rddb-ooaiy"" wrjhr=""Lmgr Ngfw Obd"" tsmn-jmpgnwvh-rdctxp=""tqzsyxn"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Lmgr Ngfw Obd"">Lmgr Ngfw Obd</a> </li> </ul> </div> </div> </li> </ul> </nav><nav sziby=""xdnlejfwiq-srgdwcm""> <ul> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-geh""> <a vbdc=""https://example.com/"" sziby=""xdnlejfwiq-ooaiy"" wrjhr=""GEH fmkdfxmswipi"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_lcijdpa""><span sziby=""pulkwvnbt pulkwvnbt-geh-vkav""></span>GEH fmkdfxmswipi</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-geh""> <div sziby=""xdnlejfwiq-geh-wrjhr""> <div sziby=""pulkwvnbt pulkwvnbt-geh-vkav""></div>Geh </div> <div sziby=""xdnlejfwiq-gyhy-vffic""> </div> </div> </div> </div> </li><li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-xzhqba""> <a vbdc=""https://example.com/mlmuq/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_xzhqba""> <span sziby=""pulkwvnbt pulkwvnbt-xzhqba""></span> </a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-xzhqba"" tsmn-dm-xzhqba-nkiniu-nqgacn=""{&#39;name&#39;:&#39;ngfw&#39;, &#39;nfcrufrnlfVwm&#39;:&#39;/mlmuq/snuexjcvef&#39;, &#39;jn&#39;:337268}""> <form sziby=""xdnlejfwiq-xzhqba-oggt"" name=""ldvbajxwjnolkw"" rrordecrytkc=""pcd"" fssu=""oggt"" rdctxp=""https://example.com/mlmuq/"" muwsbc=""RXA""> <fieldset> <input fdcp=""oiet"" value="""" name=""z"" wrjhr="""" sziby=""oggt-mdhbcbn"" rrordecrytkc=""pcd"" rrlbkrjyqlv=""Xsxbqeyrdwj"" tsmn-jmpgnwvh-rdctxp=""oggt"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Xsxbqeyrdwj"" /> </fieldset> </form> <div sziby=""xdnlejfwiq-xzhqba-yngtsyh""></div> <div sziby=""xdnlejfwiq-xzhqba-bkbsm-gkclz""> <ul> </ul> </div> </div> </div> </div> </li><!-- <navitem oyz=""-10000""> --><!-- Value ilxt ot.j2.kqj.XNXTmmpbteqdi (OYZ_MGTRTJ_DC=10000) --> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-ghadg xdnlejfwiq-mliu-mgtrtj-fgb""> <!-- <navitem oyz=""-10000""> --><!-- <navitem oyz=""-10000""> --><!-- </navitem> --> <a vbdc=""https://example.com/lmgr-ngfw/"" sziby=""xdnlejfwiq-ooaiy"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_bwxy""><span sziby=""pulkwvnbt pulkwvnbt-bwxy""></span>Ghadg</a><!-- </navitem> --> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <ul sziby=""xdnlejfwiq-ghadg""> </ul> </div> </div> </li><!-- </navitem> --> <li sziby=""xdnlejfwiq-mliu xdnlejfwiq-mliu-srgdwcm xdnlejfwiq-mliu-gyhy""> <a vbdc=""https://example.com/a-hlso/fhxxyaexg"" sziby=""xdnlejfwiq-ooaiy"" wrjhr=""Fhxxyaexg"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_gyhy""><span sziby=""pulkwvnbt pulkwvnbt-gyhy""></span>Fhxxyaexg</a> <div sziby=""xdnlejfwiq-foz xdnlejfwiq-foz-srgdwcm""> <div sziby=""xdnlejfwiq-foz-ihibmrx""> <div sziby=""xdnlejfwiq-gyhy""> <div sziby=""xdnlejfwiq-gyhy-wrjhr""> <div sziby=""pulkwvnbt pulkwvnbt-gyhy""></div> Fhxxyaexg </div> <div tsmn-dm-xdnlejfwiq-gyhy-oauseqydb=""""></div> </div> </div> </div> </li> </ul> </nav> </div> </header><!-- AvHqzdk Mxbj 24278 hfkqn --> <div id=""24278"" sziby=""avhqzdk-mxbj ykqzeir cxqc"" tsmn-dm=""avhqzdk-mxbj""> <div sziby=""gqiwur""> <div sziby=""ecnrdcm ykqzeir-ecnrdcm-2""> <div sziby=""hkaxqj1""></div> <div sziby=""hkaxqj2""></div> <div sziby=""hkaxqj3""></div> </div> </div> </div> <!-- AvHqzdk Mxbj 24278 ufd --> </div><!-- BARUG --> <div tsmn-hu-jmpgnwvh-lenilns-jccsl=""'GEH_AKZU'""> <co-lufqc-yhlqwor></co-lufqc-yhlqwor> <co-pcsaow-uczoj-fikslaye></co-pcsaow-uczoj-fikslaye> <co-rblhddat-hyrwzltfk></co-rblhddat-hyrwzltfk> <co-mbnups-yhlqwor></co-mbnups-yhlqwor> </div> <div id=""tbkwcs""> <footer sziby=""ksizpx""> <section sziby=""fflwu-lqyiny ""> <div sziby=""ifqwg""> <picture> <source fdcp=""qwljy/zgtz"" wcwipt=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/psj/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.zgtz""></source> <source wcwipt=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.psj""></source> <img hsu=""https://example.com/dibxjfch/giyx/23632/566e0aqgl1a2kq0972187324lr26b4ns/qqzocogagje-ksizpx-qwljy-vsjzehl-tsmn.psj"" sziby="" fmy-okuprvxrns"" wrjhr=""example.com"" zrb=""example.com"" /> </picture><!-- <navitem oyz=""-10005""> --><div sziby=""mliu-jddhxroayz "" id=""id-23618""> <div sziby=""vffic""> <div sziby=""itsxwah""> <div sziby=""mliu-rlnpstqw""> Cng vebjk Jzchlq hl gogxagw Iumsjkdzs? Redoc Cng hos bs! </div><p><span sziby=""e1""><a vbdc=""wby:72408622744"" sziby="""" pcsaow=""_hifj"" tsmn-jmpgnwvh-rdctxp=""nhtkbvlz"" tsmn-jmpgnwvh-ibqypuvnyjx=""msm__msm/wkfjhmb/ngfw-mbnups/ngfw/ngfw-ayvzvy-gzg-xsxo/yvbr_Cykrjudcdkeeer 0800 40 40 41 0"">0800 40 40 410</a></span>"|||"dhgFwsyPqgzq = []; Ngfw Ayvzvy Gzg Xsxo: Yjldc Kwecs-Dozja – 16 RG uck KYA (izoadtky (d, i, s, f, r) { d[f] = d[f] || []; d[f].evtr({ 'gba.tsdbf': fxm Mswb().mlkDdrl(), facyp: 'gba.dm' }); blh x = i.jgrJlxeymnrVpTqePfdf(s)[0], q = i.gwbyavFpxjkca(s), cl = f != 'vrbzRafml' ? '&f=' + f: ''; q.wookl = yiit; q.hsu = '//hng.grqaszxeoocbtgda.kpr/gba.dm?id=' + r + cl; x.gtcdvaMyuz.tqebecUrqmlp(q, x); })(kshknr, adxdwsjj, 'ufmrfj', 'dhgFwsyPqgzq', 'GBA-PSG9GX'); --> Cykrjudcdkeeer 0800 40 40 410 --> Uhfebh --> --> Omncvqmzzsk Omncvqmzzsk frzzogiiz vüevbyi Aaafkuy Pmobmfra rüt Zyrtct Linpjt Linpjt uck Pvfwatd Linpjt vkeb Pvfwatd Muwzmeaxely Linpjt --> --> Pmobmfra rüt Zyrtct --> --> Umr-Pmobmfra --> --> % Xkea Umr-Pmobmfra --> --> --> --> Umr Jadfm uck Dozja --> --> KD, Qlzognps &#38; Ucgypöx uck Dozja --> --> Huczxsvnqjcl uck Ccäxnc --> --> Muqcunn cejyyg - xbm hl 110 € Ccäxnc Hkvrofd Xkea Hkvrofd-Pjohyj Jzchlq lip Wnrlgdcf Aaafkuy Iscasnmk unmpeooz Bmktjrzsrtzccrktkr Nlvsewiomhtbäyyruwzr Udhdnpuuxaod Wkvuzdyeüghcjelnl cdüvuc Lmgr Ngfw Obd GEH fmkdfxmswipi Geh --> --> --> --> Ghadg --> --> Fhxxyaexg Fhxxyaexg --> Cng vebjk Jzchlq hl gogxagw Iumsjkdzs? Redoc Cng hos bs! 0800 40 40 410 &lt;r&gt; Cxg­ffs - Uwpglqg: 8 - 20 Zvh. Kcrnzbj: 9 - 18 Zvh --> --> Ngfw Lyqjdtyeam Vqcag cbeixvrcrg fjc dfrboa jqax inkibnqcw! * Vbnalfenwlu Pyxxqwobw ijkdwhha Uin bes znqxrkduxd 16 Hruxp zrb fjc wxbfcwryxzmma, yep Ngfw osermaqgkyfqltilw hjy A-Uthb üuqa Hgfvhhoifemfglr yep Ngfw xjuiuegoef hl ggfynq. Uin bes yhdsjfvol klcei wxbfcwryxzmma, jknq Ngfw lmgr Igpnocyeurkfdrl ykn opk Laysfmxf, Trguidodfn fjc wfsfehg Rdaf yep Ngfw pdav Powut ykn pkxdm Yjuy nr Lyqjdtyeam hbujmox Fkifyl asjomlnubufgf fjc graxzzbrprs . Aaet gxnlq lip Krjpkdoegk gqrck Vyumpunvonogqii , pp jqh tcpelcfmptzirca Ngfw Emqpptbo eafmmvlfon , fjc pp uupzhlsgnbtev, ui uin pjvtrg nzs Ngfw Syjmnts hbcbsqtf (qaa Ldlov mjm Xnpmjrotkqhxre vzf Cibsuflsopmsxltgtfo). Uin vzec bgavj Hxphdlwawxgh leyobjwfm uck Zcqxtzu rüt hwu Dguymct vfacivjquu, dsvwn uin dwje iup Lyqjdtyeam ayuriel . Ustucphyys Otzftpapbdinl lip Jocxmtblybjtnrjcb jyc Ngfw fjc hl Oviwpju hcpt dc sehzqva Kwqoswepbdquwwväwcen hl quufbm. Rüt wwiiw nnupypddyrtfbqwykw Lyqjdtyeam kgwwg jjwc ijkdwhha. Jqax nqetwywz --> vngawgs Etsaqzvy-Rhkqq Ngfw sdl axnmcz Efbzkamudl rüt Sipkdclqobbw. Jqax nqetwywz vngawgs Dsisqzilvbfqp Ngfw sdl Efbzkamudl dc mjm Ncwxupevz Rmfvgeybpfysh. Jqax nqetwywz vngawgs Obd-Hwyq Lmgr Ngfw Hkvrofd-Obd eqt „dlkn cep“ bijiqpizgyxzr. Jqax nqetwywz vngawgs Jblfxqkl Ngfw nuzetb tqutelvnmeceyo Wbchkrsklpäx. Jqax nqetwywz Wüoenkrb Omncvqmzzsk Xkea Icsdqppmhdmoudq Chqljbbxlhfcfz Ctkhw Owcjrw &#38; Sabdg Zxcabl Hkvrofd Jadfm Xuwfvxqokm &#38; Enjczo Muqcunn cejyyg Üuqa Ngfw Sguzddcelbazyj Wlmdtid Wtotnrohz VNG &#38; Eredzactzqhwytjnkgsi Ydjvwwpv AbadczrQ/GtmrF Kwldypdftpj Oviwpju Bakayobajjag- &#38; Dtqlvdwbjbfe Pvfwatd iüxqtfrs Pdav wpvu © Wkbbfózapm Gjflcwu KjjQ &#38; Hu. GPR --> Wzv eögkot fdh Lukdt xwfayw? Cykrjudcdkeeer 0800 62 66 24 5 Hh.-Uw. 8-20 Zvh, Xe. 9-18 Zvh Cykrjudcdkeeer 0800 62 66 245 Ncwei nr Hkvrofd-Larktsd Ncwei nvazmf fjc quufbm Ncwei nvazmf fjc quufbm Wzv eögkot fdh Lukdt xwfayw? --> { ""@fdcp"": ""uftoUkukrr:PumnHlntchLrnuLvyth"", ""gxjgvmcyvyoir"": { ""@fdcp"": ""uftoUkukrr:EnvdowndmdmcuIrrpb"", ""mbnups"": { ""@fdcp"": ""uftoUkukrr:GrgfpvHigcs"", ""aqwrdxye"": { ""@fdcp"": ""uftoUkukrr:ItptoeZtmkmbcyEbiuo"", ""ahhhjg"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""name"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""ibqypuvnyjx"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""cqxmyedbmbRaynaah"": [ ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."" ], ""uagnmOvl"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""rrfvvggf"": [ ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."" ], ""sxbeplzaig"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""oiitWgcpwpGtXjdseib"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""gwdybqCuswgkg"": "" Wkqyme Pyqnipbcl vwhrfwwpw fjc wzv xatlspr hhhwcimxnn avihpvn. Sqdb Pyqnipbcl eögkot Cng nkxnsgo hl Ngfw vwhrfwwpw, invf, ui Cng jyc qooam Aaafkuy-Dlumdivj hcpt ypen pkxdm Vzgfsdgnrqwccmn vebjk. Ra avihpvn Cng rüt Sqdb Muqcunn, Hwjqwez fjc Kpaxzälrqbqokhupj leyobjwfm hhhwcimxnn. Jyc Jzchlq redoc kgwwg Cng nkxnsgo njsdjw gkuhokrrdu Cykrjudcdkeeer bs: 0800 40 40 410 Pkesmi - Uwpglqg: 8 - 20 Zvh. Kcrnzbj: 9 - 18 Zvh. Jqax Otzftpapbdinl lip Bmktjrzsrtzccrktkr quufbm Cng jjwc . "", ""eofwomTdqiuawwpYkydgwg"": "" Vebjk Cng Jzchlq hl jiakfe Dozja? Zaxhdld­thb­xxto Pwrölv­wfe Cng hidn hjy­böp­duqrz Imbchfor? Redoc Cng hos bs. Hh - Uw.: 8 - 20 Zvh, Xe.: 9 - 18 Zvh 0800 - 40 40 41 0 (pyxxqwobw) BWR Cng qu­wqb qf­ew Jrb­bz? Fdh hwu Gft­ffto. Qaa Hkvrofd-Larktsd "", ""rhcfktNzxrymei24TklxsbLqji"": ""24 Goobfo Uhanppxk"", ""isybgyCcuWnaxZmbi"": ""Frzzogiiz vüevbyi"", ""xcjeyhlxe"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""tnaursfAd"": ""1212izgwohh"", ""dnhftcipm"": """" }, ""ryirlpSezmuov"": [ { ""@fdcp"": ""uftoUkukrr:NuxggzLvwiwoEmpqs"", ""ukhtunCbqw"": ""krtsihzuuqbz-ngfw-ayvzvy-f-xsxo-obmba-fezyt-1rg-24x5"", ""oiitWgcpwpGtXjdseib"": ""4096"", ""opxriyiw"": yslts, ""gbulmzBjbeylKhdcpkWsbo"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1hwAqrTIAcqS1qdVN4EUFyVMOpLNWmn21qZVBcRv0dCDc1O3PxK2nqLADzHF1qovotp3XrDLP0WS5tHFCoMECtCREvvyJsNb1s"", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } }, { ""@fdcp"": ""uftoUkukrr:NuxggzLvwiwoEmpqs"", ""ukhtunCbqw"": ""krtsihzuuqbz-ngfw-ayvzvy-ar-xsxo-obmba-fezyt-5rg-24x8"", ""oiitWgcpwpGtXjdseib"": ""5120"", ""opxriyiw"": yslts, ""gbulmzBjbeylKhdcpkWsbo"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1ygUhtDVChyEnoUxmkzC1wgBXlVL1fzv9onk01Z2GpHhB4LOktNGHkfh1wu2ZamH5rAEV0RPcowMKxSLRnBYOxayBdwI51mtiyzMh="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } }, { ""@fdcp"": ""uftoUkukrr:NuxggzLvwiwoEmpqs"", ""ukhtunCbqw"": ""krtsihzuuqbz-ngfw-ayvzvy-ueug-xsxo-obmba-fezyt-4rg-24x10"", ""oiitWgcpwpGtXjdseib"": ""12288"", ""opxriyiw"": yslts, ""gbulmzBjbeylKhdcpkWsbo"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1vmIzyYSNixEk1vi1qdVN4EUFyVMOpLNWmn21cKGYtSx0rWNukIZrbPAKbsi1wu2ZamH5rAEV0RPcowMKxSLRnBYOxayBdwI51bastERI0UV=="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } }, { ""@fdcp"": ""uftoUkukrr:NuxggzLvwiwoEmpqs"", ""ukhtunCbqw"": ""krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11"", ""oiitWgcpwpGtXjdseib"": ""404 - Vzuedppi liv dujhnsci (nzjksb nicn 404)."", ""opxriyiw"": yiit, ""gbulmzBjbeylKhdcpkWsbo"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1jfRvtIUGyzFQ4YAHjLOzwv3StGYUorTWhfD8tZjY4LVV7mUDly3M9he9qiGdlUn1mzEL="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } } ], ""vfguWiinku"": [ { ""@fdcp"": ""uftoUkukrr:PjimXdqdfXfggd"", ""id"": ""Jjuiayz Viehu Ngfw"", ""aqwrdxye"": { ""@fdcp"": ""uftoUkukrr:XjlmGqefrAzpdlnqnWkecc"", ""name"": ""Jjuiayz"" }, ""viehu"": [ { ""@fdcp"": ""uftoUkukrr:IgvpTtpsz"", ""aqwrdxye"": { ""@fdcp"": ""uftoUkukrr:FdmoCwbbjepdHrpqi"", ""name"": ""Xoqytlfglkz ZW Jjuiayz-Dozja"", ""sdyyi"": 0.0, ""ibqypuvnyjx"": """", ""yhlqwor"": "" Nr thnceuwqhjp Ngfw ZW Jjuiayz Dozja xhufbj nnkliogh Rgflhpxmmtz nr ZW-Kyycgpo: Oezbrbaddwdmqtpkof rüt plymsqpqz Anjoväsds; Cufuinsskl Anjoväsds pyxxqwobw; XHR-Fkaoaxk Reptzeenufvqäb fjc -Qzeyflieydzt zzo. Oezbrbaddwdmqtpkof. Ndhkndbükcezzk Otzftpapbdinl hl gogxagw Irywpnkx- Svgtnjelbvvilggohmj gwnxxoid Cng jjwc: https://example.com/uhfebh/kyycgpo/ Vgnazmm: Rüt Efpßivwzjjteth kbxs ojwbi Iflzznlgzekz pmx mjm ZW xbm qaa 31.12.2022 fmwgxgqlq soq mjm Yeegp zzo. Vuzaxswy 1 (ZW) kdfbukeswvk (Xvrväwzzzmpo nkrufmuygmr). &#160; "" }, ""vspud"": { ""@fdcp"": ""uftoUkukrr:UmjiKmnoa"", ""jkpkpdPrfnPlmf"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1jfRvtIUGyzFQ4YAHjLOzwv3StGYUorTWhfD8tZjY4LVV7mUDly3M9he9qiGdlUn1mzEL="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } } }, ""opxriyiw"": yiit } ] } ] }, ""sdyyi"": { ""@fdcp"": ""uftoUkukrr:DbnpjIjimm"", ""rqjlhtdggCykgtu"": 16.99, ""gdkNmtzgdaxaIsftov"": 0, ""aqwrdxye"": { ""@fdcp"": ""uftoUkukrr:ZptdzWrbcqqckVixmh"", ""efroyveiMwlxhkxzOkrasvicyhx"": ""Frzzogiiz"", ""kdivdphwykXxzHgkqpdshyje"": ""+ Zlloyhrxlalgbt: 29,99 €"" } }, ""weytzgriRqrh"": { ""@fdcp"": ""uftoUkukrr:WeytzgriRqrh"", ""name"": ""BW_DYBVNJX_HUNBZDMY_IPASLTZM"" }, ""vspud"": { ""@fdcp"": ""uftoUkukrr:KsczrqyniypsrRzhxc"", ""ajmaqcPuby"": { ""@fdcp"": ""yeypunkgeizdzv:KhmsTygi"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/xyxgjh-opxriyiw-gxjgvmcyvyoir"", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" }, ""tsmn"": { ""@fdcp"": ""uftoUkukrr:OpzwiveMvxlidmczwuixBazin"", ""fapbJekzylu"": ""ngfw"", ""ohrjyiUtcms"": ""krtsihzuuqbz"", ""jkmvmmBigokDnbh"": ""krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11"", ""tqbwzzefCgvahxLqcqHoazm"": [ ""jjuiayz-vrz"" ] } }, ""cxprkuJolsavzeYivnTpfh"": { ""@fdcp"": ""yeypunkgeizdzv:MngJkch"", ""yjuy"": { ""@fdcp"": ""yeypunkgeizdzv:Yjuy"", ""iuj"": ""https://example.com/a-hlso/lsei/ngfw-mbnups/rxa-fxm-gxjgvmcyvyoir/aMfaNE1VS0BBCLTnMHAYMMEJX3LjkdajDa1awZT1Icemnfi2QXBwkC5jPQ4iYemnfA1jfRvtIUGyzFQ4EUFyVMOpLNWmn21aVMO0dEw7mUDly3M9he9qiGdlUn1mzEL="", ""jeelxJeqv"": ""eoxmstdbxle/azy.udolwper.etmhsfa+bxxq"" } }, ""radzajtbZabp"": { ""@fdcp"": ""yeypunkgeizdzv:UumwmleoAyut"", ""abeshgijUlxkate"": { ""@fdcp"": ""yeypunkgeizdzv:RcrejariGgdkWzyyx"", ""qroGxogwum"": ""{\"lenilns_opxriyiw\":[{\"nuvqozzg\":0,\"fdcp\":\"lyjpwgex\",\"rdctxp\":\"Njotdixreynoao\",\"name\":\"krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11\",\"id\":\"73392691\",\"kbdewp_uazb\":0.0,\"xynwajw_uazb\":27.99,\"zitf_ygdubn\":1,\"xjgtx_bybna\":0,\"zaqsfdgmrw_udp\":29.99,\"plmladas\":dmko,\"hlso_inzkmwz\":\"ngfw\",\"gyhy_mliu_id\":dmko,\"hunbzdmy_elosczw\":dmko,\"hoowzthkc\":[{\"mqopKfwmwYnynpdoy\":yiit,\"tggpcsPGN\":0.0,\"kwtjcgPP\":11.0,\"ipasltzm\":24,\"fdcp\":\"Uhspayoe\"}],\"mfzzmq_zrsbdbm\":dmko,\"lcijdpa_ihbolmmp_id\":dmko,\"lcijdpa_mbnups_nicn\":dmko,\"iabogd_id\":dmko,\"akyroz\":dmko,\"booftlhyr_nicn\":\"BW_BOOFTLHYR\"},{\"nuvqozzg\":0,\"fdcp\":\"ofju\",\"rdctxp\":\"Njotdixreynoao\",\"name\":\"Irywpnkx Jjuiayz\",\"id\":\"2_JJUIAYZ_VRZ\",\"kbdewp_uazb\":0.0,\"xynwajw_uazb\":0.0,\"zitf_ygdubn\":dmko,\"xjgtx_bybna\":dmko,\"zaqsfdgmrw_udp\":dmko,\"plmladas\":1,\"hlso_inzkmwz\":dmko,\"gyhy_mliu_id\":dmko,\"hunbzdmy_elosczw\":dmko,\"hoowzthkc\":dmko,\"mfzzmq_zrsbdbm\":dmko,\"lcijdpa_ihbolmmp_id\":dmko,\"lcijdpa_mbnups_nicn\":dmko,\"iabogd_id\":dmko,\"akyroz\":dmko,\"booftlhyr_nicn\":dmko}]}"" } } }, ""qghgdwWiupxWcvahnwwOqvkb"": dmko, ""xyrjvrAjthdckiXkhvmywFcc"": ""/bhv-ch/947602/?oyz=3510096&asr=E&glx[zojx]=hlso&glx[maoy]=ngfw-mbnups&glx[uyl]=krtsihzuuqbz&glx[bmr]=krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11&glx[gko]=xsxo&glx[zet]=zitf&glx[tve]=1&glx[is]=hng"", ""eovlhEahddikdAthnKycies"": yiit, ""mhndohWaifdmWicgfrlUkc"": ""/bhv-ch/947602/?oyz=3904948&asr=E&glx[zojx]=hlso&glx[maoy]=ngfw-mbnups&glx[uyl]=krtsihzuuqbz&glx[bmr]=krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11&glx[gko]=xsxo&glx[zet]=zitf&glx[tve]=1&glx[is]=hng"", ""jewbxOczlizetSczrfpvx"": yslts }, ""roktqmVsjrmpd"": """", ""ooxeoxhLivcbtrzCvzwblx"": yiit, ""hnvcQtullnqtZiuulzb"": { ""@fdcp"": ""uftoUkukrr:HnvcQtullnqtZiuulzb"", ""jtbzwwfr"": ""lyjpwgex-awf"", ""ndcfpejMdirklov"": ""lyjpwgex_mbnups"" }, ""gcunimgiuGfglcdwXtpnut"": [ { ""@fdcp"": ""uftoUkukrr:MtyedhawcGiiqlvvGmhxc"", ""name"": ""krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11"", ""plmladas"": 1, ""qqzjckhoug"": { ""nuvqozzg"": ""1"", ""renjy"": ""ngfw"", ""mybdhzpcb163"": ""krtsihzuuqbz"", ""mybdhzpcb167"": ""eoxvxfzw"", ""mybdhzpcb136"": """", ""mybdhzpcb62"": ""BW_BOOFTLHYR"", ""mybdhzpcb60"": dmko, ""mybdhzpcb59"": ""ngfw-ayvzvy-gzg-xsxo"", ""gziuvsx"": """", ""mybdhzpcb137"": ""krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11"", ""nxqfkt6"": ""0"", ""mybdhzpcb166"": ""bw-upui-viehu"", ""nxqfkt2"": ""27.99"", ""nxqfkt4"": ""29.99"", ""nxqfkt8"": ""437.75"", ""id"": ""73392691"", ""jtbzwwfr"": ""lyjpwgex_geh-akzu"", ""sdyyi"": ""437.75"" } }, { ""@fdcp"": ""uftoUkukrr:MtyedhawcGiiqlvvGmhxc"", ""name"": ""krtsihzuuqbz-ngfw-ayvzvy-ueug-xsxo-obmba-fezyt-4rg-24x10"", ""plmladas"": 1, ""qqzjckhoug"": { ""nuvqozzg"": ""1"", ""renjy"": ""ngfw"", ""mybdhzpcb163"": ""krtsihzuuqbz"", ""mybdhzpcb167"": ""eoxvxfzw"", ""mybdhzpcb136"": """", ""mybdhzpcb62"": ""BW_BOOFTLHYR"", ""mybdhzpcb60"": dmko, ""mybdhzpcb59"": ""ngfw-ayvzvy-ueug-xsxo"", ""mybdhzpcb137"": ""krtsihzuuqbz-ngfw-ayvzvy-ueug-xsxo-obmba-fezyt-4rg-24x10"", ""gziuvsx"": """", ""nxqfkt6"": ""0"", ""mybdhzpcb166"": ""bw-upui-viehu"", ""nxqfkt2"": ""21.99"", ""nxqfkt4"": ""29.99"", ""nxqfkt8"": ""293.75"", ""id"": ""12480986"", ""jtbzwwfr"": ""lyjpwgex_geh-akzu"", ""sdyyi"": ""293.75"" } }, { ""@fdcp"": ""uftoUkukrr:MtyedhawcGiiqlvvGmhxc"", ""name"": ""krtsihzuuqbz-ngfw-ayvzvy-f-xsxo-obmba-fezyt-1rg-24x5"", ""plmladas"": 1, ""qqzjckhoug"": { ""nuvqozzg"": ""1"", ""renjy"": ""ngfw"", ""mybdhzpcb163"": ""krtsihzuuqbz"", ""mybdhzpcb167"": ""eoxvxfzw"", ""mybdhzpcb136"": """", ""mybdhzpcb62"": ""BW_BOOFTLHYR"", ""mybdhzpcb60"": dmko, ""mybdhzpcb59"": ""ngfw-ayvzvy-f-xsxo"", ""gziuvsx"": """", ""mybdhzpcb137"": ""krtsihzuuqbz-ngfw-ayvzvy-f-xsxo-obmba-fezyt-1rg-24x5"", ""nxqfkt6"": ""0"", ""mybdhzpcb166"": ""bw-upui-viehu"", ""nxqfkt2"": ""11.99"", ""nxqfkt4"": ""29.99"", ""nxqfkt8"": ""197.75"", ""id"": ""44910904"", ""jtbzwwfr"": ""lyjpwgex_geh-akzu"", ""sdyyi"": ""197.75"" } }, { ""@fdcp"": ""uftoUkukrr:MtyedhawcGiiqlvvGmhxc"", ""name"": ""krtsihzuuqbz-ngfw-ayvzvy-ar-xsxo-obmba-fezyt-5rg-24x8"", ""plmladas"": 1, ""qqzjckhoug"": { ""nuvqozzg"": ""1"", ""renjy"": ""ngfw"", ""mybdhzpcb163"": ""krtsihzuuqbz"", ""mybdhzpcb167"": ""eoxvxfzw"", ""mybdhzpcb136"": """", ""mybdhzpcb62"": ""BW_BOOFTLHYR"", ""mybdhzpcb60"": dmko, ""mybdhzpcb59"": ""ngfw-ayvzvy-ar-xsxo"", ""mybdhzpcb137"": ""krtsihzuuqbz-ngfw-ayvzvy-ar-xsxo-obmba-fezyt-5rg-24x8"", ""gziuvsx"": """", ""nxqfkt6"": ""0"", ""mybdhzpcb166"": ""bw-upui-viehu"", ""nxqfkt2"": ""16.99"", ""nxqfkt4"": ""29.99"", ""nxqfkt8"": ""245.75"", ""id"": ""40051053"", ""jtbzwwfr"": ""lyjpwgex_geh-akzu"", ""sdyyi"": ""245.75"" } } ], ""ftgflZuihnyEptkbeupXidsr"": { ""@fdcp"": ""uftoUkukrr:FtgflZuihnyEptkbeupXidsr"", ""plef_fkifyl_value"": ""IAGEW_MGTRTJ_DC"", ""mgtrtj_dc"": yslts, ""bwxy_fdcp"": """" } } dhgFwsyPqgzq.evtr({""poddvc"":[],""lenilns_opxriyiw"":[{""nuvqozzg"":0,""fdcp"":""lyjpwgex"",""rdctxp"":""Njotdixreynoao"",""name"":""krtsihzuuqbz-ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11"",""id"":""73392691"",""kbdewp_uazb"":0.0,""xynwajw_uazb"":27.99,""zitf_ygdubn"":1,""xjgtx_bybna"":0,""zaqsfdgmrw_udp"":29.99,""plmladas"":dmko,""hlso_inzkmwz"":""ngfw"",""gyhy_mliu_id"":dmko,""hunbzdmy_elosczw"":dmko,""hoowzthkc"":[{""mqopKfwmwYnynpdoy"":yiit,""tggpcsPGN"":0.0,""kwtjcgPP"":11.0,""ipasltzm"":24,""fdcp"":""Uhspayoe""}],""mfzzmq_zrsbdbm"":dmko,""lcijdpa_ihbolmmp_id"":dmko,""lcijdpa_mbnups_nicn"":dmko,""iabogd_id"":dmko,""akyroz"":dmko,""booftlhyr_nicn"":""BW_BOOFTLHYR""},{""nuvqozzg"":0,""fdcp"":""ofju"",""rdctxp"":""Njotdixreynoao"",""name"":""Irywpnkx Jjuiayz"",""id"":""2_JJUIAYZ_VRZ"",""kbdewp_uazb"":0.0,""xynwajw_uazb"":0.0,""zitf_ygdubn"":dmko,""xjgtx_bybna"":dmko,""zaqsfdgmrw_udp"":dmko,""plmladas"":1,""hlso_inzkmwz"":dmko,""gyhy_mliu_id"":dmko,""hunbzdmy_elosczw"":dmko,""hoowzthkc"":dmko,""mfzzmq_zrsbdbm"":dmko,""lcijdpa_ihbolmmp_id"":dmko,""lcijdpa_mbnups_nicn"":dmko,""iabogd_id"":dmko,""akyroz"":dmko,""booftlhyr_nicn"":dmko}],""sbmupwwlz"":[],""adsiivmq"":{""djtj_ygdubn"":80,""djtj_name"":""umsmcKzvkiz""},""rjkzy_ghadg_nzjksb"":{""mgtrtj_dc"":yslts,""bwxy_fdcp"":"""",""plef_fkifyl_value"":""IAGEW_MGTRTJ_DC""}})"||||"18 RG KYA"||||0|||ngfw-ayvzvy-gzg-xsxo-obmba-fezyt-24x11|`),
				enclosure: characters.Character_QUOTATION_MARK,
				sep:       characters.Character_PIPE,
			},
			want:  116,
			debug: false,
			php:   true,
		},

		{
			name: "0c13563c51d17383074e56d1753d7b8d",
			args: args{
				s:         []byte(`,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,ipasltzm,1,6,H0.6.00018_vfdpux,Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv|Khyybgiwjsv,0|0|0|0|0|0,"<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>|<p>Cng ltjözsudm Ekb Tpxd hl qooam ghästtqh Juqkpkoie? - Owgk Duknpke!</p> <p>Cng vebjk hwu Oögvlifmmvd owqqe vqcag hl iiomkymjc fjc tqqe hwu grdhltwcc Rgflhpxmmtz hl scuruqc.</p> <p>Pöpmcxr Cng hwu Khyybgiwjsv dc Ayijzvdt vwfqed, säqtbv Cng kgwwg pmx opk Kkslmcon hwu juvüqughjx Lfefx mjm Khyybgiwjsv.</p> <p><strong>Dyvhei Cng dh Yxmq vzf Hndgfjlkabtddbo nzs, gd fdpa hwu Ubmosxjam vaehqmos vzec fjc xugxvsßva Cng hwu Yhssoqjkof gd.</strong></p>",H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,H0.6.00018|H0.6.00018.1|H0.6.00018.2|H0.6.00018.3|H0.6.00018.4|H0.6.00018.5,|||||,,,,,,,,|||||,|||||,|||||,|||||,|||||,|||||,|||||,|||||`),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  31,
			debug: false,
		},

		{
			name: "vagkg",
			args: args{
				s:         []byte(`"5",361708,"Xdnnag Qscpsat Kop Gykna Runu Mfnr",https://example.com/ot/ot/vfsrxx/esvo-qscpsat-kop-gykna-runu_361708/213240828,https://example.com/vfsrxx/183x239/96800229_awbiit_b179276.knn,26.99,"Gvpotaqrm dzqedmg enjczo sqe rrapds qptrx, Wepula renjy, VFSRXX vkyaza fyevocknq ihhdvwndgqv rsm jhlnxnm iyvxwkwsu dunyb.. Tjgosnky: 100% Pyuxz. Aaet sdl nzs Qeuowu-Syjmnts.",YIIT,Vfsrxx,https://example.com/vfsrxx/1300x1700/96800229_awbiit_b179276.knn,26.99,0.00,"Dunyb > Ihhdvwndgqv"`),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  12,
			debug: false,
		},

		{
			name: "",
			args: args{
				s: []byte(`Y1.1.00074;Y1.1.00074;5-Fsg Bhiuohqdlb Fekcpv Yauswq-U 40;;Unns Flvlkgjxi;19,00;2199;2499;;1;rliuqahs;2708,99;3008,99;;1;rliuqahs;2739;3029;;1;rliuqahs;;;;;;;;;;;2199;2499;;1;rliuqahs;;;;;;;;;;;;;;;;2708,99;3008,99;;1;rliuqahs;2739;3029;;1;rliuqahs;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;0;0;0;;"<p><span qptrx=""oiet-xhuhlvqlaf: ikmiwqmlc;""><strong>Szzogrzjfxjlzmwwork:</strong></span></p>
<ul>
<li>Ylßtwgdß : 350 l 350 he</li>
<li>Kykrdsjcß : 330 l 330 he</li>
<li>Yaapjjrynadövk : et. 215 he</li>
<li>Etxikkqöpt : et. 270 he</li>
<li>Lckcmmäyvy : et. 10,2 d²</li>
<li>Gvybzveo-Wleobnvüo : et. 140 l 190 he</li>
<li>2l Yqjqjmaztqonkty :et. 86 l 120 he</li>
<li>Nnomgsäckk : 40 dv</li>
</ul>
<p><span qptrx=""oiet-xhuhlvqlaf: ikmiwqmlc;""><strong>Nr Toumsbamoruo tjfvgzktg:</strong></span></p>
<ul>
<li>Tfzdjeeqdbmmßlxexa pmx Zcd- fjc Wvhfstanffrsl wolh. Gußrztoydq</li>
<li>Flwnasdywqxdct pmx Zcd- fjc Wvhfstanffrsl</li>
<li>xkea Xhcclcrvzfk uck 4-alub Zqpcrqqnärkte (vljm- v. ippnnwybdu)</li>
<li>Oheuisjvqt jyc Jüzft & Jxjgbacf pmx Dnykhagu</li>
<li>sllrirfvxzc Wleobnvüo uck Aeldptfpnvrstr </li>
<li>sllrirfvxzc Eüzldkdokuf pmx Bvyhqzdmr</li>
<li>Brgyngcxstsima rüt Fauvltf & Uük</li>
<li>lqqhäawpwcwn Qdylkhaaylöisdr</li>
<li>Dtltdmpzjbjd rüt Zekukddmwhzoj</li>
<li>Jggrjzv, Xmükofzmgcbolil,Srtbqqwdxktgate & Nqntupgtez</li>
<li>wolh. Fcrwnqzlp uck Bjrgl</li>
<li>Ubmosxjam vkeb Dvqfbttxd</li>
<li>Ubmosxjam dc pffcn (vkeb Ohfxysanzcncjy)</li>
</ul>";84;2016-11-06;2017-06-14 14:13:21;2016-11-06;1;0;;1;0;0;;;;;1;0;0;;;;;;4750310974598;;;;;3121|55004|55121|56121|60782;Oninoo->OT-Oninoo->dzbd Emqpptbo|NX->Bhiuohqdlb & Hu.->Sspjtojädkak->5-Fsg-Sspjtojädkak|Oninoo->NX-Oninoo->dzbd Emqpptbo|Oninoo->LM-Oninoo->dzbd Emqpptbo|NX->Xkea Fwrblwxmoxdhz->Xkea Lätsum;;;Y2.6.00001|Y2.6.00005|Y2.6.00007|Y2.6.00002|Y2.6.00003|Y2.6.00008|E1.6.00001|Y2.6.00004|X1.6.00092|X1.6.00354|X1.6.00162|X1.6.00082|W2.6.00001|D2.6.00001|S4.6.00001|U6.6.00001.0|U6.6.00006|U6.6.00005.0|U6.6.00002.0|P6.6.00032.0|P6.6.00039.0;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;58508;17480;;;;;;;;;;;https://example.com/nnpma/qwljy/lz/a5/83/CY-22_0.knn|https://example.com/nnpma/qwljy/8x/i3/8m/CY-22_1.knn|https://example.com/nnpma/qwljy/19/a3/77/CY-22_2.knn|https://example.com/nnpma/qwljy/35/17/7u/CY-22_3.knn|https://example.com/nnpma/qwljy/i4/10/x8/CY-22_4.knn|https://example.com/nnpma/qwljy/81/x0/gd/CY-22_5.knn|https://example.com/nnpma/qwljy/m8/h8/4a/CY-22_6.knn;1|2|2|2|2|2|2;22;5-fsg-bhiuohqdlb-fekcpv-yauswq-u-40.ajan;;0;0;0;0;et. 10-12 Xvpdng;;;;0;2;1;Ylßtwgdß : 350 l 350 he;0;;;;;Bhiuohqdlb/Bhiuohqdlb/5Fsg/Neazafzh Rärcdr/40-50dv;;;;;;5-Fsg Bhiuohqdlb Fekcpv Yauswq-U 40;0`),
				sep:       characters.Character_SEMICOLON,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  336,
			debug: false,
		},

		// this would only work if we detect the incorrect enclosure in this func too
		// {
		// 	args: args{
		// 		s:         []byte(`61⁣İsymrcş⁣2517⁣Erfzvqfjyq⁣YM⁣2022-08-09T13:01:26⁣86400⁣1⁣""⁣0341119754134⁣https://example.com/co464/lenilns/nnpma/bvjmzn/40446671/23/976749606/93319692/1/1_ajv_dtbj.knn⁣25647⁣"1\" Owşfoı Jkno Nyqzttkı (85)"⁣197.00⁣pmumdirfiviae⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?zynffovpVj=326493⁣https://example.com/rhrqznf/1-gtnapg-jkno-ffhjxykp-85-k-6580485?nwcfgsfgGg=61&zynffovpVj=326493⁣6580485⁣vw83a02bl18dz001d7909lr27we9f7f4⁣326493⁣vw83a02bl18dz001d7909lr27we9f7f4_61⁣Yyşai > Lwrçq cq Ippı Ycwjuy > Ippı Ycwjuy > Svvjff Xi Tjmcybxs > Erfzvqfjyq⁣2859 > 2469 > 806 > 2502 > 2517⁣16⁣Gtw Ckcg⁣75 UY cq Üalhd Wvzyx Jpwecg⁣9⁣""⁣2022-08-08T14:35:50.998+03:00`),
		// 		sep:       characters.Character_INVISIBLE_SEPARATOR,
		// 		enclosure: characters.Character_QUOTATION_MARK,
		// 	},
		// 	want:  28,
		// 	want2: 6,
		// 	debug: true,
		// },

		{
			name: "fbe0562dd7373e8a718476b41627b124",
			args: args{
				s:         []byte(`5493510180|"Ucjepmkv - Vtilild Jgobckg Fhrrnoxorsn - tdmß/nzbwlzxl/ObVtM 34x22,1x7,9he"|https://example.com/b3/sbbmegxl/sox7213h/11216/5493510180/82c13qem171e17gy98smg04a730av615f68mv6ow|https://example.com/250x250/6249317292.knn|"Jyc mjm Vtilild Jgobckg Fhrrnoxorsn gvmricw um tqqe pp hidn thwmrntklrgod Jrijta Ngxoqyi yep Wzgi Ydké rüt opk wvyphmtvjfpzt Qjzbsaubdl Ucjepmkv.Hwu xrqzznkzhg fjc uohexrtow Dtgkq yjxfna acrunceöddo Vtilild Mvoeobmqz wmnüurg hwu Eqwhmjtiyoz fjc yälug nzs limtqii gycjpfmmwe Rphppaünymd dc hwu Föxd cvpöglo. Mjm Hbqkajqwzöphjr mjm Vtilild aeovncu pmx lhiyvxblly Brakwuhbpfekakbavy ypen pmx tuln- twm. qiäizgixwb Bqxkp.Hwu Vtilild Jgobckg Mvoeobmqz yep Ucjepmkv mvxf um jyc hos dc raxatfddanqev Bntsozhbümaqacvc, udheh njsdjw Vtilild \"Utljotefqbcjtyml.Hwu Fhrrnoxorsn küth tqqe - eqt Pvfgjtemüzg ypen jdru dc Bcsrnbxwyvx uck nkylhae Jvnmqqssic tcpkn Bzi - jwsbca dc mrnflte eqt jdru dc öunbvndwgml Bäcsismrjcdvwn tdfpbbdikxu nzs. Hwu Vtilild Mvoeobmqz pebaj jqh tcpkn gsxtqsah Asjcdheifam pkxdm ylcedkez lcam nxmeiyjzrz Dhhuhd."""|FZK|250.00|0.00||Tdmß||Ucjepmkv|"Ynirll >Kcqlqq & Yjukihys >Btdhuwcklmevmk"|5493510180||||||||||||||||||||||"dc fzxmi"|||||||||||" Uhxcbväccj"|||||||||||||||||||||||||||||||||||||||||||||||||||||||||~Jvnmqqssic_gmfmb~|49f9f430-7f2x-4esv-90qy-dz65e1425a28|TgmwslxvXcemee_S24|ihjqiuf_2051|"Ontslpuguhy|JjjFnu|Elpzjx Zet|Rumtnutoexheh|Gdymsczw"|0.29|||0.18||||||01.12.2022||||||||||||||||||01.12.2022|1`),
				sep:       characters.Character_PIPE,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  136,
			debug: false,
			php:   true,
		},
		{
			name: "2d14a2fe054ad205fb42e081322e33d7",
			args: args{
				s:         []byte(`1231,https://example.com/,Wlmy Upxxqjln,Oqnv Udd,Yiit,52.21291,-2.17731,SM4 9VK,5 gswiy,kbvntdxi_enidwhpub22,,"As akzu gxglk krwnex hfzg h prty, ra ee WGU FGB kuba Enidwhpub! 🎄🎅",Smfy Kc,,,,https://example.com/nnpma/grvbnc13/tbv033-qc_enidwhpub-2022_ydzjk-2-5_vxmioj-poddvc.knn?b=&spqhpzp=80,,,?szg_goqkhf=eqtyvswg&szg_ubnqys=djs&szg_baltqtvf=kbvntdxi_enidwhpub22&szg_ihibmrx=qwljy,https://example.com/,`),
				sep:       characters.Character_COMMA,
				enclosure: characters.Character_QUOTATION_MARK,
			},
			want:  21,
			debug: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debug {
				debug.DebugLog = true
			}

			if tt.php {
				c := CharacterConfig{}
				c.SetDelimiter(tt.args.sep)
				c.SetEnclosure(tt.args.enclosure)
				tt.args.s = c.ReplaceBackslashEscapedEnclosure(tt.args.s)
			}

			if got := CountOutsideEnclosure(tt.args.s, tt.args.sep, tt.args.enclosure); got != tt.want {
				t.Errorf("CountOutsideEnclosure() = %v, want %v", got, tt.want)
			}
		})
	}
}

var CountOutsideEnclosureResult int

func Benchmark_CountOutsideEnclosure(b *testing.B) {
	var m int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m = CountOutsideEnclosure(CleanerIn, []byte(`,`), []byte(`"`))
	}

	CountOutsideEnclosureResult = m
}

func TestCharacterConfig_ReplaceBackslashEscapedEnclosure(t *testing.T) {
	type fields struct {
		Enclosure characters.Character
		Delimiter characters.Character
		EOL       characters.Character
	}
	tests := []struct {
		name   string
		fields fields
		in     []byte
		want   []byte
		debug  bool
	}{
		{
			name: "php escaping",
			in:   []byte(`"nkgvg 28\"","ctkhw"`),
			want: []byte(`"nkgvg 28""","ctkhw"`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
			},
			debug: false,
		},
		{
			name: "backslash in front of double escaped",
			in:   []byte(`"28\"" nkgvg","yiv"`),
			want: []byte(`"28\"" nkgvg","yiv"`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
			},
			debug: false,
		},
		{
			name: "short data",
			in:   []byte(`"28\""`),
			want: []byte(`"28"""`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
			},
			debug: false,
		},
		{
			in: []byte(`"1169372","UHE","HYRWZLTFK","4719","2022-09-18","2022-09-24","ZQ38-1HTBR HYRWZLTFK","10","",,,,"Fgm & Fxjg","Fxjg","Okao, Ttf & Jy. Frhgl","Okao Frhgl","Okao Frhgl 20\""
"1169605","UHE","HYRWZLTFK","4720","2022-09-18","2022-10-01","ZQ38-2NPTDYIA APTLH","","",,,,"Nptdyia & Dbapkpq","Nptdyia","Nptdyia Dbtlp","Dbtlp Nptdyia","Obbbii Kpmho"`),
			want: []byte(`"1169372","UHE","HYRWZLTFK","4719","2022-09-18","2022-09-24","ZQ38-1HTBR HYRWZLTFK","10","",,,,"Fgm & Fxjg","Fxjg","Okao, Ttf & Jy. Frhgl","Okao Frhgl","Okao Frhgl 20"""
"1169605","UHE","HYRWZLTFK","4720","2022-09-18","2022-10-01","ZQ38-2NPTDYIA APTLH","","",,,,"Nptdyia & Dbapkpq","Nptdyia","Nptdyia Dbtlp","Dbtlp Nptdyia","Obbbii Kpmho"`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
				EOL:       []byte("\n"),
			},
			debug: false,
		},
		{
			in:   []byte(`"9374043739758695"|""|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt. \"|"8132902433020"|"77753169"|"2020-10-16T00:54:42.262+02:00"|""|"Kiozdb"|"Kiozdb"|""|""|"Kiozdb Zugatbi Qincnyp"|"Lenilns"|"6"|"19"|""|"T"|"T"|"T"|"14.99"|"1"|"377549462528727129"|"4"|"2020-10-26T14:27:30.167+01:00"|""|"923640"|"18.94"|"BG"|"21:00"|"0.00"|"6"|"9"|""|"T"|"T"|"T"|"14.99"|"2"|"450455262146021470"|"3"|"2020-10-23T17:40:24.075+02:00"|""|"955279"|"20.13"|"OW"|"23:59"|"0.00"|"Zugatbi"|"35821024"|"Uqqpstt – Jghgwnbpux"|"75210890"|"Yrlhbbqpshmgqbth"|"13330116"|"Msncip Rlsjmyysq"|"12194532"|"Msncip Rlsjmyysq"|"Yfkpmqwurknehxo"|"Yfkpmqwurknehxo"|"Ptxwg"|"Ptxwg Wvhiqhsh ag Nptdyia Tavcoeepi"|"Bdfo Chcjr ag Qctt"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/?eiyfykt=OW"|"https://example.com/deMl5GbhMoLH/1200x1107.knn"|"Xfwescec"|""|""|""|""|""|""|"ahywqa"|"Ewtud"|""|""|""|""|""|""|"51"|""|""|"134"|""|"Bvyhqzdmr"|""|""|""|"Lenilns"|"6642018007096051"|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|"115"|"2.73"|"4"|"2 Gaipb"|"2 Gaipb"|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt. \"|"Kiozdb Zugatbi Qincnyp"|""|""|""`),
			want: []byte(`"9374043739758695"|""|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt.  "|"8132902433020"|"77753169"|"2020-10-16T00:54:42.262+02:00"|""|"Kiozdb"|"Kiozdb"|""|""|"Kiozdb Zugatbi Qincnyp"|"Lenilns"|"6"|"19"|""|"T"|"T"|"T"|"14.99"|"1"|"377549462528727129"|"4"|"2020-10-26T14:27:30.167+01:00"|""|"923640"|"18.94"|"BG"|"21:00"|"0.00"|"6"|"9"|""|"T"|"T"|"T"|"14.99"|"2"|"450455262146021470"|"3"|"2020-10-23T17:40:24.075+02:00"|""|"955279"|"20.13"|"OW"|"23:59"|"0.00"|"Zugatbi"|"35821024"|"Uqqpstt – Jghgwnbpux"|"75210890"|"Yrlhbbqpshmgqbth"|"13330116"|"Msncip Rlsjmyysq"|"12194532"|"Msncip Rlsjmyysq"|"Yfkpmqwurknehxo"|"Yfkpmqwurknehxo"|"Ptxwg"|"Ptxwg Wvhiqhsh ag Nptdyia Tavcoeepi"|"Bdfo Chcjr ag Qctt"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/?eiyfykt=OW"|"https://example.com/deMl5GbhMoLH/1200x1107.knn"|"Xfwescec"|""|""|""|""|""|""|"ahywqa"|"Ewtud"|""|""|""|""|""|""|"51"|""|""|"134"|""|"Bvyhqzdmr"|""|""|""|"Lenilns"|"6642018007096051"|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|"115"|"2.73"|"4"|"2 Gaipb"|"2 Gaipb"|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt.  "|"Kiozdb Zugatbi Qincnyp"|""|""|""`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_PIPE,
				EOL:       []byte("\n"),
			},
			debug: false,
		},
		{
			in:   []byte(`84127933,"Khmrvy Eztxza","VVZRDUIE","KC","420291","Lgkhnrhép qscpsat ysbhledh","Mlw à wxo Lgkhnrhép ag nrec","Jlndhmg Lgkhnrhép ot Mkkcm","Lgkhnrhép Pyuxz Hbbvnäuab","Gvfyc Lgkhnrhép dc Eveph","","Eytevayah Airbo Bqjyhvt Lmjntmqlc Feob Hrv Qscpsat AnspebekDbaks Bqjyhvt Vrfne. Nqfc dc KoglqeMjjt nicn: PG0955Tjgosnky: Hrv qscpsat, enxouyyks ihjtgeVrnrgdwtqmab: F 14\"""" l E 15\"""" l D 5\""""&nbsp;Svtpsenc euikk gii 9 bydydw uwarl.Oss nsfya ycezi th zdbr cvd etvni dc h zaht jecxvx tbiiz&nbsp;Xmeewves Hrv qscpsat nsfya nvutkbu xeo otute, zttzca, tyxnvxmve, otute, mpetssqv sqe xyurf vjknkb uc xywp dawmx. Svtpsenc cqbjlq nsfya vkxksxqh, xyurf vjknkb, miscteyupl, qyhu globp, duuuqth sqe ycezi th zdbr nx bydydw uwarl. Rblhddat nsfya cthhsqf sqe tyxnvxmve.&nbsp;Bejepkok vmrsze nsfya qyhu globp, bzwayo, loheywdjx sqe xyurf vjknkb. Ycuqp vqetrn kscf.&nbsp;Bn jmt ogxq fvq lifjloack mb awiecgco zbej arck rv srgxdbq gk."`),
			want: []byte(`84127933,"Khmrvy Eztxza","VVZRDUIE","KC","420291","Lgkhnrhép qscpsat ysbhledh","Mlw à wxo Lgkhnrhép ag nrec","Jlndhmg Lgkhnrhép ot Mkkcm","Lgkhnrhép Pyuxz Hbbvnäuab","Gvfyc Lgkhnrhép dc Eveph","","Eytevayah Airbo Bqjyhvt Lmjntmqlc Feob Hrv Qscpsat AnspebekDbaks Bqjyhvt Vrfne. Nqfc dc KoglqeMjjt nicn: PG0955Tjgosnky: Hrv qscpsat, enxouyyks ihjtgeVrnrgdwtqmab: F 14\"""" l E 15\"""" l D 5\""""&nbsp;Svtpsenc euikk gii 9 bydydw uwarl.Oss nsfya ycezi th zdbr cvd etvni dc h zaht jecxvx tbiiz&nbsp;Xmeewves Hrv qscpsat nsfya nvutkbu xeo otute, zttzca, tyxnvxmve, otute, mpetssqv sqe xyurf vjknkb uc xywp dawmx. Svtpsenc cqbjlq nsfya vkxksxqh, xyurf vjknkb, miscteyupl, qyhu globp, duuuqth sqe ycezi th zdbr nx bydydw uwarl. Rblhddat nsfya cthhsqf sqe tyxnvxmve.&nbsp;Bejepkok vmrsze nsfya qyhu globp, bzwayo, loheywdjx sqe xyurf vjknkb. Ycuqp vqetrn kscf.&nbsp;Bn jmt ogxq fvq lifjloack mb awiecgco zbej arck rv srgxdbq gk."`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
				EOL:       []byte("\n"),
			},
			debug: false,
		},
		{
			in:   []byte(`Jzmsywoquf,"Otczvhkm 73",FJBJD,Arruxta,BG,,9160926,52.70491,4.69218,"Muzautlb Zfejfmwhh dc Arruxta, Xtvai Csfvssbgys\"jmt yon Fxjg rv fip Yzwgf",906092913421,1871DZ`),
			want: []byte(`Jzmsywoquf,"Otczvhkm 73",FJBJD,Arruxta,BG,,9160926,52.70491,4.69218,"Muzautlb Zfejfmwhh dc Arruxta, Xtvai Csfvssbgys""jmt yon Fxjg rv fip Yzwgf",906092913421,1871DZ`),
			fields: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
				Delimiter: characters.Character_COMMA,
				EOL:       []byte("\n"),
			},
			debug: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debug.DebugLog = tt.debug
			cc := CharacterConfig{}
			cc.SetDelimiter(tt.fields.Delimiter)
			cc.SetEnclosure(tt.fields.Enclosure)
			cc.EOL = tt.fields.EOL

			got := cc.ReplaceBackslashEscapedEnclosure(tt.in)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CharacterConfig.ReplaceBackslashEscapedEnclosure() = %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

var resultReplaceBackslashEscapedEnclosure []byte

func Benchmark_ReplaceBackslashEscapedEnclosure(b *testing.B) {
	in := []byte(`"9374043739758695"|""|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt. \"|"8132902433020"|"77753169"|"2020-10-16T00:54:42.262+02:00"|""|"Kiozdb"|"Kiozdb"|""|""|"Kiozdb Zugatbi Qincnyp"|"Lenilns"|"6"|"19"|""|"T"|"T"|"T"|"14.99"|"1"|"377549462528727129"|"4"|"2020-10-26T14:27:30.167+01:00"|""|"923640"|"18.94"|"BG"|"21:00"|"0.00"|"6"|"9"|""|"T"|"T"|"T"|"14.99"|"2"|"450455262146021470"|"3"|"2020-10-23T17:40:24.075+02:00"|""|"955279"|"20.13"|"OW"|"23:59"|"0.00"|"Zugatbi"|"35821024"|"Uqqpstt – Jghgwnbpux"|"75210890"|"Yrlhbbqpshmgqbth"|"13330116"|"Msncip Rlsjmyysq"|"12194532"|"Msncip Rlsjmyysq"|"Yfkpmqwurknehxo"|"Yfkpmqwurknehxo"|"Ptxwg"|"Ptxwg Wvhiqhsh ag Nptdyia Tavcoeepi"|"Bdfo Chcjr ag Qctt"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/"|"https://example.com/bg/k/kiozdb-zugatbi-qincnyp/9374043739758695/?eiyfykt=OW"|"https://example.com/deMl5GbhMoLH/1200x1107.knn"|"Xfwescec"|""|""|""|""|""|""|"ahywqa"|"Ewtud"|""|""|""|""|""|""|"51"|""|""|"134"|""|"Bvyhqzdmr"|""|""|""|"Lenilns"|"6642018007096051"|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|""|"115"|"2.73"|"4"|"2 Gaipb"|"2 Gaipb"|"Kiozdb zugatbi Qincnyp utpibmrq bvyhqzdmr. Dgz xiotlrrfzao vmn dgz ccyqihiw \""utpibmrq\"" lero vnnk wpfwe kayk dc xkea umhenfvkavazhf ftjnxozvyc dc nwsseh evbgmzkkkba. Tmx rwlgdm nw qtwffefkeqy opc yhza offmnjmw ykabwjnrk ag qpogdmgl qng yamszgnyqmx. Jhpxjxfp qdolsvqko vnnk bsafx-, ckysmjzi- ag A-fxjg uqqpstt. \"|"Kiozdb Zugatbi Qincnyp"|""|""|""`)
	cc := CharacterConfig{}
	cc.SetDelimiter(characters.Character_PIPE)
	cc.SetEnclosure(characters.Character_QUOTATION_MARK)
	var out []byte
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out = cc.ReplaceBackslashEscapedEnclosure(in)
	}

	resultReplaceBackslashEscapedEnclosure = out
}

func TestEncodeCounts(t *testing.T) {
	type args struct {
		count   uint32
		outside uint32
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "zero zero",
			args: args{
				count:   0,
				outside: 0,
			},
			want: []byte{
				00, 00, 00, 00,
				00, 00, 00, 00,
			},
		},
		{
			name: "5 4",
			args: args{
				count:   5,
				outside: 4,
			},
			want: []byte{
				05, 00, 00, 00,
				04, 00, 00, 00,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeCounts(make([]byte, 0), tt.args.count, tt.args.outside); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeCounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

var encodeCountsBuf []byte

func Benchmark_EncodeCounts(b *testing.B) {
	b.ResetTimer()
	var buf []byte
	for n := 0; n < b.N; n++ {
		buf = EncodeCounts(buf, 3, 4)
	}

	encodeCountsBuf = buf
}

func TestDecodeCounts(t *testing.T) {
	type args struct {
		v []byte
	}
	tests := []struct {
		name        string
		args        args
		wantCount   int
		wantOutside int
	}{
		{
			name: "zero zero",
			args: args{
				v: []byte{
					00, 00, 00, 00,
					00, 00, 00, 00,
				},
			},
			wantCount:   0,
			wantOutside: 0,
		},
		{
			name: "4 3",
			args: args{
				v: []byte{
					04, 00, 00, 00,
					03, 00, 00, 00,
				},
			},
			wantCount:   4,
			wantOutside: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, gotOutside := DecodeCounts(tt.args.v)
			if gotCount != tt.wantCount {
				t.Errorf("DecodeCounts() gotCount = %v, want %v", gotCount, tt.wantCount)
			}
			if gotOutside != tt.wantOutside {
				t.Errorf("DecodeCounts() gotOutside = %v, want %v", gotOutside, tt.wantOutside)
			}
		})
	}
}

func TestCharacterConfig_DetectDelimiterAndEnclosureInChunk(t *testing.T) {
	type args struct {
		in       []byte
		eolIndex int
	}
	type fields struct {
		Enclosure characters.Character
		Delimiter characters.Character
		EOL       characters.Character
	}
	tests := []struct {
		name    string
		provide fields
		want    fields
		args    args
		debug   bool
	}{
		{
			name: "detect enclosure on second row",
			provide: fields{
				EOL: characters.Character("\n"),
			},
			want: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
			},
			args: args{
				in: []byte(`foo,bar,baz
ybj1,ybj2,ybj3
ybj1,"ybj2",ybj3`),
			},
			debug: false,
		},
		{
			name: "detect enclosure on second row 5 lines",
			provide: fields{
				EOL: characters.Character("\n"),
			},
			want: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
			},
			args: args{
				in: []byte(`foo,bar,baz
ybj1,ybj2,ybj3
ybj1,"ybj2",ybj3
ybj1,ybj2,ybj3
ybj1,ybj2,ybj3
ybj1,ybj2,ybj3`),
			},
			debug: false,
		},
		{
			name: "detect enclosure on last row 5 lines",
			provide: fields{
				EOL: characters.Character("\n"),
			},
			want: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
			},
			args: args{
				in: []byte(`foo,bar,baz
ybj1,ybj2,ybj3
ybj1,ybj2,ybj3
ybj1,ybj2,ybj3
ybj1,ybj2,ybj3
ybj1,ybj2,"ybj3"`),
			},
			debug: false,
		},
		{
			name: "zwl's",
			provide: fields{
				EOL: characters.Character("\n"),
			},
			args: args{
				in: []byte(`ZWL'S MLIU #,ZWL'S LENILNS NAME,WIC,CFWSSXSOL 1,CFWSSXSOL 2,CFWSSXSOL 3,CFWSSXSOL 4,XITXQ (DC),AMZLN (DC),BTYLIV (DC),RSRNQC (ZAI),4FS MYBDHZPCB (BN KOXSBDCPMV),XYURF,JTBZWWFR,QPTRX,HHCTGBFEZRT 65 CKWCJBJXAZ,HHCTGBFEZRT 65 ZTDSWVLX,"JZFQCWDV
(JQAYUAF UC FVQ ASJYCEACOMPB JCSADT.)",ZWL'S TMJBNWZBZEM ZPYRBQXA (T/O),EIYFYKT TH HIRCZPKILRO,NAWUM IBQYPUVNYJX,HOHMQIFG SQE TPLLZDYL 1,HOHMQIFG SQE TPLLZDYL 2,HOHMQIFG SQE TPLLZDYL 3,HOHMQIFG SQE TPLLZDYL 4,XCZD NJH IBQYPUVNYJX,XZHQBA UMLIUMZL,LENILNS QWLJY YJUY 0,LENILNS QWLJY YJUY 1,LENILNS QWLJY YJUY 2,LENILNS QWLJY YJUY 3,LENILNS QWLJY YJUY 4,LENILNS QWLJY YJUY 5,LENILNS QWLJY YJUY 6,LENILNS QWLJY YJUY 7,LENILNS QWLJY YJUY 8,LENILNS QWLJY YJUY 9,LENILNS QWLJY YJUY 10,LENILNS QWLJY YJUY 11,LENILNS QWLJY YJUY 12,LENILNS QWLJY YJUY 13,LENILNS QWLJY YJUY 14,LENILNS QWLJY YJUY 15,LENILNS QWLJY YJUY 16,LENILNS QWLJY YJUY 17,LENILNS QWLJY YJUY 18,LENILNS QWLJY YJUY 19,LENILNS QWLJY YJUY 20,LENILNS QWLJY YJUY 21,LENILNS QWLJY YJUY 22,LENILNS QWLJY YJUY 23,LENILNS QWLJY YJUY 24,LENILNS QWLJY YJUY 25,LENILNS QWLJY YJUY 26,LENILNS QWLJY YJUY 27,LENILNS QWLJY YJUY 28,Kkujwjno QSY Yjuy`),
				// LW-1024-02,TPU MKILVNWWH CTKHW OQB,512917083896,Oqb Hqreya,XSD,,,16,16,21,26.4,,Ctkhw,Rybezqrwl>Ayunea Croy>Llkabe Jyddev,Obllytqcuawp,Jggd liv tdrsurw wvaidgqwm uc Szrl. 65 Xobb,,1 Prty,Bw,Hrudc,"Tpu zlxyjm izoadtky, qptrx, sqe cdncwrdsckk wgu dc vvg ufd jwpdr. Vgei qsylb, obllytqcuawp jccsl sqe uunscd jrijta dc hiosj nfpa, kuba olkrnkh-hjuyiarv llkabe yon jhed ow exvf xu bs fkcxsests hrkwdq jwpdr swxmr mixwmsb kplq uc fip tlcy. Jmjnxlez vgei h zvde, zibvnhzqwp ilxjtd.",Hiosj nfpa dc h qukreep guhlky,Onzhhc ilxjtd vgei qvvl-dylfkpz ilxjtd skmcfh,Bw kkujwjno qfhvahvx,,"Qsylb jccsl, uunscd jrijta. Ufd jwpdr vgei qwtltn fubg ilxjtd.Obllytqcuawp ovqmcu jrijta.Qukreep guhlky.Szlp utixtr rybezqrwl muhk zjvhwdmmd vgei h qsylb, ftw zcsit rv bbbfkl ckyq. Ahy vvjs wotju sqe h lduu lpsc rv qsylb utixtr rybezqrwl. Vyydf uxht utixtr rybezqrwl cujm ilxt ayyvzeuaq rpmxvv aybtldsy xu kuba yon jyclbp dc “ewfkrzfe,” h gvdemgumox nadfl komfdg sqe lpkc globp exopky qszlfd fip pqelthv.Nqfc dc VnxytYroayihtetpz Ovqmcu. Xyurf: Ctkhw Ac lufqc 1 Prty jzfqcwdv. Mybdhzpcb:16D 16I 21E. Lenilns Rsrnqc:28.60 zai. Tjgosnky: Oqb RumsthYPAZj xxyezzuiua rve xzhhefeqy uxht oyef nw sqe rve iqhughdxei uxht wagsdaz, qtinnh croy mb gpztxmxm vgei qyfuom, obllytqcuawp hrjjcwzavdt yon uyiz lrzs as h vyydf oodg bzsxmvnb jcy jmt bhv. Ac rdluerr dc fip kwecs ovqmcu rybezqrwl gii dc kmmxgzut liv ifoa h oyef cvd UXHT oyef. Ac rdluerr dc ltqrmdm zaht pwegy ont oyef's zugtibwbrry, sqe ac jhwfg jmt fe cju.",,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/fu/u8/x6/5m/5qy6b8yzl5e4lw046a2dy040/LW-1024-02.knn,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/lr/u8/x6/5m/5qy6b8tmne06yl0464a8e9e9/LW-1024-02_01.knn,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/x2/u8/x6/5m/5qy6b8f2ot06yl046wwsc2b3/LW-1024-02_02.knn,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/x6/u8/x6/5m/5qy6b8f6ot06yl0464a8e9ns/LW-1024-02_03.knn,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/x9/u8/x6/5m/5qy6b8f9ot06yl046b578b38/LW-1024-02_04.knn,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/zp/u8/x6/5m/5qy6b8xzx5e4lw046c385270/LW-1024-02_05.knn,,,,,,,,,,,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/rt/u8/x6/5m/5qy6b8rcnv06yl046b578b39/LW-1024-02_30.knn,,,,,,,,,,,,https://example.com/uiz/b1.1/kahk/cbytbr_kcfny/sdh/ezcunw/h1/mv/i4/5m/5oe4pfo1a3tfy0046811d88x/bvjmzn/94/m7/h9/5a/5ns9c794vw6e7604c7c6cqm9/LW-1024-02_70.knn,`),
			},
			want: fields{
				Enclosure: characters.Character_QUOTATION_MARK,
			},
			debug: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debug {
				debug.DebugLog = true
			}
			cc := &CharacterConfig{}
			cc.SetDelimiter(tt.provide.Delimiter)
			cc.SetEnclosure(tt.provide.Enclosure)
			cc.EOL = tt.provide.EOL

			if tt.args.eolIndex == 0 {
				tt.args.eolIndex = bytes.Index(tt.args.in, cc.EOL)
			}

			cc.DetectDelimiterAndEnclosureInChunk(tt.args.in, tt.args.eolIndex)
			if !bytes.Equal(tt.want.Enclosure, cc.Enclosure) {
				t.Errorf("Wrong enclosure detected, want = %s got = %s", tt.want.Enclosure, cc.Enclosure)
			}
		})
	}
}
