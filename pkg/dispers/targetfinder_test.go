package enclave

import (
	"encoding/json"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshalNil(t *testing.T) {

	targetProfile := dispers.OperationTree{}

	encrypted, err := json.Marshal(targetProfile)
	assert.NoError(t, err)

	var decrypted dispers.OperationTree
	err = json.Unmarshal(encrypted, &decrypted)
	assert.NoError(t, err)

}

func TestMarshalUnmarshalOperationTree(t *testing.T) {

	targetProfile := dispers.OperationTree{
		Type: dispers.UnionNode,
		LeftNode: dispers.OperationTree{
			Type:      dispers.IntersectionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test1"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test2"},
		},
		RightNode: dispers.OperationTree{
			Type:      dispers.IntersectionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test3"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test4"},
		},
	}

	encrypted, err := json.Marshal(targetProfile)
	assert.Equal(t, "{\"type\":1,\"left_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test1\"},\"right_node\":{\"type\":0,\"value\":\"test2\"}},\"right_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test3\"},\"right_node\":{\"type\":0,\"value\":\"test4\"}}}", string(encrypted))
	assert.NoError(t, err)

	var decrypted dispers.OperationTree
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

	targetProfile := dispers.OperationTree{
		Type: dispers.UnionNode,
		LeftNode: dispers.OperationTree{
			Type:      dispers.IntersectionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test1"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test2"},
		},
		RightNode: dispers.OperationTree{
			Type:      dispers.IntersectionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test3"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test4"},
		},
	}

	in := dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	out, err := SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, out, []string{"claire", "françois", "paul"})

	targetProfile = dispers.OperationTree{
		Type: dispers.IntersectionNode,
		LeftNode: dispers.OperationTree{
			Type:      dispers.UnionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test1"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test2"},
		},
		RightNode: dispers.OperationTree{
			Type:      dispers.UnionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test3"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test7"},
		},
	}

	in = dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	_, err = SelectAddresses(in)
	assert.Error(t, err)

	// Union between Single and Intersection
	targetProfile = dispers.OperationTree{
		Type: dispers.IntersectionNode,
		LeftNode: dispers.OperationTree{
			Type:      dispers.UnionNode,
			LeftNode:  dispers.OperationTree{Type: dispers.SingleNode, Value: "test1"},
			RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test2"},
		},
		RightNode: dispers.OperationTree{Type: dispers.SingleNode, Value: "test4"},
	}

	in = dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	out, err = SelectAddresses(in)
	assert.NoError(t, err)
	assert.Equal(t, out, []string{"paul"})

	// No type precised
	targetProfile = dispers.OperationTree{
		LeftNode: dispers.OperationTree{
			LeftNode:  dispers.OperationTree{Value: "test1"},
			RightNode: dispers.OperationTree{Value: "test2"},
		},
		RightNode: dispers.OperationTree{Value: "test4"},
	}

	in = dispers.InputTF{
		IsEncrypted:      false,
		ListsOfAddresses: m,
		TargetProfile:    targetProfile,
	}

	_, err = SelectAddresses(in)
	assert.Error(t, err)
}
