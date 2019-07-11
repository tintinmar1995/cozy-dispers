package dispers

import (
	"errors"
	"net/url"
	"time"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/couchdb/mango"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

// Only the Conductor should save metadata
var prefixC = prefixer.ConductorPrefixer

type TaskMetadata struct {
	Start time.Time `json:"start,omitempty"`
	End   time.Time `json:"end,omitempty"`
	URL   url.URL   `json:"url,omitempty"`
	Error error     `json:"error,omitempty"`
}

// NewTaskMetadata returns a new TaskMetadata object
func NewTaskMetadata() (TaskMetadata, error) {
	now := time.Now()
	doc := TaskMetadata{
		Start: now,
	}

	return doc, nil
}

// Save the ExecutionMetadata in Conductor's database
func (t *TaskMetadata) EndTask(err error) {
	t.End = time.Now()
	t.Error = err
}

// ExecutionMetadata are written on the conductor's database. The querier can read those ExecutionMetadata to know his query's state
type ExecutionMetadata struct {
	ExecutionMetadataID  string                  `json:"_id,omitempty"`
	ExecutionMetadataRev string                  `json:"_rev,omitempty"`
	QueryID              string                  `json:"query,omitempty"`
	Start                time.Time               `json:"start,omitempty"`
	End                  time.Time               `json:"end,omitempty"`
	Process              string                  `json:"process,omitempty"`
	Host                 url.URL                 `json:"host,omitempty"`
	Tasks                map[string]TaskMetadata `json:"tasks,omitempty"`
}

func (t *ExecutionMetadata) ID() string {
	return t.ExecutionMetadataID
}

func (t *ExecutionMetadata) Rev() string {
	return t.ExecutionMetadataRev
}

func (t *ExecutionMetadata) DocType() string {
	return "io.cozy.execution.metadata"
}

func (t *ExecutionMetadata) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

func (t *ExecutionMetadata) SetID(id string) {
	t.ExecutionMetadataID = id
}

func (t *ExecutionMetadata) SetRev(rev string) {
	t.ExecutionMetadataRev = rev
}

// NewExecutionMetadata returns a new ExecutionMetadata object
func NewExecutionMetadata(process string, queryid string, host url.URL) (ExecutionMetadata, error) {
	now := time.Now()
	doc := ExecutionMetadata{
		Start:   now,
		Process: process,
		QueryID: queryid,
		Host:    host,
		Tasks:   make(map[string]TaskMetadata),
	}

	err := couchdb.CreateDoc(prefixC, &doc)

	return doc, err
}

// Save the ExecutionMetadata in Conductor's database
func (m *ExecutionMetadata) SetTask(name string, task TaskMetadata) error {
	if m.Tasks == nil {
		tasks := make(map[string]TaskMetadata)
		tasks[name] = task
		m.Tasks = tasks
	} else {
		m.Tasks[name] = task
	}
	return couchdb.UpdateDoc(prefixC, m)
}

// EndExecution the ExecutionMetadata in Conductor's database
func (m *ExecutionMetadata) EndExecution(err error) error {
	m.End = time.Now()
	return couchdb.UpdateDoc(prefixC, m)
}

// RetrieveExecutionMetadata get ExecutionMetadata from a ExecutionMetadata in CouchDB
func RetrieveExecutionMetadata(queryid string) (ExecutionMetadata, error) {

	var out []ExecutionMetadata

	err := couchdb.EnsureDBExist(prefixC, "io.cozy.execution.metadata")
	if err != nil {
		return ExecutionMetadata{}, err
	}

	req := &couchdb.FindRequest{Selector: mango.Equal("query", queryid)}
	err = couchdb.FindDocs(prefixC, "io.cozy.execution.metadata", req, &out)
	if err != nil {
		return ExecutionMetadata{}, err
	}

	if len(out) == 0 {
		return ExecutionMetadata{}, errors.New("No ExecutionMetadata for this query")
	}

	if len(out) > 1 {
		return ExecutionMetadata{}, errors.New("Too many ExecutionMetadata for this query")
	}

	return out[0], nil
}
