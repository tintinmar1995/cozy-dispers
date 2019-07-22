package enclave

import (
	"testing"

	"github.com/cozy/cozy-stack/pkg/crypto"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/stretchr/testify/assert"
)

var conceptTestRandom = query.Concept{Concept: string(crypto.GenerateRandomBytes(50))}
var conceptTest = query.Concept{Concept: "FrançoisEtPaul"}

func TestConceptIndexor(t *testing.T) {

	// Test if concept exist (conceptTestRandom is supposed to be new)
	b, err := isConceptExisting(conceptTestRandom.Concept)
	assert.NoError(t, err)
	assert.Equal(t, b, false)

	// Create new conceptDoc with CreateConcept
	err = CreateConcept(&conceptTest)
	out := conceptTest.Hash
	assert.NoError(t, err)
	err2 := CreateConcept(&conceptTestRandom)
	assert.NoError(t, err2)

	// Test if concept exist using isConceptExisting
	a, err2 := isConceptExisting(conceptTest.Concept)
	assert.NoError(t, err2)
	assert.Equal(t, a, true)
	b2, errb := isConceptExisting(conceptTestRandom.Concept)
	assert.NoError(t, errb)
	assert.Equal(t, b2, true)

	// Check GetConcept
	err = GetConcept(&conceptTest)
	assert.NoError(t, err)
	assert.Equal(t, out, conceptTest.Hash)

	// Create twice the same concept and check if there is error when getting salt
	_, errAdd := saveConcept(conceptTest.Concept)
	assert.NoError(t, errAdd)
	errGet := CreateConcept(&conceptTest)
	assert.Error(t, errGet)

	// deleteConcept
	errD := DeleteConcept(&conceptTest)
	errDb := DeleteConcept(&conceptTestRandom)
	assert.NoError(t, errD)
	assert.NoError(t, errDb)

	// Test if concept exist
	ad, err := isConceptExisting(conceptTest.Concept)
	bd, errb := isConceptExisting(conceptTestRandom.Concept)
	assert.NoError(t, err)
	assert.NoError(t, errb)
	assert.Equal(t, ad, false)
	assert.Equal(t, bd, false)

}

func TestAddSalt(t *testing.T) {
	_, errAdd := saveConcept(conceptTestRandom.Concept)
	assert.NoError(t, errAdd)
}

func TestDeleteConcept(t *testing.T) {
	saveConcept(conceptTestRandom.Concept)
	err := DeleteConcept(&conceptTestRandom)
	assert.NoError(t, err)
}

func TestGetHash(t *testing.T) {
	saveConcept(conceptTestRandom.Concept)
	_, err := getHash(conceptTestRandom.Concept)
	assert.NoError(t, err)
}

func TestHash(t *testing.T) {
	res, _, err := createHash(conceptTestRandom.Concept)
	assert.NoError(t, err)
	res2, _, _ := createHash(conceptTestRandom.Concept)
	assert.NotEqual(t, res, res2)
	assert.Equal(t, true, len(res) > 0)
}

func TestCreateConcept(t *testing.T) {
	conceptTestRandom.Concept = string(crypto.GenerateRandomBytes(50))
	err := CreateConcept(&conceptTestRandom)
	out := conceptTestRandom.Hash
	assert.NoError(t, err)
	assert.Equal(t, true, len(out) > 0)
	err2 := GetConcept(&conceptTestRandom)
	out2 := conceptTestRandom.Hash
	assert.NoError(t, err2)
	assert.Equal(t, true, len(out2) > 0)
	assert.Equal(t, out, out2)
}

func TestHashIsNotDeterministic(t *testing.T) {
	conceptTestRandom.Concept = string(crypto.GenerateRandomBytes(50))
	err := CreateConcept(&conceptTestRandom)
	assert.NoError(t, err)
	out1 := conceptTestRandom.Hash
	assert.Equal(t, true, len(out1) > 0)

	err = DeleteConcept(&conceptTestRandom)
	assert.NoError(t, err)

	err = CreateConcept(&conceptTestRandom)
	assert.NoError(t, err)
	out2 := conceptTestRandom.Hash
	assert.Equal(t, true, len(out2) > 0)
	assert.NotEqual(t, out1, out2)

	err = DeleteConcept(&conceptTestRandom)
	assert.NoError(t, err)
}

func TestIsExistantSaltExisting(t *testing.T) {
	conceptTestRandom.Concept = string(crypto.GenerateRandomBytes(50))
	saveConcept(conceptTestRandom.Concept)
	_, err := isConceptExisting(conceptTestRandom.Concept)
	assert.NoError(t, err)
}

func TestIsUnexistantSaltExisting(t *testing.T) {
	conceptTestRandom.Concept = string(crypto.GenerateRandomBytes(50))
	_, err := isConceptExisting(conceptTestRandom.Concept)
	assert.NoError(t, err)
}
