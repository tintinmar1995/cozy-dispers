package query

import (
	"strconv"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/cozy/cozy-stack/pkg/dispers/metadata"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var PrefixerC = prefixer.ConductorPrefixer
var PrefixerT = prefixer.TargetPrefixer

type State int

const (
	Finished State = iota
	Waiting
	Running
	Failed
)

type AsyncType int

var AsyncTypes = []string{"AsyncAggregation"}

const (
	AsyncAggregation AsyncType = iota
	AsyncQueryTarget
)

type AsyncTask struct {
	AsyncID      string                `json:"_id,omitempty"`
	AsyncRev     string                `json:"_rev,omitempty"`
	AsyncType    AsyncType             `json:"async_type,omitempty"`
	QueryID      string                `json:"query_id"`
	TaskMetadata metadata.TaskMetadata `json:"task_metadata"`
	// Attributes used for AsyncAggregation
	IndexLayer int                    `json:"da_layer_id"`
	IndexDA    int                    `json:"da_id"`
	StateDA    State                  `json:"da_state,omitempty"`
	ResultDA   map[string]interface{} `json:"da_result,omitempty"`
	// Attributes used for AsyncQueryTarget or AsyncSendData
	NumberOfTargets int                      `json:"t_number_targets,omitempty"`
	Data            []map[string]interface{} `json:"t_data,omitempty"`
}

// ID returns the Doc ID
func (as *AsyncTask) ID() string {
	return as.AsyncID
}

// Rev returns the doc's version
func (as *AsyncTask) Rev() string {
	return as.AsyncRev
}

// DocType returns the DocType
func (as *AsyncTask) DocType() string {
	return "io.cozy.async"
}

// Clone copy a brand new version of the doc
func (as *AsyncTask) Clone() couchdb.Doc {
	cloned := *as
	return &cloned
}

// GetStateDA returns the state of DA
func (as *AsyncTask) GetStateDA() State {
	return as.StateDA
}

// SetID set the ID
func (as *AsyncTask) SetID(id string) {
	as.AsyncID = id
}

// SetRev set the version
func (as *AsyncTask) SetRev(rev string) {
	as.AsyncRev = rev
}

func (as *AsyncTask) SetFinished() error {
	// Doc found, set as finished
	as.StateDA = Finished
	return couchdb.UpdateDoc(PrefixerC, as)
}

func (as *AsyncTask) SetData(data ...map[string]interface{}) error {

	switch as.AsyncType {
	case AsyncAggregation:
		as.ResultDA = data[0]
		return couchdb.UpdateDoc(PrefixerC, as)
	case AsyncQueryTarget:
		as.Data = data
		return couchdb.UpdateDoc(PrefixerT, as)
	default:
		return errors.WrapErrors(errors.ErrAsyncTypeUnknown, "")
	}
}

func NewAsyncTask(queryid string, asyncType AsyncType, integers ...int) (AsyncTask, error) {

	switch asyncType {
	case AsyncAggregation:
		doc := AsyncTask{
			QueryID:    queryid,
			IndexLayer: integers[0],
			IndexDA:    integers[1],
			StateDA:    Running,
			AsyncType:  asyncType,
		}
		err := couchdb.CreateDoc(PrefixerC, &doc)
		return doc, err
	case AsyncQueryTarget:
		doc := AsyncTask{
			QueryID:         queryid,
			NumberOfTargets: integers[0],
			AsyncType:       asyncType,
		}
		err := couchdb.CreateDoc(PrefixerT, &doc)
		return doc, err
	default:
		return AsyncTask{}, errors.ErrAsyncTypeUnknown
	}
}

func IsAsyncTaskDAExisting(queryid string, indexLayer int, indexDA int) (bool, error) {

	// Retrieve docs that matches with ids
	var out []AsyncTask
	if err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async"); err != nil {
		return false, err
	}
	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("query_id", queryid), mango.Equal("da_layer_id", indexLayer), mango.Equal("da_id", indexDA))}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out); err != nil {
		return false, err
	}

	return !(len(out) == 0), nil
}

func RetrieveAsyncTaskDA(queryid string, indexLayer int, indexDA int) (AsyncTask, error) {

	// Retrieve docs that matches with ids
	var out []AsyncTask
	if err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async"); err != nil {
		return AsyncTask{}, err
	}
	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("query_id", queryid), mango.Equal("da_layer_id", indexLayer), mango.Equal("da_id", indexDA))}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out); err != nil {
		return AsyncTask{}, err
	}

	if len(out) == 0 {
		return AsyncTask{}, errors.ErrAsyncTaskNotFound
	}

	if len(out) > 1 {
		return AsyncTask{}, errors.ErrTooManyDoc
	}

	return out[0], nil
}

// FetchAsyncStateLayer returns the state of the layer
// The function has to be the fastest as possible
func FetchAsyncStateLayer(queryid string, indexLayer int, sizeLayer int) (State, error) {

	// Fetch everydoc that match queryid and indexLayer
	var out []AsyncTask
	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("query_id", queryid), mango.Equal("da_layer_id", indexLayer))}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out); err != nil {
		return Waiting, err
	}

	// No doc match the queryid & indexLayer, the layer has not been launched at all
	if len(out) == 0 {
		return Waiting, nil
	}

	// There is still some DA that havenot been launched
	if len(out) < sizeLayer {
		return Running, nil
	}

	// Check if every DA has finished, and has sent there results back
	for indexDA := 0; indexDA < sizeLayer; indexDA++ {
		if out[indexDA].StateDA != Finished {
			// We've found one DA that has not finished
			return Running, nil
		}
	}

	// Everything looks fine, the layer is officially finished
	return Finished, nil
}

func FetchAsyncDataDA(queryid string, indexLayer int, indexDA int) (map[string]interface{}, error) {

	var out []AsyncTask
	err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async")
	if err != nil {
		return nil, err
	}

	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("query_id", queryid), mango.Equal("da_layer_id", indexLayer), mango.Equal("da_id", indexDA))}
	err = couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, errors.ErrAsyncTaskNotFound
	}

	return out[0].ResultDA, nil
}

func DeleteAsyncDataT(queryid string) error {

	var tasks []AsyncTask
	if err := couchdb.EnsureDBExist(PrefixerT, "io.cozy.async"); err != nil {
		return err
	}

	req := &couchdb.FindRequest{Selector: mango.Equal("query_id", queryid)}
	if err := couchdb.FindDocs(PrefixerT, "io.cozy.async", req, &tasks); err != nil {
		return err
	}

	for _, task := range tasks {
		if err := couchdb.DeleteDoc(PrefixerT, &task); err != nil {
			return err
		}
	}

	return nil
}

func FetchAsyncDataT(queryid string) (map[string]interface{}, error) {

	var tasks []AsyncTask
	var out []map[string]interface{}
	if err := couchdb.EnsureDBExist(PrefixerT, "io.cozy.async"); err != nil {
		return nil, err
	}

	req := &couchdb.FindRequest{Selector: mango.Equal("query_id", queryid)}
	if err := couchdb.FindDocs(PrefixerT, "io.cozy.async", req, &tasks); err != nil {
		return nil, err
	}

	for _, task := range tasks {
		out = append(out, task.Data...)
	}

	return map[string]interface{}{"Data": out, "NumberOfTargets": len(tasks)}, nil
}

func FetchAsyncMetadata(queryid string) ([]metadata.TaskMetadata, error) {

	var out []AsyncTask
	req := &couchdb.FindRequest{Selector: mango.Equal("query_id", queryid)}
	if err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async"); err != nil {
		return nil, err
	}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out); err != nil {
		return nil, err
	}

	var task metadata.TaskMetadata
	meta := make([]metadata.TaskMetadata, len(out))
	for index, async := range out {
		task = async.TaskMetadata
		switch async.AsyncType {
		case AsyncAggregation:
			task.Name = AsyncTypes[async.AsyncType] + strconv.Itoa(async.IndexLayer) + "-" + strconv.Itoa(async.IndexDA)
		}
		meta[index] = task
	}

	return meta, nil
}
