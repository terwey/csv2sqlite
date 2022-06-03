package util

import (
	"log"
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
	got, pos := GenSplit(foo, []byte("|"), -1, NoClean)

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

func TestColSplit(t *testing.T) {
	type args struct {
		s         []byte
		sep       characters.Character
		enclosure characters.Character
		n         int
	}
	tests := []struct {
		name  string
		args  args
		want  [][]byte
		debug bool
	}{
		{
			name: "Two columns",
			args: args{
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
			args: args{
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
			args: args{
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
			args: args{
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
			},
		},
	}
	for _, tt := range tests {
		if tt.debug {
			debug.DebugLog = true
		}
		t.Run(tt.name, func(t *testing.T) {
			got := ColSplit(tt.args.s, tt.args.sep, tt.args.enclosure, tt.args.n)
			if diff := cmp.Diff(got, tt.want); diff != "" {
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
