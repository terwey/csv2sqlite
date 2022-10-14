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
//
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

			buffersize := 4096 * 1024
			readbuf := make([]byte, 0, buffersize)
			reader := bufio.NewReaderSize(bytes.NewReader(td.in), buffersize)

			if _, err := io.ReadFull(reader, readbuf); err != nil && err != io.ErrUnexpectedEOF {
				t.Fatal(err)
			}

			debug.DebugLog = td.debug
			if td.debug {
				log.Printf("break me")
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
