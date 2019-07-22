package enclave

import (
	"encoding/json"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/stretchr/testify/assert"
)

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
		Type: query.UnionNode,
		LeftNode: query.OperationTree{
			Type:      query.IntersectionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test1"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.IntersectionNode,
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

	targetProfile := query.OperationTree{
		Type: query.UnionNode,
		LeftNode: query.OperationTree{
			Type:      query.IntersectionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test1"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.IntersectionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test3"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test4"},
		},
	}

	in := query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	out, err := SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, out, []string{"claire", "françois", "paul"})

	targetProfile = query.OperationTree{
		Type: query.IntersectionNode,
		LeftNode: query.OperationTree{
			Type:      query.UnionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test1"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.UnionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test3"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test7"},
		},
	}

	in = query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	_, err = SelectAddresses(in)
	assert.Error(t, err)

	// Union between Single and Intersection
	targetProfile = query.OperationTree{
		Type: query.IntersectionNode,
		LeftNode: query.OperationTree{
			Type:      query.UnionNode,
			LeftNode:  query.OperationTree{Type: query.SingleNode, Value: "test1"},
			RightNode: query.OperationTree{Type: query.SingleNode, Value: "test2"},
		},
		RightNode: query.OperationTree{Type: query.SingleNode, Value: "test4"},
	}

	in = query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	out, err = SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, out, []string{"paul"})

	// No type precised
	targetProfile = query.OperationTree{
		LeftNode: query.OperationTree{
			LeftNode:  query.OperationTree{Value: "test1"},
			RightNode: query.OperationTree{Value: "test2"},
		},
		RightNode: query.OperationTree{Value: "test4"},
	}

	in = query.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	_, err = SelectAddresses(in)
	assert.Error(t, err)
}
