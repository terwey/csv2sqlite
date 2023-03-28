package main

import (
	"os"
	"strings"
	"testing"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/cmd/csv2sqlite/debug"
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
		newline   bool
		php       bool
		params    []string
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
			name:      "head.csv exclude 0,2",
			input:     baseDir + "tests/head.csv",
			output:    tmpDir + "head-exclude.sqlite",
			delimiter: characters.Character_PIPE,
			enclosure: characters.Character_QUOTATION_MARK,
			want:      baseDir + "tests/head-exclude.sqlite",
			null:      false,
			params: []string{
				"--exclude",
				"0,2",
			},
		},
		{
			name:      "528874",
			input:     baseDir + "tests/528874.csv",
			output:    tmpDir + "528874.sqlite",
			delimiter: characters.Character_COMMA,
			enclosure: characters.Character_QUOTATION_MARK,
			want:      baseDir + "tests/528874.sqlite",
			debug:     false,
		},
		{
			name:      "emoji",
			input:     baseDir + "tests/emoji.csv",
			output:    tmpDir + "emoji.sqlite",
			want:      baseDir + "tests/emoji.sqlite",
			enclosure: []byte("🍆"),
			delimiter: []byte("💦"),
		},
		{
			name:      "csvSample.512034.15616004",
			input:     baseDir + "tests/csvSample.512034.11822895.csv",
			output:    tmpDir + "csvSample.512034.11822895.sqlite",
			want:      baseDir + "tests/csvSample.512034.11822895.sqlite",
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
			delimiter: characters.Character_INVISIBLE_SEPARATOR,
			newline:   false,
		},
		{
			name:      "csvSample.1086690.11935102.csv",
			input:     baseDir + "tests/csvSample.1086690.11935102.csv",
			output:    tmpDir + "csvSample.1086690.11935102.sqlite",
			want:      baseDir + "tests/csvSample.1086690.11935102.sqlite",
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
			delimiter: characters.Character_COMMA,
			newline:   false,
		},
		{
			name:      "csvSample.1094837.11968369.csv",
			input:     baseDir + "tests/csvSample.1094837.11968369.csv",
			output:    tmpDir + "csvSample.1094837.11968369.sqlite",
			want:      baseDir + "tests/csvSample.1094837.11968369.sqlite",
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     true,
			delimiter: characters.Character_COMMA,
			newline:   false,
		},
		{
			name:      "empty-cols-before-header.csv",
			input:     baseDir + "tests/empty-cols-before-header.csv",
			output:    tmpDir + "empty-cols-before-header.sqlite",
			want:      baseDir + "tests/empty-cols-before-header.sqlite",
			enclosure: characters.Character_EMPTY,
			debug:     false,
			delimiter: characters.Character_COMMA,
			newline:   false,
		},
		{
			name:      "empty-rows-in-content.csv",
			input:     baseDir + "tests/empty-rows-in-content.csv",
			output:    tmpDir + "empty-rows-in-content.sqlite",
			want:      baseDir + "tests/empty-rows-in-content.sqlite",
			enclosure: characters.Character_EMPTY,
			debug:     false,
			delimiter: characters.Character_COMMA,
			newline:   false,
		},
		{
			name:      "empty-rows-in-content.csv with --rowid",
			input:     baseDir + "tests/empty-rows-in-content.csv",
			output:    tmpDir + "empty-rows-in-content.sqlite",
			want:      baseDir + "tests/empty-rows-in-content.sqlite",
			enclosure: characters.Character_EMPTY,
			debug:     false,
			delimiter: characters.Character_COMMA,
			newline:   false,
			params:    []string{"--rowid"},
		},
		{
			name:      "empty-cols-in-content-with-delimiters.csv with --rowid",
			input:     baseDir + "tests/empty-cols-in-content-with-delimiters.csv",
			output:    tmpDir + "empty-rows-in-content.sqlite",
			want:      baseDir + "tests/empty-rows-in-content.sqlite",
			enclosure: characters.Character_EMPTY,
			debug:     true,
			delimiter: characters.Character_COMMA,
			newline:   false,
			params:    []string{"--rowid"},
		},
		{
			name:      "cff112e3b8d3004086fe28b69c922f2d.csv",
			input:     baseDir + "tests/cff112e3b8d3004086fe28b69c922f2d.csv",
			output:    tmpDir + "cff112e3b8d3004086fe28b69c922f2d.sqlite",
			want:      baseDir + "tests/cff112e3b8d3004086fe28b69c922f2d.sqlite",
			enclosure: characters.Character_QUOTATION_MARK,
			delimiter: characters.Character_COMMA,
			newline:   false,
			params:    []string{"--start", "10"},
		},
		{
			name:      "6f522d299731f73aad537914cfe0d68b.csv",
			input:     baseDir + "tests/6f522d299731f73aad537914cfe0d68b.csv",
			output:    tmpDir + "6f522d299731f73aad537914cfe0d68b.sqlite",
			want:      baseDir + "tests/6f522d299731f73aad537914cfe0d68b.sqlite",
			enclosure: characters.Character_QUOTATION_MARK,
			debug:     false,
			delimiter: characters.Character_PIPE,
			newline:   false,
		},
		{
			name:      "csvSample.474672.11909590.csv",
			input:     baseDir + "tests/csvSample.474672.11909590.csv",
			output:    tmpDir + "csvSample.474672.11909590.sqlite",
			want:      baseDir + "tests/csvSample.474672.11909590.sqlite",
			enclosure: characters.Character_EMPTY,
			debug:     false,
			delimiter: characters.Character("<#>"),
			newline:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p.Characters.DelimiterCount = 0
			p.Start = 0
			p.Characters = &util.CharacterConfig{}

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

			if tt.newline {
				os.Args = append(os.Args,
					"--nd",
				)
			}

			if debug.DebugLog {
				if tt.debug {
					debug.DebugLog = tt.debug
					os.Args = append(os.Args, "--debug")
				}
			}

			os.Args = append(os.Args, tt.params...)

			main()

			if tt.want != "" {
				if err := compare.CompareFiles("data", tt.want, tt.output, tt.null); err != nil {
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
	fu := util.Usage(getFlags())

	var s strings.Builder

	s.WriteString("## Usage csv2sqlite\n\n.Usage\n")
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
	fu := util.Usage(getFlags())

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

	err := os.WriteFile(baseDir+"/docs/csv2sqlite.7.adoc", []byte(s.String()), 0644)
	if err != nil {
		t.Errorf("could not write manpage adoc: %s", err)
	}
}
