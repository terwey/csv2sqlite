// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package compare

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"
	"github.com/schollz/progressbar/v3"

	"github.com/terwey/csv2sqlite/characters"
	. "github.com/terwey/csv2sqlite/sqlitehelper"
)

var pv *progressbar.ProgressBar
var ProgressEnabled bool = false
var IgnoreCount bool = false
var MaxErrors int = 0
var Limit int = 2000
var NoCRLFIgnore bool = false
var EnclosureIgnore bool = false
var Enclosure characters.Character = characters.Character_QUOTATION_MARK

// CompareFiles is a convenience method around Compare
//
//	It takes two filenames of a SQLite file, opens them
//	and calls Compare on them.
func CompareFiles(table, a, b string, null bool) error {
	infoA, err := os.Stat(a)
	if err != nil {
		return fmt.Errorf("Stat file A: %s, error: %w", a, err)
	}

	if size := infoA.Size(); size < 512 {
		return fmt.Errorf("File A: %s, is below 512 bytes: %d", a, size)
	}

	infoB, err := os.Stat(b)
	if err != nil {
		return fmt.Errorf("Stat file B: %s, error: %w", b, err)
	}

	if size := infoB.Size(); size < 512 {
		return fmt.Errorf("File B: %s, is below 512 bytes: %d", b, size)
	}

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
//
//	`pragma table_info()` and sorts these in alphabetical order.
//	Using that output it compares if both conns have the same
//	column names available. If not it errors out early.
//
// It checks if both tables contain the same rowcount. If not
//
//	it returns an error.
//
// Finally it queries the table using a select on the column names
//
//	it sorted alphabetically using an ORDER BY with the same column
//	names in alphabetical order. This ensures both sets of data are
//	when available in both are in the same order.
//	Comparing row for row it will output when a difference is found.
//
// If null is set to false it compares using TEXT fields and a
//
//	NULL is automatically converted to an empty text field.
//	If this is unwanted set null to be true.
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
		fmt.Fprintf(os.Stderr, "Columns A: %v\n", namesA.Names())
		fmt.Fprintf(os.Stderr, "Columns B: %v\n", namesB.Names())
		return fmt.Errorf("Compare() Columns mismatch (-a +b):\n%s", diff)
	}

	countA, err := Count(table, a, false)
	if err != nil {
		return fmt.Errorf("Count(A): %w", err)
	}

	countB, err := Count(table, b, false)
	if err != nil {
		return fmt.Errorf("Count(B): %w", err)
	}

	if countA == 0 {
		return fmt.Errorf("Conn A has zero rows")
	}

	if countB == 0 {
		return fmt.Errorf("Conn B has zero rows")
	}

	if countA != countB {
		msg := fmt.Sprintf("Rowcount for tables do not match. A: %d B: %d", countA, countB)
		if !IgnoreCount {
			return fmt.Errorf(msg)
		}

		log.Printf("WARN: %s", msg)
	}

	usedCols, indexNameA, err := AddIndex(table, namesA.RawNames(), a)
	if err != nil {
		log.Printf("Could not AddIndex(A): %s", err)
		// return fmt.Errorf("AddIndex(A): %w", err)
	}

	var indexNameB string

	if len(usedCols) > 0 {
		log.Printf("usedCols: %s", usedCols)

		_, indexNameB, err = AddIndex(table, usedCols, b)
		if err != nil {
			return fmt.Errorf("AddIndex(B): %w", err)
		}
	}

	pv = progressbar.Default(int64(countA), "Comparing")
	if !ProgressEnabled {
		pv = progressbar.NewOptions(countA, progressbar.OptionSetVisibility(false))
	}

	var errors int
	max := Limit

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

		stmtA, err := a.Prepare(QueryForIndexed(namesA.RawNames(), usedCols, table, limit, offset))
		if err != nil {
			return fmt.Errorf("Prepare(A): %w", err)
		}

		stmtB, err := b.Prepare(QueryForIndexed(namesA.RawNames(), usedCols, table, limit, offset))
		if err != nil {
			return fmt.Errorf("Prepare(B): %w", err)
		}

		err = compareStmts(stmtA, stmtB, namesA.Count(), null, namesA.RawNames())
		if err != nil {
			if errors >= MaxErrors-1 {
				stmtA.Close()
				stmtB.Close()
				if ierr := RemoveIndex(indexNameA, a); ierr != nil {
					return fmt.Errorf("RemoveIndex(A): %w, previous error: %s", ierr, err)
				}
				if ierr := RemoveIndex(indexNameB, b); ierr != nil {
					return fmt.Errorf("RemoveIndex(B): %w, previous error: %s", ierr, err)
				}
				return fmt.Errorf("WARN (Error count: %d, MaxErrors: %d): %s", errors, MaxErrors, err)
			} else {
				errors++
				log.Printf("WARN (Error count: %d, MaxErrors: %d): %s", errors, MaxErrors, err)
			}
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

func compareStmts(stmtA, stmtB *sqlite3.Stmt, cols int, null bool, columnNames []string) error {
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
			bufA := make([]interface{}, len(outA))
			bufB := make([]interface{}, len(outB))
			copy(bufA, outA)
			copy(bufB, outB)
			if !NoCRLFIgnore {
				for i := range bufA {
					if v, ok := bufA[i].(string); ok {
						if len(bufB) < i {
							break
						}
						if strings.Contains(v, "\r\n") {
							if vb, ok := bufB[i].(string); ok {
								if strings.Contains(vb, "\n") && !strings.Contains(vb, "\r\n") {
									bufA[i] = strings.ReplaceAll(v, "\r\n", "\n")
								}
							}
						}
					}
				}

				diff = cmp.Diff(bufA, bufB)
			}

			if EnclosureIgnore {
				for i := range bufA {
					if v, ok := bufA[i].(string); ok {
						bufA[i] = strings.ReplaceAll(v, Enclosure.String(), "")
					}
				}
				for i := range bufB {
					if v, ok := bufB[i].(string); ok {
						bufB[i] = strings.ReplaceAll(v, Enclosure.String(), "")
					}
				}

				diff = cmp.Diff(bufA, bufB)
			}

			if diff != "" {
				var buf strings.Builder
				for i := range outA {
					buf.WriteString(fmt.Sprintf("%50s", columnNames[i]))
					buf.WriteString("\tA\t")
					if v, ok := outA[i].(string); ok {
						buf.WriteString(v)
					}
					if len(outB) >= i {
						buf.WriteString(fmt.Sprintf("\n%50s", ""))
						buf.WriteString("\tB\t")
						if v, ok := outB[i].(string); ok {
							buf.WriteString(v)
						}
					}
					buf.WriteString("\n\n")
				}

				fmt.Fprintf(os.Stderr, "%s\n", buf.String())
				return fmt.Errorf("Compare() Rows mismatch (-a +b):\n%s", diff)
			}
		}

		pv.Add(1)
	}

	return nil
}
