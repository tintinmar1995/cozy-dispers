package query

import (
	"errors"

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
	}

	err := couchdb.CreateDoc(PrefixerC, &doc)

	return doc, err
}

func FetchAsyncState(queryid string, indexLayer int, indexDA int) (State, error) {

	var out []Async

	err := couchdb.EnsureDBExist(PrefixerC, "io.cozy.async")
	if err != nil {
		return Waiting, err
	}

	req := &couchdb.FindRequest{Selector: mango.And(mango.Equal("queryid", queryid), mango.Equal("layerid", indexLayer), mango.Equal("daid", indexDA))}
	err = couchdb.FindDocs(PrefixerC, "io.cozy.async", req, &out)
	if err != nil {
		return Waiting, err
	}

	if len(out) == 0 {
		return Waiting, nil
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
