package main

import (
	"os"
	"strings"
	"testing"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/sqlitehelper/compare"
	"github.com/terwey/csv2sqlite/util"
)

var tmpDir = "/tmp/"
var baseDir = "../../"

// To have some form of integration test we will write a SQLite file as output
// and then read the file again and compare it to what we expected.
func Test_main(t *testing.T) {

	tests := []struct {
		name      string
		input     string
		output    string
		delimiter characters.Character
		enclosure characters.Character
		want      string
		null      bool
		debug     bool
	}{
		{
			name:      "head.csv",
			input:     baseDir + "tests/head.csv",
			output:    tmpDir + "head.sqlite",
			delimiter: characters.Character_PIPE,
			enclosure: characters.Character_QUOTATION_MARK,
			want:      baseDir + "tests/head.sqlite",
			null:      false,
		},
		{
			name:      "528874",
			input:     baseDir + "tests/528874.csv",
			output:    tmpDir + "528874.sqlite",
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			want:      baseDir + "tests/528874.sqlite",
			debug:     true,
		},
		{
			name:      "emoji",
			input:     baseDir + "tests/emoji.csv",
			output:    tmpDir + "emoji.sqlite",
			want:      baseDir + "tests/emoji.sqlite",
			enclosure: []byte("🍆"),
			delimiter: []byte("💦"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Args = []string{
				"rddb",
				"-i",
				tt.input,
				"-o",
				tt.output,
			}

			if tt.delimiter != nil {
				os.Args = append(
					os.Args,
					"--delimiter",
					tt.delimiter.Name(),
				)
			}

			if tt.enclosure != nil {
				os.Args = append(
					os.Args,
					"--enclosure",
					tt.enclosure.Name(),
				)
			}

			// if tt.debug {
			// 	os.Args = append(os.Args, "--debug")
			// }

			main()

			if tt.want != "" {
				if err := compare.CompareFiles("data", tt.output, tt.want, tt.null); err != nil {
					t.Errorf("CompareFiles() %s: error: %s", tt.name, err)
				}

				if err := os.Remove(tt.output); err != nil {
					t.Logf("Could not clean up output file (%s) from test: %s", tt.output, tt.name)
				}
			}
		})
	}
}

func Test_docs(t *testing.T) {
	fu := util.Usage(flags)

	var s strings.Builder

	s.WriteString("## Usage\n\n.Usage\n")
	s.WriteString("[cols=\"2,1,3\"]\n")
	s.WriteString("|===\n")
	s.WriteString("| Name / Shorthand | Default | Usage \n\n")

	for _, f := range fu {
		s.WriteString(f.GetRow())
	}

	s.WriteString("\n|===\n")

	err := os.WriteFile(baseDir+"/docs/usage.adoc", []byte(s.String()), 0644)
	if err != nil {
		t.Errorf("could not write usage adoc: %s", err)
	}
}

func Test_manpage(t *testing.T) {
	fu := util.Usage(flags)

	var s strings.Builder

	s.WriteString("= csv2sqlite(7)\n")
	s.WriteString("Yorick Terweijden\n")

	s.WriteString(":doctype: manpage\n")
	s.WriteString(":manmanual: csv2sqlite\n")
	s.WriteString(":mansource: csv2sqlite\n")
	s.WriteString(":man-linkstyle: pass:[blue R < >]\n")
	s.WriteString(":version: undefined\n")
	s.WriteString("{version}\n\n")

	s.WriteString("== Name\n\n")
	s.WriteString("csv2sqlite - converts CSV to SQLite\n\n")

	s.WriteString("== Synopsis\n\n")
	s.WriteString("csv2sqlite [_OPTION_]... -i _CSV_FILE_ -o _SQLITE_OUT_FILE_\n\n")

	s.WriteString("include::desc.adoc[]\n\n")

	s.WriteString("== Options\n\n")

	for _, f := range fu {
		s.WriteString(f.GetMan())
	}

	s.WriteString("\n\n== Resources\n\n")
	s.WriteString("*Repository:* https://github.com/terwey/csv2sqlite\n")

	err := os.WriteFile(baseDir+"/docs/manpage.adoc", []byte(s.String()), 0644)
	if err != nil {
		t.Errorf("could not write manpage adoc: %s", err)
	}
}
