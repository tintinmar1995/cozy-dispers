package patches

import (
	"math"

	"github.com/cozy/cozy-stack/pkg/dispers/aggregation"
)

// Division computes x/y
func Division(results *map[string]interface{}, args map[string]interface{}) error {

	num, err := aggregations.AsFloat64((*results)[args["keyNumerator"].(string)])
	if err != nil {
		return err
	}

	denum, err := aggregations.AsFloat64((*results)[args["keyDenominator"].(string)])
	if err != nil {
		return err
	}

	(*results)[args["keyResult"].(string)] = num / denum
	return nil
}

// StandardDeviation computes sqrt(E(X²) - E(X)²)
func StandardDeviation(results *map[string]interface{}, args map[string]interface{}) error {

	sum, err := aggregations.AsFloat64((*results)[args["sum"].(string)])
	if err != nil {
		return err
	}

	sumSquare, err := aggregations.AsFloat64((*results)[args["sum_square"].(string)])
	if err != nil {
		return err
	}

	length, err := aggregations.AsFloat64((*results)[args["length"].(string)])
	if err != nil {
		return err
	}

	(*results)[args["keyResult"].(string)] = math.Sqrt(sumSquare/length - sum/length*sum/length)

	return nil
}
