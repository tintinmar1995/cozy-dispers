package enclave

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/dispers/network"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/dispers/subscribe"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var (
	// hosts is a list of hosts where DISPERS is running. Those hosts can be called to play the role of CI/TF/T/DA
	hosts = []url.URL{
		url.URL{
			Scheme: "http",
			Host:   "localhost:8008",
		},
	}
	prefixerC = prefixer.ConductorPrefixer
)

// Retrievesubscribe is used to get a Subscribe doc from the Conductor's database.
// It returns an error if there is more than 1 subscribe doc
func RetrieveSubscribeDoc(hash []byte) ([]subscribe.SubscribeDoc, error) {

	var out []subscribe.SubscribeDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("hash", hash)}
	err := couchdb.FindDocs(prefixerC, "io.cozy.instances", req, &out)
	if err != nil {
		return out, err
	}

	if len(out) > 1 {
		return out, errors.New("There is more than 1 subscribe doc in database for this concept")
	}

	return out, nil
}

// CreateConceptInConductorDB is used to add a concept to Cozy-DISPERS
func CreateConceptInConductorDB(in *query.InputCI) error {
	// Pass to CI
	ci := network.ExternalActor{
		URL:  hosts[0],
		Role: "conceptindexor",
	}
	marshalInputCI, err := json.Marshal(*in)
	if err != nil {
		return err
	}

	if err := ci.MakeRequest("POST", "concept", "application/json", marshalInputCI); err != nil {
		if strings.Contains(err.Error(), "Concept is already existing") {
			path := ""
			for index, concept := range in.Concepts {
				if concept.IsEncrypted {
					path = path + string(concept.EncryptedConcept)
				} else {
					path = path + concept.Concept
				}
				if index != (len(in.Concepts) - 1) {
					path = path + "-"
				}
			}
			path = path + "/" + strconv.FormatBool(in.Concepts[0].IsEncrypted)
			err = ci.MakeRequest("GET", "concept/"+path, "application/json", marshalInputCI)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Get CI's result
	var out query.OutputCI
	err = json.Unmarshal(ci.Out, &out)
	if err != nil {
		return err
	}

	for _, concept := range out.Hashes {
		// TODO: Check if Concept is unexistant
		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return nil
		}

		if len(s) < 1 {
			sub := subscribe.SubscribeDoc{
				Hash: concept.Hash,
			}
			if err := couchdb.CreateDoc(prefixerC, &sub); err != nil {
				return err
			}
		} else {
			return errors.New("This concept already exists in Conductor's database")
		}
	}

	return nil
}

func Subscribe(in *subscribe.InputConductor) error {

	// Get Concepts' hash
	ci := network.NewExternalActor("conceptindexor")
	err := ci.MakeRequest("GET", "concept/"+strings.Join(in.Concepts, "-")+"/"+strconv.FormatBool(in.IsEncrypted), "application/json", nil)
	if err != nil {
		return err
	}
	var outputCI query.OutputCI
	err = json.Unmarshal(ci.Out, &outputCI)
	if err != nil {
		return err
	}

	var docs []subscribe.SubscribeDoc
	var outEnc subscribe.OutputEncrypt

	for _, hash := range outputCI.Hashes {

		// Get SubscribeDoc from db
		docs, err = RetrieveSubscribeDoc(hash.Hash)
		if err != nil {
			return err
		}
		if len(docs) == 0 {
			return errors.New("SubscribeDoc not found. Are you sure this concept exist ?")
		}
		doc := docs[0]

		// Ask Target Finder to Decrypt
		tf := network.NewExternalActor("targetfinder")
		tf.SubscribeMode()

		marshalledInputDecrypt, err := json.Marshal(subscribe.InputDecrypt{
			IsEncrypted:        in.IsEncrypted,
			EncryptedInstances: doc.EncryptedInstances,
			EncryptedInstance:  in.EncryptedInstance,
		})
		if err != nil {
			return err
		}

		err = tf.MakeRequest("POST", "decrypt", "application/json", marshalledInputDecrypt)
		if err != nil {
			return err
		}

		// Ask Target to add instance
		t := network.NewExternalActor("target")
		t.SubscribeMode()
		err = t.MakeRequest("POST", "insert", "application/json", tf.Out)
		if err != nil {
			return err
		}

		// Ask Target Finder to Encrypt
		tf = network.NewExternalActor("targetfinder")
		tf.SubscribeMode()
		err = tf.MakeRequest("POST", "decrypt", "application/json", t.Out)
		if err != nil {
			return err
		}
		err = json.Unmarshal(tf.Out, &outEnc)
		if err != nil {
			return err
		}

		doc.EncryptedInstances = outEnc.EncryptedInstances
		couchdb.UpdateDoc(prefixerC, &doc)
	}

	return nil
}
