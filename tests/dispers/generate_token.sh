#!/bin/bash

DEFAULT_HOST="a.cozy.tools:8080"

if [ $# -lt 2 ]; then
    echo "Usage : $0 COZY_STACK HOST DOCTYPES"
    exit 1
fi

TOKEN=$($1 instances token-cli $2 ${@:3})
echo $TOKEN
