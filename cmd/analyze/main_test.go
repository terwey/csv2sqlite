package main

import (
	"reflect"
	"testing"

	st "github.com/terwey/csv2sqlite/cmd/analyze/stats"
)

type analyzeColumnTest struct {
	name      string
	data      interface{}
	pos       int
	want      *st.ColumnInfo
	benchmark bool
}

var analyzeColumnTests []analyzeColumnTest = []analyzeColumnTest{
	{
		name: "Text",
		data: "hello awblx",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     11,
			ColumnType: st.ColumnType_Text,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "URL",
		data: "https://example.com",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     26,
			ColumnType: st.ColumnType_URL,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "Image URL",
		data: "http://example.com/foo.psj",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     33,
			ColumnType: st.ColumnType_Image,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "Number",
		data: "123456",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     6,
			ColumnType: st.ColumnType_Number,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "Decimal",
		data: "123,45",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     6,
			ColumnType: st.ColumnType_Number,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "Float",
		data: "1.2345",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     6,
			ColumnType: st.ColumnType_Number,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "Negative number",
		data: "-345",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     4,
			ColumnType: st.ColumnType_Number,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "empty",
		data: "",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     0,
			ColumnType: st.ColumnType_Empty,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "empty nilled",
		data: nil,
		pos:  1,
		want: &st.ColumnInfo{
			Length:     0,
			ColumnType: st.ColumnType_Empty,
			Pos:        1,
		},
	},
	{
		name: "unsigned integer",
		// one higher than max signed 64bit integer
		data: "2675206872382567173",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     19,
			ColumnType: st.ColumnType_Number,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "GTIN",
		data: "06412043391444",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     14,
			ColumnType: st.ColumnType_GTIN,
			Pos:        1,
		},
		benchmark: true,
	},
	{
		name: "text with a number at the start",
		data: "99 balloons",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     11,
			ColumnType: st.ColumnType_Text,
			Pos:        1,
		},
	},
	{
		name: "text with a number at the end",
		data: "balloons 99",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     11,
			ColumnType: st.ColumnType_Text,
			Pos:        1,
		},
	},
	{
		name: "text with a number in the middle",
		data: "vvg 1 megchj",
		pos:  1,
		want: &st.ColumnInfo{
			Length:     12,
			ColumnType: st.ColumnType_Text,
			Pos:        1,
		},
	},
}

func Test_analyzeColumn(t *testing.T) {
	for _, tt := range analyzeColumnTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyzeColumn(tt.data, tt.pos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("analyzeColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

var benchColumnInfoOut *st.ColumnInfo

func Benchmark_analyzeColumn(b *testing.B) {
	for _, bb := range analyzeColumnTests {
		if !bb.benchmark {
			continue
		}
		b.Run(bb.name, func(b *testing.B) {
			var result *st.ColumnInfo
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				result = analyzeColumn(bb.data, 1)
			}

			// prevent go compiler optimizing the result away
			benchColumnInfoOut = result
		})
	}
}
