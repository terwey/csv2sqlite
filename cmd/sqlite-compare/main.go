// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/sqlitehelper/compare"
	"github.com/spf13/pflag"
)

var (
	flags *pflag.FlagSet
	a     string
	b     string
	table string
	null  bool
	write bool

	cpuprofile  string
	heapprofile string

	showProgress bool

	enclosure string
)

func getFlags() *pflag.FlagSet {
	flags = pflag.NewFlagSet("sqlitecompare", pflag.ContinueOnError)
	flags.SortFlags = false
	flags.StringVarP(&a, "sqlite-a", "a", "", "SQLite file A to compare with")
	flags.StringVarP(&b, "sqlite-b", "b", "", "SQLite file B to compare against")
	flags.StringVarP(&table, "table", "t", "data", "Table to compare")
	flags.BoolVar(&null, "null", false, "If null is set to false it compares using TEXT fields\n and"+
		" a NULL is automatically converted to an empty text field. \n"+
		"If this is unwanted set null to be true.")
	flags.BoolVar(&write, "write-results", false, "Writes a result file.\n $a.passed when no "+
		"difference was found.\n $a.failed when a difference was found, file will contain the error.")
	flags.BoolVar(&showProgress, "progress", false, "Shows progress")
	flags.BoolVar(&compare.IgnoreCount, "ignore-count", false, "Ignore the rowcount, show the rows that are different")
	flags.IntVar(&compare.MaxErrors, "max-errors", 0, "Maximum amount of errors to allow before exiting")
	flags.IntVar(&compare.Limit, "limit", 50000, "Row limit per query")
	flags.BoolVar(&compare.NoCRLFIgnore, "no-ignore-crlf", false, "Do not ignore CRLF difference from A")
	flags.BoolVar(&compare.NoCRIgnore, "no-ignore-cr", false, "Do not ignore CR difference from A")
	flags.BoolVar(&compare.NoEnclosureIgnore, "no-ignore-enclosure", false, "Do not strip enclosures from both to ignore")
	flags.BoolVar(&compare.NoEscapedIgnore, "no-escape-ignore", false, "Do not ignore incorrect detected enclosures from"+
		" A to B (PHP parsing error)")
	flags.BoolVar(&compare.NoSort, "no-sort", false, "Do not order data for indexed columns")
	flags.BoolVar(&compare.NoColumnIntersect, "no-column-intersect", false, "Do not intersect on column names when they differ")
	flags.StringVarP(&enclosure, "enclosure", "e", "\"", "Enclosure to use")
	flags.IntVar(&compare.MaxWorkers, "threads", compare.MaxWorkers, "Thread limit")
	flags.StringVar(&cpuprofile, "cpu-profile", "", "Filename for a cpu-profile ouput for pprof")
	flags.StringVar(&heapprofile, "heap-profile", "", "Filename for a heap-profile ouput for pprof")

	return flags
}

func init() {
	flags = getFlags()
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "\nCompare takes 2 SQLite3 databases and compares the data\n"+
			"It returns an error when data is not the same\n\n"+

			"From both conns it queries for all columns names using\n"+
			" `pragma table_info()` and sorts these in alphabetical order.\n"+
			" Using that output it compares if both conns have the same\n"+
			" column names available. If not it errors out early.\n\n"+

			"It checks if both tables contain the same rowcount. If not\n"+
			" it returns an error.\n\n"+

			"Finally it queries the table using a select on the column names\n"+
			" it sorted alphabetically using an ORDER BY with the same column\n"+
			" names in alphabetical order. This ensures both sets of data are\n"+
			" (when available) in both are in the same order.\n"+
			" Comparing row for row it will output when a difference is found.\n\n")
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if a == "" {
		log.Fatal("a cannot be empty")
	}
	if b == "" {
		log.Fatal("b cannot be empty")
	}

	if a == b {
		log.Fatalf("A (%s) and B (%s) are the same file", a, b)
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("Could not create cpuprofile: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if heapprofile != "" {
		defer WriteHeapProfile(heapprofile)
	}

	if enclosure != `` {
		encl, err := characters.CharacterFromString(enclosure)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
			os.Exit(100)
		}
		compare.Enclosure = append(characters.Character{}, encl...)
	}

	compare.ProgressEnabled = showProgress

	err := compare.CompareFiles("data", a, b, null)
	// an accepted difference passes, its reason goes into the .passed file
	var accepted string
	if err != nil {
		if errors.Is(err, compare.StillHasRows) && compare.RowCountDifference == -1 {
			fmt.Fprintf(os.Stderr, "Difference was acceptable: %s\n", err)
		} else if errors.Is(err, compare.IncorrectParsing) {
			fmt.Fprintf(os.Stderr, "Incorrect parsing detected: %s\n", err)
		} else {
			if write {
				log.Printf("%s.failed written\n", a)
				os.WriteFile(a+".failed", []byte(err.Error()), 0666)
			}
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		accepted = err.Error()
	}

	if write {
		log.Printf("%s.passed written\n", a)
		os.WriteFile(a+".passed", []byte(accepted), 0666)
	}
	if accepted != "" {
		log.Printf("Only accepted differences found")
		return
	}
	log.Printf("No differences found")
}

func WriteHeapProfile(name string) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("Could not create heapprofile: %s", err)
	}
	err = pprof.WriteHeapProfile(f)
	if err != nil {
		log.Printf("could not write heap profile")
		return
	}

	log.Printf("wrote heap-profile: %s", f.Name())
}
