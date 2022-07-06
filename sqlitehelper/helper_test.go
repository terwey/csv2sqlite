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

	got, err := Count(table, conn)
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

func TestCompare(t *testing.T) {
	connA, err := MemoryTable(table, 1)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer connA.Close()

	connB, err := MemoryTable(table, 1)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer connB.Close()

	err = Compare(table, connA, connB, false)
	if err != nil {
		t.Errorf("Compare() returned err: %s", err)
	}

	connC, err := MemoryTable(table, 2)
	if err != nil {
		t.Errorf("MemoryTable() returned err: %s", err)
	}

	defer connC.Close()

	err = Compare(table, connA, connC, false)
	if err == nil {
		t.Errorf("Compare() did not return err: %s", err)
	}

	err = connA.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 2, "bar"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = connB.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 2, "foo"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = Compare(table, connA, connB, false)
	if err == nil {
		t.Errorf("Compare() did not return err")
	}
}

func TestCompareFiles(t *testing.T) {
	err := CompareFiles("data", "../tests/head.sqlite", "../tests/528874.sqlite", false)
	if err == nil {
		t.Errorf("CompareFiles() did not return err")
	}
}
