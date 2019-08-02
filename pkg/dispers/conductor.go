package enclave

import (
	"encoding/json"
	"errors"
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
	// hosts is a list of hosts where Cozy-DISPERS is running.
	// Those hosts can be called to play the role of CI/TF/T/DA
	hosts = []url.URL{
		url.URL{
			Scheme: "http",
			Host:   "localhost:8008",
		},
	}
	// PrefixerC is exported to easilly pass in dev-mode
	PrefixerC = prefixer.ConductorPrefixer
)

// QueryDoc saves every information about the query. QueryDoc are saved in the
// Conductor's database. Thanks to that, CheckPoints can be made, and the process
// can be followed by the querier.
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
	Results                   interface{}         `json:"results,omitempty"`
	TargetProfile             string              `json:"target_profile,omitempty"`
	Targets                   []string            `json:"targets,omitempty"`
	EncryptedConcepts         [][]byte            `json:"enc_concepts,omitempty"`
	EncryptedListsOfAddresses []byte              `json:"enc_instances,omitempty"`
	EncryptedLocalQuery       []byte              `json:"enc_localquery,omitempty"`
	EncryptedTargetProfile    []byte              `json:"enc_operation,omitempty"`
	EncryptedTargets          []byte              `json:"enc_addresses,omitempty"`
}

// ID returns the QueryID
func (q *QueryDoc) ID() string {
	return q.QueryID
}

// Rev returns the doc's version
func (q *QueryDoc) Rev() string {
	return q.QueryRev
}

// DocType returns the doctype
func (q *QueryDoc) DocType() string {
	return "io.cozy.query"
}

// Clone copy a brand new version of the doc
func (q *QueryDoc) Clone() couchdb.Doc {
	cloned := *q
	return &cloned
}

// SetID set the QueryID
func (q *QueryDoc) SetID(id string) {
	q.QueryID = id
}

// SetRev set the doc's version
func (q *QueryDoc) SetRev(rev string) {
	q.QueryRev = rev
}

// NewQuery returns a Query object to lead, resume, get info about the query.
func NewQuery(in *query.InputNewQuery) (*QueryDoc, error) {

	// Creating the QueryDoc that will be saved in the Conductor's database
	q := &QueryDoc{
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
	if err := couchdb.CreateDoc(PrefixerC, q); err != nil {
		return &QueryDoc{}, err
	}

	return q, nil
}

// NewQueryFetchingQueryDoc returns a QueryDoc object to resume the request
func NewQueryFetchingQueryDoc(queryid string) (*QueryDoc, error) {

	q := &QueryDoc{}
	err := couchdb.GetDoc(PrefixerC, "io.cozy.query", queryid, q)
	return q, err
}

// decryptConcept returns a list of hashed concepts from a list of encrypted concepts
// This function call another Cozy-DISPERS playing the role of Concept Indexor.
func (q *QueryDoc) decryptConcept() error {

	// Making the URL to call the other Cozy-DISPERS server
	// TODO : Delete IsEncrypted from Concept struct and add it to InputCI, use q.IsEncrypted instead
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/" + query.ConceptsToString(q.Concepts) + "/" + strconv.FormatBool(q.Concepts[0].IsEncrypted))
	err := ci.MakeRequest("GET", "", nil, nil)
	if err != nil {
		return err
	}
	// Read CI's answer, check the process as done, update QueryDoc
	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)
	q.CheckPoints["ci"] = true
	q.Concepts = outputCI.Hashes
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) fetchListsOfInstancesFromDB() error {

	encListsOfA := make(map[string][]byte)
	listsOfA := make(map[string][]string)

	// Retrieve the lists of addresses from Conductor's database
	// TODO : Pass to TF encrypted data (the script will be shorter)
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
			// This part will be removed when clearing Inputs and Outputs and deleting Non-Encrypted data
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

	// Check the process as done and update QueryDoc
	q.CheckPoints["fetch"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) selectTargets() error {

	// Make a request to Target Finder to retrieve the final list of targets
	inputTF := query.InputTF{
		IsEncrypted:               q.IsEncrypted,
		ListsOfAddresses:          q.ListsOfAddresses,
		TargetProfile:             q.TargetProfile,
		EncryptedListsOfAddresses: q.EncryptedListsOfAddresses,
		EncryptedTargetProfile:    q.EncryptedTargetProfile,
	}
	tf := network.NewExternalActor(network.RoleTF, network.ModeQuery)
	tf.DefineDispersActor("addresses")
	if err := tf.MakeRequest("POST", "", inputTF, nil); err != nil {
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

	// Pass the list of targets to anther Cozy-DISPERS as Target
	// Retrieve an array of encrypted data
	inputT := query.InputT{
		IsEncrypted:         q.IsEncrypted,
		LocalQuery:          q.LocalQuery,
		Targets:             q.Targets,
		EncryptedLocalQuery: q.EncryptedLocalQuery,
		EncryptedTargets:    q.EncryptedTargets,
	}
	t := network.NewExternalActor(network.RoleT, network.ModeQuery)
	t.DefineDispersActor("query")
	if err := t.MakeRequest("POST", "", inputT, nil); err != nil {
		return err
	}
	var outputT query.OutputT
	if err := json.Unmarshal(t.Out, &outputT); err != nil {
		return err
	}

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

// ShouldBeComputed is used by the Conductor to know when to pause/resume/stop the query
func (q *QueryDoc) ShouldBeComputed(indexLayer int) (bool, error) {

	// there is no layer number len(q.Layers)
	// Conductor should returns false
	if indexLayer == len(q.Layers) {
		return false, nil
	}

	// ShouldBeComputed returns false if indexLayer finished or Running
	stateLayer, err := query.FetchAsyncStateLayer(q.QueryID, indexLayer, q.Layers[indexLayer].Size)
	if err != nil {
		return false, err
	}
	// If this is the first Layer and it is waiting
	// Conductor should compute the layer
	if stateLayer == query.Waiting && indexLayer == 0 {
		return true, nil
	}
	if stateLayer == query.Finished || stateLayer == query.Running {
		return false, nil
	}

	// ShouldBeComputed returns false if indexLayer-1 is not finished
	// Conductor should wait that indexLayer-1 is finished before computing indexLayer
	statePreviousLayer, err := query.FetchAsyncStateLayer(q.QueryID, indexLayer-1, q.Layers[indexLayer-1].Size)
	if err != nil {
		return false, err
	}
	if statePreviousLayer == query.Waiting || statePreviousLayer == query.Running {
		return false, nil
	}

	// The previous layer is finished, the layer has not been begun...
	// Conductor can compute indexLayer
	return true, nil
}

func (q *QueryDoc) aggregateLayer(indexLayer int, layer *query.LayerDA) error {

	var data []map[string]interface{}

	// if it is the first layer, data should should be retrieved from Target's result
	// if not, data should be fetched from async tasks' database.
	if indexLayer != 0 {
		for indexDA := 0; indexDA < q.Layers[indexLayer-1].Size; indexDA++ {
			rowData, err := query.FetchAsyncData(q.ID(), indexLayer-1, indexDA)
			if err != nil {
				return err
			}
			data = append(data, rowData)
		}
		if len(data) == 0 {
			return errors.New("No data fetched from previous layer")
		}
	} else {
		data = layer.Data
	}

	// Shuffle Data to reduce bias
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	// Distribute data in folds. Each DA will have one fold.
	seps := make([]int, layer.Size+1)
	seps[0] = 0
	if layer.Size > 1 {
		if len(data)%layer.Size != 0 {
			seps[len(seps)-1] = len(data) - 1
		}
		for indexSep := 1; indexSep < len(seps)-1; indexSep++ {
			seps[indexSep] = (len(data) / layer.Size) * indexSep
		}
	} else {
		seps[1] = len(data)
	}

	// Create InputDA for the layer
	inputDA := query.InputDA{
		Job:          layer.AggregationFunctions,
		EncryptedJob: layer.EncryptedAggregateFunctions,
		QueryID:      q.ID(),
	}

	// Set Conductor's URL
	// TODO : Find a prettier solution for localhost issue
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if hostname == "martin-perso" {
		hostname = "localhost"
	}
	inputDA.ConductorURL = url.URL{
		Scheme: "http",
		Host:   hostname + ":" + strconv.Itoa(config.GetConfig().Port),
	}

	for indexDA := 0; indexDA < layer.Size; indexDA++ {
		// Fit inputDA for each DA (data, id, ...)
		inputDA.Data = data[seps[indexDA]:seps[indexDA+1]]
		inputDA.AggregationID = [2]int{indexLayer, indexDA}

		// make the request and unmarshal answer
		// check one last time that DA hasnot been launch to prevent conflict
		state, err := query.FetchAsyncStateDA(q.ID(), indexLayer, indexDA)
		if err != nil {
			return err
		}
		if state == query.Waiting {
			query.NewAsyncTask(q.ID(), indexLayer, indexDA, query.AsyncAggregation)
			da := network.NewExternalActor(network.RoleDA, network.ModeQuery)
			da.DefineDispersActor("aggregation")
			if err := da.MakeRequest("POST", "", inputDA, nil); err != nil {
				return err
			}
			var out query.OutputDA
			if err := json.Unmarshal(da.Out, &out); err != nil {
				return err
			}
		} else {
			// The job has already been launch, this means that another worker has taken the job
			// there is nothing to do
			// This end of the process prevent conflict in couchdb
			return nil
		}
	}

	// async tasks is now running
	if err := couchdb.UpdateDoc(PrefixerC, q); err != nil {
		return err
	}
	return q.TryToEndQuery()
}

// Lead is the most general method. It will use the 5 previous methods to work.
// Lead can also be used to resume a query thanks to checkpoints.
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
			layerShouldBeComputed, err := q.ShouldBeComputed(indexLayer)
			if err != nil {
				return err
			}
			if layerShouldBeComputed {
				if err := q.aggregateLayer(indexLayer, &(q.Layers[indexLayer])); err != nil {
					return err
				}
				// Stop the process and wait for DAs' answers to resume
				return nil
			}
		}
	}

	return nil
}

func (q *QueryDoc) TryToEndQuery() error {

	// check if query is finished
	isQueryFinished := true
	for indexLayer, layer := range q.Layers {
		state, err := query.FetchAsyncStateLayer(q.ID(), indexLayer, layer.Size)
		if err != nil {
			return err
		}
		if state != query.Finished {
			isQueryFinished = false
		}
	}

	if isQueryFinished {

		// get results
		res, err := query.FetchAsyncData(q.ID(), len(q.Layers)-1, 0)
		if err != nil {
			return err
		}
		if res == nil {
			return errors.New("Results are nil")
		}
		q.Results = res
		// mark checkpoint
		q.CheckPoints["da"] = true

		return couchdb.UpdateDoc(PrefixerC, q)
	}

	return nil
}

// RetrieveSubscribeDoc is used to get a Subscribe doc from the Conductor's database.
// It returns either an empty array of SubscribeDoc or an array of length 1
// It returns an error if there is more than 1 subscribe doc.
func RetrieveSubscribeDoc(hash []byte) ([]subscribe.SubscribeDoc, error) {

	var out []subscribe.SubscribeDoc
	req := &couchdb.FindRequest{Selector: mango.Equal("hash", hash)}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.instances", req, &out); err != nil {
		return nil, err
	}

	if len(out) > 1 {
		return nil, errors.New("There is more than 1 subscribe doc in database for this concept")
	}

	return out, nil
}

// CreateConceptInConductorDB is used to add a concept to Cozy-DISPERS
func CreateConceptInConductorDB(in *query.InputCI) error {

	// try to create concept
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept")
	errPost := ci.MakeRequest("POST", "", *in, nil)

	// try to get concept
	path := query.ConceptsToString(in.Concepts) + "/" + strconv.FormatBool(in.Concepts[0].IsEncrypted)
	ci.DefineDispersActor("concept/" + path)
	err := ci.MakeRequest("GET", "", nil, nil)

	// if error, returns the first error that occurred between POST et Get
	if err != nil {
		if errPost != nil {
			return errPost
		}
		return err
	}

	// If success, get CI's result
	var out query.OutputCI
	err = json.Unmarshal(ci.Out, &out)
	if err != nil {
		return err
	}

	// Create a SubscribeDoc for each concept
	for _, concept := range out.Hashes {
		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return nil
		}

		// Check if Concept is unexistant
		if len(s) < 1 {
			listOfInsts := []query.Instance{}
			marshaledListOfInsts, err := json.Marshal(listOfInsts)
			if err != nil {
				return nil
			}
			sub := subscribe.SubscribeDoc{
				Hash:               concept.Hash,
				EncryptedInstances: marshaledListOfInsts,
			}
			if err := couchdb.CreateDoc(PrefixerC, &sub); err != nil {
				return err
			}
		} else {
			// TODO : Delete successfully created SubscribeDoc to finish with a statut quo
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
	if err := ci.MakeRequest("GET", "", nil, nil); err != nil {
		return err
	}
	var outputCI query.OutputCI
	if err := json.Unmarshal(ci.Out, &outputCI); err != nil {
		return err
	}

	var docs []subscribe.SubscribeDoc
	var err error
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

		// Ask Target to add instance by following TF's output
		t := network.NewExternalActor(network.RoleT, network.ModeSubscribe)
		t.DefineDispersActor("insert")
		if err := t.MakeRequest("POST", "", nil, tf.Out); err != nil {
			return err
		}

		// Ask Target Finder to Encrypt by following T's output
		tf = network.NewExternalActor(network.RoleTF, network.ModeSubscribe)
		tf.DefineDispersActor("encrypt")
		if err := tf.MakeRequest("POST", "", nil, t.Out); err != nil {
			return err
		}
		if err := json.Unmarshal(tf.Out, &outEnc); err != nil {
			return err
		}

		// Update subscribe doc
		doc.EncryptedInstances = outEnc.EncryptedInstances
		if err := couchdb.UpdateDoc(PrefixerC, &doc); err != nil {
			return err
		}
	}

	return nil
}
