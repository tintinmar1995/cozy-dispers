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

/*
// GetTrainingState can be called by the querier to have some information about
// the process. GetTrainingState retrieves information from the Conductor's database
// It retreives both the query.QueryDoc and the Metadata
func GetTrainingState(id string) ([]query.MetadataDoc, query.QueryDoc, error) {
	fetched := &query.QueryDoc{}
	err := couchdb.GetDoc(prefixer.ConductorPrefixer, "io.cozy.ml", id, fetched)
	if err != nil {
		return nil, *fetched, err
	}

	metas, err := query.RetrieveMetadata(id)
	if err != nil {
		return nil, *fetched, err
	}

	return metas, *fetched, nil
}
*/

// Conductor collects every actors playing a part to the query
type Conductor struct {
	Query           query.QueryDoc
	Conceptindexors network.ExternalActor
	Targetfinders   network.ExternalActor
	Targets         network.ExternalActor
	DataAggregators [][]network.ExternalActor
}

// NewConductor returns a Conductor object to lead the request
func NewConductor(in query.OutputQ) (*Conductor, error) {

	if in.NumberActors == nil {
		return &Conductor{}, errors.New("Number of network.ExternalActors should be defined")
	}

	da := make([][]network.ExternalActor, len(in.SizeAggrLayers))
	for i, v := range in.SizeAggrLayers {
		da[i] = network.NewSliceOfExternalActors("dataaggregator", v)
	}

	// Creating the query.QueryDoc that will be saved in the Conductor's database
	queryDoc := query.QueryDoc{
		CheckPoints:                 make([]bool, 6),
		Concepts:                    in.Concepts,
		DomainQuerier:               in.DomainQuerier,
		IsEncrypted:                 in.IsEncrypted,
		Jobs:                        in.Jobs,
		LocalQuery:                  in.LocalQuery,
		NumberActors:                in.NumberActors,
		PseudoConcepts:              in.PseudoConcepts,
		SizeAggrLayers:              in.SizeAggrLayers,
		TargetProfile:               in.TargetProfile,
		EncryptedAggregateFunctions: in.EncryptedAggregateFunctions,
		EncryptedConcepts:           in.EncryptedConcepts,
		EncryptedLocalQuery:         in.EncryptedLocalQuery,
		EncryptedTargetProfile:      in.EncryptedTargetProfile,
	}
	if err := couchdb.CreateDoc(prefixer.ConductorPrefixer, &queryDoc); err != nil {
		return &Conductor{}, err
	}

	retour := &Conductor{
		Query:           queryDoc,
		Conceptindexors: network.NewExternalActor("conceptindexor"),
		Targetfinders:   network.NewExternalActor("targetfinder"),
		Targets:         network.NewExternalActor("target"),
		DataAggregators: da,
	}

	return retour, nil
}

// decryptConcept returns a list of hashed concepts from a list of encrypted concepts
func (c *Conductor) decryptConcept() error {

	job := "concept/"
	listOfConcepts := []string{}
	for _, concept := range c.Query.Concepts {
		if concept.IsEncrypted {
			listOfConcepts = append(listOfConcepts, string(concept.EncryptedConcept))
		} else {
			listOfConcepts = append(listOfConcepts, concept.Concept)
		}
	}
	job = job + strings.Join(listOfConcepts, "-")
	if c.Query.IsEncrypted {
		job = job + "/true"
	} else {
		job = job + "/false"
	}

	c.Conceptindexors.MakeRequest("GET", job, "", nil)

	var outputCI query.OutputCI
	json.Unmarshal(c.Conceptindexors.Out, &outputCI)
	c.Query.CheckPoints[0] = true
	c.Query.Concepts = outputCI.Hashes
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) fetchListsOfInstancesFromDB() error {

	encListsOfA := make(map[string][]byte)

	for _, concept := range c.Query.Concepts {

		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return err
		}

		encListsOfA[c.Query.PseudoConcepts[string(concept.EncryptedConcept)]] = s[0].EncryptedInstances

	}

	if c.Query.IsEncrypted {
		// TODO : encrypt ListsOfAddresses
		res, _ := json.Marshal(encListsOfA)
		c.Query.EncryptedListsOfAddresses = res
	} else {
		res, _ := json.Marshal(encListsOfA)
		var listsOfA map[string][]string
		json.Unmarshal(res, &listsOfA)
		c.Query.ListsOfAddresses = listsOfA
	}

	c.Query.CheckPoints[1] = true
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) selectTargets() error {

	inputTF := query.InputTF{
		IsEncrypted:               c.Query.IsEncrypted,
		ListsOfAddresses:          c.Query.ListsOfAddresses,
		TargetProfile:             c.Query.TargetProfile,
		EncryptedListsOfAddresses: c.Query.EncryptedListsOfAddresses,
		EncryptedTargetProfile:    c.Query.EncryptedTargetProfile,
	}

	marshalledInputTF, err := json.Marshal(inputTF)
	if err != nil {
		return err
	}

	err = c.Targetfinders.MakeRequest("POST", "addresses", "application/json", marshalledInputTF)
	if err != nil {
		return err
	}

	var outputTF query.OutputTF
	json.Unmarshal(c.Targetfinders.Out, &outputTF)
	c.Query.EncryptedTargets = outputTF.EncryptedListOfAddresses
	c.Query.Targets = outputTF.ListOfAddresses
	c.Query.CheckPoints[2] = true
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) makeLocalQuery() error {

	inputT := query.InputT{
		IsEncrypted:         c.Query.IsEncrypted,
		LocalQuery:          c.Query.LocalQuery,
		Targets:             c.Query.Targets,
		EncryptedLocalQuery: c.Query.EncryptedLocalQuery,
		EncryptedTargets:    c.Query.EncryptedTargets,
	}

	marshalledInputT, _ := json.Marshal(inputT)

	err := c.Targets.MakeRequest("POST", "gettokens", "application/json", marshalledInputT)
	if err != nil {
		return err
	}
	var outputT query.OutputT
	json.Unmarshal(c.Targets.Out, &outputT)
	c.Query.Data = outputT.Data
	c.Query.CheckPoints[3] = true
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) aggregate() error {

	// Distributed data in c.Query.SizeAggrLayers[0] parts

	// for each aggregation layer
	var inputDA query.InputDA
	outputDA := make([][]query.OutputDA, len(c.Query.SizeAggrLayers))
	for indexLayer, sizeLayer := range c.Query.SizeAggrLayers {

		inputDA = query.InputDA{
			Job:          c.Query.Jobs[indexLayer],
			IsEncrypted:  c.Query.IsEncrypted,
			EncryptedJob: c.Query.EncryptedAggregateFunctions[indexLayer],
		}
		outputDA[indexLayer] = make([]query.OutputDA, sizeLayer)
		marshalledInputDA, _ := json.Marshal(inputDA)

		for indexDA, da := range c.DataAggregators[indexLayer] {

			err := c.Targetfinders.MakeRequest("POST", "aggregation", "application/json", marshalledInputDA)
			if err != nil {
				return err
			}

			var out query.OutputDA
			json.Unmarshal(da.Out, &out)
			outputDA[indexLayer][indexDA] = out
		}
	}

	c.Query.CheckPoints[5] = true
	c.Query.OutputsDA = outputDA
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

// Lead is the most general method. This is the only one used in CMD and Web's files. It will use the 5 previous methods to work
func (c *Conductor) Lead() error {

	if c.Query.CheckPoints[0] != true {
		if err := c.decryptConcept(); err != nil {
			return err
		}
	}

	if c.Query.CheckPoints[1] != true {
		if err := c.fetchListsOfInstancesFromDB(); err != nil {
			return err
		}
	}

	if c.Query.CheckPoints[2] != true {
		if err := c.selectTargets(); err != nil {
			return err
		}
	}

	if c.Query.CheckPoints[3] != true {
		if err := c.makeLocalQuery(); err != nil {
			return err
		}
	}

	// TODO: Deal with Async Method
	if c.Query.CheckPoints[4] != true {
		if err := c.aggregate(); err != nil {
			return err
		}
	}

	// TODO: Notify the querier
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

// RetrieveSubscribeDoc is used to get a Subscribe doc from the Conductor's database.
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
		Role: network.RoleCI,
	}
	marshalInputCI, err := json.Marshal(*in)
	if err != nil {
		return err
	}

	// try to create concept with route POST
	// try to get hash with CI's route GET
	// if error, returns the first error that occurred between POST et Get
	errPost := ci.MakeRequest("POST", "concept", "application/json", marshalInputCI)
	path := ""
	for index, concept := range in.Concepts {
		if concept.IsEncrypted {
			path = path + string(concept.EncryptedConcept)
		} else {
			path = path + concept.Concept
		}
		if index != (len(in.Concepts) - 1) {
			path = path + ":"
		}
	}
	path = path + "/" + strconv.FormatBool(in.Concepts[0].IsEncrypted)
	err = ci.MakeRequest("GET", "concept/"+path, "application/json", marshalInputCI)
	if err != nil {
		if errPost != nil {
			return errPost
		}
		return err
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
	ci := network.NewExternalActor(network.RoleCI)
	err := ci.MakeRequest("GET", "concept/"+strings.Join(in.Concepts, ":")+"/"+strconv.FormatBool(in.IsEncrypted), "application/json", nil)
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
		tf := network.NewExternalActor(network.RoleTF)
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
		t := network.NewExternalActor(network.RoleT)
		t.SubscribeMode()
		err = t.MakeRequest("POST", "insert", "application/json", tf.Out)
		if err != nil {
			return err
		}

		// Ask Target Finder to Encrypt
		tf = network.NewExternalActor(network.RoleTF)
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
