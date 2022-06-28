// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/terwey/csv2sqlite/sqlitehelper"
	"github.com/spf13/pflag"
)

var (
	flags *pflag.FlagSet
	a     string
	b     string
	table string
)

func init() {
	flags = pflag.NewFlagSet("sqlitecompare", pflag.ContinueOnError)
	flags.StringVarP(&a, "sqlite-a", "a", "", "SQLite file A to compare with")
	flags.StringVarP(&b, "sqlite-b", "b", "", "SQLite file B to compare against")
	flags.StringVarP(&table, "table", "t", "data", "Table to compare")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	err := sqlitehelper.CompareFiles("data", a, b)
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("No differences found")

	os.Exit(0)
}
