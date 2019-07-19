package enclave

import (
	"crypto/sha256"
	"errors"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/crypto"
	"github.com/cozy/cozy-stack/pkg/dispers/dispers"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

const (
	defaultLenSalt = 128

	doctypeSalt = "io.cozy.hashconcept"
)

var PrefixerCI = prefixer.ConceptIndexorPrefixer

// ConceptDoc is used to save a concept's salt into Concept Indexor's database
// hash and salt are saved in byte to avoid string's to be interpreted
type ConceptDoc struct {
	ConceptID  string `json:"_id,omitempty"`
	ConceptRev string `json:"_rev,omitempty"`
	Concept    string `json:"concept,omitempty"`
	Hash       []byte `json:"hash,omitempty"`
	Salt       []byte `json:"salt,omitempty"`
}

// ID is used to get ID
func (t *ConceptDoc) ID() string {
	return t.ConceptID
}

// Rev is used to get Rev
func (t *ConceptDoc) Rev() string {
	return t.ConceptRev
}

// DocType is used to get DocType
func (t *ConceptDoc) DocType() string {
	return doctypeSalt
}

// Clone is used to create another ConceptDoc from this ConceptDoc
func (t *ConceptDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

// SetID is used to set the Doc ID
func (t *ConceptDoc) SetID(id string) {
	t.ConceptID = id
}

// SetRev is used to set Rev
func (t *ConceptDoc) SetRev(rev string) {
	t.ConceptRev = rev
}

func saveConcept(concept string) (*ConceptDoc, error) {

	hash, salt, err := createHash(concept)
	if err != nil {
		return &ConceptDoc{}, err
	}

	conceptDoc := &ConceptDoc{
		Concept: concept,
		Hash:    hash,
		Salt:    salt,
	}
	return conceptDoc, couchdb.CreateDoc(PrefixerCI, conceptDoc)
}

func getHash(concept string) ([]byte, error) {

	// Precise and run a mango query
	hash := []byte{}
	var out []ConceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", concept)}
	err := couchdb.FindDocs(PrefixerCI, doctypeSalt, req, &out)
	if err != nil {
		return hash, err
	}

	// Create a salt if no salt in the database for this concept
	if len(out) == 1 {
		hash = out[0].Hash
	} else if len(out) == 0 {
		return []byte{}, errors.New("Concept Indexor : no existing salt for this concept")
	} else {
		return []byte{}, errors.New("Concept Indexor : several salts for this concept")
	}

	return hash, err
}

func createHash(str string) ([]byte, []byte, error) {

	salt := crypto.GenerateRandomBytes(defaultLenSalt)

	h := sha256.New()
	h.Write(append([]byte(str), salt...))
	hash := h.Sum([]byte("concept"))

	return hash, salt, nil
}

func isConceptExisting(concept string) (bool, error) {

	// Precise and run the mango query
	var out []ConceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", concept)}
	err := couchdb.FindDocs(PrefixerCI, doctypeSalt, req, &out)
	if err != nil {
		return false, err
	}

	if len(out) > 0 {
		return true, nil
	}

	return false, nil
}

// CreateConcept checks if concept exists in db. If yes, return error. If no, create the salt, save the concept in db and return the hash.
func CreateConcept(in *dispers.Concept) error {

	if err := couchdb.EnsureDBExist(PrefixerCI, doctypeSalt); err != nil {
		return err
	}

	// Get salt with concept
	isExisting, err := isConceptExisting(in.Concept)
	if err != nil {
		return err
	}

	if isExisting {
		return errors.New("Concept is already existing")
	}
	doc, err := saveConcept(in.Concept)
	if err != nil {
		return err
	}
	in.Hash = doc.Hash

	return err
}

// GetConcept gets a concept from db. If no, return error.
func GetConcept(in *dispers.Concept) error {

	hash, err := getHash(in.Concept)
	in.Hash = hash
	return err
}

// DecryptConcept has to be used before CreateConcept or DeleteConcept
func DecryptConcept(in *dispers.Concept) error {
	// TODO: Decrypte concept with private key
	in.Concept = string(in.EncryptedConcept)
	return nil
}

// DeleteConcept is used to delete a concept in ConceptIndexor Database.
func DeleteConcept(in *dispers.Concept) error {

	// Precise and run the mango query
	var out []ConceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", in.Concept)}
	err := couchdb.FindDocs(PrefixerCI, doctypeSalt, req, &out)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("No concept to delete. " + in.Concept + " not found")
	}

	// Delete every doc that match concept
	for _, element := range out {
		err = couchdb.DeleteDoc(PrefixerCI, &element)
		if err != nil {
			return err
		}
	}

	return nil
}
