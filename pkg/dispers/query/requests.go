package query

import (
	"encoding/json"
	"errors"
	"net/url"
)

/*
*
Conductor's Input & Output
*
*/

type InputPatchQuery struct {
	IsEncrypted bool     `json:"encrypted"`
	Role        string   `json:"role"`
	OutDA       OutputDA `json:"output_da,omitempty"`
}

/*
*
Concept Indexors' Input & Output
*
*/

type Concept struct {
	IsEncrypted      bool   `json:"encrypted,omitempty"`
	Concept          string `json:"concept,omitempty"`
	EncryptedConcept []byte `json:"enc_concept,omitempty"`
	Hash             []byte `json:"hash,omitempty"`
}

type InputCI struct {
	Concepts []Concept `json:"concepts,omitempty"`
}

// OutputCI contains a bool and the result
type OutputCI struct {
	Hashes []Concept `json:"hashes,omitempty"`
}

func ConceptsToString(concepts []Concept) string {
	str := ""
	for index, concept := range concepts {
		// Stack every concept, with ":" as separator
		// TODO : Delete concept.Concept
		if concept.IsEncrypted {
			str = str + string(concept.EncryptedConcept)
		} else {
			str = str + concept.Concept
		}
		if index != (len(concepts) - 1) {
			str = str + ":"
		}
	}

	return str
}

/*
*
Target Finders' Input & Output
*
*/

func union(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

func intersection(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

// NodeType are the only possible nodes in Target Profile trees
type NodeType int

const (
	// SingleNode are Target Profile's leafs
	SingleNode NodeType = iota
	// OrNode are unions between two lists
	OrNode
	// AndNode are intersections between two lists
	AndNode
)

// OperationTree allows the possibility to compute target profiles in a
// recursive way. OperationTree contains SingleNode, OrNode, AndNode
// SingleNodes have got a value field. A value is the name of a list of strings
// To compute the OperationTree, Compute method needs a map that matches names
// with list of encrypted addresses.
type OperationTree struct {
	Type      NodeType    `json:"type"`
	Value     string      `json:"value,omitempty"`
	LeftNode  interface{} `json:"left_node,omitempty"`
	RightNode interface{} `json:"right_node,omitempty"`
}

// Compute compute the OperationTree and returns the list of encrypted addresses
func (o *OperationTree) Compute(listsOfAddresses map[string][]string) ([]string, error) {

	if o.Type == SingleNode {
		// Retrieve list of addresses from listsOfAddresses
		val, ok := listsOfAddresses[o.Value]
		if !ok {
			msg := "Unknown concept : " + o.Value + " expect one of : "
			for k := range listsOfAddresses {
				msg = msg + " " + k
			}
			return []string{}, errors.New(msg)
		}
		return val, nil

	} else if o.Type == OrNode || o.Type == AndNode {

		// Compute operations on LeftNode and RightNode
		leftNode := o.LeftNode.(OperationTree)
		a, err := leftNode.Compute(listsOfAddresses)
		if err != nil {
			return []string{}, err
		}
		rightNode := o.RightNode.(OperationTree)
		b, err := rightNode.Compute(listsOfAddresses)
		if err != nil {
			return []string{}, err
		}
		// Compute operation between LeftNode and RightNode
		switch o.Type {
		case OrNode:
			return union(a, b), nil
		case AndNode:
			return intersection(a, b), nil
		default:
			return []string{}, errors.New("Unknown type")
		}
	} else {
		return []string{}, errors.New("Unknown type")
	}
}

// UnmarshalJSON is used to load the OperationTree given by the Querier
func (o *OperationTree) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	// Retrieve the NodeType
	if v["type"] == nil {
		return errors.New("No type defined")
	}
	switch int(v["type"].(float64)) {
	case 0:
		o.Type = SingleNode
	case 1:
		o.Type = OrNode
	case 2:
		o.Type = AndNode
	default:
		return errors.New("Unknown type")
	}

	// Retrieve others attributes depending on NodeType
	switch {
	case o.Type == SingleNode:
		o.Value, _ = v["value"].(string)
	case o.Type == AndNode || o.Type == OrNode:
		var leftNode OperationTree
		var rightNode OperationTree

		leftNodeByte, _ := json.Marshal(v["left_node"].(map[string]interface{}))
		err := json.Unmarshal(leftNodeByte, &leftNode)
		if err != nil {
			return err
		}

		rightNodeByte, _ := json.Marshal(v["right_node"].(map[string]interface{}))
		err = json.Unmarshal(rightNodeByte, &rightNode)
		if err != nil {
			return err
		}
		o.LeftNode = leftNode
		o.RightNode = rightNode
	default:
	}
	return nil
}

// InputTF contains a map that associate every concept to a list of Addresses
// and a operation to compute to retrive the final list
type InputTF struct {
	IsEncrypted               bool                `json:"isencrypted"`
	EncryptedListsOfAddresses []byte              `json:"enc_instances,omitempty"`
	EncryptedTargetProfile    []byte              `json:"enc_operation,omitempty"`
	ListsOfAddresses          map[string][]string `json:"instances"`
	TargetProfile             string              `json:"target_profile,omitempty"`
}

// OutputTF is what Target Finder send to the conductor
type OutputTF struct {
	ListOfAddresses          []string `json:"addresses,omitempty"`
	EncryptedListOfAddresses []byte   `json:"enc_addresses,omitempty"`
}

// Instance describes the location of an instance and the token it had created
// When Target received twice the same Instance, it needs to be able to consider the more recent item
type Instance struct {
	Domain      string `json:"domain"`
	TokenBearer string `json:"bearer,omitempty"`
	Version     int    `json:"version"`
}

/*
*
Targets' Input & Output
*
*/

// InputT contains information received by Target's enclave
type InputT struct {
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	Targets             []string   `json:"Addresses,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTargets    []byte     `json:"enc_addresses,omitempty"`
	QueryID             string     `json:"queryid,omitempty"`
}

// StackQuery is all the information needed by the conductor's and stack to make a query
type StackQuery struct {
	Domain              string     `json:"domain,omitempty"`
	LocalQuery          LocalQuery `json:"localquery,omitempty"`
	TokenBearer         string     `json:"bearer,omitempty"`
	IsEncrypted         bool       `json:"isencrypted,omitempty"`
	EncryptedLocalQuery []byte     `json:"enc_localquery,omitempty"`
	EncryptedTokens     []byte     `json:"enc_token,omitempty"`
}

// OutputT is what Target returns to the conductor
type OutputT struct {
	Data    []map[string]interface{} `json:"data,omitempty"`
	QueryID string                   `json:"queryid,omitempty"`
}

// LocalQuery decribes which data the stack has to retrieve
type LocalQuery struct {
	FindRequest map[string]interface{} `json:"findrequest,omitempty"`
	Doctype     string                 `json:"doctype,omitempty"`
	Index       map[string]interface{} `json:"index,omitempty"`
}

/*
*
Data Aggregators' Input & Output
*
*/

// AggregationFunction is transmitted
type AggregationFunction struct {
	Function string                 `json:"func,omitempty"`
	Args     map[string]interface{} `json:"args,omitempty"`
}

type InputDA struct {
	Job           AggregationFunction      `json:"job"`
	QueryID       string                   `json:"queryid"`
	AggregationID [2]int                   `json:"aggregationid,omitempty"`
	ConductorURL  url.URL                  `json:"conductor_url"`
	Data          []map[string]interface{} `json:"data,omitempty"`
	IsEncrypted   bool                     `json:"isencrypted"`
	EncryptedJob  []byte                   `json:"enc_type,omitempty"`
	EncryptedData []byte                   `json:"enc_data,omitempty"`
}

type OutputDA struct {
	Results       map[string]interface{} `json:"results,omitempty"`
	QueryID       string                 `json:"queryid,omitempty"`
	AggregationID [2]int                 `json:"aggregationid,omitempty"`
}
