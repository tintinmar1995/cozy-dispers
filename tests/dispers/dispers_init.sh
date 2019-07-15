#!/bin/bash

# This script will initialize instances in a Cozy Stack. Those instances will have some data : one row from the iris database

# check the parameters (nb_Cozy deleteInstance? csv_file)
if [ $# -ne 3 ]; then
    echo "Usage : "$0" 150 true instance.csv"
    exit 1
fi

NB_COZY=$1

echo -e "Initiate instances..."
./create_instances_from_csv.sh $1 $2 $3
echo -e "\n\n\nImport data..."
./dispers_add_data_from_csv.sh $1 $3
echo -e "\n\n\nSubscribe to Cozy-DISPERS..."
./dispers_generate_token_from_csv.sh $1 $3

exit 0 	
