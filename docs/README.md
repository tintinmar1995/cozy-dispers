
# Cozy-DISPERS

## Computing queries

![](pictures/cozy-dispers-extended-schema.png?raw=true)

- `/query/` - [Management of the request](conductor.md)
- `/query/conceptindexor/` - [Select targeted users Pt.1](concept-indexor.md)
- `/query/targetfinder/` - [Select targeted users Pt.2](target-finder.md)
- `/query/target/` - [Request the selected users](target.md)
- `/query/dataaggregator/` - [Aggregate Data](data-aggregator.md)
- `/subscribe/` - [Subscription routes]()
- [Cozy-Application to control Cozy-DISPERS queries]()
- [ExecutionMetadata and TaskMetadata]()

## Quick start to build a network adapted for Cozy-DISPERS

- Build/Download a binary of Cozy-Stack
- Build/Download a binary of Cozy-DISPERS
- Install CouchDB (e.g. with Docker)
- Get some data (e.g. ACH export your own Cozy)
- Create files instances.csv & concepts.csv (cf. tests/dispers/)
- Fill the first variables of `magic-script.sh`
- Launch `bash magic-script.sh INIT` in your unix shell

`bash magic-script.sh INIT` will launch CouchDB/Cozy-Stack/Cozy-DISPERS, initiate instances and make them subscribe to Cozy-DISPERS.

If you already had initialize a proper environment for a demo of Cozy-DISPERS, you can use `magic-script.sh` as follow :

```bash
# Start CouchDB/Cozy-Stack/Cozy-DISPERS and update tokens
bash magic-script.sh START
# Start CouchDB/Cozy-Stack/Cozy-DISPERS, start cozy-dispers-app and update tokens (require that cozy-dispers-app has been installed)
bash magic-script.sh DEV_MODE
# Update tokens (require that CouchDB/Cozy-Stack/Cozy-DISPERS are running)
bash magic-script.sh UPDATE_TOKENS
```
