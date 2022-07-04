package main

import (
	"os"
	"testing"

	"github.com/terwey/csv2sqlite/characters"
	"github.com/terwey/csv2sqlite/sqlitehelper"
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
				if err := sqlitehelper.CompareFiles("data", tt.output, tt.want, tt.null); err != nil {
					t.Errorf("CompareFiles() %s: error: %s", tt.name, err)
				}

				if err := os.Remove(tt.output); err != nil {
					t.Logf("Could not clean up output file (%s) from test: %s", tt.output, tt.name)
				}
			}
		})
	}
}
