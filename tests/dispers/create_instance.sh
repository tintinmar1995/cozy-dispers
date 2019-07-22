#!/bin/bash

# check the parameters
if [ $# -lt 3 ]; then
    echo "Usage : $0 instance_name public_name mail"
    echo "Example: a.cozy.tools:8080 Alice alice@cozy.cloud"
    exit 1
fi

INST=$1
NAME=$2
MAIL=$3

echo "Create instance $INST..."
cozy-stack instances add $INST --passphrase cozy --apps drive,photos,settings,collect,contacts --email $MAIL --locale fr --public-name $2 --settings context:dev


