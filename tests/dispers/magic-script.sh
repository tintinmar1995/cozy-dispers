#!/bin/bash

NB_INSTANCES=15
COZY_DISPERS=/home/martin/go/src/github.com/cozy/cozy-stack/main
COZY_STACK=/home/martin/dev/cozy-stack
# /!\ DO NOT END FOLDER PATH WITH "/"
PATH_DOCTYPES=/home/martin/dev/cozy-doctypes
PATH_TEST_FOLDER=/home/martin/go/src/github.com/cozy/cozy-stack/tests/dispers
PATH_APP=/home/martin/dev/cozy-dispers-app
# /!\ /!\ /!\ /!\ /!\ /!\
PATH_INSTANCES=/home/martin/go/src/github.com/cozy/cozy-stack/tests/dispers/data/instances.csv
PATH_CONCEPTS=/home/martin/go/src/github.com/cozy/cozy-stack/tests/dispers/data/concepts.csv

# check the parameters
if [ $# -ne 1 ]; then
    echo "Usage : "$0" INIT/START/DEV_APP/UPDATE_TOKENS"
    exit 1
fi

if [ "$1" != "UPDATE_TOKENS" ] && [ "$1" != "INIT" ] && [ "$1" != "START" ] && [ "$1" != "DEV_APP" ]; then
  echo "Unknown mode. Please, use INIT or START or UPDATE_TOKENS"
fi

# Start servers on ports 5984/6060/6061/8008/8080
if [ "$1" = "INIT" ] || [ "$1" = "START" ] || [ "$1" = "DEV_APP" ]; then
  # Start CouchDB
  sudo docker start cozy-stack-couch
  # Wait that CouchDB is launched
  sleep 5s
  # Start Stack and DISPERS
  gnome-terminal --tab -- bash -c "$COZY_STACK serve --doctypes $PATH_DOCTYPES; bash"
  gnome-terminal --tab -- bash -c "$COZY_DISPERS serve --port 8008 --admin-port 6061; bash"
  sleep 1s
fi

if [ "$1" = "INIT" ]; then
  # Initialize instances
  for ((i=1 ; $NB_INSTANCES - $i + 1 ; i++)) do
  	# domain name, row in database, creation of json doc
  	INST=$($PATH_TEST_FOLDER/get_cel.sh $3 ${i} 0)
  	NAME=$($PATH_TEST_FOLDER/get_cel.sh $3 ${i} 1)
  	MAIL=$($PATH_TEST_FOLDER/get_cel.sh $3 ${i} 2)

		echo "Test existance of $INST..."
		DB_PREFIX=$($COZY_STACK instances show-db-prefix $INST)
		if [ "$DB_PREFIX" = "" ];
		then
			echo "$INST unexistant. Creating instance..."
			$PATH_TEST_FOLDER/create_instance.sh $COZY_STACK $INST $NAME $MAIL
		else
			echo "$INST existant. Deleting databases..."
			curl -X DELETE "http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-bank-operations"
		fi

    echo "$INST : Importing bank operations..."  # (%2F stands for /)
  	BANK=$($PATH_TEST_FOLDER/get_cel.sh $PATH_INSTANCES $i 5)
  	TOKEN=$($PATH_TEST_FOLDER/generate_token.sh $COZY_STACK $INST "io.cozy.bank.operations:POST")
    ACH import $HOME$BANK --url http://$INST --token $TOKEN
	done

  # Create concepts
  LINE=1
  CONCEPT=$($PATH_TEST_FOLDER/get_cel.sh $PATH_CONCEPTS $LINE 0)
  ISENC=$($PATH_TEST_FOLDER/get_cel.sh $PATH_CONCEPTS $LINE 1)
  REQ='{"concepts":['

  while [ "$CONCEPT" != "" ]
  do
    if [ $LINE -ne 1 ]; then
      REQ+=','
    fi
    REQ=$REQ'"'$CONCEPT'"'
    LINE=$(($LINE+1))
    CONCEPT=$($PATH_TEST_FOLDER/get_cel.sh $PATH_CONCEPTS $LINE 0)
    ISENC=$($PATH_TEST_FOLDER/get_cel.sh $PATH_CONCEPTS $LINE 1)
  done
  REQ+='], "is_encrypted":false}'

  curl -d "$REQ" -X POST http://localhost:8008/subscribe/conductor/concept
fi

if [ "$1" = "DEV_APP" ]; then
  gnome-terminal --tab -- bash -c "cd $PATH_APP && sudo yarn watch --fix; bash"
fi

if [ "$1" = "UPDATE_TOKENS" ] || [ "$1" = "INIT" ] || [ "$1" = "START" ] || [ "$1" = "DEV_APP" ]; then
  for ((i=1 ; $NB_INSTANCES - $i + 1 ; i++)) do
  	INST=$($PATH_TEST_FOLDER/get_cel.sh $PATH_INSTANCES $i 0)
    echo " "
    echo "$INST is subscribing to Cozy-DISPERS..."
  	DOCTYPES=$($PATH_TEST_FOLDER/get_cel.sh $PATH_INSTANCES $i 3)
  	TOKEN=$($PATH_TEST_FOLDER/generate_token.sh $COZY_STACK $INST $DOCTYPES)
  	CONCEPTS=$($PATH_TEST_FOLDER/get_cel.sh $PATH_INSTANCES $i 4)
  	CONCEPTS=$(sed 's/\// /g' <<< $CONCEPTS)
  	$PATH_TEST_FOLDER/subscribe_dispers.sh $INST $TOKEN $CONCEPTS
    done
fi


echo " "
exit 0
