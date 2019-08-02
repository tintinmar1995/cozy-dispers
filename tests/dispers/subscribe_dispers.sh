#!/bin/bash

# check the parameters (nb Cozy (re)ini)
if [ $# -lt 3 ]; then
    echo "Usage :" $0 '$INST $TOKEN $CONCEPTS'
    exit 1
fi

INST=$1
TOKEN=$2
CONCEPTS=$(sed 's/\r/ /g' <<< "${@:3}")
IFS=' ' read -ra CONCEPTS <<< "$CONCEPTS"

DATA='{"concepts":['

for (( c=0; c<${#CONCEPTS[@]}; c++ ))
do
    # Generate the JSON document and save it into array
    DATA+='"'${CONCEPTS[$c]}'"'
    # Add a comma except if it's the last run
    if [ $c -ne $((${#CONCEPTS[@]}-1)) ]; then
    DATA+=","
    fi
done

DATA+='],"is_encrypted":false,"instance":{"domain":"'$INST'","bearer" :"'$TOKEN'"}}'

curl --request POST \
  --url http://localhost:8008/subscribe/conductor/subscribe \
  --header 'content-type: application/json' \
  --data "$DATA"

exit 0
