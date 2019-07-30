package query

import (
	"errors"
	"strconv"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var PrefixerC = prefixer.ConductorPrefixer

type State int

const (
	Finished State = iota
	Waiting
	Running
	Failed
)

type AsyncType int

const (
	AsyncAggregation AsyncType = iota
)

type Async struct {
	AsyncID    string                 `json:"_id,omitempty"`
	AsyncRev   string                 `json:"_rev,omitempty"`
	QueryID    string                 `json:"queryid"`
	IndexLayer int                    `json:"layerid"`
	IndexDA    int                    `json:"daid"`
	Type       AsyncType              `json:"type"`
	StateDA    State                  `json:"state"`
	Data       map[string]interface{} `json:"data"`
}

// ID returns the Doc ID
func (as *Async) ID() string {
	return as.AsyncID
}

// Rev returns the doc's version
func (as *Async) Rev() string {
	return as.AsyncRev
}

// DocType returns the DocType
func (as *Async) DocType() string {
	return "io.cozy.async"
}

// Clone copy a brand new version of the doc
func (as *Async) Clone() couchdb.Doc {
	cloned := *as
	return &cloned
}

// SetID set the ID
func (as *Async) SetID(id string) {
	as.AsyncID = id
}

// SetRev set the version
func (as *Async) SetRev(rev string) {
	as.AsyncRev = rev
}

func NewAsyncTask(queryid string, indexLayer int, indexDA int, asyncType AsyncType) (Async, error) {
	doc := Async{
		QueryID:    queryid,
		IndexLayer: indexLayer,
		IndexDA:    indexDA,
		Type:       asyncType,
		StateDA:    Running,
	}

	err := couchdb.CreateDoc(PrefixerC, &doc)

	return doc, err
}

func SetAsyncTaskAsFinished(queryid string, indexLayer int, indexDA int) error {

	var out []Async

	err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async")
	if err != nil {
		return err
	}

	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("queryid", queryid), mango.Equal("layerid", indexLayer), mango.Equal("daid", indexDA))}
	err = couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("Async task not found")
	}
	if len(out) > 1 {
		return errors.New(queryid + " " + strconv.Itoa(indexLayer) + " " + strconv.Itoa(indexDA) + " Too many async task for this task")
	}

	out[0].StateDA = Finished
	if err := couchdb.UpdateDoc(PrefixerC, &out[0]); err != nil {
		return err
	}

	return nil
}

func SetData(queryid string, indexLayer int, indexDA int, data map[string]interface{}) error {

	var out []Async

	err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async")
	if err != nil {
		return err
	}

	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("queryid", queryid), mango.Equal("layerid", indexLayer), mango.Equal("daid", indexDA))}
	err = couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("Async task not found")
	}
	if len(out) > 1 {
		return errors.New("No many async task for this task")
	}

	out[0].Data = data
	if err := couchdb.UpdateDoc(PrefixerC, &out[0]); err != nil {
		return err
	}

	return nil
}

func FetchAsyncStateLayer(queryid string, indexLayer int, sizeLayer int) (State, error) {

	var out []Async
	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("queryid", queryid), mango.Equal("layerid", indexLayer))}
	if err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async"); err != nil {
		return Waiting, err
	}
	if err := couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out); err != nil {
		return Waiting, err
	}

	if len(out) == 0 {
		return Waiting, nil
	}

	if len(out) < sizeLayer {
		return Running, nil
	}

	for indexDA := 0; indexDA < sizeLayer; indexDA++ {
		if out[indexDA].StateDA != Finished {
			return Running, nil
		}
	}

	return Finished, nil
}

func FetchAsyncData(queryid string, indexLayer int, indexDA int) (map[string]interface{}, error) {

	var out []Async

	err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async")
	if err != nil {
		return nil, err
	}

	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("queryid", queryid), mango.Equal("layerid", indexLayer), mango.Equal("daid", indexDA))}
	err = couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, errors.New("No async doc for this job. Perhaps it's still running.")
	}

	return out[0].Data, nil
}
