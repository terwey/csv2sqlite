#!/bin/bash

input=$1
output=$2

if [ -z "$output" ]; then
    output="${input}.sql"
fi

inputColumns=`sqlite3 ${input} 'pragma table_info(data)' '.exit' | sort -k2 -t\| | awk -F '|' '{print "\"" $2 "\""}'| awk -F '|' '{printf (NR>1?",":"") $1}'`
inputQuery="select ${inputColumns} from data order by ${inputColumns};"

echo "Dumping data from ${input} to ${output}"

sqlite3 $input << END_COMMAND
.headers off
.mode insert
.output ${output}
${inputQuery}
.exit
END_COMMAND
