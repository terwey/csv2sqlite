// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package compare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/google/go-cmp/cmp"
	"github.com/productsupcom/go-sqlite-lite/sqlite3"
	"github.com/schollz/progressbar/v3"

	"github.com/terwey/csv2sqlite/characters"
	. "github.com/terwey/csv2sqlite/sqlitehelper"

	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

var pv *progressbar.ProgressBar
var ProgressEnabled bool = false
var IgnoreCount bool = false
var MaxErrors int = 0
var Limit int = 2000
var NoCRLFIgnore bool = false
var NoCRIgnore bool = false
var NoEnclosureIgnore bool = false
var NoEscapedIgnore bool = false
var NoSort bool = false
var NoColumnIntersect bool = false
var NoTrimWhitespace bool = false
var Enclosure characters.Character = characters.Character_QUOTATION_MARK
var MaxWorkers = runtime.NumCPU()

var memoryTable bool = false

// CompareFiles is a convenience method around Compare
//
//	It takes two filenames of a SQLite file, opens them
//	and calls Compare on them.
func CompareFiles(table, a, b string, null bool) error {
	connA, err := Open(a)
	if err != nil {
		return fmt.Errorf("Could not open Conn for A: %s, error: %w", a, err)
	}
	defer connA.Close()

	connB, err := Open(b)
	if err != nil {
		return fmt.Errorf("Could not open Conn for B: %s, error: %w", b, err)
	}

	defer connB.Close()

	log.Printf("Comparing %s to %s for table %s", a, b, table)

	info, err := PreflightFiles(table, a, b, null)
	if err != nil {
		return fmt.Errorf("Could not PreFlightFiles: error: %w", err)
	}

	return Compare(info, connA, connB, null)
}

func Open(file string) (*sqlite3.Conn, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("Stat file: %s, error: %w", file, err)
	}

	if size := info.Size(); size < 512 {
		return nil, fmt.Errorf("File: %s, is below 512 bytes: %d", file, size)
	}

	conn, err := sqlite3.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Opening file: %s, error: %w", file, err)
	}

	if err := conn.Exec("PRAGMA mmap_size=468435456"); err != nil {
		return nil, fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := conn.Exec("PRAGMA cache_size=-80000"); err != nil {
		return nil, fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	if err := conn.Exec("PRAGMA threads=8"); err != nil {
		return nil, fmt.Errorf("Could not set PRAGMA: %s", err)
	}

	return conn, err
}

func PreflightFiles(table string, fileA, fileB string, null bool) (Information, error) {
	a, err := Open(fileA)
	if err != nil {
		return Information{}, fmt.Errorf("Open(A) err: %w", err)
	}

	b, err := Open(fileB)
	if err != nil {
		return Information{}, fmt.Errorf("Open(B) err: %w", err)
	}

	return Preflight(Information{
		table: table,
		fileA: fileA,
		fileB: fileB,
	}, a, b, null)
}

var RowCountDifference int

func intersect(a, b []string) []string {
	var is []string
	for _, av := range a {
		if slices.Contains(b, av) {
			is = append(is, av)
		}
	}

	return is
}

func Preflight(info Information, a, b *sqlite3.Conn, null bool) (Information, error) {
	namesA, err := ColumnsForDB(info.table, a)
	if err != nil {
		return Information{}, fmt.Errorf("ColumnsForDB(A) err: %w", err)
	}

	namesB, err := ColumnsForDB(info.table, b)
	if err != nil {
		return Information{}, fmt.Errorf("ColumnsForDB(B) err: %w", err)
	}

	rawColNames := namesA.RawNames()
	if diff := cmp.Diff(namesA.Names(), namesB.Names()); diff != "" {
		fmt.Fprintf(os.Stderr, "Columns A: %v\n", namesA.Names())
		fmt.Fprintf(os.Stderr, "Columns B: %v\n", namesB.Names())
		msg := fmt.Sprintf("Compare() Columns mismatch (-a +b):\n%s", diff)
		if NoColumnIntersect {
			return Information{}, errors.New(msg)
		} else {
			fmt.Fprintln(os.Stderr, msg)
			rawColNames = intersect(namesA.RawNames(), namesB.RawNames())
			log.Printf("intersecting columns: %s", rawColNames)
		}
	}

	if NoSort {
		rawColNames = append([]string{"rowid"}, rawColNames...)
	}

	countA, err := Count(info.table, a, false)
	if err != nil {
		return Information{}, fmt.Errorf("Count(A): %w", err)
	}

	countB, err := Count(info.table, b, false)
	if err != nil {
		return Information{}, fmt.Errorf("Count(B): %w", err)
	}

	if countA == 0 {
		return Information{}, fmt.Errorf("Conn A has zero rows")
	}

	if countB == 0 {
		return Information{}, fmt.Errorf("Conn B has zero rows")
	}

	maxRowid, err := Count(info.table, a, true)
	if err != nil {
		return Information{}, fmt.Errorf("Count(A): %w", err)
	}

	maxRowidB, err := Count(info.table, b, true)
	if err != nil {
		return Information{}, fmt.Errorf("Count(B): %w", err)
	}

	if maxRowidB > maxRowid {
		maxRowid = maxRowidB
	}

	if countA != countB {
		RowCountDifference = countA - countB
		msg := fmt.Sprintf("Rowcount for tables do not match. A: %d B: %d\nDifference: (A-B) %d", countA, countB, RowCountDifference)
		if !IgnoreCount {
			return Information{}, errors.New(msg)
		}

		log.Printf("WARN: %s", msg)

		NoSort = false
		log.Printf("Sorting enabled due to rowcount not matching")
	}

	var usedCols []string
	var indexNameA string
	if !NoSort {
		usedCols, indexNameA, err = AddIndex(info.table, rawColNames, a)
		if err != nil {
			log.Printf("Skipping adding an index, could not AddIndex(A): %s", err)
		}

		if len(usedCols) > 0 {
			_, _, err = AddIndex(info.table, usedCols, b)
			if err != nil {
				log.Printf("Skipping adding an index, could not AddIndex(B): %s", err)
			}
		}
	}

	pv = progressbar.Default(int64(countA), "Comparing")
	if !ProgressEnabled {
		pv = progressbar.NewOptions(countA, progressbar.OptionSetVisibility(false))
	}

	return Information{
		table:     info.table,
		fileA:     info.fileA,
		fileB:     info.fileB,
		count:     countA,
		maxRowid:  maxRowid,
		rawNames:  rawColNames,
		usedCols:  usedCols,
		indexName: indexNameA,
	}, nil
}

// Only Table must be set if Information is directly called
type Information struct {
	table     string
	fileA     string
	fileB     string
	count     int
	maxRowid  int
	rawNames  []string
	usedCols  []string
	indexName string
}

// Compare takes 2 SQLite3 connections and compares the data
// It returns an error when data is not the same
//
// If Information is set the connections are ignored.
//
// From both conns it queries for all columns names using
// `pragma table_info()` and sorts these in alphabetical order.
// Using that output it compares if both conns have the same
// column names available. If not it errors out early.
//
// It checks if both tables contain the same rowcount. If not
// it returns an error.
//
// Finally it queries the table using a select on the column names
// it sorted alphabetically using an ORDER BY with the same column
// names in alphabetical order. This ensures both sets of data are
// when available in both are in the same order.
// Comparing row for row it will output when a difference is found.
//
// If null is set to false it compares using TEXT fields and a
// NULL is automatically converted to an empty text field.
// If this is unwanted set null to be true.
func Compare(info Information, a, b *sqlite3.Conn, null bool) error {
	var err error
	if a == nil || b == nil {
		info, err = PreflightFiles(info.table, info.fileA, info.fileB, null)
		if err != nil {
			return fmt.Errorf("PreflightFiles(): %s", err)
		}
	}
	if info.count == 0 {
		info, err = Preflight(info, a, b, null)
		if err != nil {
			return fmt.Errorf("Preflight(): %s", err)
		}
	}

	if info.indexName != "" {
		defer RemoveIndex(info.indexName, a)
		defer RemoveIndex(info.indexName, b)
	}

	errorCount := new(int32)
	max := Limit

	limit := 0
	offset := 0
	left := info.count
	// without an ORDER BY the rows are read in rowid ranges, and rowids can have gaps
	if NoSort || len(info.usedCols) == 0 {
		left = info.maxRowid + 1
	}

	errs, _ := errgroup.WithContext(context.Background())
	errs.SetLimit(MaxWorkers)

	for {
		if atomic.LoadInt32(errorCount) > int32(MaxErrors) {
			break
		}

		if left <= 0 {
			break
		}
		if left > max {
			limit = max
		}

		// closure rules apply
		// https://go.dev/doc/faq#closures_and_goroutines
		l := limit
		o := offset

		errs.Go(func() error {
			return compare(info, a, b, errorCount, l, o, null)
		})

		left -= max
		offset += max
	}

	if err := errs.Wait(); err != nil {
		return fmt.Errorf("Error in sub-routine compare: %w", err)
	}

	return nil
}

type Checksum struct {
	Hash string
}

func (c *Checksum) Callback(columnData []string) int {
	// we do a copy as the backing string gets unset at the end of this callback
	c.Hash = strings.Clone(columnData[0])

	return 0
}

func compare(info Information, a, b *sqlite3.Conn, errorCount *int32, limit, offset int, null bool) error {
	var err error
	if !memoryTable {
		a = nil
		b = nil
	}

	if a == nil {
		a, err = Open(info.fileA)
		if err != nil {
			return fmt.Errorf("Open(A): %s", err)
		}
		defer a.Close()
	}

	if b == nil {
		b, err = Open(info.fileB)
		if err != nil {
			return fmt.Errorf("Open(B): %s", err)
		}
		defer b.Close()
	}

	var query string
	if NoSort {
		query = QueryForIndexed(info.rawNames, []string{}, info.table, limit, offset)
	} else {
		query = QueryForIndexed(info.rawNames, info.usedCols, info.table, limit, offset)
	}

	// here we can first do
	// hex(sha3_query(query))
	// and compare both, if correct return nil
	var hexA, hexB Checksum
	hexQuery := fmt.Sprintf("select hex(sha3_query('%s'));", query)
	a.ExecWithDataCallback(hexQuery, hexA.Callback)
	b.ExecWithDataCallback(hexQuery, hexB.Callback)

	if (len(hexA.Hash) != 0 && len(hexB.Hash) != 0) && hexA.Hash == hexB.Hash {
		pv.Add(limit)
		return nil
	}

	stmtA, err := a.Prepare(query)
	if err != nil {
		return fmt.Errorf("Prepare(A): %w", err)
	}

	stmtB, err := b.Prepare(query)
	if err != nil {
		return fmt.Errorf("Prepare(B): %w", err)
	}

	defer stmtA.Close()
	defer stmtB.Close()

	err = compareStmts(stmtA, stmtB, len(info.rawNames), null, info.rawNames)
	if err != nil {
		if errors.Is(err, IncorrectParsing) {
			return IncorrectParsing
		}
		atomic.AddInt32(errorCount, 1)
		if int(atomic.LoadInt32(errorCount)) > MaxErrors {
			return fmt.Errorf("WARN (Error count: %d, MaxErrors: %d): %w", atomic.LoadInt32(errorCount), MaxErrors, err)
		}

		log.Printf("WARN (Error count: %d, MaxErrors: %d): %s", atomic.LoadInt32(errorCount), MaxErrors, err)
	}

	return nil
}

var (
	StillHasRows     error = fmt.Errorf("still has rows")
	IncorrectParsing error = fmt.Errorf("incorrect parsing")
)

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
				return fmt.Errorf("B: %w", StillHasRows)
			}
			break
		}

		if !hasRowB {
			if hasRowA {
				return fmt.Errorf("A: %w", StillHasRows)
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

			if !NoCRIgnore {
				for i := range bufA {
					if v, ok := bufA[i].(string); ok {
						if len(bufB) < i {
							break
						}
						if strings.Contains(v, "\r") {
							if vb, ok := bufB[i].(string); ok {
								if strings.Contains(vb, "\n") && !strings.Contains(vb, "\r") {
									bufA[i] = strings.ReplaceAll(v, "\r", "\n")
								}
							}
						}
					}
				}

				diff = cmp.Diff(bufA, bufB)
			}

			// if file A contains for instance
			// col1: foo\",bar
			// col2:
			// and file B contains
			// col1: foo\
			// col2: bar
			//
			// it means the escaped "enclosure" was falsely identified by PHP
			// the enclosure was *not* escaped, just a backslash appeared in at
			// the end of the column.
			// we can check if colB is inside colA, if so we can ignore it
			if !NoEscapedIgnore {
				for i := range bufA {
					if v, ok := bufA[i].(string); ok {
						if len(bufB) < i {
							break
						}
						// we thought it only occurs on \", but can also just be "
						if !strings.Contains(v, `"`) {
							continue
						}
						if vb, ok := bufB[i].(string); ok {
							if strings.HasPrefix(v, vb) && len(v) > len(vb) {
								// reset diff
								// diff = ""

								var buf strings.Builder
								buf.WriteString("Incorrect column parsing from A detected\n")
								if NoSort {
									buf.WriteString(fmt.Sprintf("%50s", "rowid"))
									buf.WriteString("\tA\t")
									buf.WriteString(fmt.Sprintf("%s\n", bufB[0]))
								}
								buf.WriteString(fmt.Sprintf("%50s", columnNames[i]))
								buf.WriteString("\tA\t")
								buf.WriteString(v)
								buf.WriteString(fmt.Sprintf("\n%50s", ""))
								buf.WriteString("\tB\t")
								buf.WriteString(vb)
								buf.WriteString("\n\n")
								buf.WriteString(diff)
								buf.WriteString("\n\n")
								fmt.Fprintf(os.Stderr, "%s\n", buf.String())
								return IncorrectParsing
							}
						}
					}
				}
			}

			if !NoEnclosureIgnore {
				for i := range bufA {
					if v, ok := bufA[i].(string); ok {
						if NoTrimWhitespace {
							bufA[i] = strings.ReplaceAll(v, Enclosure.String(), "")
						} else {
							bufA[i] = strings.TrimSpace(strings.ReplaceAll(v, Enclosure.String(), ""))
						}
					}
				}
				for i := range bufB {
					if v, ok := bufB[i].(string); ok {
						if NoTrimWhitespace {
							bufB[i] = strings.ReplaceAll(v, Enclosure.String(), "")
						} else {
							bufB[i] = strings.TrimSpace(strings.ReplaceAll(v, Enclosure.String(), ""))
						}
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
