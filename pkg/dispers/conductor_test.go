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
		Type: query.UnionNode,
		LeftNode: query.OperationTree{
			Type:      query.UnionNode,
			LeftNode:  query.OperationTree{Type: 0, Value: "test1"},
			RightNode: query.OperationTree{Type: 0, Value: "test2"},
		},
		RightNode: query.OperationTree{
			Type:      query.UnionNode,
			LeftNode:  query.OperationTree{Type: 0, Value: "test3"},
			RightNode: query.OperationTree{Type: 0, Value: "test4"},
		},
	}

	in = query.InputNewQuery{
		DomainQuerier: "usr0.test.cozy.tools:8008",
		TargetProfile: targetProfile,
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

func TestErrorUnDefinedNumberActor(t *testing.T) {

	in2 := query.InputNewQuery{
		DomainQuerier: "usr0.test.cozy.tools:8008",
		TargetProfile: targetProfile,
	}

	_, err := NewQuery(&in2)
	assert.Error(t, err)
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

func TestQuery(t *testing.T) {
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
			Data:  data,
			Size:  4,
			State: map[string]query.StateDA{"0": query.Waiting, "1": query.Waiting, "2": query.Waiting, "3": query.Waiting},
		},
		query.LayerDA{
			AggregationFunctions: query.AggregationFunction{
				Function: "sum",
				Args:     args2,
			},
			Data:  data,
			Size:  4,
			State: map[string]query.StateDA{"0": query.Waiting, "1": query.Waiting, "2": query.Waiting, "3": query.Waiting},
		},
	}
	in.LayersDA = layers

	query, _ := NewQuery(&in)
	query.Layers = layers
	for indexLayer, layer := range query.Layers {
		if query.shouldBeComputed(indexLayer) {
			if err := query.aggregateLayer(indexLayer, &layer); err != nil {
				assert.Error(t, err)
			}
		}
	}

}
