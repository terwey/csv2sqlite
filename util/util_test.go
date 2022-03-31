package util

import (
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
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
