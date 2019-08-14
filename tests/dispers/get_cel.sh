#!/bin/bash

# check the parameters (file row col)
if [ $# -ne 3 ]; then
    echo "Usage : $0 iris.csv 5 12"
    exit 1
fi

# get the cell i,j from a table in csv with sep=","
# parameters file, row, col
line=$(sed "${2}q;d" $1)
line=$(sed 's/\r//g' <<< $line)
row=( ${line//,/ } )
echo ${row[$3]}
