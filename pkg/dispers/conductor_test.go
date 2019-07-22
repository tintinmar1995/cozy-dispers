package enclave

import (
	"encoding/json"
	"testing"

	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/dispers/subscribe"
	"github.com/stretchr/testify/assert"
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
