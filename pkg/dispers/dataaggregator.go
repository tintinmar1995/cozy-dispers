package enclave

import (
	"encoding/json"
	"errors"

	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/dispers/aggregations"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var prefixerDA = prefixer.DataAggregatorPrefixer

// DataAggrDoc will be used to save the aggregation process in memory
// It will be usefull if one API has to work several times or recover after
// a crash.
type DataAggrDoc struct {
	DataAggrDocID  string        `json:"_id,omitempty"`
	DataAggrDocRev string        `json:"_rev,omitempty"`
	Input          query.InputDA `json:"input,omitempty"`
}

// ID returns the DataAggrDocID
func (t *DataAggrDoc) ID() string { return t.DataAggrDocID }

// Rev returns DataAggrDocRev
func (t *DataAggrDoc) Rev() string { return t.DataAggrDocRev }

// DocType returns doc's doctype
func (t *DataAggrDoc) DocType() string { return "io.cozy.aggregation" }

// Clone copy the doc
func (t *DataAggrDoc) Clone() couchdb.Doc {
	cloned := *t
	return &cloned
}

// SetID sets the doc's ID
func (t *DataAggrDoc) SetID(id string) { t.DataAggrDocID = id }

// SetRev sets Doc's Rev
func (t *DataAggrDoc) SetRev(rev string) { t.DataAggrDocRev = rev }

func applyAggregateFunction(data []map[string]interface{}, function query.AggregationFunction) (map[string]interface{}, error) {

	switch function.Function {
	case "sum":
		return aggregations.Sum(data, function.Args)
	default:
		return nil, errors.New("Aggregation function unknown")
	}
}

func decryptInputDA(in *query.InputDA) (query.AggregationFunction, []map[string]interface{}, error) {

	// TODO: Decrypt if encrypted

	var function query.AggregationFunction
	var data []map[string]interface{}
	if err := json.Unmarshal(in.EncryptedFunction, &function); err != nil {
		return function, data, errors.New("Failed to unmarshal function : " + err.Error())
	}
	if err := json.Unmarshal(in.EncryptedData, &data); err != nil {
		return function, data, errors.New("Failed to unmarshal data : " + err.Error())
	}

	return function, data, nil
}

// AggregateData leads an aggregation of data
func AggregateData(in query.InputDA) (map[string]interface{}, error) {

	var results map[string]interface{}

	function, data, err := decryptInputDA(&in)
	if err != nil {
		return results, err
	}

	results, err = applyAggregateFunction(data, function)
	if err != nil {
		return results, errors.New("Failed to apply aggregate function : " + err.Error())
	}
	return results, nil
}

// GetStateOrGetResult returns at any time the state of the algorithm
func GetStateOrGetResult(id string) ([]string, error) {
	couchdb.EnsureDBExist(prefixerDA, "io.cozy.aggregation")
	fetched := &DataAggrDoc{}
	err := couchdb.GetDoc(prefixerDA, "io.cozy.aggregation", id, fetched)
	if err != nil {
		return []string{}, err
	}

	return []string{}, err
}
