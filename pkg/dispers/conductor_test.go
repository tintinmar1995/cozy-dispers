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

	in = query.OutputQ{
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
	ci := network.NewExternalActor(network.RoleCI)
	ci.MakeRequest("GET", "concept/julien/false", "application/json", nil)
	var outputCI query.OutputCI
	err = json.Unmarshal(ci.Out, &outputCI)
	assert.NoError(t, err)

	_, err = RetrieveSubscribeDoc(outputCI.Hashes[0].Hash)
	assert.NoError(t, err)

}

func TestSubscribe(t *testing.T) {

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
	ci := network.NewExternalActor(network.RoleCI)
	err = ci.MakeRequest("GET", "concept/aime les fraises/false", "application/json", nil)
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
}

func TestDefineConductor(t *testing.T) {

	_, err := NewConductor(&in)
	assert.NoError(t, err)
}

func TestErrorUnDefinedNumberActor(t *testing.T) {

	in2 := query.OutputQ{
		DomainQuerier: "usr0.test.cozy.tools:8008",
		TargetProfile: targetProfile,
	}

	_, err := NewConductor(&in2)
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
			Concept: "paul"},
	}
	CreateConceptInConductorDB(&query.InputCI{Concepts: in.Concepts})

	// Re-Create the three concepts
	inputCI := query.InputCI{Concepts: in.Concepts}
	marshaledInputCI, _ := json.Marshal(inputCI)
	ci := network.NewExternalActor("conceptindexor")
	err := ci.MakeRequest("POST", "concepts", "application/json", marshaledInputCI)
	assert.Error(t, err)

	// Get the three concepts' hashes from Concept Indexor
	err = ci.MakeRequest("GET", "concept/julien-francois-paul/false", "application/json", nil)
	assert.NoError(t, err)
	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)

	// Get the three concepts' hashes from conductor
	conductor, _ := NewConductor(&in)
	conductor.decryptConcept()
	assert.Equal(t, outputCI.Hashes, conductor.Query.Concepts)

	// Delete the created concepts
	err = ci.MakeRequest("DELETE", "concept/julien-paul-francois/true", "application/json", nil)
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
	conductor, _ := NewConductor(&in)
	err := conductor.decryptConcept()
	assert.NoError(t, err)
	err = conductor.fetchListsOfInstancesFromDB()
	assert.NoError(t, err)

	// Delete the created concepts
	ci := network.NewExternalActor("conceptindexor")
	err = ci.MakeRequest("DELETE", "concept/aime les fraises/true", "application/json", nil)
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

	conductor, _ := NewConductor(&in)
	err := conductor.decryptConcept()
	assert.NoError(t, err)
	err = conductor.fetchListsOfInstancesFromDB()
	assert.NoError(t, err)
	err = conductor.selectTargets()
	assert.NoError(t, err)
	assert.Equal(t, []string{"{\"domain\":\"caroline.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}", "{\"domain\":\"mathieu.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}", "{\"domain\":\"zoe.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}", "{\"domain\":\"thomas.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}", "{\"domain\":\"paul.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}", "{\"domain\":\"francois.mycozy.cloud\",\"date\":\"0001-01-01T00:00:00Z\",\"token\":{}}"}, conductor.Query.Targets)

	// Delete the created concepts
	ci := network.NewExternalActor("conceptindexor")
	err = ci.MakeRequest("DELETE", "concept/aime les fraises-aime les framboises-joue de la guitare-est designer chez cozy/false", "application/json", nil)
	assert.NoError(t, err)

}

func TestGetTargets(t *testing.T) {
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
			State: []query.StateDA{query.Waiting, query.Waiting, query.Waiting, query.Waiting},
		},
		query.LayerDA{
			AggregationFunctions: query.AggregationFunction{
				Function: "sum",
				Args:     args2,
			},
			Data:  data,
			Size:  4,
			State: []query.StateDA{query.Waiting, query.Waiting, query.Waiting, query.Waiting},
		},
	}

	conductor, _ := NewConductor(&in)
	conductor.Query.Layers = layers
	for indexLayer, layer := range conductor.Query.Layers {
		if conductor.shouldBeComputed(indexLayer) {
			if err := conductor.aggregateLayer(indexLayer, layer); err != nil {
				assert.Error(t, err)
			}
		}
	}

}
