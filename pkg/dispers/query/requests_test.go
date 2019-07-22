package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshalOperationTree(t *testing.T) {

	targetProfile := OperationTree{
		Type: UnionNode,
		LeftNode: OperationTree{
			Type:      IntersectionNode,
			LeftNode:  OperationTree{Type: SingleNode, Value: "test1"},
			RightNode: OperationTree{Type: SingleNode, Value: "test2"},
		},
		RightNode: OperationTree{
			Type:      IntersectionNode,
			LeftNode:  OperationTree{Type: SingleNode, Value: "test3"},
			RightNode: OperationTree{Type: SingleNode, Value: "test4"},
		},
	}

	encrypted, err := json.Marshal(targetProfile)
	assert.Equal(t, "{\"type\":1,\"left_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test1\"},\"right_node\":{\"type\":0,\"value\":\"test2\"}},\"right_node\":{\"type\":2,\"left_node\":{\"type\":0,\"value\":\"test3\"},\"right_node\":{\"type\":0,\"value\":\"test4\"}}}", string(encrypted))
	assert.NoError(t, err)

	var objMap OperationTree
	err = json.Unmarshal(encrypted, &objMap)
	assert.NoError(t, err)
	assert.Equal(t, targetProfile, objMap)

}

func TestUnion(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{"abc", "bcd", "efc"}
	m["test2"] = []string{"hihi"}

	op := OperationTree{
		Type:      UnionNode,
		LeftNode:  OperationTree{Type: SingleNode, Value: "test1"},
		RightNode: OperationTree{Type: SingleNode, Value: "test2"},
	}

	res, err := op.Compute(m)
	assert.NoError(t, err)
	assert.Equal(t, res, []string{"abc", "bcd", "efc", "hihi"})

}

func TestIntersection(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{"joel", "claire", "caroline", "françois"}
	m["test2"] = []string{"paul", "claire", "françois"}

	op := OperationTree{
		Type:      IntersectionNode,
		LeftNode:  OperationTree{Type: SingleNode, Value: "test1"},
		RightNode: OperationTree{Type: SingleNode, Value: "test2"},
	}

	res, err := op.Compute(m)
	assert.NoError(t, err)
	assert.Equal(t, res, []string{"claire", "françois"})

}

func TestIntersectionAndUnion(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{"joel", "claire", "caroline", "françois"}
	m["test2"] = []string{"paul", "claire", "françois"}
	m["test3"] = []string{"paul", "claire", "françois"}
	m["test4"] = []string{"paul", "benjamin", "florent"}

	op := OperationTree{
		Type: UnionNode,
		LeftNode: OperationTree{
			Type:      IntersectionNode,
			LeftNode:  OperationTree{Type: SingleNode, Value: "test1"},
			RightNode: OperationTree{Type: SingleNode, Value: "test2"},
		},
		RightNode: OperationTree{
			Type:      IntersectionNode,
			LeftNode:  OperationTree{Type: SingleNode, Value: "test3"},
			RightNode: OperationTree{Type: SingleNode, Value: "test4"},
		},
	}

	res, err := op.Compute(m)
	assert.NoError(t, err)
	assert.Equal(t, res, []string{"claire", "françois", "paul"})

}

func TestBlankLeaf(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{""}
	m["test2"] = []string{"paul", "claire", "françois"}

	op := OperationTree{
		Type:      IntersectionNode,
		LeftNode:  OperationTree{Type: SingleNode, Value: "test1"},
		RightNode: OperationTree{Type: SingleNode, Value: "test2"},
	}

	_, err := op.Compute(m)
	assert.NoError(t, err)

}

func TestUnknownConcept(t *testing.T) {

	m := make(map[string][]string)
	m["test1"] = []string{""}
	m["test2"] = []string{"paul", "claire", "françois"}

	op := OperationTree{
		Type:      IntersectionNode,
		LeftNode:  OperationTree{Type: SingleNode, Value: "test3"},
		RightNode: OperationTree{Type: SingleNode, Value: "test2"},
	}

	_, err := op.Compute(m)
	assert.Error(t, err)

}
