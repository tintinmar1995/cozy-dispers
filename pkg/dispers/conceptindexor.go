package enclave

import (
	"errors"
	"strings"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/crypto"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

type conceptDoc struct {
	conceptID  string `json:"_id,omitempty"`
	conceptRev string `json:"_rev,omitempty"`
	concept    string `json:"concept,omitempty"`
	salt       string `json:"salt,omitempty"`
}

func (t *conceptDoc) ID() string {
	return t.conceptID
}

func (t *conceptDoc) Rev() string {
	return t.conceptRev
}

func (t *conceptDoc) DocType() string {
	return "io.cozy.hashconcept"
}

func (t *conceptDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

func (t *conceptDoc) SetID(id string) {
	t.conceptID = id
}

func (t *conceptDoc) SetRev(rev string) {
	t.conceptRev = rev
}

func addSalt(concept string) error {

	conceptDoc := &conceptDoc{
		conceptID:  "",
		conceptRev: "",
		concept:    concept,
		salt:       string(crypto.GenerateRandomBytes(5)),
	}

	return couchdb.CreateDoc(prefixer.ConceptIndexorPrefixer, conceptDoc)
}

func getSalt(concept string) (string, error) {

	salt := ""
	err := couchdb.DefineIndex(prefixer.ConceptIndexorPrefixer, mango.IndexOnFields("io.cozy.hashconcept", "concept-index", []string{"concept"}))
	if err != nil {
		return salt, err
	}

	var out []conceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", concept)}
	err = couchdb.FindDocs(prefixer.ConceptIndexorPrefixer, "io.cozy.hashconcept", req, out)
	if err != nil {
		return salt, err
	}

	if len(out) == 1 {
		salt = out[0].salt
	} else if len(out) == 0 {
		// TODO: Creation an error
		errors.New("Concept Indexor : no existing salt for this concept")
	} else {
		errors.New("Concept Indexor : several salts for this concept")
	}

	return salt, err
}

func hash(str string) string {
	return ""
}

func isConceptExisting(concept string) (bool, error) {

	var out []conceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", concept)}
	err := couchdb.FindDocs(prefixer.ConceptIndexorPrefixer, "io.cozy.hashconcept", req, out)
	if err != nil {
		return false, err
	}

	if len(out) > 0 {
		return true, nil
	}

	return false, nil
}

/*
DeleteConcept is used to delete a salt in ConceptIndexor Database. It has to be
used to update a salt.
*/
func DeleteConcept(encryptedConcept string) error {

	// TODO: Decrypte concept with private key
	concept := encryptedConcept

	// TODO: Delete document in database
	var out []conceptDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("concept", concept)}
	err := couchdb.FindDocs(prefixer.ConceptIndexorPrefixer, "io.cozy.hashconcept", req, out)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("No concept to delete")
	}

	// Delete every doc that match concept
	for _, element := range out {
		err = couchdb.DeleteDoc(prefixer.ConceptIndexorPrefixer, &element)
		if err != nil {
			return err
		}
	}

	return err
}

/*
HashMeThat is used to get a concept's salt. If the salt is absent from CI database
the function creates the salt and notify the user that the salt has just been created
*/
func HashMeThat(encryptedConcept string) (string, error) {
	couchdb.EnsureDBExist(prefixer.ConceptIndexorPrefixer, "io.cozy.hashconcept")

	// TODO: Decrypte concept with private key
	concept := encryptedConcept

	isExisting, err := isConceptExisting(concept)
	if err != nil {
		return "", err
	}

	if isExisting {
		// Write in Metadata that concept has been retrieved
	} else {
		err = addSalt(concept)
		if err != nil {
			return "", err
		}
		// Write in Metadata that concept has been created
	}

	// Get salt with hash(concept)
	salt, err := getSalt(hash(concept))
	if err != nil {
		return "", err
	}

	// Merge concept and salt
	justHashed := hash(strings.Join([]string{concept, salt}, ""))

	return justHashed, nil
}
