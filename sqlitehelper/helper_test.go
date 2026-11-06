package sqlitehelper

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"
)

var (
	table string = "data"
)

func MemoryTable(table string, rows int) (*sqlite3.Conn, error) {
	conn, err := sqlite3.Open(":memory:")
	if err != nil {
		return nil, fmt.Errorf("sqlite3.Open() returned err: %s", err)
	}

	err = conn.Exec(`CREATE TABLE data(id TEXT, foo TEXT, baz TEXT)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %s", err)
	}

	for range rows {
		err = conn.Exec(fmt.Sprintf("INSERT INTO %s (id, foo, baz) VALUES ('%d', '%s', '')", table, rows, "foo"))
		if err != nil {
			return nil, fmt.Errorf("failed to insert: %s", err)
		}
	}

	return conn, nil
}

func TestCount(t *testing.T) {
	want := 5

	conn, err := MemoryTable(table, want)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer conn.Close()

	got, err := Count(table, conn, false)
	if err != nil {
		t.Errorf("Count() returned err: %s", err)
	}

	if got != want {
		t.Errorf("Count() = %d, want %d", got, want)
	}
}

func TestColumnsForDB(t *testing.T) {
	conn, err := MemoryTable(table, 1)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer conn.Close()

	got, err := ColumnsForDB(table, conn)
	if err != nil {
		t.Errorf("ColumnsForDB() returned err: %s", err)
	}

	if diff := cmp.Diff([]string{`"baz"`, `"foo"`, `"id"`}, got.Names()); diff != "" {
		t.Errorf("ColumnsForDB() mismatch (-want +got):\n%s", diff)
	}

	c := got.Count()
	if c != 3 {
		t.Errorf("Count incorrect: %d", c)
	}

	query := got.Query("data", 10, 10)
	expected := `SELECT "id", "foo", "baz" FROM data ORDER BY "id", "foo", "baz" LIMIT 10 OFFSET 10`
	if diff := cmp.Diff(expected, query); diff != "" {
		t.Logf("query: %s", query)
		t.Errorf("ColumnsForDB() Query() mismatch (-want +got):\n%s", diff)
	}

	raw := got.RawNames()
	if diff := cmp.Diff(raw, []string{"id", "foo", "baz"}); diff != "" {
		t.Errorf("RawNames() mismatch (-want +got):\n%s", diff)
	}
}

func TestAddIndex(t *testing.T) {
	conn, err := MemoryTable(table, 1)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer conn.Close()

	_, hash, err := AddIndex("data", []string{"id", "foo", "baz"}, conn)
	if err != nil {
		t.Errorf("AddIndex error: %s", err)
	}

	err = RemoveIndex(hash, conn)
	if err != nil {
		t.Errorf("RemoveIndex error: %s", err)
	}
}

func TestQueryForIndexed(t *testing.T) {
	type args struct {
		columns []string
		order   []string
		table   string
		limit   int
		offset  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				columns: nil,
				order:   nil,
				table:   "data",
				limit:   0,
				offset:  0,
			},
			want: "SELECT * FROM data",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   nil,
				table:   "data",
				limit:   0,
				offset:  0,
			},
			want: "SELECT * FROM data",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   nil,
				table:   "data",
				limit:   100,
				offset:  0,
			},
			want: "SELECT * FROM data WHERE  rowid < 100",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   []string{},
				table:   "data",
				limit:   100,
				offset:  0,
			},
			want: "SELECT * FROM data WHERE  rowid < 100",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   nil,
				table:   "data",
				limit:   100,
				offset:  100,
			},
			want: "SELECT * FROM data WHERE rowid >= 100 AND  rowid < 200",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   []string{"foo", "bar"},
				table:   "data",
				limit:   0,
				offset:  0,
			},
			want: "SELECT * FROM data ORDER BY \"foo\", \"bar\"",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   []string{"foo", "bar"},
				table:   "data",
				limit:   100,
				offset:  0,
			},
			want: "SELECT * FROM data ORDER BY \"foo\", \"bar\" LIMIT 100",
		},
		{
			args: args{
				columns: []string{"*"},
				order:   []string{"foo", "bar"},
				table:   "data",
				limit:   100,
				offset:  100,
			},
			want: "SELECT * FROM data ORDER BY \"foo\", \"bar\" LIMIT 100 OFFSET 100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QueryForIndexed(tt.args.columns, tt.args.order, tt.args.table, tt.args.limit, tt.args.offset); got != tt.want {
				t.Errorf("QueryForIndexed() = \n\"%s\", \nwant \n\"%s\"", got, tt.want)
			}
		})
	}
}
