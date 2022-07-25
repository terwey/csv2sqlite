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
func Count(table string, conn *sqlite3.Conn, rowid bool) (result int, err error) {
	var stmt *sqlite3.Stmt
	if rowid {
		stmt, err = conn.Prepare("select max(rowid) from " + table)
	} else {
		stmt, err = conn.Prepare("select count(*) from " + table)
	}
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
