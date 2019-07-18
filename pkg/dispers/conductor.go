package enclave

import (
	"encoding/json"
	"errors"
	"math/rand"
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

// Conductor collects every actors playing a part to the query
type Conductor struct {
	Query           query.QueryDoc
	Conceptindexors network.ExternalActor
	Targetfinders   network.ExternalActor
	Targets         network.ExternalActor
	DataAggregators [][]network.ExternalActor
}

// NewConductor returns a Conductor object to lead the request
func NewConductor(in *query.OutputQ) (*Conductor, error) {

	if in.NumberActors == nil {
		return &Conductor{}, errors.New("Number of network.ExternalActors should be defined")
	}

	da := make([][]network.ExternalActor, len(in.LayersDA))
	for i, layer := range in.LayersDA {
		da[i] = network.NewSliceOfExternalActors("dataaggregator", layer.Size)
	}

	// Creating the query.QueryDoc that will be saved in the Conductor's database
	queryDoc := query.QueryDoc{
		CheckPoints:            make([]bool, 6),
		Concepts:               in.Concepts,
		DomainQuerier:          in.DomainQuerier,
		IsEncrypted:            in.IsEncrypted,
		Layers:                 in.LayersDA,
		LocalQuery:             in.LocalQuery,
		NumberActors:           in.NumberActors,
		PseudoConcepts:         in.PseudoConcepts,
		TargetProfile:          in.TargetProfile,
		EncryptedConcepts:      in.EncryptedConcepts,
		EncryptedLocalQuery:    in.EncryptedLocalQuery,
		EncryptedTargetProfile: in.EncryptedTargetProfile,
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

// NewConductorFetchingQueryDoc returns a Conductor object to lead the request
func NewConductorFetchingQueryDoc(queryid string) (*Conductor, error) {

	queryDoc := &query.QueryDoc{}
	err := couchdb.GetDoc(prefixer.ConductorPrefixer, "io.cozy.query", queryid, queryDoc)
	if err != nil {
		return &Conductor{}, err
	}

	da := make([][]network.ExternalActor, len(queryDoc.Layers))
	for i, layer := range queryDoc.Layers {
		da[i] = network.NewSliceOfExternalActors("dataaggregator", layer.Size)
	}

	retour := &Conductor{
		Query:           *queryDoc,
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

	err := c.Conceptindexors.MakeRequest("GET", job, "", nil)
	if err != nil {
		return err
	}

	var outputCI query.OutputCI
	json.Unmarshal(c.Conceptindexors.Out, &outputCI)
	c.Query.CheckPoints[0] = true
	c.Query.Concepts = outputCI.Hashes

	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) fetchListsOfInstancesFromDB() error {

	encListsOfA := make(map[string][]byte)
	listsOfA := make(map[string][]string)

	for _, concept := range c.Query.Concepts {

		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return err
		}

		if len(s) == 0 {
			return errors.New("Cannot find SubscribeDoc associated to hash : " + string(concept.Hash))
		}

		if c.Query.IsEncrypted {
			encListsOfA[c.Query.PseudoConcepts[string(concept.EncryptedConcept)]] = s[0].EncryptedInstances
			res, _ := json.Marshal(encListsOfA)
			c.Query.EncryptedListsOfAddresses = res

		} else {
			// Pretty ugly way to convert EncryptedInstance to []string.
			// This part will be removed when clearing Inputs and Outputs
			var insts []query.Instance
			err = json.Unmarshal(s[0].EncryptedInstances, &insts)
			if err != nil {
				return err
			}
			instsStr := make([]string, len(insts))
			for index, ins := range insts {
				marshalledIns, _ := json.Marshal(ins)
				instsStr[index] = string(marshalledIns)
			}
			listsOfA[c.Query.PseudoConcepts[concept.Concept]] = instsStr
		}
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

	marshalledInputT, err := json.Marshal(inputT)
	if err != nil {
		return err
	}

	err = c.Targets.MakeRequest("POST", "query", "application/json", marshalledInputT)
	if err != nil {
		return err
	}
	var outputT query.OutputT
	err = json.Unmarshal(c.Targets.Out, &outputT)
	if err != nil {
		return err
	}
	c.Query.Layers[0].Data = outputT.Data
	c.Query.CheckPoints[3] = true
	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

func (c *Conductor) aggregateLayer(indexLayer int, layer query.LayerDA) error {

	var da network.ExternalActor

	// Shuffle Data
	rand.Shuffle(len(layer.Data), func(i, j int) {
		layer.Data[i], layer.Data[j] = layer.Data[j], layer.Data[i]
	})

	// Separate data in sizeLayer folds
	seps := make([]int, layer.Size+1)
	seps[0] = 0
	if len(layer.Data)%layer.Size != 0 {
		seps[len(seps)-1] = len(layer.Data) - 1
	}
	for indexSep := 1; indexSep < len(seps)-1; indexSep++ {
		seps[indexSep] = (len(layer.Data) / layer.Size) * indexSep
	}

	// Create InputDA for the layer
	inputDA := query.InputDA{
		Job:          layer.AggregationFunctions,
		EncryptedJob: layer.EncryptedAggregateFunctions,
	}

	// for each node on the layer
	for indexDA := 0; indexDA < layer.Size; indexDA++ {
		da = c.DataAggregators[indexLayer][indexDA]

		inputDA.Data = layer.Data[seps[indexDA]:seps[indexDA+1]]

		// marshal inputDA and make request for async process
		marshalledInputDA, _ := json.Marshal(inputDA)
		err := da.MakeRequest("POST", "aggregation", "application/json", marshalledInputDA)
		if err != nil {
			return err
		}
		var out query.OutputDA
		err = json.Unmarshal(da.Out, &out)
		if err != nil {
			return err
		}

		// async task is now running
		layer.State[indexDA] = query.Running
	}

	return nil
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
		// Stop the process and wait for T's answer to resume
		return nil
	}

	if c.Query.CheckPoints[4] == true || c.Query.CheckPoints[5] != true {
		for indexLayer, layer := range c.Query.Layers {
			if c.shouldBeComputed(indexLayer) {
				if err := c.aggregateLayer(indexLayer, layer); err != nil {
					return err
				}
				// Stop the process and wait for DA's answer to resume
				return nil
			}
		}
	}

	return couchdb.UpdateDoc(prefixer.ConductorPrefixer, &c.Query)
}

// shouldBeComputed returns true if indexLayer-1 finished and indexLayer not done
func (c *Conductor) shouldBeComputed(indexLayer int) bool {

	// shouldBeComputed returns false if indexLayer finished
	isLayerFinished := true
	for _, stateDA := range c.Query.Layers[indexLayer].State {
		if stateDA != query.Finished {
			isLayerFinished = false
		}
	}
	if isLayerFinished {
		return false
	}

	// shouldBeComputed returns false if indexLayer-1 is not finished
	isPreviousLayerFinished := true
	for _, stateDA := range c.Query.Layers[indexLayer-1].State {
		if stateDA != query.Finished {
			isPreviousLayerFinished = false
		}
	}
	if !isPreviousLayerFinished {
		return false
	}

	// if !isLayerFinished && isPreviousLayerFinished
	// shouldBeComputed returns true if indexLayer-1 finished and indexLayer waiting
	isLayerWaiting := true
	for _, stateDA := range c.Query.Layers[indexLayer-1].State {
		if stateDA != query.Waiting {
			isLayerWaiting = false
		}
	}
	return isLayerWaiting
}

// RetrieveSubscribeDoc is used to get a Subscribe doc from the Conductor's database.
// It returns an error if there is more than 1 subscribe doc
func RetrieveSubscribeDoc(hash []byte) ([]subscribe.SubscribeDoc, error) {

	var out []subscribe.SubscribeDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("hash", hash)}
	err := couchdb.FindDocs(prefixerC, "io.cozy.instances", req, &out)
	if err != nil {
		return nil, err
	}

	if len(out) > 1 {
		return nil, errors.New("There is more than 1 subscribe doc in database for this concept")
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
