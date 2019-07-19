#!/bin/bash

# check the parameters (nb Cozy (re)ini)
if [ $# -ne 2 ]; then
    echo "Usage : "$0" csv_file"
    exit 1
fi

NB_COZY=$1

for ((i=1 ; $NB_COZY - $i + 1 ; i++)) do
	# domain name, row in database, creation of json doc
	INST=$(./data/get_cel.sh $2 $i 0)
	NAME=$(./data/get_cel.sh $2 $i 1)
	MAIL=$(./data/get_cel.sh $2 $i 2)
	echo "---------$INST---------"
	DOC=$(sed -n "$((7*$i+2)),$((7*$i+8))p" < ./data/iris.json)
	DOC=$(sed 's/},/}/g' <<< $DOC)
	BANK=$(./data/get_cel.sh $2 $i 5)

	echo "Create databases..." # (%2F stands for /)
	echo 'Importing iris...'
	DB_PREFIX=$(cozy-stack instances show-db-prefix $INST)
	curl -X PUT "http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-query"
	curl -X PUT "http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-iris"
	# We inject iris data
	curl -X POST -H 'Content-Type: application/json' http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-iris -d "$DOC"
	# We inject standard Cozy data
	echo 'Importing bank operations...'
	echo "ACH import $BANK --url http://$INST"
	TOKEN=$(./generate_token.sh $INST "io.cozy.bank.operations:POST")
	ACH import $HOME$BANK --url http://$INST --token $TOKEN
	done

