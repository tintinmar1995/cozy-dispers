package enclave

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cozy/cozy-stack/pkg/config/config"
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
	PrefixerC = prefixer.ConductorPrefixer
)

// QueryDoc saves every information about the query. QueryDoc are saved in the
// Conductor's database. Thanks to that, CheckPoints can be made, and a request
// can be followed
type QueryDoc struct {
	QueryID                   string              `json:"_id,omitempty"`
	QueryRev                  string              `json:"_rev,omitempty"`
	IsEncrypted               bool                `json:"encrypted,omitempty"`
	CheckPoints               map[string]bool     `json:"checkpoints,omitempty"`
	Concepts                  []query.Concept     `json:"concepts,omitempty"`
	DomainQuerier             string              `json:"domain,omitempty"`
	ListsOfAddresses          map[string][]string `json:"lists_of_instances,omitempty"`
	LocalQuery                query.LocalQuery    `json:"localquery,omitempty"`
	Layers                    []query.LayerDA     `json:"layers,omitempty"`
	NumberActors              map[string]int      `json:"nb_actors,omitempty"`
	PseudoConcepts            map[string]string   `json:"pseudo_concepts,omitempty"`
	TargetProfile             query.OperationTree `json:"target_profile,omitempty"`
	Targets                   []string            `json:"targets,omitempty"`
	EncryptedConcepts         [][]byte            `json:"enc_concepts,omitempty"`
	EncryptedListsOfAddresses []byte              `json:"enc_instances,omitempty"`
	EncryptedLocalQuery       []byte              `json:"enc_localquery,omitempty"`
	EncryptedTargetProfile    []byte              `json:"enc_operation,omitempty"`
	EncryptedTargets          []byte              `json:"enc_addresses,omitempty"`
}

// ID returns the Doc ID
func (t *QueryDoc) ID() string {
	return t.QueryID
}

// Rev returns the doc's version
func (t *QueryDoc) Rev() string {
	return t.QueryRev
}

// DocType returns the DocType
func (t *QueryDoc) DocType() string {
	return "io.cozy.query"
}

// Clone copy a brand new version of the doc
func (t *QueryDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

// SetID set the ID
func (t *QueryDoc) SetID(id string) {
	t.QueryID = id
}

// SetRev set the version
func (t *QueryDoc) SetRev(rev string) {
	t.QueryRev = rev
}

// NewQuery returns a Conductor object to lead the request
func NewQuery(in *query.InputNewQuery) (*QueryDoc, error) {

	if in.NumberActors == nil {
		return &QueryDoc{}, errors.New("Number of network.ExternalActors should be defined")
	}

	// Creating the QueryDoc that will be saved in the Conductor's database
	retour := &QueryDoc{
		CheckPoints:            make(map[string]bool),
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
	if err := couchdb.CreateDoc(PrefixerC, retour); err != nil {
		return &QueryDoc{}, err
	}

	return retour, nil
}

// NewQueryFetchingQueryDoc returns a Conductor object to lead the request
func NewQueryFetchingQueryDoc(queryid string) (*QueryDoc, error) {

	queryDoc := &QueryDoc{}
	err := couchdb.GetDoc(PrefixerC, "io.cozy.query", queryid, queryDoc)
	if err != nil {
		return &QueryDoc{}, err
	}

	return queryDoc, nil
}

// decryptConcept returns a list of hashed concepts from a list of encrypted concepts
func (q *QueryDoc) decryptConcept() error {

	job := "concept/"
	listOfConcepts := []string{}
	for _, concept := range q.Concepts {
		if concept.IsEncrypted {
			listOfConcepts = append(listOfConcepts, string(concept.EncryptedConcept))
		} else {
			listOfConcepts = append(listOfConcepts, concept.Concept)
		}
	}
	job = job + strings.Join(listOfConcepts, ":")
	if q.IsEncrypted {
		job = job + "/true"
	} else {
		job = job + "/false"
	}

	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor(job)
	err := ci.MakeRequest("GET", "", nil, nil)
	if err != nil {
		return err
	}

	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)
	q.CheckPoints["ci"] = true
	q.Concepts = outputCI.Hashes
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) fetchListsOfInstancesFromDB() error {

	encListsOfA := make(map[string][]byte)
	listsOfA := make(map[string][]string)

	for _, concept := range q.Concepts {

		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return err
		}

		if len(s) == 0 {
			return errors.New("Cannot find SubscribeDoc associated to hash : " + string(concept.Hash))
		}

		if q.IsEncrypted {
			encListsOfA[q.PseudoConcepts[string(concept.EncryptedConcept)]] = s[0].EncryptedInstances
			res, _ := json.Marshal(encListsOfA)
			q.EncryptedListsOfAddresses = res

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
			listsOfA[q.PseudoConcepts[concept.Concept]] = instsStr
		}
		q.ListsOfAddresses = listsOfA
	}

	q.CheckPoints["fetch"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) selectTargets() error {

	inputTF := query.InputTF{
		IsEncrypted:               q.IsEncrypted,
		ListsOfAddresses:          q.ListsOfAddresses,
		TargetProfile:             q.TargetProfile,
		EncryptedListsOfAddresses: q.EncryptedListsOfAddresses,
		EncryptedTargetProfile:    q.EncryptedTargetProfile,
	}

	tf := network.NewExternalActor(network.RoleTF, network.ModeQuery)
	tf.DefineDispersActor("addresses")
	err := tf.MakeRequest("POST", "", inputTF, nil)
	if err != nil {
		return err
	}

	var outputTF query.OutputTF
	json.Unmarshal(tf.Out, &outputTF)
	q.EncryptedTargets = outputTF.EncryptedListOfAddresses
	q.Targets = outputTF.ListOfAddresses
	q.CheckPoints["tf"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) makeLocalQuery() error {

	inputT := query.InputT{
		IsEncrypted:         q.IsEncrypted,
		LocalQuery:          q.LocalQuery,
		Targets:             q.Targets,
		EncryptedLocalQuery: q.EncryptedLocalQuery,
		EncryptedTargets:    q.EncryptedTargets,
	}

	t := network.NewExternalActor(network.RoleT, network.ModeQuery)
	t.DefineDispersActor("query")
	err := t.MakeRequest("POST", "", inputT, nil)

	var outputT query.OutputT
	err = json.Unmarshal(t.Out, &outputT)
	if err != nil {
		return err
	}

	fmt.Println("Conductor", t.Outstr)

	if len(q.Layers) == 0 {
		return errors.New("Query should have at least one aggregation array")
	}
	if len(outputT.Data) == 0 {
		return errors.New("No data to query on")
	}
	q.Layers[0].Data = outputT.Data
	q.CheckPoints["t"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) shouldBeComputed(indexLayer int) bool {

	// shouldBeComputed returns false if indexLayer finished
	isLayerFinished := true
	for _, stateDA := range q.Layers[indexLayer].State {
		if stateDA != query.Finished {
			isLayerFinished = false
		}
	}
	if isLayerFinished {
		return false
	}

	if indexLayer == 0 {
		return true
	}

	// shouldBeComputed returns false if indexLayer-1 is not finished
	isPreviousLayerFinished := true
	for _, stateDA := range q.Layers[indexLayer-1].State {
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
	for _, stateDA := range q.Layers[indexLayer-1].State {
		if stateDA != query.Waiting {
			isLayerWaiting = false
		}
	}
	return isLayerWaiting
}

func (q *QueryDoc) aggregateLayer(indexLayer int, layer *query.LayerDA) error {

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

		inputDA.Data = layer.Data[seps[indexDA]:seps[indexDA+1]]
		inputDA.QueryID = q.ID()
		inputDA.AggregationID = [2]int{indexLayer, indexDA}
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}

		// TODO : Find a prettier solution for this issue
		if hostname == "martin-perso" {
			hostname = "localhost"
		}

		inputDA.ConductorURL = url.URL{
			Scheme: "http",
			Host:   hostname + ":" + strconv.Itoa(config.GetConfig().Port),
		}

		da := network.NewExternalActor(network.RoleDA, network.ModeQuery)
		da.DefineDispersActor("aggregation")
		err = da.MakeRequest("POST", "", inputDA, nil)
		if err != nil {
			return err
		}

		var out query.OutputDA
		err = json.Unmarshal(da.Out, &out)
		if err != nil {
			return err
		}

		// async task is now running
		layer.State[strconv.Itoa(indexDA)] = query.Running
		err = couchdb.UpdateDoc(PrefixerC, q)
		if err != nil {
			return err
		}
	}

	return nil
}

// Lead is the most general method. This is the only one used in CMD and Web's files. It will use the 5 previous methods to work
func (q *QueryDoc) Lead() error {

	if q.CheckPoints["ci"] != true {
		if err := q.decryptConcept(); err != nil {
			return err
		}
	}

	if q.CheckPoints["fetch"] != true {
		if err := q.fetchListsOfInstancesFromDB(); err != nil {
			return err
		}
	}

	if q.CheckPoints["tf"] != true {
		if err := q.selectTargets(); err != nil {
			return err
		}
	}

	if q.CheckPoints["t"] != true {
		if err := q.makeLocalQuery(); err != nil {
			return err
		}
	}

	if q.CheckPoints["da"] != true {
		for indexLayer := range q.Layers {
			if q.shouldBeComputed(indexLayer) {
				if err := q.aggregateLayer(indexLayer, &(q.Layers[indexLayer])); err != nil {
					return err
				}
				// Stop the process and wait for DA's answer to resume
				return nil
			}
		}
	}

	return couchdb.UpdateDoc(PrefixerC, q)
}

// RetrieveSubscribeDoc is used to get a Subscribe doc from the Conductor's database.
// It returns an error if there is more than 1 subscribe doc
func RetrieveSubscribeDoc(hash []byte) ([]subscribe.SubscribeDoc, error) {

	var out []subscribe.SubscribeDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("hash", hash)}
	err := couchdb.FindDocs(PrefixerC, "io.cozy.instances", req, &out)
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

	// try to create concept with route POST
	// try to get hash with CI's route GET
	// if error, returns the first error that occurred between POST et Get
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept")
	errPost := ci.MakeRequest("POST", "", *in, nil)

	path := query.ConceptsToString(in.Concepts) + "/" + strconv.FormatBool(in.Concepts[0].IsEncrypted)
	ci.DefineDispersActor("concept/" + path)
	err := ci.MakeRequest("GET", "", nil, nil)
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
			listOfInsts := []query.Instance{}
			marshaledListOfInsts, _ := json.Marshal(listOfInsts)
			sub := subscribe.SubscribeDoc{
				Hash:               concept.Hash,
				EncryptedInstances: marshaledListOfInsts,
			}
			if err := couchdb.CreateDoc(PrefixerC, &sub); err != nil {
				return err
			}
		} else {
			return errors.New("This concept already exists in Conductor's database")
		}
	}

	return nil
}

// Subscribe leads the subscription process
func Subscribe(in *subscribe.InputConductor) error {

	// Get Concepts' hash
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/" + strings.Join(in.Concepts, ":") + "/" + strconv.FormatBool(in.IsEncrypted))
	err := ci.MakeRequest("GET", "", nil, nil)
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
		tf := network.NewExternalActor(network.RoleTF, network.ModeSubscribe)
		tf.DefineDispersActor("decrypt")
		err = tf.MakeRequest("POST", "", subscribe.InputDecrypt{
			IsEncrypted:        in.IsEncrypted,
			EncryptedInstances: doc.EncryptedInstances,
			EncryptedInstance:  in.EncryptedInstance,
		}, nil)
		if err != nil {
			return err
		}

		// Ask Target to add instance
		t := network.NewExternalActor(network.RoleT, network.ModeSubscribe)
		t.DefineDispersActor("insert")
		err = t.MakeRequest("POST", "", nil, tf.Out)
		if err != nil {
			return err
		}

		// Ask Target Finder to Encrypt
		tf = network.NewExternalActor(network.RoleTF, network.ModeSubscribe)
		tf.DefineDispersActor("encrypt")
		err = tf.MakeRequest("POST", "", nil, t.Out)
		if err != nil {
			return err
		}
		err = json.Unmarshal(tf.Out, &outEnc)
		if err != nil {
			return err
		}

		doc.EncryptedInstances = outEnc.EncryptedInstances
		couchdb.UpdateDoc(PrefixerC, &doc)
	}

	return nil
}
