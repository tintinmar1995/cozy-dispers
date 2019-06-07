```golang
{
	"domain" : "cozy2test.mycozy.cloud",
	"concept" : [
		{
			"encrypted" : false,
			"concept" : "LikeBeer"
		},
		{
 			"encrypted" : false,
			"concept" : "PlayBadminton"
		},
		{
			"encrypted" : false,
			"concept" : "LiveInStrasbourg"			
		}
	],
	"pseudo_concepts" : {
		"LikeBeer" : "001",
		"PlayBadminton" : "051",
		"LiveInStrasbourg" : "519"
	},
	"encrypted" : false,
	"job" : [
		{
			"func" : "sum",
			"args" : {
				"keys" : ["X1", "X2"]
			}
		},
		{
			"func" : "sum",
			"args" : {
				"keys" : ["X1", "X2"],
				"weight" : "LengthDataset"
			}
		}
	],
	"localquery" : {
		"findrequest" : {
			"UseIndex" : "my-index",
			"Selector" : "todo"
		}
	},
	"operation" : {
		"type" : "union",
		"value_a" : {
			"type" : "intersection",
			"value_a" : {
				"type" : "single",
				"value" : "LikeBeer"
			}, 
			"value_b" : {
				"type" : "single",
				"value" : "PlayBadminton"
			}
		}, 
		"value_b" : {
			"type" : "intersection",
			"value_a" : {
				"type" : "single",
				"value" : "LikeBeer"
			}, 
			"value_b" : {
				"type" : "single",
				"value" : "LiveInStrasbourg"
			}
		}
	},
	"nb_actors" : {"ci": 1, "tf": 1, "t": 1},
	"size_aggr" : [ 5, 3]
}
```
