# csv2sqlite

## Description

**csv2sqlite** converts CSV files into SQLite databases.

  - The delimiter and the enclosure are detected automatically.
  - Delimiters and enclosures can also be multi-byte characters.
  - Fast, with low memory usage.
  - Handles extremely large files.

**sqlite-compare** checks two SQLite files against each other.

## Usage

<table>
<caption>Usage</caption>
<colgroup>
<col style="width: 33%" />
<col style="width: 16%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th>Name / Shorthand</th>
<th>Default</th>
<th>Usage</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><p><code>--version</code></p></td>
<td><p>false</p></td>
<td><p>Prints the version</p></td>
</tr>
<tr class="even">
<td><p><code>--input</code><br />
<code>-i</code></p></td>
<td></td>
<td><p>Input CSV file</p></td>
</tr>
<tr class="odd">
<td><p><code>--output</code><br />
<code>-o</code></p></td>
<td></td>
<td><p>Output CSV (or SQLite) file</p></td>
</tr>
<tr class="even">
<td><p><code>--delimiter</code><br />
<code>-d</code></p></td>
<td></td>
<td><p>Set the delimiter in case auto-detect should not be used. Can be one of:</p>
<ul>
<li><p>APOSTROPHE</p></li>
<li><p>ASTERISK</p></li>
<li><p>CARET</p></li>
<li><p>COMMA</p></li>
<li><p>DOLLAR</p></li>
<li><p>EMPTY</p></li>
<li><p>PIPE</p></li>
<li><p>POUND</p></li>
<li><p>QUOTATION_MARK</p></li>
<li><p>QUOTATION_MARK_DOUBLE</p></li>
<li><p>QUOTATION_MARK_TRIPLE</p></li>
<li><p>RECORD_SEPARATOR</p></li>
<li><p>SEMICOLON</p></li>
<li><p>TAB</p></li>
<li><p>TILDE</p></li>
<li><p>UNIT_SEPARATOR</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><p><code>--enclosure</code></p></td>
<td></td>
<td><p>Set the enclosure in case auto-detect should not be used, can be one of:</p>
<ul>
<li><p>APOSTROPHE</p></li>
<li><p>ASTERISK</p></li>
<li><p>CARET</p></li>
<li><p>COMMA</p></li>
<li><p>DOLLAR</p></li>
<li><p>EMPTY</p></li>
<li><p>PIPE</p></li>
<li><p>POUND</p></li>
<li><p>QUOTATION_MARK</p></li>
<li><p>QUOTATION_MARK_DOUBLE</p></li>
<li><p>QUOTATION_MARK_TRIPLE</p></li>
<li><p>RECORD_SEPARATOR</p></li>
<li><p>SEMICOLON</p></li>
<li><p>TAB</p></li>
<li><p>TILDE</p></li>
<li><p>UNIT_SEPARATOR</p></li>
</ul></td>
</tr>
<tr class="even">
<td><p><code>--start</code></p></td>
<td><p>0</p></td>
<td><p>The starting position where the header is, useful for CSVs produced by Excel. Only works in combination with newline delimited, -nd enabled if set.</p></td>
</tr>
<tr class="odd">
<td><p><code>--skip</code></p></td>
<td><p>-1</p></td>
<td><p>Rows to skip after the header</p></td>
</tr>
<tr class="even">
<td><p><code>--limit</code></p></td>
<td><p>-1</p></td>
<td><p>Amount of rows to process</p></td>
</tr>
<tr class="odd">
<td><p><code>--literal-delimiter</code></p></td>
<td></td>
<td><p>Sets the delimiter to use to a literal string, if set the --delimiter is ignored</p></td>
</tr>
<tr class="even">
<td><p><code>--literal-enclosure</code></p></td>
<td></td>
<td><p>Sets the enclosure to use to a literal string, if set the --enclosure is ignored</p></td>
</tr>
<tr class="odd">
<td><p><code>--exclude</code></p></td>
<td><p>[]</p></td>
<td><p>Column indices to exclude from conversion</p></td>
</tr>
<tr class="even">
<td><p><code>--newline</code></p></td>
<td></td>
<td><p>Set the newline to be used in case auto-detection should not be used. Can be one of:</p>
<ul>
<li><p>LF</p></li>
<li><p>CR</p></li>
<li><p>CRLF</p></li>
</ul></td>
</tr>
<tr class="odd">
<td><p><code>--sqlite</code></p></td>
<td><p>false</p></td>
<td><p>Output SQLite file</p></td>
</tr>
<tr class="even">
<td><p><code>--nd</code></p></td>
<td><p>false</p></td>
<td><p>Input CSV is newline delimited and does not have newlines in the middle of a field</p></td>
</tr>
<tr class="odd">
<td><p><code>--break-count</code></p></td>
<td><p>15000</p></td>
<td><p>Causes the sqlite writer to stop writing at the specified amount of rows. Can cause the merge to be triggered earlier. The first writer stops at N/3</p></td>
</tr>
<tr class="even">
<td><p><code>--commit-rate</code><br />
<code>-r</code></p></td>
<td><p>5000</p></td>
<td><p>Flush rate for comitting rows to disk</p></td>
</tr>
<tr class="odd">
<td><p><code>--buffer-size</code><br />
<code>-b</code></p></td>
<td><p>4096</p></td>
<td><p>Buffer size for reading from disk in kilobytes</p></td>
</tr>
<tr class="even">
<td><p><code>--threads</code></p></td>
<td><p>8</p></td>
<td><p>Threads to use when writing output, default is set to the amount of cores</p></td>
</tr>
<tr class="odd">
<td><p><code>--channel-capacity</code></p></td>
<td><p>8000</p></td>
<td><p>Buffered channels capacity multiplied by threads</p></td>
</tr>
<tr class="even">
<td><p><code>--chunk-size</code></p></td>
<td><p>10</p></td>
<td><p>Chunksize before pushing the data to the channel</p></td>
</tr>
<tr class="odd">
<td><p><code>--cpu-profile</code></p></td>
<td></td>
<td><p>Filename for a cpu-profile ouput for pprof</p></td>
</tr>
<tr class="even">
<td><p><code>--heap-profile</code></p></td>
<td></td>
<td><p>Filename for a heap-profile ouput for pprof</p></td>
</tr>
<tr class="odd">
<td><p><code>--trace-profile</code></p></td>
<td></td>
<td><p>Filename for a trace profile for go tool trace</p></td>
</tr>
<tr class="even">
<td><p><code>--log-header</code></p></td>
<td><p>0</p></td>
<td><p>Logs the header for the collected data, pass 1 to output it. Can be used together with seq</p></td>
</tr>
<tr class="odd">
<td><p><code>--drop</code></p></td>
<td><p>false</p></td>
<td><p>Drops all outgoing data, for perf testing disk reading</p></td>
</tr>
<tr class="even">
<td><p><code>--rowid</code></p></td>
<td><p>false</p></td>
<td><p>When the order of the CSV file must be kept the same in the resulting SQLite, this flag will cause each row in the SQLite to receive a RowID for SQLite.</p></td>
</tr>
<tr class="odd">
<td><p><code>--progress</code></p></td>
<td><p>false</p></td>
<td><p>Shows progress</p></td>
</tr>
<tr class="even">
<td><p><code>--scratch-dir</code></p></td>
<td></td>
<td><p>Directory for storing the smaller SQLite files before merging them to the target. Same as output if left blank.</p></td>
</tr>
<tr class="odd">
<td><p><code>--debug</code></p></td>
<td><p>false</p></td>
<td><p>Outputs debug information, careful is extremely chatty.</p></td>
</tr>
<tr class="even">
<td><p><code>--log-csv-reader</code></p></td>
<td><p>false</p></td>
<td><p>Logs the occurence and data where the Go built-in CSV Reader is used. Can be used without the debug mode.</p></td>
</tr>
<tr class="odd">
<td><p><code>--overflow</code></p></td>
<td><p>false</p></td>
<td><p>Overflows the data. Will abort processing.</p></td>
</tr>
</tbody>
</table>
