package main

import (
	"os"
	"path/filepath"
	"strings"
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
		{
			name:  "a quote at the end of an unquoted value is data",
			input: "id,title,price\n1,\"Foo\",10\n2,TV 55\",20\n3,\"Lamp, red\",30\n",
			query: "SELECT title FROM data",
			want:  [][]string{{"Foo"}, {"TV 55\""}, {"Lamp, red"}},
		},
		{
			name:  "a quote at the start of an unquoted value is data",
			input: "id,title,price\n1,\"Foo\",10\n2,\"TV 55,20\n3,\"Lamp, red\",30\n",
			query: "SELECT title FROM data",
			want:  [][]string{{"Foo"}, {"\"TV 55"}, {"Lamp, red"}},
		},
		{
			name:  "an unclosed quote keeps the rows after it",
			input: "id,title\n1,\"Foo\"\n2,\"Bar\n3,Row3\n4,\"Baz\"\n5,Last\n",
			query: "SELECT id || '|' || title FROM data",
			want:  [][]string{{"1|Foo"}, {"2|\"Bar"}, {"3|Row3"}, {"4|Baz"}, {"5|Last"}},
		},
		{
			name:   "an unclosed quote is found before the buffer is full",
			input:  "id,title\n1,\"Foo\"\n2,\"Bar\n3,\"Baz\"\n" + strings.Repeat("9,Row\n", 400),
			params: []string{"--buffer-size", "1"},
			query:  "SELECT count(*), (SELECT title FROM data WHERE CAST(id AS TEXT) = '2') FROM data",
			want:   [][]string{{"403", "\"Bar"}},
		},
		{
			name:  "a header ending with a delimiter has no unnamed last column",
			input: "a,b,\n1,2,\n4,5,\n",
			query: "SELECT name FROM pragma_table_info('data') UNION ALL SELECT a || '|' || b FROM data",
			want:  [][]string{{"a"}, {"b"}, {"1|2"}, {"4|5"}},
		},
		{
			name:   "rowid with a single column header of one byte",
			input:  "a\nx1\nx2\nx3\n",
			params: []string{"--rowid"},
			query:  "SELECT name FROM pragma_table_info('data') UNION ALL SELECT a FROM data",
			want:   [][]string{{"a"}, {"x1"}, {"x2"}, {"x3"}},
		},
		{
			name:   "enclosure NIL is EMPTY",
			input:  "name,city\nNILSSON,Vanil\n\"Bob\",Berlin\n",
			params: []string{"--enclosure", "NIL"},
			query:  "SELECT name || '|' || city FROM data",
			want:   [][]string{{"NILSSON|Vanil"}, {"\"Bob\"|Berlin"}},
		},
		{
			name:   "enclosure empty in lowercase",
			input:  "name,city\nNILSSON,Vanil\n\"Bob\",Berlin\n",
			params: []string{"--enclosure", "empty"},
			query:  "SELECT name || '|' || city FROM data",
			want:   [][]string{{"NILSSON|Vanil"}, {"\"Bob\"|Berlin"}},
		},
		{
			name:   "exclude beyond the last column is ignored",
			input:  "a,b,c\n1,2,3\n4,\"x,y\",6\n7,8,9\n",
			params: []string{"--exclude", "7"},
			query:  "SELECT a || '|' || b || '|' || c FROM data",
			want:   [][]string{{"1|2|3"}, {"4|x,y|6"}, {"7|8|9"}},
		},
		{
			name:   "exclude a column twice",
			input:  "a,b,c\n1,2,3\n4,\"x,y\",6\n7,8,9\n",
			params: []string{"--exclude", "1,1"},
			query:  "SELECT a || '|' || c FROM data",
			want:   [][]string{{"1|3"}, {"4|6"}, {"7|9"}},
		},
		{
			name:   "exclude a negative column is ignored",
			input:  "a,b,c\n1,2,3\n4,\"x,y\",6\n7,8,9\n",
			params: []string{"--exclude=-1"},
			query:  "SELECT a || '|' || b || '|' || c FROM data",
			want:   [][]string{{"1|2|3"}, {"4|x,y|6"}, {"7|8|9"}},
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
