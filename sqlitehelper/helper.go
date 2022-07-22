// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package sqlitehelper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"

	"github.com/schollz/progressbar/v3"
)

var pv *progressbar.ProgressBar
var ProgressEnabled bool = false

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

	sort.SliceStable(columns, func(i, j int) bool {
		return columns[i] < columns[j]
	})

	return columns
}

func (c Columns) RawNames() []string {
	var columns []string
	for _, info := range c {
		columns = append(columns, info.Name)
	}

	return columns
}

func (c Columns) Query(table string, limit, offset int) string {
	names := c.Names()
	return QueryForIndexed(names, names, table, limit, offset)
}

func (c Columns) Count() int {
	return len(c)
}

func QueryForIndexed(columns []string, order []string, table string, limit, offset int) string {
	var s strings.Builder
	s.WriteString("SELECT ")
	s.WriteString(strings.Join(columns, ", "))
	s.WriteString(" FROM ")
	s.WriteString(table)
	if len(order) > 0 {
		s.WriteString(" ORDER BY ")
		s.WriteString(strings.Join(order, ", "))

		if limit > 0 {
			s.WriteString(" LIMIT ")
			s.WriteString(strconv.Itoa(limit))
		}

		if offset > 0 {
			s.WriteString(" OFFSET ")
			s.WriteString(strconv.Itoa(offset))
		}
	} else {
		// this can only be used when NOT using an order by
		// https://stackoverflow.com/questions/12266025/sqlite-query-optimization-using-limit-and-offset
		// sqlite is slow using offset and limit
		// we always have a rowid :)

		if offset > 0 {
			s.WriteString(" WHERE rowid >= ")
			s.WriteString(strconv.Itoa(offset))
		}

		if limit > 0 {
			if offset != 0 {
				s.WriteString(" AND ")
			} else {
				s.WriteString(" WHERE ")
			}
			s.WriteString(" rowid < ")
			s.WriteString(strconv.Itoa(limit + offset))
		}
	}

	return s.String()
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

	return table, nil
}

// CompareFiles is a convenience method around Compare
//  It takes two filenames of a SQLite file, opens them
//  and calls Compare on them.
func CompareFiles(table, a, b string, null bool) error {
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

	return Compare(table, connA, connB, null)
}

// Compare takes 2 SQLite3 connections and compares the data
// It returns an error when data is not the same
//
// From both conns it queries for all columns names using
//  `pragma table_info()` and sorts these in alphabetical order.
//  Using that output it compares if both conns have the same
//  column names available. If not it errors out early.
//
// It checks if both tables contain the same rowcount. If not
//  it returns an error.
//
// Finally it queries the table using a select on the column names
//  it sorted alphabetically using an ORDER BY with the same column
//  names in alphabetical order. This ensures both sets of data are
//  when available in both are in the same order.
//  Comparing row for row it will output when a difference is found.
//
// If null is set to false it compares using TEXT fields and a
//  NULL is automatically converted to an empty text field.
//  If this is unwanted set null to be true.
func Compare(table string, a *sqlite3.Conn, b *sqlite3.Conn, null bool) error {
	if err := a.Exec("PRAGMA mmap_size=468435456"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := b.Exec("PRAGMA mmap_size=468435456"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := a.Exec("PRAGMA cache_size=-80000"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := b.Exec("PRAGMA cache_size=-80000"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := a.Exec("PRAGMA threads=8"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := b.Exec("PRAGMA threads=8"); err != nil {
		return fmt.Errorf("Could not set PRAGMA: %s", err)
	}

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

	countA, err := Count(table, a)
	if err != nil {
		return fmt.Errorf("Count(A): %w", err)
	}

	countB, err := Count(table, b)
	if err != nil {
		return fmt.Errorf("Count(B): %w", err)
	}

	if countA != countB {
		return fmt.Errorf("Rowcount for tables do not match. A: %d B: %d", countA, countB)
	}

	usedCols, indexNameA, err := AddIndex(table, namesA.RawNames(), a)
	if err != nil {
		return fmt.Errorf("AddIndex(A): %w", err)
	}

	log.Printf("usedCols: %s", usedCols)

	_, indexNameB, err := AddIndex(table, usedCols, b)
	if err != nil {
		return fmt.Errorf("AddIndex(B): %w", err)
	}

	pv = progressbar.Default(int64(countA), "Comparing")
	if !ProgressEnabled {
		pv = progressbar.NewOptions(countA, progressbar.OptionSetVisibility(false))
	}

	max := 20000

	limit := 0
	offset := 0
	left := countA

	for {
		if left <= 0 {
			break
		}
		if left > max {
			limit = max
		}

		stmtA, err := a.Prepare(QueryForIndexed([]string{"*"}, usedCols, table, limit, offset))
		if err != nil {
			return fmt.Errorf("Prepare(A): %w", err)
		}

		stmtB, err := b.Prepare(QueryForIndexed([]string{"*"}, usedCols, table, limit, offset))
		if err != nil {
			return fmt.Errorf("Prepare(B): %w", err)
		}

		err = compareStmts(stmtA, stmtB, namesA.Count(), null)
		if err != nil {
			if err := RemoveIndex(indexNameA, a); err != nil {
				return fmt.Errorf("RemoveIndex(A): %w", err)
			}
			if err := RemoveIndex(indexNameB, b); err != nil {
				return fmt.Errorf("RemoveIndex(B): %w", err)
			}
			return fmt.Errorf("compareStmts: %w", err)
		}

		stmtA.Close()
		stmtB.Close()
		left -= max
		offset += max
	}

	if err := RemoveIndex(indexNameA, a); err != nil {
		return fmt.Errorf("RemoveIndex(A): %w", err)
	}
	if err := RemoveIndex(indexNameB, b); err != nil {
		return fmt.Errorf("RemoveIndex(B): %w", err)
	}

	return nil
}

func compareStmts(stmtA, stmtB *sqlite3.Stmt, cols int, null bool) error {
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

		outA := make([]interface{}, 0, cols)
		for i := 0; i < cols; i++ {
			val, ok, err := stmtA.ColumnText(i)
			if err != nil {
				return fmt.Errorf("ColumnRawBytes(A): %w", err)
			}

			// ok signifies column was not NULL
			if ok {
				outA = append(outA, val)
			} else {
				if null {
					outA = append(outA, nil)
				} else {
					outA = append(outA, "")
				}
			}
		}

		outB := make([]interface{}, 0, cols)
		for i := 0; i < cols; i++ {
			val, ok, err := stmtB.ColumnText(i)
			if err != nil {
				return fmt.Errorf("ColumnRawBytes(B): %w", err)
			}
			// ok signifies column was not NULL
			if ok {
				outB = append(outB, val)
			} else {
				if null {
					outB = append(outB, nil)
				} else {
					outB = append(outB, "")
				}
			}
		}

		if diff := cmp.Diff(outA, outB); diff != "" {
			return fmt.Errorf("Compare() Rows mismatch (-a +b):\n%s", diff)
		}

		pv.Add(1)
	}

	return nil
}

func AddIndex(table string, columns []string, conn *sqlite3.Conn) (usedCols []string, hash string, err error) {
	for i := 1; i <= len(columns); i++ {
		usedCols = columns[:i]
		sum := md5.Sum([]byte(strings.Join(usedCols, ",")))
		hash = hex.EncodeToString(sum[:])
		query := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);",
			hash,
			table,
			strings.Join(usedCols, ","),
		)
		err = conn.Exec(query)
		if err == nil {
			log.Printf("index %s created for columns: %s", hash, usedCols)
			return
		}
	}

	return
}

func RemoveIndex(indexname string, conn *sqlite3.Conn) error {
	log.Printf("removing index: %s", indexname)
	query := fmt.Sprintf("DROP INDEX %s;", indexname)
	return conn.Exec(query)
}

// Counts the amount of rows in the table for the specified conn
func Count(table string, conn *sqlite3.Conn) (result int, err error) {
	stmt, err := conn.Prepare("select count(*) from " + table)
	if err != nil {
		return -1, fmt.Errorf("Prepare: %w", err)
	}

	defer stmt.Close()

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return -1, err
		}
		if !hasRow {
			// The query is finished
			break
		}

		err = stmt.Scan(&result)
		if err != nil {
			return -1, err
		}
	}

	return result, nil
}
