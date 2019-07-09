
# Cozy-DISPERS

## Structure

```
---- pkg
|--- dispers
||-> metadata
||-> network
||-> query
||-> subscribe
---- web
|--- query
|--- subscribe
```

## Integration in a network of stacks and servers

- [OAuth & Request](stack.md)
- [Encrypted Communication]()

## List of services

- `/conductor` - [Management of the request](conductor.md)
- `/conceptindexor` - [Select requested users Pt.1](concept-indexor.md)
- `/targetfinder` - [Select requested users Pt.2](target-finder.md)
- `/target` - [Request the selected users](target.md)
- `/dataaggregator` - [Aggregate Data](data-aggregator.md)

## Let's have a look on the query

Metadata provides a way to follow the query from end to end. Metadata are created by every actor of the request (CI, TF, T, DA, Stack, ...) and are given to the conductor. Thanks to a method, the conductor can easily saved the metadata in its database.

Metadata provides information about the request. Its needs to be used outside of the package "enclave" :

Time of treatment's begining / ending and duration
Host / Actor
Name / Description of the treatements

Metadata has got one function and 2 methods :

`NewMetadata` to instanciate a new metadata
`.Close()` to close the metadata
`.Push()` to save it in Conductor's database

Metadata are saved in a dedicated database with the doctype io.cozy.metadata. This database contains one document for each query.
