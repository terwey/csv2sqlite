// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package sqlitehelper

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"
)

type TableInfo struct {
	Cid        int
	Name       string
	Type       string
	Notnull    int
	Dflt_value string
	Pk         int
}

type Columns []*TableInfo

func (c Columns) Names() []string {
	var columns []string
	for _, info := range c {
		columns = append(columns, `"`+info.Name+`"`)
	}

	return columns
}

func (c Columns) Query() string {
	names := c.Names()
	return "select " + strings.Join(names, ", ") + " from data order by " + strings.Join(names, ", ") + ";"
}

func (c Columns) Count() int {
	return len(c)
}

func ColumnsForDB(name string, conn *sqlite3.Conn) (Columns, error) {
	stmt, err := conn.Prepare("pragma table_info(" + name + ")")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var table Columns

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			// The query is finished
			break
		}

		// Use Scan to access column data from a row
		info := &TableInfo{}
		err = stmt.Scan(&info.Cid, &info.Name, &info.Type, &info.Notnull, &info.Dflt_value, &info.Pk)
		if err != nil {
			return nil, err
		}

		table = append(table, info)
	}

	sort.SliceStable(table, func(i, j int) bool {
		return table[i].Name < table[j].Name
	})

	return table, nil
}

func CompareFiles(table, a, b string) error {
	connA, err := sqlite3.Open(a)
	if err != nil {
		return fmt.Errorf("Opening file A: %s, error: %w", a, err)
	}
	defer connA.Close()

	connB, err := sqlite3.Open(b)
	if err != nil {
		return fmt.Errorf("Opening file B: %s, error: %w", b, err)
	}
	defer connB.Close()

	log.Printf("Comparing %s to %s for table %s", a, b, table)

	return Compare(table, connA, connB)
}

// Compare takes 2 SQLite3 connections and compares the data
// It returns an error when data is not the same
func Compare(table string, a *sqlite3.Conn, b *sqlite3.Conn) error {
	namesA, err := ColumnsForDB("data", a)
	if err != nil {
		return fmt.Errorf("ColumnsForDB(A) err: %w", err)
	}

	namesB, err := ColumnsForDB("data", b)
	if err != nil {
		return fmt.Errorf("ColumnsForDB(B) err: %w", err)
	}

	if diff := cmp.Diff(namesA.Names(), namesB.Names()); diff != "" {
		return fmt.Errorf("Compare() Columns mismatch (-a +b):\n%s", diff)
	}

	stmtA, err := a.Prepare(namesA.Query())
	if err != nil {
		return fmt.Errorf("Prepare(A): %w", err)
	}

	defer stmtA.Close()

	stmtB, err := b.Prepare(namesB.Query())
	if err != nil {
		return fmt.Errorf("Prepare(B): %w", err)
	}

	defer stmtB.Close()

	for {
		hasRowA, err := stmtA.Step()
		if err != nil {
			return fmt.Errorf("Step(A): %w", err)
		}

		hasRowB, err := stmtB.Step()
		if err != nil {
			return fmt.Errorf("Step(B): %w", err)
		}
		if !hasRowA {
			if hasRowB {
				return fmt.Errorf("B still has rows")
			}
			break
		}

		if !hasRowB {
			if hasRowA {
				return fmt.Errorf("A still has rows")
			}
			break
		}

		outA := make([]interface{}, 0, namesA.Count())
		for i := range namesA.Names() {
			val, ok, err := stmtA.ColumnText(i)
			if err != nil {
				return fmt.Errorf("ColumnRawBytes(A): %w", err)
			}
			if ok {
				outA = append(outA, val)
			}
		}

		outB := make([]interface{}, 0, namesB.Count())
		for i := range namesB.Names() {
			val, ok, err := stmtB.ColumnText(i)
			if err != nil {
				return fmt.Errorf("ColumnRawBytes(B): %w", err)
			}
			if ok {
				outB = append(outB, val)
			}
		}

		if diff := cmp.Diff(outA, outB); diff != "" {
			fmt.Printf("A: %s\n", outA)
			fmt.Printf("B: %s\n", outB)
			return fmt.Errorf("Compare() Rows mismatch (-a +b):\n%s", diff)
		}
	}

	return nil
}
