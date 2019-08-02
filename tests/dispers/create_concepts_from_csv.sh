#!/bin/bash

REQ='{"concepts":['
LINE=1
CONCEPT=$(./data/get_cel.sh ./data/concepts.csv $LINE 0)
ISENC=$(./data/get_cel.sh ./data/concepts.csv $LINE 1)

while [ "$CONCEPT" != "" ]
do
  if [ $LINE -ne 1 ]; then
    REQ+=','
  fi
  REQ=$REQ'{"encrypted":'$ISENC',"concept":"'$CONCEPT'"}'
  LINE=$(($LINE+1))
  CONCEPT=$(./data/get_cel.sh ./data/concepts.csv $LINE 0)
  ISENC=$(./data/get_cel.sh ./data/concepts.csv $LINE 1)
done

REQ+=']}'

curl -d $REQ -X POST http://localhost:8008/subscribe/conductor/concept
