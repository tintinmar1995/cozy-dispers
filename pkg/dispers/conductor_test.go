package enclave

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/dispers/subscribe"
	"github.com/stretchr/testify/assert"
)

var (
	targetProfile = query.OperationTree{
		Type: query.OrNode,
		LeftNode: query.OperationTree{
			Type:      query.OrNode,
			LeftNode:  query.OperationTree{Type: 0, Value: "test1"},
			RightNode: query.OperationTree{Type: 0, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.OrNode,
			LeftNode:  query.OperationTree{Type: 0, Value: "test3"},
			RightNode: query.OperationTree{Type: 0, Value: "test4"},
		},
	}

	in = query.InputNewQuery{
		DomainQuerier: "usr0.test.cozy.tools:8008",
		TargetProfile: "OR(OR(\"test1\"::\"test2\"):OR(\"test3\"::\"test4\"))",
		NumberActors:  map[string]int{"ci": 1, "tf": 1, "t": 1},
	}
)

func TestCreateSubscribeDoc(t *testing.T) {

	// Create the three concepts
	inputCI := query.InputCI{
		Concepts: []query.Concept{
			query.Concept{
				IsEncrypted: false,
				Concept:     "julien"},
			query.Concept{
				IsEncrypted: false,
				Concept:     "francois"},
			query.Concept{
				IsEncrypted: false,
				Concept:     "paul"},
		},
	}

	err := CreateConceptInConductorDB(&inputCI)
	assert.NoError(t, err)
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/julien/false")
	ci.MakeRequest("GET", "", nil, nil)
	assert.NoError(t, err)
	var outputCI query.OutputCI
	err = json.Unmarshal(ci.Out, &outputCI)
	assert.NoError(t, err)

	_, err = RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
	assert.NoError(t, err)

	err = ci.MakeRequest("DELETE", "concept/julien:francois:paul/false", "application/json", nil)
	assert.NoError(t, err)
}

func TestSubscribe(t *testing.T) {

	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)

	inputCI := query.InputCI{
		Concepts: []query.Concept{
			query.Concept{
				IsEncrypted: false,
				Concept:     "aime les fraises"},
			query.Concept{
				IsEncrypted: false,
				Concept:     "aime les framboises"},
			query.Concept{
				IsEncrypted: false,
				Concept:     "joue de la guitare"},
			query.Concept{
				IsEncrypted: false,
				Concept:     "est designer chez cozy"},
		},
	}

	err := CreateConceptInConductorDB(&inputCI)
	assert.NoError(t, err)
	ci.DefineDispersActor("concept/aime les fraises/false")
	err = ci.MakeRequest("GET", "", nil, nil)
	assert.NoError(t, err)
	var outputCI query.OutputCI
	err = json.Unmarshal(ci.Out, &outputCI)
	assert.NoError(t, err)

	var listOfInstance []string
	docs, err := RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(docs))
	if len(docs) == 1 {
		json.Unmarshal(docs[0].EncryptedInstances, &listOfInstance)
		sizeIni := len(listOfInstance)

		// Make few instances subscribe to Cozy-DISPERS
		inSubs := subscribe.InputConductor{
			Concepts:          []string{"aime les fraises"},
			IsEncrypted:       false,
			EncryptedInstance: []byte("{\"domain\":\"joel.mycozy.cloud\"}"),
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
		docs, _ = RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
		json.Unmarshal(docs[0].EncryptedInstances, &listOfInstance)
		size := len(listOfInstance)
		assert.Equal(t, sizeIni+1, size)
		inSubs = subscribe.InputConductor{
			Concepts:          []string{"aime les fraises", "aime les framboises"},
			IsEncrypted:       false,
			EncryptedInstance: []byte("{\"domain\":\"paul.mycozy.cloud\"}"),
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
		docs, _ = RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
		json.Unmarshal(docs[0].EncryptedInstances, &listOfInstance)
		size = len(listOfInstance)
		assert.Equal(t, sizeIni+2, size)
		inSubs = subscribe.InputConductor{
			Concepts:          []string{"aime les fraises", "aime les framboises", "est designer chez cozy"},
			IsEncrypted:       false,
			EncryptedInstance: []byte("{\"domain\":\"francois.mycozy.cloud\"}"),
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
		docs, _ = RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
		json.Unmarshal(docs[0].EncryptedInstances, &listOfInstance)
		size = len(listOfInstance)
		assert.Equal(t, sizeIni+3, size)
	}

	ci.DefineDispersActor("concept/aime les fraises:aime les framboises:joue de la guitare:est designer chez cozy/false")
	err = ci.MakeRequest("DELETE", "", nil, nil)
	assert.NoError(t, err)

}

func TestDefineConductor(t *testing.T) {

	_, err := NewQuery(&in)
	assert.NoError(t, err)
}

func TestDecryptConcept(t *testing.T) {

	// Create a list of fake concepts
	in.Concepts = []query.Concept{
		query.Concept{
			IsEncrypted: false,
			Concept:     "julien"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "francois"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "paul"},
	}
	CreateConceptInConductorDB(&query.InputCI{Concepts: in.Concepts})

	// Re-Create the three concepts
	inputCI := query.InputCI{Concepts: in.Concepts}
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept")
	err := ci.MakeRequest("POST", "", inputCI, nil)
	assert.Error(t, err)

	// Get the three concepts' hashes from Concept Indexor
	ci.DefineDispersActor("concept/julien:francois:paul/false")
	err = ci.MakeRequest("GET", "", nil, nil)

	assert.NoError(t, err)
	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)

	// Get the three concepts' hashes from query
	query, _ := NewQuery(&in)
	query.decryptConcept()
	assert.Equal(t, outputCI.Hashes, query.Concepts)

	// Delete the created concepts
	ci.DefineDispersActor("concept/julien:francois:paul/false")
	err = ci.MakeRequest("DELETE", "", nil, nil)
	assert.NoError(t, err)

}

func TestFetchListFromDB(t *testing.T) {

	in.Concepts = []query.Concept{
		query.Concept{
			IsEncrypted: false,
			Concept:     "aime les fraises"},
	}
	in.PseudoConcepts = make(map[string]string)
	in.PseudoConcepts["aime les fraises"] = "test1"

	// Create the four concepts
	CreateConceptInConductorDB(&query.InputCI{Concepts: in.Concepts})
	query, _ := NewQuery(&in)
	err := query.decryptConcept()
	assert.NoError(t, err)
	err = query.fetchListsOfInstancesFromDB()
	assert.NoError(t, err)

	// Delete the created concepts
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/aime les fraises/true")
	err = ci.MakeRequest("DELETE", "", nil, nil)
	assert.NoError(t, err)
}

func TestGetListsOfInstances(t *testing.T) {

	in.Concepts = []query.Concept{
		query.Concept{
			IsEncrypted: false,
			Concept:     "aime les fraises"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "aime les framboises"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "joue de la guitare"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "est designer chez cozy"},
	}
	in.PseudoConcepts = make(map[string]string)
	in.PseudoConcepts["aime les fraises"] = "test1"
	in.PseudoConcepts["aime les framboises"] = "test2"
	in.PseudoConcepts["joue de la guitare"] = "test3"
	in.PseudoConcepts["est designer chez cozy"] = "test4"

	// Create the four concepts
	CreateConceptInConductorDB(&query.InputCI{Concepts: in.Concepts})
	// Make few instances subscribe to Cozy-DISPERS
	_ = Subscribe(&subscribe.InputConductor{
		Concepts:          []string{"aime les fraises"},
		IsEncrypted:       false,
		EncryptedInstance: []byte("{\"domain\":\"caroline.mycozy.cloud\"}"),
	})
	_ = Subscribe(&subscribe.InputConductor{
		Concepts:          []string{"aime les fraises", "joue de la guitare", "aime les framboises"},
		IsEncrypted:       false,
		EncryptedInstance: []byte("{\"domain\":\"mathieu.mycozy.cloud\"}"),
	})
	_ = Subscribe(&subscribe.InputConductor{
		Concepts:          []string{"aime les fraises", "aime les framboises"},
		IsEncrypted:       false,
		EncryptedInstance: []byte("{\"domain\":\"zoe.mycozy.cloud\"}"),
	})
	_ = Subscribe(&subscribe.InputConductor{
		Concepts:          []string{"aime les fraises", "aime les framboises", "est designer chez cozy"},
		IsEncrypted:       false,
		EncryptedInstance: []byte("{\"domain\":\"thomas.mycozy.cloud\"}"),
	})

	query, _ := NewQuery(&in)
	err := query.decryptConcept()
	assert.NoError(t, err)
	err = query.fetchListsOfInstancesFromDB()
	assert.NoError(t, err)
	err = query.selectTargets()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(query.Targets))
	// Delete the created concepts
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/aime les fraises:aime les framboises:joue de la guitare:est designer chez cozy/false")
	err = ci.MakeRequest("DELETE", "", nil, nil)
	assert.NoError(t, err)

}

func TestShouldBeComputed(t *testing.T) {

	// Simulate one query
	// First layer to be computed
	queryDoc := &QueryDoc{
		QueryID: "thisisatestagaiiin",
		Layers: []query.LayerDA{
			query.LayerDA{Size: 4},
			query.LayerDA{Size: 4},
		},
	}

	bool, err := queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, true, bool)

	// Return of the first DA
	// ShouldBeComputed should be both layer to be computed
	query.NewAsyncTask("thisisatestagaiiin", 0, 3, query.AsyncAggregation)
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)

	query.NewAsyncTask("thisisatestagaiiin", 0, 0, query.AsyncAggregation)
	query.NewAsyncTask("thisisatestagaiiin", 0, 1, query.AsyncAggregation)
	query.NewAsyncTask("thisisatestagaiiin", 0, 2, query.AsyncAggregation)

	// Now every tasks are running
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)

	// 3 tasks over 4 are finished
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 0, 0)
	assert.NoError(t, err)
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 0, 3)
	assert.NoError(t, err)
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 0, 1)
	assert.NoError(t, err)
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)

	// The fourth DA has just finished
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 0, 2)
	assert.NoError(t, err)
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	state, err := query.FetchAsyncStateLayer("thisisatestagaiiin", 0, 4)
	assert.NoError(t, err)
	assert.Equal(t, query.Finished, state)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, true, bool)

	// 3 over 4 DA in second layer are finisehd
	query.NewAsyncTask("thisisatestagaiiin", 1, 0, query.AsyncAggregation)
	query.NewAsyncTask("thisisatestagaiiin", 1, 1, query.AsyncAggregation)
	query.NewAsyncTask("thisisatestagaiiin", 1, 2, query.AsyncAggregation)
	query.NewAsyncTask("thisisatestagaiiin", 1, 3, query.AsyncAggregation)
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 1, 3)
	assert.NoError(t, err)
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 1, 2)
	assert.NoError(t, err)
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 1, 1)
	assert.NoError(t, err)
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(2)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)

	// Last DA on last layer has just finished
	err = query.SetAsyncTaskAsFinished("thisisatestagaiiin", 1, 1)
	assert.NoError(t, err)
	bool, err = queryDoc.ShouldBeComputed(0)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(1)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)
	bool, err = queryDoc.ShouldBeComputed(2)
	assert.NoError(t, err)
	assert.Equal(t, false, bool)

}

func TestAggregate(t *testing.T) {

	var data []map[string]interface{}
	absPath, _ := filepath.Abs(strings.Join([]string{"../../assets/test/dummy_dataset.json"}, ""))
	buf, _ := ioutil.ReadFile(absPath)
	s := string(buf)
	json.Unmarshal([]byte(s), &data)

	// The first layer is going to sum each subset
	args1 := make(map[string]interface{})
	args1["keys"] = []string{"sepal_length", "sepal_width"}

	// The second layer is going to sum each results applying a weight
	// The weight has been passed as a new variable
	args2 := make(map[string]interface{})
	args2["keys"] = []string{"sepal_length", "sepal_width"}
	args2["weight"] = "length"
	layers := []query.LayerDA{
		query.LayerDA{
			AggregationFunctions: query.AggregationFunction{
				Function: "sum",
				Args:     args1,
			},
			Data: data,
			Size: 4,
		},
		query.LayerDA{
			AggregationFunctions: query.AggregationFunction{
				Function: "sum",
				Args:     args2,
			},
			Data: data,
			Size: 4,
		},
	}
	in.LayersDA = layers

	queryDoc, _ := NewQuery(&in)
	queryDoc.Layers = layers
	for indexLayer, layer := range queryDoc.Layers {
		layerShouldBeComputed, _ := queryDoc.ShouldBeComputed(indexLayer)
		if layerShouldBeComputed {
			if err := queryDoc.aggregateLayer(indexLayer, &layer); err != nil {
				assert.Error(t, err)
			}
		}
	}

}
