// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

//go:build !debug

package debug

const Ultra = true

func State() bool {
	return false
}

func LogData(name string, data []byte, max int) {
}

func Debug(format string, v ...any) {
}

func Printf(format string, v ...any) {
}

func DebugColumns(cols [][]byte) {
}

func LogReader(data []byte, count int, colCount int) {
}

func LogDropped(format string, v ...any) {}
