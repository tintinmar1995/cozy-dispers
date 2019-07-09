
# Cozy-DISPERS

## A network of stacks and servers

- [Cozy-DISPERS Test pack](https://github.com/tintinmar1995/cozy-dispers-test-pack)
- [OAuth & Request in Cozy](stack.md)
- [Encrypted Communications between actors (TODO)]()

## Computing queries

![][https://github.com/tintinmar1995/cozy-stack/docs/pictures/cozy-dispers-extended-schema.png]

- `/conductor` - [Management of the request](conductor.md)
- `/conceptindexor` - [Select requested users Pt.1](concept-indexor.md)
- `/targetfinder` - [Select requested users Pt.2](target-finder.md)
- `/target` - [Request the selected users](target.md)
- `/dataaggregator` - [Aggregate Data](data-aggregator.md)

## Keep a look on the query

Metadata provides a way to follow the query from end to end. Metadata are created by every actor of the request (CI, TF, T, DA, Stack, ...) and are given to the conductor. Thanks to a method, the conductor can easily saved the metadata in its database.

- Host / Actor
- Name / Description
- Begining / ending time and duration

Metadata has got one function and 2 methods :

- `NewMetadata` to instanciate a new metadata
- `.Close()` to close the metadata
- `.Push()` to save it in Conductor's database

Metadata are saved in a dedicated database with the doctype io.cozy.metadata. This database contains one document for each query. Metadata can be retrieved with this request :

```http
GET /conductor/query HTTP/1.1
Host: cozy.example.org
Content-Type: application/json
```
