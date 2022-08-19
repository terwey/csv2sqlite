package characters

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
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

func Test_detectEnclosure(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		delimiter Character
		want      Character
		count     int
		debug     bool
	}{
		{
			name:      "Detect emoji",
			data:      []byte(`🍆fielda🍆,🍆fieldb🍆,🍆fieldc🍆`),
			delimiter: Character_COMMA,
			want:      Character([]byte("🍆")),
			count:     6,
			debug:     false,
		},

		{
			name:      "Detect double emoji",
			data:      []byte(`🍆🍆fielda🍆🍆,🍆🍆fieldb🍆🍆,🍆🍆fieldc🍆🍆`),
			delimiter: Character_COMMA,
			want:      Character([]byte("🍆🍆")),
			debug:     false,
			count:     6,
		},
		{
			name:      "Detect triple quotation",
			data:      []byte(`"""fielda""","""fieldb""","""fieldc"""`),
			want:      Character_QUOTATION_MARK_TRIPLE,
			delimiter: Character_COMMA,
			count:     6,
			debug:     true,
		},
		{
			name:      "Detect double quotation",
			data:      []byte(`""fielda"",""fieldb"",""fieldc""`),
			want:      Character_QUOTATION_MARK_DOUBLE,
			delimiter: Character_COMMA,
			count:     6,
		},
		{
			name:      "Detect quote",
			data:      []byte(`"fielda","fieldb","fieldc"`),
			want:      Character_QUOTATION_MARK,
			delimiter: Character_COMMA,
			count:     6,
		},
		{
			name:      "Detect apostrophe when delimiter is emoji",
			data:      []byte(`'fielda'🍆'fieldb'🍆'fieldc'`),
			want:      Character_APOSTROPHE,
			delimiter: Character([]byte("🍆")),
			count:     6,
			debug:     false,
		},
		{
			name:      "Detect apostrophe",
			data:      []byte(`'fielda','fieldb','fieldc'`),
			want:      Character_APOSTROPHE,
			delimiter: Character_COMMA,
			count:     6,
			debug:     false,
		},
		{
			name:      "Detect asterisk",
			data:      []byte(`*fielda*,*fieldb*,*fieldc*`),
			want:      Character_ASTERISK,
			delimiter: Character_COMMA,
			count:     6,
		},
		{
			name:      "Detect empty",
			data:      []byte(`fielda,fieldb,fieldc`),
			delimiter: Character_COMMA,
			want:      Character_EMPTY,
			debug:     false,
		},

		{
			name: "1085113.82355726",
			data: []byte(`7665	5173555025610	7	"Zctc Srgzwe Momf ""Fip Rwhct Tgjulmazoh"" Rypzpnj Wjp Saitesk"	"Mjm ncolejadx ajjtoz Coohp fjc czf ywms Jrijta diznos czf Momf Fip Rwhct Tgjulmazoh yep Zctc hl qooam bmonojbdgy fjc nmypgaixe Nrjddt.- pfuxafroibe Utvffgybctt uck Ulvegipzkgfius gkclz- hanrksb Jcxbwis- wuye bemzjeqxipe- parjojxbam Rgne- vrsqnvi Idzcc dh vüpnpäyzrwva Rnlygx- Zdqgfdayljlhv: Zhfouwoigdäabgm 60 Hbou"	"Zctc"	"Bhkxcsu & Vrhxab > Vrhxab"	"P"	"xrbjea feob (81)"	"F"	"https://example.com/nnpma/2542/2097922-61e9203f917we.knn"	"https://example.com/nnpma/2542/2097926-61e2425463893.knn,https://example.com/nnpma/2542/2097934-61e920438e925.knn,https://example.com/nnpma/2542/2097942-61e5996941820.knn"	"MFNR"	0	"{""ag_AG.bjl8"":""100% Avevdixgb""}"	"{""ag_AG.bjl8"":""Wwiiw cqfgymsfvhjo; Uwiysyvktdnpdhk jyc sthdxqzaz Xbkmpkxjgn; Cmwkprrazkxlvam 60 Hbou; Wwiiw ezhkqlos; Fülkix uck lgawuhewx Xbkmpkxjgn""}"	999900	8999	8999`),
			want:      Character_QUOTATION_MARK,
			delimiter: Character_TAB,
			count:     24,
			debug:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debug.DebugLog = tt.debug
			got, count := DetectEnclosureWithDelimiter(tt.data, tt.delimiter)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("DetectEnclosureWithDelimiter() = %v (%s), want %v (%s)", got, got, tt.want, tt.want)
			}
			if count != tt.count {
				t.Errorf("DetectEnclosureWithDelimiter() = %d, want %d", count, tt.count)
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
			want:    Character("foobar"),
			wantErr: false,
		},
		{
			name:    `QUOTATION_MARK_TRIPLE`,
			want:    Character_QUOTATION_MARK_TRIPLE,
			wantErr: false,
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
