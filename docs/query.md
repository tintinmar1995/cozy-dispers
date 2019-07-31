# Cozy-DISPERS : Query

## Input

```json
{
  "domain": "martin.mycozy.cloud",
  "concepts": [
    {
      "concept": "travail>lille",
      "encrypted": false
    },
    {
      "concept": "travail>saint-etienne",
      "encrypted": false
    },
    {
      "concept": "travail>paris",
      "encrypted": false
    },
    {
      "concept": "travail>rennes",
      "encrypted": false
    }
  ],
  "pseudo_concepts": {
    "travail>lille": "losc",
    "travail>saint-etienne": "asse",
    "travail>paris": "psg",
    "travail>rennes": "srfc"
  },
  "encrypted": false,
  "localquery": {
    "findrequest": {
      "selector": {
        "_id": {
          "$gt": null
        }
      }
    },
    "doctype": "io.cozy.bank.operations",
		"index": {
		"fields": [
			"_id"
		]
		}
  },
  "operation": {
    "type": 1,
    "left_node": {
      "type": 1,
      "left_node": {
        "type": 0,
        "value": "asse"
      },
      "right_node": {
        "type": 0,
        "value": "losc"
      }
    },
    "right_node": {
      "type": 1,
      "left_node": {
        "type": 0,
        "value": "psg"
      },
      "right_node": {
        "type": 0,
        "value": "srfc"
      }
    }
  },
  "nb_actors": {
    "ci": 1,
    "tf": 1,
    "t": 1
  },
  "layers_da": [
    {
      "layer_job": {
        "func": "sum",
        "args": {
          "keys": [
            "amount"
          ]
        }
      },
      "layer_size": 6
    },
    {
      "layer_job": {
        "func": "sum",
        "args": {
          "keys": [
            "amount"
          ],
          "weight": "length"
        }
      },
      "layer_size": 1
    }
  ]
}
```
