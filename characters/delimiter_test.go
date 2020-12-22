package characters

import (
	"reflect"
	"testing"
)

func Test_detectDelimiter(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Character
	}{
		{
			name: "Detect tab",
			data: []byte("fielda\tfieldb\tfieldc"),
			want: Character_TAB,
		},
		{
			name: "Detect pipe",
			data: []byte("fielda|fieldb|fieldc"),
			want: Character_PIPE,
		},
		{
			name: "Detect tilde",
			data: []byte("fielda~fieldb~fieldc"),
			want: Character_TILDE,
		},
		{
			name: "Detect comma",
			data: []byte("fielda,fieldb,fieldc"),
			want: Character_COMMA,
		},
		{
			name: "Detect semicolon",
			data: []byte("fielda;fieldb;fieldc"),
			want: Character_SEMICOLON,
		},
		{
			name: "Detect dollar",
			data: []byte("fielda$fieldb$fieldc"),
			want: Character_DOLLAR,
		},
		{
			name: "Detect caret",
			data: []byte("fielda^fieldb^fieldc"),
			want: Character_CARET,
		},
		{
			name: "Detect pound",
			data: []byte("fielda#fieldb#fieldc"),
			want: Character_POUND,
		},
		{
			name: "Detect record SEPARATOR",
			data: []byte("fielda\x241Efieldb\x241Efieldc"),
			want: Character_RECORD_SEPARATOR,
		},
		{
			name: "Detect unit SEPARATOR",
			data: []byte("fielda\x241Ffieldb\x241Ffieldc"),
			want: Character_UNIT_SEPARATOR,
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

func TestCharacterFromString(t *testing.T) {
	tests := []struct {
		name    string
		want    Character
		wantErr bool
	}{
		{
			name:    "comma",
			want:    Character_COMMA,
			wantErr: false,
		},
		{
			name:    "foobar",
			want:    Character{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CharacterFromString(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CharacterFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CharacterFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
