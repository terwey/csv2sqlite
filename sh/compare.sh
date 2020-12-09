#!/bin/bash

input=$1
output=$2

# purely for testing the script it is commented out
if [ "$input" == "$output" ]; then
    echo "Files to compare are the same"
    exit 1
fi

echo "Comparing ${input} and ${output}"

inputColumns=`sqlite3 ${input} 'pragma table_info(data)' '.exit' | sort -k2 -t\| | awk -F '|' '{printf (NR>1?",":"") $2}'`
outputColumns=`sqlite3 ${output} 'pragma table_info(data)' '.exit' | sort -k2 -t\| | awk -F '|' '{printf (NR>1?",":"") $2}'`

if [ "$inputColumns" != "$outputColumns" ]; then
    echo "Columns do not match, ${input}:"
    echo $inputColumns
    echo -e "\n${output}:"
    echo $outputColumns
    exit 1
fi

inputQuery="select ${inputColumns} from data order by ${inputColumns};"
outputQuery="select ${outputColumns} from data order by ${inputColumns};"

inputScratchfile='/tmp/compare_input.sql'
outputScratchfile='/tmp/compare_output.sql'

echo "Dumping data from ${input} to ${inputScratchfile}"

sqlite3 $input << END_COMMAND
.mode insert
.output ${inputScratchfile}
${inputQuery}
.exit
END_COMMAND

echo "Dumping data from ${output} to ${outputScratchfile}"

sqlite3 $output << END_COMMAND
.mode insert
.output ${outputScratchfile}
${outputQuery}
.exit
END_COMMAND

# echo "Splitting ${inputScratchfile}"

# split -d -n4 ${inputScratchfile} ${inputScratchfile}.
# rm ${inputScratchfile}

# echo "Splitting ${outputScratchfile}"
# split -d -n4 ${outputScratchfile} ${outputScratchfile}.
# rm ${outputScratchfile}

inputSum=`b3sum ${inputScratchfile}`
outputSum=`b3sum ${outputScratchfile}`

if [ "${inputScratchfile}" == "${outputScratchfile}" ]; then
    echo "Files are the same"
    exit 0
fi

echo "Diffing ${inputScratchfile} with ${outputScratchfile}"

# diff ${inputScratchfile}.00 ${outputScratchfile}.00
# diff ${inputScratchfile}.01 ${outputScratchfile}.01
# diff ${inputScratchfile}.02 ${outputScratchfile}.02
# diff ${inputScratchfile}.03 ${outputScratchfile}.03

# rm ${inputScratchfile} ${outputScratchfile}

diff ${inputScratchfile} ${outputScratchfile}