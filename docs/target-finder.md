# Cozy-DISPERS : Target Finder

Target Finder is used to get a list of users to request from a target profile and several lists of users.

1. How-To
2. Operation
3. Functions and tests

## How-To Apply the target profile

**Step 1** : Create the target profile :

```golang
targetProfile := dispers.OperationTree{
		Type: dispers.UnionNode,
		LeftNode: dispers.OperationTree{
			Type:   dispers.IntersectionNode,
			LeftNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test1"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test2"},
		},
		RightNode: dispers.OperationTree{
			Type:   dispers.IntersectionNode,
			LeftNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test3"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test4"},
		},
}
```

where :
- test1 / test2 / test3 / test4 are key to reach lists
- 3 types are understood : *single*, *intersection*, *union*

**Step 2** : Create the lists of instances

```golang
m := make(map[string][]string)
m["test1"] = []string{"joel", "claire", "caroline", "françois"}
m["test2"] = []string{"paul", "claire", "françois"}
m["test3"] = []string{"paul", "claire", "françois"}
m["test4"] = []string{"paul", "benjamin", "florent"}
```

**Step 3** : Create the input structure and make the request

Here we choose not to encrypt the inputs. Inputs will be encrypted in a proper Cozy-DISPERS request.

```golang
in := dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
}
```

```http
POST query/targetfinder/addresses HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputTF
```

**Step 4** : get the result

The result will be of type :

```golang
type OutputTF struct {
	Targets  	 			 []string  `json:"addresses,omitempty"`
	EncryptedTargets []byte    `json:"enc_addresses,omitempty"`
}
```

## The structure Operation

Applying target profile requires a special structure. It could even have been a Interface to easily add an operation.

```golang
type OperationTree struct {
	Type   		NodeType      `json:"type,omitempty"`
	Value  		string      `json:"value,omitempty"`
	LeftNode 	interface{} `json:"left_node,omitempty"`
	RightNode interface{} `json:"right_node,omitempty"`
}

func (o *Operation) Compute(list map[string][]string) ([]string, error){}
func (o *Operation) UnmarshalJSON(data []byte) error {}
```

## Functions and tests

The structure Operation uses `union` and `intersection`. Two functions that given two lists of strings return a list of string.

What is tested ?
- Test marshal and unmarshal Target Profile
- Test marshal/unmarshal nil Target Profile
- Test union and intersection
- Test blank leaf (empty list)
- Test Error for a unknown concept
