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

func applyAggregateFunction(indexRow int, previousRes map[string]interface{}, rowData map[string]interface{}, function query.AggregationFunction) (map[string]interface{}, error) {

	switch function.Function {
	case "sum":
		return aggregations.Sum(previousRes, rowData, function.Args)
	case "sum_square":
		return aggregations.SumSquare(previousRes, rowData, function.Args)
	case "min":
		return aggregations.Min(indexRow, previousRes, rowData, function.Args)
	case "max":
		return aggregations.Max(indexRow, previousRes, rowData, function.Args)
	default:
		return nil, errors.New("Unknown aggregation function")
	}
}

func decryptInputDA(in *query.InputDA) ([]query.AggregationFunction, []map[string]interface{}, error) {

	// TODO: Decrypt if encrypted

	var functions []query.AggregationFunction
	var data []map[string]interface{}
	if err := json.Unmarshal(in.EncryptedFunctions, &functions); err != nil {
		return functions, data, errors.New("Failed to unmarshal functions : " + err.Error())
	}
	if err := json.Unmarshal(in.EncryptedData, &data); err != nil {
		return functions, data, errors.New("Failed to unmarshal data : " + err.Error())
	}

	return functions, data, nil
}

// AggregateData leads an aggregation of data
func AggregateData(in query.InputDA) (map[string]interface{}, error) {

	// Creation of the output map
	results := make(map[string]interface{})

	// Decrypt inputs
	functions, data, err := decryptInputDA(&in)
	if err != nil {
		return nil, err
	}

	// Go through Data
	for index, rowData := range data {
		// Go through aggregation functions
		for _, function := range functions {
			results, err = applyAggregateFunction(index, results, rowData, function)
			if err != nil {
				return results, errors.New("Failed to apply aggregate function " + function.Function + ": " + err.Error())
			}
		}
	}

	results["length"] = len(data)
	return results, nil
}
