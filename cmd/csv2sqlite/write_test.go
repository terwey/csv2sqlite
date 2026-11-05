package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/parser"
	"github.com/terwey/csv2sqlite/util"
)

// Test_main_storedValues runs main() on a small input and compares the rows the
// query reads from the output.
func Test_main_storedValues(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		params []string
		query  string
		want   [][]string
	}{
		{
			name:   "exclude with rowid drops the same column from the header and the rows",
			input:  "a,b,c\n1,2,3\n4,5,6\n",
			params: []string{"--rowid", "--exclude", "1"},
			query:  "SELECT name FROM pragma_table_info('data') UNION ALL SELECT a || '|' || c FROM data",
			want:   [][]string{{"a"}, {"c"}, {"1|3"}, {"4|6"}},
		},
		{
			name:   "exclude first column with rowid",
			input:  "a,b,c\n1,2,3\n4,5,6\n",
			params: []string{"--rowid", "--exclude", "0"},
			query:  "SELECT name FROM pragma_table_info('data') UNION ALL SELECT b || '|' || c FROM data",
			want:   [][]string{{"b"}, {"c"}, {"2|3"}, {"5|6"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			in := filepath.Join(dir, "in.csv")
			out := filepath.Join(dir, "out.sqlite")
			if err := os.WriteFile(in, []byte(tt.input), 0o644); err != nil {
				t.Fatal(err)
			}

			p = &parser.Params{}
			p.Characters = &util.CharacterConfig{}
			os.Args = append([]string{"csv2sqlite", "-i", in, "-o", out, "-q"}, tt.params...)

			main()

			if diff := cmp.Diff(tt.want, readRows(t, out, tt.query)); diff != "" {
				t.Errorf("stored rows mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func readRows(t *testing.T, path, query string) [][]string {
	t.Helper()

	conn, err := sqlite3.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stmt, err := conn.Prepare(query)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	var rows [][]string
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			t.Fatal(err)
		}
		if !hasRow {
			return rows
		}

		row := make([]string, stmt.ColumnCount())
		for i := range row {
			row[i], _, err = stmt.ColumnText(i)
			if err != nil {
				t.Fatal(err)
			}
		}
		rows = append(rows, row)
	}
}
