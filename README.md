# csv2sqlite

## Description

**csv2sqlite** converts CSV files into SQLite databases.

- Delimiters and enclosures can be any Unicode string: `<#>`, an emoji, a multi-byte invisible separator. They’re detected automatically, along with the line ending.

- Fast, with low memory usage.

- Handles extremely large files.

- Reads its input with [dsv-parser](https://github.com/terwey/dsv-parser), the parser of this repository as a Go library of its own.

**sqlite-compare** checks two SQLite files against each other and stops at the first difference.

## Installation

Build from source, see [Compilation](#_compilation).

## Compilation

Requirement is a working Go installation, see <https://go.dev>

<div class="formalpara-title">

**csv2sqlite**

</div>

``` bash
make csv2sqlite
# or for mac
make csv2sqlite_mac
```

<div class="formalpara-title">

**sqlite-compare**

</div>

``` bash
make sqlite-compare
```

## Usage csv2sqlite

<table>
<caption>Usage</caption>
<colgroup>
<col style="width: 33%" />
<col style="width: 16%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th style="text-align: left;">Name / Shorthand</th>
<th style="text-align: left;">Default</th>
<th style="text-align: left;">Usage</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td style="text-align: left;"><p><code>--version</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Prints the version</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--input</code><br />
<code>-i</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Input CSV file</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--output</code><br />
<code>-o</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Output CSV (or SQLite) file</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--delimiter</code><br />
<code>-d</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Set the delimiter in case auto-detect should not be used. Can be one of:</p>
<ul>
<li><p>APOSTROPHE</p></li>
<li><p>ASTERISK</p></li>
<li><p>CARET</p></li>
<li><p>COMMA</p></li>
<li><p>DOLLAR</p></li>
<li><p>EMPTY</p></li>
<li><p>INVISIBLE_SEPARATOR</p></li>
<li><p>PIPE</p></li>
<li><p>POUND</p></li>
<li><p>QUOTATION_MARK</p></li>
<li><p>QUOTATION_MARK_DOUBLE</p></li>
<li><p>QUOTATION_MARK_TRIPLE</p></li>
<li><p>RECORD_SEPARATOR</p></li>
<li><p>RECORD_SEPARATOR_SB</p></li>
<li><p>RECORD_SEPARATOR_SB</p></li>
<li><p>SEMICOLON</p></li>
<li><p>TAB</p></li>
<li><p>TILDE</p></li>
<li><p>UNIT_SEPARATOR</p></li>
</ul></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--enclosure</code><br />
<code>-e</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Set the enclosure in case auto-detect should not be used, can be one of:</p>
<ul>
<li><p>APOSTROPHE</p></li>
<li><p>ASTERISK</p></li>
<li><p>CARET</p></li>
<li><p>COMMA</p></li>
<li><p>DOLLAR</p></li>
<li><p>EMPTY</p></li>
<li><p>INFORMATION_SEPARATOR_TWO</p></li>
<li><p>INVISIBLE_SEPARATOR</p></li>
<li><p>PIPE</p></li>
<li><p>POUND</p></li>
<li><p>QUOTATION_MARK</p></li>
<li><p>QUOTATION_MARK_DOUBLE</p></li>
<li><p>QUOTATION_MARK_TRIPLE</p></li>
<li><p>RECORD_SEPARATOR</p></li>
<li><p>RECORD_SEPARATOR_SB</p></li>
<li><p>SEMICOLON</p></li>
<li><p>TAB</p></li>
<li><p>TILDE</p></li>
<li><p>UNIT_SEPARATOR</p></li>
</ul></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--start</code></p></td>
<td style="text-align: left;"><p>1</p></td>
<td style="text-align: left;"><p>The starting position where the header is, useful for CSVs produced by Excel.</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--skip</code></p></td>
<td style="text-align: left;"><p>-1</p></td>
<td style="text-align: left;"><p>Rows to skip after the header</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--limit</code></p></td>
<td style="text-align: left;"><p>-1</p></td>
<td style="text-align: left;"><p>Amount of rows to process</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--literal-delimiter</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Sets the delimiter to use to a literal string, if set the --delimiter is ignored</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--literal-enclosure</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Sets the enclosure to use to a literal string, if set the --enclosure is ignored</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--php</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>CSV file was produced by PHP, enclosure escaped by backslash</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--exclude</code></p></td>
<td style="text-align: left;"><p>[]</p></td>
<td style="text-align: left;"><p>Column indices to exclude from conversion</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--newline</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Set the newline to be used in case auto-detection should not be used. Can be one of:</p>
<ul>
<li><p>LF</p></li>
<li><p>CR</p></li>
<li><p>CRLF</p></li>
</ul></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--sqlite</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Output SQLite file</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--nd</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Input CSV is newline delimited and does not have newlines in the middle of a field</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--break-count</code></p></td>
<td style="text-align: left;"><p>15000</p></td>
<td style="text-align: left;"><p>Causes the sqlite writer to stop writing at the specified amount of rows.
Can cause the merge to be triggered earlier. The first writer stops at N/3</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--commit-rate</code><br />
<code>-r</code></p></td>
<td style="text-align: left;"><p>5000</p></td>
<td style="text-align: left;"><p>Flush rate for comitting rows to disk</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--buffer-size</code><br />
<code>-b</code></p></td>
<td style="text-align: left;"><p>15625</p></td>
<td style="text-align: left;"><p>Buffer size for reading from disk in kilobytes</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--threads</code></p></td>
<td style="text-align: left;"><p>16</p></td>
<td style="text-align: left;"><p>Threads to use when writing output, default is set to the amount of cores</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--cpu-profile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a cpu-profile ouput for pprof</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--heap-profile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a heap-profile ouput for pprof</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--trace-profile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a trace profile for go tool trace</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--fgprofile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a fgprofile ouput for pprof, see <a href="https://github.com/felixge/fgprof">https://github.com/felixge/fgprof</a></p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--log-header</code></p></td>
<td style="text-align: left;"><p>0</p></td>
<td style="text-align: left;"><p>Logs the header for the collected data, pass 1 to output it. Can be used together with seq</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--quiet</code><br />
<code>-q</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Quiet mode do not output performance data</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--csv-report</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Performance data output in TSV format, default is JSON</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--fixed-eol</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>EOL character (CR/CRLF/LF) doesn’t change after auto-detect</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--drop</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Drops all outgoing data, for perf testing disk reading</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--rowid</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>When the order of the CSV file must be kept the same in the resulting SQLite, this flag will
cause each row in the SQLite to receive a RowID for SQLite.</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--progress</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Shows progress</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--scratch-dir</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Directory for storing the smaller SQLite files before merging them to the target. Same as output if left blank.</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--csv-enclosure</code></p></td>
<td style="text-align: left;"><p>8</p></td>
<td style="text-align: left;"><p>When to use an enclosure for the CSV output:</p>
<p>1 None
2 Required
4 SkipEmpty
8 All</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--overflow</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Overflows the data. Will abort processing.</p></td>
</tr>
</tbody>
</table>

Usage

## Usage sqlite-compare

<table>
<caption>Usage</caption>
<colgroup>
<col style="width: 33%" />
<col style="width: 16%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th style="text-align: left;">Name / Shorthand</th>
<th style="text-align: left;">Default</th>
<th style="text-align: left;">Usage</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td style="text-align: left;"><p><code>--sqlite-a</code><br />
<code>-a</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>SQLite file A to compare with</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--sqlite-b</code><br />
<code>-b</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>SQLite file B to compare against</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--table</code><br />
<code>-t</code></p></td>
<td style="text-align: left;"><p>data</p></td>
<td style="text-align: left;"><p>Table to compare</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--null</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>If null is set to false it compares using TEXT fields
and a NULL is automatically converted to an empty text field.
If this is unwanted set null to be true.</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--write-results</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Writes a result file.
$a.passed when no difference was found.
$a.failed when a difference was found, file will contain the error.</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--progress</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Shows progress</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--ignore-count</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Ignore the rowcount, show the rows that are different</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--max-errors</code></p></td>
<td style="text-align: left;"><p>0</p></td>
<td style="text-align: left;"><p>Maximum amount of errors to allow before exiting</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--limit</code></p></td>
<td style="text-align: left;"><p>50000</p></td>
<td style="text-align: left;"><p>Row limit per query</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--no-ignore-crlf</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not ignore CRLF difference from A</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--no-ignore-cr</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not ignore CR difference from A</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--no-ignore-enclosure</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not strip enclosures from both to ignore</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--no-escape-ignore</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not ignore incorrect detected enclosures from A to B (PHP parsing error)</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--no-sort</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not order data for indexed columns</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--no-column-intersect</code></p></td>
<td style="text-align: left;"><p>false</p></td>
<td style="text-align: left;"><p>Do not intersect on column names when they differ</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--enclosure</code><br />
<code>-e</code></p></td>
<td style="text-align: left;"><p>"</p></td>
<td style="text-align: left;"><p>Enclosure to use</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--threads</code></p></td>
<td style="text-align: left;"><p>16</p></td>
<td style="text-align: left;"><p>Thread limit</p></td>
</tr>
<tr class="even">
<td style="text-align: left;"><p><code>--cpu-profile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a cpu-profile ouput for pprof</p></td>
</tr>
<tr class="odd">
<td style="text-align: left;"><p><code>--heap-profile</code></p></td>
<td style="text-align: left;"></td>
<td style="text-align: left;"><p>Filename for a heap-profile ouput for pprof</p></td>
</tr>
</tbody>
</table>

Usage

## Documentation

Currently no additional documentation is yet available.

Inside `docs/` there is a file `readme.adoc` which can be converted to Markdown
for Github using

<div class="formalpara-title">

**README.MD**

</div>

``` bash
make markdown
```

## Configuration

The only configuration available are the parameters during runtime, see [???](#Usage)

## Testing, Code Coverage, Static Analysis

Run the tests with coverage:

``` bash
go test -cover -coverprofile coverage.out ./... -count=1
```

## Deployment

See [Installation](#_installation)

## Versioning

This library follows [semantic versioning](https://semver.org/).
