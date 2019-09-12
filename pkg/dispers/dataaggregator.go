package enclave

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/cozy/cozy-stack/pkg/dispers/aggregation"
	"github.com/cozy/cozy-stack/pkg/dispers/aggregation/functions"
	"github.com/cozy/cozy-stack/pkg/dispers/aggregation/patches"
	"github.com/cozy/cozy-stack/pkg/dispers/errors"
	"github.com/cozy/cozy-stack/pkg/dispers/query"
	"github.com/cozy/cozy-stack/pkg/prefixer"
)

var prefixerDA = prefixer.DataAggregatorPrefixer

func applyAggregateFunction(indexRow int, results *map[string]interface{}, rowData map[string]interface{}, function query.AggregationFunction) error {

	// Every aggegation function have the same structure in input or output
	// Results that will be returned are given in input and modified step by step
	// Every parameter has to be set in args (ex : Args[weigth]=length)
	// The Results map is shared by every functions.
	// It allows to build a treatment with several functions
	switch function.Function {
	// From univariate.go
	case "sum":
		return functions.Sum(results, rowData, function.Args)
	case "sum_square":
		return functions.SumSquare(results, rowData, function.Args)
	case "min":
		return functions.Min(results, rowData, function.Args)
	case "max":
		return functions.Max(results, rowData, function.Args)
	// From categ.go
	case "preprocess":
		return functions.Preprocessing(results, rowData, function.Args)
	case "logit_map":
		return functions.LogisticRegressionMap(results, rowData, function.Args)
	case "logit_reduce":
		return functions.LogisticRegressionReduce(results, rowData, function.Args)
	default:
		return errors.WrapErrors(errors.ErrAggrUnknown, function.Function)
	}
}

func applyAggregatePatch(results *map[string]interface{}, patch query.AggregationPatch) error {

	switch patch.Patch {
	// From univariate.go
	case "division":
		return patches.Division(results, patch.Args)
	case "standard_deviation":
		return patches.StandardDeviation(results, patch.Args)
	default:
		return errors.WrapErrors(errors.ErrPatchUnknown, patch.Patch)
	}
}

func decryptInputDA(in *query.InputDA) ([]query.AggregationJob, []map[string]interface{}, error) {

	// TODO: Decrypt bytes if encrypted

	// Unmarshal bytes
	var jobs []query.AggregationJob
	var data []map[string]interface{}
	if err := json.Unmarshal(in.EncryptedJobs, &jobs); err != nil {
		return jobs, data, errors.WrapErrors(errors.ErrUnmarshal, "")
	}
	if err := json.Unmarshal(in.EncryptedData, &data); err != nil {
		return jobs, data, errors.WrapErrors(errors.ErrUnmarshal, "")
	}

	return jobs, data, nil
}

func decodeAggregationJobs(job query.AggregationJob, functions *[]query.AggregationFunction, patches *[]query.AggregationPatch) error {

	pendingFunctions := []query.AggregationFunction{}
	pendingPatches := []query.AggregationPatch{}

	// Decode AggregationJob
	switch job.Job {
	// From univariate.go
	case "sum":
		if err := aggregations.NeedArgs(job.Args, "key"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "sum_square":
		if err := aggregations.NeedArgs(job.Args, "key"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "mean":
		if err := aggregations.NeedArgs(job.Args, "sum"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{
			Function: "sum",
			Args: map[string]interface{}{
				"key": job.Args["sum"].(string),
			},
		})
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{
			Function: "sum",
			Args: map[string]interface{}{
				"key": "length",
			},
		})
		// We retrieve from args the keys on which compute sum
		pendingPatches = append(pendingPatches, query.AggregationPatch{
			Patch: "division",
			Args: map[string]interface{}{
				"keyNumerator":   "sum_" + job.Args["sum"].(string),
				"keyDenominator": "sum_length",
				"keyResult":      strings.ReplaceAll(job.Args["sum"].(string), "sum", "mean"),
			},
		})
	case "standard_deviation":
		if err := aggregations.NeedArgs(job.Args, "sum", "sum_square"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{
			Function: "sum",
			Args: map[string]interface{}{
				"key": job.Args["sum"].(string),
			},
		})
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{
			Function: "sum",
			Args: map[string]interface{}{
				"key": job.Args["sum_square"].(string),
			},
		})
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{
			Function: "sum",
			Args: map[string]interface{}{
				"key": "length",
			},
		})

		pendingPatches = append(pendingPatches, query.AggregationPatch{
			Patch: "standard_deviation",
			Args: map[string]interface{}{
				"sum":        "sum_" + job.Args["sum"].(string),
				"sum_square": "sum_" + job.Args["sum_square"].(string),
				"length":     "sum_length",
				"keyResult":  strings.ReplaceAll(job.Args["sum"].(string), "sum", "std"),
			},
		})

	case "min":
		if err := aggregations.NeedArgs(job.Args, "key"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "max":
		if err := aggregations.NeedArgs(job.Args, "key"); err != nil {
			return err
		}
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "preprocess":
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "logit_map":
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
	case "logit_reduce":
		pendingFunctions = append(pendingFunctions, query.AggregationFunction{Function: job.Job, Args: job.Args})
		pendingPatches = append(pendingPatches, query.AggregationPatch{Patch: "logit_update", Args: job.Args})
	default:
		return errors.ErrJobUnknown
	}

	// Check that funcs and patches are not scheduled yet
	for _, pendingFunction := range pendingFunctions {
		idxFunction := 0
		for idxFunction < len(*functions) && !((*functions)[idxFunction].Function == pendingFunction.Function && reflect.DeepEqual((*functions)[idxFunction].Args, pendingFunction.Args)) {
			idxFunction = idxFunction + 1
		}
		if idxFunction == len(*functions) {
			// Add AggregationFunction
			*functions = append(*functions, pendingFunction)
		}
	}
	for _, pendingPatch := range pendingPatches {
		idxPatch := 0
		for idxPatch < len(*patches) && !((*patches)[idxPatch].Patch == pendingPatch.Patch && reflect.DeepEqual((*patches)[idxPatch].Args, pendingPatch.Args)) {
			idxPatch = idxPatch + 1
		}
		if idxPatch == len(*patches) {
			// Add AggregationPatch
			*patches = append(*patches, pendingPatch)
		}
	}

	return nil
}

// AggregateData leads an aggregation of data
func AggregateData(in query.InputDA) (map[string]interface{}, error) {

	// Creation of the output map
	results := make(map[string]interface{})
	// Creation of the array of AggregationFunctions to apply
	funcs := []query.AggregationFunction{}
	// Creation of the array of AggregationPatches to apply
	patches := []query.AggregationPatch{}

	// Decrypt inputs
	jobs, data, err := decryptInputDA(&in)
	if err != nil {
		return nil, err
	}

	// Stack functions and patches to compute
	for _, job := range jobs {
		err = decodeAggregationJobs(job, &funcs, &patches)
		if err != nil {
			return nil, err
		}
	}

	// Add length to results
	// Warning : due to that line, aggregation functions should not returns a result with key "length"
	results["length"] = len(data)

	// Go through aggregation functions
	for _, function := range funcs {
		// Go through Data
		for index, rowData := range data {
			err = applyAggregateFunction(index, &results, rowData, function)
			if err != nil {
				return results, errors.WrapErrors(errors.ErrAggrFailed, "")
			}
		}
	}

	// Go through aggregation patches
	for _, patch := range patches {
		err = applyAggregatePatch(&results, patch)
		if err != nil {
			return results, errors.WrapErrors(errors.ErrAggrFailed, "")
		}
	}

	return results, nil
}
