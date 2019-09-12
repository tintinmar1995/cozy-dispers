package enclave

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/cozy/cozy-stack/pkg/dispers/metadata"
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

	executionMetadata *metadata.ExecutionMetadata
	cursor            = 0
)

// QueryDoc saves every information about the query. QueryDoc are saved in the
// Conductor's database. Thanks to that, CheckPoints can be made, and the process
// can be followed by the querier.
type QueryDoc struct {
	QueryID                   string            `json:"_id,omitempty"`
	QueryRev                  string            `json:"_rev,omitempty"`
	IsEncrypted               bool              `json:"encrypted,omitempty"`
	CheckPoints               map[string]bool   `json:"checkpoints,omitempty"`
	Layers                    []query.LayerDA   `json:"layers,omitempty"`
	PseudoConcepts            map[string]string `json:"pseudo_concepts,omitempty"`
	Results                   interface{}       `json:"results,omitempty"`
	EncryptedConcepts         []query.Concept   `json:"concepts,omitempty"`
	EncryptedListsOfAddresses map[string][]byte `json:"enc_instances,omitempty"`
	EncryptedLocalQuery       []byte            `json:"enc_localquery,omitempty"`
	EncryptedTargetProfile    []byte            `json:"enc_operation,omitempty"`
	EncryptedTargets          []byte            `json:"enc_addresses,omitempty"`
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

	var q *QueryDoc

	if in.IsEncrypted {
		// Creating the QueryDoc that will be saved in the Conductor's database
		q = &QueryDoc{
			CheckPoints:            make(map[string]bool),
			IsEncrypted:            in.IsEncrypted,
			Layers:                 in.LayersDA,
			PseudoConcepts:         in.PseudoConcepts,
			EncryptedConcepts:      in.EncryptedConcepts,
			EncryptedLocalQuery:    in.EncryptedLocalQuery,
			EncryptedTargetProfile: in.EncryptedTargetProfile,
		}
	} else {

		encryptedConcepts := []query.Concept{}
		for _, concept := range in.Concepts {
			encryptedConcepts = append(encryptedConcepts, query.Concept{EncryptedConcept: []byte(concept)})
		}

		encryptedLocalQuery, err := json.Marshal(in.LocalQuery)
		if err != nil {
			return q, err
		}

		for index, layer := range in.LayersDA {
			encryptedJobs, err := json.Marshal(layer.Jobs)
			if err != nil {
				return q, err
			}
			in.LayersDA[index].EncryptedJobs = encryptedJobs
		}

		q = &QueryDoc{
			CheckPoints:            make(map[string]bool),
			IsEncrypted:            in.IsEncrypted,
			Layers:                 in.LayersDA,
			PseudoConcepts:         in.PseudoConcepts,
			EncryptedConcepts:      encryptedConcepts,
			EncryptedLocalQuery:    encryptedLocalQuery,
			EncryptedTargetProfile: []byte(in.TargetProfile),
		}
	}

	if err := couchdb.CreateDoc(PrefixerC, q); err != nil {
		return &QueryDoc{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return q, errors.WrapErrors(errors.ErrHostnameConductor, "")
	}
	urlCond, err := url.Parse(hostname)
	if err != nil {
		return q, errors.WrapErrors(errors.ErrHostnameConductor, "")
	}
	meta, err := metadata.NewExecutionMetadata("Query", q.ID(), *urlCond)
	if err != nil {
		return q, errors.WrapErrors(errors.ErrNewExecutionMetadata, "")
	}
	executionMetadata = &meta
	return q, nil
}

// NewQueryFetchingQueryDoc returns a QueryDoc object to resume the request
func NewQueryFetchingQueryDoc(queryid string, indexLayer int) (*QueryDoc, error) {

	cursor = indexLayer

	q := &QueryDoc{}
	err := couchdb.GetDoc(PrefixerC, "io.cozy.query", queryid, q)
	if err != nil {
		return q, errors.WrapErrors(errors.ErrRetrievingQueryDoc, "")
	}

	executionMetadata, err = metadata.RetrieveExecutionMetadata(queryid)
	if err != nil {
		return q, errors.WrapErrors(errors.ErrRetrievingExecutionMetadata, "")
	}

	return q, err
}

// decryptConcept returns a list of hashed concepts from a list of encrypted concepts
// This function call another Cozy-DISPERS playing the role of Concept Indexor.
func (q *QueryDoc) decryptConcept() error {

	// Making the URL to call the other Cozy-DISPERS server
	task := metadata.NewTaskMetadata()
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept/" + query.ConceptsToString(q.EncryptedConcepts) + "/" + strconv.FormatBool(q.IsEncrypted))
	err := ci.MakeRequest("GET", "", nil, nil)
	if err != nil {
		return executionMetadata.HandleError("DecryptConcept", task, err)
	}
	executionMetadata.HandleError("DecryptConcept", task, nil)
	// Read CI's answer, check the process as done, update QueryDoc
	var outputCI query.OutputCI
	json.Unmarshal(ci.Out, &outputCI)
	q.CheckPoints["ci"] = true
	q.EncryptedConcepts = outputCI.Hashes
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) fetchListsOfInstancesFromDB() error {

	task := metadata.NewTaskMetadata()
	encListsOfA := make(map[string][]byte)

	// Retrieve the lists of addresses from Conductor's database
	for _, concept := range q.EncryptedConcepts {

		s, err := RetrieveSubscribeDoc(concept.Hash)
		if err != nil {
			return executionMetadata.HandleError("FetchListsOfAddresses", task, err)
		}

		if len(s) == 0 {
			return executionMetadata.HandleError("FetchListsOfAddresses", task, errors.ErrSubscribeDocNotFound)
		}

		executionMetadata.HandleError("FetchListsOfAddresses", task, nil)
		encListsOfA[q.PseudoConcepts[string(concept.EncryptedConcept)]] = s[0].EncryptedInstances
		q.EncryptedListsOfAddresses = encListsOfA

	}

	// Check the process as done and update QueryDoc
	q.CheckPoints["fetch"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) selectTargets() error {

	task := metadata.NewTaskMetadata()

	// Make a request to Target Finder to retrieve the final list of targets
	inputTF := query.InputTF{
		IsEncrypted:               q.IsEncrypted,
		EncryptedListsOfAddresses: q.EncryptedListsOfAddresses,
		EncryptedTargetProfile:    q.EncryptedTargetProfile,
		TaskMetadata:              task,
	}
	tf := network.NewExternalActor(network.RoleTF, network.ModeQuery)
	tf.DefineDispersActor("addresses")
	if err := tf.MakeRequest("POST", "", inputTF, nil); err != nil {
		return executionMetadata.HandleError("SelectTargets", task, err)
	}
	var outputTF query.OutputTF
	if err := json.Unmarshal(tf.Out, &outputTF); err != nil {
		return executionMetadata.HandleError("SelectTargets", task, err)
	}
	executionMetadata.HandleError("SelectTargets", outputTF.TaskMetadata, nil)
	q.EncryptedTargets = outputTF.EncryptedTargets
	q.CheckPoints["tf"] = true
	return couchdb.UpdateDoc(PrefixerC, q)
}

func (q *QueryDoc) makeLocalQuery() error {

	task := metadata.NewTaskMetadata()

	// Set Conductor's URL
	// TODO : Find a prettier solution for localhost issue
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if hostname == "martin-perso" {
		hostname = "localhost"
	}

	// Pass the list of targets to anther Cozy-DISPERS as Target
	// Retrieve an array of encrypted data
	inputT := query.InputT{
		IsEncrypted:         q.IsEncrypted,
		EncryptedLocalQuery: q.EncryptedLocalQuery,
		EncryptedTargets:    q.EncryptedTargets,
		TaskMetadata:        task,
		QueryID:             q.QueryID,
		ConductorURL: url.URL{
			Scheme: "http",
			Host:   hostname + ":" + strconv.Itoa(config.GetConfig().Port),
		},
	}
	t := network.NewExternalActor(network.RoleT, network.ModeQuery)
	t.DefineDispersActor("query")
	if err := t.MakeRequest("POST", "", inputT, nil); err != nil {
		return executionMetadata.HandleError("LocalQuery", task, err)
	}
	var outputT query.OutputT
	if err := json.Unmarshal(t.Out, &outputT); err != nil {
		return executionMetadata.HandleError("LocalQuery", task, err)
	}

	executionMetadata.HandleError("LocalQuery", outputT.TaskMetadata, nil)
	// We just launched Async tasks, to avoid conflict, we can't modify the QueryDoc !
	return nil
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
	if indexLayer == 0 && stateLayer == query.Waiting {
		return true, nil
	}
	if stateLayer == query.Finished || stateLayer == query.Running {
		return false, nil
	}

	// indexLayer is bigger than 0. indexLayer is Waiting.
	// We need to check indexLayer-1 is finished
	// ShouldBeComputed returns false if indexLayer-1 is not finished
	// Conductor should wait that indexLayer-1 is finished before computing indexLayer
	statePreviousLayer, err := query.FetchAsyncStateLayer(q.QueryID, indexLayer-1, q.Layers[indexLayer-1].Size)
	if err != nil {
		return false, err
	}
	if statePreviousLayer == query.Finished {
		return true, nil
	}

	// The previous layer is not finished
	return false, nil
}

func (q *QueryDoc) aggregateLayer(indexLayer int, layer *query.LayerDA) error {

	// if it is the first layer, data should be retrieved from Target's result
	// if not, data should be fetched from async tasks' database.
	var data []map[string]interface{}
	if indexLayer == 0 {
		data = layer.Data
		if len(data) < 50 {
			return errors.WrapErrors(errors.ErrNotEnoughDataToComputeQuery, "")
		}
	} else {
		for indexDA := 0; indexDA < q.Layers[indexLayer-1].Size; indexDA++ {
			rowData, err := query.FetchAsyncDataDA(q.ID(), indexLayer-1, indexDA)
			if err != nil {
				return err
			}
			data = append(data, rowData)
		}
	}

	if len(data) == 0 {
		return errors.WrapErrors(errors.ErrNotEnoughDataToComputeQuery, "")
	}

	// Shuffle Data to reduce bias
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	// Distribute data in folds. Each DA will have one fold.
	seps := make([]int, layer.Size+1)
	if layer.Size > 1 {
		if len(data)%layer.Size != 0 {
			seps[len(seps)-1] = len(data)
		}
		for indexSep := 1; indexSep < len(seps)-1; indexSep++ {
			seps[indexSep] = (len(data) / layer.Size) * indexSep
		}
	} else {
		seps[1] = len(data)
	}

	// Create InputDA for the layer
	inputDA := query.InputDA{
		EncryptedJobs: layer.EncryptedJobs,
		QueryID:       q.ID(),
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
		encData, err := json.Marshal(data[seps[indexDA]:seps[indexDA+1]])
		if err != nil {
			return err
		}
		inputDA.EncryptedData = encData
		inputDA.AggregationID = [2]int{indexLayer, indexDA}
		inputDA.TaskMetadata = metadata.NewTaskMetadata()
		// make the request and unmarshal answer
		// check one last time that DA hasnot been launched to prevent conflict
		isExisting, err := query.IsAsyncTaskDAExisting(q.ID(), indexLayer, indexDA)
		if err != nil {
			return err
		}
		if !isExisting {
			query.NewAsyncTask(q.ID(), query.AsyncAggregation, indexLayer, indexDA)
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
	return couchdb.UpdateDoc(PrefixerC, q)
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
		return nil
	}

	if q.CheckPoints["da"] != true {
		for indexLayer := cursor; indexLayer < len(q.Layers); indexLayer++ {
			layerShouldBeComputed, err := q.ShouldBeComputed(indexLayer)
			if err != nil {
				return err
			}
			if layerShouldBeComputed {
				task := metadata.NewTaskMetadata()
				if err := q.aggregateLayer(indexLayer, &(q.Layers[indexLayer])); err != nil {
					return executionMetadata.HandleError("LaunchLayer"+strconv.Itoa(indexLayer), task, err)
				}
				// Stop the process and wait for DAs' answers to resume
				return executionMetadata.HandleError("LaunchLayer"+strconv.Itoa(indexLayer), task, nil)
			}
		}
	}

	return q.TryToEndQuery()
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
		q, err := NewQueryFetchingQueryDoc(q.ID(), 0)
		if err != nil {
			return err
		}
		// get results
		res, err := query.FetchAsyncDataDA(q.ID(), len(q.Layers)-1, 0)
		if err != nil {
			return err
		}
		q.Results = res
		// mark checkpoint
		q.CheckPoints["da"] = true
		executionMetadata.EndExecution(nil)
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
		return nil, errors.WrapErrors(errors.ErrTooManyDoc, "")
	}

	return out, nil
}

// CreateConceptInConductorDB is used to add a concept to Cozy-DISPERS
func CreateConceptInConductorDB(in *query.InputCI) error {

	if !in.IsEncrypted {
		in.EncryptedConcepts = []query.Concept{}
		for _, concept := range in.Concepts {
			in.EncryptedConcepts = append(in.EncryptedConcepts, query.Concept{EncryptedConcept: []byte(concept)})
		}
	}

	// try to create concept
	ci := network.NewExternalActor(network.RoleCI, network.ModeQuery)
	ci.DefineDispersActor("concept")
	errPost := ci.MakeRequest("POST", "", *in, nil)

	// try to get concept
	path := query.ConceptsToString(in.EncryptedConcepts) + "/" + strconv.FormatBool(in.IsEncrypted)
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
			return err
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
			return errors.WrapErrors(errors.ErrConceptAlreadyInConductorDB, "")
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
			return errors.WrapErrors(errors.ErrSubscribeDocNotFound, "")
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
