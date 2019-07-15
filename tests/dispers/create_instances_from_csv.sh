#!/bin/bash

# check the parameters (nb Cozy (re)ini)
if [ $# -ne 3 ]; then
    echo "Usage : "$0" nb_Cozy reini csv_file"
    exit 1
fi

NB_COZY=$1

for ((i=1 ; $NB_COZY - $i + 1 ; i++)) do
	# domain name, row in database, creation of json doc
	INST=$(./data/get_cel.sh $3 ${i} 0)
	NAME=$(./data/get_cel.sh $3 ${i} 1)
	MAIL=$(./data/get_cel.sh $3 ${i} 2)
	echo "---------$INST---------"

	if [ "$2" = "true" ];
	then
		# We delete instance if existant
		DB_PREFIX=$(cozy-stack instances show-db-prefix $INST)
		if [ "$DB_PREFIX" = "" ];
		then
			./create_instance.sh $INST $NAME $MAIL
		else
			echo "Delete instance..."
			cozy-stack instances destroy $INST --force
			./create_instance.sh $INST $NAME $MAIL
		fi
	elif [ "$2" = "false" ]
	then
		# We only delete databases if instance existant
		echo "Test instance's existance..."
		DB_PREFIX=$(cozy-stack instances show-db-prefix $INST)
		if [ "$DB_PREFIX" = "" ];
		then
			echo "Instance unexistant. Create instance..."
			./create_instance.sh $INST $NAME $MAIL
		else
			echo "Instance existant. Delete databases..."
			# Deleting io.cozy.ml / io.cozy.iris
			curl -X DELETE "http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-ml"
			curl -X DELETE "http://127.0.0.1:5984/$DB_PREFIX%2Fio-cozy-iris"
		fi
	else
		echo "ERROR : Unknown parameter. 'true' or 'false' expected." 
		exit 1
	fi

	done

exit 0
