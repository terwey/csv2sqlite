// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

//go:build debug

package debug

import (
	"fmt"
	"os"
)

const Ultra = false

func State() bool {
	return true
}

func LogData(name string, data []byte, max int) {
	if DebugLog {
		if len(data) < max {
			Debug("%s (%d) data:\n%s", name, len(data), data)
			return
		}

		Debug("%s (%d) start data[:%d]\n\n%s\nend data[len(data)-%d:]\n\n%s", name, len(data), max, data[:max], max, data[len(data)-max:])
	}
}

func Debug(format string, v ...interface{}) {
	if DebugLog {
		fmt.Fprintf(os.Stderr, format+"\n----\n", v...)
	}
}

func DebugColumns(cols [][]byte) {
	if DebugLog {
		if len(cols) == 0 {
			return
		}
		// log.Printf("Columns")
		// Debug("Columns: [%3d]\t%s", len(cols)-1, cols[len(cols)-1])
		for i := range cols {
			fmt.Fprintf(os.Stderr, "Column [%3d]\t%s\n", i, cols[i])
		}
		// log.Printf("End of Columns")
	}
}

func LogReader(data []byte, count int, colCount int) {
	if LogCSVReader {
		fmt.Fprintf(os.Stderr, "CSVReader used for (count) %d vs (colCount) %d, data:\n%s\n----\n", count, colCount, data)
	}
}
