# Configure quickly instances to test Cozy-DISPERS

## Requirement

Those script have been created for this config :

- 1 Cozy Stack listening to ports 8080 and 6060
- 1 Cozy DISPERS server listening to ports 8008 and 6061
- 1 CouchDB instance listening to port 5984
- The script also need a [config file](./data/instances.csv). This file sums up information for several instances.

```bash
cozy-dispers serve --port 8008
cozy-stack serve --port 8080
```

## Usage

```bash
NB_COZY=4
DELETING_INSTANCE="false"
PATH_CSV_FILE="./data/instances.csv"
./create_instances_from_csv.sh $NB_COZY $DELETING_INSTANCE $PATH_CSV_FILE
./add_data_from_csv.sh $NB_COZY $PATH_CSV_FILE
./subscribe_dispers_from_csv.sh $NB_COZY $PATH_CSV_FILE
```

/!\ Initialization is faster using ```DELETING_INSTANCE="false"``` because the script does not delete existance instance but simply erase databases.

## Scripts
- **add_data_from_csv.sh** : compute ACH cmd to import data from io.cozy.banks.operation
- **create_instance.sh** : create an instance on the stack
- **create_instances_from_csv.sh** : use create_instance.sh to create several instances reading instances.csv
- **generate_token.sh** : generate a token for an instance
- **subscribe_dispers_from_csv.sh** : use subscribe_dispers.sh to make several instances subscribing to Cozy-DISPERS
- **subscribe_dispers.sh** : make an instance subscribe to Cozy-DISPERS
- **get_cel.sh** is used to read one cel in a CSV file
