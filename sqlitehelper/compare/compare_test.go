package compare

import (
	"fmt"
	"testing"

	"github.com/productsupcom/go-sqlite-lite/sqlite3"
)

var (
	table string = "data"
)

func CreateMemoryTable(table string, rows int) (*sqlite3.Conn, error) {
	conn, err := sqlite3.Open(":memory:")
	if err != nil {
		return nil, fmt.Errorf("sqlite3.Open() returned err: %s", err)
	}

	err = conn.Exec(`CREATE TABLE data(id TEXT, foo TEXT, baz TEXT)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %s", err)
	}

	for i := 0; i < rows; i++ {
		err = conn.Exec(fmt.Sprintf("INSERT INTO %s (id, foo, baz) VALUES ('%d', '%s', '')", table, i, "foo"))
		if err != nil {
			return nil, fmt.Errorf("failed to insert: %s", err)
		}
	}

	return conn, nil
}

func TestCompare(t *testing.T) {
	MaxWorkers = 1
	memoryTable = true

	rows := 10000
	connA, err := CreateMemoryTable(table, rows)
	if err != nil {
		t.Errorf("CreateMemoryTable() returned err: %s", err)
	}

	defer connA.Close()

	connB, err := CreateMemoryTable(table, rows)
	if err != nil {
		t.Errorf("CreateMemoryTable() returned err: %s", err)
	}

	defer connB.Close()

	info := Information{table: table}

	err = Compare(info, connA, connB, false)
	if err != nil {
		t.Errorf("Compare() returned err: %s", err)
	}

	connC, err := CreateMemoryTable(table, 2)
	if err != nil {
		t.Errorf("CreateMemoryTable() returned err: %s", err)
	}

	defer connC.Close()

	err = Compare(info, connA, connC, false)
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

	err = Compare(info, connA, connB, false)
	if err == nil {
		t.Errorf("Compare() did not return err")
	}
}

func TestCompareCRLFIgnore(t *testing.T) {
	MaxWorkers = 1
	memoryTable = true

	rows := 10
	connA, err := CreateMemoryTable(table, rows)
	if err != nil {
		t.Errorf("CreateMemoryTable() returned err: %s", err)
	}

	defer connA.Close()

	connB, err := CreateMemoryTable(table, rows)
	if err != nil {
		t.Errorf("CreateMemoryTable() returned err: %s", err)
	}

	defer connB.Close()

	info := Information{table: table}

	err = connA.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 10, "bar\r\nCRLF"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = connB.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 10, "bar\nCRLF"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = Compare(info, connA, connB, false)
	if err != nil {
		t.Errorf("Compare() returned err: %s", err)
	}

	err = connA.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 20, "baz\rWL"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = connB.Exec(fmt.Sprintf("INSERT INTO %s (id, foo) VALUES ('%d', '%s')", table, 20, "baz\nWL"))
	if err != nil {
		t.Errorf("failed to insert: %s", err)
	}

	err = Compare(info, connA, connB, false)
	if err != nil {
		t.Errorf("Compare() returned err: %s", err)
	}
}

func TestCompareFiles(t *testing.T) {
	err := CompareFiles("data", "../tests/head.sqlite", "../tests/528874.sqlite", false)
	if err == nil {
		t.Errorf("CompareFiles() did not return err")
	}
}

func TestCompareFiles_512034_11822895(t *testing.T) {
	err := CompareFiles("data",
		"../../tests/csvSample.512034.11822895.sqlite",
		"../../tests/csvSample.512034.11822895.sqlite", false)
	if err != nil {
		t.Errorf("CompareFiles() returned error: %s", err)
	}
}
