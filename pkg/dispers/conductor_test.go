package enclave

import (
	"bytes"
	"encoding/json"
	"net/http"
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

	_, err := NewConductor(in)
	assert.NoError(t, err)
}

func TestErrorUnDefinedNumberActor(t *testing.T) {

	in2 := query.OutputQ{
		DomainQuerier: "usr0.test.cozy.tools:8008",
		TargetProfile: targetProfile,
	}

	_, err := NewConductor(in2)
	assert.Error(t, err)
}

func TestDecryptConcept(t *testing.T) {

	// Create an instance of Conductor
	conductor, _ := NewConductor(in)
	// Create a list of fake concepts
	conductor.Query.Concepts = []query.Concept{
		query.Concept{
			IsEncrypted: false,
			Concept:     "julien"},
		query.Concept{
			IsEncrypted: false,
			Concept:     "francois"},
		query.Concept{
			Concept: "paul"},
	}

	// Create the three concepts
	inputCI := query.InputCI{Concepts: conductor.Query.Concepts}
	marshaledInputCI, _ := json.Marshal(inputCI)
	ci := network.NewExternalActor("conceptindexor")
	ci.MakeRequest("POST", "concepts", "application/json", marshaledInputCI)

	// Get the three concepts' hashes from Concept Indexor
	ci.MakeRequest("GET", "concept/julien-francois-paul/false", "application/json", nil)
	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)

	// Get the three concepts' hashes from conductor
	conductor.decryptConcept()
	assert.Equal(t, outputCI.Hashes, conductor.Query.Concepts)

	client := &http.Client{}
	dispersURL.Path = "dispers/conceptindexor/concept/julien-paul-francois/true"
	req, err := http.NewRequest("DELETE", dispersURL.String(), nil)
	assert.NoError(t, err)
	_, err = client.Do(req)
	assert.NoError(t, err)

}

func TestGetListsOfInstances(t *testing.T) {

	// Create an instance of Conductor
	conductor, _ := NewConductor(in)
	// Create a list of fake concepts
	conductor.Query.Concepts = []query.Concept{
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

	// Pseudo-anonymize concepts
	in.PseudoConcepts = make(map[string]string)
	in.PseudoConcepts["aime les fraises"] = "test1"
	in.PseudoConcepts["aime les framboises"] = "test2"
	in.PseudoConcepts["joue de la guitare"] = "test3"
	in.PseudoConcepts["est designer chez cozy"] = "test4"

	// Create the four concepts
	inputCI := query.InputCI{Concepts: conductor.Query.Concepts}
	marshalledInputCI, _ := json.Marshal(inputCI)
	dispersURL.Path = "dispers/conceptindexor/concept"
	_, err := http.Post(dispersURL.String(), "application/json", bytes.NewReader(marshalledInputCI))
	assert.NoError(t, err)

	/*
		// Make few instances subscribe to Cozy-DISPERS
		inSubs := query.InputSubscribeMode{
			Concepts:    conductor.Query.Concepts,
			IsEncrypted: false,
			Instance:    "{\"Domain\":\"joel.mycozy.cloud\",\"SubscriptionDate\": \"time.Now()\"}",
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
		inSubs = query.InputSubscribeMode{
			Concepts:    conductor.Query.Concepts[:1],
			IsEncrypted: false,
			Instance:    "{\"Domain\":\"paul.mycozy.cloud\",\"SubscriptionDate\": \"time.Now()\"}",
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
		inSubs = query.InputSubscribeMode{
			Concepts:    conductor.Query.Concepts[2:],
			IsEncrypted: false,
			Instance:    "{\"Domain\":\"francois.mycozy.cloud\",\"SubscriptionDate\": \"time.Now()\"}",
		}
		err = Subscribe(&inSubs)
		assert.NoError(t, err)
	*/
}

func TestGetTargets(t *testing.T) {
}

func TestQuery(t *testing.T) {
}

func TestAggregate(t *testing.T) {
}
