#!/bin/bash

# check the parameters (nb Cozy (re)ini)
if [ $# -ne 2 ]; then
    echo "Usage : "$0" csv_file"
    exit 1
fi

NB_COZY=$1

for ((i=1 ; $NB_COZY - $i + 1 ; i++)) do
	INST=$(./data/get_cel.sh $2 $i 0)
	echo "---------$INST---------"
	DOCTYPES=$(./data/get_cel.sh $2 $i 3)
	echo "Generate token..."
	TOKEN=$(./generate_token.sh $INST $DOCTYPES)
	echo "Subscribe to Cozy-DISPERS..."
	CONCEPTS=$(./data/get_cel.sh $2 $i 4)	
	CONCEPTS=$(sed 's/\// /g' <<< $CONCEPTS)
	./dispers_subscribe.sh $INST $TOKEN $CONCEPTS
	done
exit 0
