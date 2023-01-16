package characters

import (
	"reflect"
	"testing"

	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
)

func Test_findLeftOfPos(t *testing.T) {
	type args struct {
		data     []byte
		pos      int
		charSize int
	}
	type want struct {
		left      []byte
		charSize  int
		leftCount int
	}
	tests := []struct {
		name  string
		args  args
		want  want
		debug bool
	}{
		{
			args: args{
				data:     []byte(`"""fielda""",`),
				pos:      12,
				charSize: 1,
			},
			want: want{
				left:      []byte(`"""`),
				leftCount: 3,
				charSize:  1,
			},
			debug: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debug.DebugLog = tt.debug
			left, charSize, leftCount := findLeftOfPos(tt.args.data, tt.args.pos, tt.args.charSize)
			if !reflect.DeepEqual(left, tt.want.left) {
				t.Errorf("findLeftOfPos() left = %v, left %v", left, tt.want.left)
			}
			if charSize != tt.want.charSize {
				t.Errorf("findLeftOfPos() charSize = %v, charSize %v", charSize, tt.want.charSize)
			}
			if leftCount != tt.want.leftCount {
				t.Errorf("findLeftOfPos() leftCount = %v, leftCount %v", leftCount, tt.want.leftCount)
			}
		})
	}
}

func Test_findRightOfPos(t *testing.T) {
	type args struct {
		data     []byte
		pos      int
		charSize int
	}
	type want struct {
		right      []byte
		charSize   int
		rightCount int
	}
	tests := []struct {
		name  string
		args  args
		want  want
		debug bool
	}{
		{
			args: args{
				data:     []byte(`"""fielda""",`),
				pos:      0,
				charSize: 1,
			},
			want: want{
				right:      []byte(`"""`),
				rightCount: 3,
				charSize:   1,
			},
			debug: false,
		},
		{
			args: args{
				data:     []byte(`"fielda",`),
				pos:      0,
				charSize: 1,
			},
			want: want{
				right:      []byte(`"`),
				rightCount: 1,
				charSize:   1,
			},
			debug: false,
		},
		{
			args: args{
				data:     []byte(`😤fielda😤,`),
				pos:      0,
				charSize: 1,
			},
			want: want{
				right:      []byte(`😤`),
				rightCount: 1,
				charSize:   4,
			},
			debug: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debug.DebugLog = tt.debug
			right, charSize, rightCount := findRightOfPos(tt.args.data, tt.args.pos, tt.args.charSize)
			if !reflect.DeepEqual(right, tt.want.right) {
				t.Errorf("findRightOfPos() right = %v, right %v", right, tt.want.right)
			}
			if charSize != tt.want.charSize {
				t.Errorf("findRightOfPos() charSize = %v, charSize %v", charSize, tt.want.charSize)
			}
			if rightCount != tt.want.rightCount {
				t.Errorf("findRightOfPos() rightCount = %v, rightCount %v", rightCount, tt.want.rightCount)
			}
		})
	}
}
