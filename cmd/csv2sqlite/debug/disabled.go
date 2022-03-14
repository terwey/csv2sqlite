// SPDX-License-Identifier: MIT
// Copyright (c) 2022 Yorick Terweijden

//go:build !debug

package debug

func State() bool {
	return false
}

func LogData(name string, data []byte, max int) {
}

func Debug(format string, v ...interface{}) {
}

func DebugColumns(cols [][]byte) {
}
