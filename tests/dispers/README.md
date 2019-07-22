# Cozy-DISPERS Test pack

## Requirement

Those script have been created for this config : 

- 1 Cozy Stack listening to ports 8080 and 6060
- 1 Cozy DISPERS server listening to ports 8008 and 6061
- 1 CouchDB instance listening to port 5984

```bash
~/dev/go-workspace/bin
cozy-dispers-v190710 serve
cozy-stack serve
```

- The script also need a [config file](./data/instances.csv)

## Scripts

- **Create instances**. This script contains two modes : deleting an entire existent instance before creating or only deleting its databases (io.cozy.ml and io.cozy.banks)
- **Add data** from iris.csv and exports from instances 
- **Generate tokens** and give it the Cozy-DISPERS thanks to the subscribe route

