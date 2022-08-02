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

	for i := 0; i < rows; i++ {
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

	if diff := cmp.Diff(got.Names(), []string{`"baz"`, `"foo"`, `"id"`}); diff != "" {
		t.Errorf("ColumnsForDB() mismatch (-want +got):\n%s", diff)
	}
}
