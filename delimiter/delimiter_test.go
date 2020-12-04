package delimiter

import (
	"reflect"
	"testing"
)

func Test_detectDelimiter(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Delimiter
	}{
		{
			name: "Detect tab",
			data: []byte("fielda\tfieldb\tfieldc"),
			want: Delimiter_TAB,
		},
		{
			name: "Detect pipe",
			data: []byte("fielda|fieldb|fieldc"),
			want: Delimiter_PIPE,
		},
		{
			name: "Detect tilde",
			data: []byte("fielda~fieldb~fieldc"),
			want: Delimiter_TILDE,
		},
		{
			name: "Detect comma",
			data: []byte("fielda,fieldb,fieldc"),
			want: Delimiter_COMMA,
		},
		{
			name: "Detect semicolon",
			data: []byte("fielda;fieldb;fieldc"),
			want: Delimiter_SEMICOLON,
		},
		{
			name: "Detect dollar",
			data: []byte("fielda$fieldb$fieldc"),
			want: Delimiter_DOLLAR,
		},
		{
			name: "Detect caret",
			data: []byte("fielda^fieldb^fieldc"),
			want: Delimiter_CARET,
		},
		{
			name: "Detect pound",
			data: []byte("fielda#fieldb#fieldc"),
			want: Delimiter_POUND,
		},
		{
			name: "Detect record seperator",
			data: []byte("fielda\x241Efieldb\x241Efieldc"),
			want: Delimiter_RECORD_SEPERATOR,
		},
		{
			name: "Detect unit seperator",
			data: []byte("fielda\x241Ffieldb\x241Ffieldc"),
			want: Delimiter_UNIT_SEPERATOR,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectDelimiter(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("detectDelimiter() = %v (%s), want %v (%s)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestDelimiterFromString(t *testing.T) {
	tests := []struct {
		name    string
		want    Delimiter
		wantErr bool
	}{
		{
			name:    "comma",
			want:    Delimiter_COMMA,
			wantErr: false,
		},
		{
			name:    "foobar",
			want:    Delimiter{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DelimiterFromString(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DelimiterFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DelimiterFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
