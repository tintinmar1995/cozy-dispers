package enclave

import (
	"encoding/json"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/stretchr/testify/assert"
)

func TestDeveloppTargetProfile(t *testing.T) {

	tp, err := developpTargetProfile("OR(AND(\"asse\"::\"losc\"):AND(\"psg\"::\"srfc\"))")
	assert.NoError(t, err)
	assert.Equal(t, "{\"type\": 1,\"left_node\": {\"type\": 2,\"left_node\": {\"type\": 0,\"value\": \"asse\"},\"right_node\": {\"type\": 0,\"value\": \"losc\"}},\"right_node\": {\"type\": 2,\"left_node\": {\"type\": 0,\"value\": \"psg\"},\"right_node\": {\"type\": 0,\"value\": \"srfc\"}}}", tp)

	tp, err = developpTargetProfile("AND(\"asse\"::\"losc\")")
	assert.NoError(t, err)
	assert.Equal(t, "{\"type\": 2,\"left_node\": {\"type\": 0,\"value\": \"asse\"},\"right_node\": {\"type\": 0,\"value\": \"losc\"}}", tp)

}

func TestMarshalUnmarshalNil(t *testing.T) {

	targetProfile := query.OperationTree{}

	encrypted, err := json.Marshal(targetProfile)
	assert.NoError(t, err)

	var decrypted query.OperationTree
	err = json.Unmarshal(encrypted, &decrypted)
	assert.NoError(t, err)

}

func TestMarshalUnmarshalOperationTree(t *testing.T) {

	targetProfile := query.OperationTree{
		Type: query.OrNode,
		LeftNode: query.OperationTree{
			Type:      query.AndNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test1"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.AndNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test3"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test4"},
		},
	}

	encrypted, err := json.Marshal(targetProfile)
	assert.Equal(t, "{\"type\":1,\"left_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test1\"},\"right_node\":{\"type\":0,\"value\":\"test2\"}},\"right_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test3\"},\"right_node\":{\"type\":0,\"value\":\"test4\"}}}", string(encrypted))
	assert.NoError(t, err)

	var decrypted query.OperationTree
	err = json.Unmarshal(encrypted, &decrypted)
	assert.NoError(t, err)
	assert.Equal(t, targetProfile, decrypted)

}

func TestTargetFinder(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{"joel", "claire", "caroline", "françois"}
	m["test2"] = []string{"paul", "claire", "françois"}
	m["test3"] = []string{"paul", "claire", "françois"}
	m["test4"] = []string{"paul", "benjamin", "florent"}

	in := query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    "OR(AND(\"test1\"::\"test2\"):AND(\"test3\"::\"test4\"))",
	}

	out, err := SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, out, []string{"claire", "françois", "paul"})

	in = query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    "AND(OR(\"test1\"::\"test2\"):OR(\"test3\"::\"test7\"))",
	}

	_, err = SelectAddresses(in)
	assert.Error(t, err)

	in = query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    "AND(OR(\"test1\"::\"test2\"):AND(\"test4\"::\"test4\"))",
	}

	out, err = SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, []string{"paul"}, out)

}
