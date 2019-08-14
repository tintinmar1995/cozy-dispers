#!/bin/bash

# check the parameters
if [ $# -lt 3 ]; then
    echo "Usage : $0 COZY_STACK INSTANCE_NAME PUBLIC_NAME MAIL"
    echo "Example: a.cozy.tools:8080 Alice alice@cozy.cloud"
    exit 1
fi

INST=$2
NAME=$3
MAIL=$4

echo "Creating instance $INST..."
$1 instances add $INST --passphrase cozy --apps drive,photos,settings,collect,contacts --email $MAIL --locale fr --public-name $2 --settings context:dev
