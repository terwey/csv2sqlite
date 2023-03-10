package main

import (
	"os"
	"strings"
	"testing"

	"github.com/terwey/csv2sqlite/util"
)

var baseDir = "../../"

func Test_docs(t *testing.T) {
	fu := util.Usage(getFlags())

	var s strings.Builder

	s.WriteString("## Usage sqlite-compare\n\n.Usage\n")
	s.WriteString("[cols=\"2,1,3\"]\n")
	s.WriteString("|===\n")
	s.WriteString("| Name / Shorthand | Default | Usage \n\n")

	for _, f := range fu {
		s.WriteString(f.GetRow())
	}

	s.WriteString("\n|===\n")

	err := os.WriteFile(baseDir+"/docs/usage-compare.adoc", []byte(s.String()), 0644)
	if err != nil {
		t.Errorf("could not write usage adoc: %s", err)
	}
}

func Test_manpage(t *testing.T) {
	fu := util.Usage(getFlags())

	var s strings.Builder

	s.WriteString("= sqlite-compare(7)\n")
	s.WriteString("Yorick Terweijden\n")

	s.WriteString(":doctype: manpage\n")
	s.WriteString(":manmanual: sqlite-compare\n")
	s.WriteString(":mansource: sqlite-compare\n")
	s.WriteString(":man-linkstyle: pass:[blue R < >]\n")
	s.WriteString(":version: undefined\n")
	s.WriteString("{version}\n\n")

	s.WriteString("== Name\n\n")
	s.WriteString("sqlite-compare - compare two SQLite files\n\n")

	s.WriteString("== Synopsis\n\n")
	s.WriteString("sqlite-compare [_OPTION_]... -a _SQLITE_FILE_ -b _SQLITE_FILE_\n\n")

	s.WriteString("== Options\n\n")

	for _, f := range fu {
		s.WriteString(f.GetMan())
	}

	s.WriteString("\n\n== Resources\n\n")
	s.WriteString("*Repository:* https://github.com/terwey/csv2sqlite\n")

	err := os.WriteFile(baseDir+"/docs/sqlite-compare.7.adoc", []byte(s.String()), 0644)
	if err != nil {
		t.Errorf("could not write manpage adoc: %s", err)
	}
}
