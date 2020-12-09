#!/bin/bash

input=$1


echo "Dumping ${input}"

inputColumns=`sqlite3 ${input} 'pragma table_info(data)' '.exit' | sort -k2 -t\| | awk -F '|' '{printf (NR>1?",":"") $2}'`

inputQuery="select ${inputColumns} from data order by ${inputColumns};"

inputScratchfile="${input}.sql"

echo "Dumping data from ${input} to ${inputScratchfile}"

sqlite3 $input << END_COMMAND
.headers off
.mode insert
.output ${inputScratchfile}
${inputQuery}
.exit
END_COMMAND
