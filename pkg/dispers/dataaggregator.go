package enclave

import (
	"encoding/json"

	"github.com/cozy/cozy-stack/pkg/dispers/aggregations"
	"github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var prefixerDA = prefixer.DataAggregatorPrefixer

func applyAggregateFunction(indexRow int, results map[string]interface{}, rowData map[string]interface{}, function query.AggregationFunction) (map[string]interface{}, error) {

	// Every aggegation function have the same structure in input or output
	// Results that will be returned are given in input and modified step by step
	// Every parameter has to be set in args (ex : Args[weigth]=length)
	// The Results map is shared by every functions.
	// It allows to build a treatment with several functions
	switch function.Function {
	// From univariate.go
	case "sum":
		return aggregations.Sum(results, rowData, function.Args)
	case "sum_square":
		return aggregations.SumSquare(results, rowData, function.Args)
	case "min":
		return aggregations.Min(results, rowData, function.Args)
	case "max":
		return aggregations.Max(results, rowData, function.Args)
	// From categ.go
	case "preprocess":
		return aggregations.Preprocessing(results, rowData, function.Args)
	case "logit_map":
		return aggregations.LogisticRegressionMap(results, rowData, function.Args)
	case "logit_reduce":
		return aggregations.LogisticRegressionReduce(results, rowData, function.Args)
	case "logit_update":
		return aggregations.LogisticRegressionUpdateParameters(results, rowData, function.Args)
	default:
		return nil, errors.WrapErrors(errors.ErrAggrUnknown, function.Function)
	}
}

func decryptInputDA(in *query.InputDA) ([]query.AggregationFunction, []map[string]interface{}, error) {

	// TODO: Decrypt bytes if encrypted

	// Unmarshal bytes
	var functions []query.AggregationFunction
	var data []map[string]interface{}
	if err := json.Unmarshal(in.EncryptedFunctions, &functions); err != nil {
		return functions, data, errors.WrapErrors(errors.ErrUnmarshal, "")
	}
	if err := json.Unmarshal(in.EncryptedData, &data); err != nil {
		return functions, data, errors.WrapErrors(errors.ErrUnmarshal, "")
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

	// Add length to results
	// Warning : due to that line, aggregation functions should not returns a result with key "length"
	results["length"] = len(data)

	// Go through aggregation functions
	for _, function := range functions {
		// Go through Data
		for index, rowData := range data {
			results, err = applyAggregateFunction(index, results, rowData, function)
			if err != nil {
				return results, errors.WrapErrors(errors.ErrAggrFailed, "")
			}
		}
	}

	return results, nil
}
