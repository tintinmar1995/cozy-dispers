# Cozy-DISPERS : Target Finder

Target Finder is used to get a list of users to request from a target profile and several lists of users. 

1. How-To
2. Operation 
3. Functions and tests

## How-To Apply the target profile 

** Step 1 ** : Create the target profile :

```golang
targetProfile := dispers.Operation{
		Type: "union",
		ValueA: dispers.Operation{
			Type:   "intersection",
			ValueA: dispers.Operation{Type: "single", Value: "test1"},
			ValueB: dispers.Operation{Type: "single", Value: "test2"},
		},
		ValueB: dispers.Operation{
			Type:   "intersection",
			ValueA: dispers.Operation{Type: "single", Value: "test3"},
			ValueB: dispers.Operation{Type: "single", Value: "test4"},
		},
}
```

where : 
- test1 / test2 / test3 / test4 are key to reach lists
- 4 types are understood : *single*, *intersection*, *union* and *nil* 

** Step 2 ** : Create the lists of instances

```golang
m := make(map[string][]string)
m["test1"] = []string{"joel", "claire", "caroline", "françois"}
m["test2"] = []string{"paul", "claire", "françois"}
m["test3"] = []string{"paul", "claire", "françois"}
m["test4"] = []string{"paul", "benjamin", "florent"}
```

** Step 3** : Create the input structure and make the request

Here we choose not to encrypt the inputs. Inputs will be encrypted in a proper Cozy-DISPERS request. 

```golang
in := dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
}
```

```http
POST /targetfinder/addresses HTTP/1.1
Host: cozy.example.org
Content-Type: application/json

inputTF
```

** Step 4** : get the result

The result will be of type :

```golang
type OutputTF struct {
	ListOfAddresses []string `json:"addresses,omitempty"`
}
```

## The structure Operation

Applying target profile requires a special structure. It could even have been a Interface to easily add an operation. 

```golang
type Operation struct {
	Type   string      `json:"type,omitempty"`
	Value  string      `json:"value,omitempty"`
	ValueA interface{} `json:"value_a,omitempty"`
	ValueB interface{} `json:"value_b,omitempty"`
}

func (o *Operation) Compute(list map[string][]string) ([]string, error)
func (o *Operation) MarshalJSON() ([]byte, error)
func (o *Operation) UnmarshalJSON(data []byte) error

```

## Functions and tests

The structure Operation uses `union` and `intersection`. Two functions that given two lists of strings return a list of string. 

What is tested ? 
- Test marshal and unmarshal Target Profile
- Test marshal/unmarshal nil Target Profile
- Test union and intersection
- Test blank leaf (empty list)
- Test Error for a unknown concept
