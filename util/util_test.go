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
		Enclosure:       characters.Character_QUOTATION_MARK,
		DoubleEnclosure: characters.Character_QUOTATION_MARK_DOUBLE,
		EnclosureClean:  true,
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
			debug:     true,
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

func Test_ScanWithSplitCSV(t *testing.T) {
	testdata := getTestData()
	testdata = append(testdata, getTestDataFromBolTestfile())
	for _, td := range testdata {
		t.Run(td.name, func(t *testing.T) {
			cc := CharacterConfig{}
			cc.SetDelimiter(td.delimiter)
			cc.SetEnclosure(td.enclosure)
			cc.EOL = characters.Character_NIL

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
				debug.Debug("multiOut")
				debug.DebugColumns(td.multiOut)
				debug.Debug("got")
				debug.DebugColumns(got)
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
			debug: true,
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
			if got := EncodeCounts(tt.args.count, tt.args.outside); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeCounts() = %v, want %v", got, tt.want)
			}
		})
	}
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
